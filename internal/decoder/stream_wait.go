package decoder

import (
	"errors"
	"io"

	"github.com/goccy/go-json/jsontext"
)

// A value which the buffer doesn't hold whole is read on until its end ( scanValue ). A read may wait for more
// input, as one of a network connection does, which may never come when the value is not valid: the syntax error
// of the bytes read so far is then to be reported before the read, as encoding/json, which reads a value byte by
// byte, reports it. Once a read returned fewer bytes than it could take, as a reader which had no more input at
// that time does, the end of the value is found by the decoder of jsontext, which checks the value as the input
// comes, at the cost of a scan of the value: a reader which fills every read, as one of a file or of a buffer
// does, is never waited for, and the stream finds the end by its own scan alone.

// textDecodeOptions are the options of the decoder of jsontext: what the decoders of the stream take.
var textDecodeOptions = []jsontext.Options{jsontext.AllowDuplicateNames(true), jsontext.AllowInvalidUTF8(true)}

// streamText is the decoder of jsontext of a stream and what it reads from. It doesn't refer to the stream, which
// may then be on the stack of its user.
type streamText struct {
	dec *jsontext.Decoder
	// prefix is what the stream read of the value, and r the reader, whose bytes are kept in read and given to
	// the stream after the decoder read the value.
	prefix  []byte
	r       io.Reader
	read    []byte
	readErr error
}

func (t *streamText) Read(p []byte) (int, error) {
	if len(t.prefix) > 0 {
		n := copy(p, t.prefix)
		t.prefix = t.prefix[n:]
		return n, nil
	}
	if t.readErr != nil {
		return 0, t.readErr
	}
	n, err := t.r.Read(p)
	t.read = append(t.read, p[:n]...)
	if err != nil {
		t.readErr = err
	}
	return n, err
}

// scanByText finds the end of the value at the cursor by the decoder of jsontext, which reports a syntax error as
// soon as it has read its byte.
//
//go:noinline
func (s *Stream) scanByText() (int64, error) {
	t := s.text
	if t == nil {
		t = &streamText{}
		s.text = t
	}
	t.prefix, t.r, t.read, t.readErr = s.buf[s.cursor:s.length], s.r, t.read[:0], nil
	if t.dec == nil {
		t.dec = jsontext.NewDecoder(t, textDecodeOptions...)
	} else {
		t.dec.Reset(t, textDecodeOptions...)
	}
	_, err := t.dec.ReadValue()
	// what the decoder read from the reader is the stream's, as if it had read it
	s.appendRead(t.read)
	if t.readErr != nil {
		s.readErr = t.readErr
	}
	t.prefix, t.r = nil, nil
	if err == nil {
		return s.cursor + t.dec.InputOffset(), nil
	}
	var serr *jsontext.SyntacticError
	if errors.As(err, &serr) && serr.Err != io.ErrUnexpectedEOF {
		// the error of the bytes read, as the stream reports it
		if verr := s.valueSyntaxError(); verr != nil {
			return 0, verr
		}
	}
	return 0, s.truncatedValueError()
}

// appendRead appends b, read from the reader, to the data of the buffer.
func (s *Stream) appendRead(b []byte) {
	for len(b) > 0 {
		if int64(len(s.buf))-s.length <= bufPadding {
			s.grow()
		}
		n := copy(s.buf[s.length:int64(len(s.buf))-bufPadding], b)
		s.length += int64(n)
		b = b[n:]
	}
	s.buf[s.length] = nul
}

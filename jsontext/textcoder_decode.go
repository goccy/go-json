package jsontext

import (
	"bytes"
	"errors"
	"io"
	"unsafe"

	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/internal/textcoder"
)

func init() {
	textcoder.AttachDecoder = attachDecoder
	textcoder.DetachDecoder = detachDecoder
	textcoder.UnmarshalPlaceAt = unmarshalPlaceAt
	textcoder.ReadError = func(err error) error { return &ioError{err: err} }
	textcoder.IsReadError = func(err error) bool {
		var e *ioError
		return errors.As(err, &e)
	}
	textcoder.AfterTopLevelError = func(in []byte, base int64, opts *options.Config) error {
		dec := NewDecoder(bytes.NewBuffer(in), opts)
		dec.d.base = base
		return dec.d.charError(0, placeAfterTopLevel, pointAt)
	}
	textcoder.ReleaseDecoder = func(dec unsafe.Pointer) {
		d := &(*Decoder)(dec).d
		d.cfg = config{}
		d.st.outer = nil
		if d.r != &attachedInput {
			d.r, d.bb = nil, nil // the reader of an input which is read on, which the decoder must not keep
		}
	}
	textcoder.DecoderPointer = func(dec unsafe.Pointer, where int) string {
		st := &(*Decoder)(dec).d.st
		switch where {
		case -1:
			return string(st.pointer(pointLast))
		case +1:
			return string(st.pointer(pointNext))
		}
		return string(st.pointer(pointAt))
	}
}

// attachedInput is the reader of an attached decoder, whose input is all in its buffer: an empty bytes.Buffer, which
// is never written to, which makes the decoder read its buffer in place, and nothing after it.
var attachedInput bytes.Buffer

// attachDecoder makes the Decoder dec read p.In from p.Pos, at the place whose innermost level is the one of p ( see
// textcoder.AttachDecoder ). The input is read in place: it is the array of a bytes.Buffer for the decoder, which
// never writes to it, and, if p has a reader, the buffer of the decoder, which it moves and reads into, once it
// needs more.
func attachDecoder(dec unsafe.Pointer, p *textcoder.DecoderAttachment) {
	d := &(*Decoder)(dec).d
	if !(p.Same && p.R == nil && d.ready() && d.r == &attachedInput) {
		// set up as reset sets up a decoder of a reader, by the options of the call, without the options of a
		// reset, which would be allocated.
		if !d.ready() {
			d.st, d.vs.scratch = stack{}, nil
		}
		d.cfg = *p.Opts
		d.r, d.bb = &attachedInput, &attachedInput
		if p.R != nil {
			d.r, d.bb = p.R, nil
		}
		d.raw = tokenSource{dec: d, form: formRaw}
		d.vs = valueScanner{
			st:         &d.st,
			validUTF8:  !d.cfg.Has(allowInvalidUTF8),
			checkNames: !d.cfg.Has(allowDuplicateNames),
			decoding:   true,
			scratch:    d.vs.scratch[:0],
		}
	}
	st := &d.st
	st.reset()
	if !p.Top {
		_ = st.push(p.Object)
	}
	// the names of an object, which the value read at its place can't have, are not given.
	st.last().count = p.Count
	st.outer, st.outerDepth, st.outerCount = p.Outer, len(st.levels)-1, p.Count
	st.base = p.OuterDepth - st.outerDepth
	d.buf, d.shared = p.In, !p.Own
	d.base, d.pos, d.prevStart, d.prevEnd = p.Base, p.Pos, p.Pos, p.Pos
	d.rerr, d.peekErr = nil, nil
	d.vs.scratch = d.vs.scratch[:0]
	p.Depth = st.depth()
}

// detachDecoder sets where dec stopped reading in p, and drops its input ( see textcoder.DetachDecoder ).
func detachDecoder(dec unsafe.Pointer, p *textcoder.DecoderAttachment) {
	d := &(*Decoder)(dec).d
	start := d.prevStart
	if start == d.prevEnd && start > 0 {
		// a delimiter or a literal, which ends there: the buffer doesn't keep it as the last token.
		if c := d.buf[start-1]; c == '{' || c == '}' || c == '[' || c == ']' {
			start--
		} else {
			for start > 0 && 'a' <= d.buf[start-1] && d.buf[start-1] <= 'z' {
				start--
			}
		}
	}
	p.Depth, p.Count, p.Pos, p.End, p.PrevStart = d.st.depth(), d.st.last().count, d.pos, d.base+int64(d.pos), d.base+int64(start)
	p.Rest = nil
	if !d.shared {
		// the decoder read on into a buffer of its own, whose bytes after the place go back to the caller
		p.Rest = d.buf[d.pos:]
		if p.Own {
			p.In = d.buf
		}
	}
	d.buf, d.vs.in = nil, nil
	d.pos, d.prevStart, d.prevEnd = 0, 0, 0
	d.peekErr = nil
}

// unmarshalPlace is textcoder.UnmarshalPlace: the input is read by a Decoder, token by token, as encoding/json/v2
// reads it, which finds its syntax errors and the pointer of the value at stop as encoding/json/v2 does.
func unmarshalPlace(in []byte, opts *options.Config, stop int64, read textcoder.Read) (string, error) {
	dec := NewDecoder(bytes.NewBuffer(in), opts)
	d := &dec.d
	for {
		if stop >= 0 {
			if pos, err := d.beforeToken(true); err == nil && d.base+int64(pos) >= stop {
				return readPlace(dec, read)
			}
		}
		if _, err := dec.ReadToken(); err != nil {
			if err == io.EOF {
				// an input of white space only, which has no value
				return "", &SyntacticError{ByteOffset: int64(len(in)), Err: io.ErrUnexpectedEOF}
			}
			return "", err
		}
		if d.st.depth() == 0 {
			// the value is read: only white space may follow it
			pos, err := d.skip(d.pos)
			if err == nil {
				return "", d.charError(pos, placeAfterTopLevel, pointAt)
			}
			return "", nil
		}
	}
}

// unmarshalPlaceAt is textcoder.UnmarshalPlaceAt: the part of the input is read by a Decoder attached at its place,
// whose stack is made the one of the whole input, so that the levels of the place may end in the part.
func unmarshalPlaceAt(p *textcoder.DecoderAttachment, stop int64, read textcoder.Read) (string, error) {
	dec := new(Decoder)
	attachDecoder(unsafe.Pointer(dec), p)
	d := &dec.d
	if d.st.outer != nil {
		d.st = *d.st.full()
	}
	for {
		if stop >= 0 {
			if pos, err := d.beforeToken(true); err == nil && d.base+int64(pos) >= stop {
				return readPlace(dec, read)
			}
		}
		if _, err := dec.ReadToken(); err != nil {
			if err == io.EOF {
				return "", nil
			}
			return "", err
		}
		if d.st.depth() == 0 {
			return "", nil
		}
	}
}

// validValueOptions are the options of validValue, by AllowDuplicateNames and AllowInvalidUTF8, made once so that a
// call makes no Options.
var validValueOptions = func() [2][2][]Options {
	var opts [2][2][]Options
	for i, dup := range []bool{false, true} {
		for j, utf8 := range []bool{false, true} {
			opts[i][j] = []Options{AllowDuplicateNames(dup), AllowInvalidUTF8(utf8)}
		}
	}
	return opts
}()

// validValue is textcoder.ValidValue: v is checked as Value.IsValid checks it.
func validValue(v []byte, allowDuplicateNames, allowInvalidUTF8 bool) bool {
	var i, j int
	if allowDuplicateNames {
		i = 1
	}
	if allowInvalidUTF8 {
		j = 1
	}
	e := getEncoder(validValueOptions[i][j])
	defer putEncoder(e)
	e.vs.write, e.vs.spaced = false, false
	return e.writeValue(v) == nil
}

// readPlace reads the value at the start of the input of dec as read says, and returns the syntax error it meets,
// or else the pointer of the value.
func readPlace(dec *Decoder, read textcoder.Read) (string, error) {
	switch read {
	case textcoder.ReadNothing:
		return string(dec.d.st.pointer(pointNext)), nil
	case textcoder.ReadValue:
		if _, err := dec.ReadValue(); err != nil {
			return "", err
		}
	default:
		if _, err := dec.ReadToken(); err != nil {
			return "", err
		}
	}
	return string(dec.d.st.pointer(pointLast)), nil
}

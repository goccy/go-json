package jsontext

import (
	"bytes"
	"io"

	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/internal/textcoder"
)

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

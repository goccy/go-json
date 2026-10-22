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

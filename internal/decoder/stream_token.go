package decoder

import (
	"encoding/json"
	"strconv"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

// The tokens of a stream are read as encoding/json reads them: the stream keeps where it is in the objects and
// the arrays which Token opened, so that a delimiter is checked by the grammar, and a value which Decode reads
// between the tokens follows the comma or the colon before it.

// tokenState is where the stream is between the tokens of the objects and the arrays which Token opened.
type tokenState uint8

const (
	// tokenTopValue is out of any: the values of the stream follow each other.
	tokenTopValue tokenState = iota
	tokenArrayStart
	tokenArrayValue
	tokenArrayComma
	tokenObjectStart
	tokenObjectKey
	tokenObjectColon
	tokenObjectValue
	tokenObjectComma
	// tokenFailed is after a read which failed: every read returns its error.
	tokenFailed
)

// tokenValueEnd moves the state after a value, which is the key of an object where one is waited for, as
// encoding/json of Go 1.27 reads it.
func (s *Stream) tokenValueEnd() {
	switch s.tokenState {
	case tokenArrayStart, tokenArrayValue:
		s.tokenState = tokenArrayComma
	case tokenObjectValue:
		s.tokenState = tokenObjectComma
	case tokenObjectStart, tokenObjectKey:
		s.tokenState = tokenObjectColon
	}
}

// tokenValue returns the value at the cursor, which starts with c, as the token of a string, a number, a bool or
// null.
func (s *Stream) tokenValue(c byte) (any, error) {
	switch c {
	case '"':
		return s.tokenString()
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		end, err := s.scanValue()
		if err != nil {
			return nil, err
		}
		literal := s.buf[s.cursor:end]
		s.cursor = end
		f64, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&literal)), 64)
		if err != nil {
			// A number out of the range of float64 is still a number.
			if numErr, ok := err.(*strconv.NumError); !ok || numErr.Err != strconv.ErrRange {
				return nil, errors.ErrSyntax(err.Error(), s.TotalOffset())
			}
		}
		if (s.Option.Flags & UseNumberOption) != 0 {
			return json.Number(literal), nil
		}
		return f64, nil
	case 't', 'f', 'n':
		end, err := s.scanValue()
		if err != nil {
			return nil, err
		}
		literal := string(s.buf[s.cursor:end])
		s.cursor = end
		switch literal {
		case "true":
			return true, nil
		case "false":
			return false, nil
		case "null":
			return nil, nil
		}
		return nil, errors.ErrInvalidCharacter(literal[0], "token", s.TotalOffset())
	}
	return nil, s.tokenError(c)
}

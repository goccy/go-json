//go:build go1.27 && goexperiment.jsonv2

package decoder

import (
	"encoding/json"
	"io"

	"github.com/goccy/go-json/internal/errors"
)

// The tokens of a stream as encoding/json of Go 1.27 reads them, by encoding/json/jsontext: a token or a value is
// read after the comma or the colon which the grammar has before it, which is checked against the kind of the
// token, and a read which fails leaves the stream where it was. A value may be read where a key is waited for,
// as the key.

// tokenEnded is after Token read the end of the input: every read returns io.EOF.
const tokenEnded = tokenFailed + 1

// needDelim returns the comma or the colon which the grammar has before a token of the kind which starts with c,
// or 0 for none.
func (s *Stream) needDelim(c byte) byte {
	switch s.tokenState {
	case tokenObjectColon:
		return ':'
	case tokenArrayComma, tokenObjectComma:
		if c != '}' && c != ']' {
			return ','
		}
	}
	return 0
}

// readDelim reads, from start, the white spaces and the comma or the colon before the next token, which it checks
// against the token, and moves the state after it. It returns the byte which starts the token.
func (s *Stream) readDelim(start int64) (byte, error) {
	if !s.skipWhiteSpace() {
		return 0, s.tokenEnd(start)
	}
	var delim byte
	at := s.TotalOffset()
	if c := s.buf[s.cursor]; c == ',' || c == ':' {
		delim = c
		s.cursor++
		if !s.skipWhiteSpace() {
			// the next token is not known: a string, which follows any comma or colon, is taken for it
			if s.needDelim('"') != delim {
				return 0, s.delimError(delim, '"', at)
			}
			return 0, s.tokenEnd(start)
		}
	}
	c := s.buf[s.cursor]
	if s.needDelim(c) != delim {
		// the byte after the white spaces: the comma or the colon, or else the token
		bad := c
		if delim != 0 {
			bad = delim
		}
		return 0, s.delimError(bad, c, at)
	}
	switch {
	case delim == ',' && s.tokenState == tokenArrayComma:
		s.tokenState = tokenArrayValue
	case delim == ',':
		s.tokenState = tokenObjectKey
	case delim == ':':
		s.tokenState = tokenObjectValue
	}
	return c, nil
}

// delimError returns the error of the byte c at the total offset at, which is not the comma or the colon which the
// grammar has before a token which starts with next.
func (s *Stream) delimError(c, next byte, at int64) error {
	where := "looking for beginning of value"
	switch s.needDelim(next) {
	case ':':
		where = "after object key"
	case ',':
		if s.tokenState == tokenObjectComma {
			where = "after object key:value pair"
		} else {
			where = "after array element"
		}
	}
	return errors.ErrSyntax("invalid character "+quoteChar(c)+" "+where, at+1)
}

// tokenEnd returns the error of the end of the input before a token: io.EOF if what is left is white spaces,
// commas and colons, which a value may follow, as encoding/json reports it.
func (s *Stream) tokenEnd(start int64) error {
	if err := s.endError(); err != io.EOF {
		return err
	}
	for i := start - s.offset; i < s.length; i++ {
		switch s.buf[i] {
		case ' ', '\t', '\n', '\r', ',', ':':
		default:
			return io.ErrUnexpectedEOF
		}
	}
	return io.EOF
}

// restore moves the stream back to the total offset start, where a read which failed started, which InputOffset
// reports, and returns err.
func (s *Stream) restore(start int64, err error) error {
	s.moveBack(start)
	return err
}

// moveBack moves the stream back to the total offset start, where a read which failed started, which InputOffset
// reports.
func (s *Stream) moveBack(start int64) {
	if start >= s.offset {
		s.cursor = start - s.offset
	}
	s.errOffset = start
}

// prepareInTokens moves the cursor to the value which Decode reads between the tokens: after the comma or the
// colon which the grammar has before it. A read which fails is kept, as encoding/json keeps it. It is not
// inlined, so that Decode keeps the size it had.
//
//go:noinline
func (s *Stream) prepareInTokens() error {
	if s.tokenState == tokenFailed || s.tokenState == tokenEnded {
		return s.err
	}
	start := s.TotalOffset()
	c, err := s.readDelim(start)
	if err == nil {
		err = s.valueStartError(c)
	}
	if err != nil {
		if err == io.EOF {
			// the end of the input in an object or an array
			err = io.ErrUnexpectedEOF
		}
		s.moveBack(start)
		s.err, s.tokenState = err, tokenFailed
		return err
	}
	return nil
}

// valueStartError returns the error of a value which starts with c where it can't be read: the end of an object or
// an array, or a value which is not a string where a key is waited for.
func (s *Stream) valueStartError(c byte) error {
	at := s.TotalOffset()
	switch {
	case c == '}' || c == ']':
		return s.mismatchError(c, at)
	case kindOf(c) == noValue:
		// a byte which starts no token
		return errors.ErrSyntax("invalid character "+quoteChar(c)+" looking for beginning of value", at+1)
	case c != '"' && (s.tokenState == tokenObjectStart || s.tokenState == tokenObjectKey):
		return errors.ErrSyntax("object member name must be a string", at+1)
	}
	return nil
}

// mismatchError returns the error of the end of an object or an array c at the total offset at, which doesn't end
// the one which is open.
func (s *Stream) mismatchError(c byte, at int64) error {
	where := "looking for beginning of value"
	switch s.tokenState {
	case tokenArrayComma:
		where = "after array element"
	case tokenObjectComma:
		where = "after object key:value pair"
	}
	return errors.ErrSyntax("invalid character "+quoteChar(c)+" "+where, at+1)
}

// Token returns the next token of the stream: a delimiter, a string, a number, a bool or nil. The commas and the
// colons are consumed where the grammar has them.
func (s *Stream) Token() (any, error) {
	if s.tokenState == tokenFailed || s.tokenState == tokenEnded {
		return nil, s.err
	}
	start := s.TotalOffset()
	c, err := s.readDelim(start)
	if err != nil {
		if err == io.EOF {
			// InputOffset reports the white spaces before the end, and a comma or a colon after them
			s.moveBack(start)
			s.errOffset = s.TotalOffset() + int64(len(s.buf[s.cursor:s.length])) - int64(len(trimLeftSpaces(s.buf[s.cursor:s.length])))
			if rest := trimLeftSpaces(s.buf[s.cursor:s.length]); len(rest) > 0 && (rest[0] == ',' || rest[0] == ':') {
				s.errOffset++
			}
			s.tokenState, s.err = tokenEnded, err
			return nil, err
		}
		return nil, s.restore(start, err)
	}
	at := s.TotalOffset()
	switch c {
	case '[', '{':
		if s.tokenState == tokenObjectStart || s.tokenState == tokenObjectKey {
			return nil, s.restore(start, errors.ErrSyntax("object member name must be a string", at+1))
		}
		s.cursor++
		s.tokenStack = append(s.tokenStack, s.tokenState)
		if c == '[' {
			s.tokenState = tokenArrayStart
		} else {
			s.tokenState = tokenObjectStart
		}
		return json.Delim(c), nil
	case ']', '}':
		open := tokenArrayStart
		if c == '}' {
			open = tokenObjectStart
		}
		if len(s.tokenStack) == 0 || !s.inOpen(open) {
			return nil, s.restore(start, s.mismatchError(c, at))
		}
		if s.tokenState == tokenObjectValue {
			return nil, s.restore(start, errors.ErrSyntax("missing value after object key", at+1))
		}
		s.cursor++
		s.tokenState = s.tokenStack[len(s.tokenStack)-1]
		s.tokenStack = s.tokenStack[:len(s.tokenStack)-1]
		s.tokenValueEnd()
		return json.Delim(c), nil
	case '"':
		if s.tokenState == tokenObjectStart || s.tokenState == tokenObjectKey {
			key, err := s.tokenString()
			if err != nil {
				return nil, s.restore(start, err)
			}
			s.tokenState = tokenObjectColon
			return key, nil
		}
	default:
		if kindOf(c) != noValue && (s.tokenState == tokenObjectStart || s.tokenState == tokenObjectKey) {
			return nil, s.restore(start, errors.ErrSyntax("object member name must be a string", at+1))
		}
	}
	v, err := s.tokenValue(c)
	if err != nil {
		return nil, s.restore(start, err)
	}
	s.tokenValueEnd()
	return v, nil
}

// inOpen reports whether the innermost object or array which is open is of the kind of the state open.
func (s *Stream) inOpen(open tokenState) bool {
	switch s.tokenState {
	case tokenArrayStart, tokenArrayValue, tokenArrayComma:
		return open == tokenArrayStart
	}
	return open == tokenObjectStart
}

// tokenError returns the syntax error of the byte c at the cursor, which the grammar doesn't have in the state.
func (s *Stream) tokenError(c byte) error {
	return errors.ErrSyntax("invalid character "+quoteChar(c)+" looking for beginning of value", s.TotalOffset()+1)
}

// InputOffset returns the offset of the stream which Decoder.InputOffset reports: the end of the last token or
// value read, or where a read which failed started.
func (s *Stream) InputOffset() int64 {
	if s.tokenState == tokenFailed || s.tokenState == tokenEnded {
		return s.errOffset
	}
	return s.TotalOffset()
}

// More reports whether the current array or object has another element: at the end of the input in one, it
// reports true, and the error of the end is returned by the next read, as encoding/json of Go 1.27 does.
func (s *Stream) More() bool {
	if s.tokenState == tokenFailed || s.tokenState == tokenEnded {
		return s.err != io.EOF
	}
	if !s.skipWhiteSpace() {
		if len(s.tokenStack) > 0 {
			s.err, s.errOffset, s.tokenState = errors.ErrSyntax("unexpected end of JSON input", s.TotalOffset()), s.TotalOffset(), tokenFailed
			return true
		}
		return false
	}
	switch s.buf[s.cursor] {
	case ']', '}':
		return false
	}
	return true
}

func trimLeftSpaces(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t' || b[0] == '\n' || b[0] == '\r') {
		b = b[1:]
	}
	return b
}

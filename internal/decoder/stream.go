package decoder

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

const (
	initBufSize = 512
)

// Token() comma/colon elision state, matching encoding/json.Decoder.Token.
const (
	tokenTopValue = iota
	tokenArrayStart
	tokenArrayValue
	tokenArrayComma
	tokenObjectStart
	tokenObjectKey
	tokenObjectColon
	tokenObjectValue
	tokenObjectComma
)

type Stream struct {
	buf                   []byte
	bufSize               int64
	length                int64
	r                     io.Reader
	offset                int64
	cursor                int64
	filledBuffer          bool
	allRead               bool
	UseNumber             bool
	DisallowUnknownFields bool
	Option                *Option
	tokenState            int
	tokenStack            []int
}

func NewStream(r io.Reader) *Stream {
	return &Stream{
		r:       r,
		bufSize: initBufSize,
		buf:     make([]byte, initBufSize),
		Option:  &Option{},
	}
}

func (s *Stream) TotalOffset() int64 {
	return s.totalOffset()
}

func (s *Stream) Buffered() io.Reader {
	buflen := int64(len(s.buf))
	for i := s.cursor; i < buflen; i++ {
		if s.buf[i] == nul {
			return bytes.NewReader(s.buf[s.cursor:i])
		}
	}
	return bytes.NewReader(s.buf[s.cursor:])
}

func (s *Stream) skipSpace() error {
	for {
		switch s.char() {
		case ' ', '\t', '\r', '\n':
			s.cursor++
		case nul:
			if s.read() {
				continue
			}
			return io.EOF
		default:
			return nil
		}
	}
}

func (s *Stream) tokenValueAllowed() bool {
	switch s.tokenState {
	case tokenTopValue, tokenArrayStart, tokenArrayValue, tokenObjectValue:
		return true
	}
	return false
}

func (s *Stream) TokenValueAllowed() bool { return s.tokenValueAllowed() }

func (s *Stream) tokenValueEnd() {
	switch s.tokenState {
	case tokenArrayStart, tokenArrayValue:
		s.tokenState = tokenArrayComma
	case tokenObjectValue:
		s.tokenState = tokenObjectComma
	}
}

func (s *Stream) TokenValueEnd() { s.tokenValueEnd() }

func quoteChar(c byte) string {
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}
	q := strconv.Quote(string([]byte{c}))
	return "'" + q[1:len(q)-1] + "'"
}

func (s *Stream) tokenError(c byte) error {
	var context string
	switch s.tokenState {
	case tokenTopValue, tokenArrayStart, tokenArrayValue, tokenObjectValue:
		context = " looking for beginning of value"
	case tokenArrayComma:
		context = " after array element"
	case tokenObjectStart, tokenObjectKey:
		context = " looking for beginning of object key string"
	case tokenObjectColon:
		context = " after object key"
	case tokenObjectComma:
		context = " after object key:value pair"
	}
	return errors.ErrSyntax("invalid character "+quoteChar(c)+context, s.totalOffset())
}

func (s *Stream) PrepareForDecode() error {
	if err := s.skipSpace(); err != nil {
		return err
	}
	switch s.tokenState {
	case tokenArrayComma:
		if s.char() != ',' {
			return errors.ErrExpected("comma after array element", s.totalOffset())
		}
		s.cursor++
		s.tokenState = tokenArrayValue
		return nil
	case tokenObjectColon:
		if s.char() != ':' {
			return errors.ErrExpected("colon after object key", s.totalOffset())
		}
		s.cursor++
		s.tokenState = tokenObjectValue
		return nil
	default:
		switch s.char() {
		case ',', ':':
			s.cursor++
		}
		return nil
	}
}

func (s *Stream) totalOffset() int64 {
	return s.offset + s.cursor
}

func (s *Stream) char() byte {
	return s.buf[s.cursor]
}

func (s *Stream) equalChar(c byte) bool {
	cur := s.buf[s.cursor]
	if cur == nul {
		s.read()
		cur = s.buf[s.cursor]
	}
	return cur == c
}

func (s *Stream) stat() ([]byte, int64, unsafe.Pointer) {
	return s.buf, s.cursor, (*sliceHeader)(unsafe.Pointer(&s.buf)).data
}

func (s *Stream) bufptr() unsafe.Pointer {
	return (*sliceHeader)(unsafe.Pointer(&s.buf)).data
}

func (s *Stream) statForRetry() ([]byte, int64, unsafe.Pointer) {
	s.cursor-- // for retry ( because caller progress cursor position in each loop )
	return s.buf, s.cursor, (*sliceHeader)(unsafe.Pointer(&s.buf)).data
}

func (s *Stream) Reset() {
	s.reset()
	s.bufSize = int64(len(s.buf))
}

func (s *Stream) More() bool {
	for {
		switch s.char() {
		case ' ', '\n', '\r', '\t':
			s.cursor++
			continue
		case '}', ']':
			return false
		case nul:
			if s.read() {
				continue
			}
			return false
		}
		break
	}
	return true
}

func (s *Stream) Token() (interface{}, error) {
	for {
		if err := s.skipSpace(); err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, err
		}
		c := s.char()
		switch c {
		case '[':
			if !s.tokenValueAllowed() {
				return nil, s.tokenError(c)
			}
			s.cursor++
			s.tokenStack = append(s.tokenStack, s.tokenState)
			s.tokenState = tokenArrayStart
			return json.Delim('['), nil
		case ']':
			if s.tokenState != tokenArrayStart && s.tokenState != tokenArrayComma {
				return nil, s.tokenError(c)
			}
			s.cursor++
			s.tokenState = s.tokenStack[len(s.tokenStack)-1]
			s.tokenStack = s.tokenStack[:len(s.tokenStack)-1]
			s.tokenValueEnd()
			return json.Delim(']'), nil
		case '{':
			if !s.tokenValueAllowed() {
				return nil, s.tokenError(c)
			}
			s.cursor++
			s.tokenStack = append(s.tokenStack, s.tokenState)
			s.tokenState = tokenObjectStart
			return json.Delim('{'), nil
		case '}':
			if s.tokenState != tokenObjectStart && s.tokenState != tokenObjectComma {
				return nil, s.tokenError(c)
			}
			s.cursor++
			s.tokenState = s.tokenStack[len(s.tokenStack)-1]
			s.tokenStack = s.tokenStack[:len(s.tokenStack)-1]
			s.tokenValueEnd()
			return json.Delim('}'), nil
		case ':':
			if s.tokenState != tokenObjectColon {
				return nil, s.tokenError(c)
			}
			s.cursor++
			s.tokenState = tokenObjectValue
			continue
		case ',':
			if s.tokenState == tokenArrayComma {
				s.cursor++
				s.tokenState = tokenArrayValue
				continue
			}
			if s.tokenState == tokenObjectComma {
				s.cursor++
				s.tokenState = tokenObjectKey
				continue
			}
			return nil, s.tokenError(c)
		case '"':
			if s.tokenState == tokenObjectStart || s.tokenState == tokenObjectKey {
				b, err := stringBytes(s)
				if err != nil {
					return nil, err
				}
				s.tokenState = tokenObjectColon
				return string(b), nil
			}
			if !s.tokenValueAllowed() {
				return nil, s.tokenError(c)
			}
			b, err := stringBytes(s)
			if err != nil {
				return nil, err
			}
			s.tokenValueEnd()
			return string(b), nil
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if !s.tokenValueAllowed() {
				return nil, s.tokenError(c)
			}
			b := floatBytes(s)
			str := *(*string)(unsafe.Pointer(&b))
			s.tokenValueEnd()
			if s.UseNumber {
				return json.Number(str), nil
			}
			f64, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return nil, err
			}
			return f64, nil
		case 't':
			if !s.tokenValueAllowed() {
				return nil, s.tokenError(c)
			}
			if err := trueBytes(s); err != nil {
				return nil, err
			}
			s.tokenValueEnd()
			return true, nil
		case 'f':
			if !s.tokenValueAllowed() {
				return nil, s.tokenError(c)
			}
			if err := falseBytes(s); err != nil {
				return nil, err
			}
			s.tokenValueEnd()
			return false, nil
		case 'n':
			if !s.tokenValueAllowed() {
				return nil, s.tokenError(c)
			}
			if err := nullBytes(s); err != nil {
				return nil, err
			}
			s.tokenValueEnd()
			return nil, nil
		case nul:
			if s.read() {
				continue
			}
			return nil, io.EOF
		default:
			return nil, s.tokenError(c)
		}
	}
}

func (s *Stream) reset() {
	s.offset += s.cursor
	s.buf = s.buf[s.cursor:]
	s.length -= s.cursor
	s.cursor = 0
}

func (s *Stream) readBuf() []byte {
	if s.filledBuffer {
		s.bufSize *= 2
		remainBuf := s.buf
		s.buf = make([]byte, s.bufSize)
		copy(s.buf, remainBuf)
	}
	remainLen := s.length - s.cursor
	remainNotNulCharNum := int64(0)
	for i := int64(0); i < remainLen; i++ {
		if s.buf[s.cursor+i] == nul {
			break
		}
		remainNotNulCharNum++
	}
	s.length = s.cursor + remainNotNulCharNum
	return s.buf[s.cursor+remainNotNulCharNum:]
}

func (s *Stream) read() bool {
	if s.allRead {
		return false
	}
	buf := s.readBuf()
	last := len(buf) - 1
	buf[last] = nul
	n, err := s.r.Read(buf[:last])
	s.length += int64(n)
	if n == last {
		s.filledBuffer = true
	} else {
		s.filledBuffer = false
	}
	if err == io.EOF {
		s.allRead = true
	} else if err != nil {
		return false
	}
	return true
}

func (s *Stream) skipWhiteSpace() byte {
	p := s.bufptr()
LOOP:
	c := char(p, s.cursor)
	switch c {
	case ' ', '\n', '\t', '\r':
		s.cursor++
		goto LOOP
	case nul:
		if s.read() {
			p = s.bufptr()
			goto LOOP
		}
	}
	return c
}

func (s *Stream) skipObject(depth int64) error {
	braceCount := 1
	_, cursor, p := s.stat()
	for {
		switch char(p, cursor) {
		case '{':
			braceCount++
			depth++
			if depth > maxDecodeNestingDepth {
				return errors.ErrExceededMaxDepth(s.char(), s.cursor)
			}
		case '}':
			braceCount--
			depth--
			if braceCount == 0 {
				s.cursor = cursor + 1
				return nil
			}
		case '[':
			depth++
			if depth > maxDecodeNestingDepth {
				return errors.ErrExceededMaxDepth(s.char(), s.cursor)
			}
		case ']':
			depth--
		case '"':
			for {
				cursor++
				switch char(p, cursor) {
				case '\\':
					cursor++
					if char(p, cursor) == nul {
						s.cursor = cursor
						if s.read() {
							_, cursor, p = s.stat()
							continue
						}
						return errors.ErrUnexpectedEndOfJSON("string of object", cursor)
					}
				case '"':
					goto SWITCH_OUT
				case nul:
					s.cursor = cursor
					if s.read() {
						_, cursor, p = s.statForRetry()
						continue
					}
					return errors.ErrUnexpectedEndOfJSON("string of object", cursor)
				}
			}
		case nul:
			s.cursor = cursor
			if s.read() {
				_, cursor, p = s.stat()
				continue
			}
			return errors.ErrUnexpectedEndOfJSON("object of object", cursor)
		}
	SWITCH_OUT:
		cursor++
	}
}

func (s *Stream) skipArray(depth int64) error {
	bracketCount := 1
	_, cursor, p := s.stat()
	for {
		switch char(p, cursor) {
		case '[':
			bracketCount++
			depth++
			if depth > maxDecodeNestingDepth {
				return errors.ErrExceededMaxDepth(s.char(), s.cursor)
			}
		case ']':
			bracketCount--
			depth--
			if bracketCount == 0 {
				s.cursor = cursor + 1
				return nil
			}
		case '{':
			depth++
			if depth > maxDecodeNestingDepth {
				return errors.ErrExceededMaxDepth(s.char(), s.cursor)
			}
		case '}':
			depth--
		case '"':
			for {
				cursor++
				switch char(p, cursor) {
				case '\\':
					cursor++
					if char(p, cursor) == nul {
						s.cursor = cursor
						if s.read() {
							_, cursor, p = s.stat()
							continue
						}
						return errors.ErrUnexpectedEndOfJSON("string of object", cursor)
					}
				case '"':
					goto SWITCH_OUT
				case nul:
					s.cursor = cursor
					if s.read() {
						_, cursor, p = s.statForRetry()
						continue
					}
					return errors.ErrUnexpectedEndOfJSON("string of object", cursor)
				}
			}
		case nul:
			s.cursor = cursor
			if s.read() {
				_, cursor, p = s.stat()
				continue
			}
			return errors.ErrUnexpectedEndOfJSON("array of object", cursor)
		}
	SWITCH_OUT:
		cursor++
	}
}

func (s *Stream) skipValue(depth int64) error {
	_, cursor, p := s.stat()
	for {
		switch char(p, cursor) {
		case ' ', '\n', '\t', '\r':
			cursor++
			continue
		case nul:
			s.cursor = cursor
			if s.read() {
				_, cursor, p = s.stat()
				continue
			}
			return errors.ErrUnexpectedEndOfJSON("value of object", s.totalOffset())
		case '{':
			s.cursor = cursor + 1
			return s.skipObject(depth + 1)
		case '[':
			s.cursor = cursor + 1
			return s.skipArray(depth + 1)
		case '"':
			for {
				cursor++
				switch char(p, cursor) {
				case '\\':
					cursor++
					if char(p, cursor) == nul {
						s.cursor = cursor
						if s.read() {
							_, cursor, p = s.stat()
							continue
						}
						return errors.ErrUnexpectedEndOfJSON("value of string", s.totalOffset())
					}
				case '"':
					s.cursor = cursor + 1
					return nil
				case nul:
					s.cursor = cursor
					if s.read() {
						_, cursor, p = s.statForRetry()
						continue
					}
					return errors.ErrUnexpectedEndOfJSON("value of string", s.totalOffset())
				}
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			for {
				cursor++
				c := char(p, cursor)
				if floatTable[c] {
					continue
				} else if c == nul {
					if s.read() {
						_, cursor, p = s.stat()
						continue
					}
				}
				s.cursor = cursor
				return nil
			}
		case 't':
			s.cursor = cursor
			if err := trueBytes(s); err != nil {
				return err
			}
			return nil
		case 'f':
			s.cursor = cursor
			if err := falseBytes(s); err != nil {
				return err
			}
			return nil
		case 'n':
			s.cursor = cursor
			if err := nullBytes(s); err != nil {
				return err
			}
			return nil
		}
		cursor++
	}
}

func nullBytes(s *Stream) error {
	// current cursor's character is 'n'
	s.cursor++
	if s.char() != 'u' {
		if err := retryReadNull(s); err != nil {
			return err
		}
	}
	s.cursor++
	if s.char() != 'l' {
		if err := retryReadNull(s); err != nil {
			return err
		}
	}
	s.cursor++
	if s.char() != 'l' {
		if err := retryReadNull(s); err != nil {
			return err
		}
	}
	s.cursor++
	return nil
}

func retryReadNull(s *Stream) error {
	if s.char() == nul && s.read() {
		return nil
	}
	return errors.ErrInvalidCharacter(s.char(), "null", s.totalOffset())
}

func trueBytes(s *Stream) error {
	// current cursor's character is 't'
	s.cursor++
	if s.char() != 'r' {
		if err := retryReadTrue(s); err != nil {
			return err
		}
	}
	s.cursor++
	if s.char() != 'u' {
		if err := retryReadTrue(s); err != nil {
			return err
		}
	}
	s.cursor++
	if s.char() != 'e' {
		if err := retryReadTrue(s); err != nil {
			return err
		}
	}
	s.cursor++
	return nil
}

func retryReadTrue(s *Stream) error {
	if s.char() == nul && s.read() {
		return nil
	}
	return errors.ErrInvalidCharacter(s.char(), "bool(true)", s.totalOffset())
}

func falseBytes(s *Stream) error {
	// current cursor's character is 'f'
	s.cursor++
	if s.char() != 'a' {
		if err := retryReadFalse(s); err != nil {
			return err
		}
	}
	s.cursor++
	if s.char() != 'l' {
		if err := retryReadFalse(s); err != nil {
			return err
		}
	}
	s.cursor++
	if s.char() != 's' {
		if err := retryReadFalse(s); err != nil {
			return err
		}
	}
	s.cursor++
	if s.char() != 'e' {
		if err := retryReadFalse(s); err != nil {
			return err
		}
	}
	s.cursor++
	return nil
}

func retryReadFalse(s *Stream) error {
	if s.char() == nul && s.read() {
		return nil
	}
	return errors.ErrInvalidCharacter(s.char(), "bool(false)", s.totalOffset())
}

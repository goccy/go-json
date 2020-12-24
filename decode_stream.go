package json

import (
	"bytes"
	"io"
)

const (
	initBufSize = 512
)

type stream struct {
	buf                   []byte
	bufSize               int64
	length                int64
	r                     io.Reader
	offset                int64
	cursor                int64
	readPos               int64
	allRead               bool
	useNumber             bool
	disallowUnknownFields bool
}

func newStream(r io.Reader) *stream {
	return &stream{
		r:       r,
		bufSize: initBufSize,
		buf:     []byte{nul},
	}
}

func (s *stream) buffered() io.Reader {
	buflen := int64(len(s.buf))
	for i := s.cursor; i < buflen; i++ {
		if s.buf[i] == nul {
			return bytes.NewReader(s.buf[s.cursor:i])
		}
	}
	return bytes.NewReader(s.buf[s.cursor:])
}

func (s *stream) totalOffset() int64 {
	return s.offset + s.cursor
}

func (s *stream) prevChar() byte {
	return s.buf[s.cursor-1]
}

func (s *stream) char() byte {
	return s.buf[s.cursor]
}

func (s *stream) reset() {
	s.offset += s.cursor
	s.buf = s.buf[s.cursor:]
	s.cursor = 0
	s.length = int64(len(s.buf))
}

func (s *stream) readBuf() []byte {
	s.bufSize *= 2
	remainBuf := s.buf
	s.buf = make([]byte, s.bufSize)
	copy(s.buf, remainBuf)
	return s.buf[s.cursor:]
}

func (s *stream) read() bool {
	if s.allRead {
		return false
	}
	buf := s.readBuf()
	last := len(buf) - 1
	buf[last] = nul
	n, err := s.r.Read(buf[:last])
	s.length = s.cursor + int64(n)
	if n < last || err == io.EOF {
		s.allRead = true
	} else if err != nil {
		return false
	}
	return true
}

func (s *stream) skipWhiteSpace() {
LOOP:
	c := s.char()
	if isWhiteSpace[c] {
		s.cursor++
		goto LOOP
	} else if c == nul {
		if s.read() {
			goto LOOP
		}
	}
}

func (s *stream) skipValue() error {
	s.skipWhiteSpace()
	braceCount := 0
	bracketCount := 0
	start := s.cursor
	for {
		switch s.char() {
		case nul:
			if s.read() {
				continue
			}
			if start == s.cursor {
				return errUnexpectedEndOfJSON("value of object", s.totalOffset())
			}
			if braceCount == 0 && bracketCount == 0 {
				return nil
			}
			return errUnexpectedEndOfJSON("value of object", s.totalOffset())
		case '{':
			braceCount++
		case '[':
			bracketCount++
		case '}':
			braceCount--
			if braceCount == -1 && bracketCount == 0 {
				return nil
			}
		case ']':
			bracketCount--
			if braceCount == 0 && bracketCount == -1 {
				return nil
			}
		case ',':
			if bracketCount == 0 && braceCount == 0 {
				return nil
			}
		case '"':
			for {
				s.cursor++
				c := s.char()
				if c == nul {
					if !s.read() {
						return errUnexpectedEndOfJSON("value of string", s.totalOffset())
					}
					c = s.char()
				}
				if c != '"' {
					continue
				}
				if s.prevChar() == '\\' {
					continue
				}
				if bracketCount == 0 && braceCount == 0 {
					s.cursor++
					return nil
				}
				break
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			for {
				s.cursor++
				c := s.char()
				if floatTable[c] {
					continue
				} else if c == nul {
					if s.read() {
						s.cursor-- // for retry current character
						continue
					}
				}
				break
			}
			if bracketCount == 0 && braceCount == 0 {
				return nil
			}
			continue
		case 't':
			if err := trueBytes(s); err != nil {
				return err
			}
			if bracketCount == 0 && braceCount == 0 {
				return nil
			}
			continue
		case 'f':
			if err := falseBytes(s); err != nil {
				return err
			}
			if bracketCount == 0 && braceCount == 0 {
				return nil
			}
			continue
		case 'n':
			if err := nullBytes(s); err != nil {
				return err
			}
			if bracketCount == 0 && braceCount == 0 {
				return nil
			}
			continue
		}
		s.cursor++
	}
	return errUnexpectedEndOfJSON("value of object", s.offset)
}

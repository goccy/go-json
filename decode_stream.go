package json

import (
	"bytes"
	"io"
)

const (
	readChunkSize = 2
)

type stream struct {
	buf     []byte
	length  int64
	r       io.Reader
	offset  int64
	cursor  int64
	allRead bool
}

func (s *stream) buffered() io.Reader {
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

func (s *stream) end() bool {
	return s.allRead && s.length <= s.cursor
}

func (s *stream) progress() bool {
	if s.cursor < s.length-1 || s.read() {
		s.cursor++
		return true
	}
	s.cursor = s.length
	return false
}

func (s *stream) progressN(n int64) bool {
	if s.cursor+n < s.length-1 || s.read() {
		s.cursor += n
		return true
	}
	s.cursor = s.length
	return false
}

func (s *stream) reset() {
	s.buf = s.buf[s.cursor:]
	s.length -= s.cursor
	s.cursor = 0
}

func (s *stream) read() bool {
	buf := make([]byte, readChunkSize)
	n, err := s.r.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	remain := s.length
	newBuf := make([]byte, remain+int64(n)+1)
	copy(newBuf, s.buf)
	copy(newBuf[remain:], buf)
	s.buf = newBuf
	s.length = int64(len(newBuf)) - 1
	s.offset += s.cursor
	if n == 0 || err == io.EOF {
		s.allRead = true
		return false
	}
	return true
}

func (s *stream) skipWhiteSpace() {
LOOP:
	if isWhiteSpace[s.char()] {
		s.progress()
		goto LOOP
	}
}

func (s *stream) skipValue() error {
	s.skipWhiteSpace()
	braceCount := 0
	bracketCount := 0
	for {
		switch s.char() {
		case '\000':
			return errUnexpectedEndOfJSON("value of object", s.offset)
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
		case ',':
			if bracketCount == 0 && braceCount == 0 {
				return nil
			}
		case '"':
			for s.progress() {
				if s.char() != '"' {
					continue
				}
				if s.prevChar() == '\\' {
					continue
				}
				if bracketCount == 0 && braceCount == 0 {
					s.progress()
					return nil
				}
				break
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			for s.progress() {
				tk := int(s.char())
				if (int('0') <= tk && tk <= int('9')) || tk == '.' || tk == 'e' || tk == 'E' {
					continue
				}
				break
			}
			if bracketCount == 0 && braceCount == 0 {
				return nil
			}
			continue
		}
		s.progress()
	}
	return errUnexpectedEndOfJSON("value of object", s.offset)
}

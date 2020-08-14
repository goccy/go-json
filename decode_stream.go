package json

import (
	"bytes"
	"io"
)

const (
	readChunkSize = 512
)

type stream struct {
	buf                   []byte
	length                int64
	r                     io.Reader
	offset                int64
	cursor                int64
	allRead               bool
	useNumber             bool
	disallowUnknownFields bool
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
	if s.allRead {
		return false
	}
	buf := make([]byte, readChunkSize)
	n, err := s.r.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	if n < readChunkSize || err == io.EOF {
		s.allRead = true
	}
	// extend buffer (2) is protect ( s.cursor++ x2 )
	// e.g.) decodeEscapeString
	const extendBufLength = int64(2)

	totalSize := s.length + int64(n) + extendBufLength
	if totalSize > readChunkSize {
		newBuf := make([]byte, totalSize)
		copy(newBuf, s.buf)
		copy(newBuf[s.length:], buf)
		s.buf = newBuf
		s.length = totalSize - extendBufLength
	} else if s.length > 0 {
		copy(buf[s.length:], buf)
		copy(buf, s.buf[:s.length])
		s.buf = buf
		s.length = totalSize - extendBufLength
	} else {
		s.buf = buf
		s.length = totalSize - extendBufLength
	}
	s.offset += s.cursor
	if n == 0 {
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
	for {
		switch s.char() {
		case nul:
			if s.read() {
				continue
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

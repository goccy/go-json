package json

import (
	"bytes"
	"io"
)

const (
	readChunkSize = 1024
)

type stream struct {
	buf        []byte
	length     int64
	r          io.Reader
	decodedPos int64
	offset     int64
	cursor     int64
}

func (s *stream) buffered() io.Reader {
	return bytes.NewReader(s.buf[s.cursor:])
}

func (s *stream) totalOffset() int64 {
	return s.offset + s.cursor
}

func (s *stream) char() byte {
	return s.buf[s.cursor]
}

func (s *stream) read() bool {
	buf := make([]byte, readChunkSize)
	n, err := s.r.Read(buf)
	if n == 0 || err == io.EOF {
		return false
	}
	remain := s.length - s.decodedPos
	newBuf := make([]byte, remain+int64(n))
	copy(newBuf, s.buf[s.decodedPos:])
	copy(newBuf[remain:], buf)
	s.buf = newBuf
	s.length = int64(len(newBuf))
	s.offset += s.decodedPos
	s.cursor = 0
	s.decodedPos = 0
	return true
}

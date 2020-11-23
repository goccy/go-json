package json

import (
	"encoding/base64"
	"unsafe"
)

type bytesDecoder struct {
	structName string
	fieldName  string
}

func newBytesDecoder(structName string, fieldName string) *bytesDecoder {
	return &bytesDecoder{structName: structName, fieldName: fieldName}
}

func (d *bytesDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	bytes, err := d.decodeStreamBinary(s)
	if err != nil {
		return err
	}
	decodedLen := base64.StdEncoding.DecodedLen(len(bytes))
	buf := make([]byte, decodedLen)
	if _, err := base64.StdEncoding.Decode(buf, bytes); err != nil {
		return err
	}
	*(*[]byte)(p) = buf
	return nil
}

func (d *bytesDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	bytes, c, err := d.decodeBinary(buf, cursor)
	if err != nil {
		return 0, err
	}
	cursor = c
	decodedLen := base64.StdEncoding.DecodedLen(len(bytes))
	b := make([]byte, decodedLen)
	if _, err := base64.StdEncoding.Decode(b, bytes); err != nil {
		return 0, err
	}
	*(*[]byte)(p) = b
	return cursor, nil
}

func binaryBytes(s *stream) ([]byte, error) {
	s.cursor++
	start := s.cursor
	for {
		switch s.char() {
		case '"':
			literal := s.buf[start:s.cursor]
			s.cursor++
			s.reset()
			return literal, nil
		case nul:
			if s.read() {
				continue
			}
			goto ERROR
		}
		s.cursor++
	}
ERROR:
	return nil, errUnexpectedEndOfJSON("[]byte", s.totalOffset())
}

func (d *bytesDecoder) decodeStreamBinary(s *stream) ([]byte, error) {
	for {
		switch s.char() {
		case ' ', '\n', '\t', '\r':
			s.cursor++
			continue
		case '"':
			return binaryBytes(s)
		case 'n':
			if err := nullBytes(s); err != nil {
				return nil, err
			}
			return []byte{}, nil
		case nul:
			if s.read() {
				continue
			}
		}
		break
	}
	return nil, errNotAtBeginningOfValue(s.totalOffset())
}

func (d *bytesDecoder) decodeBinary(buf []byte, cursor int64) ([]byte, int64, error) {
	for {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			cursor++
		case '"':
			cursor++
			start := cursor
			for {
				switch buf[cursor] {
				case '"':
					literal := buf[start:cursor]
					cursor++
					return literal, cursor, nil
				case nul:
					return nil, 0, errUnexpectedEndOfJSON("[]byte", cursor)
				}
				cursor++
			}
			return nil, 0, errUnexpectedEndOfJSON("[]byte", cursor)
		case 'n':
			buflen := int64(len(buf))
			if cursor+3 >= buflen {
				return nil, 0, errUnexpectedEndOfJSON("null", cursor)
			}
			if buf[cursor+1] != 'u' {
				return nil, 0, errInvalidCharacter(buf[cursor+1], "null", cursor)
			}
			if buf[cursor+2] != 'l' {
				return nil, 0, errInvalidCharacter(buf[cursor+2], "null", cursor)
			}
			if buf[cursor+3] != 'l' {
				return nil, 0, errInvalidCharacter(buf[cursor+3], "null", cursor)
			}
			cursor += 4
			return []byte{}, cursor, nil
		default:
			goto ERROR
		}
	}
ERROR:
	return nil, 0, errNotAtBeginningOfValue(cursor)
}

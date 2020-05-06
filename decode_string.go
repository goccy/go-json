package json

import (
	"errors"
	"unsafe"
)

type stringDecoder struct {
}

func newStringDecoder() *stringDecoder {
	return &stringDecoder{}
}

func (d *stringDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	bytes, c, err := d.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	cursor = c
	*(*string)(unsafe.Pointer(p)) = *(*string)(unsafe.Pointer(&bytes))
	return cursor, nil
}

func (d *stringDecoder) decodeByte(buf []byte, cursor int) ([]byte, int, error) {
	buflen := len(buf)
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case '"':
			cursor++
			start := cursor
			for ; cursor < buflen; cursor++ {
				switch buf[cursor] {
				case '\\':
					cursor++
				case '"':
					literal := buf[start:cursor]
					cursor++
					return literal, cursor, nil
				}
			}
			return nil, 0, errors.New("unexpected error string")
		case 'n':
			if cursor+3 >= buflen {
				return nil, 0, errors.New("unexpected error. invalid bool character")
			}
			if buf[cursor+1] != 'u' {
				return nil, 0, errors.New("unexpected error. invalid bool character")
			}
			if buf[cursor+2] != 'l' {
				return nil, 0, errors.New("unexpected error. invalid bool character")
			}
			if buf[cursor+3] != 'l' {
				return nil, 0, errors.New("unexpected error. invalid bool character")
			}
			cursor += 5
			return []byte{}, cursor, nil
		}
	}
	return nil, 0, errors.New("unexpected error key delimiter")
}

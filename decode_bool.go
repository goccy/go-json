package json

import (
	"errors"
	"unsafe"
)

type boolDecoder struct{}

func newBoolDecoder() *boolDecoder {
	return &boolDecoder{}
}

func (d *boolDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	buflen := len(buf)
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 't':
		if cursor+3 >= buflen {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+1] != 'r' {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+2] != 'u' {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+3] != 'e' {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		cursor += 4
		*(*bool)(unsafe.Pointer(p)) = true
	case 'f':
		if cursor+4 >= buflen {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+1] != 'a' {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+2] != 'l' {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+3] != 's' {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+4] != 'e' {
			return 0, errors.New("unexpected error. invalid bool character")
		}
		cursor += 5
		*(*bool)(unsafe.Pointer(p)) = false
	}
	return cursor, nil
}

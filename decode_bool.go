package json

import (
	"unsafe"
)

type boolDecoder struct{}

func newBoolDecoder() *boolDecoder {
	return &boolDecoder{}
}

func (d *boolDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	buflen := int64(len(buf))
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 't':
		if cursor+3 >= buflen {
			return 0, errUnexpectedEndOfJSON("bool(true)", cursor)
		}
		if buf[cursor+1] != 'r' {
			return 0, errInvalidCharacter(buf[cursor+1], "bool(true)", cursor)
		}
		if buf[cursor+2] != 'u' {
			return 0, errInvalidCharacter(buf[cursor+2], "bool(true)", cursor)
		}
		if buf[cursor+3] != 'e' {
			return 0, errInvalidCharacter(buf[cursor+3], "bool(true)", cursor)
		}
		cursor += 4
		*(*bool)(unsafe.Pointer(p)) = true
	case 'f':
		if cursor+4 >= buflen {
			return 0, errUnexpectedEndOfJSON("bool(false)", cursor)
		}
		if buf[cursor+1] != 'a' {
			return 0, errInvalidCharacter(buf[cursor+1], "bool(false)", cursor)
		}
		if buf[cursor+2] != 'l' {
			return 0, errInvalidCharacter(buf[cursor+2], "bool(false)", cursor)
		}
		if buf[cursor+3] != 's' {
			return 0, errInvalidCharacter(buf[cursor+3], "bool(false)", cursor)
		}
		if buf[cursor+4] != 'e' {
			return 0, errInvalidCharacter(buf[cursor+4], "bool(false)", cursor)
		}
		cursor += 5
		*(*bool)(unsafe.Pointer(p)) = false
	}
	return cursor, nil
}

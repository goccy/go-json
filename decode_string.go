package json

import (
	"unsafe"
)

type stringDecoder struct {
}

func newStringDecoder() *stringDecoder {
	return &stringDecoder{}
}

func (d *stringDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	bytes, c, err := d.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	cursor = c
	*(*string)(unsafe.Pointer(p)) = *(*string)(unsafe.Pointer(&bytes))
	return cursor, nil
}

func (d *stringDecoder) decodeByte(buf []byte, cursor int64) ([]byte, int64, error) {
	for {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			cursor++
		case '"':
			cursor++
			start := cursor
			for {
				switch buf[cursor] {
				case '\\':
					cursor++
				case '"':
					literal := buf[start:cursor]
					cursor++
					return literal, cursor, nil
				case '\000':
					return nil, 0, errUnexpectedEndOfJSON("string", cursor)
				}
				cursor++
			}
			return nil, 0, errUnexpectedEndOfJSON("string", cursor)
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
			cursor += 5
			return []byte{'n', 'u', 'l', 'l'}, cursor, nil
		default:
			goto ERROR
		}
	}
ERROR:
	return nil, 0, errNotAtBeginningOfValue(cursor)
}

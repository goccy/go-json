package json

import (
	"errors"
	"unsafe"
)

type structFieldSet struct {
	dec    decoder
	offset uintptr
}

type structDecoder struct {
	fieldMap   map[string]*structFieldSet
	keyDecoder *stringDecoder
}

func newStructDecoder(fieldMap map[string]*structFieldSet) *structDecoder {
	return &structDecoder{
		fieldMap:   fieldMap,
		keyDecoder: newStringDecoder(),
	}
}

func (d *structDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	buflen := len(buf)
	cursor = skipWhiteSpace(buf, cursor)
	if buflen < 2 {
		return 0, errors.New("unexpected error {}")
	}
	if buf[cursor] != '{' {
		return 0, errors.New("unexpected error {")
	}
	cursor++
	for ; cursor < buflen; cursor++ {
		key, c, err := d.keyDecoder.decodeByte(buf, cursor)
		if err != nil {
			return 0, err
		}
		cursor = c
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] != ':' {
			return 0, errors.New("unexpected error invalid delimiter for object")
		}
		cursor++
		if cursor >= buflen {
			return 0, errors.New("unexpected error missing value")
		}
		k := *(*string)(unsafe.Pointer(&key))
		field, exists := d.fieldMap[k]
		if exists {
			c, err := field.dec.decode(buf, cursor, p+field.offset)
			if err != nil {
				return 0, err
			}
			cursor = c
		} else {
			c, err := skipValue(buf, cursor)
			if err != nil {
				return 0, err
			}
			cursor = c
		}
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] == '}' {
			cursor++
			return cursor, nil
		}
		if buf[cursor] != ',' {
			return 0, errors.New("unexpected error ,")
		}
	}
	return cursor, nil
}

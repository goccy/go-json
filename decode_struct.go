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

func (d *structDecoder) skipValue(buf []byte, cursor int) (int, error) {
	cursor = skipWhiteSpace(buf, cursor)
	braceCount := 0
	bracketCount := 0
	buflen := len(buf)
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case '{':
			braceCount++
		case '[':
			bracketCount++
		case '}':
			braceCount--
			if braceCount == -1 && bracketCount == 0 {
				return cursor, nil
			}
		case ']':
			bracketCount--
		case ',':
			if bracketCount == 0 && braceCount == 0 {
				return cursor, nil
			}
		case '"':
			cursor++
			for ; cursor < buflen; cursor++ {
				switch buf[cursor] {
				case '\\':
					cursor++
				case '"':
					if bracketCount == 0 && braceCount == 0 {
						return cursor + 1, nil
					}
					goto QUOTE_END
				}
			}
		QUOTE_END:
		}
	}
	return cursor, errors.New("unexpected error value")
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
			c, err := d.skipValue(buf, cursor)
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

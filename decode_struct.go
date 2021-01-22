package json

import (
	"fmt"
	"unsafe"
)

type structFieldSet struct {
	dec         decoder
	offset      uintptr
	isTaggedKey bool
}

type structDecoder struct {
	fieldMap   map[string]*structFieldSet
	keyDecoder *stringDecoder
	structName string
	fieldName  string
}

func newStructDecoder(structName, fieldName string, fieldMap map[string]*structFieldSet) *structDecoder {
	return &structDecoder{
		fieldMap:   fieldMap,
		keyDecoder: newStringDecoder(structName, fieldName),
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *structDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	s.skipWhiteSpace()
	switch s.char() {
	case 'n':
		if err := nullBytes(s); err != nil {
			return err
		}
		return nil
	case nul:
		s.read()
	default:
		if s.char() != '{' {
			return errNotAtBeginningOfValue(s.totalOffset())
		}
	}
	s.cursor++
	if s.char() == '}' {
		s.cursor++
		return nil
	}
	for {
		s.reset()
		key, err := d.keyDecoder.decodeStreamByte(s)
		if err != nil {
			return err
		}
		s.skipWhiteSpace()
		if s.char() == nul {
			s.read()
		}
		if s.char() != ':' {
			return errExpected("colon after object key", s.totalOffset())
		}
		s.cursor++
		if s.char() == nul {
			if !s.read() {
				return errExpected("object value after colon", s.totalOffset())
			}
		}
		k := *(*string)(unsafe.Pointer(&key))
		field, exists := d.fieldMap[k]
		if exists {
			if err := field.dec.decodeStream(s, unsafe.Pointer(uintptr(p)+field.offset)); err != nil {
				return err
			}
		} else if s.disallowUnknownFields {
			return fmt.Errorf("json: unknown field %q", k)
		} else {
			if err := s.skipValue(); err != nil {
				return err
			}
		}
		s.skipWhiteSpace()
		if s.char() == nul {
			s.read()
		}
		c := s.char()
		if c == '}' {
			s.cursor++
			return nil
		}
		if c != ',' {
			return errExpected("comma after object element", s.totalOffset())
		}
		s.cursor++
	}
	return nil
}

func (d *structDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	buflen := int64(len(buf))
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 'n':
		buflen := int64(len(buf))
		if cursor+3 >= buflen {
			return 0, errUnexpectedEndOfJSON("null", cursor)
		}
		if buf[cursor+1] != 'u' {
			return 0, errInvalidCharacter(buf[cursor+1], "null", cursor)
		}
		if buf[cursor+2] != 'l' {
			return 0, errInvalidCharacter(buf[cursor+2], "null", cursor)
		}
		if buf[cursor+3] != 'l' {
			return 0, errInvalidCharacter(buf[cursor+3], "null", cursor)
		}
		cursor += 4
		return cursor, nil
	case '{':
	default:
		return 0, errNotAtBeginningOfValue(cursor)
	}
	if buflen < 2 {
		return 0, errUnexpectedEndOfJSON("object", cursor)
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
			return 0, errExpected("colon after object key", cursor)
		}
		cursor++
		if cursor >= buflen {
			return 0, errExpected("object value after colon", cursor)
		}
		k := *(*string)(unsafe.Pointer(&key))
		field, exists := d.fieldMap[k]
		if exists {
			c, err := field.dec.decode(buf, cursor, unsafe.Pointer(uintptr(p)+field.offset))
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
			return 0, errExpected("comma after object element", cursor)
		}
	}
	return cursor, nil
}

package json

import (
	"reflect"
	"unsafe"
)

type interfaceDecoder struct {
	typ   *rtype
	dummy unsafe.Pointer // for escape value
}

func newInterfaceDecoder(typ *rtype) *interfaceDecoder {
	return &interfaceDecoder{typ: typ}
}

var (
	interfaceMapType = type2rtype(
		reflect.TypeOf((*map[interface{}]interface{})(nil)).Elem(),
	)
)

func (d *interfaceDecoder) decodeStream(s *stream, p uintptr) error {
	s.skipWhiteSpace()
	switch s.char() {
	case '{':
		var v map[interface{}]interface{}
		ptr := unsafe.Pointer(&v)
		d.dummy = ptr
		dec := newMapDecoder(interfaceMapType, newInterfaceDecoder(d.typ), newInterfaceDecoder(d.typ))
		if err := dec.decodeStream(s, uintptr(ptr)); err != nil {
			return err
		}
		*(*interface{})(unsafe.Pointer(p)) = v
		return nil
	case '[':
		var v []interface{}
		ptr := unsafe.Pointer(&v)
		d.dummy = ptr // escape ptr
		dec := newSliceDecoder(newInterfaceDecoder(d.typ), d.typ, d.typ.Size())
		if err := dec.decodeStream(s, uintptr(ptr)); err != nil {
			return err
		}
		*(*interface{})(unsafe.Pointer(p)) = v
		return nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return newFloatDecoder(func(p uintptr, v float64) {
			*(*interface{})(unsafe.Pointer(p)) = v
		}).decodeStream(s, p)
	case '"':
		s.progress()
		start := s.cursor
		for {
			switch s.char() {
			case '\\':
				s.progress()
			case '"':
				literal := s.buf[start:s.cursor]
				s.progress()
				*(*interface{})(unsafe.Pointer(p)) = *(*string)(unsafe.Pointer(&literal))
				return nil
			case '\000':
				return errUnexpectedEndOfJSON("string", s.totalOffset())
			}
			s.progress()
		}
		return errUnexpectedEndOfJSON("string", s.totalOffset())
	case 't':
		s.progress()
		if s.char() != 'r' {
			return errInvalidCharacter(s.char(), "bool(true)", s.totalOffset())
		}
		s.progress()
		if s.char() != 'u' {
			return errInvalidCharacter(s.char(), "bool(true)", s.totalOffset())
		}
		s.progress()
		if s.char() != 'e' {
			return errInvalidCharacter(s.char(), "bool(true)", s.totalOffset())
		}
		s.progress()
		*(*interface{})(unsafe.Pointer(p)) = true
		return nil
	case 'f':
		s.progress()
		if s.char() != 'a' {
			return errInvalidCharacter(s.char(), "bool(false)", s.totalOffset())
		}
		s.progress()
		if s.char() != 'l' {
			return errInvalidCharacter(s.char(), "bool(false)", s.totalOffset())
		}
		s.progress()
		if s.char() != 's' {
			return errInvalidCharacter(s.char(), "bool(false)", s.totalOffset())
		}
		s.progress()
		if s.char() != 'e' {
			return errInvalidCharacter(s.char(), "bool(false)", s.totalOffset())
		}
		s.progress()
		*(*interface{})(unsafe.Pointer(p)) = false
		return nil
	case 'n':
		s.progress()
		if s.char() != 'u' {
			return errInvalidCharacter(s.char(), "null", s.totalOffset())
		}
		s.progress()
		if s.char() != 'l' {
			return errInvalidCharacter(s.char(), "null", s.totalOffset())
		}
		s.progress()
		if s.char() != 'l' {
			return errInvalidCharacter(s.char(), "null", s.totalOffset())
		}
		s.progress()
		*(*interface{})(unsafe.Pointer(p)) = nil
		return nil
	}
	return errNotAtBeginningOfValue(s.totalOffset())
}

func (d *interfaceDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case '{':
		var v map[interface{}]interface{}
		ptr := unsafe.Pointer(&v)
		d.dummy = ptr
		dec := newMapDecoder(interfaceMapType, newInterfaceDecoder(d.typ), newInterfaceDecoder(d.typ))
		cursor, err := dec.decode(buf, cursor, uintptr(ptr))
		if err != nil {
			return 0, err
		}
		*(*interface{})(unsafe.Pointer(p)) = v
		return cursor, nil
	case '[':
		var v []interface{}
		ptr := unsafe.Pointer(&v)
		d.dummy = ptr // escape ptr
		dec := newSliceDecoder(newInterfaceDecoder(d.typ), d.typ, d.typ.Size())
		cursor, err := dec.decode(buf, cursor, uintptr(ptr))
		if err != nil {
			return 0, err
		}
		*(*interface{})(unsafe.Pointer(p)) = v
		return cursor, nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return newFloatDecoder(func(p uintptr, v float64) {
			*(*interface{})(unsafe.Pointer(p)) = v
		}).decode(buf, cursor, p)
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
				*(*interface{})(unsafe.Pointer(p)) = *(*string)(unsafe.Pointer(&literal))
				return cursor, nil
			case '\000':
				return 0, errUnexpectedEndOfJSON("string", cursor)
			}
			cursor++
		}
		return 0, errUnexpectedEndOfJSON("string", cursor)
	case 't':
		if cursor+3 >= int64(len(buf)) {
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
		*(*interface{})(unsafe.Pointer(p)) = true
		return cursor, nil
	case 'f':
		if cursor+4 >= int64(len(buf)) {
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
		*(*interface{})(unsafe.Pointer(p)) = false
		return cursor, nil
	case 'n':
		if cursor+3 >= int64(len(buf)) {
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
		*(*interface{})(unsafe.Pointer(p)) = nil
		return cursor, nil
	}
	return cursor, errNotAtBeginningOfValue(cursor)
}

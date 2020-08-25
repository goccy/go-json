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
	return &interfaceDecoder{
		typ: typ,
	}
}

func (d *interfaceDecoder) numDecoder(s *stream) decoder {
	if s.useNumber {
		return newNumberDecoder(func(p uintptr, v Number) {
			*(*interface{})(unsafe.Pointer(p)) = v
		})
	}
	return newFloatDecoder(func(p uintptr, v float64) {
		*(*interface{})(unsafe.Pointer(p)) = v
	})
}

var (
	interfaceMapType = type2rtype(
		reflect.TypeOf((*map[string]interface{})(nil)).Elem(),
	)
)

func (d *interfaceDecoder) decodeStream(s *stream, p uintptr) error {
	s.skipWhiteSpace()
	for {
		switch s.char() {
		case '{':
			var v map[string]interface{}
			ptr := unsafe.Pointer(&v)
			d.dummy = ptr
			if err := newMapDecoder(
				interfaceMapType,
				newStringDecoder(),
				newInterfaceDecoder(d.typ),
			).decodeStream(s, uintptr(ptr)); err != nil {
				return err
			}
			*(*interface{})(unsafe.Pointer(p)) = v
			return nil
		case '[':
			var v []interface{}
			ptr := unsafe.Pointer(&v)
			d.dummy = ptr // escape ptr
			if err := newSliceDecoder(
				newInterfaceDecoder(d.typ),
				d.typ,
				d.typ.Size(),
			).decodeStream(s, uintptr(ptr)); err != nil {
				return err
			}
			*(*interface{})(unsafe.Pointer(p)) = v
			return nil
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return d.numDecoder(s).decodeStream(s, p)
		case '"':
			s.cursor++
			start := s.cursor
			for {
				switch s.char() {
				case '\\':
					if err := decodeEscapeString(s); err != nil {
						return err
					}
				case '"':
					literal := s.buf[start:s.cursor]
					s.cursor++
					*(*interface{})(unsafe.Pointer(p)) = *(*string)(unsafe.Pointer(&literal))
					return nil
				case nul:
					if s.read() {
						continue
					}
					return errUnexpectedEndOfJSON("string", s.totalOffset())
				}
				s.cursor++
			}
			return errUnexpectedEndOfJSON("string", s.totalOffset())
		case 't':
			if err := trueBytes(s); err != nil {
				return err
			}
			*(*interface{})(unsafe.Pointer(p)) = true
			return nil
		case 'f':
			if err := falseBytes(s); err != nil {
				return err
			}
			*(*interface{})(unsafe.Pointer(p)) = false
			return nil
		case 'n':
			if err := nullBytes(s); err != nil {
				return err
			}
			*(*interface{})(unsafe.Pointer(p)) = nil
			return nil
		case nul:
			if s.read() {
				continue
			}
		}
		break
	}
	return errNotAtBeginningOfValue(s.totalOffset())
}

func (d *interfaceDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case '{':
		var v map[string]interface{}
		ptr := unsafe.Pointer(&v)
		d.dummy = ptr
		dec := newMapDecoder(
			interfaceMapType,
			newStringDecoder(),
			newInterfaceDecoder(d.typ),
		)
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
		dec := newSliceDecoder(
			newInterfaceDecoder(d.typ),
			d.typ,
			d.typ.Size(),
		)
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
			case nul:
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

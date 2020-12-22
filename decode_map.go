package json

import (
	"unsafe"
)

type mapDecoder struct {
	mapType      *rtype
	keyType      *rtype
	valueType    *rtype
	keyDecoder   decoder
	valueDecoder decoder
	structName   string
	fieldName    string
}

func newMapDecoder(mapType *rtype, keyType *rtype, keyDec decoder, valueType *rtype, valueDec decoder, structName, fieldName string) *mapDecoder {
	return &mapDecoder{
		mapType:      mapType,
		keyDecoder:   keyDec,
		keyType:      keyType,
		valueType:    valueType,
		valueDecoder: valueDec,
		structName:   structName,
		fieldName:    fieldName,
	}
}

//go:linkname makemap reflect.makemap
func makemap(*rtype, int) unsafe.Pointer

//go:linkname mapassign reflect.mapassign
//go:noescape
func mapassign(t *rtype, m unsafe.Pointer, key, val unsafe.Pointer)

func (d *mapDecoder) setKey(buf []byte, cursor int64, key interface{}) (int64, error) {
	header := (*interfaceHeader)(unsafe.Pointer(&key))
	return d.keyDecoder.decode(buf, cursor, header.ptr)
}

func (d *mapDecoder) setValue(buf []byte, cursor int64, key interface{}) (int64, error) {
	header := (*interfaceHeader)(unsafe.Pointer(&key))
	return d.valueDecoder.decode(buf, cursor, header.ptr)
}

func (d *mapDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	s.skipWhiteSpace()
	switch s.char() {
	case 'n':
		if err := nullBytes(s); err != nil {
			return err
		}
		return nil
	case '{':
	default:
		return errExpected("{ character for map value", s.totalOffset())
	}
	s.skipWhiteSpace()
	mapValue := makemap(d.mapType, 0)
	if s.buf[s.cursor+1] == '}' {
		*(*unsafe.Pointer)(p) = mapValue
		s.cursor += 2
		return nil
	}
	for {
		s.cursor++
		k := unsafe_New(d.keyType)
		if err := d.keyDecoder.decodeStream(s, k); err != nil {
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
		v := unsafe_New(d.valueType)
		if err := d.valueDecoder.decodeStream(s, v); err != nil {
			return err
		}
		mapassign(d.mapType, mapValue, k, v)
		s.skipWhiteSpace()
		if s.char() == nul {
			s.read()
		}
		if s.char() == '}' {
			**(**unsafe.Pointer)(unsafe.Pointer(&p)) = mapValue
			s.cursor++
			return nil
		}
		if s.char() != ',' {
			return errExpected("comma after object value", s.totalOffset())
		}
	}
	return nil
}

func (d *mapDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	buflen := int64(len(buf))
	if buflen < 2 {
		return 0, errExpected("{} for map", cursor)
	}
	switch buf[cursor] {
	case 'n':
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
		return 0, errExpected("{ character for map value", cursor)
	}
	cursor++
	cursor = skipWhiteSpace(buf, cursor)
	mapValue := makemap(d.mapType, 0)
	if buf[cursor] == '}' {
		**(**unsafe.Pointer)(unsafe.Pointer(&p)) = mapValue
		cursor++
		return cursor, nil
	}
	for ; cursor < buflen; cursor++ {
		var key interface{}
		keyCursor, err := d.setKey(buf, cursor, &key)
		if err != nil {
			return 0, err
		}
		cursor = keyCursor
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] != ':' {
			return 0, errExpected("colon after object key", cursor)
		}
		cursor++
		if cursor >= buflen {
			return 0, errUnexpectedEndOfJSON("map", cursor)
		}
		var value interface{}
		valueCursor, err := d.setValue(buf, cursor, &value)
		if err != nil {
			return 0, err
		}
		cursor = valueCursor
		mapassign(d.mapType, mapValue, unsafe.Pointer(&key), unsafe.Pointer(&value))
		cursor = skipWhiteSpace(buf, valueCursor)
		if buf[cursor] == '}' {
			**(**unsafe.Pointer)(unsafe.Pointer(&p)) = mapValue
			cursor++
			return cursor, nil
		}
		if buf[cursor] != ',' {
			return 0, errExpected("comma after object value", cursor)
		}
	}
	return cursor, nil
}

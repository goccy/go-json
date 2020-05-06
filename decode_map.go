package json

import (
	"errors"
	"unsafe"
)

type mapDecoder struct {
	mapType      *rtype
	keyDecoder   decoder
	valueDecoder decoder
}

func newMapDecoder(mapType *rtype, keyDec decoder, valueDec decoder) *mapDecoder {
	return &mapDecoder{
		mapType:      mapType,
		keyDecoder:   keyDec,
		valueDecoder: valueDec,
	}
}

//go:linkname makemap reflect.makemap
func makemap(*rtype, int) unsafe.Pointer

//go:linkname mapassign reflect.mapassign
//go:noescape
func mapassign(t *rtype, m unsafe.Pointer, key, val unsafe.Pointer)

func (d *mapDecoder) setKey(buf []byte, cursor int, key interface{}) (int, error) {
	header := (*interfaceHeader)(unsafe.Pointer(&key))
	return d.keyDecoder.decode(buf, cursor, uintptr(header.ptr))
}

func (d *mapDecoder) setValue(buf []byte, cursor int, key interface{}) (int, error) {
	header := (*interfaceHeader)(unsafe.Pointer(&key))
	return d.valueDecoder.decode(buf, cursor, uintptr(header.ptr))
}

func (d *mapDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	cursor = skipWhiteSpace(buf, cursor)
	buflen := len(buf)
	if buflen < 2 {
		return 0, errors.New("unexpected error {}")
	}
	if buf[cursor] != '{' {
		return 0, errors.New("unexpected error {")
	}
	cursor++
	mapValue := makemap(d.mapType, 0)
	for ; cursor < buflen; cursor++ {
		var key interface{}
		keyCursor, err := d.setKey(buf, cursor, &key)
		if err != nil {
			return 0, err
		}
		cursor = keyCursor
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] != ':' {
			return 0, errors.New("unexpected error invalid delimiter for object")
		}
		cursor++
		if cursor >= buflen {
			return 0, errors.New("unexpected error missing value")
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
			*(*unsafe.Pointer)(unsafe.Pointer(p)) = mapValue
			return cursor, nil
		}
		if buf[cursor] != ',' {
			return 0, errors.New("unexpected error ,")
		}
	}
	return cursor, nil
}

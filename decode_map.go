package json

import (
	"errors"
	"fmt"
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

func (d *mapDecoder) setKey(ctx *context, key interface{}) error {
	header := (*interfaceHeader)(unsafe.Pointer(&key))
	if err := d.keyDecoder.decode(ctx, uintptr(header.ptr)); err != nil {
		return err
	}
	//fmt.Println("key = ", *(*string)(header.ptr))
	//fmt.Println("Key = ", key.(*string))
	//fmt.Println("key = ", *(*string)(unsafe.Pointer(key)))
	return nil
}

func (d *mapDecoder) setValue(ctx *context, key interface{}) error {
	header := (*interfaceHeader)(unsafe.Pointer(&key))
	return d.valueDecoder.decode(ctx, uintptr(header.ptr))
}

func (d *mapDecoder) decode(ctx *context, p uintptr) error {
	ctx.skipWhiteSpace()
	buf := ctx.buf
	buflen := ctx.buflen
	cursor := ctx.cursor
	if buflen < 2 {
		return errors.New("unexpected error {}")
	}
	if buf[cursor] != '{' {
		return errors.New("unexpected error {")
	}
	cursor++
	mapValue := makemap(d.mapType, 0)
	fmt.Println("mapValue = ", mapValue)
	for ; cursor < buflen; cursor++ {
		ctx.cursor = cursor
		var key interface{}
		if err := d.setKey(ctx, &key); err != nil {
			return err
		}
		cursor = ctx.skipWhiteSpace()
		if buf[cursor] != ':' {
			return errors.New("unexpected error invalid delimiter for object")
		}
		cursor++
		if cursor >= buflen {
			return errors.New("unexpected error missing value")
		}
		ctx.cursor = cursor
		var value interface{}
		if err := d.setValue(ctx, &value); err != nil {
			return err
		}
		mapassign(d.mapType, mapValue, unsafe.Pointer(&key), unsafe.Pointer(&value))
		cursor = ctx.skipWhiteSpace()
		if buf[cursor] == '}' {
			*(*unsafe.Pointer)(unsafe.Pointer(p)) = mapValue
			return nil
		}
		if buf[cursor] != ',' {
			return errors.New("unexpected error ,")
		}
	}
	return nil
}

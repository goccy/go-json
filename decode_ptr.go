package json

import (
	"reflect"
	"unsafe"
)

type ptrDecoder struct {
	dec decoder
	typ reflect.Type
}

func newPtrDecoder(dec decoder, typ reflect.Type) *ptrDecoder {
	return &ptrDecoder{dec: dec, typ: typ}
}

func (d *ptrDecoder) decode(ctx *context, p uintptr) error {
	newptr := uintptr(reflect.New(d.typ).Pointer())
	if err := d.dec.decode(ctx, newptr); err != nil {
		return err
	}
	*(*uintptr)(unsafe.Pointer(p)) = newptr
	return nil
}

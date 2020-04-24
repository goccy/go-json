package json

import (
	"unsafe"
)

type ptrDecoder struct {
	dec decoder
	typ *rtype
}

func newPtrDecoder(dec decoder, typ *rtype) *ptrDecoder {
	return &ptrDecoder{dec: dec, typ: typ}
}

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(*rtype) uintptr

func (d *ptrDecoder) decode(ctx *context, p uintptr) error {
	newptr := unsafe_New(d.typ)
	if err := d.dec.decode(ctx, newptr); err != nil {
		return err
	}
	*(*uintptr)(unsafe.Pointer(p)) = newptr
	return nil
}

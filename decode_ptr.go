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

func (d *ptrDecoder) decodeStream(s *stream, p uintptr) error {
	newptr := unsafe_New(d.typ)
	if err := d.dec.decodeStream(s, newptr); err != nil {
		return err
	}
	*(*uintptr)(unsafe.Pointer(p)) = newptr
	return nil
}

func (d *ptrDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	newptr := unsafe_New(d.typ)
	c, err := d.dec.decode(buf, cursor, newptr)
	if err != nil {
		return 0, err
	}
	cursor = c
	*(*uintptr)(unsafe.Pointer(p)) = newptr
	return cursor, nil
}

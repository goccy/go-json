package json

import (
	"unsafe"
)

type ptrDecoder struct {
	dec        decoder
	typ        *rtype
	structName string
	fieldName  string
}

func newPtrDecoder(dec decoder, typ *rtype, structName, fieldName string) *ptrDecoder {
	return &ptrDecoder{
		dec:        dec,
		typ:        typ,
		structName: structName,
		fieldName:  fieldName,
	}
}

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(*rtype) unsafe.Pointer

func (d *ptrDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	newptr := unsafe_New(d.typ)
	*(*unsafe.Pointer)(p) = newptr
	if err := d.dec.decodeStream(s, newptr); err != nil {
		return err
	}
	return nil
}

func (d *ptrDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	newptr := unsafe_New(d.typ)
	*(*unsafe.Pointer)(p) = newptr
	c, err := d.dec.decode(buf, cursor, newptr)
	if err != nil {
		return 0, err
	}
	cursor = c
	return cursor, nil
}

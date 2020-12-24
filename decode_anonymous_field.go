package json

import (
	"unsafe"
)

type anonymousFieldDecoder struct {
	structType *rtype
	offset     uintptr
	dec        decoder
}

func newAnonymousFieldDecoder(structType *rtype, offset uintptr, dec decoder) *anonymousFieldDecoder {
	return &anonymousFieldDecoder{
		structType: structType,
		offset:     offset,
		dec:        dec,
	}
}

func (d *anonymousFieldDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	if *(*unsafe.Pointer)(p) == nil {
		*(*unsafe.Pointer)(p) = unsafe_New(d.structType)
	}
	p = *(*unsafe.Pointer)(p)
	return d.dec.decodeStream(s, unsafe.Pointer(uintptr(p)+d.offset))
}

func (d *anonymousFieldDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	if *(*unsafe.Pointer)(p) == nil {
		*(*unsafe.Pointer)(p) = unsafe_New(d.structType)
	}
	p = *(*unsafe.Pointer)(p)
	return d.dec.decode(buf, cursor, unsafe.Pointer(uintptr(p)+d.offset))
}

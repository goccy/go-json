package json

import (
	"sync"
	"unsafe"
)

type ptrDecoder struct {
	dec      decoder
	typ      *rtype
	mu       sync.Mutex
	keepRefs []unsafe.Pointer
}

func newPtrDecoder(dec decoder, typ *rtype) *ptrDecoder {
	return &ptrDecoder{dec: dec, typ: typ}
}

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(*rtype) unsafe.Pointer

func (d *ptrDecoder) decodeStream(s *stream, p uintptr) error {
	newptr := unsafe_New(d.typ)
	if err := d.dec.decodeStream(s, uintptr(newptr)); err != nil {
		return err
	}
	**(**unsafe.Pointer)(unsafe.Pointer(&p)) = newptr
	return nil
}

func (d *ptrDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	d.mu.Lock()
	newptr := unsafe_New(d.typ)
	d.keepRefs = append(d.keepRefs, newptr)
	**(**unsafe.Pointer)(unsafe.Pointer(&p)) = newptr
	d.mu.Unlock()
	c, err := d.dec.decode(buf, cursor, uintptr(newptr))
	if err != nil {
		return 0, err
	}
	cursor = c
	return cursor, nil
}

package json

import (
	"reflect"
	"sync"
	"unsafe"
)

type sliceDecoder struct {
	elemType     *rtype
	valueDecoder decoder
	size         uintptr
	arrayPool    sync.Pool
}

func newSliceDecoder(dec decoder, elemType *rtype, size uintptr) *sliceDecoder {
	return &sliceDecoder{
		valueDecoder: dec,
		elemType:     elemType,
		size:         size,
		arrayPool: sync.Pool{
			New: func() interface{} {
				cap := 2
				return &reflect.SliceHeader{
					Data: uintptr(newArray(elemType, cap)),
					Len:  0,
					Cap:  cap,
				}
			},
		},
	}
}

func (d *sliceDecoder) newSlice() *reflect.SliceHeader {
	slice := d.arrayPool.Get().(*reflect.SliceHeader)
	slice.Len = 0
	return slice
}

func (d *sliceDecoder) releaseSlice(p *reflect.SliceHeader) {
	d.arrayPool.Put(p)
}

//go:linkname copySlice reflect.typedslicecopy
func copySlice(elemType *rtype, dst, src reflect.SliceHeader) int

//go:linkname newArray reflect.unsafe_NewArray
func newArray(*rtype, int) unsafe.Pointer

func (d *sliceDecoder) decodeStream(s *stream, p uintptr) error {
	for {
		switch s.char() {
		case ' ', '\n', '\t', '\r':
			s.progress()
			continue
		case '[':
			idx := 0
			slice := d.newSlice()
			cap := slice.Cap
			data := slice.Data
			for s.progress() {
				if cap <= idx {
					src := reflect.SliceHeader{Data: data, Len: idx, Cap: cap}
					cap *= 2
					data = uintptr(newArray(d.elemType, cap))
					dst := reflect.SliceHeader{Data: data, Len: idx, Cap: cap}
					copySlice(d.elemType, dst, src)
				}
				if err := d.valueDecoder.decodeStream(s, data+uintptr(idx)*d.size); err != nil {
					return err
				}
				s.skipWhiteSpace()
				switch s.char() {
				case ']':
					slice.Cap = cap
					slice.Len = idx + 1
					slice.Data = data
					dstCap := idx + 1
					dst := reflect.SliceHeader{
						Data: uintptr(newArray(d.elemType, dstCap)),
						Len:  idx + 1,
						Cap:  dstCap,
					}
					copySlice(d.elemType, dst, *slice)
					*(*reflect.SliceHeader)(unsafe.Pointer(p)) = dst
					d.releaseSlice(slice)
					s.progress()
					return nil
				case ',':
					idx++
					continue
				default:
					slice.Cap = cap
					slice.Data = data
					d.releaseSlice(slice)
					return errInvalidCharacter(s.char(), "slice", s.totalOffset())
				}
			}
		}
	}
	return errUnexpectedEndOfJSON("slice", s.totalOffset())
}

func (d *sliceDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	buflen := int64(len(buf))
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case '[':
			idx := 0
			slice := d.newSlice()
			cap := slice.Cap
			data := slice.Data
			for {
				cursor++
				if cap <= idx {
					src := reflect.SliceHeader{Data: data, Len: idx, Cap: cap}
					cap *= 2
					data = uintptr(newArray(d.elemType, cap))
					dst := reflect.SliceHeader{Data: data, Len: idx, Cap: cap}
					copySlice(d.elemType, dst, src)
				}
				c, err := d.valueDecoder.decode(buf, cursor, data+uintptr(idx)*d.size)
				if err != nil {
					return 0, err
				}
				cursor = c
				cursor = skipWhiteSpace(buf, cursor)
				switch buf[cursor] {
				case ']':
					slice.Cap = cap
					slice.Len = idx + 1
					slice.Data = data
					dstCap := idx + 1
					dst := reflect.SliceHeader{
						Data: uintptr(newArray(d.elemType, dstCap)),
						Len:  idx + 1,
						Cap:  dstCap,
					}
					copySlice(d.elemType, dst, *slice)
					*(*reflect.SliceHeader)(unsafe.Pointer(p)) = dst
					d.releaseSlice(slice)
					cursor++
					return cursor, nil
				case ',':
					idx++
					continue
				default:
					slice.Cap = cap
					slice.Data = data
					d.releaseSlice(slice)
					return 0, errInvalidCharacter(buf[cursor], "slice", cursor)
				}
			}
		}
	}
	return 0, errUnexpectedEndOfJSON("slice", cursor)
}

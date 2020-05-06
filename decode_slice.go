package json

import (
	"errors"
	"reflect"
	"unsafe"
)

type sliceDecoder struct {
	elemType     *rtype
	valueDecoder decoder
	size         uintptr
}

func newSliceDecoder(dec decoder, elemType *rtype, size uintptr) *sliceDecoder {
	return &sliceDecoder{
		valueDecoder: dec,
		elemType:     elemType,
		size:         size,
	}
}

//go:linkname copySlice reflect.typedslicecopy
func copySlice(elemType *rtype, dst, src reflect.SliceHeader) int

//go:linkname newArray reflect.unsafe_NewArray
func newArray(*rtype, int) unsafe.Pointer

func (d *sliceDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	buflen := len(buf)
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case '[':
			idx := 0
			cap := 2
			data := uintptr(newArray(d.elemType, cap))
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
					*(*reflect.SliceHeader)(unsafe.Pointer(p)) = reflect.SliceHeader{
						Data: data,
						Len:  idx + 1,
						Cap:  cap,
					}
					cursor++
					return cursor, nil
				case ',':
					idx++
					continue
				default:
					return 0, errors.New("syntax error slice")
				}
			}
		}
	}
	return 0, errors.New("unexpected error slice")
}

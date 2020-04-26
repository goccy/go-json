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

func (d *sliceDecoder) decode(ctx *context, p uintptr) error {
	buf := ctx.buf
	buflen := ctx.buflen
	cursor := ctx.cursor
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case '[':
			idx := 0
			cap := 2
			data := uintptr(newArray(d.elemType, cap))
			for {
				ctx.cursor = cursor + 1
				if cap <= idx {
					src := reflect.SliceHeader{Data: data, Len: idx, Cap: cap}
					cap *= 2
					data = uintptr(newArray(d.elemType, cap))
					dst := reflect.SliceHeader{Data: data, Len: idx, Cap: cap}
					copySlice(d.elemType, dst, src)
				}
				if err := d.valueDecoder.decode(ctx, data+uintptr(idx)*d.size); err != nil {
					return err
				}
				cursor = ctx.skipWhiteSpace()
				switch buf[cursor] {
				case ']':
					*(*reflect.SliceHeader)(unsafe.Pointer(p)) = reflect.SliceHeader{
						Data: data,
						Len:  idx + 1,
						Cap:  cap,
					}
					ctx.cursor++
					return nil
				case ',':
					idx++
					continue
				default:
					return errors.New("syntax error slice")
				}
			}
		}
	}
	return errors.New("unexpected error slice")
}

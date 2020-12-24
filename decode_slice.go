package json

import (
	"sync"
	"unsafe"
)

type sliceDecoder struct {
	elemType     *rtype
	valueDecoder decoder
	size         uintptr
	arrayPool    sync.Pool
	structName   string
	fieldName    string
}

// If use reflect.SliceHeader, data type is uintptr.
// In this case, Go compiler cannot trace reference created by newArray().
// So, define using unsafe.Pointer as data type
type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

func newSliceDecoder(dec decoder, elemType *rtype, size uintptr, structName, fieldName string) *sliceDecoder {
	return &sliceDecoder{
		valueDecoder: dec,
		elemType:     elemType,
		size:         size,
		arrayPool: sync.Pool{
			New: func() interface{} {
				cap := 2
				return &sliceHeader{
					data: newArray(elemType, cap),
					len:  0,
					cap:  cap,
				}
			},
		},
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *sliceDecoder) newSlice() *sliceHeader {
	slice := d.arrayPool.Get().(*sliceHeader)
	slice.len = 0
	return slice
}

func (d *sliceDecoder) releaseSlice(p *sliceHeader) {
	d.arrayPool.Put(p)
}

//go:linkname copySlice reflect.typedslicecopy
func copySlice(elemType *rtype, dst, src sliceHeader) int

//go:linkname newArray reflect.unsafe_NewArray
func newArray(*rtype, int) unsafe.Pointer

func (d *sliceDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	for {
		switch s.char() {
		case ' ', '\n', '\t', '\r':
			s.cursor++
			continue
		case 'n':
			if err := nullBytes(s); err != nil {
				return err
			}
			return nil
		case '[':
			s.cursor++
			s.skipWhiteSpace()
			if s.char() == ']' {
				*(*sliceHeader)(p) = sliceHeader{
					data: newArray(d.elemType, 0),
					len:  0,
					cap:  0,
				}
				s.cursor++
				return nil
			}
			idx := 0
			slice := d.newSlice()
			cap := slice.cap
			data := slice.data
			for {
				if cap <= idx {
					src := sliceHeader{data: data, len: idx, cap: cap}
					cap *= 2
					data = newArray(d.elemType, cap)
					dst := sliceHeader{data: data, len: idx, cap: cap}
					copySlice(d.elemType, dst, src)
				}
				if err := d.valueDecoder.decodeStream(s, unsafe.Pointer(uintptr(data)+uintptr(idx)*d.size)); err != nil {
					return err
				}
				s.skipWhiteSpace()
			RETRY:
				switch s.char() {
				case ']':
					slice.cap = cap
					slice.len = idx + 1
					slice.data = data
					dstCap := idx + 1
					dst := sliceHeader{
						data: newArray(d.elemType, dstCap),
						len:  idx + 1,
						cap:  dstCap,
					}
					copySlice(d.elemType, dst, sliceHeader{
						data: slice.data,
						len:  slice.len,
						cap:  slice.cap,
					})
					*(*sliceHeader)(p) = dst
					d.releaseSlice(slice)
					s.cursor++
					return nil
				case ',':
					idx++
				case nul:
					if s.read() {
						goto RETRY
					}
					slice.cap = cap
					slice.data = data
					d.releaseSlice(slice)
					goto ERROR
				default:
					slice.cap = cap
					slice.data = data
					d.releaseSlice(slice)
					goto ERROR
				}
				s.cursor++
			}
		case nul:
			if s.read() {
				continue
			}
			goto ERROR
		}
	}
ERROR:
	return errUnexpectedEndOfJSON("slice", s.totalOffset())
}

func (d *sliceDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	buflen := int64(len(buf))
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case 'n':
			buflen := int64(len(buf))
			if cursor+3 >= buflen {
				return 0, errUnexpectedEndOfJSON("null", cursor)
			}
			if buf[cursor+1] != 'u' {
				return 0, errInvalidCharacter(buf[cursor+1], "null", cursor)
			}
			if buf[cursor+2] != 'l' {
				return 0, errInvalidCharacter(buf[cursor+2], "null", cursor)
			}
			if buf[cursor+3] != 'l' {
				return 0, errInvalidCharacter(buf[cursor+3], "null", cursor)
			}
			cursor += 4
			return cursor, nil
		case '[':
			cursor++
			cursor = skipWhiteSpace(buf, cursor)
			if buf[cursor] == ']' {
				**(**sliceHeader)(unsafe.Pointer(&p)) = sliceHeader{
					data: newArray(d.elemType, 0),
					len:  0,
					cap:  0,
				}
				cursor++
				return cursor, nil
			}
			idx := 0
			slice := d.newSlice()
			cap := slice.cap
			data := slice.data
			for {
				if cap <= idx {
					src := sliceHeader{data: data, len: idx, cap: cap}
					cap *= 2
					data = newArray(d.elemType, cap)
					dst := sliceHeader{data: data, len: idx, cap: cap}
					copySlice(d.elemType, dst, src)
				}
				c, err := d.valueDecoder.decode(buf, cursor, unsafe.Pointer(uintptr(data)+uintptr(idx)*d.size))
				if err != nil {
					return 0, err
				}
				cursor = c
				cursor = skipWhiteSpace(buf, cursor)
				switch buf[cursor] {
				case ']':
					slice.cap = cap
					slice.len = idx + 1
					slice.data = data
					dstCap := idx + 1
					dst := sliceHeader{
						data: newArray(d.elemType, dstCap),
						len:  idx + 1,
						cap:  dstCap,
					}
					copySlice(d.elemType, dst, sliceHeader{
						data: slice.data,
						len:  slice.len,
						cap:  slice.cap,
					})
					**(**sliceHeader)(unsafe.Pointer(&p)) = dst
					d.releaseSlice(slice)
					cursor++
					return cursor, nil
				case ',':
					idx++
				default:
					slice.cap = cap
					slice.data = data
					d.releaseSlice(slice)
					return 0, errInvalidCharacter(buf[cursor], "slice", cursor)
				}
				cursor++
			}
		}
	}
	return 0, errUnexpectedEndOfJSON("slice", cursor)
}

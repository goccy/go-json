package decoder

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

type sliceDecoder struct {
	elemType reflect.Type
	// slicePtrType is the type descriptor of the pointer to the slice.
	slicePtrType unsafe.Pointer
	valueDecoder Decoder
	size         uintptr
	// bufPool holds the buffers into which the elements are decoded before the length of the slice is known.
	// A buffer in the pool has only zero values, so that an element is decoded into a zero value.
	bufPool    sync.Pool
	structName string
	fieldName  string
}

// If use reflect.SliceHeader, data type is uintptr.
// In this case, Go compiler cannot trace reference created by newArray().
// So, define using unsafe.Pointer as data type
type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

// sliceBuf is a buffer of the elements of a slice, allocated and grown by reflect.Value.Grow,
// so that its memory has the type of the elements.
type sliceBuf struct {
	hdr sliceHeader
}

const (
	// defaultSliceCapacity is the capacity of a new buffer of the elements.
	defaultSliceCapacity = 4
)

func newSliceDecoder(dec Decoder, elemType reflect.Type, size uintptr, structName, fieldName string) *sliceDecoder {
	return &sliceDecoder{
		valueDecoder: dec,
		elemType:     elemType,
		slicePtrType: ptrTypeOf(reflect.SliceOf(elemType)),
		size:         size,
		bufPool: sync.Pool{
			New: func() any {
				return &sliceBuf{}
			},
		},
		structName: structName,
		fieldName:  fieldName,
	}
}

// sliceValue returns the reflect.Value of the slice whose header is at p.
func (d *sliceDecoder) sliceValue(p *sliceHeader) reflect.Value {
	return valueAt(d.slicePtrType, unsafe.Pointer(p))
}

// grow makes the capacity of the buffer n or more, keeping its first length elements.
// The elements after them are zero values.
func (d *sliceDecoder) grow(buf *sliceBuf, length, n int) {
	if n <= buf.hdr.cap {
		return
	}
	buf.hdr.len = length
	if n < defaultSliceCapacity {
		n = defaultSliceCapacity
	}
	// The capacity grows as the one of append does: the elements are copied a few times at most.
	d.sliceValue(&buf.hdr).Grow(n - length)
}

// takeBuf returns a buffer which has a copy of the elements of the slice at dst,
// into which the elements are decoded: an element of the slice is decoded into its existing value.
func (d *sliceDecoder) takeBuf(dst *sliceHeader) *sliceBuf {
	buf := d.bufPool.Get().(*sliceBuf)
	if dst.len > 0 {
		d.grow(buf, 0, dst.len)
		buf.hdr.len = dst.len
		reflect.Copy(d.sliceValue(&buf.hdr), d.sliceValue(dst))
	}
	return buf
}

// releaseBuf clears the first n elements of the buffer, which are all it may have used,
// and puts it back to the pool: it keeps nothing the decoded value refers to.
func (d *sliceDecoder) releaseBuf(buf *sliceBuf, n int) {
	if n > 0 {
		buf.hdr.len = n
		d.sliceValue(&buf.hdr).Clear()
	}
	buf.hdr.len = 0
	d.bufPool.Put(buf)
}

// store copies the n elements of the buffer to the slice at dst: into its array when it has room
// for them, or else into a new array of their length.
func (d *sliceDecoder) store(dst *sliceHeader, buf *sliceBuf, n int) {
	if dst.cap < n {
		*dst = sliceHeader{}
		d.sliceValue(dst).Grow(n)
	}
	dst.len = n
	buf.hdr.len = n
	reflect.Copy(d.sliceValue(dst), d.sliceValue(&buf.hdr))
}

func (d *sliceDecoder) errNumber(offset int64) *errors.UnmarshalTypeError {
	return &errors.UnmarshalTypeError{
		Value:  "number",
		Type:   reflect.SliceOf(d.elemType),
		Struct: d.structName,
		Field:  d.fieldName,
		Offset: offset,
	}
}

func (d *sliceDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	for {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			cursor++
			continue
		case 'n':
			if err := validateNull(buf, cursor); err != nil {
				return 0, err
			}
			cursor += 4
			*(*sliceHeader)(p) = sliceHeader{}
			return cursor, nil
		case '[':
			cursor++
			cursor = skipWhiteSpace(buf, cursor)
			dst := (*sliceHeader)(p)
			if buf[cursor] == ']' {
				if dst.data == nil {
					dst.data = unsafe.Pointer(&zeroBase)
				} else {
					dst.len = 0
				}
				cursor++
				return cursor, nil
			}
			elems := d.takeBuf(dst)
			srcLen := dst.len
			idx := 0
			for {
				d.grow(elems, idx, idx+1)
				ep := unsafe.Add(elems.hdr.data, uintptr(idx)*d.size)
				c, err := d.valueDecoder.Decode(ctx, cursor, depth, ep)
				if err != nil {
					d.releaseBuf(elems, max(idx+1, srcLen))
					return 0, err
				}
				cursor = skipWhiteSpace(buf, c)
				switch buf[cursor] {
				case ']':
					d.store(dst, elems, idx+1)
					d.releaseBuf(elems, max(idx+1, srcLen))
					cursor++
					return cursor, nil
				case ',':
					idx++
				default:
					d.releaseBuf(elems, max(idx+1, srcLen))
					return 0, errors.ErrInvalidCharacter(buf[cursor], "slice", cursor)
				}
				cursor++
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return 0, d.errNumber(cursor)
		default:
			return 0, errors.ErrUnexpectedEndOfJSON("slice", cursor)
		}
	}
}

func (d *sliceDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return nil, 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	ret := [][]byte{}
	for {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			cursor++
			continue
		case 'n':
			if err := validateNull(buf, cursor); err != nil {
				return nil, 0, err
			}
			cursor += 4
			return [][]byte{nullbytes}, cursor, nil
		case '[':
			cursor++
			cursor = skipWhiteSpace(buf, cursor)
			if buf[cursor] == ']' {
				cursor++
				return ret, cursor, nil
			}
			idx := 0
			for {
				child, found, err := ctx.Option.Path.node.Index(idx)
				if err != nil {
					return nil, 0, err
				}
				if found {
					if child != nil {
						oldPath := ctx.Option.Path.node
						ctx.Option.Path.node = child
						paths, c, err := d.valueDecoder.DecodePath(ctx, cursor, depth)
						if err != nil {
							return nil, 0, err
						}
						ctx.Option.Path.node = oldPath
						ret = append(ret, paths...)
						cursor = c
					} else {
						start := cursor
						end, err := skipValue(buf, cursor, depth)
						if err != nil {
							return nil, 0, err
						}
						ret = append(ret, buf[start:end])
						cursor = end
					}
				} else {
					c, err := skipValue(buf, cursor, depth)
					if err != nil {
						return nil, 0, err
					}
					cursor = c
				}
				cursor = skipWhiteSpace(buf, cursor)
				switch buf[cursor] {
				case ']':
					cursor++
					return ret, cursor, nil
				case ',':
					idx++
				default:
					return nil, 0, errors.ErrInvalidCharacter(buf[cursor], "slice", cursor)
				}
				cursor++
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return nil, 0, d.errNumber(cursor)
		default:
			return nil, 0, errors.ErrUnexpectedEndOfJSON("slice", cursor)
		}
	}
}

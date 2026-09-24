package decoder

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

type arrayDecoder struct {
	elemType     reflect.Type
	size         uintptr
	valueDecoder Decoder
	alen         int
	structName   string
	fieldName    string
	zeroValue    unsafe.Pointer
}

func newArrayDecoder(dec Decoder, elemType reflect.Type, alen int, structName, fieldName string) *arrayDecoder {
	// workaround to avoid checkptr errors. cannot use `*(*unsafe.Pointer)(unsafe_New(elemType))` directly.
	zeroValuePtr := unsafe_New(runtime.TypePtr(elemType))
	zeroValue := **(**unsafe.Pointer)(unsafe.Pointer(&zeroValuePtr))
	return &arrayDecoder{
		valueDecoder: dec,
		elemType:     elemType,
		size:         elemType.Size(),
		alen:         alen,
		structName:   structName,
		fieldName:    fieldName,
		zeroValue:    zeroValue,
	}
}

func (d *arrayDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
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
			return cursor, nil
		case '[':
			idx := 0
			cursor++
			cursor = skipWhiteSpace(buf, cursor)
			if buf[cursor] == ']' {
				for idx < d.alen {
					*(*unsafe.Pointer)(unsafe.Add(p, uintptr(idx)*d.size)) = d.zeroValue
					idx++
				}
				cursor++
				return cursor, nil
			}
			for {
				if idx < d.alen {
					c, err := d.valueDecoder.Decode(ctx, cursor, depth, unsafe.Add(p, uintptr(idx)*d.size))
					if err != nil {
						return 0, err
					}
					cursor = c
				} else {
					c, err := skipValue(buf, cursor, depth)
					if err != nil {
						return 0, err
					}
					cursor = c
				}
				idx++
				cursor = skipWhiteSpace(buf, cursor)
				switch buf[cursor] {
				case ']':
					for idx < d.alen {
						*(*unsafe.Pointer)(unsafe.Add(p, uintptr(idx)*d.size)) = d.zeroValue
						idx++
					}
					cursor++
					return cursor, nil
				case ',':
					cursor++
					continue
				default:
					return 0, errors.ErrInvalidCharacter(buf[cursor], "array", cursor)
				}
			}
		default:
			return 0, errors.ErrUnexpectedEndOfJSON("array", cursor)
		}
	}
}

func (d *arrayDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: array decoder does not support decode path")
}

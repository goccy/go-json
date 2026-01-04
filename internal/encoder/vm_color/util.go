package vm_color

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/runtime"
)

const uintptrSize = 4 << (^uintptr(0) >> 63)

var (
	errUnsupportedValue = encoder.ErrUnsupportedValue
	errUnsupportedFloat = encoder.ErrUnsupportedFloat
	mapiterinit         = encoder.MapIterInit
	mapiterkey          = encoder.MapIterKey
	mapitervalue        = encoder.MapIterValue
	mapiternext         = encoder.MapIterNext
	maplen              = encoder.MapLen
)

type emptyInterface struct {
	typ *runtime.Type
	ptr unsafe.Pointer
}

type nonEmptyInterface struct {
	itab *struct {
		ityp *runtime.Type // static interface type
		typ  *runtime.Type // dynamic concrete type
		// unused fields...
	}
	ptr unsafe.Pointer
}

func errUnimplementedOp(op encoder.OpType) error {
	return fmt.Errorf("encoder: opcode %s has not been implemented", op)
}

func load(base unsafe.Pointer, idx uint32) unsafe.Pointer {
	addr := unsafe.Add(base, idx)
	return *(*unsafe.Pointer)(addr)
}

func store(base unsafe.Pointer, idx uint32, p unsafe.Pointer) {
	addr := unsafe.Add(base, idx)
	*(*unsafe.Pointer)(addr) = p
}

func loadNPtr(base unsafe.Pointer, idx uint32, ptrNum uint8) unsafe.Pointer {
	addr := unsafe.Add(base, idx)
	p := *(*unsafe.Pointer)(addr)
	for i := uint8(0); i < ptrNum; i++ {
		if p == nil {
			return nil
		}
		p = ptrToPtr(p)
	}
	return p
}

func ptrToUint64(p unsafe.Pointer, bitSize uint8) uint64 {
	switch bitSize {
	case 8:
		return uint64(*(*uint8)(p))
	case 16:
		return uint64(*(*uint16)(p))
	case 32:
		return uint64(*(*uint32)(p))
	case 64:
		return *(*uint64)(p)
	}
	return 0
}

func ptrToInterface(code *encoder.Opcode, p unsafe.Pointer) interface{} {
	return *(*interface{})(unsafe.Pointer(&emptyInterface{
		typ: (*runtime.Type)(((*emptyInterface)(unsafe.Pointer(&code.Type))).ptr),
		ptr: p,
	}))
}

// unsafe.Pointer versions of ptr functions
func ptrToFloat32(p unsafe.Pointer) float32            { return *(*float32)(p) }
func ptrToFloat64(p unsafe.Pointer) float64            { return *(*float64)(p) }
func ptrToBool(p unsafe.Pointer) bool                  { return *(*bool)(p) }
func ptrToBytes(p unsafe.Pointer) []byte               { return *(*[]byte)(p) }
func ptrToNumber(p unsafe.Pointer) json.Number         { return *(*json.Number)(p) }
func ptrToString(p unsafe.Pointer) string              { return *(*string)(p) }
func ptrToSlice(p unsafe.Pointer) *runtime.SliceHeader { return (*runtime.SliceHeader)(p) }

func ptrToPtr(p unsafe.Pointer) unsafe.Pointer {
	if p == nil {
		return nil
	}
	return *(*unsafe.Pointer)(p)
}

func ptrToNPtr(p unsafe.Pointer, ptrNum uint8) unsafe.Pointer {
	for i := uint8(0); i < ptrNum; i++ {
		if p == nil {
			return nil
		}
		p = ptrToPtr(p)
	}
	return p
}

func storeUintptr(base unsafe.Pointer, idx uint32, v uintptr) {
	addr := unsafe.Add(base, idx)
	*(*uintptr)(addr) = v
}

func appendInt(ctx *encoder.RuntimeContext, b []byte, p unsafe.Pointer, code *encoder.Opcode) []byte {
	format := ctx.Option.ColorScheme.Int
	b = append(b, format.Header...)
	b = encoder.AppendInt(ctx, b, p, code)
	return append(b, format.Footer...)
}

func appendUint(ctx *encoder.RuntimeContext, b []byte, p unsafe.Pointer, code *encoder.Opcode) []byte {
	format := ctx.Option.ColorScheme.Uint
	b = append(b, format.Header...)
	b = encoder.AppendUint(ctx, b, p, code)
	return append(b, format.Footer...)
}

func appendFloat32(ctx *encoder.RuntimeContext, b []byte, v float32) []byte {
	format := ctx.Option.ColorScheme.Float
	b = append(b, format.Header...)
	b = encoder.AppendFloat32(ctx, b, v)
	return append(b, format.Footer...)
}

func appendFloat64(ctx *encoder.RuntimeContext, b []byte, v float64) []byte {
	format := ctx.Option.ColorScheme.Float
	b = append(b, format.Header...)
	b = encoder.AppendFloat64(ctx, b, v)
	return append(b, format.Footer...)
}

func appendString(ctx *encoder.RuntimeContext, b []byte, v string) []byte {
	format := ctx.Option.ColorScheme.String
	b = append(b, format.Header...)
	b = encoder.AppendString(ctx, b, v)
	return append(b, format.Footer...)
}

func appendByteSlice(ctx *encoder.RuntimeContext, b []byte, src []byte) []byte {
	format := ctx.Option.ColorScheme.Binary
	b = append(b, format.Header...)
	b = encoder.AppendByteSlice(ctx, b, src)
	return append(b, format.Footer...)
}

func appendNumber(ctx *encoder.RuntimeContext, b []byte, n json.Number) ([]byte, error) {
	format := ctx.Option.ColorScheme.Int
	b = append(b, format.Header...)
	bb, err := encoder.AppendNumber(ctx, b, n)
	if err != nil {
		return nil, err
	}
	return append(bb, format.Footer...), nil
}

func appendBool(ctx *encoder.RuntimeContext, b []byte, v bool) []byte {
	format := ctx.Option.ColorScheme.Bool
	b = append(b, format.Header...)
	if v {
		b = append(b, "true"...)
	} else {
		b = append(b, "false"...)
	}
	return append(b, format.Footer...)
}

func appendNull(ctx *encoder.RuntimeContext, b []byte) []byte {
	format := ctx.Option.ColorScheme.Null
	b = append(b, format.Header...)
	b = append(b, "null"...)
	return append(b, format.Footer...)
}

func appendComma(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, ',')
}

func appendNullComma(ctx *encoder.RuntimeContext, b []byte) []byte {
	format := ctx.Option.ColorScheme.Null
	b = append(b, format.Header...)
	b = append(b, "null"...)
	return append(append(b, format.Footer...), ',')
}

func appendColon(_ *encoder.RuntimeContext, b []byte) []byte {
	last := len(b) - 1
	b[last] = ':'
	return b
}

func appendMapKeyValue(_ *encoder.RuntimeContext, _ *encoder.Opcode, b, key, value []byte) []byte {
	b = append(b, key[:len(key)-1]...)
	b = append(b, ':')
	return append(b, value...)
}

func appendMapEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	b[last] = '}'
	b = append(b, ',')
	return b
}

func appendMarshalJSON(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalJSON(ctx, code, b, v)
}

func appendMarshalText(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	format := ctx.Option.ColorScheme.String
	b = append(b, format.Header...)
	bb, err := encoder.AppendMarshalText(ctx, code, b, v)
	if err != nil {
		return nil, err
	}
	return append(bb, format.Footer...), nil
}

func appendArrayHead(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	return append(b, '[')
}

func appendArrayEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	b[last] = ']'
	return append(b, ',')
}

func appendEmptyArray(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, '[', ']', ',')
}

func appendEmptyObject(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, '{', '}', ',')
}

func appendObjectEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	b[last] = '}'
	return append(b, ',')
}

func appendStructHead(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, '{')
}

func appendStructKey(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	format := ctx.Option.ColorScheme.ObjectKey
	b = append(b, format.Header...)
	b = append(b, code.Key[:len(code.Key)-1]...)
	b = append(b, format.Footer...)

	return append(b, ':')
}

func appendStructEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	return append(b, '}', ',')
}

func appendStructEndSkipLast(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	if b[last] == ',' {
		b[last] = '}'
		return appendComma(ctx, b)
	}
	return appendStructEnd(ctx, code, b)
}

func restoreIndent(_ *encoder.RuntimeContext, _ *encoder.Opcode, _ unsafe.Pointer)        {}
func storeIndent(_ unsafe.Pointer, _ *encoder.Opcode, _ uintptr)                          {}
func appendMapKeyIndent(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte    { return b }
func appendArrayElemIndent(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte { return b }

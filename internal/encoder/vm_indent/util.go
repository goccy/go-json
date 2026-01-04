package vm_indent

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/runtime"
)

const uintptrSize = 4 << (^uintptr(0) >> 63)

var (
	appendInt           = encoder.AppendInt
	appendUint          = encoder.AppendUint
	appendFloat32       = encoder.AppendFloat32
	appendFloat64       = encoder.AppendFloat64
	appendString        = encoder.AppendString
	appendByteSlice     = encoder.AppendByteSlice
	appendNumber        = encoder.AppendNumber
	appendStructEnd     = encoder.AppendStructEndIndent
	appendIndent        = encoder.AppendIndent
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
	return fmt.Errorf("encoder (indent): opcode %s has not been implemented", op)
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

func loadPtr(base unsafe.Pointer, idx uint32) unsafe.Pointer {
	addr := unsafe.Add(base, idx)
	return unsafe.Pointer(*(*uintptr)(addr))
}

func loadNPtrPtr(base unsafe.Pointer, idx uint32, ptrNum uint8) unsafe.Pointer {
	addr := unsafe.Add(base, idx)
	p := unsafe.Pointer(*(*uintptr)(addr))
	for i := uint8(0); i < ptrNum; i++ {
		if p == nil {
			return nil
		}
		p = ptrToPtr(p)
	}
	return p
}


func storePtr(base unsafe.Pointer, idx uint32, p unsafe.Pointer) {
	addr := unsafe.Add(base, idx)
	*(*unsafe.Pointer)(addr) = p
}

func storePtrAsUintptr(base unsafe.Pointer, idx uint32, p unsafe.Pointer) {
	addr := unsafe.Add(base, idx)
	*(*uintptr)(addr) = uintptr(p)
}

func storeUintptr(base unsafe.Pointer, idx uint32, v uintptr) {
	addr := unsafe.Add(base, idx)
	*(*uintptr)(addr) = v
}

func appendBool(_ *encoder.RuntimeContext, b []byte, v bool) []byte {
	if v {
		return append(b, "true"...)
	}
	return append(b, "false"...)
}

func appendNull(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, "null"...)
}

func appendComma(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, ',', '\n')
}

func appendNullComma(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, "null,\n"...)
}

func appendColon(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b[:len(b)-2], ':', ' ')
}

func appendMapKeyValue(ctx *encoder.RuntimeContext, code *encoder.Opcode, b, key, value []byte) []byte {
	b = appendIndent(ctx, b, code.Indent+1)
	b = append(b, key...)
	b[len(b)-2] = ':'
	b[len(b)-1] = ' '
	return append(b, value...)
}

func appendMapEnd(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	b = b[:len(b)-2]
	b = append(b, '\n')
	b = appendIndent(ctx, b, code.Indent)
	return append(b, '}', ',', '\n')
}

func appendArrayHead(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	b = append(b, '[', '\n')
	return appendIndent(ctx, b, code.Indent+1)
}

func appendArrayEnd(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	b = b[:len(b)-2]
	b = append(b, '\n')
	b = appendIndent(ctx, b, code.Indent)
	return append(b, ']', ',', '\n')
}

func appendEmptyArray(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, '[', ']', ',', '\n')
}

func appendEmptyObject(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, '{', '}', ',', '\n')
}

func appendObjectEnd(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	// replace comma to newline
	b[last-1] = '\n'
	b = appendIndent(ctx, b[:last], code.Indent)
	return append(b, '}', ',', '\n')
}

func appendMarshalJSON(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalJSONIndent(ctx, code, b, v)
}

func appendMarshalText(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalTextIndent(ctx, code, b, v)
}

func appendStructHead(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, '{', '\n')
}

func appendStructKey(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	b = appendIndent(ctx, b, code.Indent)
	b = append(b, code.Key...)
	return append(b, ' ')
}

func appendStructEndSkipLast(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	if b[last-1] == '{' {
		b[last] = '}'
	} else {
		if b[last] == '\n' {
			// to remove ',' and '\n' characters
			b = b[:len(b)-2]
		}
		b = append(b, '\n')
		b = appendIndent(ctx, b, code.Indent-1)
		b = append(b, '}')
	}
	return appendComma(ctx, b)
}

func restoreIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, ctxptr unsafe.Pointer) {
	ctx.BaseIndent = uint32(uintptr(load(ctxptr, code.Length)))
}

func storeIndent(ctxptr unsafe.Pointer, code *encoder.Opcode, indent uintptr) {
	storeUintptr(ctxptr, code.Length, indent)
}

func appendArrayElemIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	return appendIndent(ctx, b, code.Indent+1)
}

func appendMapKeyIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	return appendIndent(ctx, b, code.Indent)
}

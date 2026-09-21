package vm_indent

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/runtime"
)

// slotSize is the size of a slot of the VM.
const slotSize = unsafe.Sizeof(encoder.Slot{})

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
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

type nonEmptyInterface struct {
	itab *struct {
		ityp unsafe.Pointer // static interface type
		typ  unsafe.Pointer // dynamic concrete type
		// unused fields...
	}
	ptr unsafe.Pointer
}

func errUnimplementedOp(op encoder.OpType) error {
	return fmt.Errorf("encoder (indent): opcode %s has not been implemented", op)
}

// The slots of the VM are separated by what they hold: load / store are for the slots of the pointers,
// which the GC sees, and loadInt / storeInt are for the slots of the other values ( an index, a length, ... ).

func load(slots *encoder.Slots, offset uintptr, idx uint32) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(slots), offset+uintptr(idx)))
}

func store(slots *encoder.Slots, offset uintptr, idx uint32, p unsafe.Pointer) {
	*(*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(slots), offset+uintptr(idx))) = p
}

func loadInt(slots *encoder.Slots, offset uintptr, idx uint32) uintptr {
	return *(*uintptr)(unsafe.Add(unsafe.Pointer(slots), offset+uintptr(idx)))
}

func storeInt(slots *encoder.Slots, offset uintptr, idx uint32, v uintptr) {
	*(*uintptr)(unsafe.Add(unsafe.Pointer(slots), offset+uintptr(idx))) = v
}

func loadNPtr(slots *encoder.Slots, offset uintptr, idx uint32, ptrNum uint8) unsafe.Pointer {
	return ptrToNPtr(load(slots, offset, idx), ptrNum)
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
func ptrToFloat32(p unsafe.Pointer) float32            { return *(*float32)(p) }
func ptrToFloat64(p unsafe.Pointer) float64            { return *(*float64)(p) }
func ptrToBool(p unsafe.Pointer) bool                  { return *(*bool)(p) }
func ptrToBytes(p unsafe.Pointer) []byte               { return *(*[]byte)(p) }
func ptrToNumber(p unsafe.Pointer) json.Number         { return *(*json.Number)(p) }
func ptrToString(p unsafe.Pointer) string              { return *(*string)(p) }
func ptrToSlice(p unsafe.Pointer) *runtime.SliceHeader { return (*runtime.SliceHeader)(p) }
func ptrToPtr(p unsafe.Pointer) unsafe.Pointer         { return *(*unsafe.Pointer)(p) }
func ptrToNPtr(p unsafe.Pointer, ptrNum uint8) unsafe.Pointer {
	for i := uint8(0); i < ptrNum; i++ {
		if p == nil {
			return nil
		}
		p = ptrToPtr(p)
	}
	return p
}

func ptrToInterface(code *encoder.Opcode, p unsafe.Pointer) interface{} {
	return *(*interface{})(unsafe.Pointer(&emptyInterface{
		typ: code.Type,
		ptr: p,
	}))
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

func restoreIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, slots *encoder.Slots, offset uintptr) {
	ctx.BaseIndent = uint32(loadInt(slots, offset, code.Length))
}

func storeIndent(slots *encoder.Slots, offset uintptr, code *encoder.Opcode, indent uintptr) {
	storeInt(slots, offset, code.Length, indent)
}

func appendArrayElemIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	return appendIndent(ctx, b, code.Indent+1)
}

func appendMapKeyIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	return appendIndent(ctx, b, code.Indent)
}

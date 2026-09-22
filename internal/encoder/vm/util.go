package vm

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/runtime"
)

var (
	appendInt           = encoder.AppendInt
	appendUint          = encoder.AppendUint
	appendFloat32       = encoder.AppendFloat32
	appendFloat64       = encoder.AppendFloat64
	appendString        = encoder.AppendString
	appendByteSlice     = encoder.AppendByteSlice
	appendNumber        = encoder.AppendNumber
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

func errUnimplementedOp(op encoder.OpType) error {
	return fmt.Errorf("encoder: opcode %s has not been implemented", op)
}

// load / store are for the half of a slot for a pointer, and loadInt / storeInt are for the other half.
// A pointer is stored as uintptr, which needs no write barrier: see encoder.Slot.

func load(base unsafe.Pointer, idx uint32) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Add(base, idx))
}

func store(base unsafe.Pointer, idx uint32, p unsafe.Pointer) {
	*(*uintptr)(unsafe.Add(base, idx)) = uintptr(p)
}

func loadInt(base unsafe.Pointer, idx uint32) uintptr {
	return *(*uintptr)(unsafe.Add(base, idx))
}

func storeInt(base unsafe.Pointer, idx uint32, v uintptr) {
	*(*uintptr)(unsafe.Add(base, idx)) = v
}

func loadNPtr(base unsafe.Pointer, idx uint32, ptrNum uint8) unsafe.Pointer {
	return ptrToNPtr(load(base, idx), ptrNum)
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
	return append(b, ',')
}

func appendNullComma(_ *encoder.RuntimeContext, b []byte) []byte {
	return append(b, "null,"...)
}

func appendColon(_ *encoder.RuntimeContext, b []byte) []byte {
	last := len(b) - 1
	b[last] = ':'
	return b
}

func appendMapKeyValue(_ *encoder.RuntimeContext, _ *encoder.Opcode, b, key, value []byte) []byte {
	b = append(b, key...)
	b[len(b)-1] = ':'
	return append(b, value...)
}

func appendMapEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	b[len(b)-1] = '}'
	b = append(b, ',')
	return b
}

func appendMarshalJSON(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalJSON(ctx, code, b, v)
}

func appendMarshalText(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalText(ctx, code, b, v)
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

// appendStructKey copies a key up to a chunk by a copy of a whole chunk, not by a call of memmove:
// a call from the VM makes it spill and restore its variables, which costs more than the copy of a key.
// The memory of a key has the bytes after it up to a chunk: see encoder.PaddedKey.
func appendStructKey(_ *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	key := code.Key
	n := len(b)
	if len(key) <= encoder.KeyChunkSize && cap(b)-n >= encoder.KeyChunkSize {
		b = b[:n+encoder.KeyChunkSize]
		*(*[encoder.KeyChunkSize]byte)(unsafe.Pointer(&b[n])) = *(*[encoder.KeyChunkSize]byte)(unsafe.Pointer(unsafe.StringData(key)))
		return b[:n+len(key)]
	}
	return append(b, key...)
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

func appendMapKeyIndent(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte    { return b }
func appendArrayElemIndent(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte { return b }

// appendMapKeyString appends the key of a map which is a plain string, with what follows it:
// the same bytes as the opcode of a string and appendColon append.
func appendMapKeyString(ctx *encoder.RuntimeContext, b []byte, key string) []byte {
	b = appendString(ctx, b, key)
	return append(b, ':')
}

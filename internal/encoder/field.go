package encoder

import (
	"unsafe"
)

// The functions here encode a whole field of a struct ( the key, the value and what follows it ) by a call.
//
// The VM is a function too large for the compiler to keep its variables in the registers, so every call from it
// spills and restores them, and amd64 has few registers. A field encoded here costs the VM a single call.

// AppendIntField appends the key, the integer at p and what follows the field.
func AppendIntField(b []byte, p unsafe.Pointer, code *Opcode, isEnd bool) []byte {
	b = append(b, code.Key...)
	b = AppendInt(nil, b, p, code)
	if isEnd {
		return append(b, '}', ',')
	}
	return append(b, ',')
}

// AppendStringField appends the key, the string and what follows the field.
func AppendStringField(ctx *RuntimeContext, b []byte, s string, code *Opcode, isEnd bool) []byte {
	b = append(b, code.Key...)
	b = AppendString(ctx, b, s)
	if isEnd {
		return append(b, '}', ',')
	}
	return append(b, ',')
}

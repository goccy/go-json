// Package run drives the encoder: it compiles the opcodes of a value, sets up the runtime context and runs the
// VM which the options choose. The v1 and the v2 json packages encode through it.
package run

import (
	"runtime"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/encoder/vm"
	"github.com/goccy/go-json/internal/encoder/vm_color"
	"github.com/goccy/go-json/internal/encoder/vm_color_indent"
	"github.com/goccy/go-json/internal/encoder/vm_indent"
)

type emptyInterface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

// Encode appends v to the buffer of ctx as JSON, with a comma after it, by the options of ctx.
func Encode(ctx *encoder.RuntimeContext, v any) ([]byte, error) {
	b := ctx.Buf[:0]
	if v == nil {
		b = encoder.AppendNull(ctx, b)
		b = encoder.AppendComma(ctx, b)
		return b, nil
	}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ

	typeptr := uintptr(typ)
	codeSet, err := encoder.CompileToGetCodeSet(ctx, typeptr)
	if err != nil {
		return nil, err
	}

	p := ctx.ValueAddr(codeSet, header.ptr)
	if p == nil {
		// only a nil pointer has no address of the value.
		b = encoder.AppendNull(ctx, b)
		b = encoder.AppendComma(ctx, b)
		return b, nil
	}
	ctx.Init(p, codeSet.CodeLength)

	buf, err := Code(ctx, b, codeSet)
	// the VM refers to the value by uintptr.
	runtime.KeepAlive(v)
	if err != nil {
		return nil, err
	}
	ctx.Buf = buf
	return buf, nil
}

// EncodeIndent is Encode with the indentation and the prefix of each line.
func EncodeIndent(ctx *encoder.RuntimeContext, v any, prefix, indent string) ([]byte, error) {
	b := ctx.Buf[:0]
	if v == nil {
		b = encoder.AppendNull(ctx, b)
		b = encoder.AppendCommaIndent(ctx, b)
		return b, nil
	}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ

	typeptr := uintptr(typ)
	codeSet, err := encoder.CompileToGetCodeSet(ctx, typeptr)
	if err != nil {
		return nil, err
	}

	p := ctx.ValueAddr(codeSet, header.ptr)
	if p == nil {
		// only a nil pointer has no address of the value.
		b = encoder.AppendNull(ctx, b)
		b = encoder.AppendCommaIndent(ctx, b)
		return b, nil
	}
	ctx.Init(p, codeSet.CodeLength)
	buf, err := IndentCode(ctx, b, codeSet, prefix, indent)
	// the VM refers to the value by uintptr.
	runtime.KeepAlive(v)

	if err != nil {
		return nil, err
	}

	ctx.Buf = buf
	return buf, nil
}

// Code runs the opcodes of a value, which ctx is set up for, by the VM of the options of ctx.
func Code(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet) ([]byte, error) {
	if (ctx.Option.Flag & encoder.DebugOption) != 0 {
		if (ctx.Option.Flag & encoder.ColorizeOption) != 0 {
			return vm_color.DebugRun(ctx, b, codeSet)
		}
		return vm.DebugRun(ctx, b, codeSet)
	}
	if (ctx.Option.Flag & encoder.ColorizeOption) != 0 {
		return vm_color.Run(ctx, b, codeSet)
	}
	return vm.Run(ctx, b, codeSet)
}

// IndentCode is Code with the indentation and the prefix of each line.
func IndentCode(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, prefix, indent string) ([]byte, error) {
	ctx.Prefix = []byte(prefix)
	ctx.IndentStr = []byte(indent)
	if (ctx.Option.Flag & encoder.DebugOption) != 0 {
		if (ctx.Option.Flag & encoder.ColorizeOption) != 0 {
			return vm_color_indent.DebugRun(ctx, b, codeSet)
		}
		return vm_indent.DebugRun(ctx, b, codeSet)
	}
	if (ctx.Option.Flag & encoder.ColorizeOption) != 0 {
		return vm_color_indent.Run(ctx, b, codeSet)
	}
	return vm_indent.Run(ctx, b, codeSet)
}

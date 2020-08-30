package json

import (
	"reflect"
	"unsafe"
)

type encodeCompileContext struct {
	typ         *rtype
	withIndent  bool
	root        bool
	opcodeIndex int
	indent      int

	parent *encodeCompileContext
}

func (c *encodeCompileContext) context() *encodeCompileContext {
	return &encodeCompileContext{
		typ:         c.typ,
		withIndent:  c.withIndent,
		root:        c.root,
		opcodeIndex: c.opcodeIndex,
		indent:      c.indent,
		parent:      c,
	}
}

func (c *encodeCompileContext) withType(typ *rtype) *encodeCompileContext {
	ctx := c.context()
	ctx.typ = typ
	return ctx
}

func (c *encodeCompileContext) incIndent() *encodeCompileContext {
	ctx := c.context()
	ctx.indent++
	return ctx
}

func (c *encodeCompileContext) decIndent() *encodeCompileContext {
	ctx := c.context()
	ctx.indent--
	return ctx
}

func (c *encodeCompileContext) incOpcodeIndex() {
	c.opcodeIndex++
	if c.parent != nil {
		c.parent.incOpcodeIndex()
	}
}

func (c *encodeCompileContext) decOpcodeIndex() {
	c.opcodeIndex--
	if c.parent != nil {
		c.parent.decOpcodeIndex()
	}
}

type encodeRuntimeContext struct {
	ptrs []uintptr
}

func (c *encodeRuntimeContext) init(p uintptr) {
	c.ptrs[0] = p
}

func (c *encodeRuntimeContext) ptr() uintptr {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&c.ptrs))
	return header.Data
}

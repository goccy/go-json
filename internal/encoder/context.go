package encoder

import (
	"github.com/goccy/go-json/internal/runtime"
)

type compileContext struct {
	typ                      *runtime.Type
	opcodeIndex              int
	ptrIndex                 int
	indent                   int
	structTypeToCompiledCode map[uintptr]*CompiledCode

	parent *compileContext
}

func (c *compileContext) context() *compileContext {
	return &compileContext{
		typ:                      c.typ,
		opcodeIndex:              c.opcodeIndex,
		ptrIndex:                 c.ptrIndex,
		indent:                   c.indent,
		structTypeToCompiledCode: c.structTypeToCompiledCode,
		parent:                   c,
	}
}

func (c *compileContext) withType(typ *runtime.Type) *compileContext {
	ctx := c.context()
	ctx.typ = typ
	return ctx
}

func (c *compileContext) incIndent() *compileContext {
	ctx := c.context()
	ctx.indent++
	return ctx
}

func (c *compileContext) decIndent() *compileContext {
	ctx := c.context()
	ctx.indent--
	return ctx
}

func (c *compileContext) incIndex() {
	c.incOpcodeIndex()
	c.incPtrIndex()
}

func (c *compileContext) decIndex() {
	c.decOpcodeIndex()
	c.decPtrIndex()
}

func (c *compileContext) incOpcodeIndex() {
	c.opcodeIndex++
	if c.parent != nil {
		c.parent.incOpcodeIndex()
	}
}

func (c *compileContext) decOpcodeIndex() {
	c.opcodeIndex--
	if c.parent != nil {
		c.parent.decOpcodeIndex()
	}
}

func (c *compileContext) incPtrIndex() {
	c.ptrIndex++
	if c.parent != nil {
		c.parent.incPtrIndex()
	}
}

func (c *compileContext) decPtrIndex() {
	c.ptrIndex--
	if c.parent != nil {
		c.parent.decPtrIndex()
	}
}

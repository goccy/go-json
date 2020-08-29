package json

type encodeCompileContext struct {
	typ         *rtype
	withIndent  bool
	root        bool
	opcodeIndex int
	indent      int
}

func (c *encodeCompileContext) context() *encodeCompileContext {
	return &encodeCompileContext{
		typ:         c.typ,
		withIndent:  c.withIndent,
		root:        c.root,
		opcodeIndex: c.opcodeIndex,
		indent:      c.indent,
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

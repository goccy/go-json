package json

import (
	"fmt"
	"strings"
	"unsafe"
)

var uintptrSize = unsafe.Sizeof(uintptr(0))

type opcode struct {
	op           opType // operation type
	typ          *rtype // go type
	displayIdx   int    // opcode index
	key          []byte // struct field key
	displayKey   string // key text to display
	isTaggedKey  bool   // whether tagged key
	anonymousKey bool   // whether anonymous key
	root         bool   // whether root
	indent       int    // indent number

	idx     uintptr // offset to access ptr
	headIdx uintptr // offset to access slice/struct head
	elemIdx uintptr // offset to access array/slice/map elem
	length  uintptr // offset to access slice/map length or array length
	mapIter uintptr // offset to access map iterator
	offset  uintptr // offset size from struct header
	size    uintptr // array/slice elem size

	mapKey    *opcode       // map key
	mapValue  *opcode       // map value
	elem      *opcode       // array/slice elem
	end       *opcode       // array/slice/struct/map end
	nextField *opcode       // next struct field
	next      *opcode       // next opcode
	jmp       *compiledCode // for recursive call
}

func newOpCode(ctx *encodeCompileContext, op opType) *opcode {
	return newOpCodeWithNext(ctx, op, newEndOp(ctx))
}

func opcodeOffset(idx int) uintptr {
	return uintptr(idx) * uintptrSize
}

func newOpCodeWithNext(ctx *encodeCompileContext, op opType, next *opcode) *opcode {
	return &opcode{
		op:         op,
		typ:        ctx.typ,
		displayIdx: ctx.opcodeIndex,
		indent:     ctx.indent,
		idx:        opcodeOffset(ctx.opcodeIndex),
		next:       next,
	}
}

func newEndOp(ctx *encodeCompileContext) *opcode {
	return newOpCodeWithNext(ctx, opEnd, nil)
}

func (c *opcode) beforeLastCode() *opcode {
	code := c
	for {
		var nextCode *opcode
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem, codeMapKey:
			nextCode = code.end
		default:
			nextCode = code.next
		}
		if nextCode.op == opEnd {
			return code
		}
		code = nextCode
	}
	return nil
}

func (c *opcode) totalLength() int {
	var idx int
	for code := c; code.op != opEnd; {
		idx = code.displayIdx
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem, codeMapKey:
			code = code.end
		default:
			code = code.next
		}
	}
	return idx + 1
}

func (c *opcode) decOpcodeIndex() {
	for code := c; code.op != opEnd; {
		code.displayIdx--
		code.idx -= 8
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem, codeMapKey:
			code = code.end
		default:
			code = code.next
		}
	}
}

func (c *opcode) dumpElem(code *opcode) string {
	return fmt.Sprintf(
		`[%d]%s%s ([headIdx:%d][elemIdx:%d][length:%d][size:%d])`,
		code.displayIdx,
		strings.Repeat("-", code.indent),
		code.op,
		code.headIdx,
		code.elemIdx,
		code.length,
		code.size,
	)
}

func (c *opcode) dumpField(code *opcode) string {
	return fmt.Sprintf(
		`[%d]%s%s ([key:%s][offset:%d][headIdx:%d])`,
		code.displayIdx,
		strings.Repeat("-", code.indent),
		code.op,
		code.displayKey,
		code.offset,
		code.headIdx,
	)
}

func (c *opcode) dumpKey(code *opcode) string {
	return fmt.Sprintf(
		`[%d]%s%s ([elemIdx:%d][length:%d][mapIter:%d])`,
		code.displayIdx,
		strings.Repeat("-", code.indent),
		code.op,
		code.elemIdx,
		code.length,
		code.mapIter,
	)
}

func (c *opcode) dump() string {
	codes := []string{}
	for code := c; code.op != opEnd; {
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem:
			codes = append(codes, c.dumpElem(code))
			code = code.end
		case codeMapKey:
			codes = append(codes, c.dumpKey(code))
			code = code.end
		case codeStructField:
			codes = append(codes, c.dumpField(code))
			code = code.next
		default:
			codes = append(codes, fmt.Sprintf(
				"[%d]%s%s",
				code.displayIdx,
				strings.Repeat("-", code.indent),
				code.op,
			))
			code = code.next
		}
	}
	return strings.Join(codes, "\n")
}

func linkPrevToNextField(prev, cur *opcode) {
	prev.nextField = cur.nextField
	code := prev
	fcode := cur
	for {
		var nextCode *opcode
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem, codeMapKey:
			nextCode = code.end
		default:
			nextCode = code.next
		}
		if nextCode == fcode {
			code.next = fcode.next
			break
		} else if nextCode.op == opEnd {
			break
		}
		code = nextCode
	}
}

func newSliceHeaderCode(ctx *encodeCompileContext) *opcode {
	return &opcode{
		op:         opSliceHead,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		indent:     ctx.indent,
	}
}

func newSliceElemCode(ctx *encodeCompileContext, size uintptr) *opcode {
	return &opcode{
		op:         opSliceElem,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		indent:     ctx.indent,
		size:       size,
	}
}

func newArrayHeaderCode(ctx *encodeCompileContext, alen int) *opcode {
	return &opcode{
		op:         opArrayHead,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		indent:     ctx.indent,
		length:     uintptr(alen),
	}
}

func newArrayElemCode(ctx *encodeCompileContext, alen int, size uintptr) *opcode {
	return &opcode{
		op:         opArrayElem,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		length:     uintptr(alen),
		size:       size,
	}
}

func newMapHeaderCode(ctx *encodeCompileContext, withLoad bool) *opcode {
	var op opType
	if withLoad {
		op = opMapHeadLoad
	} else {
		op = opMapHead
	}
	return &opcode{
		op:         op,
		typ:        ctx.typ,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		indent:     ctx.indent,
	}
}

func newMapKeyCode(ctx *encodeCompileContext) *opcode {
	return &opcode{
		op:         opMapKey,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		indent:     ctx.indent,
	}
}

func newMapValueCode(ctx *encodeCompileContext) *opcode {
	return &opcode{
		op:         opMapValue,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		indent:     ctx.indent,
	}
}

func newInterfaceCode(ctx *encodeCompileContext) *opcode {
	return &opcode{
		op:         opInterface,
		typ:        ctx.typ,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		indent:     ctx.indent,
		next:       newEndOp(ctx),
		root:       ctx.root,
	}
}

func newRecursiveCode(ctx *encodeCompileContext, jmp *compiledCode) *opcode {
	return &opcode{
		op:         opStructFieldRecursive,
		typ:        ctx.typ,
		displayIdx: ctx.opcodeIndex,
		idx:        opcodeOffset(ctx.opcodeIndex),
		indent:     ctx.indent,
		next:       newEndOp(ctx),
		jmp:        jmp,
	}
}

//func newRecursiveCode(recursive *recursiveCode) *opcode {
//code := copyOpcode(recursive.jmp.code)
//head := (*structFieldCode)(unsafe.Pointer(code))
//head.end.next = newEndOp(&encodeCompileContext{})

//code.op = code.op.ptrHeadToHead()
//	return code
//}

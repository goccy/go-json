package json

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

func copyOpcode(code *opcode) *opcode {
	codeMap := map[uintptr]*opcode{}
	return code.copy(codeMap)
}

type opcodeHeader struct {
	op     opType
	typ    *rtype
	ptr    uintptr
	indent int
	next   *opcode
}

func (h *opcodeHeader) copy(codeMap map[uintptr]*opcode) *opcodeHeader {
	return &opcodeHeader{
		op:     h.op,
		typ:    h.typ,
		ptr:    h.ptr,
		indent: h.indent,
		next:   h.next.copy(codeMap),
	}
}

type opcode struct {
	*opcodeHeader
}

func newOpCode(op opType, typ *rtype, indent int, next *opcode) *opcode {
	return &opcode{
		opcodeHeader: &opcodeHeader{
			op:     op,
			typ:    typ,
			indent: indent,
			next:   next,
		},
	}
}

func newEndOp(indent int) *opcode {
	return newOpCode(opEnd, nil, indent, nil)
}

func (c *opcode) beforeLastCode() *opcode {
	code := c
	for {
		var nextCode *opcode
		switch code.op.codeType() {
		case codeArrayElem:
			nextCode = code.toArrayElemCode().end
		case codeSliceElem:
			nextCode = code.toSliceElemCode().end
		case codeMapKey:
			nextCode = code.toMapKeyCode().end
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

func (c *opcode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	var code *opcode
	switch c.op.codeType() {
	case codeArrayHead:
		code = c.toArrayHeaderCode().copy(codeMap)
	case codeArrayElem:
		code = c.toArrayElemCode().copy(codeMap)
	case codeSliceHead:
		code = c.toSliceHeaderCode().copy(codeMap)
	case codeSliceElem:
		code = c.toSliceElemCode().copy(codeMap)
	case codeMapHead:
		code = c.toMapHeadCode().copy(codeMap)
	case codeMapKey:
		code = c.toMapKeyCode().copy(codeMap)
	case codeMapValue:
		code = c.toMapValueCode().copy(codeMap)
	case codeStructFieldRecursive:
		code = c.toRecursiveCode().copy(codeMap)
	case codeStructField:
		code = c.toStructFieldCode().copy(codeMap)
	default:
		code = &opcode{}
		codeMap[addr] = code

		code.opcodeHeader = c.opcodeHeader.copy(codeMap)
	}
	return code
}

func (c *opcode) dump() string {
	codes := []string{}
	for code := c; code.op != opEnd; {
		indent := strings.Repeat(" ", code.indent)
		codes = append(codes, fmt.Sprintf("%s%s ( %p )", indent, code.op, unsafe.Pointer(code)))
		switch code.op.codeType() {
		case codeArrayElem:
			code = code.toArrayElemCode().end
		case codeSliceElem:
			code = code.toSliceElemCode().end
		case codeMapKey:
			code = code.toMapKeyCode().end
		default:
			code = code.next
		}
	}
	return strings.Join(codes, "\n")
}

func (c *opcode) toSliceHeaderCode() *sliceHeaderCode {
	return (*sliceHeaderCode)(unsafe.Pointer(c))
}

func (c *opcode) toSliceElemCode() *sliceElemCode {
	return (*sliceElemCode)(unsafe.Pointer(c))
}

func (c *opcode) toArrayHeaderCode() *arrayHeaderCode {
	return (*arrayHeaderCode)(unsafe.Pointer(c))
}

func (c *opcode) toArrayElemCode() *arrayElemCode {
	return (*arrayElemCode)(unsafe.Pointer(c))
}

func (c *opcode) toStructFieldCode() *structFieldCode {
	return (*structFieldCode)(unsafe.Pointer(c))
}

func (c *opcode) toMapHeadCode() *mapHeaderCode {
	return (*mapHeaderCode)(unsafe.Pointer(c))
}

func (c *opcode) toMapKeyCode() *mapKeyCode {
	return (*mapKeyCode)(unsafe.Pointer(c))
}

func (c *opcode) toMapValueCode() *mapValueCode {
	return (*mapValueCode)(unsafe.Pointer(c))
}

func (c *opcode) toInterfaceCode() *interfaceCode {
	return (*interfaceCode)(unsafe.Pointer(c))
}

func (c *opcode) toRecursiveCode() *recursiveCode {
	return (*recursiveCode)(unsafe.Pointer(c))
}

type sliceHeaderCode struct {
	*opcodeHeader
	elem *sliceElemCode
	end  *opcode
}

func newSliceHeaderCode(indent int) *sliceHeaderCode {
	return &sliceHeaderCode{
		opcodeHeader: &opcodeHeader{
			op:     opSliceHead,
			indent: indent,
		},
	}
}

func (c *sliceHeaderCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	header := &sliceHeaderCode{}
	code := (*opcode)(unsafe.Pointer(header))
	codeMap[addr] = code

	header.opcodeHeader = c.opcodeHeader.copy(codeMap)
	header.elem = (*sliceElemCode)(unsafe.Pointer(c.elem.copy(codeMap)))
	header.end = c.end.copy(codeMap)
	return code
}

type sliceElemCode struct {
	*opcodeHeader
	idx  uintptr
	len  uintptr
	size uintptr
	data uintptr
	end  *opcode
}

func (c *sliceElemCode) set(header *reflect.SliceHeader) {
	c.idx = uintptr(0)
	c.len = uintptr(header.Len)
	c.data = header.Data
}

func (c *sliceElemCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	elem := &sliceElemCode{
		idx:  c.idx,
		len:  c.len,
		size: c.size,
		data: c.data,
	}
	code := (*opcode)(unsafe.Pointer(elem))
	codeMap[addr] = code

	elem.opcodeHeader = c.opcodeHeader.copy(codeMap)
	elem.end = c.end.copy(codeMap)
	return code
}

type arrayHeaderCode struct {
	*opcodeHeader
	len  uintptr
	elem *arrayElemCode
	end  *opcode
}

func newArrayHeaderCode(indent, alen int) *arrayHeaderCode {
	return &arrayHeaderCode{
		opcodeHeader: &opcodeHeader{
			op:     opArrayHead,
			indent: indent,
		},
		len: uintptr(alen),
	}
}

func (c *arrayHeaderCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	header := &arrayHeaderCode{}
	code := (*opcode)(unsafe.Pointer(header))
	codeMap[addr] = code

	header.opcodeHeader = c.opcodeHeader.copy(codeMap)
	header.len = c.len
	header.elem = (*arrayElemCode)(unsafe.Pointer(c.elem.copy(codeMap)))
	header.end = c.end.copy(codeMap)
	return code
}

type arrayElemCode struct {
	*opcodeHeader
	idx  uintptr
	len  uintptr
	size uintptr
	end  *opcode
}

func (c *arrayElemCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	elem := &arrayElemCode{
		idx:  c.idx,
		len:  c.len,
		size: c.size,
	}
	code := (*opcode)(unsafe.Pointer(elem))
	codeMap[addr] = code

	elem.opcodeHeader = c.opcodeHeader.copy(codeMap)
	elem.end = c.end.copy(codeMap)
	return code
}

type structFieldCode struct {
	*opcodeHeader
	key          []byte
	displayKey   string
	offset       uintptr
	anonymousKey bool
	nextField    *opcode
	end          *opcode
}

func linkPrevToNextField(prev, cur *structFieldCode) {
	prev.nextField = cur.nextField
	code := prev.toOpcode()
	fcode := cur.toOpcode()
	for {
		var nextCode *opcode
		switch code.op.codeType() {
		case codeArrayElem:
			nextCode = code.toArrayElemCode().end
		case codeSliceElem:
			nextCode = code.toSliceElemCode().end
		case codeMapKey:
			nextCode = code.toMapKeyCode().end
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

func (c *structFieldCode) toOpcode() *opcode {
	return (*opcode)(unsafe.Pointer(c))
}

func (c *structFieldCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	field := &structFieldCode{
		key:          c.key,
		displayKey:   c.displayKey,
		anonymousKey: c.anonymousKey,
		offset:       c.offset,
	}
	code := (*opcode)(unsafe.Pointer(field))
	codeMap[addr] = code

	field.opcodeHeader = c.opcodeHeader.copy(codeMap)
	field.nextField = c.nextField.copy(codeMap)
	field.end = c.end.copy(codeMap)
	return code
}

type mapHeaderCode struct {
	*opcodeHeader
	key   *mapKeyCode
	value *mapValueCode
	end   *opcode
}

func (c *mapHeaderCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	header := &mapHeaderCode{}
	code := (*opcode)(unsafe.Pointer(header))
	codeMap[addr] = code

	header.opcodeHeader = c.opcodeHeader.copy(codeMap)
	header.key = (*mapKeyCode)(unsafe.Pointer(c.key.copy(codeMap)))
	header.value = (*mapValueCode)(unsafe.Pointer(c.value.copy(codeMap)))
	header.end = c.end.copy(codeMap)
	return code
}

type mapKeyCode struct {
	*opcodeHeader
	idx  int
	len  int
	iter unsafe.Pointer
	end  *opcode
}

func (c *mapKeyCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	key := &mapKeyCode{
		idx:  c.idx,
		len:  c.len,
		iter: c.iter,
	}
	code := (*opcode)(unsafe.Pointer(key))
	codeMap[addr] = code

	key.opcodeHeader = c.opcodeHeader.copy(codeMap)
	key.end = c.end.copy(codeMap)
	return code
}

func (c *mapKeyCode) set(len int, iter unsafe.Pointer) {
	c.idx = 0
	c.len = len
	c.iter = iter
}

type mapValueCode struct {
	*opcodeHeader
	iter unsafe.Pointer
}

func (c *mapValueCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	value := &mapValueCode{
		iter: c.iter,
	}
	code := (*opcode)(unsafe.Pointer(value))
	codeMap[addr] = code

	value.opcodeHeader = c.opcodeHeader.copy(codeMap)
	return code
}

func (c *mapValueCode) set(iter unsafe.Pointer) {
	c.iter = iter
}

func newMapHeaderCode(typ *rtype, withLoad bool, indent int) *mapHeaderCode {
	var op opType
	if withLoad {
		op = opMapHeadLoad
	} else {
		op = opMapHead
	}
	return &mapHeaderCode{
		opcodeHeader: &opcodeHeader{
			op:     op,
			typ:    typ,
			indent: indent,
		},
	}
}

func newMapKeyCode(indent int) *mapKeyCode {
	return &mapKeyCode{
		opcodeHeader: &opcodeHeader{
			op:     opMapKey,
			indent: indent,
		},
	}
}

func newMapValueCode(indent int) *mapValueCode {
	return &mapValueCode{
		opcodeHeader: &opcodeHeader{
			op:     opMapValue,
			indent: indent,
		},
	}
}

type interfaceCode struct {
	*opcodeHeader
	root bool
}

func (c *interfaceCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	iface := &interfaceCode{}
	code := (*opcode)(unsafe.Pointer(iface))
	codeMap[addr] = code

	iface.opcodeHeader = c.opcodeHeader.copy(codeMap)
	return code
}

type recursiveCode struct {
	*opcodeHeader
	jmp     *compiledCode
	seenPtr uintptr
}

func (c *recursiveCode) copy(codeMap map[uintptr]*opcode) *opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	recur := &recursiveCode{seenPtr: c.seenPtr}
	code := (*opcode)(unsafe.Pointer(recur))
	codeMap[addr] = code

	recur.opcodeHeader = c.opcodeHeader.copy(codeMap)
	recur.jmp = &compiledCode{
		code: c.jmp.code.copy(codeMap),
	}
	return code
}

func newRecursiveCode(recursive *recursiveCode) *opcode {
	code := copyOpcode(recursive.jmp.code)
	head := (*structFieldCode)(unsafe.Pointer(code))
	head.end.next = newEndOp(0)
	code.ptr = recursive.ptr

	code.op = code.op.ptrHeadToHead()
	return code
}

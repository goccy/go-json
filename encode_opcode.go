package json

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

type opType int

const (
	opEnd opType = iota
	opInt
	opInt8
	opInt16
	opInt32
	opInt64
	opUint
	opUint8
	opUint16
	opUint32
	opUint64
	opFloat32
	opFloat64
	opString
	opBool
	opPtr
	opSliceHead
	opSliceElemFirst
	opSliceElem
	opSliceEnd
	opStructHead
	opStructFieldFirst
	opStructField
	opStructFieldFirstInt
	opStructFieldInt
	opStructFieldFirstString
	opStructFieldString
	opStructEnd
)

func (t opType) String() string {
	switch t {
	case opEnd:
		return "END"
	case opInt:
		return "INT"
	case opInt8:
		return "INT8"
	case opInt16:
		return "INT16"
	case opInt32:
		return "INT32"
	case opInt64:
		return "INT64"
	case opUint:
		return "UINT"
	case opUint8:
		return "UINT8"
	case opUint16:
		return "UINT16"
	case opUint32:
		return "UINT32"
	case opUint64:
		return "UINT64"
	case opFloat32:
		return "FLOAT32"
	case opFloat64:
		return "FLOAT64"
	case opString:
		return "STRING"
	case opBool:
		return "BOOL"
	case opPtr:
		return "PTR"
	case opSliceHead:
		return "SLICE_HEAD"
	case opSliceElemFirst:
		return "SLICE_ELEM_FIRST"
	case opSliceElem:
		return "SLICE_ELEM"
	case opSliceEnd:
		return "SLICE_END"
	case opStructHead:
		return "STRUCT_HEAD"
	case opStructFieldFirst:
		return "STRUCT_FIELD_FIRST"
	case opStructField:
		return "STRUCT_FIELD"
	case opStructFieldFirstInt:
		return "STRUCT_FIELD_FIRST_INT"
	case opStructFieldInt:
		return "STRUCT_FIELD_INT"
	case opStructFieldFirstString:
		return "STRUCT_FIELD_FIRST_STRING"
	case opStructFieldString:
		return "STRUCT_FIELD_STRING"
	case opStructEnd:
		return "STRUCT_END"
	}
	return ""
}

type opcodeHeader struct {
	op   opType
	typ  *rtype
	ptr  uintptr
	next *opcode
}

type opcode struct {
	*opcodeHeader
}

func newOpCode(op opType, typ *rtype, next *opcode) *opcode {
	return &opcode{
		opcodeHeader: &opcodeHeader{
			op:   op,
			typ:  typ,
			next: next,
		},
	}
}

func newEndOp() *opcode {
	return newOpCode(opEnd, nil, nil)
}

func (c *opcode) beforeLastCode() *opcode {
	code := c
	for {
		var nextCode *opcode
		if code.op == opSliceElem {
			nextCode = code.toSliceElemCode().end
		} else {
			nextCode = code.next
		}
		if nextCode.op == opEnd {
			return code
		}
		code = nextCode
	}
	return nil
}

func (c *opcode) dump() string {
	codes := []string{}
	for code := c; code.op != opEnd; {
		codes = append(codes, fmt.Sprintf("%s", code.op))
		if code.op == opSliceElem {
			code = code.toSliceElemCode().end
		} else {
			code = code.next
		}
	}
	return strings.Join(codes, "\n")
}

func (c *opcode) toSliceElemCode() *sliceElemCode {
	return (*sliceElemCode)(unsafe.Pointer(c))
}

func (c *opcode) toStructHeaderCode() *structHeaderCode {
	return (*structHeaderCode)(unsafe.Pointer(c))
}

func (c *opcode) toStructFieldCode() *structFieldCode {
	return (*structFieldCode)(unsafe.Pointer(c))
}

type sliceElemCode struct {
	*opcodeHeader
	idx  uintptr
	len  uintptr
	size uintptr
	data uintptr
	elem *sliceElemCode // first => elem
	end  *opcode
}

func (c *sliceElemCode) set(header *reflect.SliceHeader) {
	c.idx = uintptr(0)
	c.len = uintptr(header.Len)
	c.data = header.Data
}

type structHeaderCode struct {
	*opcodeHeader
	end *opcode
}

type structFieldCode struct {
	*opcodeHeader
	key       string
	offset    uintptr
	nextField *opcode
}

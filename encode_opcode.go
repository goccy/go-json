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
	opInterface
	opPtr
	opSliceHead
	opSliceElem
	opSliceEnd
	opArrayHead
	opArrayElem
	opArrayEnd
	opMapHead
	opMapKey
	opMapValue
	opMapEnd

	// StructFieldHead
	opStructFieldHead
	opStructFieldHeadInt
	opStructFieldHeadInt8
	opStructFieldHeadInt16
	opStructFieldHeadInt32
	opStructFieldHeadInt64
	opStructFieldHeadUint
	opStructFieldHeadUint8
	opStructFieldHeadUint16
	opStructFieldHeadUint32
	opStructFieldHeadUint64
	opStructFieldHeadFloat32
	opStructFieldHeadFloat64
	opStructFieldHeadString
	opStructFieldHeadBool

	// StructFieldHead with omitempty
	opStructFieldHeadOmitEmpty
	opStructFieldHeadIntOmitEmpty
	opStructFieldHeadInt8OmitEmpty
	opStructFieldHeadInt16OmitEmpty
	opStructFieldHeadInt32OmitEmpty
	opStructFieldHeadInt64OmitEmpty
	opStructFieldHeadUintOmitEmpty
	opStructFieldHeadUint8OmitEmpty
	opStructFieldHeadUint16OmitEmpty
	opStructFieldHeadUint32OmitEmpty
	opStructFieldHeadUint64OmitEmpty
	opStructFieldHeadFloat32OmitEmpty
	opStructFieldHeadFloat64OmitEmpty
	opStructFieldHeadStringOmitEmpty
	opStructFieldHeadBoolOmitEmpty

	// StructFieldHead for pointer structure
	opStructFieldPtrHead
	opStructFieldPtrHeadInt
	opStructFieldPtrHeadInt8
	opStructFieldPtrHeadInt16
	opStructFieldPtrHeadInt32
	opStructFieldPtrHeadInt64
	opStructFieldPtrHeadUint
	opStructFieldPtrHeadUint8
	opStructFieldPtrHeadUint16
	opStructFieldPtrHeadUint32
	opStructFieldPtrHeadUint64
	opStructFieldPtrHeadFloat32
	opStructFieldPtrHeadFloat64
	opStructFieldPtrHeadString
	opStructFieldPtrHeadBool

	// StructFieldPtrHead with omitempty
	opStructFieldPtrHeadOmitEmpty
	opStructFieldPtrHeadIntOmitEmpty
	opStructFieldPtrHeadInt8OmitEmpty
	opStructFieldPtrHeadInt16OmitEmpty
	opStructFieldPtrHeadInt32OmitEmpty
	opStructFieldPtrHeadInt64OmitEmpty
	opStructFieldPtrHeadUintOmitEmpty
	opStructFieldPtrHeadUint8OmitEmpty
	opStructFieldPtrHeadUint16OmitEmpty
	opStructFieldPtrHeadUint32OmitEmpty
	opStructFieldPtrHeadUint64OmitEmpty
	opStructFieldPtrHeadFloat32OmitEmpty
	opStructFieldPtrHeadFloat64OmitEmpty
	opStructFieldPtrHeadStringOmitEmpty
	opStructFieldPtrHeadBoolOmitEmpty

	// StructField
	opStructField
	opStructFieldInt
	opStructFieldInt8
	opStructFieldInt16
	opStructFieldInt32
	opStructFieldInt64
	opStructFieldUint
	opStructFieldUint8
	opStructFieldUint16
	opStructFieldUint32
	opStructFieldUint64
	opStructFieldFloat32
	opStructFieldFloat64
	opStructFieldString
	opStructFieldBool

	// StructField with omitempty
	opStructFieldOmitEmpty
	opStructFieldIntOmitEmpty
	opStructFieldInt8OmitEmpty
	opStructFieldInt16OmitEmpty
	opStructFieldInt32OmitEmpty
	opStructFieldInt64OmitEmpty
	opStructFieldUintOmitEmpty
	opStructFieldUint8OmitEmpty
	opStructFieldUint16OmitEmpty
	opStructFieldUint32OmitEmpty
	opStructFieldUint64OmitEmpty
	opStructFieldFloat32OmitEmpty
	opStructFieldFloat64OmitEmpty
	opStructFieldStringOmitEmpty
	opStructFieldBoolOmitEmpty

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
	case opInterface:
		return "INTERFACE"
	case opPtr:
		return "PTR"
	case opSliceHead:
		return "SLICE_HEAD"
	case opSliceElem:
		return "SLICE_ELEM"
	case opSliceEnd:
		return "SLICE_END"
	case opArrayHead:
		return "ARRAY_HEAD"
	case opArrayElem:
		return "ARRAY_ELEM"
	case opArrayEnd:
		return "ARRAY_END"
	case opMapHead:
		return "MAP_HEAD"
	case opMapKey:
		return "MAP_KEY"
	case opMapValue:
		return "MAP_VALUE"
	case opMapEnd:
		return "MAP_END"

	case opStructFieldHead:
		return "STRUCT_FIELD_HEAD"
	case opStructFieldHeadInt:
		return "STRUCT_FIELD_HEAD_INT"
	case opStructFieldHeadInt8:
		return "STRUCT_FIELD_HEAD_INT8"
	case opStructFieldHeadInt16:
		return "STRUCT_FIELD_HEAD_INT16"
	case opStructFieldHeadInt32:
		return "STRUCT_FIELD_HEAD_INT32"
	case opStructFieldHeadInt64:
		return "STRUCT_FIELD_HEAD_INT64"
	case opStructFieldHeadUint:
		return "STRUCT_FIELD_HEAD_UINT"
	case opStructFieldHeadUint8:
		return "STRUCT_FIELD_HEAD_UINT8"
	case opStructFieldHeadUint16:
		return "STRUCT_FIELD_HEAD_UINT16"
	case opStructFieldHeadUint32:
		return "STRUCT_FIELD_HEAD_UINT32"
	case opStructFieldHeadUint64:
		return "STRUCT_FIELD_HEAD_UINT64"
	case opStructFieldHeadFloat32:
		return "STRUCT_FIELD_HEAD_FLOAT32"
	case opStructFieldHeadFloat64:
		return "STRUCT_FIELD_HEAD_FLOAT64"
	case opStructFieldHeadString:
		return "STRUCT_FIELD_HEAD_STRING"
	case opStructFieldHeadBool:
		return "STRUCT_FIELD_HEAD_BOOL"

	case opStructFieldHeadOmitEmpty:
		return "STRUCT_FIELD_HEAD_OMIT_EMPTY"
	case opStructFieldHeadIntOmitEmpty:
		return "STRUCT_FIELD_HEAD_INT_OMIT_EMPTY"
	case opStructFieldHeadInt8OmitEmpty:
		return "STRUCT_FIELD_HEAD_INT8_OMIT_EMPTY"
	case opStructFieldHeadInt16OmitEmpty:
		return "STRUCT_FIELD_HEAD_INT16_OMIT_EMPTY"
	case opStructFieldHeadInt32OmitEmpty:
		return "STRUCT_FIELD_HEAD_INT32_OMIT_EMPTY"
	case opStructFieldHeadInt64OmitEmpty:
		return "STRUCT_FIELD_HEAD_INT64_OMIT_EMPTY"
	case opStructFieldHeadUintOmitEmpty:
		return "STRUCT_FIELD_HEAD_UINT_OMIT_EMPTY"
	case opStructFieldHeadUint8OmitEmpty:
		return "STRUCT_FIELD_HEAD_UINT8_OMIT_EMPTY"
	case opStructFieldHeadUint16OmitEmpty:
		return "STRUCT_FIELD_HEAD_UINT16_OMIT_EMPTY"
	case opStructFieldHeadUint32OmitEmpty:
		return "STRUCT_FIELD_HEAD_UINT32_OMIT_EMPTY"
	case opStructFieldHeadUint64OmitEmpty:
		return "STRUCT_FIELD_HEAD_UINT64_OMIT_EMPTY"
	case opStructFieldHeadFloat32OmitEmpty:
		return "STRUCT_FIELD_HEAD_FLOAT32_OMIT_EMPTY"
	case opStructFieldHeadFloat64OmitEmpty:
		return "STRUCT_FIELD_HEAD_FLOAT64_OMIT_EMPTY"
	case opStructFieldHeadStringOmitEmpty:
		return "STRUCT_FIELD_HEAD_STRING_OMIT_EMPTY"
	case opStructFieldHeadBoolOmitEmpty:
		return "STRUCT_FIELD_HEAD_BOOL_OMIT_EMPTY"

	case opStructFieldPtrHead:
		return "STRUCT_FIELD_PTR_HEAD"
	case opStructFieldPtrHeadInt:
		return "STRUCT_FIELD_PTR_HEAD_INT"
	case opStructFieldPtrHeadInt8:
		return "STRUCT_FIELD_PTR_HEAD_INT8"
	case opStructFieldPtrHeadInt16:
		return "STRUCT_FIELD_PTR_HEAD_INT16"
	case opStructFieldPtrHeadInt32:
		return "STRUCT_FIELD_PTR_HEAD_INT32"
	case opStructFieldPtrHeadInt64:
		return "STRUCT_FIELD_PTR_HEAD_INT64"
	case opStructFieldPtrHeadUint:
		return "STRUCT_FIELD_PTR_HEAD_UINT"
	case opStructFieldPtrHeadUint8:
		return "STRUCT_FIELD_PTR_HEAD_UINT8"
	case opStructFieldPtrHeadUint16:
		return "STRUCT_FIELD_PTR_HEAD_UINT16"
	case opStructFieldPtrHeadUint32:
		return "STRUCT_FIELD_PTR_HEAD_UINT32"
	case opStructFieldPtrHeadUint64:
		return "STRUCT_FIELD_PTR_HEAD_UINT64"
	case opStructFieldPtrHeadFloat32:
		return "STRUCT_FIELD_PTR_HEAD_FLOAT32"
	case opStructFieldPtrHeadFloat64:
		return "STRUCT_FIELD_PTR_HEAD_FLOAT64"
	case opStructFieldPtrHeadString:
		return "STRUCT_FIELD_PTR_HEAD_STRING"
	case opStructFieldPtrHeadBool:
		return "STRUCT_FIELD_PTR_HEAD_BOOL"

	case opStructFieldPtrHeadOmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_OMIT_EMPTY"
	case opStructFieldPtrHeadIntOmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_INT_OMIT_EMPTY"
	case opStructFieldPtrHeadInt8OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_INT8_OMIT_EMPTY"
	case opStructFieldPtrHeadInt16OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_INT16_OMIT_EMPTY"
	case opStructFieldPtrHeadInt32OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_INT32_OMIT_EMPTY"
	case opStructFieldPtrHeadInt64OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_INT64_OMIT_EMPTY"
	case opStructFieldPtrHeadUintOmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_UINT_OMIT_EMPTY"
	case opStructFieldPtrHeadUint8OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_UINT8_OMIT_EMPTY"
	case opStructFieldPtrHeadUint16OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_UINT16_OMIT_EMPTY"
	case opStructFieldPtrHeadUint32OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_UINT32_OMIT_EMPTY"
	case opStructFieldPtrHeadUint64OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_UINT64_OMIT_EMPTY"
	case opStructFieldPtrHeadFloat32OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_FLOAT32_OMIT_EMPTY"
	case opStructFieldPtrHeadFloat64OmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_FLOAT64_OMIT_EMPTY"
	case opStructFieldPtrHeadStringOmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_STRING_OMIT_EMPTY"
	case opStructFieldPtrHeadBoolOmitEmpty:
		return "STRUCT_FIELD_PTR_HEAD_BOOL_OMIT_EMPTY"

	case opStructField:
		return "STRUCT_FIELD"
	case opStructFieldInt:
		return "STRUCT_FIELD_INT"
	case opStructFieldInt8:
		return "STRUCT_FIELD_INT8"
	case opStructFieldInt16:
		return "STRUCT_FIELD_INT16"
	case opStructFieldInt32:
		return "STRUCT_FIELD_INT32"
	case opStructFieldInt64:
		return "STRUCT_FIELD_INT64"
	case opStructFieldUint:
		return "STRUCT_FIELD_UINT"
	case opStructFieldUint8:
		return "STRUCT_FIELD_UINT8"
	case opStructFieldUint16:
		return "STRUCT_FIELD_UINT16"
	case opStructFieldUint32:
		return "STRUCT_FIELD_UINT32"
	case opStructFieldUint64:
		return "STRUCT_FIELD_UINT64"
	case opStructFieldFloat32:
		return "STRUCT_FIELD_FLOAT32"
	case opStructFieldFloat64:
		return "STRUCT_FIELD_FLOAT64"
	case opStructFieldString:
		return "STRUCT_FIELD_STRING"
	case opStructFieldBool:
		return "STRUCT_FIELD_BOOL"

	case opStructFieldOmitEmpty:
		return "STRUCT_FIELD_OMIT_EMPTY"
	case opStructFieldIntOmitEmpty:
		return "STRUCT_FIELD_INT_OMIT_EMPTY"
	case opStructFieldInt8OmitEmpty:
		return "STRUCT_FIELD_INT8_OMIT_EMPTY"
	case opStructFieldInt16OmitEmpty:
		return "STRUCT_FIELD_INT16_OMIT_EMPTY"
	case opStructFieldInt32OmitEmpty:
		return "STRUCT_FIELD_INT32_OMIT_EMPTY"
	case opStructFieldInt64OmitEmpty:
		return "STRUCT_FIELD_INT64_OMIT_EMPTY"
	case opStructFieldUintOmitEmpty:
		return "STRUCT_FIELD_UINT_OMIT_EMPTY"
	case opStructFieldUint8OmitEmpty:
		return "STRUCT_FIELD_UINT8_OMIT_EMPTY"
	case opStructFieldUint16OmitEmpty:
		return "STRUCT_FIELD_UINT16_OMIT_EMPTY"
	case opStructFieldUint32OmitEmpty:
		return "STRUCT_FIELD_UINT32_OMIT_EMPTY"
	case opStructFieldUint64OmitEmpty:
		return "STRUCT_FIELD_UINT64_OMIT_EMPTY"
	case opStructFieldFloat32OmitEmpty:
		return "STRUCT_FIELD_FLOAT32_OMIT_EMPTY"
	case opStructFieldFloat64OmitEmpty:
		return "STRUCT_FIELD_FLOAT64_OMIT_EMPTY"
	case opStructFieldStringOmitEmpty:
		return "STRUCT_FIELD_STRING_OMIT_EMPTY"
	case opStructFieldBoolOmitEmpty:
		return "STRUCT_FIELD_BOOL_OMIT_EMPTY"

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
		switch code.op {
		case opArrayElem:
			nextCode = code.toArrayElemCode().end
		case opSliceElem:
			nextCode = code.toSliceElemCode().end
		case opMapKey:
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

func (c *opcode) dump() string {
	codes := []string{}
	for code := c; code.op != opEnd; {
		codes = append(codes, fmt.Sprintf("%s", code.op))
		switch code.op {
		case opArrayElem:
			code = code.toArrayElemCode().end
		case opSliceElem:
			code = code.toSliceElemCode().end
		case opMapKey:
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

type sliceHeaderCode struct {
	*opcodeHeader
	elem *sliceElemCode
	end  *opcode
}

func newSliceHeaderCode() *sliceHeaderCode {
	return &sliceHeaderCode{
		opcodeHeader: &opcodeHeader{
			op: opSliceHead,
		},
	}
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

type arrayHeaderCode struct {
	*opcodeHeader
	len  uintptr
	elem *arrayElemCode
	end  *opcode
}

func newArrayHeaderCode(alen int) *arrayHeaderCode {
	return &arrayHeaderCode{
		opcodeHeader: &opcodeHeader{
			op: opArrayHead,
		},
		len: uintptr(alen),
	}
}

type arrayElemCode struct {
	*opcodeHeader
	idx  uintptr
	len  uintptr
	size uintptr
	end  *opcode
}

type structFieldCode struct {
	*opcodeHeader
	key       []byte
	offset    uintptr
	nextField *opcode
	end       *opcode
}

type mapHeaderCode struct {
	*opcodeHeader
	key   *mapKeyCode
	value *mapValueCode
	end   *opcode
}

type mapKeyCode struct {
	*opcodeHeader
	idx  int
	len  int
	iter unsafe.Pointer
	end  *opcode
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

func (c *mapValueCode) set(iter unsafe.Pointer) {
	c.iter = iter
}

func newMapHeaderCode(typ *rtype) *mapHeaderCode {
	return &mapHeaderCode{
		opcodeHeader: &opcodeHeader{
			op:  opMapHead,
			typ: typ,
		},
	}
}

func newMapKeyCode() *mapKeyCode {
	return &mapKeyCode{
		opcodeHeader: &opcodeHeader{
			op: opMapKey,
		},
	}
}

func newMapValueCode() *mapValueCode {
	return &mapValueCode{
		opcodeHeader: &opcodeHeader{
			op: opMapValue,
		},
	}
}

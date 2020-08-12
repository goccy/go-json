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
	opMarshalJSON
	opMarshalText

	opSliceHead
	opSliceElem
	opSliceEnd

	opSliceHeadIndent
	opRootSliceHeadIndent
	opSliceElemIndent
	opRootSliceElemIndent
	opSliceEndIndent

	opArrayHead
	opArrayElem
	opArrayEnd

	opArrayHeadIndent
	opArrayElemIndent
	opArrayEndIndent

	opMapHead
	opMapHeadLoad
	opMapKey
	opMapValue

	opMapHeadIndent
	opRootMapHeadIndent
	opMapHeadLoadIndent
	opMapKeyIndent
	opRootMapKeyIndent
	opMapValueIndent
	opMapEnd
	opMapEndIndent

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

	opStructFieldHeadIndent
	opStructFieldHeadIntIndent
	opStructFieldHeadInt8Indent
	opStructFieldHeadInt16Indent
	opStructFieldHeadInt32Indent
	opStructFieldHeadInt64Indent
	opStructFieldHeadUintIndent
	opStructFieldHeadUint8Indent
	opStructFieldHeadUint16Indent
	opStructFieldHeadUint32Indent
	opStructFieldHeadUint64Indent
	opStructFieldHeadFloat32Indent
	opStructFieldHeadFloat64Indent
	opStructFieldHeadStringIndent
	opStructFieldHeadBoolIndent

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

	opStructFieldHeadOmitEmptyIndent
	opStructFieldHeadIntOmitEmptyIndent
	opStructFieldHeadInt8OmitEmptyIndent
	opStructFieldHeadInt16OmitEmptyIndent
	opStructFieldHeadInt32OmitEmptyIndent
	opStructFieldHeadInt64OmitEmptyIndent
	opStructFieldHeadUintOmitEmptyIndent
	opStructFieldHeadUint8OmitEmptyIndent
	opStructFieldHeadUint16OmitEmptyIndent
	opStructFieldHeadUint32OmitEmptyIndent
	opStructFieldHeadUint64OmitEmptyIndent
	opStructFieldHeadFloat32OmitEmptyIndent
	opStructFieldHeadFloat64OmitEmptyIndent
	opStructFieldHeadStringOmitEmptyIndent
	opStructFieldHeadBoolOmitEmptyIndent

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

	opStructFieldPtrHeadIndent
	opStructFieldPtrHeadIntIndent
	opStructFieldPtrHeadInt8Indent
	opStructFieldPtrHeadInt16Indent
	opStructFieldPtrHeadInt32Indent
	opStructFieldPtrHeadInt64Indent
	opStructFieldPtrHeadUintIndent
	opStructFieldPtrHeadUint8Indent
	opStructFieldPtrHeadUint16Indent
	opStructFieldPtrHeadUint32Indent
	opStructFieldPtrHeadUint64Indent
	opStructFieldPtrHeadFloat32Indent
	opStructFieldPtrHeadFloat64Indent
	opStructFieldPtrHeadStringIndent
	opStructFieldPtrHeadBoolIndent

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

	opStructFieldPtrHeadOmitEmptyIndent
	opStructFieldPtrHeadIntOmitEmptyIndent
	opStructFieldPtrHeadInt8OmitEmptyIndent
	opStructFieldPtrHeadInt16OmitEmptyIndent
	opStructFieldPtrHeadInt32OmitEmptyIndent
	opStructFieldPtrHeadInt64OmitEmptyIndent
	opStructFieldPtrHeadUintOmitEmptyIndent
	opStructFieldPtrHeadUint8OmitEmptyIndent
	opStructFieldPtrHeadUint16OmitEmptyIndent
	opStructFieldPtrHeadUint32OmitEmptyIndent
	opStructFieldPtrHeadUint64OmitEmptyIndent
	opStructFieldPtrHeadFloat32OmitEmptyIndent
	opStructFieldPtrHeadFloat64OmitEmptyIndent
	opStructFieldPtrHeadStringOmitEmptyIndent
	opStructFieldPtrHeadBoolOmitEmptyIndent

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

	opStructFieldIndent
	opStructFieldIntIndent
	opStructFieldInt8Indent
	opStructFieldInt16Indent
	opStructFieldInt32Indent
	opStructFieldInt64Indent
	opStructFieldUintIndent
	opStructFieldUint8Indent
	opStructFieldUint16Indent
	opStructFieldUint32Indent
	opStructFieldUint64Indent
	opStructFieldFloat32Indent
	opStructFieldFloat64Indent
	opStructFieldStringIndent
	opStructFieldBoolIndent

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

	opStructFieldOmitEmptyIndent
	opStructFieldIntOmitEmptyIndent
	opStructFieldInt8OmitEmptyIndent
	opStructFieldInt16OmitEmptyIndent
	opStructFieldInt32OmitEmptyIndent
	opStructFieldInt64OmitEmptyIndent
	opStructFieldUintOmitEmptyIndent
	opStructFieldUint8OmitEmptyIndent
	opStructFieldUint16OmitEmptyIndent
	opStructFieldUint32OmitEmptyIndent
	opStructFieldUint64OmitEmptyIndent
	opStructFieldFloat32OmitEmptyIndent
	opStructFieldFloat64OmitEmptyIndent
	opStructFieldStringOmitEmptyIndent
	opStructFieldBoolOmitEmptyIndent

	opStructEnd
	opStructEndIndent
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
	case opMarshalJSON:
		return "MARSHAL_JSON"
	case opMarshalText:
		return "MARSHAL_TEXT"

	case opSliceHead:
		return "SLICE_HEAD"
	case opSliceElem:
		return "SLICE_ELEM"
	case opSliceEnd:
		return "SLICE_END"

	case opSliceHeadIndent:
		return "SLICE_HEAD_INDENT"
	case opRootSliceHeadIndent:
		return "ROOT_SLICE_HEAD_INDENT"
	case opSliceElemIndent:
		return "SLICE_ELEM_INDENT"
	case opRootSliceElemIndent:
		return "ROOT_SLICE_ELEM_INDENT"
	case opSliceEndIndent:
		return "SLICE_END_INDENT"

	case opArrayHead:
		return "ARRAY_HEAD"
	case opArrayElem:
		return "ARRAY_ELEM"
	case opArrayEnd:
		return "ARRAY_END"

	case opArrayHeadIndent:
		return "ARRAY_HEAD_INDENT"
	case opArrayElemIndent:
		return "ARRAY_ELEM_INDENT"
	case opArrayEndIndent:
		return "ARRAY_END_INDENT"
	case opMapHead:
		return "MAP_HEAD"
	case opMapHeadLoad:
		return "MAP_HEAD_LOAD"
	case opMapKey:
		return "MAP_KEY"
	case opMapValue:
		return "MAP_VALUE"
	case opMapEnd:
		return "MAP_END"

	case opMapHeadIndent:
		return "MAP_HEAD_INDENT"
	case opRootMapHeadIndent:
		return "ROOT_MAP_HEAD_INDENT"
	case opMapHeadLoadIndent:
		return "MAP_HEAD_LOAD_INDENT"
	case opMapKeyIndent:
		return "MAP_KEY_INDENT"
	case opRootMapKeyIndent:
		return "ROOT_MAP_KEY_INDENT"
	case opMapValueIndent:
		return "MAP_VALUE_INDENT"
	case opMapEndIndent:
		return "MAP_END_INDENT"

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

	case opStructFieldHeadIndent:
		return "STRUCT_FIELD_HEAD_INDENT"
	case opStructFieldHeadIntIndent:
		return "STRUCT_FIELD_HEAD_INT_INDENT"
	case opStructFieldHeadInt8Indent:
		return "STRUCT_FIELD_HEAD_INT8_INDENT"
	case opStructFieldHeadInt16Indent:
		return "STRUCT_FIELD_HEAD_INT16_INDENT"
	case opStructFieldHeadInt32Indent:
		return "STRUCT_FIELD_HEAD_INT32_INDENT"
	case opStructFieldHeadInt64Indent:
		return "STRUCT_FIELD_HEAD_INT64_INDENT"
	case opStructFieldHeadUintIndent:
		return "STRUCT_FIELD_HEAD_UINT_INDENT"
	case opStructFieldHeadUint8Indent:
		return "STRUCT_FIELD_HEAD_UINT8_INDENT"
	case opStructFieldHeadUint16Indent:
		return "STRUCT_FIELD_HEAD_UINT16_INDENT"
	case opStructFieldHeadUint32Indent:
		return "STRUCT_FIELD_HEAD_UINT32_INDENT"
	case opStructFieldHeadUint64Indent:
		return "STRUCT_FIELD_HEAD_UINT64_INDENT"
	case opStructFieldHeadFloat32Indent:
		return "STRUCT_FIELD_HEAD_FLOAT32_INDENT"
	case opStructFieldHeadFloat64Indent:
		return "STRUCT_FIELD_HEAD_FLOAT64_INDENT"
	case opStructFieldHeadStringIndent:
		return "STRUCT_FIELD_HEAD_STRING_INDENT"
	case opStructFieldHeadBoolIndent:
		return "STRUCT_FIELD_HEAD_BOOL_INDENT"

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

	case opStructFieldHeadOmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_OMIT_EMPTY_INDENT"
	case opStructFieldHeadIntOmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_INT_OMIT_EMPTY_INDENT"
	case opStructFieldHeadInt8OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_INT8_OMIT_EMPTY_INDENT"
	case opStructFieldHeadInt16OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_INT16_OMIT_EMPTY_INDENT"
	case opStructFieldHeadInt32OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_INT32_OMIT_EMPTY_INDENT"
	case opStructFieldHeadInt64OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_INT64_OMIT_EMPTY_INDENT"
	case opStructFieldHeadUintOmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_UINT_OMIT_EMPTY_INDENT"
	case opStructFieldHeadUint8OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_UINT8_OMIT_EMPTY_INDENT"
	case opStructFieldHeadUint16OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_UINT16_OMIT_EMPTY_INDENT"
	case opStructFieldHeadUint32OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_UINT32_OMIT_EMPTY_INDENT"
	case opStructFieldHeadUint64OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_UINT64_OMIT_EMPTY_INDENT"
	case opStructFieldHeadFloat32OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_FLOAT32_OMIT_EMPTY_INDENT"
	case opStructFieldHeadFloat64OmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_FLOAT64_OMIT_EMPTY_INDENT"
	case opStructFieldHeadStringOmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_STRING_OMIT_EMPTY_INDENT"
	case opStructFieldHeadBoolOmitEmptyIndent:
		return "STRUCT_FIELD_HEAD_BOOL_OMIT_EMPTY_INDENT"

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

	case opStructFieldPtrHeadIndent:
		return "STRUCT_FIELD_PTR_HEAD_INDENT"
	case opStructFieldPtrHeadIntIndent:
		return "STRUCT_FIELD_PTR_HEAD_INT_INDENT"
	case opStructFieldPtrHeadInt8Indent:
		return "STRUCT_FIELD_PTR_HEAD_INT8_INDENT"
	case opStructFieldPtrHeadInt16Indent:
		return "STRUCT_FIELD_PTR_HEAD_INT16_INDENT"
	case opStructFieldPtrHeadInt32Indent:
		return "STRUCT_FIELD_PTR_HEAD_INT32_INDENT"
	case opStructFieldPtrHeadInt64Indent:
		return "STRUCT_FIELD_PTR_HEAD_INT64_INDENT"
	case opStructFieldPtrHeadUintIndent:
		return "STRUCT_FIELD_PTR_HEAD_UINT_INDENT"
	case opStructFieldPtrHeadUint8Indent:
		return "STRUCT_FIELD_PTR_HEAD_UINT8_INDENT"
	case opStructFieldPtrHeadUint16Indent:
		return "STRUCT_FIELD_PTR_HEAD_UINT16_INDENT"
	case opStructFieldPtrHeadUint32Indent:
		return "STRUCT_FIELD_PTR_HEAD_UINT32_INDENT"
	case opStructFieldPtrHeadUint64Indent:
		return "STRUCT_FIELD_PTR_HEAD_UINT64_INDENT"
	case opStructFieldPtrHeadFloat32Indent:
		return "STRUCT_FIELD_PTR_HEAD_FLOAT32_INDENT"
	case opStructFieldPtrHeadFloat64Indent:
		return "STRUCT_FIELD_PTR_HEAD_FLOAT64_INDENT"
	case opStructFieldPtrHeadStringIndent:
		return "STRUCT_FIELD_PTR_HEAD_STRING_INDENT"
	case opStructFieldPtrHeadBoolIndent:
		return "STRUCT_FIELD_PTR_HEAD_BOOL_INDENT"

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

	case opStructFieldPtrHeadOmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadIntOmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_INT_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadInt8OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_INT8_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadInt16OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_INT16_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadInt32OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_INT32_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadInt64OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_INT64_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadUintOmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_UINT_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadUint8OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_UINT8_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadUint16OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_UINT16_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadUint32OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_UINT32_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadUint64OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_UINT64_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadFloat32OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_FLOAT32_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadFloat64OmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_FLOAT64_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadStringOmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_STRING_OMIT_EMPTY_INDENT"
	case opStructFieldPtrHeadBoolOmitEmptyIndent:
		return "STRUCT_FIELD_PTR_HEAD_BOOL_OMIT_EMPTY_INDENT"

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

	case opStructFieldIndent:
		return "STRUCT_FIELD_INDENT"
	case opStructFieldIntIndent:
		return "STRUCT_FIELD_INT_INDENT"
	case opStructFieldInt8Indent:
		return "STRUCT_FIELD_INT8_INDENT"
	case opStructFieldInt16Indent:
		return "STRUCT_FIELD_INT16_INDENT"
	case opStructFieldInt32Indent:
		return "STRUCT_FIELD_INT32_INDENT"
	case opStructFieldInt64Indent:
		return "STRUCT_FIELD_INT64_INDENT"
	case opStructFieldUintIndent:
		return "STRUCT_FIELD_UINT_INDENT"
	case opStructFieldUint8Indent:
		return "STRUCT_FIELD_UINT8_INDENT"
	case opStructFieldUint16Indent:
		return "STRUCT_FIELD_UINT16_INDENT"
	case opStructFieldUint32Indent:
		return "STRUCT_FIELD_UINT32_INDENT"
	case opStructFieldUint64Indent:
		return "STRUCT_FIELD_UINT64_INDENT"
	case opStructFieldFloat32Indent:
		return "STRUCT_FIELD_FLOAT32_INDENT"
	case opStructFieldFloat64Indent:
		return "STRUCT_FIELD_FLOAT64_INDENT"
	case opStructFieldStringIndent:
		return "STRUCT_FIELD_STRING_INDENT"
	case opStructFieldBoolIndent:
		return "STRUCT_FIELD_BOOL_INDENT"

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

	case opStructFieldOmitEmptyIndent:
		return "STRUCT_FIELD_OMIT_EMPTY_INDENT"
	case opStructFieldIntOmitEmptyIndent:
		return "STRUCT_FIELD_INT_OMIT_EMPTY_INDENT"
	case opStructFieldInt8OmitEmptyIndent:
		return "STRUCT_FIELD_INT8_OMIT_EMPTY_INDENT"
	case opStructFieldInt16OmitEmptyIndent:
		return "STRUCT_FIELD_INT16_OMIT_EMPTY_INDENT"
	case opStructFieldInt32OmitEmptyIndent:
		return "STRUCT_FIELD_INT32_OMIT_EMPTY_INDENT"
	case opStructFieldInt64OmitEmptyIndent:
		return "STRUCT_FIELD_INT64_OMIT_EMPTY_INDENT"
	case opStructFieldUintOmitEmptyIndent:
		return "STRUCT_FIELD_UINT_OMIT_EMPTY_INDENT"
	case opStructFieldUint8OmitEmptyIndent:
		return "STRUCT_FIELD_UINT8_OMIT_EMPTY_INDENT"
	case opStructFieldUint16OmitEmptyIndent:
		return "STRUCT_FIELD_UINT16_OMIT_EMPTY_INDENT"
	case opStructFieldUint32OmitEmptyIndent:
		return "STRUCT_FIELD_UINT32_OMIT_EMPTY_INDENT"
	case opStructFieldUint64OmitEmptyIndent:
		return "STRUCT_FIELD_UINT64_OMIT_EMPTY_INDENT"
	case opStructFieldFloat32OmitEmptyIndent:
		return "STRUCT_FIELD_FLOAT32_OMIT_EMPTY_INDENT"
	case opStructFieldFloat64OmitEmptyIndent:
		return "STRUCT_FIELD_FLOAT64_OMIT_EMPTY_INDENT"
	case opStructFieldStringOmitEmptyIndent:
		return "STRUCT_FIELD_STRING_OMIT_EMPTY_INDENT"
	case opStructFieldBoolOmitEmptyIndent:
		return "STRUCT_FIELD_BOOL_OMIT_EMPTY_INDENT"

	case opStructEnd:
		return "STRUCT_END"
	case opStructEndIndent:
		return "STRUCT_END_INDENT"

	}
	return ""
}

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
		switch code.op {
		case opArrayElem, opArrayElemIndent:
			nextCode = code.toArrayElemCode().end
		case opSliceElem, opSliceElemIndent, opRootSliceElemIndent:
			nextCode = code.toSliceElemCode().end
		case opMapKey, opMapKeyIndent, opRootMapKeyIndent:
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
	switch c.op {
	case opArrayHead, opArrayHeadIndent:
		code = c.toArrayHeaderCode().copy(codeMap)
	case opArrayElem, opArrayElemIndent:
		code = c.toArrayElemCode().copy(codeMap)
	case opSliceHead, opSliceHeadIndent, opRootSliceHeadIndent:
		code = c.toSliceHeaderCode().copy(codeMap)
	case opSliceElem, opSliceElemIndent, opRootSliceElemIndent:
		code = c.toSliceElemCode().copy(codeMap)
	case opMapHead, opMapHeadLoad, opMapHeadIndent, opMapHeadLoadIndent, opRootMapHeadIndent:
		code = c.toMapHeadCode().copy(codeMap)
	case opMapKey, opMapKeyIndent, opRootMapKeyIndent:
		code = c.toMapKeyCode().copy(codeMap)
	case opMapValue, opMapValueIndent:
		code = c.toMapValueCode().copy(codeMap)
	case opStructFieldHead,
		opStructFieldHeadInt,
		opStructFieldHeadInt8,
		opStructFieldHeadInt16,
		opStructFieldHeadInt32,
		opStructFieldHeadInt64,
		opStructFieldHeadUint,
		opStructFieldHeadUint8,
		opStructFieldHeadUint16,
		opStructFieldHeadUint32,
		opStructFieldHeadUint64,
		opStructFieldHeadFloat32,
		opStructFieldHeadFloat64,
		opStructFieldHeadString,
		opStructFieldHeadBool,
		opStructFieldHeadIndent,
		opStructFieldHeadIntIndent,
		opStructFieldHeadInt8Indent,
		opStructFieldHeadInt16Indent,
		opStructFieldHeadInt32Indent,
		opStructFieldHeadInt64Indent,
		opStructFieldHeadUintIndent,
		opStructFieldHeadUint8Indent,
		opStructFieldHeadUint16Indent,
		opStructFieldHeadUint32Indent,
		opStructFieldHeadUint64Indent,
		opStructFieldHeadFloat32Indent,
		opStructFieldHeadFloat64Indent,
		opStructFieldHeadStringIndent,
		opStructFieldHeadBoolIndent,
		opStructFieldHeadOmitEmpty,
		opStructFieldHeadIntOmitEmpty,
		opStructFieldHeadInt8OmitEmpty,
		opStructFieldHeadInt16OmitEmpty,
		opStructFieldHeadInt32OmitEmpty,
		opStructFieldHeadInt64OmitEmpty,
		opStructFieldHeadUintOmitEmpty,
		opStructFieldHeadUint8OmitEmpty,
		opStructFieldHeadUint16OmitEmpty,
		opStructFieldHeadUint32OmitEmpty,
		opStructFieldHeadUint64OmitEmpty,
		opStructFieldHeadFloat32OmitEmpty,
		opStructFieldHeadFloat64OmitEmpty,
		opStructFieldHeadStringOmitEmpty,
		opStructFieldHeadBoolOmitEmpty,
		opStructFieldHeadOmitEmptyIndent,
		opStructFieldHeadIntOmitEmptyIndent,
		opStructFieldHeadInt8OmitEmptyIndent,
		opStructFieldHeadInt16OmitEmptyIndent,
		opStructFieldHeadInt32OmitEmptyIndent,
		opStructFieldHeadInt64OmitEmptyIndent,
		opStructFieldHeadUintOmitEmptyIndent,
		opStructFieldHeadUint8OmitEmptyIndent,
		opStructFieldHeadUint16OmitEmptyIndent,
		opStructFieldHeadUint32OmitEmptyIndent,
		opStructFieldHeadUint64OmitEmptyIndent,
		opStructFieldHeadFloat32OmitEmptyIndent,
		opStructFieldHeadFloat64OmitEmptyIndent,
		opStructFieldHeadStringOmitEmptyIndent,
		opStructFieldHeadBoolOmitEmptyIndent,
		opStructFieldPtrHead,
		opStructFieldPtrHeadInt,
		opStructFieldPtrHeadInt8,
		opStructFieldPtrHeadInt16,
		opStructFieldPtrHeadInt32,
		opStructFieldPtrHeadInt64,
		opStructFieldPtrHeadUint,
		opStructFieldPtrHeadUint8,
		opStructFieldPtrHeadUint16,
		opStructFieldPtrHeadUint32,
		opStructFieldPtrHeadUint64,
		opStructFieldPtrHeadFloat32,
		opStructFieldPtrHeadFloat64,
		opStructFieldPtrHeadString,
		opStructFieldPtrHeadBool,
		opStructFieldPtrHeadIndent,
		opStructFieldPtrHeadIntIndent,
		opStructFieldPtrHeadInt8Indent,
		opStructFieldPtrHeadInt16Indent,
		opStructFieldPtrHeadInt32Indent,
		opStructFieldPtrHeadInt64Indent,
		opStructFieldPtrHeadUintIndent,
		opStructFieldPtrHeadUint8Indent,
		opStructFieldPtrHeadUint16Indent,
		opStructFieldPtrHeadUint32Indent,
		opStructFieldPtrHeadUint64Indent,
		opStructFieldPtrHeadFloat32Indent,
		opStructFieldPtrHeadFloat64Indent,
		opStructFieldPtrHeadStringIndent,
		opStructFieldPtrHeadBoolIndent,
		opStructFieldPtrHeadOmitEmpty,
		opStructFieldPtrHeadIntOmitEmpty,
		opStructFieldPtrHeadInt8OmitEmpty,
		opStructFieldPtrHeadInt16OmitEmpty,
		opStructFieldPtrHeadInt32OmitEmpty,
		opStructFieldPtrHeadInt64OmitEmpty,
		opStructFieldPtrHeadUintOmitEmpty,
		opStructFieldPtrHeadUint8OmitEmpty,
		opStructFieldPtrHeadUint16OmitEmpty,
		opStructFieldPtrHeadUint32OmitEmpty,
		opStructFieldPtrHeadUint64OmitEmpty,
		opStructFieldPtrHeadFloat32OmitEmpty,
		opStructFieldPtrHeadFloat64OmitEmpty,
		opStructFieldPtrHeadStringOmitEmpty,
		opStructFieldPtrHeadBoolOmitEmpty,
		opStructFieldPtrHeadOmitEmptyIndent,
		opStructFieldPtrHeadIntOmitEmptyIndent,
		opStructFieldPtrHeadInt8OmitEmptyIndent,
		opStructFieldPtrHeadInt16OmitEmptyIndent,
		opStructFieldPtrHeadInt32OmitEmptyIndent,
		opStructFieldPtrHeadInt64OmitEmptyIndent,
		opStructFieldPtrHeadUintOmitEmptyIndent,
		opStructFieldPtrHeadUint8OmitEmptyIndent,
		opStructFieldPtrHeadUint16OmitEmptyIndent,
		opStructFieldPtrHeadUint32OmitEmptyIndent,
		opStructFieldPtrHeadUint64OmitEmptyIndent,
		opStructFieldPtrHeadFloat32OmitEmptyIndent,
		opStructFieldPtrHeadFloat64OmitEmptyIndent,
		opStructFieldPtrHeadStringOmitEmptyIndent,
		opStructFieldPtrHeadBoolOmitEmptyIndent,
		opStructField,
		opStructFieldInt,
		opStructFieldInt8,
		opStructFieldInt16,
		opStructFieldInt32,
		opStructFieldInt64,
		opStructFieldUint,
		opStructFieldUint8,
		opStructFieldUint16,
		opStructFieldUint32,
		opStructFieldUint64,
		opStructFieldFloat32,
		opStructFieldFloat64,
		opStructFieldString,
		opStructFieldBool,
		opStructFieldIndent,
		opStructFieldIntIndent,
		opStructFieldInt8Indent,
		opStructFieldInt16Indent,
		opStructFieldInt32Indent,
		opStructFieldInt64Indent,
		opStructFieldUintIndent,
		opStructFieldUint8Indent,
		opStructFieldUint16Indent,
		opStructFieldUint32Indent,
		opStructFieldUint64Indent,
		opStructFieldFloat32Indent,
		opStructFieldFloat64Indent,
		opStructFieldStringIndent,
		opStructFieldBoolIndent,
		opStructFieldOmitEmpty,
		opStructFieldIntOmitEmpty,
		opStructFieldInt8OmitEmpty,
		opStructFieldInt16OmitEmpty,
		opStructFieldInt32OmitEmpty,
		opStructFieldInt64OmitEmpty,
		opStructFieldUintOmitEmpty,
		opStructFieldUint8OmitEmpty,
		opStructFieldUint16OmitEmpty,
		opStructFieldUint32OmitEmpty,
		opStructFieldUint64OmitEmpty,
		opStructFieldFloat32OmitEmpty,
		opStructFieldFloat64OmitEmpty,
		opStructFieldStringOmitEmpty,
		opStructFieldBoolOmitEmpty,
		opStructFieldOmitEmptyIndent,
		opStructFieldIntOmitEmptyIndent,
		opStructFieldInt8OmitEmptyIndent,
		opStructFieldInt16OmitEmptyIndent,
		opStructFieldInt32OmitEmptyIndent,
		opStructFieldInt64OmitEmptyIndent,
		opStructFieldUintOmitEmptyIndent,
		opStructFieldUint8OmitEmptyIndent,
		opStructFieldUint16OmitEmptyIndent,
		opStructFieldUint32OmitEmptyIndent,
		opStructFieldUint64OmitEmptyIndent,
		opStructFieldFloat32OmitEmptyIndent,
		opStructFieldFloat64OmitEmptyIndent,
		opStructFieldStringOmitEmptyIndent,
		opStructFieldBoolOmitEmptyIndent:
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
		codes = append(codes, fmt.Sprintf("%s%s", indent, code.op))
		switch code.op {
		case opArrayElem, opArrayElemIndent:
			code = code.toArrayElemCode().end
		case opSliceElem, opSliceElemIndent, opRootSliceElemIndent:
			code = code.toSliceElemCode().end
		case opMapKey, opMapKeyIndent, opRootMapKeyIndent:
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
	key       []byte
	offset    uintptr
	nextField *opcode
	end       *opcode
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
		key:    c.key,
		offset: c.offset,
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

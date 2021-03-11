package json

import (
	"encoding"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unsafe"
)

type compiledCode struct {
	code    *opcode
	linked  bool // whether recursive code already have linked
	curLen  uintptr
	nextLen uintptr
}

type opcodeSet struct {
	code       *opcode
	codeLength int
}

var (
	marshalJSONType = reflect.TypeOf((*Marshaler)(nil)).Elem()
	marshalTextType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	jsonNumberType  = reflect.TypeOf(Number(""))
)

func encodeCompileToGetCodeSetSlowPath(typeptr uintptr) (*opcodeSet, error) {
	opcodeMap := loadOpcodeMap()
	if codeSet, exists := opcodeMap[typeptr]; exists {
		return codeSet, nil
	}

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := *(**rtype)(unsafe.Pointer(&typeptr))

	code, err := encodeCompileHead(&encodeCompileContext{
		typ:                      copiedType,
		structTypeToCompiledCode: map[uintptr]*compiledCode{},
	})
	if err != nil {
		return nil, err
	}
	code = copyOpcode(code)
	codeLength := code.totalLength()
	codeSet := &opcodeSet{
		code:       code,
		codeLength: codeLength,
	}
	storeOpcodeSet(typeptr, codeSet, opcodeMap)
	return codeSet, nil
}

func encodeCompileHead(ctx *encodeCompileContext) (*opcode, error) {
	typ := ctx.typ
	switch {
	case encodeImplementsMarshalJSON(typ):
		return encodeCompileMarshalJSON(ctx)
	case encodeImplementsMarshalText(typ):
		return encodeCompileMarshalText(ctx)
	}

	isPtr := false
	orgType := typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		isPtr = true
	}
	switch {
	case encodeImplementsMarshalJSON(typ):
		return encodeCompileMarshalJSON(ctx)
	case encodeImplementsMarshalText(typ):
		return encodeCompileMarshalText(ctx)
	}
	if typ.Kind() == reflect.Map {
		if isPtr {
			return encodeCompilePtr(ctx.withType(rtype_ptrTo(typ)))
		}
		return encodeCompileMap(ctx.withType(typ))
	} else if typ.Kind() == reflect.Struct {
		code, err := encodeCompileStruct(ctx.withType(typ), isPtr)
		if err != nil {
			return nil, err
		}
		encodeOptimizeStructEnd(code)
		encodeLinkRecursiveCode(code)
		return code, nil
	} else if isPtr && typ.Implements(marshalTextType) {
		typ = orgType
	}
	code, err := encodeCompile(ctx.withType(typ), isPtr)
	if err != nil {
		return nil, err
	}
	encodeOptimizeStructEnd(code)
	encodeLinkRecursiveCode(code)
	return code, nil
}

func encodeLinkRecursiveCode(c *opcode) {
	for code := c; code.op != opEnd && code.op != opStructFieldRecursiveEnd; {
		switch code.op {
		case opStructFieldRecursive, opStructFieldRecursivePtr:
			if code.jmp.linked {
				code = code.next
				continue
			}
			code.jmp.code = copyOpcode(code.jmp.code)
			c := code.jmp.code
			c.end.next = newEndOp(&encodeCompileContext{})
			c.op = c.op.ptrHeadToHead()

			beforeLastCode := c.end
			lastCode := beforeLastCode.next

			lastCode.idx = beforeLastCode.idx + uintptrSize
			lastCode.elemIdx = lastCode.idx + uintptrSize

			// extend length to alloc slot for elemIdx
			totalLength := uintptr(code.totalLength() + 1)
			nextTotalLength := uintptr(c.totalLength() + 1)

			c.end.next.op = opStructFieldRecursiveEnd

			code.jmp.curLen = totalLength
			code.jmp.nextLen = nextTotalLength
			code.jmp.linked = true

			encodeLinkRecursiveCode(code.jmp.code)
			code = code.next
			continue
		}
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem, codeMapKey:
			code = code.end
		default:
			code = code.next
		}
	}
}

func encodeOptimizeStructEnd(c *opcode) {
	for code := c; code.op != opEnd; {
		if code.op == opStructFieldRecursive || code.op == opStructFieldRecursivePtr {
			// ignore if exists recursive operation
			return
		}
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem, codeMapKey:
			code = code.end
		default:
			code = code.next
		}
	}

	for code := c; code.op != opEnd; {
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem, codeMapKey:
			code = code.end
		case codeStructEnd:
			switch code.op {
			case opStructEnd:
				prev := code.prevField
				prevOp := prev.op.String()
				if strings.Contains(prevOp, "Head") ||
					strings.Contains(prevOp, "Slice") ||
					strings.Contains(prevOp, "Array") ||
					strings.Contains(prevOp, "Map") ||
					strings.Contains(prevOp, "MarshalJSON") ||
					strings.Contains(prevOp, "MarshalText") {
					// not exists field
					code = code.next
					break
				}
				if prev.op != prev.op.fieldToEnd() {
					prev.op = prev.op.fieldToEnd()
					prev.next = code.next
				}
				code = code.next
			default:
				code = code.next
			}
		default:
			code = code.next
		}
	}
}

func encodeImplementsMarshalJSON(typ *rtype) bool {
	if !typ.Implements(marshalJSONType) {
		return false
	}
	if typ.Kind() != reflect.Ptr {
		return true
	}
	// type kind is reflect.Ptr
	if !typ.Elem().Implements(marshalJSONType) {
		return true
	}
	// needs to dereference
	return false
}

func encodeImplementsMarshalText(typ *rtype) bool {
	if !typ.Implements(marshalTextType) {
		return false
	}
	if typ.Kind() != reflect.Ptr {
		return true
	}
	// type kind is reflect.Ptr
	if !typ.Elem().Implements(marshalTextType) {
		return true
	}
	// needs to dereference
	return false
}

func encodeCompile(ctx *encodeCompileContext, isPtr bool) (*opcode, error) {
	typ := ctx.typ
	switch {
	case encodeImplementsMarshalJSON(typ):
		return encodeCompileMarshalJSON(ctx)
	case encodeImplementsMarshalText(typ):
		return encodeCompileMarshalText(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return encodeCompilePtr(ctx)
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := rtype_ptrTo(elem)
			if !p.Implements(marshalJSONType) && !p.Implements(marshalTextType) {
				return encodeCompileBytes(ctx)
			}
		}
		return encodeCompileSlice(ctx)
	case reflect.Array:
		return encodeCompileArray(ctx)
	case reflect.Map:
		return encodeCompileMap(ctx)
	case reflect.Struct:
		return encodeCompileStruct(ctx, isPtr)
	case reflect.Interface:
		return encodeCompileInterface(ctx)
	case reflect.Int:
		return encodeCompileInt(ctx)
	case reflect.Int8:
		return encodeCompileInt8(ctx)
	case reflect.Int16:
		return encodeCompileInt16(ctx)
	case reflect.Int32:
		return encodeCompileInt32(ctx)
	case reflect.Int64:
		return encodeCompileInt64(ctx)
	case reflect.Uint:
		return encodeCompileUint(ctx)
	case reflect.Uint8:
		return encodeCompileUint8(ctx)
	case reflect.Uint16:
		return encodeCompileUint16(ctx)
	case reflect.Uint32:
		return encodeCompileUint32(ctx)
	case reflect.Uint64:
		return encodeCompileUint64(ctx)
	case reflect.Uintptr:
		return encodeCompileUint(ctx)
	case reflect.Float32:
		return encodeCompileFloat32(ctx)
	case reflect.Float64:
		return encodeCompileFloat64(ctx)
	case reflect.String:
		return encodeCompileString(ctx)
	case reflect.Bool:
		return encodeCompileBool(ctx)
	}
	return nil, &UnsupportedTypeError{Type: rtype2type(typ)}
}

func encodeConvertPtrOp(code *opcode) opType {
	ptrHeadOp := code.op.headToPtrHead()
	if code.op != ptrHeadOp {
		return ptrHeadOp
	}
	switch code.op {
	case opInt:
		return opIntPtr
	case opUint:
		return opUintPtr
	case opFloat32:
		return opFloat32Ptr
	case opFloat64:
		return opFloat64Ptr
	case opString:
		return opStringPtr
	case opBool:
		return opBoolPtr
	case opBytes:
		return opBytesPtr
	case opArray:
		return opArrayPtr
	case opSlice:
		return opSlicePtr
	case opMap:
		return opMapPtr
	case opMarshalJSON:
		return opMarshalJSONPtr
	case opMarshalText:
		return opMarshalTextPtr
	case opInterface:
		return opInterfacePtr
	case opStructFieldRecursive:
		return opStructFieldRecursivePtr
	}
	return code.op
}

func encodeCompileKey(ctx *encodeCompileContext) (*opcode, error) {
	typ := ctx.typ
	switch {
	case encodeImplementsMarshalJSON(typ):
		return encodeCompileMarshalJSON(ctx)
	case encodeImplementsMarshalText(typ):
		return encodeCompileMarshalText(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return encodeCompilePtr(ctx)
	case reflect.Interface:
		return encodeCompileInterface(ctx)
	case reflect.String:
		return encodeCompileString(ctx)
	case reflect.Int:
		return encodeCompileIntString(ctx)
	case reflect.Int8:
		return encodeCompileInt8String(ctx)
	case reflect.Int16:
		return encodeCompileInt16String(ctx)
	case reflect.Int32:
		return encodeCompileInt32String(ctx)
	case reflect.Int64:
		return encodeCompileInt64String(ctx)
	case reflect.Uint:
		return encodeCompileUintString(ctx)
	case reflect.Uint8:
		return encodeCompileUint8String(ctx)
	case reflect.Uint16:
		return encodeCompileUint16String(ctx)
	case reflect.Uint32:
		return encodeCompileUint32String(ctx)
	case reflect.Uint64:
		return encodeCompileUint64String(ctx)
	case reflect.Uintptr:
		return encodeCompileUintString(ctx)
	}
	return nil, &UnsupportedTypeError{Type: rtype2type(typ)}
}

func encodeCompilePtr(ctx *encodeCompileContext) (*opcode, error) {
	code, err := encodeCompile(ctx.withType(ctx.typ.Elem()), true)
	if err != nil {
		return nil, err
	}
	code.op = encodeConvertPtrOp(code)
	code.ptrNum++
	return code, nil
}

func encodeCompileMarshalJSON(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opMarshalJSON)
	typ := ctx.typ
	if !typ.Implements(marshalJSONType) && rtype_ptrTo(typ).Implements(marshalJSONType) {
		code.addrForMarshaler = true
	}
	ctx.incIndex()
	return code, nil
}

func encodeCompileMarshalText(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opMarshalText)
	typ := ctx.typ
	if !typ.Implements(marshalTextType) && rtype_ptrTo(typ).Implements(marshalTextType) {
		code.addrForMarshaler = true
	}
	ctx.incIndex()
	return code, nil
}

const intSize = 32 << (^uint(0) >> 63)

func encodeCompileInt(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt)
	switch intSize {
	case 32:
		code.mask = math.MaxUint32
		code.rshiftNum = 31
	default:
		code.mask = math.MaxUint64
		code.rshiftNum = 63
	}
	ctx.incIndex()
	return code, nil
}

func encodeCompileInt8(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt)
	code.mask = math.MaxUint8
	code.rshiftNum = 7
	ctx.incIndex()
	return code, nil
}

func encodeCompileInt16(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt)
	code.mask = math.MaxUint16
	code.rshiftNum = 15
	ctx.incIndex()
	return code, nil
}

func encodeCompileInt32(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt)
	code.mask = math.MaxUint32
	code.rshiftNum = 31
	ctx.incIndex()
	return code, nil
}

func encodeCompileInt64(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt)
	code.mask = math.MaxUint64
	code.rshiftNum = 63
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint)
	switch intSize {
	case 32:
		code.mask = math.MaxUint32
		code.rshiftNum = 31
	default:
		code.mask = math.MaxUint64
		code.rshiftNum = 63
	}
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint8(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint)
	code.mask = math.MaxUint8
	code.rshiftNum = 7
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint16(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint)
	code.mask = math.MaxUint16
	code.rshiftNum = 15
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint32(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint)
	code.mask = math.MaxUint32
	code.rshiftNum = 31
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint64(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint)
	code.mask = math.MaxUint64
	code.rshiftNum = 63
	ctx.incIndex()
	return code, nil
}

func encodeCompileIntString(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opIntString)
	switch intSize {
	case 32:
		code.mask = math.MaxUint32
		code.rshiftNum = 31
	default:
		code.mask = math.MaxUint64
		code.rshiftNum = 63
	}
	ctx.incIndex()
	return code, nil
}

func encodeCompileInt8String(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opIntString)
	code.mask = math.MaxUint8
	code.rshiftNum = 7
	ctx.incIndex()
	return code, nil
}

func encodeCompileInt16String(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opIntString)
	code.mask = math.MaxUint16
	code.rshiftNum = 15
	ctx.incIndex()
	return code, nil
}

func encodeCompileInt32String(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opIntString)
	code.mask = math.MaxUint32
	code.rshiftNum = 31
	ctx.incIndex()
	return code, nil
}

func encodeCompileInt64String(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opIntString)
	code.mask = math.MaxUint64
	code.rshiftNum = 63
	ctx.incIndex()
	return code, nil
}

func encodeCompileUintString(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUintString)
	switch intSize {
	case 32:
		code.mask = math.MaxUint32
		code.rshiftNum = 31
	default:
		code.mask = math.MaxUint64
		code.rshiftNum = 63
	}
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint8String(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUintString)
	code.mask = math.MaxUint8
	code.rshiftNum = 7
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint16String(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUintString)
	code.mask = math.MaxUint16
	code.rshiftNum = 15
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint32String(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUintString)
	code.mask = math.MaxUint32
	code.rshiftNum = 31
	ctx.incIndex()
	return code, nil
}

func encodeCompileUint64String(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUintString)
	code.mask = math.MaxUint64
	code.rshiftNum = 63
	ctx.incIndex()
	return code, nil
}

func encodeCompileFloat32(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opFloat32)
	ctx.incIndex()
	return code, nil
}

func encodeCompileFloat64(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opFloat64)
	ctx.incIndex()
	return code, nil
}

func encodeCompileString(ctx *encodeCompileContext) (*opcode, error) {
	var op opType
	if ctx.typ == type2rtype(jsonNumberType) {
		op = opNumber
	} else {
		op = opString
	}
	code := newOpCode(ctx, op)
	ctx.incIndex()
	return code, nil
}

func encodeCompileBool(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opBool)
	ctx.incIndex()
	return code, nil
}

func encodeCompileBytes(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opBytes)
	ctx.incIndex()
	return code, nil
}

func encodeCompileInterface(ctx *encodeCompileContext) (*opcode, error) {
	code := newInterfaceCode(ctx)
	ctx.incIndex()
	return code, nil
}

func encodeCompileSlice(ctx *encodeCompileContext) (*opcode, error) {
	elem := ctx.typ.Elem()
	size := elem.Size()

	header := newSliceHeaderCode(ctx)
	ctx.incIndex()

	code, err := encodeCompileSliceElem(ctx.withType(elem).incIndent())
	if err != nil {
		return nil, err
	}

	// header => opcode => elem => end
	//             ^        |
	//             |________|

	elemCode := newSliceElemCode(ctx, header, size)
	ctx.incIndex()

	end := newOpCode(ctx, opSliceEnd)
	ctx.incIndex()

	header.elem = elemCode
	header.end = end
	header.next = code
	code.beforeLastCode().next = (*opcode)(unsafe.Pointer(elemCode))
	elemCode.next = code
	elemCode.end = end
	return (*opcode)(unsafe.Pointer(header)), nil
}

func encodeCompileSliceElem(ctx *encodeCompileContext) (*opcode, error) {
	typ := ctx.typ
	switch {
	case !typ.Implements(marshalJSONType) && rtype_ptrTo(typ).Implements(marshalJSONType):
		return encodeCompileMarshalJSON(ctx)
	case !typ.Implements(marshalTextType) && rtype_ptrTo(typ).Implements(marshalTextType):
		return encodeCompileMarshalText(ctx)
	default:
		return encodeCompile(ctx, false)
	}
}

func encodeCompileArray(ctx *encodeCompileContext) (*opcode, error) {
	typ := ctx.typ
	elem := typ.Elem()
	alen := typ.Len()
	size := elem.Size()

	header := newArrayHeaderCode(ctx, alen)
	ctx.incIndex()

	code, err := encodeCompile(ctx.withType(elem).incIndent(), false)
	if err != nil {
		return nil, err
	}
	// header => opcode => elem => end
	//             ^        |
	//             |________|

	elemCode := newArrayElemCode(ctx, header, alen, size)
	ctx.incIndex()

	end := newOpCode(ctx, opArrayEnd)
	ctx.incIndex()

	header.elem = elemCode
	header.end = end
	header.next = code
	code.beforeLastCode().next = (*opcode)(unsafe.Pointer(elemCode))
	elemCode.next = code
	elemCode.end = end
	return (*opcode)(unsafe.Pointer(header)), nil
}

//go:linkname mapiterinit reflect.mapiterinit
//go:noescape
func mapiterinit(mapType *rtype, m unsafe.Pointer) unsafe.Pointer

//go:linkname mapiterkey reflect.mapiterkey
//go:noescape
func mapiterkey(it unsafe.Pointer) unsafe.Pointer

//go:linkname mapiternext reflect.mapiternext
//go:noescape
func mapiternext(it unsafe.Pointer)

//go:linkname maplen reflect.maplen
//go:noescape
func maplen(m unsafe.Pointer) int

func encodeCompileMap(ctx *encodeCompileContext) (*opcode, error) {
	// header => code => value => code => key => code => value => code => end
	//                                     ^                       |
	//                                     |_______________________|
	ctx = ctx.incIndent()
	header := newMapHeaderCode(ctx)
	ctx.incIndex()

	typ := ctx.typ
	keyType := ctx.typ.Key()
	keyCode, err := encodeCompileKey(ctx.withType(keyType))
	if err != nil {
		return nil, err
	}

	value := newMapValueCode(ctx, header)
	ctx.incIndex()

	valueType := typ.Elem()
	valueCode, err := encodeCompile(ctx.withType(valueType), false)
	if err != nil {
		return nil, err
	}

	key := newMapKeyCode(ctx, header)
	ctx.incIndex()

	ctx = ctx.decIndent()

	header.mapKey = key
	header.mapValue = value

	end := newMapEndCode(ctx, header)
	ctx.incIndex()

	header.next = keyCode
	keyCode.beforeLastCode().next = (*opcode)(unsafe.Pointer(value))
	value.next = valueCode
	valueCode.beforeLastCode().next = (*opcode)(unsafe.Pointer(key))
	key.next = keyCode

	header.end = end
	key.end = end
	value.end = end

	return (*opcode)(unsafe.Pointer(header)), nil
}

func encodeTypeToHeaderType(code *opcode) opType {
	switch code.op {
	case opInt:
		return opStructFieldHeadInt
	case opIntPtr:
		return opStructFieldHeadIntPtr
	case opUint:
		return opStructFieldHeadUint
	case opUintPtr:
		return opStructFieldHeadUintPtr
	case opFloat32:
		return opStructFieldHeadFloat32
	case opFloat32Ptr:
		return opStructFieldHeadFloat32Ptr
	case opFloat64:
		return opStructFieldHeadFloat64
	case opFloat64Ptr:
		return opStructFieldHeadFloat64Ptr
	case opString:
		return opStructFieldHeadString
	case opStringPtr:
		return opStructFieldHeadStringPtr
	case opNumber:
		return opStructFieldHeadNumber
	case opNumberPtr:
		return opStructFieldHeadNumberPtr
	case opBool:
		return opStructFieldHeadBool
	case opBoolPtr:
		return opStructFieldHeadBoolPtr
	case opMap:
		return opStructFieldHeadMap
	case opMapPtr:
		code.op = opMap
		return opStructFieldHeadMapPtr
	case opArray:
		return opStructFieldHeadArray
	case opArrayPtr:
		code.op = opArray
		return opStructFieldHeadArrayPtr
	case opSlice:
		return opStructFieldHeadSlice
	case opSlicePtr:
		code.op = opSlice
		return opStructFieldHeadSlicePtr
	case opMarshalJSON:
		return opStructFieldHeadMarshalJSON
	case opMarshalJSONPtr:
		return opStructFieldHeadMarshalJSONPtr
	case opMarshalText:
		return opStructFieldHeadMarshalText
	case opMarshalTextPtr:
		return opStructFieldHeadMarshalTextPtr
	}
	return opStructFieldHead
}

func encodeTypeToFieldType(code *opcode) opType {
	switch code.op {
	case opInt:
		return opStructFieldInt
	case opIntPtr:
		return opStructFieldIntPtr
	case opUint:
		return opStructFieldUint
	case opUintPtr:
		return opStructFieldUintPtr
	case opFloat32:
		return opStructFieldFloat32
	case opFloat32Ptr:
		return opStructFieldFloat32Ptr
	case opFloat64:
		return opStructFieldFloat64
	case opFloat64Ptr:
		return opStructFieldFloat64Ptr
	case opString:
		return opStructFieldString
	case opStringPtr:
		return opStructFieldStringPtr
	case opNumber:
		return opStructFieldNumber
	case opNumberPtr:
		return opStructFieldNumberPtr
	case opBool:
		return opStructFieldBool
	case opBoolPtr:
		return opStructFieldBoolPtr
	case opMap:
		return opStructFieldMap
	case opMapPtr:
		code.op = opMap
		return opStructFieldMapPtr
	case opArray:
		return opStructFieldArray
	case opArrayPtr:
		code.op = opArray
		return opStructFieldArrayPtr
	case opSlice:
		return opStructFieldSlice
	case opSlicePtr:
		code.op = opSlice
		return opStructFieldSlicePtr
	case opMarshalJSON:
		return opStructFieldMarshalJSON
	case opMarshalJSONPtr:
		return opStructFieldMarshalJSONPtr
	case opMarshalText:
		return opStructFieldMarshalText
	case opMarshalTextPtr:
		return opStructFieldMarshalTextPtr
	}
	return opStructField
}

func encodeOptimizeStructHeader(code *opcode, tag *structTag) opType {
	headType := encodeTypeToHeaderType(code)
	switch {
	case tag.isOmitEmpty:
		headType = headType.headToOmitEmptyHead()
	case tag.isString:
		headType = headType.headToStringTagHead()
	}
	return headType
}

func encodeOptimizeStructField(code *opcode, tag *structTag) opType {
	fieldType := encodeTypeToFieldType(code)
	switch {
	case tag.isOmitEmpty:
		fieldType = fieldType.fieldToOmitEmptyField()
	case tag.isString:
		fieldType = fieldType.fieldToStringTagField()
	}
	return fieldType
}

func encodeRecursiveCode(ctx *encodeCompileContext, jmp *compiledCode) *opcode {
	code := newRecursiveCode(ctx, jmp)
	ctx.incIndex()
	return code
}

func encodeCompiledCode(ctx *encodeCompileContext) *opcode {
	typ := ctx.typ
	typeptr := uintptr(unsafe.Pointer(typ))
	if compiledCode, exists := ctx.structTypeToCompiledCode[typeptr]; exists {
		return encodeRecursiveCode(ctx, compiledCode)
	}
	return nil
}

func encodeStructHeader(ctx *encodeCompileContext, fieldCode *opcode, valueCode *opcode, tag *structTag) *opcode {
	fieldCode.indent--
	op := encodeOptimizeStructHeader(valueCode, tag)
	fieldCode.op = op
	fieldCode.mask = valueCode.mask
	fieldCode.rshiftNum = valueCode.rshiftNum
	fieldCode.ptrNum = valueCode.ptrNum
	switch op {
	case opStructFieldHead,
		opStructFieldHeadSlice,
		opStructFieldHeadArray,
		opStructFieldHeadMap,
		opStructFieldHeadStruct,
		opStructFieldHeadOmitEmpty,
		opStructFieldHeadOmitEmptySlice,
		opStructFieldHeadStringTagSlice,
		opStructFieldHeadOmitEmptyArray,
		opStructFieldHeadStringTagArray,
		opStructFieldHeadOmitEmptyMap,
		opStructFieldHeadStringTagMap,
		opStructFieldHeadOmitEmptyStruct,
		opStructFieldHeadStringTag:
		return valueCode.beforeLastCode()
	case opStructFieldHeadSlicePtr,
		opStructFieldHeadOmitEmptySlicePtr,
		opStructFieldHeadStringTagSlicePtr,
		opStructFieldHeadArrayPtr,
		opStructFieldHeadOmitEmptyArrayPtr,
		opStructFieldHeadStringTagArrayPtr,
		opStructFieldHeadMapPtr,
		opStructFieldHeadOmitEmptyMapPtr,
		opStructFieldHeadStringTagMapPtr:
		return valueCode.beforeLastCode()
	case opStructFieldHeadMarshalJSONPtr,
		opStructFieldHeadOmitEmptyMarshalJSONPtr,
		opStructFieldHeadStringTagMarshalJSONPtr,
		opStructFieldHeadMarshalTextPtr,
		opStructFieldHeadOmitEmptyMarshalTextPtr,
		opStructFieldHeadStringTagMarshalTextPtr:
		ctx.decOpcodeIndex()
		return (*opcode)(unsafe.Pointer(fieldCode))
	}
	ctx.decOpcodeIndex()
	return (*opcode)(unsafe.Pointer(fieldCode))
}

func encodeStructField(ctx *encodeCompileContext, fieldCode *opcode, valueCode *opcode, tag *structTag) *opcode {
	code := (*opcode)(unsafe.Pointer(fieldCode))
	op := encodeOptimizeStructField(valueCode, tag)
	fieldCode.op = op
	fieldCode.ptrNum = valueCode.ptrNum
	fieldCode.mask = valueCode.mask
	fieldCode.rshiftNum = valueCode.rshiftNum
	fieldCode.jmp = valueCode.jmp
	switch op {
	case opStructField,
		opStructFieldSlice,
		opStructFieldArray,
		opStructFieldMap,
		opStructFieldStruct,
		opStructFieldOmitEmpty,
		opStructFieldOmitEmptySlice,
		opStructFieldStringTagSlice,
		opStructFieldOmitEmptyArray,
		opStructFieldStringTagArray,
		opStructFieldOmitEmptyMap,
		opStructFieldStringTagMap,
		opStructFieldOmitEmptyStruct,
		opStructFieldStringTag:
		return valueCode.beforeLastCode()
	case opStructFieldSlicePtr,
		opStructFieldOmitEmptySlicePtr,
		opStructFieldStringTagSlicePtr,
		opStructFieldArrayPtr,
		opStructFieldOmitEmptyArrayPtr,
		opStructFieldStringTagArrayPtr,
		opStructFieldMapPtr,
		opStructFieldOmitEmptyMapPtr,
		opStructFieldStringTagMapPtr:
		return valueCode.beforeLastCode()
	}
	ctx.decIndex()
	return code
}

func encodeIsNotExistsField(head *opcode) bool {
	if head == nil {
		return false
	}
	if head.op != opStructFieldHead {
		return false
	}
	if !head.anonymousHead {
		return false
	}
	if head.next == nil {
		return false
	}
	if head.nextField == nil {
		return false
	}
	if head.nextField.op != opStructAnonymousEnd {
		return false
	}
	if head.next.op == opStructAnonymousEnd {
		return true
	}
	if head.next.op.codeType() != codeStructField {
		return false
	}
	return encodeIsNotExistsField(head.next)
}

func encodeOptimizeAnonymousFields(head *opcode) {
	code := head
	var prev *opcode
	removedFields := map[*opcode]struct{}{}
	for {
		if code.op == opStructEnd {
			break
		}
		if code.op == opStructField {
			codeType := code.next.op.codeType()
			if codeType == codeStructField {
				if encodeIsNotExistsField(code.next) {
					code.next = code.nextField
					diff := code.next.displayIdx - code.displayIdx
					for i := 0; i < diff; i++ {
						code.next.decOpcodeIndex()
					}
					encodeLinkPrevToNextField(code, removedFields)
					code = prev
				}
			}
		}
		prev = code
		code = code.nextField
	}
}

type structFieldPair struct {
	prevField   *opcode
	curField    *opcode
	isTaggedKey bool
	linked      bool
}

func encodeAnonymousStructFieldPairMap(tags structTags, named string, valueCode *opcode) map[string][]structFieldPair {
	anonymousFields := map[string][]structFieldPair{}
	f := valueCode
	var prevAnonymousField *opcode
	removedFields := map[*opcode]struct{}{}
	for {
		existsKey := tags.existsKey(f.displayKey)
		isHeadOp := strings.Contains(f.op.String(), "Head")
		if existsKey && strings.Contains(f.op.String(), "Recursive") {
			// through
		} else if isHeadOp && !f.anonymousHead {
			if existsKey {
				// TODO: need to remove this head
				f.op = opStructFieldHead
				f.anonymousKey = true
				f.anonymousHead = true
			} else if named == "" {
				f.anonymousHead = true
			}
		} else if named == "" && f.op == opStructEnd {
			f.op = opStructAnonymousEnd
		} else if existsKey {
			diff := f.nextField.displayIdx - f.displayIdx
			for i := 0; i < diff; i++ {
				f.nextField.decOpcodeIndex()
			}
			encodeLinkPrevToNextField(f, removedFields)
		}

		if f.displayKey == "" {
			if f.nextField == nil {
				break
			}
			prevAnonymousField = f
			f = f.nextField
			continue
		}

		key := fmt.Sprintf("%s.%s", named, f.displayKey)
		anonymousFields[key] = append(anonymousFields[key], structFieldPair{
			prevField:   prevAnonymousField,
			curField:    f,
			isTaggedKey: f.isTaggedKey,
		})
		if f.next != nil && f.nextField != f.next && f.next.op.codeType() == codeStructField {
			for k, v := range encodeAnonymousFieldPairRecursively(named, f.next) {
				anonymousFields[k] = append(anonymousFields[k], v...)
			}
		}
		if f.nextField == nil {
			break
		}
		prevAnonymousField = f
		f = f.nextField
	}
	return anonymousFields
}

func encodeAnonymousFieldPairRecursively(named string, valueCode *opcode) map[string][]structFieldPair {
	anonymousFields := map[string][]structFieldPair{}
	f := valueCode
	var prevAnonymousField *opcode
	for {
		if f.displayKey != "" && f.anonymousHead {
			key := fmt.Sprintf("%s.%s", named, f.displayKey)
			anonymousFields[key] = append(anonymousFields[key], structFieldPair{
				prevField:   prevAnonymousField,
				curField:    f,
				isTaggedKey: f.isTaggedKey,
			})
			if f.next != nil && f.nextField != f.next && f.next.op.codeType() == codeStructField {
				for k, v := range encodeAnonymousFieldPairRecursively(named, f.next) {
					anonymousFields[k] = append(anonymousFields[k], v...)
				}
			}
		}
		if f.nextField == nil {
			break
		}
		prevAnonymousField = f
		f = f.nextField
	}
	return anonymousFields
}

func encodeOptimizeConflictAnonymousFields(anonymousFields map[string][]structFieldPair) {
	removedFields := map[*opcode]struct{}{}
	for _, fieldPairs := range anonymousFields {
		if len(fieldPairs) == 1 {
			continue
		}
		// conflict anonymous fields
		taggedPairs := []structFieldPair{}
		for _, fieldPair := range fieldPairs {
			if fieldPair.isTaggedKey {
				taggedPairs = append(taggedPairs, fieldPair)
			} else {
				if !fieldPair.linked {
					if fieldPair.prevField == nil {
						// head operation
						fieldPair.curField.op = opStructFieldHead
						fieldPair.curField.anonymousHead = true
						fieldPair.curField.anonymousKey = true
					} else {
						diff := fieldPair.curField.nextField.displayIdx - fieldPair.curField.displayIdx
						for i := 0; i < diff; i++ {
							fieldPair.curField.nextField.decOpcodeIndex()
						}
						removedFields[fieldPair.curField] = struct{}{}
						encodeLinkPrevToNextField(fieldPair.curField, removedFields)
					}
					fieldPair.linked = true
				}
			}
		}
		if len(taggedPairs) > 1 {
			for _, fieldPair := range taggedPairs {
				if !fieldPair.linked {
					if fieldPair.prevField == nil {
						// head operation
						fieldPair.curField.op = opStructFieldHead
						fieldPair.curField.anonymousHead = true
						fieldPair.curField.anonymousKey = true
					} else {
						diff := fieldPair.curField.nextField.displayIdx - fieldPair.curField.displayIdx
						removedFields[fieldPair.curField] = struct{}{}
						for i := 0; i < diff; i++ {
							fieldPair.curField.nextField.decOpcodeIndex()
						}
						encodeLinkPrevToNextField(fieldPair.curField, removedFields)
					}
					fieldPair.linked = true
				}
			}
		} else {
			for _, fieldPair := range taggedPairs {
				fieldPair.curField.isTaggedKey = false
			}
		}
	}
}

func encodeIsNilableType(typ *rtype) bool {
	switch typ.Kind() {
	case reflect.Ptr:
		return true
	case reflect.Interface:
		return true
	case reflect.Slice:
		return true
	case reflect.Map:
		return true
	default:
		return false
	}
}

func encodeCompileStruct(ctx *encodeCompileContext, isPtr bool) (*opcode, error) {
	if code := encodeCompiledCode(ctx); code != nil {
		return code, nil
	}
	typ := ctx.typ
	typeptr := uintptr(unsafe.Pointer(typ))
	compiled := &compiledCode{}
	ctx.structTypeToCompiledCode[typeptr] = compiled
	// header => code => structField => code => end
	//                        ^          |
	//                        |__________|
	fieldNum := typ.NumField()
	indirect := ifaceIndir(typ)
	fieldIdx := 0
	disableIndirectConversion := false
	var (
		head      *opcode
		code      *opcode
		prevField *opcode
	)
	ctx = ctx.incIndent()
	tags := structTags{}
	anonymousFields := map[string][]structFieldPair{}
	for i := 0; i < fieldNum; i++ {
		field := typ.Field(i)
		if isIgnoredStructField(field) {
			continue
		}
		tags = append(tags, structTagFromField(field))
	}
	for i, tag := range tags {
		field := tag.field
		fieldType := type2rtype(field.Type)
		fieldOpcodeIndex := ctx.opcodeIndex
		fieldPtrIndex := ctx.ptrIndex
		ctx.incIndex()

		nilcheck := true
		addrForMarshaler := false
		isIndirectSpecialCase := isPtr && i == 0 && fieldNum == 1
		isNilableType := encodeIsNilableType(fieldType)

		var valueCode *opcode
		switch {
		case isIndirectSpecialCase && !isNilableType && encodeIsPtrMarshalJSONType(fieldType):
			// *struct{ field T } => struct { field *T }
			// func (*T) MarshalJSON() ([]byte, error)
			// move pointer position from head to first field
			code, err := encodeCompileMarshalJSON(ctx.withType(rtype_ptrTo(fieldType)))
			if err != nil {
				return nil, err
			}
			valueCode = code
			nilcheck = false
			indirect = false
			disableIndirectConversion = true
		case isIndirectSpecialCase && !isNilableType && encodeIsPtrMarshalTextType(fieldType):
			// *struct{ field T } => struct { field *T }
			// func (*T) MarshalText() ([]byte, error)
			// move pointer position from head to first field
			code, err := encodeCompileMarshalText(ctx.withType(rtype_ptrTo(fieldType)))
			if err != nil {
				return nil, err
			}
			valueCode = code
			nilcheck = false
			indirect = false
			disableIndirectConversion = true
		case isPtr && encodeIsPtrMarshalJSONType(fieldType):
			// *struct{ field T }
			// func (*T) MarshalJSON() ([]byte, error)
			code, err := encodeCompileMarshalJSON(ctx.withType(fieldType))
			if err != nil {
				return nil, err
			}
			addrForMarshaler = true
			nilcheck = false
			valueCode = code
		case isPtr && encodeIsPtrMarshalTextType(fieldType):
			// *struct{ field T }
			// func (*T) MarshalText() ([]byte, error)
			code, err := encodeCompileMarshalText(ctx.withType(fieldType))
			if err != nil {
				return nil, err
			}
			addrForMarshaler = true
			nilcheck = false
			valueCode = code
		default:
			code, err := encodeCompile(ctx.withType(fieldType), isPtr)
			if err != nil {
				return nil, err
			}
			valueCode = code
		}

		if field.Anonymous {
			tagKey := ""
			if tag.isTaggedKey {
				tagKey = tag.key
			}
			for k, v := range encodeAnonymousStructFieldPairMap(tags, tagKey, valueCode) {
				anonymousFields[k] = append(anonymousFields[k], v...)
			}
			valueCode.decIndent()

			// fix issue144
			if !(isPtr && strings.Contains(valueCode.op.String(), "Marshal")) {
				valueCode.indirect = indirect
			}
		} else {
			valueCode.indirect = indirect
		}
		key := fmt.Sprintf(`"%s":`, tag.key)
		escapedKey := fmt.Sprintf(`%s:`, string(encodeEscapedString([]byte{}, tag.key)))
		fieldCode := &opcode{
			typ:              valueCode.typ,
			displayIdx:       fieldOpcodeIndex,
			idx:              opcodeOffset(fieldPtrIndex),
			next:             valueCode,
			indent:           ctx.indent,
			anonymousKey:     field.Anonymous,
			key:              []byte(key),
			escapedKey:       []byte(escapedKey),
			isTaggedKey:      tag.isTaggedKey,
			displayKey:       tag.key,
			offset:           field.Offset,
			indirect:         indirect,
			nilcheck:         nilcheck,
			addrForMarshaler: addrForMarshaler,
		}
		if fieldIdx == 0 {
			fieldCode.headIdx = fieldCode.idx
			code = encodeStructHeader(ctx, fieldCode, valueCode, tag)
			head = fieldCode
			prevField = fieldCode
		} else {
			fieldCode.headIdx = head.headIdx
			code.next = fieldCode
			code = encodeStructField(ctx, fieldCode, valueCode, tag)
			prevField.nextField = fieldCode
			fieldCode.prevField = prevField
			prevField = fieldCode
		}
		fieldIdx++
	}
	ctx = ctx.decIndent()

	structEndCode := &opcode{
		op:     opStructEnd,
		typ:    nil,
		indent: ctx.indent,
		next:   newEndOp(ctx),
	}

	// no struct field
	if head == nil {
		head = &opcode{
			op:         opStructFieldHead,
			typ:        typ,
			displayIdx: ctx.opcodeIndex,
			idx:        opcodeOffset(ctx.ptrIndex),
			headIdx:    opcodeOffset(ctx.ptrIndex),
			indent:     ctx.indent,
			nextField:  structEndCode,
		}
		structEndCode.prevField = head
		ctx.incIndex()
		code = head
	}

	structEndCode.displayIdx = ctx.opcodeIndex
	structEndCode.idx = opcodeOffset(ctx.ptrIndex)
	ctx.incIndex()

	if prevField != nil && prevField.nextField == nil {
		prevField.nextField = structEndCode
		structEndCode.prevField = prevField
	}

	head.end = structEndCode
	code.next = structEndCode
	encodeOptimizeConflictAnonymousFields(anonymousFields)
	encodeOptimizeAnonymousFields(head)
	ret := (*opcode)(unsafe.Pointer(head))
	compiled.code = ret

	delete(ctx.structTypeToCompiledCode, typeptr)

	if !disableIndirectConversion && !head.indirect && isPtr {
		head.indirect = true
	}

	return ret, nil
}

func encodeIsPtrMarshalJSONType(typ *rtype) bool {
	return !typ.Implements(marshalJSONType) && rtype_ptrTo(typ).Implements(marshalJSONType)
}

func encodeIsPtrMarshalTextType(typ *rtype) bool {
	return !typ.Implements(marshalTextType) && rtype_ptrTo(typ).Implements(marshalTextType)
}

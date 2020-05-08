package json

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

func (e *Encoder) compileHead(typ *rtype, withIndent bool) (*opcode, error) {
	if typ.Implements(marshalJSONType) {
		return newOpCode(opMarshalJSON, typ, e.indent, newEndOp(e.indent)), nil
	} else if typ.Implements(marshalTextType) {
		return newOpCode(opMarshalText, typ, e.indent, newEndOp(e.indent)), nil
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return e.compile(typ, withIndent)
}

func (e *Encoder) compile(typ *rtype, withIndent bool) (*opcode, error) {
	if typ.Implements(marshalJSONType) {
		return newOpCode(opMarshalJSON, typ, e.indent, newEndOp(e.indent)), nil
	} else if typ.Implements(marshalTextType) {
		return newOpCode(opMarshalText, typ, e.indent, newEndOp(e.indent)), nil
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return e.compilePtr(typ, withIndent)
	case reflect.Slice:
		return e.compileSlice(typ, withIndent)
	case reflect.Array:
		return e.compileArray(typ, withIndent)
	case reflect.Map:
		return e.compileMap(typ, withIndent)
	case reflect.Struct:
		return e.compileStruct(typ, withIndent)
	case reflect.Int:
		return e.compileInt(typ)
	case reflect.Int8:
		return e.compileInt8(typ)
	case reflect.Int16:
		return e.compileInt16(typ)
	case reflect.Int32:
		return e.compileInt32(typ)
	case reflect.Int64:
		return e.compileInt64(typ)
	case reflect.Uint:
		return e.compileUint(typ)
	case reflect.Uint8:
		return e.compileUint8(typ)
	case reflect.Uint16:
		return e.compileUint16(typ)
	case reflect.Uint32:
		return e.compileUint32(typ)
	case reflect.Uint64:
		return e.compileUint64(typ)
	case reflect.Uintptr:
		return e.compileUint(typ)
	case reflect.Float32:
		return e.compileFloat32(typ)
	case reflect.Float64:
		return e.compileFloat64(typ)
	case reflect.String:
		return e.compileString(typ)
	case reflect.Bool:
		return e.compileBool(typ)
	case reflect.Interface:
		return e.compileInterface(typ)
	}
	return nil, &UnsupportedTypeError{Type: rtype2type(typ)}
}

func (e *Encoder) optimizeStructFieldPtrHead(typ *rtype, code *opcode) *opcode {
	switch code.op {
	case opStructFieldHead:
		code.op = opStructFieldPtrHead
	case opStructFieldHeadInt:
		code.op = opStructFieldPtrHeadInt
	case opStructFieldHeadInt8:
		code.op = opStructFieldPtrHeadInt8
	case opStructFieldHeadInt16:
		code.op = opStructFieldPtrHeadInt16
	case opStructFieldHeadInt32:
		code.op = opStructFieldPtrHeadInt32
	case opStructFieldHeadInt64:
		code.op = opStructFieldPtrHeadInt64
	case opStructFieldHeadUint:
		code.op = opStructFieldPtrHeadUint
	case opStructFieldHeadUint8:
		code.op = opStructFieldPtrHeadUint8
	case opStructFieldHeadUint16:
		code.op = opStructFieldPtrHeadUint16
	case opStructFieldHeadUint32:
		code.op = opStructFieldPtrHeadUint32
	case opStructFieldHeadUint64:
		code.op = opStructFieldPtrHeadUint64
	case opStructFieldHeadFloat32:
		code.op = opStructFieldPtrHeadFloat32
	case opStructFieldHeadFloat64:
		code.op = opStructFieldPtrHeadFloat64
	case opStructFieldHeadString:
		code.op = opStructFieldPtrHeadString
	case opStructFieldHeadBool:
		code.op = opStructFieldPtrHeadBool
	case opStructFieldHeadOmitEmpty:
		code.op = opStructFieldPtrHeadOmitEmpty
	case opStructFieldHeadIntOmitEmpty:
		code.op = opStructFieldPtrHeadIntOmitEmpty
	case opStructFieldHeadInt8OmitEmpty:
		code.op = opStructFieldPtrHeadInt8OmitEmpty
	case opStructFieldHeadInt16OmitEmpty:
		code.op = opStructFieldPtrHeadInt16OmitEmpty
	case opStructFieldHeadInt32OmitEmpty:
		code.op = opStructFieldPtrHeadInt32OmitEmpty
	case opStructFieldHeadInt64OmitEmpty:
		code.op = opStructFieldPtrHeadInt64OmitEmpty
	case opStructFieldHeadUintOmitEmpty:
		code.op = opStructFieldPtrHeadUintOmitEmpty
	case opStructFieldHeadUint8OmitEmpty:
		code.op = opStructFieldPtrHeadUint8OmitEmpty
	case opStructFieldHeadUint16OmitEmpty:
		code.op = opStructFieldPtrHeadUint16OmitEmpty
	case opStructFieldHeadUint32OmitEmpty:
		code.op = opStructFieldPtrHeadUint32OmitEmpty
	case opStructFieldHeadUint64OmitEmpty:
		code.op = opStructFieldPtrHeadUint64OmitEmpty
	case opStructFieldHeadFloat32OmitEmpty:
		code.op = opStructFieldPtrHeadFloat32OmitEmpty
	case opStructFieldHeadFloat64OmitEmpty:
		code.op = opStructFieldPtrHeadFloat64OmitEmpty
	case opStructFieldHeadStringOmitEmpty:
		code.op = opStructFieldPtrHeadStringOmitEmpty
	case opStructFieldHeadBoolOmitEmpty:
		code.op = opStructFieldPtrHeadBoolOmitEmpty

	case opStructFieldHeadIndent:
		code.op = opStructFieldPtrHeadIndent
	case opStructFieldHeadIntIndent:
		code.op = opStructFieldPtrHeadIntIndent
	case opStructFieldHeadInt8Indent:
		code.op = opStructFieldPtrHeadInt8Indent
	case opStructFieldHeadInt16Indent:
		code.op = opStructFieldPtrHeadInt16Indent
	case opStructFieldHeadInt32Indent:
		code.op = opStructFieldPtrHeadInt32Indent
	case opStructFieldHeadInt64Indent:
		code.op = opStructFieldPtrHeadInt64Indent
	case opStructFieldHeadUintIndent:
		code.op = opStructFieldPtrHeadUintIndent
	case opStructFieldHeadUint8Indent:
		code.op = opStructFieldPtrHeadUint8Indent
	case opStructFieldHeadUint16Indent:
		code.op = opStructFieldPtrHeadUint16Indent
	case opStructFieldHeadUint32Indent:
		code.op = opStructFieldPtrHeadUint32Indent
	case opStructFieldHeadUint64Indent:
		code.op = opStructFieldPtrHeadUint64Indent
	case opStructFieldHeadFloat32Indent:
		code.op = opStructFieldPtrHeadFloat32Indent
	case opStructFieldHeadFloat64Indent:
		code.op = opStructFieldPtrHeadFloat64Indent
	case opStructFieldHeadStringIndent:
		code.op = opStructFieldPtrHeadStringIndent
	case opStructFieldHeadBoolIndent:
		code.op = opStructFieldPtrHeadBoolIndent
	case opStructFieldHeadOmitEmptyIndent:
		code.op = opStructFieldPtrHeadOmitEmptyIndent
	case opStructFieldHeadIntOmitEmptyIndent:
		code.op = opStructFieldPtrHeadIntOmitEmptyIndent
	case opStructFieldHeadInt8OmitEmptyIndent:
		code.op = opStructFieldPtrHeadInt8OmitEmptyIndent
	case opStructFieldHeadInt16OmitEmptyIndent:
		code.op = opStructFieldPtrHeadInt16OmitEmptyIndent
	case opStructFieldHeadInt32OmitEmptyIndent:
		code.op = opStructFieldPtrHeadInt32OmitEmptyIndent
	case opStructFieldHeadInt64OmitEmptyIndent:
		code.op = opStructFieldPtrHeadInt64OmitEmptyIndent
	case opStructFieldHeadUintOmitEmptyIndent:
		code.op = opStructFieldPtrHeadUintOmitEmptyIndent
	case opStructFieldHeadUint8OmitEmptyIndent:
		code.op = opStructFieldPtrHeadUint8OmitEmptyIndent
	case opStructFieldHeadUint16OmitEmptyIndent:
		code.op = opStructFieldPtrHeadUint16OmitEmptyIndent
	case opStructFieldHeadUint32OmitEmptyIndent:
		code.op = opStructFieldPtrHeadUint32OmitEmptyIndent
	case opStructFieldHeadUint64OmitEmptyIndent:
		code.op = opStructFieldPtrHeadUint64OmitEmptyIndent
	case opStructFieldHeadFloat32OmitEmptyIndent:
		code.op = opStructFieldPtrHeadFloat32OmitEmptyIndent
	case opStructFieldHeadFloat64OmitEmptyIndent:
		code.op = opStructFieldPtrHeadFloat64OmitEmptyIndent
	case opStructFieldHeadStringOmitEmptyIndent:
		code.op = opStructFieldPtrHeadStringOmitEmptyIndent
	case opStructFieldHeadBoolOmitEmptyIndent:
		code.op = opStructFieldPtrHeadBoolOmitEmptyIndent
	default:
		return newOpCode(opPtr, typ, e.indent, code)
	}
	return code
}

func (e *Encoder) compilePtr(typ *rtype, withIndent bool) (*opcode, error) {
	code, err := e.compile(typ.Elem(), withIndent)
	if err != nil {
		return nil, err
	}
	return e.optimizeStructFieldPtrHead(typ, code), nil
}

func (e *Encoder) compileInt(typ *rtype) (*opcode, error) {
	return newOpCode(opInt, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileInt8(typ *rtype) (*opcode, error) {
	return newOpCode(opInt8, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileInt16(typ *rtype) (*opcode, error) {
	return newOpCode(opInt16, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileInt32(typ *rtype) (*opcode, error) {
	return newOpCode(opInt32, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileInt64(typ *rtype) (*opcode, error) {
	return newOpCode(opInt64, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileUint(typ *rtype) (*opcode, error) {
	return newOpCode(opUint, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileUint8(typ *rtype) (*opcode, error) {
	return newOpCode(opUint8, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileUint16(typ *rtype) (*opcode, error) {
	return newOpCode(opUint16, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileUint32(typ *rtype) (*opcode, error) {
	return newOpCode(opUint32, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileUint64(typ *rtype) (*opcode, error) {
	return newOpCode(opUint64, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileFloat32(typ *rtype) (*opcode, error) {
	return newOpCode(opFloat32, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileFloat64(typ *rtype) (*opcode, error) {
	return newOpCode(opFloat64, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileString(typ *rtype) (*opcode, error) {
	return newOpCode(opString, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileBool(typ *rtype) (*opcode, error) {
	return newOpCode(opBool, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileInterface(typ *rtype) (*opcode, error) {
	return newOpCode(opInterface, typ, e.indent, newEndOp(e.indent)), nil
}

func (e *Encoder) compileSlice(typ *rtype, withIndent bool) (*opcode, error) {
	elem := typ.Elem()
	size := elem.Size()

	e.indent++
	code, err := e.compile(elem, withIndent)
	e.indent--

	if err != nil {
		return nil, err
	}

	// header => opcode => elem => end
	//             ^        |
	//             |________|

	header := newSliceHeaderCode(e.indent)
	elemCode := &sliceElemCode{
		opcodeHeader: &opcodeHeader{
			op:     opSliceElem,
			indent: e.indent,
		},
		size: size,
	}
	end := newOpCode(opSliceEnd, nil, e.indent, newEndOp(e.indent))
	if withIndent {
		header.op = opSliceHeadIndent
		elemCode.op = opSliceElemIndent
		end.op = opSliceEndIndent
	}

	header.elem = elemCode
	header.end = end
	header.next = code
	code.beforeLastCode().next = (*opcode)(unsafe.Pointer(elemCode))
	elemCode.next = code
	elemCode.end = end
	return (*opcode)(unsafe.Pointer(header)), nil
}

func (e *Encoder) compileArray(typ *rtype, withIndent bool) (*opcode, error) {
	elem := typ.Elem()
	alen := typ.Len()
	size := elem.Size()

	e.indent++
	code, err := e.compile(elem, withIndent)
	e.indent--

	if err != nil {
		return nil, err
	}
	// header => opcode => elem => end
	//             ^        |
	//             |________|

	header := newArrayHeaderCode(e.indent, alen)
	elemCode := &arrayElemCode{
		opcodeHeader: &opcodeHeader{
			op: opArrayElem,
		},
		len:  uintptr(alen),
		size: size,
	}
	end := newOpCode(opArrayEnd, nil, e.indent, newEndOp(e.indent))

	if withIndent {
		header.op = opArrayHeadIndent
		elemCode.op = opArrayElemIndent
		end.op = opArrayEndIndent
	}

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

func (e *Encoder) compileMap(typ *rtype, withIndent bool) (*opcode, error) {
	// header => code => value => code => key => code => value => code => end
	//                                     ^                       |
	//                                     |_______________________|
	e.indent++
	keyType := typ.Key()
	keyCode, err := e.compile(keyType, withIndent)
	if err != nil {
		return nil, err
	}
	valueType := typ.Elem()
	valueCode, err := e.compile(valueType, withIndent)
	if err != nil {
		return nil, err
	}

	key := newMapKeyCode(e.indent)
	value := newMapValueCode(e.indent)

	e.indent--

	header := newMapHeaderCode(typ, e.indent)
	header.key = key
	header.value = value
	end := newOpCode(opMapEnd, nil, e.indent, newEndOp(e.indent))

	if withIndent {
		header.op = opMapHeadIndent
		key.op = opMapKeyIndent
		value.op = opMapValueIndent
		end.op = opMapEndIndent
	}

	header.next = keyCode
	keyCode.beforeLastCode().next = (*opcode)(unsafe.Pointer(value))
	value.next = valueCode
	valueCode.beforeLastCode().next = (*opcode)(unsafe.Pointer(key))
	key.next = keyCode

	header.end = end
	key.end = end

	return (*opcode)(unsafe.Pointer(header)), nil
}

func (e *Encoder) getTag(field reflect.StructField) string {
	return field.Tag.Get("json")
}

func (e *Encoder) isIgnoredStructField(field reflect.StructField) bool {
	if field.PkgPath != "" && !field.Anonymous {
		// private field
		return true
	}
	tag := e.getTag(field)
	if tag == "-" {
		return true
	}
	return false
}

func (e *Encoder) optimizeStructHeaderOmitEmptyIndent(op opType) opType {
	switch op {
	case opInt:
		return opStructFieldHeadIntOmitEmptyIndent
	case opInt8:
		return opStructFieldHeadInt8OmitEmptyIndent
	case opInt16:
		return opStructFieldHeadInt16OmitEmptyIndent
	case opInt32:
		return opStructFieldHeadInt32OmitEmptyIndent
	case opInt64:
		return opStructFieldHeadInt64OmitEmptyIndent
	case opUint:
		return opStructFieldHeadUintOmitEmptyIndent
	case opUint8:
		return opStructFieldHeadUint8OmitEmptyIndent
	case opUint16:
		return opStructFieldHeadUint16OmitEmptyIndent
	case opUint32:
		return opStructFieldHeadUint32OmitEmptyIndent
	case opUint64:
		return opStructFieldHeadUint64OmitEmptyIndent
	case opFloat32:
		return opStructFieldHeadFloat32OmitEmptyIndent
	case opFloat64:
		return opStructFieldHeadFloat64OmitEmptyIndent
	case opString:
		return opStructFieldHeadStringOmitEmptyIndent
	case opBool:
		return opStructFieldHeadBoolOmitEmptyIndent
	}
	return opStructFieldHeadOmitEmptyIndent
}

func (e *Encoder) optimizeStructHeaderIndent(op opType, isOmitEmpty bool) opType {
	if isOmitEmpty {
		return e.optimizeStructHeaderOmitEmptyIndent(op)
	}
	switch op {
	case opInt:
		return opStructFieldHeadIntIndent
	case opInt8:
		return opStructFieldHeadInt8Indent
	case opInt16:
		return opStructFieldHeadInt16Indent
	case opInt32:
		return opStructFieldHeadInt32Indent
	case opInt64:
		return opStructFieldHeadInt64Indent
	case opUint:
		return opStructFieldHeadUintIndent
	case opUint8:
		return opStructFieldHeadUint8Indent
	case opUint16:
		return opStructFieldHeadUint16Indent
	case opUint32:
		return opStructFieldHeadUint32Indent
	case opUint64:
		return opStructFieldHeadUint64Indent
	case opFloat32:
		return opStructFieldHeadFloat32Indent
	case opFloat64:
		return opStructFieldHeadFloat64Indent
	case opString:
		return opStructFieldHeadStringIndent
	case opBool:
		return opStructFieldHeadBoolIndent
	}
	return opStructFieldHeadIndent
}

func (e *Encoder) optimizeStructHeaderOmitEmpty(op opType) opType {
	switch op {
	case opInt:
		return opStructFieldHeadIntOmitEmpty
	case opInt8:
		return opStructFieldHeadInt8OmitEmpty
	case opInt16:
		return opStructFieldHeadInt16OmitEmpty
	case opInt32:
		return opStructFieldHeadInt32OmitEmpty
	case opInt64:
		return opStructFieldHeadInt64OmitEmpty
	case opUint:
		return opStructFieldHeadUintOmitEmpty
	case opUint8:
		return opStructFieldHeadUint8OmitEmpty
	case opUint16:
		return opStructFieldHeadUint16OmitEmpty
	case opUint32:
		return opStructFieldHeadUint32OmitEmpty
	case opUint64:
		return opStructFieldHeadUint64OmitEmpty
	case opFloat32:
		return opStructFieldHeadFloat32OmitEmpty
	case opFloat64:
		return opStructFieldHeadFloat64OmitEmpty
	case opString:
		return opStructFieldHeadStringOmitEmpty
	case opBool:
		return opStructFieldHeadBoolOmitEmpty
	}
	return opStructFieldHeadOmitEmpty
}

func (e *Encoder) optimizeStructHeader(op opType, isOmitEmpty, withIndent bool) opType {
	if withIndent {
		return e.optimizeStructHeaderIndent(op, isOmitEmpty)
	}
	if isOmitEmpty {
		return e.optimizeStructHeaderOmitEmpty(op)
	}
	switch op {
	case opInt:
		return opStructFieldHeadInt
	case opInt8:
		return opStructFieldHeadInt8
	case opInt16:
		return opStructFieldHeadInt16
	case opInt32:
		return opStructFieldHeadInt32
	case opInt64:
		return opStructFieldHeadInt64
	case opUint:
		return opStructFieldHeadUint
	case opUint8:
		return opStructFieldHeadUint8
	case opUint16:
		return opStructFieldHeadUint16
	case opUint32:
		return opStructFieldHeadUint32
	case opUint64:
		return opStructFieldHeadUint64
	case opFloat32:
		return opStructFieldHeadFloat32
	case opFloat64:
		return opStructFieldHeadFloat64
	case opString:
		return opStructFieldHeadString
	case opBool:
		return opStructFieldHeadBool
	}
	return opStructFieldHead
}

func (e *Encoder) optimizeStructFieldOmitEmptyIndent(op opType) opType {
	switch op {
	case opInt:
		return opStructFieldIntOmitEmptyIndent
	case opInt8:
		return opStructFieldInt8OmitEmptyIndent
	case opInt16:
		return opStructFieldInt16OmitEmptyIndent
	case opInt32:
		return opStructFieldInt32OmitEmptyIndent
	case opInt64:
		return opStructFieldInt64OmitEmptyIndent
	case opUint:
		return opStructFieldUintOmitEmptyIndent
	case opUint8:
		return opStructFieldUint8OmitEmptyIndent
	case opUint16:
		return opStructFieldUint16OmitEmptyIndent
	case opUint32:
		return opStructFieldUint32OmitEmptyIndent
	case opUint64:
		return opStructFieldUint64OmitEmptyIndent
	case opFloat32:
		return opStructFieldFloat32OmitEmptyIndent
	case opFloat64:
		return opStructFieldFloat64OmitEmptyIndent
	case opString:
		return opStructFieldStringOmitEmptyIndent
	case opBool:
		return opStructFieldBoolOmitEmptyIndent
	}
	return opStructFieldOmitEmptyIndent
}

func (e *Encoder) optimizeStructFieldIndent(op opType, isOmitEmpty bool) opType {
	if isOmitEmpty {
		return e.optimizeStructFieldOmitEmptyIndent(op)
	}
	switch op {
	case opInt:
		return opStructFieldIntIndent
	case opInt8:
		return opStructFieldInt8Indent
	case opInt16:
		return opStructFieldInt16Indent
	case opInt32:
		return opStructFieldInt32Indent
	case opInt64:
		return opStructFieldInt64Indent
	case opUint:
		return opStructFieldUintIndent
	case opUint8:
		return opStructFieldUint8Indent
	case opUint16:
		return opStructFieldUint16Indent
	case opUint32:
		return opStructFieldUint32Indent
	case opUint64:
		return opStructFieldUint64Indent
	case opFloat32:
		return opStructFieldFloat32Indent
	case opFloat64:
		return opStructFieldFloat64Indent
	case opString:
		return opStructFieldStringIndent
	case opBool:
		return opStructFieldBoolIndent
	}
	return opStructFieldIndent
}

func (e *Encoder) optimizeStructFieldOmitEmpty(op opType) opType {
	switch op {
	case opInt:
		return opStructFieldIntOmitEmpty
	case opInt8:
		return opStructFieldInt8OmitEmpty
	case opInt16:
		return opStructFieldInt16OmitEmpty
	case opInt32:
		return opStructFieldInt32OmitEmpty
	case opInt64:
		return opStructFieldInt64OmitEmpty
	case opUint:
		return opStructFieldUintOmitEmpty
	case opUint8:
		return opStructFieldUint8OmitEmpty
	case opUint16:
		return opStructFieldUint16OmitEmpty
	case opUint32:
		return opStructFieldUint32OmitEmpty
	case opUint64:
		return opStructFieldUint64OmitEmpty
	case opFloat32:
		return opStructFieldFloat32OmitEmpty
	case opFloat64:
		return opStructFieldFloat64OmitEmpty
	case opString:
		return opStructFieldStringOmitEmpty
	case opBool:
		return opStructFieldBoolOmitEmpty
	}
	return opStructFieldOmitEmpty
}

func (e *Encoder) optimizeStructField(op opType, isOmitEmpty, withIndent bool) opType {
	if withIndent {
		return e.optimizeStructFieldIndent(op, isOmitEmpty)
	}
	if isOmitEmpty {
		return e.optimizeStructFieldOmitEmpty(op)
	}
	switch op {
	case opInt:
		return opStructFieldInt
	case opInt8:
		return opStructFieldInt8
	case opInt16:
		return opStructFieldInt16
	case opInt32:
		return opStructFieldInt32
	case opInt64:
		return opStructFieldInt64
	case opUint:
		return opStructFieldUint
	case opUint8:
		return opStructFieldUint8
	case opUint16:
		return opStructFieldUint16
	case opUint32:
		return opStructFieldUint32
	case opUint64:
		return opStructFieldUint64
	case opFloat32:
		return opStructFieldFloat32
	case opFloat64:
		return opStructFieldFloat64
	case opString:
		return opStructFieldString
	case opBool:
		return opStructFieldBool
	}
	return opStructField
}

func (e *Encoder) compileStruct(typ *rtype, withIndent bool) (*opcode, error) {
	// header => code => structField => code => end
	//                        ^          |
	//                        |__________|
	fieldNum := typ.NumField()
	fieldIdx := 0
	var (
		head      *structFieldCode
		code      *opcode
		prevField *structFieldCode
	)
	e.indent++
	for i := 0; i < fieldNum; i++ {
		field := typ.Field(i)
		if e.isIgnoredStructField(field) {
			continue
		}
		keyName := field.Name
		tag := e.getTag(field)
		opts := strings.Split(tag, ",")
		if len(opts) > 0 {
			if opts[0] != "" {
				keyName = opts[0]
			}
		}
		isOmitEmpty := false
		if len(opts) > 1 {
			isOmitEmpty = opts[1] == "omitempty"
		}
		fieldType := type2rtype(field.Type)
		valueCode, err := e.compile(fieldType, withIndent)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf(`"%s":`, keyName)
		fieldCode := &structFieldCode{
			opcodeHeader: &opcodeHeader{
				typ:    fieldType,
				next:   valueCode,
				indent: e.indent,
			},
			key:    []byte(key),
			offset: field.Offset,
		}
		if fieldIdx == 0 {
			fieldCode.indent--
			head = fieldCode
			code = (*opcode)(unsafe.Pointer(fieldCode))
			prevField = fieldCode
			op := e.optimizeStructHeader(valueCode.op, isOmitEmpty, withIndent)
			fieldCode.op = op
			switch op {
			case opStructFieldHead,
				opStructFieldHeadOmitEmpty,
				opStructFieldHeadIndent,
				opStructFieldHeadOmitEmptyIndent:
				code = valueCode.beforeLastCode()
			}
		} else {
			code.next = (*opcode)(unsafe.Pointer(fieldCode))
			prevField.nextField = (*opcode)(unsafe.Pointer(fieldCode))
			prevField = fieldCode
			code = (*opcode)(unsafe.Pointer(fieldCode))
			op := e.optimizeStructField(valueCode.op, isOmitEmpty, withIndent)
			fieldCode.op = op
			switch op {
			case opStructField,
				opStructFieldOmitEmpty,
				opStructFieldIndent,
				opStructFieldOmitEmptyIndent:
				code = valueCode.beforeLastCode()
			}
		}
		fieldIdx++
	}
	e.indent--

	structEndCode := newOpCode(opStructEnd, nil, e.indent, nil)

	if withIndent {
		structEndCode.op = opStructEndIndent
	}

	if prevField != nil && prevField.nextField == nil {
		prevField.nextField = structEndCode
	}

	// no struct field
	if head == nil {
		head = &structFieldCode{
			opcodeHeader: &opcodeHeader{
				op:     opStructFieldHead,
				typ:    typ,
				indent: e.indent,
			},
			nextField: structEndCode,
		}
		if withIndent {
			head.op = opStructFieldHeadIndent
		}
		code = (*opcode)(unsafe.Pointer(head))
	}
	head.end = structEndCode
	code.next = structEndCode
	structEndCode.next = newEndOp(e.indent)
	return (*opcode)(unsafe.Pointer(head)), nil
}

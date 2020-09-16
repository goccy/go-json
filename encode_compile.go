package json

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"
)

func (e *Encoder) compileHead(ctx *encodeCompileContext) (*opcode, error) {
	typ := ctx.typ
	switch {
	case typ.Implements(marshalJSONType):
		return e.compileMarshalJSON(ctx)
	case rtype_ptrTo(typ).Implements(marshalJSONType):
		return e.compileMarshalJSONPtr(ctx)
	case typ.Implements(marshalTextType):
		return e.compileMarshalText(ctx)
	case rtype_ptrTo(typ).Implements(marshalTextType):
		return e.compileMarshalTextPtr(ctx)
	}
	isPtr := false
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		isPtr = true
	}
	if typ.Kind() == reflect.Map {
		return e.compileMap(ctx.withType(typ), isPtr)
	} else if typ.Kind() == reflect.Struct {
		return e.compileStruct(ctx.withType(typ), isPtr)
	}
	return e.compile(ctx.withType(typ))
}

func (e *Encoder) implementsMarshaler(typ *rtype) bool {
	switch {
	case typ.Implements(marshalJSONType):
		return true
	case rtype_ptrTo(typ).Implements(marshalJSONType):
		return true
	case typ.Implements(marshalTextType):
		return true
	case rtype_ptrTo(typ).Implements(marshalTextType):
		return true
	}
	return false
}

func (e *Encoder) compile(ctx *encodeCompileContext) (*opcode, error) {
	typ := ctx.typ
	switch {
	case typ.Implements(marshalJSONType):
		return e.compileMarshalJSON(ctx)
	case rtype_ptrTo(typ).Implements(marshalJSONType):
		return e.compileMarshalJSONPtr(ctx)
	case typ.Implements(marshalTextType):
		return e.compileMarshalText(ctx)
	case rtype_ptrTo(typ).Implements(marshalTextType):
		return e.compileMarshalTextPtr(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return e.compilePtr(ctx)
	case reflect.Slice:
		elem := typ.Elem()
		if !e.implementsMarshaler(elem) && elem.Kind() == reflect.Uint8 {
			return e.compileBytes(ctx)
		}
		return e.compileSlice(ctx)
	case reflect.Array:
		return e.compileArray(ctx)
	case reflect.Map:
		return e.compileMap(ctx, true)
	case reflect.Struct:
		return e.compileStruct(ctx, false)
	case reflect.Interface:
		return e.compileInterface(ctx)
	case reflect.Int:
		return e.compileInt(ctx)
	case reflect.Int8:
		return e.compileInt8(ctx)
	case reflect.Int16:
		return e.compileInt16(ctx)
	case reflect.Int32:
		return e.compileInt32(ctx)
	case reflect.Int64:
		return e.compileInt64(ctx)
	case reflect.Uint:
		return e.compileUint(ctx)
	case reflect.Uint8:
		return e.compileUint8(ctx)
	case reflect.Uint16:
		return e.compileUint16(ctx)
	case reflect.Uint32:
		return e.compileUint32(ctx)
	case reflect.Uint64:
		return e.compileUint64(ctx)
	case reflect.Uintptr:
		return e.compileUint(ctx)
	case reflect.Float32:
		return e.compileFloat32(ctx)
	case reflect.Float64:
		return e.compileFloat64(ctx)
	case reflect.String:
		return e.compileString(ctx)
	case reflect.Bool:
		return e.compileBool(ctx)
	}
	return nil, &UnsupportedTypeError{Type: rtype2type(typ)}
}

func (e *Encoder) compileKey(ctx *encodeCompileContext) (*opcode, error) {
	typ := ctx.typ
	switch {
	case typ.Implements(marshalJSONType):
		return e.compileMarshalJSON(ctx)
	case rtype_ptrTo(typ).Implements(marshalJSONType):
		return e.compileMarshalJSONPtr(ctx)
	case typ.Implements(marshalTextType):
		return e.compileMarshalText(ctx)
	case rtype_ptrTo(typ).Implements(marshalTextType):
		return e.compileMarshalTextPtr(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return e.compilePtr(ctx)
	case reflect.Interface:
		return e.compileInterface(ctx)
	case reflect.String:
		return e.compileString(ctx)
	}
	return nil, &UnsupportedTypeError{Type: rtype2type(typ)}
}

func (e *Encoder) compilePtr(ctx *encodeCompileContext) (*opcode, error) {
	ptrOpcodeIndex := ctx.opcodeIndex
	ctx.incIndex()
	code, err := e.compile(ctx.withType(ctx.typ.Elem()))
	if err != nil {
		return nil, err
	}
	ptrHeadOp := code.op.headToPtrHead()
	if code.op != ptrHeadOp {
		code.op = ptrHeadOp
		code.decOpcodeIndex()
		ctx.decIndex()
		return code, nil
	}
	c := ctx.context()
	c.opcodeIndex = ptrOpcodeIndex
	return newOpCodeWithNext(c, opPtr, code), nil
}

func (e *Encoder) compileMarshalJSON(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opMarshalJSON)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileMarshalJSONPtr(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx.withType(rtype_ptrTo(ctx.typ)), opMarshalJSON)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileMarshalText(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opMarshalText)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileMarshalTextPtr(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx.withType(rtype_ptrTo(ctx.typ)), opMarshalText)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileInt(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileInt8(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt8)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileInt16(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt16)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileInt32(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt32)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileInt64(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opInt64)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileUint(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileUint8(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint8)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileUint16(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint16)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileUint32(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint32)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileUint64(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opUint64)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileFloat32(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opFloat32)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileFloat64(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opFloat64)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileString(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opString)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileBool(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opBool)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileBytes(ctx *encodeCompileContext) (*opcode, error) {
	code := newOpCode(ctx, opBytes)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileInterface(ctx *encodeCompileContext) (*opcode, error) {
	code := newInterfaceCode(ctx)
	ctx.incIndex()
	return code, nil
}

func (e *Encoder) compileSlice(ctx *encodeCompileContext) (*opcode, error) {
	ctx.root = false
	elem := ctx.typ.Elem()
	size := elem.Size()

	header := newSliceHeaderCode(ctx)
	ctx.incIndex()

	code, err := e.compile(ctx.withType(ctx.typ.Elem()).incIndent())
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
	if ctx.withIndent {
		if ctx.root {
			header.op = opRootSliceHeadIndent
			elemCode.op = opRootSliceElemIndent
		} else {
			header.op = opSliceHeadIndent
			elemCode.op = opSliceElemIndent
		}
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

func (e *Encoder) compileArray(ctx *encodeCompileContext) (*opcode, error) {
	ctx.root = false
	typ := ctx.typ
	elem := typ.Elem()
	alen := typ.Len()
	size := elem.Size()

	header := newArrayHeaderCode(ctx, alen)
	ctx.incIndex()

	code, err := e.compile(ctx.withType(elem).incIndent())
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

	if ctx.withIndent {
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

func (e *Encoder) compileMap(ctx *encodeCompileContext, withLoad bool) (*opcode, error) {
	// header => code => value => code => key => code => value => code => end
	//                                     ^                       |
	//                                     |_______________________|
	ctx = ctx.incIndent()
	header := newMapHeaderCode(ctx, withLoad)
	ctx.incIndex()

	typ := ctx.typ
	keyType := ctx.typ.Key()
	keyCode, err := e.compileKey(ctx.withType(keyType))
	if err != nil {
		return nil, err
	}

	value := newMapValueCode(ctx, header)
	ctx.incIndex()

	valueType := typ.Elem()
	valueCode, err := e.compile(ctx.withType(valueType))
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

	if ctx.withIndent {
		header.op = header.op.toIndent()
		key.op = key.op.toIndent()
		value.op = value.op.toIndent()
		end.op = end.op.toIndent()
	}

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

func (e *Encoder) typeToHeaderType(op opType) opType {
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
	case opMapHead:
		return opStructFieldHeadMap
	case opMapHeadLoad:
		return opStructFieldHeadMapLoad
	case opMapHeadIndent:
		return opStructFieldHeadMapIndent
	case opMapHeadLoadIndent:
		return opStructFieldHeadMapLoadIndent
	case opArrayHead:
		return opStructFieldHeadArray
	case opArrayHeadIndent:
		return opStructFieldHeadArrayIndent
	case opSliceHead:
		return opStructFieldHeadSlice
	case opSliceHeadIndent:
		return opStructFieldHeadSliceIndent
	case opStructFieldHead:
		return opStructFieldHeadStruct
	case opStructFieldHeadIndent:
		return opStructFieldHeadStructIndent
	case opMarshalJSON:
		return opStructFieldHeadMarshalJSON
	case opMarshalText:
		return opStructFieldHeadMarshalText
	}
	return opStructFieldHead
}

func (e *Encoder) typeToFieldType(op opType) opType {
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
	case opMapHead:
		return opStructFieldMap
	case opMapHeadLoad:
		return opStructFieldMapLoad
	case opMapHeadIndent:
		return opStructFieldMapIndent
	case opMapHeadLoadIndent:
		return opStructFieldMapLoadIndent
	case opArrayHead:
		return opStructFieldArray
	case opArrayHeadIndent:
		return opStructFieldArrayIndent
	case opSliceHead:
		return opStructFieldSlice
	case opSliceHeadIndent:
		return opStructFieldSliceIndent
	case opStructFieldHead:
		return opStructFieldStruct
	case opStructFieldHeadIndent:
		return opStructFieldStructIndent
	case opMarshalJSON:
		return opStructFieldMarshalJSON
	case opMarshalText:
		return opStructFieldMarshalText
	}
	return opStructField
}

func (e *Encoder) optimizeStructHeader(op opType, tag *structTag, withIndent bool) opType {
	headType := e.typeToHeaderType(op)
	switch {
	case tag.isOmitEmpty:
		headType = headType.headToOmitEmptyHead()
	case tag.isString:
		headType = headType.headToStringTagHead()
	}
	if withIndent {
		return headType.toIndent()
	}
	return headType
}

func (e *Encoder) optimizeStructField(op opType, tag *structTag, withIndent bool) opType {
	fieldType := e.typeToFieldType(op)
	switch {
	case tag.isOmitEmpty:
		fieldType = fieldType.fieldToOmitEmptyField()
	case tag.isString:
		fieldType = fieldType.fieldToStringTagField()
	}
	if withIndent {
		return fieldType.toIndent()
	}
	return fieldType
}

func (e *Encoder) recursiveCode(ctx *encodeCompileContext, jmp *compiledCode) *opcode {
	code := newRecursiveCode(ctx, jmp)
	ctx.incIndex()
	return code
}

func (e *Encoder) compiledCode(ctx *encodeCompileContext) *opcode {
	typ := ctx.typ
	typeptr := uintptr(unsafe.Pointer(typ))
	if ctx.withIndent {
		if compiledCode, exists := e.structTypeToCompiledIndentCode[typeptr]; exists {
			return e.recursiveCode(ctx, compiledCode)
		}
	} else {
		if compiledCode, exists := e.structTypeToCompiledCode[typeptr]; exists {
			return e.recursiveCode(ctx, compiledCode)
		}
	}
	return nil
}

func (e *Encoder) structHeader(ctx *encodeCompileContext, fieldCode *opcode, valueCode *opcode, tag *structTag) *opcode {
	fieldCode.indent--
	op := e.optimizeStructHeader(valueCode.op, tag, ctx.withIndent)
	fieldCode.op = op
	switch op {
	case opStructFieldHead,
		opStructFieldHeadSlice,
		opStructFieldHeadArray,
		opStructFieldHeadMap,
		opStructFieldHeadMapLoad,
		opStructFieldHeadStruct,
		opStructFieldHeadOmitEmpty,
		opStructFieldHeadOmitEmptySlice,
		opStructFieldHeadOmitEmptyArray,
		opStructFieldHeadOmitEmptyMap,
		opStructFieldHeadOmitEmptyMapLoad,
		opStructFieldHeadOmitEmptyStruct,
		opStructFieldHeadStringTag,
		opStructFieldHeadIndent,
		opStructFieldHeadSliceIndent,
		opStructFieldHeadArrayIndent,
		opStructFieldHeadMapIndent,
		opStructFieldHeadMapLoadIndent,
		opStructFieldHeadStructIndent,
		opStructFieldHeadOmitEmptyIndent,
		opStructFieldHeadOmitEmptySliceIndent,
		opStructFieldHeadOmitEmptyArrayIndent,
		opStructFieldHeadOmitEmptyMapIndent,
		opStructFieldHeadOmitEmptyMapLoadIndent,
		opStructFieldHeadOmitEmptyStructIndent,
		opStructFieldHeadStringTagIndent:
		return valueCode.beforeLastCode()
	}
	ctx.decOpcodeIndex()
	return (*opcode)(unsafe.Pointer(fieldCode))
}

func (e *Encoder) structField(ctx *encodeCompileContext, fieldCode *opcode, valueCode *opcode, tag *structTag) *opcode {
	code := (*opcode)(unsafe.Pointer(fieldCode))
	op := e.optimizeStructField(valueCode.op, tag, ctx.withIndent)
	fieldCode.op = op
	switch op {
	case opStructField,
		opStructFieldSlice,
		opStructFieldArray,
		opStructFieldMap,
		opStructFieldMapLoad,
		opStructFieldStruct,
		opStructFieldOmitEmpty,
		opStructFieldOmitEmptySlice,
		opStructFieldOmitEmptyArray,
		opStructFieldOmitEmptyMap,
		opStructFieldOmitEmptyMapLoad,
		opStructFieldOmitEmptyStruct,
		opStructFieldStringTag,
		opStructFieldIndent,
		opStructFieldSliceIndent,
		opStructFieldArrayIndent,
		opStructFieldMapIndent,
		opStructFieldMapLoadIndent,
		opStructFieldStructIndent,
		opStructFieldOmitEmptyIndent,
		opStructFieldOmitEmptySliceIndent,
		opStructFieldOmitEmptyArrayIndent,
		opStructFieldOmitEmptyMapIndent,
		opStructFieldOmitEmptyMapLoadIndent,
		opStructFieldOmitEmptyStructIndent,
		opStructFieldStringTagIndent:
		return valueCode.beforeLastCode()
	}
	ctx.decIndex()
	return code
}

func (e *Encoder) isNotExistsField(head *opcode) bool {
	if head == nil {
		return false
	}
	if head.op != opStructFieldAnonymousHead {
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
	return e.isNotExistsField(head.next)
}

func (e *Encoder) optimizeAnonymousFields(head *opcode) {
	code := head
	var prev *opcode
	for {
		if code.op == opStructEnd || code.op == opStructEndIndent {
			break
		}
		if code.op == opStructField || code.op == opStructFieldIndent {
			codeType := code.next.op.codeType()
			if codeType == codeStructField {
				if e.isNotExistsField(code.next) {
					code.next = code.nextField
					diff := code.next.displayIdx - code.displayIdx
					for i := 0; i < diff; i++ {
						code.next.decOpcodeIndex()
					}
					linkPrevToNextField(prev, code)
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

func (e *Encoder) anonymousStructFieldPairMap(typ *rtype, tags structTags, valueCode *opcode) map[string][]structFieldPair {
	anonymousFields := map[string][]structFieldPair{}
	f := valueCode
	var prevAnonymousField *opcode
	for {
		existsKey := tags.existsKey(f.displayKey)
		op := f.op.headToAnonymousHead()
		if op != f.op {
			if existsKey {
				f.op = opStructFieldAnonymousHead
			} else {
				f.op = op
			}
		} else if f.op == opStructEnd {
			f.op = opStructAnonymousEnd
		} else if existsKey {
			diff := f.nextField.displayIdx - f.displayIdx
			for i := 0; i < diff; i++ {
				f.nextField.decOpcodeIndex()
			}
			linkPrevToNextField(prevAnonymousField, f)
		}

		if f.displayKey == "" {
			if f.nextField == nil {
				break
			}
			prevAnonymousField = f
			f = f.nextField
			continue
		}

		anonymousFields[f.displayKey] = append(anonymousFields[f.displayKey], structFieldPair{
			prevField:   prevAnonymousField,
			curField:    f,
			isTaggedKey: f.isTaggedKey,
		})
		if f.next != nil && f.nextField != f.next && f.next.op.codeType() == codeStructField {
			for k, v := range e.anonymousStructFieldPairMap(typ, tags, f.next) {
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

func (e *Encoder) optimizeConflictAnonymousFields(anonymousFields map[string][]structFieldPair) {
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
						fieldPair.curField.op = opStructFieldAnonymousHead
					} else {
						diff := fieldPair.curField.nextField.displayIdx - fieldPair.curField.displayIdx
						for i := 0; i < diff; i++ {
							fieldPair.curField.nextField.decOpcodeIndex()
						}
						linkPrevToNextField(fieldPair.prevField, fieldPair.curField)
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
						fieldPair.curField.op = opStructFieldAnonymousHead
					} else {
						diff := fieldPair.curField.nextField.displayIdx - fieldPair.curField.displayIdx
						for i := 0; i < diff; i++ {
							fieldPair.curField.nextField.decOpcodeIndex()
						}
						linkPrevToNextField(fieldPair.prevField, fieldPair.curField)
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

func (e *Encoder) compileStruct(ctx *encodeCompileContext, isPtr bool) (*opcode, error) {
	ctx.root = false
	if code := e.compiledCode(ctx); code != nil {
		return code, nil
	}
	typ := ctx.typ
	typeptr := uintptr(unsafe.Pointer(typ))
	compiled := &compiledCode{}
	if ctx.withIndent {
		e.structTypeToCompiledIndentCode[typeptr] = compiled
	} else {
		e.structTypeToCompiledCode[typeptr] = compiled
	}
	// header => code => structField => code => end
	//                        ^          |
	//                        |__________|
	fieldNum := typ.NumField()
	fieldIdx := 0
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
		if isPtr && i == 0 {
			// head field of pointer structure at top level
			// if field type is pointer and implements MarshalJSON or MarshalText,
			// it need to operation of dereference of pointer.
			if field.Type.Kind() == reflect.Ptr &&
				(field.Type.Implements(marshalJSONType) || field.Type.Implements(marshalTextType)) {
				fieldType = rtype_ptrTo(fieldType)
			}
		}
		fieldOpcodeIndex := ctx.opcodeIndex
		fieldPtrIndex := ctx.ptrIndex
		ctx.incIndex()
		valueCode, err := e.compile(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}

		if field.Anonymous {
			for k, v := range e.anonymousStructFieldPairMap(typ, tags, valueCode) {
				anonymousFields[k] = append(anonymousFields[k], v...)
			}
		}
		if fieldNum == 1 && valueCode.op == opPtr {
			// if field number is one and primitive pointer type,
			// it should encode as **not** pointer .
			switch valueCode.next.op {
			case opInt, opInt8, opInt16, opInt32, opInt64,
				opUint, opUint8, opUint16, opUint32, opUint64,
				opFloat32, opFloat64, opBool, opString, opBytes:
				valueCode = valueCode.next
				ctx.decOpcodeIndex()
			}
		}
		key := fmt.Sprintf(`"%s":`, tag.key)

		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		enc.encodeEscapedString(tag.key)
		escapedKey := fmt.Sprintf(`%s:`, string(enc.buf))
		enc.release()
		fieldCode := &opcode{
			typ:          valueCode.typ,
			displayIdx:   fieldOpcodeIndex,
			idx:          opcodeOffset(fieldPtrIndex),
			next:         valueCode,
			indent:       ctx.indent,
			anonymousKey: field.Anonymous,
			key:          []byte(key),
			escapedKey:   []byte(escapedKey),
			isTaggedKey:  tag.isTaggedKey,
			displayKey:   tag.key,
			offset:       field.Offset,
		}
		if fieldIdx == 0 {
			fieldCode.headIdx = fieldCode.idx
			code = e.structHeader(ctx, fieldCode, valueCode, tag)
			head = fieldCode
			prevField = fieldCode
		} else {
			fieldCode.headIdx = head.headIdx
			code.next = fieldCode
			code = e.structField(ctx, fieldCode, valueCode, tag)
			prevField.nextField = fieldCode
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
		ctx.incIndex()
		if ctx.withIndent {
			head.op = opStructFieldHeadIndent
		}
		code = head
	}

	structEndCode.displayIdx = ctx.opcodeIndex
	structEndCode.idx = opcodeOffset(ctx.ptrIndex)
	ctx.incIndex()

	if ctx.withIndent {
		structEndCode.op = opStructEndIndent
	}

	if prevField != nil && prevField.nextField == nil {
		prevField.nextField = structEndCode
	}

	head.end = structEndCode
	code.next = structEndCode

	e.optimizeConflictAnonymousFields(anonymousFields)
	e.optimizeAnonymousFields(head)

	ret := (*opcode)(unsafe.Pointer(head))
	compiled.code = ret

	if ctx.withIndent {
		delete(e.structTypeToCompiledIndentCode, typeptr)
	} else {
		delete(e.structTypeToCompiledCode, typeptr)
	}

	return ret, nil
}

package encoder

import (
	"context"
	"encoding"
	"encoding/json"
	"reflect"
	"sync/atomic"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

type marshalerContext interface {
	MarshalJSON(context.Context) ([]byte, error)
}

var (
	marshalJSONType        = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	marshalJSONContextType = reflect.TypeOf((*marshalerContext)(nil)).Elem()
	marshalTextType        = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	jsonNumberType         = reflect.TypeOf(json.Number(""))
	cachedOpcodeSets       []*OpcodeSet
	cachedOpcodeMap        unsafe.Pointer // map[uintptr]*OpcodeSet
	typeAddr               *runtime.TypeAddr
)

func init() {
	typeAddr = runtime.AnalyzeTypeAddr()
	if typeAddr == nil {
		typeAddr = &runtime.TypeAddr{}
	}
	cachedOpcodeSets = make([]*OpcodeSet, typeAddr.AddrRange>>typeAddr.AddrShift)
}

func loadOpcodeMap() map[uintptr]*OpcodeSet {
	p := atomic.LoadPointer(&cachedOpcodeMap)
	return *(*map[uintptr]*OpcodeSet)(unsafe.Pointer(&p))
}

func storeOpcodeSet(typ uintptr, set *OpcodeSet, m map[uintptr]*OpcodeSet) {
	newOpcodeMap := make(map[uintptr]*OpcodeSet, len(m)+1)
	newOpcodeMap[typ] = set

	for k, v := range m {
		newOpcodeMap[k] = v
	}

	atomic.StorePointer(&cachedOpcodeMap, *(*unsafe.Pointer)(unsafe.Pointer(&newOpcodeMap)))
}

func compileToGetCodeSetSlowPath(typeptr uintptr) (*OpcodeSet, error) {
	opcodeMap := loadOpcodeMap()
	if codeSet, exists := opcodeMap[typeptr]; exists {
		return codeSet, nil
	}

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := *(**runtime.Type)(unsafe.Pointer(&typeptr))

	noescapeKeyCode, err := compile(&compileContext{
		typ:               copiedType,
		structTypeToCode:  map[uintptr]*StructCode{},
		structTypeToCodes: map[uintptr]Opcodes{},
	})
	if err != nil {
		return nil, err
	}
	escapeKeyCode, err := compile(&compileContext{
		typ:               copiedType,
		structTypeToCode:  map[uintptr]*StructCode{},
		structTypeToCodes: map[uintptr]Opcodes{},
		escapeKey:         true,
	})
	if err != nil {
		return nil, err
	}
	noescapeKeyCode = copyOpcode(noescapeKeyCode)
	escapeKeyCode = copyOpcode(escapeKeyCode)
	setTotalLengthToInterfaceOp(noescapeKeyCode)
	setTotalLengthToInterfaceOp(escapeKeyCode)
	interfaceNoescapeKeyCode := copyToInterfaceOpcode(noescapeKeyCode)
	interfaceEscapeKeyCode := copyToInterfaceOpcode(escapeKeyCode)
	codeLength := noescapeKeyCode.TotalLength()
	codeSet := &OpcodeSet{
		Type:                     copiedType,
		NoescapeKeyCode:          noescapeKeyCode,
		EscapeKeyCode:            escapeKeyCode,
		InterfaceNoescapeKeyCode: interfaceNoescapeKeyCode,
		InterfaceEscapeKeyCode:   interfaceEscapeKeyCode,
		CodeLength:               codeLength,
		EndCode:                  ToEndCode(interfaceNoescapeKeyCode),
	}
	storeOpcodeSet(typeptr, codeSet, opcodeMap)
	return codeSet, nil
}

func compile(ctx *compileContext) (*Opcode, error) {
	code, err := type2code(ctx)
	if err != nil {
		return nil, err
	}
	derefctx := *ctx
	newCtx := &derefctx
	codes := code.ToOpcode(newCtx)
	codes.Last().Next = newEndOp(newCtx)
	linkRecursiveCode(newCtx)
	return codes.First(), nil
}

func implementsMarshalJSON(typ *runtime.Type) bool {
	if !implementsMarshalJSONType(typ) {
		return false
	}
	if typ.Kind() != reflect.Ptr {
		return true
	}
	// type kind is reflect.Ptr
	if !implementsMarshalJSONType(typ.Elem()) {
		return true
	}
	// needs to dereference
	return false
}

func implementsMarshalText(typ *runtime.Type) bool {
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

func convertPtrOp(code *Opcode) OpType {
	ptrHeadOp := code.Op.HeadToPtrHead()
	if code.Op != ptrHeadOp {
		if code.PtrNum > 0 {
			// ptr field and ptr head
			code.PtrNum--
		}
		return ptrHeadOp
	}
	switch code.Op {
	case OpInt:
		return OpIntPtr
	case OpUint:
		return OpUintPtr
	case OpFloat32:
		return OpFloat32Ptr
	case OpFloat64:
		return OpFloat64Ptr
	case OpString:
		return OpStringPtr
	case OpBool:
		return OpBoolPtr
	case OpBytes:
		return OpBytesPtr
	case OpNumber:
		return OpNumberPtr
	case OpArray:
		return OpArrayPtr
	case OpSlice:
		return OpSlicePtr
	case OpMap:
		return OpMapPtr
	case OpMarshalJSON:
		return OpMarshalJSONPtr
	case OpMarshalText:
		return OpMarshalTextPtr
	case OpInterface:
		return OpInterfacePtr
	case OpRecursive:
		return OpRecursivePtr
	}
	return code.Op
}

const intSize = 32 << (^uint(0) >> 63)

func optimizeStructHeader(code *Opcode, tag *runtime.StructTag) OpType {
	headType := code.ToHeaderType(tag.IsString)
	if tag.IsOmitEmpty {
		headType = headType.HeadToOmitEmptyHead()
	}
	return headType
}

func optimizeStructField(code *Opcode, tag *runtime.StructTag) OpType {
	fieldType := code.ToFieldType(tag.IsString)
	if tag.IsOmitEmpty {
		fieldType = fieldType.FieldToOmitEmptyField()
	}
	return fieldType
}

func isNilableType(typ *runtime.Type) bool {
	switch typ.Kind() {
	case reflect.Ptr:
		return true
	case reflect.Map:
		return true
	case reflect.Func:
		return true
	default:
		return false
	}
}

func implementsMarshalJSONType(typ *runtime.Type) bool {
	return typ.Implements(marshalJSONType) || typ.Implements(marshalJSONContextType)
}

func isPtrMarshalJSONType(typ *runtime.Type) bool {
	return !implementsMarshalJSONType(typ) && implementsMarshalJSONType(runtime.PtrTo(typ))
}

func isPtrMarshalTextType(typ *runtime.Type) bool {
	return !typ.Implements(marshalTextType) && runtime.PtrTo(typ).Implements(marshalTextType)
}

func linkRecursiveCode(ctx *compileContext) {
	for _, recursive := range *ctx.recursiveCodes {
		typeptr := uintptr(unsafe.Pointer(recursive.Type))
		codes := ctx.structTypeToCodes[typeptr]
		compiled := recursive.Jmp
		compiled.Code = copyOpcode(codes.First())
		code := compiled.Code
		code.End.Next = newEndOp(&compileContext{})
		code.Op = code.Op.PtrHeadToHead()

		beforeLastCode := code.End
		lastCode := beforeLastCode.Next

		totalLength := code.TotalLength()
		lastCode.Idx = uint32((totalLength + 1) * uintptrSize)
		lastCode.ElemIdx = lastCode.Idx + uintptrSize
		lastCode.Length = lastCode.Idx + 2*uintptrSize
		code.End.Next.Op = OpRecursiveEnd

		// extend length to alloc slot for elemIdx + length
		curTotalLength := uintptr(recursive.TotalLength()) + 3
		nextTotalLength := uintptr(totalLength) + 3
		compiled.CurLen = curTotalLength
		compiled.NextLen = nextTotalLength
		compiled.Linked = true
	}
}

func type2code(ctx *compileContext) (Code, error) {
	typ := ctx.typ
	switch {
	case implementsMarshalJSON(typ):
		return compileMarshalJSON(ctx)
	case implementsMarshalText(typ):
		return compileMarshalText(ctx)
	}

	isPtr := false
	orgType := typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		isPtr = true
	}
	switch {
	case implementsMarshalJSON(typ):
		return compileMarshalJSON(ctx)
	case implementsMarshalText(typ):
		return compileMarshalText(ctx)
	}
	switch typ.Kind() {
	case reflect.Slice:
		ctx := ctx.withType(typ)
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := runtime.PtrTo(elem)
			if !implementsMarshalJSONType(p) && !p.Implements(marshalTextType) {
				return compileBytes(ctx, isPtr)
			}
		}
		return compileSlice(ctx)
	case reflect.Map:
		if isPtr {
			return compilePtr(ctx.withType(runtime.PtrTo(typ)))
		}
		return compileMap(ctx.withType(typ))
	case reflect.Struct:
		return compileStruct(ctx.withType(typ), isPtr)
	case reflect.Int:
		return compileInt(ctx.withType(typ), isPtr)
	case reflect.Int8:
		return compileInt8(ctx.withType(typ), isPtr)
	case reflect.Int16:
		return compileInt16(ctx.withType(typ), isPtr)
	case reflect.Int32:
		return compileInt32(ctx.withType(typ), isPtr)
	case reflect.Int64:
		return compileInt64(ctx.withType(typ), isPtr)
	case reflect.Uint, reflect.Uintptr:
		return compileUint(ctx.withType(typ), isPtr)
	case reflect.Uint8:
		return compileUint8(ctx.withType(typ), isPtr)
	case reflect.Uint16:
		return compileUint16(ctx.withType(typ), isPtr)
	case reflect.Uint32:
		return compileUint32(ctx.withType(typ), isPtr)
	case reflect.Uint64:
		return compileUint64(ctx.withType(typ), isPtr)
	case reflect.Float32:
		return compileFloat32(ctx.withType(typ), isPtr)
	case reflect.Float64:
		return compileFloat64(ctx.withType(typ), isPtr)
	case reflect.String:
		return compileString(ctx.withType(typ), isPtr)
	case reflect.Bool:
		return compileBool(ctx.withType(typ), isPtr)
	case reflect.Interface:
		return compileInterface(ctx.withType(typ), isPtr)
	default:
		if isPtr && typ.Implements(marshalTextType) {
			typ = orgType
		}
		return type2codeWithPtr(ctx.withType(typ), isPtr)
	}
}

func type2codeWithPtr(ctx *compileContext, isPtr bool) (Code, error) {
	typ := ctx.typ
	switch {
	case implementsMarshalJSON(typ):
		return compileMarshalJSON(ctx)
	case implementsMarshalText(typ):
		return compileMarshalText(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return compilePtr(ctx)
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := runtime.PtrTo(elem)
			if !implementsMarshalJSONType(p) && !p.Implements(marshalTextType) {
				return compileBytes(ctx, false)
			}
		}
		return compileSlice(ctx)
	case reflect.Array:
		return compileArray(ctx)
	case reflect.Map:
		return compileMap(ctx)
	case reflect.Struct:
		return compileStruct(ctx, isPtr)
	case reflect.Interface:
		return compileInterface(ctx, false)
	case reflect.Int:
		return compileInt(ctx, false)
	case reflect.Int8:
		return compileInt8(ctx, false)
	case reflect.Int16:
		return compileInt16(ctx, false)
	case reflect.Int32:
		return compileInt32(ctx, false)
	case reflect.Int64:
		return compileInt64(ctx, false)
	case reflect.Uint:
		return compileUint(ctx, false)
	case reflect.Uint8:
		return compileUint8(ctx, false)
	case reflect.Uint16:
		return compileUint16(ctx, false)
	case reflect.Uint32:
		return compileUint32(ctx, false)
	case reflect.Uint64:
		return compileUint64(ctx, false)
	case reflect.Uintptr:
		return compileUint(ctx, false)
	case reflect.Float32:
		return compileFloat32(ctx, false)
	case reflect.Float64:
		return compileFloat64(ctx, false)
	case reflect.String:
		return compileString(ctx, false)
	case reflect.Bool:
		return compileBool(ctx, false)
	}
	return nil, &errors.UnsupportedTypeError{Type: runtime.RType2Type(typ)}
}

func compileInt(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: intSize, isPtr: isPtr}, nil
}

func compileInt8(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 8, isPtr: isPtr}, nil
}

func compileInt16(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 16, isPtr: isPtr}, nil
}

func compileInt32(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 32, isPtr: isPtr}, nil
}

func compileInt64(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 64, isPtr: isPtr}, nil
}

func compileUint(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: intSize, isPtr: isPtr}, nil
}

func compileUint8(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 8, isPtr: isPtr}, nil
}

func compileUint16(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 16, isPtr: isPtr}, nil
}

func compileUint32(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 32, isPtr: isPtr}, nil
}

func compileUint64(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 64, isPtr: isPtr}, nil
}

func compileFloat32(ctx *compileContext, isPtr bool) (*FloatCode, error) {
	return &FloatCode{typ: ctx.typ, bitSize: 32, isPtr: isPtr}, nil
}

func compileFloat64(ctx *compileContext, isPtr bool) (*FloatCode, error) {
	return &FloatCode{typ: ctx.typ, bitSize: 64, isPtr: isPtr}, nil
}

func compileString(ctx *compileContext, isPtr bool) (*StringCode, error) {
	return &StringCode{typ: ctx.typ, isPtr: isPtr}, nil
}

func compileBool(ctx *compileContext, isPtr bool) (*BoolCode, error) {
	return &BoolCode{typ: ctx.typ, isPtr: isPtr}, nil
}

func compileIntString(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: intSize, isString: true}, nil
}

func compileInt8String(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 8, isString: true}, nil
}

func compileInt16String(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 16, isString: true}, nil
}

func compileInt32String(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 32, isString: true}, nil
}

func compileInt64String(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 64, isString: true}, nil
}

func compileUintString(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: intSize, isString: true}, nil
}

func compileUint8String(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 8, isString: true}, nil
}

func compileUint16String(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 16, isString: true}, nil
}

func compileUint32String(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 32, isString: true}, nil
}

func compileUint64String(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 64, isString: true}, nil
}

func compileSlice(ctx *compileContext) (*SliceCode, error) {
	elem := ctx.typ.Elem()
	code, err := compileListElem(ctx.withType(elem))
	if err != nil {
		return nil, err
	}
	if code.Type() == CodeTypeStruct {
		structCode := code.(*StructCode)
		structCode.enableIndirect()
	}
	return &SliceCode{typ: ctx.typ, value: code}, nil
}

func compileArray(ctx *compileContext) (*ArrayCode, error) {
	typ := ctx.typ
	elem := typ.Elem()
	code, err := compileListElem(ctx.withType(elem))
	if err != nil {
		return nil, err
	}
	if code.Type() == CodeTypeStruct {
		structCode := code.(*StructCode)
		structCode.enableIndirect()
	}
	return &ArrayCode{typ: ctx.typ, value: code}, nil
}

func compileMap(ctx *compileContext) (*MapCode, error) {
	typ := ctx.typ
	keyCode, err := compileMapKey(ctx.withType(typ.Key()))
	if err != nil {
		return nil, err
	}
	valueCode, err := compileMapValue(ctx.withType(typ.Elem()))
	if err != nil {
		return nil, err
	}
	if valueCode.Type() == CodeTypeStruct {
		structCode := valueCode.(*StructCode)
		structCode.enableIndirect()
	}
	return &MapCode{typ: ctx.typ, key: keyCode, value: valueCode}, nil
}

func compileBytes(ctx *compileContext, isPtr bool) (*BytesCode, error) {
	return &BytesCode{typ: ctx.typ, isPtr: isPtr}, nil
}

func compileInterface(ctx *compileContext, isPtr bool) (*InterfaceCode, error) {
	return &InterfaceCode{typ: ctx.typ, isPtr: isPtr}, nil
}

func compileMarshalJSON(ctx *compileContext) (*MarshalJSONCode, error) {
	return &MarshalJSONCode{typ: ctx.typ}, nil
}

func compileMarshalText(ctx *compileContext) (*MarshalTextCode, error) {
	return &MarshalTextCode{typ: ctx.typ}, nil
}

func compilePtr(ctx *compileContext) (*PtrCode, error) {
	code, err := type2codeWithPtr(ctx.withType(ctx.typ.Elem()), true)
	if err != nil {
		return nil, err
	}
	ptr, ok := code.(*PtrCode)
	if ok {
		return &PtrCode{typ: ctx.typ, value: ptr.value, ptrNum: ptr.ptrNum + 1}, nil
	}
	return &PtrCode{typ: ctx.typ, value: code, ptrNum: 1}, nil
}

func compileListElem(ctx *compileContext) (Code, error) {
	typ := ctx.typ
	switch {
	case isPtrMarshalJSONType(typ):
		return compileMarshalJSON(ctx)
	case !typ.Implements(marshalTextType) && runtime.PtrTo(typ).Implements(marshalTextType):
		return compileMarshalText(ctx)
	case typ.Kind() == reflect.Map:
		return compilePtr(ctx.withType(runtime.PtrTo(typ)))
	default:
		code, err := type2codeWithPtr(ctx, false)
		if err != nil {
			return nil, err
		}
		ptr, ok := code.(*PtrCode)
		if ok {
			if ptr.value.Type() == CodeTypeMap {
				ptr.ptrNum++
			}
		}
		return code, nil
	}
}

func compileMapKey(ctx *compileContext) (Code, error) {
	typ := ctx.typ
	switch {
	case implementsMarshalJSON(typ):
		return compileMarshalJSON(ctx)
	case implementsMarshalText(typ):
		return compileMarshalText(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return compilePtr(ctx)
	case reflect.String:
		return compileString(ctx, false)
	case reflect.Int:
		return compileIntString(ctx)
	case reflect.Int8:
		return compileInt8String(ctx)
	case reflect.Int16:
		return compileInt16String(ctx)
	case reflect.Int32:
		return compileInt32String(ctx)
	case reflect.Int64:
		return compileInt64String(ctx)
	case reflect.Uint:
		return compileUintString(ctx)
	case reflect.Uint8:
		return compileUint8String(ctx)
	case reflect.Uint16:
		return compileUint16String(ctx)
	case reflect.Uint32:
		return compileUint32String(ctx)
	case reflect.Uint64:
		return compileUint64String(ctx)
	case reflect.Uintptr:
		return compileUintString(ctx)
	}
	return nil, &errors.UnsupportedTypeError{Type: runtime.RType2Type(typ)}
}

func compileMapValue(ctx *compileContext) (Code, error) {
	switch ctx.typ.Kind() {
	case reflect.Map:
		return compilePtr(ctx.withType(runtime.PtrTo(ctx.typ)))
	default:
		code, err := type2codeWithPtr(ctx, false)
		if err != nil {
			return nil, err
		}
		ptr, ok := code.(*PtrCode)
		if ok {
			if ptr.value.Type() == CodeTypeMap {
				ptr.ptrNum++
			}
		}
		return code, nil
	}
}

func compileStruct(ctx *compileContext, isPtr bool) (*StructCode, error) {
	typ := ctx.typ
	typeptr := uintptr(unsafe.Pointer(typ))
	if code, exists := ctx.structTypeToCode[typeptr]; exists {
		derefCode := *code
		derefCode.isRecursive = true
		return &derefCode, nil
	}
	indirect := runtime.IfaceIndir(typ)
	code := &StructCode{typ: typ, isPtr: isPtr, isIndirect: indirect}
	ctx.structTypeToCode[typeptr] = code

	fieldNum := typ.NumField()
	tags := typeToStructTags(typ)
	fields := []*StructFieldCode{}
	for i, tag := range tags {
		isOnlyOneFirstField := i == 0 && fieldNum == 1
		field, err := code.compileStructField(ctx, tag, isPtr, isOnlyOneFirstField)
		if err != nil {
			return nil, err
		}
		if field.isAnonymous {
			structCode := field.getAnonymousStruct()
			if structCode != nil {
				structCode.removeFieldsByTags(tags)
				if isAssignableIndirect(field, isPtr) {
					if indirect {
						structCode.isIndirect = true
					} else {
						structCode.isIndirect = false
					}
				}
			}
		} else {
			structCode := field.getStruct()
			if structCode != nil {
				if indirect {
					// if parent is indirect type, set child indirect property to true
					structCode.isIndirect = true
				} else {
					// if parent is not indirect type, set child indirect property to false.
					// but if parent's indirect is false and isPtr is true, then indirect must be true.
					// Do this only if indirectConversion is enabled at the end of compileStruct.
					structCode.isIndirect = false
				}
			}
		}
		fields = append(fields, field)
	}
	fieldMap := getFieldMap(fields)
	duplicatedFieldMap := getDuplicatedFieldMap(fieldMap)
	code.fields = filteredDuplicatedFields(fields, duplicatedFieldMap)
	if !code.disableIndirectConversion && !indirect && isPtr {
		code.enableIndirect()
	}
	delete(ctx.structTypeToCode, typeptr)
	return code, nil
}

func getFieldMap(fields []*StructFieldCode) map[string][]*StructFieldCode {
	fieldMap := map[string][]*StructFieldCode{}
	for _, field := range fields {
		if field.isAnonymous {
			for k, v := range getAnonymousFieldMap(field) {
				fieldMap[k] = append(fieldMap[k], v...)
			}
			continue
		}
		fieldMap[field.key] = append(fieldMap[field.key], field)
	}
	return fieldMap
}

func getAnonymousFieldMap(field *StructFieldCode) map[string][]*StructFieldCode {
	fieldMap := map[string][]*StructFieldCode{}
	structCode := field.getAnonymousStruct()
	if structCode == nil || structCode.isRecursive {
		fieldMap[field.key] = append(fieldMap[field.key], field)
		return fieldMap
	}
	for k, v := range getFieldMapFromAnonymousParent(structCode.fields) {
		fieldMap[k] = append(fieldMap[k], v...)
	}
	return fieldMap
}

func getFieldMapFromAnonymousParent(fields []*StructFieldCode) map[string][]*StructFieldCode {
	fieldMap := map[string][]*StructFieldCode{}
	for _, field := range fields {
		if field.isAnonymous {
			for k, v := range getAnonymousFieldMap(field) {
				// Do not handle tagged key when embedding more than once
				for _, vv := range v {
					vv.isTaggedKey = false
				}
				fieldMap[k] = append(fieldMap[k], v...)
			}
			continue
		}
		fieldMap[field.key] = append(fieldMap[field.key], field)
	}
	return fieldMap
}

func getDuplicatedFieldMap(fieldMap map[string][]*StructFieldCode) map[*StructFieldCode]struct{} {
	duplicatedFieldMap := map[*StructFieldCode]struct{}{}
	for _, fields := range fieldMap {
		if len(fields) == 1 {
			continue
		}
		if isTaggedKeyOnly(fields) {
			for _, field := range fields {
				if field.isTaggedKey {
					continue
				}
				duplicatedFieldMap[field] = struct{}{}
			}
		} else {
			for _, field := range fields {
				duplicatedFieldMap[field] = struct{}{}
			}
		}
	}
	return duplicatedFieldMap
}

func filteredDuplicatedFields(fields []*StructFieldCode, duplicatedFieldMap map[*StructFieldCode]struct{}) []*StructFieldCode {
	filteredFields := make([]*StructFieldCode, 0, len(fields))
	for _, field := range fields {
		if field.isAnonymous {
			structCode := field.getAnonymousStruct()
			if structCode != nil && !structCode.isRecursive {
				structCode.fields = filteredDuplicatedFields(structCode.fields, duplicatedFieldMap)
				if len(structCode.fields) > 0 {
					filteredFields = append(filteredFields, field)
				}
				continue
			}
		}
		if _, exists := duplicatedFieldMap[field]; exists {
			continue
		}
		filteredFields = append(filteredFields, field)
	}
	return filteredFields
}

func isTaggedKeyOnly(fields []*StructFieldCode) bool {
	var taggedKeyFieldCount int
	for _, field := range fields {
		if field.isTaggedKey {
			taggedKeyFieldCount++
		}
	}
	return taggedKeyFieldCount == 1
}

func typeToStructTags(typ *runtime.Type) runtime.StructTags {
	tags := runtime.StructTags{}
	fieldNum := typ.NumField()
	for i := 0; i < fieldNum; i++ {
		field := typ.Field(i)
		if runtime.IsIgnoredStructField(field) {
			continue
		}
		tags = append(tags, runtime.StructTagFromField(field))
	}
	return tags
}

// *struct{ field T } => struct { field *T }
// func (*T) MarshalJSON() ([]byte, error)
func isMovePointerPositionFromHeadToFirstMarshalJSONFieldCase(typ *runtime.Type, isIndirectSpecialCase bool) bool {
	return isIndirectSpecialCase && !isNilableType(typ) && isPtrMarshalJSONType(typ)
}

// *struct{ field T } => struct { field *T }
// func (*T) MarshalText() ([]byte, error)
func isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(typ *runtime.Type, isIndirectSpecialCase bool) bool {
	return isIndirectSpecialCase && !isNilableType(typ) && isPtrMarshalTextType(typ)
}

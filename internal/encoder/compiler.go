package encoder

import (
	"context"
	"encoding"
	"encoding/json"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

func CompileToGetCodeSet(ctx *RuntimeContext, typeptr uintptr) (*OpcodeSet, error) {
	initEncoder()
	if typeptr > typeAddr.MaxTypeAddr || typeptr < typeAddr.BaseTypeAddr {
		codeSet, err := compileToGetCodeSetSlowPath(typeptr)
		if err != nil {
			return nil, err
		}
		return getFilteredCodeSetIfNeeded(ctx, codeSet)
	}
	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if codeSet := cachedOpcodeSets[index].Load(); codeSet != nil {
		filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
		if err != nil {
			return nil, err
		}
		return filtered, nil
	}
	codeSet, err := newCompiler().compile(typeptr)
	if err != nil {
		return nil, err
	}
	filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
	if err != nil {
		return nil, err
	}
	cachedOpcodeSets[index].Store(codeSet)
	return filtered, nil
}

// compileToGetUnfilteredCodeSet is CompileToGetCodeSet without the filter by the field query of the context.
func compileToGetUnfilteredCodeSet(typeptr uintptr) (*OpcodeSet, error) {
	initEncoder()
	if typeptr > typeAddr.MaxTypeAddr || typeptr < typeAddr.BaseTypeAddr {
		return compileToGetCodeSetSlowPath(typeptr)
	}
	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if codeSet := cachedOpcodeSets[index].Load(); codeSet != nil {
		return codeSet, nil
	}
	codeSet, err := newCompiler().compile(typeptr)
	if err != nil {
		return nil, err
	}
	cachedOpcodeSets[index].Store(codeSet)
	return codeSet, nil
}

type marshalerContext interface {
	MarshalJSON(context.Context) ([]byte, error)
}

var (
	marshalJSONType        = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	marshalJSONContextType = reflect.TypeOf((*marshalerContext)(nil)).Elem()
	marshalTextType        = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	jsonNumberType         = reflect.TypeOf(json.Number(""))
	cachedOpcodeSets       []atomic.Pointer[OpcodeSet]
	cachedOpcodeMap        unsafe.Pointer // map[uintptr]*OpcodeSet
	typeAddr               *runtime.TypeAddr
	initEncoderOnce        sync.Once
)

func initEncoder() {
	initEncoderOnce.Do(func() {
		typeAddr = runtime.AnalyzeTypeAddr()
		if typeAddr == nil {
			typeAddr = &runtime.TypeAddr{}
		}
		cachedOpcodeSets = make([]atomic.Pointer[OpcodeSet], typeAddr.AddrRange>>typeAddr.AddrShift+1)
	})
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
	codeSet, err := newCompiler().compile(typeptr)
	if err != nil {
		return nil, err
	}
	storeOpcodeSet(typeptr, codeSet, opcodeMap)
	return codeSet, nil
}

func getFilteredCodeSetIfNeeded(ctx *RuntimeContext, codeSet *OpcodeSet) (*OpcodeSet, error) {
	if (ctx.Option.Flag & ContextOption) == 0 {
		return codeSet, nil
	}
	query := FieldQueryFromContext(ctx.Option.Context)
	if query == nil {
		return codeSet, nil
	}
	ctx.Option.Flag |= FieldQueryOption
	cacheCodeSet := codeSet.getQueryCache(query.Hash())
	if cacheCodeSet != nil {
		return cacheCodeSet, nil
	}
	queryCodeSet, err := newCompiler().codeToOpcodeSet(codeSet.Type, codeSet.Code.Filter(query))
	if err != nil {
		return nil, err
	}
	codeSet.setQueryCache(query.Hash(), queryCodeSet)
	return queryCodeSet, nil
}

type Compiler struct {
	structTypeToCode map[uintptr]*StructCode
}

func newCompiler() *Compiler {
	return &Compiler{
		structTypeToCode: map[uintptr]*StructCode{},
	}
}

func (c *Compiler) compile(typeptr uintptr) (*OpcodeSet, error) {
	// noescape trick for header.typ ( reflect.*rtype )
	typ := runtime.TypeOfPtr(*(*unsafe.Pointer)(unsafe.Pointer(&typeptr)))
	if typ.Kind() == reflect.Ptr {
		// a pointer is the address of a value, so the opcodes are the ones of the value it points to.
		code, err := c.pointeeCode(typ.Elem())
		if err != nil {
			return nil, err
		}
		return c.codeToOpcodeSet(typ, code)
	}
	code, err := c.valueCode(typ)
	if err != nil {
		return nil, err
	}
	return c.codeToOpcodeSet(typ, code)
}

// pointeeCode returns the code for the value which a pointer points to.
// It is addressable, as an element of a slice is.
func (c *Compiler) pointeeCode(typ reflect.Type) (Code, error) {
	code, err := c.listElemCode(typ)
	if err != nil {
		return nil, err
	}
	if code.Kind() == CodeKindStruct {
		structCode := code.(*StructCode)
		structCode.enableIndirect()
	}
	return code, nil
}

// valueCode returns the code for a value which is not addressable and whose address is given to the opcode:
// the value passed to Marshal and the value held by an interface value.
// It is the same as the value of a map, which is not addressable either.
func (c *Compiler) valueCode(typ reflect.Type) (Code, error) {
	code, err := c.mapValueCode(typ)
	if err != nil {
		return nil, err
	}
	if code.Kind() == CodeKindStruct {
		structCode := code.(*StructCode)
		structCode.enableIndirect()
	}
	return code, nil
}

func (c *Compiler) codeToOpcodeSet(typ reflect.Type, code Code) (*OpcodeSet, error) {
	noescapeKeyCode, err := c.codeToOpcode(&compileContext{
		structTypeToCodes: map[uintptr]Opcodes{},
		recursiveCodes:    &Opcodes{},
	}, typ, code)
	if err != nil {
		return nil, err
	}
	if err := noescapeKeyCode.Validate(); err != nil {
		return nil, err
	}
	escapeKeyCode, err := c.codeToOpcode(&compileContext{
		structTypeToCodes: map[uintptr]Opcodes{},
		recursiveCodes:    &Opcodes{},
		escapeKey:         true,
	}, typ, code)
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
	return &OpcodeSet{
		Type:                     typ,
		IfaceIndir:               runtime.IfaceIndir(typ),
		DataWordIsAddr:           runtime.IfaceIndir(typ) || typ.Kind() == reflect.Ptr,
		NoescapeKeyCode:          noescapeKeyCode,
		EscapeKeyCode:            escapeKeyCode,
		InterfaceNoescapeKeyCode: interfaceNoescapeKeyCode,
		InterfaceEscapeKeyCode:   interfaceEscapeKeyCode,
		CodeLength:               codeLength,
		EndCode:                  ToEndCode(interfaceNoescapeKeyCode),
		Code:                     code,
		QueryCache:               map[string]*OpcodeSet{},
	}, nil
}

func (c *Compiler) typeToCodeWithPtr(typ reflect.Type, isPtr bool) (Code, error) {
	switch {
	case c.implementsMarshalJSON(typ):
		return c.marshalJSONCode(typ)
	case c.implementsMarshalText(typ):
		return c.marshalTextCode(typ)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return c.ptrCode(typ)
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := reflect.PointerTo(elem)
			if !c.implementsMarshalJSONType(p) && !p.Implements(marshalTextType) {
				return c.bytesCode(typ, false)
			}
		}
		return c.sliceCode(typ)
	case reflect.Array:
		return c.arrayCode(typ)
	case reflect.Map:
		return c.mapCode(typ)
	case reflect.Struct:
		return c.structCode(typ, isPtr)
	case reflect.Interface:
		return c.interfaceCode(typ, false)
	case reflect.Int:
		return c.intCode(typ, false)
	case reflect.Int8:
		return c.int8Code(typ, false)
	case reflect.Int16:
		return c.int16Code(typ, false)
	case reflect.Int32:
		return c.int32Code(typ, false)
	case reflect.Int64:
		return c.int64Code(typ, false)
	case reflect.Uint:
		return c.uintCode(typ, false)
	case reflect.Uint8:
		return c.uint8Code(typ, false)
	case reflect.Uint16:
		return c.uint16Code(typ, false)
	case reflect.Uint32:
		return c.uint32Code(typ, false)
	case reflect.Uint64:
		return c.uint64Code(typ, false)
	case reflect.Uintptr:
		return c.uintCode(typ, false)
	case reflect.Float32:
		return c.float32Code(typ, false)
	case reflect.Float64:
		return c.float64Code(typ, false)
	case reflect.String:
		return c.stringCode(typ, false)
	case reflect.Bool:
		return c.boolCode(typ, false)
	}
	return nil, &errors.UnsupportedTypeError{Type: typ}
}

const intSize = 32 << (^uint(0) >> 63)

//nolint:unparam
func (c *Compiler) intCode(typ reflect.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: intSize, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) int8Code(typ reflect.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 8, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) int16Code(typ reflect.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 16, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) int32Code(typ reflect.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 32, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) int64Code(typ reflect.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 64, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uintCode(typ reflect.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: intSize, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uint8Code(typ reflect.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 8, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uint16Code(typ reflect.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 16, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uint32Code(typ reflect.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 32, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uint64Code(typ reflect.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 64, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) float32Code(typ reflect.Type, isPtr bool) (*FloatCode, error) {
	return &FloatCode{typ: typ, bitSize: 32, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) float64Code(typ reflect.Type, isPtr bool) (*FloatCode, error) {
	return &FloatCode{typ: typ, bitSize: 64, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) stringCode(typ reflect.Type, isPtr bool) (*StringCode, error) {
	return &StringCode{typ: typ, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) boolCode(typ reflect.Type, isPtr bool) (*BoolCode, error) {
	return &BoolCode{typ: typ, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) intStringCode(typ reflect.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: intSize, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) int8StringCode(typ reflect.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 8, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) int16StringCode(typ reflect.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 16, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) int32StringCode(typ reflect.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 32, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) int64StringCode(typ reflect.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 64, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uintStringCode(typ reflect.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: intSize, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uint8StringCode(typ reflect.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 8, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uint16StringCode(typ reflect.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 16, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uint32StringCode(typ reflect.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 32, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uint64StringCode(typ reflect.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 64, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) bytesCode(typ reflect.Type, isPtr bool) (*BytesCode, error) {
	return &BytesCode{typ: typ, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) interfaceCode(typ reflect.Type, isPtr bool) (*InterfaceCode, error) {
	return &InterfaceCode{typ: typ, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) marshalJSONCode(typ reflect.Type) (*MarshalJSONCode, error) {
	return &MarshalJSONCode{
		typ:                typ,
		isAddrForMarshaler: c.isPtrMarshalJSONType(typ),
		isNilableType:      c.isNilableType(typ),
		isMarshalerContext: typ.Implements(marshalJSONContextType) || reflect.PointerTo(typ).Implements(marshalJSONContextType),
	}, nil
}

//nolint:unparam
func (c *Compiler) marshalTextCode(typ reflect.Type) (*MarshalTextCode, error) {
	return &MarshalTextCode{
		typ:                typ,
		isAddrForMarshaler: c.isPtrMarshalTextType(typ),
		isNilableType:      c.isNilableType(typ),
	}, nil
}

func (c *Compiler) ptrCode(typ reflect.Type) (*PtrCode, error) {
	code, err := c.typeToCodeWithPtr(typ.Elem(), true)
	if err != nil {
		return nil, err
	}
	ptr, ok := code.(*PtrCode)
	if ok {
		return &PtrCode{typ: typ, value: ptr.value, ptrNum: ptr.ptrNum + 1}, nil
	}
	return &PtrCode{typ: typ, value: code, ptrNum: 1}, nil
}

func (c *Compiler) sliceCode(typ reflect.Type) (*SliceCode, error) {
	elem := typ.Elem()
	code, err := c.listElemCode(elem)
	if err != nil {
		return nil, err
	}
	if code.Kind() == CodeKindStruct {
		structCode := code.(*StructCode)
		structCode.enableIndirect()
	}
	return &SliceCode{typ: typ, value: code}, nil
}

func (c *Compiler) arrayCode(typ reflect.Type) (*ArrayCode, error) {
	elem := typ.Elem()
	code, err := c.listElemCode(elem)
	if err != nil {
		return nil, err
	}
	if code.Kind() == CodeKindStruct {
		structCode := code.(*StructCode)
		structCode.enableIndirect()
	}
	return &ArrayCode{typ: typ, value: code}, nil
}

func (c *Compiler) mapCode(typ reflect.Type) (*MapCode, error) {
	keyCode, err := c.mapKeyCode(typ.Key())
	if err != nil {
		return nil, err
	}
	valueCode, err := c.mapValueCode(typ.Elem())
	if err != nil {
		return nil, err
	}
	if valueCode.Kind() == CodeKindStruct {
		structCode := valueCode.(*StructCode)
		structCode.enableIndirect()
	}
	return &MapCode{typ: typ, key: keyCode, value: valueCode}, nil
}

func (c *Compiler) listElemCode(typ reflect.Type) (Code, error) {
	switch {
	case c.isPtrMarshalJSONType(typ):
		// The opcode takes the address of the element, which is the very pointer the marshaler is called with.
		// So the opcode is the one of the pointer type, and neither a copy of the element nor reflect is needed.
		ptrType := reflect.PointerTo(typ)
		return &MarshalJSONCode{
			typ:                ptrType,
			isMarshalerContext: ptrType.Implements(marshalJSONContextType),
		}, nil
	case c.implementsMarshalJSONType(typ):
		return c.marshalJSONCode(typ)
	case !typ.Implements(marshalTextType) && reflect.PointerTo(typ).Implements(marshalTextType):
		return &MarshalTextCode{typ: reflect.PointerTo(typ)}, nil
	case typ.Kind() == reflect.Map:
		return c.ptrCode(reflect.PointerTo(typ))
	default:
		// isPtr was originally used to indicate whether the type of top level is pointer.
		// However, since the slice/array element is a specification that can get the pointer address, explicitly set isPtr to true.
		// See here for related issues: https://github.com/goccy/go-json/issues/370
		code, err := c.typeToCodeWithPtr(typ, true)
		if err != nil {
			return nil, err
		}
		ptr, ok := code.(*PtrCode)
		if ok {
			if ptr.value.Kind() == CodeKindMap {
				ptr.ptrNum++
			}
		}
		return code, nil
	}
}

func (c *Compiler) mapKeyCode(typ reflect.Type) (Code, error) {
	switch {
	case c.implementsMarshalText(typ):
		return c.marshalTextCode(typ)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return c.ptrCode(typ)
	case reflect.String:
		return c.stringCode(typ, false)
	case reflect.Int:
		return c.intStringCode(typ)
	case reflect.Int8:
		return c.int8StringCode(typ)
	case reflect.Int16:
		return c.int16StringCode(typ)
	case reflect.Int32:
		return c.int32StringCode(typ)
	case reflect.Int64:
		return c.int64StringCode(typ)
	case reflect.Uint:
		return c.uintStringCode(typ)
	case reflect.Uint8:
		return c.uint8StringCode(typ)
	case reflect.Uint16:
		return c.uint16StringCode(typ)
	case reflect.Uint32:
		return c.uint32StringCode(typ)
	case reflect.Uint64:
		return c.uint64StringCode(typ)
	case reflect.Uintptr:
		return c.uintStringCode(typ)
	}
	return nil, &errors.UnsupportedTypeError{Type: typ}
}

func (c *Compiler) mapValueCode(typ reflect.Type) (Code, error) {
	switch {
	case typ.Kind() == reflect.Map && !c.implementsMarshalJSON(typ) && !c.implementsMarshalText(typ):
		// a map which has a marshaler is encoded by the marshaler, not as a map.
		return c.ptrCode(reflect.PointerTo(typ))
	default:
		code, err := c.typeToCodeWithPtr(typ, false)
		if err != nil {
			return nil, err
		}
		ptr, ok := code.(*PtrCode)
		if ok {
			if ptr.value.Kind() == CodeKindMap {
				ptr.ptrNum++
			}
		}
		return code, nil
	}
}

func (c *Compiler) structCode(typ reflect.Type, isPtr bool) (*StructCode, error) {
	typeptr := uintptr(runtime.TypePtr(typ))
	if code, exists := c.structTypeToCode[typeptr]; exists {
		derefCode := *code
		derefCode.isRecursive = true
		return &derefCode, nil
	}
	indirect := runtime.IfaceIndir(typ)
	code := &StructCode{typ: typ, isPtr: isPtr, isIndirect: indirect}
	c.structTypeToCode[typeptr] = code

	fieldNum := typ.NumField()
	tags := c.typeToStructTags(typ)
	fields := []*StructFieldCode{}
	for i, tag := range tags {
		isOnlyOneFirstField := i == 0 && fieldNum == 1
		field, err := c.structFieldCode(code, tag, isPtr, isOnlyOneFirstField)
		if err != nil {
			return nil, err
		}
		if field.isAnonymous {
			structCode := field.getAnonymousStruct()
			if structCode != nil {
				structCode.removeFieldsByTags(tags)
				if c.isAssignableIndirect(field, isPtr) {
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
	fieldMap := c.getFieldMap(fields)
	duplicatedFieldMap := c.getDuplicatedFieldMap(fieldMap)
	code.fields = c.filteredDuplicatedFields(fields, duplicatedFieldMap)
	if !code.disableIndirectConversion && !indirect && isPtr {
		code.enableIndirect()
	}
	delete(c.structTypeToCode, typeptr)
	return code, nil
}

func toElemType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (c *Compiler) structFieldCode(structCode *StructCode, tag *runtime.StructTag, isPtr, isOnlyOneFirstField bool) (*StructFieldCode, error) {
	field := tag.Field
	fieldType := field.Type
	isIndirectSpecialCase := isPtr && isOnlyOneFirstField
	fieldCode := &StructFieldCode{
		typ:           fieldType,
		key:           tag.Key,
		tag:           tag,
		offset:        field.Offset,
		isAnonymous:   field.Anonymous && !tag.IsTaggedKey && toElemType(fieldType).Kind() == reflect.Struct,
		isTaggedKey:   tag.IsTaggedKey,
		isNilableType: c.isNilableType(fieldType),
		isNilCheck:    true,
	}
	switch {
	case c.isMovePointerPositionFromHeadToFirstMarshalJSONFieldCase(fieldType, isIndirectSpecialCase):
		code, err := c.marshalJSONCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = c.isNilCheckForAddrMarshaler(tag)
		structCode.isIndirect = false
		structCode.disableIndirectConversion = true
	case c.isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(fieldType, isIndirectSpecialCase):
		code, err := c.marshalTextCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = c.isNilCheckForAddrMarshaler(tag)
		structCode.isIndirect = false
		structCode.disableIndirectConversion = true
	case isPtr && c.isPtrMarshalJSONType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalJSON() ([]byte, error)
		code, err := c.marshalJSONCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = c.isNilCheckForAddrMarshaler(tag)
	case isPtr && c.isPtrMarshalTextType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalText() ([]byte, error)
		code, err := c.marshalTextCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = c.isNilCheckForAddrMarshaler(tag)
	default:
		code, err := c.typeToCodeWithPtr(fieldType, isPtr)
		if err != nil {
			return nil, err
		}
		switch code.Kind() {
		case CodeKindPtr, CodeKindInterface:
			fieldCode.isNextOpPtrType = true
		}
		fieldCode.value = code
	}
	return fieldCode, nil
}

// isNilCheckForAddrMarshaler returns whether a field whose marshaler is called with its address is checked.
//
// The check is what decides that the field is empty for omitempty, by the kind of the value as encoding/json
// does, so it is required for omitempty. Without omitempty it would write null instead of calling the
// marshaler, which a pointer to the field never needs.
func (c *Compiler) isNilCheckForAddrMarshaler(tag *runtime.StructTag) bool {
	return tag.IsOmitEmpty
}

func (c *Compiler) isAssignableIndirect(fieldCode *StructFieldCode, isPtr bool) bool {
	if isPtr {
		return false
	}
	codeType := fieldCode.value.Kind()
	if codeType == CodeKindMarshalJSON {
		return false
	}
	if codeType == CodeKindMarshalText {
		return false
	}
	return true
}

func (c *Compiler) getFieldMap(fields []*StructFieldCode) map[string][]*StructFieldCode {
	fieldMap := map[string][]*StructFieldCode{}
	for _, field := range fields {
		if field.isAnonymous {
			for k, v := range c.getAnonymousFieldMap(field) {
				fieldMap[k] = append(fieldMap[k], v...)
			}
			continue
		}
		fieldMap[field.key] = append(fieldMap[field.key], field)
	}
	return fieldMap
}

func (c *Compiler) getAnonymousFieldMap(field *StructFieldCode) map[string][]*StructFieldCode {
	fieldMap := map[string][]*StructFieldCode{}
	structCode := field.getAnonymousStruct()
	if structCode == nil || structCode.isRecursive {
		fieldMap[field.key] = append(fieldMap[field.key], field)
		return fieldMap
	}
	for k, v := range c.getFieldMapFromAnonymousParent(structCode.fields) {
		fieldMap[k] = append(fieldMap[k], v...)
	}
	return fieldMap
}

func (c *Compiler) getFieldMapFromAnonymousParent(fields []*StructFieldCode) map[string][]*StructFieldCode {
	fieldMap := map[string][]*StructFieldCode{}
	for _, field := range fields {
		if field.isAnonymous {
			for k, v := range c.getAnonymousFieldMap(field) {
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

func (c *Compiler) getDuplicatedFieldMap(fieldMap map[string][]*StructFieldCode) map[*StructFieldCode]struct{} {
	duplicatedFieldMap := map[*StructFieldCode]struct{}{}
	for _, fields := range fieldMap {
		if len(fields) == 1 {
			continue
		}
		if c.isTaggedKeyOnly(fields) {
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

func (c *Compiler) filteredDuplicatedFields(fields []*StructFieldCode, duplicatedFieldMap map[*StructFieldCode]struct{}) []*StructFieldCode {
	filteredFields := make([]*StructFieldCode, 0, len(fields))
	for _, field := range fields {
		if field.isAnonymous {
			structCode := field.getAnonymousStruct()
			if structCode != nil && !structCode.isRecursive {
				structCode.fields = c.filteredDuplicatedFields(structCode.fields, duplicatedFieldMap)
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

func (c *Compiler) isTaggedKeyOnly(fields []*StructFieldCode) bool {
	var taggedKeyFieldCount int
	for _, field := range fields {
		if field.isTaggedKey {
			taggedKeyFieldCount++
		}
	}
	return taggedKeyFieldCount == 1
}

func (c *Compiler) typeToStructTags(typ reflect.Type) runtime.StructTags {
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
func (c *Compiler) isMovePointerPositionFromHeadToFirstMarshalJSONFieldCase(typ reflect.Type, isIndirectSpecialCase bool) bool {
	return isIndirectSpecialCase && !c.isNilableType(typ) && c.isPtrMarshalJSONType(typ)
}

// *struct{ field T } => struct { field *T }
// func (*T) MarshalText() ([]byte, error)
func (c *Compiler) isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(typ reflect.Type, isIndirectSpecialCase bool) bool {
	return isIndirectSpecialCase && !c.isNilableType(typ) && c.isPtrMarshalTextType(typ)
}

func (c *Compiler) implementsMarshalJSON(typ reflect.Type) bool {
	if !c.implementsMarshalJSONType(typ) {
		return false
	}
	if typ.Kind() != reflect.Ptr {
		return true
	}
	// type kind is reflect.Ptr
	if !c.implementsMarshalJSONType(typ.Elem()) {
		return true
	}
	// needs to dereference
	return false
}

func (c *Compiler) implementsMarshalText(typ reflect.Type) bool {
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

func (c *Compiler) isNilableType(typ reflect.Type) bool {
	if !runtime.IfaceIndir(typ) {
		return true
	}
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

func (c *Compiler) implementsMarshalJSONType(typ reflect.Type) bool {
	return typ.Implements(marshalJSONType) || typ.Implements(marshalJSONContextType)
}

func (c *Compiler) isPtrMarshalJSONType(typ reflect.Type) bool {
	return !c.implementsMarshalJSONType(typ) && c.implementsMarshalJSONType(reflect.PointerTo(typ))
}

func (c *Compiler) isPtrMarshalTextType(typ reflect.Type) bool {
	return !typ.Implements(marshalTextType) && reflect.PointerTo(typ).Implements(marshalTextType)
}

func (c *Compiler) codeToOpcode(ctx *compileContext, typ reflect.Type, code Code) (*Opcode, error) {
	codes := code.ToOpcode(ctx)
	// the first opcode takes the address of the value, as the one of the value of a map does.
	codes.First().Flags |= IndirectFlags
	codes.Last().Next = newEndOp(ctx, typ)
	if err := c.linkRecursiveCode(ctx); err != nil {
		return nil, err
	}
	return codes.First(), nil
}

func (c *Compiler) linkRecursiveCode(ctx *compileContext) error {
	type recursiveTarget struct {
		typeptr  uintptr
		embedded bool
	}
	recursiveCodes := map[recursiveTarget]*CompiledCode{}
	// the recursive codes may increase while they are linked, so the length is evaluated every time.
	for i := 0; i < len(*ctx.recursiveCodes); i++ {
		recursive := (*ctx.recursiveCodes)[i]
		target := recursiveTarget{typeptr: uintptr(recursive.Type), embedded: recursive.Jmp.Embedded}
		typeptr := target.typeptr
		if recursiveCode, ok := recursiveCodes[target]; ok {
			*recursive.Jmp = *recursiveCode
			continue
		}
		codes, exists := ctx.structTypeToCodes[typeptr]
		if target.embedded || !exists {
			// structTypeToCodes has the opcodes of the struct itself, with the braces and the check of nil.
			// - A recursive struct which is embedded jumps to the opcodes only of the fields.
			// - A struct which has been compiled only as an embedded struct is not in structTypeToCodes.
			// In both cases the opcodes to jump to are compiled here.
			structCode, err := c.structCode(runtime.TypeOfPtr(recursive.Type), false)
			if err != nil {
				return err
			}
			structCode.enableIndirect()
			if target.embedded {
				codes = structCode.ToAnonymousOpcode(ctx)
			} else {
				codes = structCode.ToOpcode(ctx)
			}
			// the opcodes are copied up to the end, as the ones in the code which jumps to them are.
			codes.Last().Next = newEndOp(ctx, runtime.TypeOfPtr(recursive.Type))
		}

		code := copyOpcode(codes.First())
		code.Op = code.Op.PtrHeadToHead()
		lastCode := newEndOp(&compileContext{}, runtime.TypeOfPtr(recursive.Type))
		lastCode.Op = OpRecursiveEnd

		// OpRecursiveEnd must set before call TotalLength
		code.End.Next = lastCode

		totalLength := code.TotalLength()

		// Idx, ElemIdx, Length must set after call TotalLength
		lastCode.Idx = uint32((totalLength + 1) * uintptrSize)
		lastCode.ElemIdx = lastCode.Idx + uintptrSize
		lastCode.Length = lastCode.Idx + 2*uintptrSize

		// An interface in the recursive code allocates its frame after the frame it is in, which is
		// the one of the recursive code, not of the code which jumps to it. The length must include
		// the slots of OpRecursiveEnd, so it is set after they are decided.
		setTotalLengthToInterfaceOp(code)

		// extend length to alloc slot for elemIdx + length
		curTotalLength := uintptr(recursive.TotalLength()) + 3
		nextTotalLength := uintptr(totalLength) + 3

		compiled := recursive.Jmp
		compiled.Code = code
		compiled.CurLen = curTotalLength
		compiled.NextLen = nextTotalLength
		compiled.Linked = true

		recursiveCodes[target] = compiled
	}
	return nil
}

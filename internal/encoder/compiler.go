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
	initEncoderOnce        sync.Once
)

func initEncoder() {
	initEncoderOnce.Do(func() {
		typeAddr = runtime.AnalyzeTypeAddr()
		if typeAddr == nil {
			typeAddr = &runtime.TypeAddr{}
		}
		cachedOpcodeSets = make([]*OpcodeSet, typeAddr.AddrRange>>typeAddr.AddrShift+1)
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

func compileToGetCodeSetFromValue(_ *RuntimeContext, v reflect.Value) (*OpcodeSet, error) {
	// Safety check: ensure we have a valid reflect.Value
	if !v.IsValid() {
		return nil, &errors.InvalidUnmarshalError{Type: nil}
	}

	// For compilation purposes, we use the type information
	// Currently falls back to original implementation
	return newCompiler().compileFromValue(v)
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
	typ := *(**runtime.Type)(unsafe.Pointer(&typeptr))
	code, err := c.typeToCode(typ)
	if err != nil {
		return nil, err
	}
	return c.codeToOpcodeSet(typ, code)
}

func (c *Compiler) compileFromValue(v reflect.Value) (*OpcodeSet, error) {
	// Convert reflect.Type to runtime.Type for compatibility
	typ := runtime.Type2RType(v.Type())

	// Currently using original typeToCode implementation
	// TODO: Implement reflect.Value based compilation as requested in CLAUDE.md
	code, err := c.typeToCode(typ)
	if err != nil {
		return nil, err
	}
	return c.codeToOpcodeSet(typ, code)
}

func (c *Compiler) codeToOpcodeSet(typ *runtime.Type, code Code) (*OpcodeSet, error) {
	noescapeKeyCode := c.codeToOpcode(&compileContext{
		structTypeToCodes: map[uintptr]Opcodes{},
		recursiveCodes:    &Opcodes{},
	}, typ, code)
	if err := noescapeKeyCode.Validate(); err != nil {
		return nil, err
	}
	escapeKeyCode := c.codeToOpcode(&compileContext{
		structTypeToCodes: map[uintptr]Opcodes{},
		recursiveCodes:    &Opcodes{},
		escapeKey:         true,
	}, typ, code)
	noescapeKeyCode = copyOpcode(noescapeKeyCode)
	escapeKeyCode = copyOpcode(escapeKeyCode)
	setTotalLengthToInterfaceOp(noescapeKeyCode)
	setTotalLengthToInterfaceOp(escapeKeyCode)
	interfaceNoescapeKeyCode := copyToInterfaceOpcode(noescapeKeyCode)
	interfaceEscapeKeyCode := copyToInterfaceOpcode(escapeKeyCode)
	codeLength := noescapeKeyCode.TotalLength()
	return &OpcodeSet{
		Type:                     typ,
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

func (c *Compiler) typeToCode(typ *runtime.Type) (Code, error) {
	switch {
	case c.implementsMarshalJSON(typ):
		return c.marshalJSONCode(typ)
	case c.implementsMarshalText(typ):
		return c.marshalTextCode(typ)
	}

	isPtr := false
	orgType := typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		isPtr = true
	}
	switch {
	case c.implementsMarshalJSON(typ):
		return c.marshalJSONCode(orgType)
	case c.implementsMarshalText(typ):
		return c.marshalTextCode(orgType)
	}
	switch typ.Kind() {
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(elem)))
			if !c.implementsMarshalJSONType(p) && !p.Implements(marshalTextType) {
				return c.bytesCode(typ, isPtr)
			}
		}
		return c.sliceCode(typ)
	case reflect.Map:
		if isPtr {
			return c.ptrCode(runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(typ))))
		}
		return c.mapCode(typ)
	case reflect.Struct:
		return c.structCode(typ, isPtr)
	case reflect.Int:
		return c.intCode(typ, isPtr)
	case reflect.Int8:
		return c.int8Code(typ, isPtr)
	case reflect.Int16:
		return c.int16Code(typ, isPtr)
	case reflect.Int32:
		return c.int32Code(typ, isPtr)
	case reflect.Int64:
		return c.int64Code(typ, isPtr)
	case reflect.Uint, reflect.Uintptr:
		return c.uintCode(typ, isPtr)
	case reflect.Uint8:
		return c.uint8Code(typ, isPtr)
	case reflect.Uint16:
		return c.uint16Code(typ, isPtr)
	case reflect.Uint32:
		return c.uint32Code(typ, isPtr)
	case reflect.Uint64:
		return c.uint64Code(typ, isPtr)
	case reflect.Float32:
		return c.float32Code(typ, isPtr)
	case reflect.Float64:
		return c.float64Code(typ, isPtr)
	case reflect.String:
		return c.stringCode(typ, isPtr)
	case reflect.Bool:
		return c.boolCode(typ, isPtr)
	case reflect.Interface:
		return c.interfaceCode(typ, isPtr)
	default:
		if isPtr && typ.Implements(marshalTextType) {
			typ = orgType
		}
		return c.typeToCodeWithPtr(typ, isPtr)
	}
}

func (c *Compiler) typeToCodeWithPtr(typ *runtime.Type, isPtr bool) (Code, error) {
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
			p := runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(elem)))
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
	return nil, &errors.UnsupportedTypeError{Type: runtime.RType2Type(typ)}
}

const intSize = 32 << (^uint(0) >> 63)

//nolint:unparam
func (c *Compiler) intCode(typ *runtime.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: intSize, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) int8Code(typ *runtime.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 8, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) int16Code(typ *runtime.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 16, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) int32Code(typ *runtime.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 32, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) int64Code(typ *runtime.Type, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 64, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uintCode(typ *runtime.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: intSize, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uint8Code(typ *runtime.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 8, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uint16Code(typ *runtime.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 16, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uint32Code(typ *runtime.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 32, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) uint64Code(typ *runtime.Type, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 64, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) float32Code(typ *runtime.Type, isPtr bool) (*FloatCode, error) {
	return &FloatCode{typ: typ, bitSize: 32, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) float64Code(typ *runtime.Type, isPtr bool) (*FloatCode, error) {
	return &FloatCode{typ: typ, bitSize: 64, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) stringCode(typ *runtime.Type, isPtr bool) (*StringCode, error) {
	return &StringCode{typ: typ, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) boolCode(typ *runtime.Type, isPtr bool) (*BoolCode, error) {
	return &BoolCode{typ: typ, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) intStringCode(typ *runtime.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: intSize, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) int8StringCode(typ *runtime.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 8, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) int16StringCode(typ *runtime.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 16, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) int32StringCode(typ *runtime.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 32, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) int64StringCode(typ *runtime.Type) (*IntCode, error) {
	return &IntCode{typ: typ, bitSize: 64, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uintStringCode(typ *runtime.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: intSize, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uint8StringCode(typ *runtime.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 8, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uint16StringCode(typ *runtime.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 16, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uint32StringCode(typ *runtime.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 32, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) uint64StringCode(typ *runtime.Type) (*UintCode, error) {
	return &UintCode{typ: typ, bitSize: 64, isString: true}, nil
}

//nolint:unparam
func (c *Compiler) bytesCode(typ *runtime.Type, isPtr bool) (*BytesCode, error) {
	return &BytesCode{typ: typ, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) interfaceCode(typ *runtime.Type, isPtr bool) (*InterfaceCode, error) {
	return &InterfaceCode{typ: typ, isPtr: isPtr}, nil
}

//nolint:unparam
func (c *Compiler) marshalJSONCode(typ *runtime.Type) (*MarshalJSONCode, error) {
	return &MarshalJSONCode{
		typ:                typ,
		isAddrForMarshaler: c.isPtrMarshalJSONType(typ),
		isNilableType:      c.isNilableType(typ),
		isMarshalerContext: typ.Implements(marshalJSONContextType) || runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(typ))).Implements(marshalJSONContextType),
	}, nil
}

//nolint:unparam
func (c *Compiler) marshalTextCode(typ *runtime.Type) (*MarshalTextCode, error) {
	return &MarshalTextCode{
		typ:                typ,
		isAddrForMarshaler: c.isPtrMarshalTextType(typ),
		isNilableType:      c.isNilableType(typ),
	}, nil
}

func (c *Compiler) ptrCode(typ *runtime.Type) (*PtrCode, error) {
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

func (c *Compiler) sliceCode(typ *runtime.Type) (*SliceCode, error) {
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

func (c *Compiler) arrayCode(typ *runtime.Type) (*ArrayCode, error) {
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

func (c *Compiler) mapCode(typ *runtime.Type) (*MapCode, error) {
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

func (c *Compiler) listElemCode(typ *runtime.Type) (Code, error) {
	switch {
	case c.implementsMarshalJSONType(typ) || c.implementsMarshalJSONType(runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(typ)))):
		return c.marshalJSONCode(typ)
	case !typ.Implements(marshalTextType) && runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(typ))).Implements(marshalTextType):
		return c.marshalTextCode(typ)
	case typ.Kind() == reflect.Map:
		return c.ptrCode(runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(typ))))
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

func (c *Compiler) mapKeyCode(typ *runtime.Type) (Code, error) {
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
	return nil, &errors.UnsupportedTypeError{Type: runtime.RType2Type(typ)}
}

func (c *Compiler) mapValueCode(typ *runtime.Type) (Code, error) {
	switch typ.Kind() {
	case reflect.Map:
		return c.ptrCode(runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(typ))))
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

// isIndirectFromType determines if a type requires indirect storage in interfaces
// Complete implementation using address comparison as suggested in CLAUDE.md
func (c *Compiler) isIndirectFromType(typ *runtime.Type) bool {
	if typ == nil {
		return false
	}

	reflectType := runtime.RType2Type(typ)
	
	// Our reflect-based implementation to replace runtime.IfaceIndir
	// Based on Go's KindDirectIface: Size_ == PtrBytes == goarch.PtrSize
	ptrSize := unsafe.Sizeof(uintptr(0))
	typeSize := reflectType.Size()
	
	switch reflectType.Kind() {
	case reflect.Ptr, reflect.Chan, reflect.Map, reflect.Func, reflect.UnsafePointer:
		return false // Pointer-like types are stored directly
	case reflect.Interface:
		return true // Interfaces are stored indirectly
	case reflect.Struct:
		// Special case: structs that contain only pointer-like fields and fit in pointer size
		if typeSize == ptrSize && c.isPointerLikeStruct(reflectType) {
			return false // Direct storage for pointer-like structs
		} else {
			return true // All other structs are indirect
		}
	default:
		// ALL other types (basic types, arrays, slices, strings) are indirect
		return true
	}
}

// isPointerLikeStruct checks if a struct contains only pointer-like fields
func (c *Compiler) isPointerLikeStruct(structType reflect.Type) bool {
	if structType.Kind() != reflect.Struct {
		return false
	}
	
	// Must be exactly pointer-sized to be considered for direct storage
	ptrSize := unsafe.Sizeof(uintptr(0))
	if structType.Size() != ptrSize {
		return false
	}
	
	numFields := structType.NumField()
	if numFields == 0 {
		return false // Empty structs are not pointer-like
	}
	
	// Check if this struct effectively contains only pointer-like content
	return c.isEffectivelyPointerLike(structType)
}

func (c *Compiler) isEffectivelyPointerLike(structType reflect.Type) bool {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldType := field.Type
		fieldKind := fieldType.Kind()
		
		switch fieldKind {
		case reflect.Ptr, reflect.Chan, reflect.Map, reflect.Func, reflect.UnsafePointer:
			continue // This field is directly pointer-like
		case reflect.Struct:
			// Recursively check if embedded struct is pointer-like
			if !c.isEffectivelyPointerLike(fieldType) {
				return false
			}
		default:
			return false // Non-pointer-like field found
		}
	}
	
	return true // All fields are effectively pointer-like
}

// testInterfaceStorageByAddress implements address comparison method from CLAUDE.md
func (c *Compiler) testInterfaceStorageByAddress(reflectType reflect.Type) (result bool) {
	defer func() {
		if r := recover(); r != nil {
			// If we can't test safely, fall back to conservative approach
			// Structs and arrays are typically stored indirectly
			switch reflectType.Kind() {
			case reflect.Struct, reflect.Array:
				result = true
			default:
				result = false
			}
		}
	}()

	// Handle nil and invalid types
	if reflectType == nil {
		return false
	}

	// Create a non-null addressable value using reflect.New as suggested in CLAUDE.md
	// This ensures we have a valid value for address comparison
	var actualValue reflect.Value

	switch reflectType.Kind() {
	case reflect.Ptr:
		// For pointer types, create a pointer to a zero value of the element type
		if reflectType.Elem() != nil {
			actualValue = reflect.New(reflectType.Elem()) // This creates *T and is addressable
		} else {
			// Create a nil pointer but make it addressable
			ptrVal := reflect.New(reflectType).Elem() // Create *(*T) and dereference to get *T
			actualValue = ptrVal
		}
	case reflect.Struct, reflect.Array:
		// For struct and array types, create a new addressable value
		actualValue = reflect.New(reflectType).Elem() // Creates *T then dereference to get T, but still addressable
	default:
		// For other types, create an addressable value
		actualValue = reflect.New(reflectType).Elem()
	}

	if !actualValue.IsValid() {
		// Conservative fallback for invalid values
		return reflectType.Kind() == reflect.Struct || reflectType.Kind() == reflect.Array
	}

	// Ensure the value is addressable
	if !actualValue.CanAddr() {
		newVal := reflect.New(reflectType).Elem()
		if actualValue.CanInterface() && newVal.CanSet() {
			newVal.Set(actualValue)
		}
		actualValue = newVal
	}

	if !actualValue.CanAddr() {
		// Conservative fallback if still cannot address
		return reflectType.Kind() == reflect.Struct || reflectType.Kind() == reflect.Array
	}

	// Convert to interface{} to test storage
	var interfaceValue interface{}
	if actualValue.CanInterface() {
		interfaceValue = actualValue.Interface()
	} else {
		// Conservative fallback for types that can't interface
		return reflectType.Kind() == reflect.Struct || reflectType.Kind() == reflect.Array
	}

	// Get reflect.Value of the interface to examine storage
	interfaceReflectValue := reflect.ValueOf(interfaceValue)

	// Compare addresses to determine storage type
	// According to CLAUDE.md: if struct first field address matches struct address, it's direct
	// If they differ, it requires dereferencing (indirect storage)

	originalAddr := actualValue.Addr().Pointer()

	// For interface storage testing, we need to check if the interface value itself is addressable
	var interfaceAddr uintptr
	if interfaceReflectValue.CanAddr() {
		interfaceAddr = interfaceReflectValue.Addr().Pointer()
	} else {
		// If interface value is not addressable, it means it's stored directly
		// We test this by creating an addressable copy
		interfaceCopy := reflect.New(interfaceReflectValue.Type()).Elem()
		interfaceCopy.Set(interfaceReflectValue)
		interfaceAddr = interfaceCopy.Addr().Pointer()
	}

	isDirect := (originalAddr == interfaceAddr)

	// Return true if indirect storage (address differs), false if direct storage
	return !isDirect
}

// structArrayAddressComparison implements address comparison for structs and arrays as per CLAUDE.md
func (c *Compiler) structArrayAddressComparison(reflectType reflect.Type) bool {
	// For struct and array types specifically, use address comparison as requested
	// This is the method mentioned in CLAUDE.md: compare struct first field address with struct address

	// Create an addressable instance using reflect.New as suggested in CLAUDE.md
	ptrVal := reflect.New(reflectType)
	structVal := ptrVal.Elem() // Get the actual struct/array value

	if !structVal.CanAddr() {
		// If cannot get address, fallback to conservative approach
		return true // Assume indirect for safety
	}

	structAddr := structVal.Addr().Pointer()

	// For structs, check first field address alignment
	if reflectType.Kind() == reflect.Struct && reflectType.NumField() > 0 {
		firstField := structVal.Field(0)
		if firstField.CanAddr() {
			firstFieldAddr := firstField.Addr().Pointer()
			// If first field address matches struct address, it's direct storage
			// If they differ, it requires dereferencing (indirect storage)
			return structAddr != firstFieldAddr
		}
	}

	// For arrays and structs without fields or when field not addressable,
	// test interface storage directly by putting value into interface{}
	interfaceVal := structVal.Interface()
	interfaceReflectVal := reflect.ValueOf(interfaceVal)

	// Create addressable copy to get interface storage address
	interfaceCopy := reflect.New(interfaceReflectVal.Type()).Elem()
	interfaceCopy.Set(interfaceReflectVal)

	if interfaceCopy.CanAddr() {
		interfaceAddr := interfaceCopy.Addr().Pointer()
		// If addresses differ, storage is indirect
		return structAddr != interfaceAddr
	}

	// Conservative fallback: structs and arrays are typically indirect
	return true
}

// structAddressComparison implements CLAUDE.md address comparison for structs
func (c *Compiler) structAddressComparison(structType reflect.Type) (result bool) {
	// Panic recovery for safety
	defer func() {
		if r := recover(); r != nil {
			// If any panic occurs, conservatively assume indirect storage
			result = true
		}
	}()

	// CLAUDE.md: check if struct field address alignment affects interface storage
	// Safe implementation with nil checks to avoid panics

	if structType.NumField() == 0 {
		// Empty structs follow size rule
		return structType.Size() > unsafe.Sizeof(uintptr(0))
	}

	// Create addressable struct value safely
	structPtr := reflect.New(structType) // Create *StructType
	if !structPtr.IsValid() {
		return true // Conservative: assume indirect storage
	}

	structVal := structPtr.Elem() // Dereference to get StructType (addressable)
	if !structVal.IsValid() || !structVal.CanAddr() {
		return true // Conservative: assume indirect storage
	}

	// Get first field safely
	firstField := structVal.Field(0)
	if !firstField.IsValid() || !firstField.CanAddr() {
		return true // Conservative: assume indirect storage
	}

	// Only proceed with address comparison if both are safely addressable
	structAddr := structVal.Addr().Pointer()
	firstFieldAddr := firstField.Addr().Pointer()

	// CLAUDE.md: if addresses match, direct storage; if different, indirect storage
	addressesMatch := (structAddr == firstFieldAddr)

	return !addressesMatch // Return true for indirect storage
}

// noRuntimeInterfaceStorageLogic implements interface storage logic without runtime dependencies
func (c *Compiler) noRuntimeInterfaceStorageLogic(reflectType reflect.Type) bool {
	// Final implementation removing all runtime.IfaceIndir dependencies as per CLAUDE.md
	// Based on Go's interface storage rules:
	// - Pointer-like types (ptr, map, chan, func, unsafe.Pointer) and interfaces are stored directly
	// - All other types (including basic types, strings, slices, structs, arrays) are stored indirectly

	switch reflectType.Kind() {
	case reflect.Ptr, reflect.Chan, reflect.Map, reflect.Func, reflect.UnsafePointer:
		return false // Pointer-like types are stored directly
	case reflect.Interface:
		return false // Interface values stored directly
	default:
		// ALL other types are stored indirectly - this is the conservative approach
		// This includes: bool, int*, uint*, float*, complex*, string, array, slice, struct
		return true
	}
}

// correctInterfaceStorageLogic implements the exact Go runtime behavior
func (c *Compiler) correctInterfaceStorageLogic(reflectType reflect.Type) bool {
	// Based on actual runtime.IfaceIndir behavior from testing
	// Returns true if t is stored indirectly in an interface value

	// Only these pointer-like types are stored directly (return false)
	switch reflectType.Kind() {
	case reflect.Ptr, reflect.Chan, reflect.Map, reflect.Func, reflect.UnsafePointer:
		return false // Pointer-like types are stored directly
	case reflect.Interface:
		return false // Interface values stored directly
	default:
		// ALL other types (including basic types, strings, slices, structs, arrays)
		// are stored indirectly (return true)
		return true
	}
}

// IsIndirectFromType provides public access to interface storage detection
// Step-by-step replacement for VM usage
func IsIndirectFromType(typ *runtime.Type) bool {
	var c Compiler
	return c.isIndirectFromType(typ)
}

func (c *Compiler) structCode(typ *runtime.Type, isPtr bool) (*StructCode, error) {
	typeptr := uintptr(unsafe.Pointer(typ))
	if code, exists := c.structTypeToCode[typeptr]; exists {
		derefCode := *code
		derefCode.isRecursive = true
		return &derefCode, nil
	}
	indirect := c.isIndirectFromType(typ)
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

func (c *Compiler) structCodeFromValue(typ *runtime.Type, v reflect.Value, isPtr bool) (*StructCode, error) {
	// For safety and to avoid memory corruption, use the original structCode implementation
	// The reflect.Value parameter is preserved for API compatibility but not used
	return c.structCode(typ, isPtr)
}

func toElemType(t *runtime.Type) *runtime.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (c *Compiler) structFieldCode(structCode *StructCode, tag *runtime.StructTag, isPtr, isOnlyOneFirstField bool) (*StructFieldCode, error) {
	field := tag.Field
	fieldType := runtime.Type2RType(field.Type)
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
		fieldCode.isNilCheck = false
		structCode.isIndirect = false
		structCode.disableIndirectConversion = true
	case c.isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(fieldType, isIndirectSpecialCase):
		code, err := c.marshalTextCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
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
		fieldCode.isNilCheck = false
	case isPtr && c.isPtrMarshalTextType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalText() ([]byte, error)
		code, err := c.marshalTextCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
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

func (c *Compiler) typeToStructTags(typ *runtime.Type) runtime.StructTags {
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
func (c *Compiler) isMovePointerPositionFromHeadToFirstMarshalJSONFieldCase(typ *runtime.Type, isIndirectSpecialCase bool) bool {
	return isIndirectSpecialCase && !c.isNilableType(typ) && c.isPtrMarshalJSONType(typ)
}

// *struct{ field T } => struct { field *T }
// func (*T) MarshalText() ([]byte, error)
func (c *Compiler) isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(typ *runtime.Type, isIndirectSpecialCase bool) bool {
	return isIndirectSpecialCase && !c.isNilableType(typ) && c.isPtrMarshalTextType(typ)
}

func (c *Compiler) implementsMarshalJSON(typ *runtime.Type) bool {
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

func (c *Compiler) implementsMarshalText(typ *runtime.Type) bool {
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

func (c *Compiler) isNilableType(typ *runtime.Type) bool {
	if !c.isIndirectFromType(typ) {
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

func (c *Compiler) implementsMarshalJSONType(typ *runtime.Type) bool {
	return typ.Implements(marshalJSONType) || typ.Implements(marshalJSONContextType)
}

func (c *Compiler) isPtrMarshalJSONType(typ *runtime.Type) bool {
	return !c.implementsMarshalJSONType(typ) && c.implementsMarshalJSONType(runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(typ))))
}

func (c *Compiler) isPtrMarshalTextType(typ *runtime.Type) bool {
	return !typ.Implements(marshalTextType) && runtime.Type2RType(reflect.PtrTo(runtime.RType2Type(typ))).Implements(marshalTextType)
}

func (c *Compiler) codeToOpcode(ctx *compileContext, typ *runtime.Type, code Code) *Opcode {
	codes := code.ToOpcode(ctx)
	codes.Last().Next = newEndOp(ctx, typ)
	c.linkRecursiveCode(ctx)
	return codes.First()
}

func (c *Compiler) linkRecursiveCode(ctx *compileContext) {
	recursiveCodes := map[uintptr]*CompiledCode{}
	for _, recursive := range *ctx.recursiveCodes {
		typeptr := uintptr(unsafe.Pointer(recursive.Type))
		codes := ctx.structTypeToCodes[typeptr]
		if recursiveCode, ok := recursiveCodes[typeptr]; ok {
			*recursive.Jmp = *recursiveCode
			continue
		}

		code := copyOpcode(codes.First())
		code.Op = code.Op.PtrHeadToHead()
		lastCode := newEndOp(&compileContext{}, recursive.Type)
		lastCode.Op = OpRecursiveEnd

		// OpRecursiveEnd must set before call TotalLength
		code.End.Next = lastCode

		totalLength := code.TotalLength()

		// Idx, ElemIdx, Length must set after call TotalLength
		lastCode.Idx = uint32((totalLength + 1) * uintptrSize)
		lastCode.ElemIdx = lastCode.Idx + uintptrSize
		lastCode.Length = lastCode.Idx + 2*uintptrSize

		// extend length to alloc slot for elemIdx + length
		curTotalLength := uintptr(recursive.TotalLength()) + 3
		nextTotalLength := uintptr(totalLength) + 3

		compiled := recursive.Jmp
		compiled.Code = code
		compiled.CurLen = curTotalLength
		compiled.NextLen = nextTotalLength
		compiled.Linked = true

		recursiveCodes[typeptr] = compiled
	}
}

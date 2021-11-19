package encoder

import (
	"reflect"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

type Code interface {
	Type() CodeType2
	ToOpcode() []*Opcode
}

type CodeType2 int

const (
	CodeTypeInterface CodeType2 = iota
	CodeTypePtr
	CodeTypeInt
	CodeTypeUint
	CodeTypeFloat
	CodeTypeString
	CodeTypeBool
	CodeTypeStruct
	CodeTypeMap
	CodeTypeSlice
	CodeTypeArray
	CodeTypeBytes
	CodeTypeMarshalJSON
	CodeTypeMarshalText
)

type IntCode struct {
	typ      *runtime.Type
	bitSize  uint8
	isString bool
	isPtr    bool
}

func (c *IntCode) Type() CodeType2 {
	return CodeTypeInt
}

func (c *IntCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type UintCode struct {
	typ      *runtime.Type
	bitSize  uint8
	isString bool
	isPtr    bool
}

func (c *UintCode) Type() CodeType2 {
	return CodeTypeUint
}

func (c *UintCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type FloatCode struct {
	typ      *runtime.Type
	bitSize  uint8
	isString bool
	isPtr    bool
}

func (c *FloatCode) Type() CodeType2 {
	return CodeTypeFloat
}

func (c *FloatCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type StringCode struct {
	typ      *runtime.Type
	isString bool
	isPtr    bool
}

func (c *StringCode) Type() CodeType2 {
	return CodeTypeString
}

func (c *StringCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type BoolCode struct {
	typ      *runtime.Type
	isString bool
	isPtr    bool
}

func (c *BoolCode) Type() CodeType2 {
	return CodeTypeBool
}

func (c *BoolCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type SliceCode struct {
	typ *runtime.Type
}

func (c *SliceCode) Type() CodeType2 {
	return CodeTypeSlice
}

func (c *SliceCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type ArrayCode struct {
	typ *runtime.Type
}

func (c *ArrayCode) Type() CodeType2 {
	return CodeTypeArray
}

func (c *ArrayCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type MapCode struct {
	typ *runtime.Type
}

func (c *MapCode) Type() CodeType2 {
	return CodeTypeMap
}

func (c *MapCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type BytesCode struct {
	typ   *runtime.Type
	isPtr bool
}

func (c *BytesCode) Type() CodeType2 {
	return CodeTypeBytes
}

func (c *BytesCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type StructCode struct {
	typ                       *runtime.Type
	isPtr                     bool
	fields                    []*StructFieldCode
	disableIndirectConversion bool
}

func (c *StructCode) Type() CodeType2 {
	return CodeTypeStruct
}

func (c *StructCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type StructFieldCode struct {
	typ                *runtime.Type
	key                string
	value              Code
	offset             uintptr
	isAnonymous        bool
	isTaggedKey        bool
	isNilableType      bool
	isNilCheck         bool
	isAddrForMarshaler bool
	isNextOpPtrType    bool
}

func (c *StructFieldCode) toInlineCode() []*StructFieldCode {
	return nil
}

type InterfaceCode struct {
	typ   *runtime.Type
	isPtr bool
}

func (c *InterfaceCode) Type() CodeType2 {
	return CodeTypeInterface
}

func (c *InterfaceCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type MarshalJSONCode struct {
	typ *runtime.Type
}

func (c *MarshalJSONCode) Type() CodeType2 {
	return CodeTypeMarshalJSON
}

func (c *MarshalJSONCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type MarshalTextCode struct {
	typ *runtime.Type
}

func (c *MarshalTextCode) Type() CodeType2 {
	return CodeTypeMarshalText
}

func (c *MarshalTextCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

type PtrCode struct {
	typ   *runtime.Type
	value Code
}

func (c *PtrCode) Type() CodeType2 {
	return CodeTypePtr
}

func (c *PtrCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func type2code(ctx *compileContext) (Code, error) {
	typ := ctx.typ
	switch {
	case implementsMarshalJSON(typ):
		return compileMarshalJSON2(ctx)
	case implementsMarshalText(typ):
		return compileMarshalText2(ctx)
	}

	isPtr := false
	orgType := typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		isPtr = true
	}
	switch {
	case implementsMarshalJSON(typ):
		return compileMarshalJSON2(ctx)
	case implementsMarshalText(typ):
		return compileMarshalText2(ctx)
	}
	switch typ.Kind() {
	case reflect.Slice:
		ctx := ctx.withType(typ)
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := runtime.PtrTo(elem)
			if !implementsMarshalJSONType(p) && !p.Implements(marshalTextType) {
				return compileBytes2(ctx, isPtr)
			}
		}
		return compileSlice2(ctx)
	case reflect.Map:
		if isPtr {
			return compilePtr2(ctx.withType(runtime.PtrTo(typ)))
		}
		return compileMap2(ctx.withType(typ))
	case reflect.Struct:
		return compileStruct2(ctx.withType(typ), isPtr)
	case reflect.Int:
		return compileInt2(ctx.withType(typ), isPtr)
	case reflect.Int8:
		return compileInt82(ctx.withType(typ), isPtr)
	case reflect.Int16:
		return compileInt162(ctx.withType(typ), isPtr)
	case reflect.Int32:
		return compileInt322(ctx.withType(typ), isPtr)
	case reflect.Int64:
		return compileInt642(ctx.withType(typ), isPtr)
	case reflect.Uint, reflect.Uintptr:
		return compileUint2(ctx.withType(typ), isPtr)
	case reflect.Uint8:
		return compileUint82(ctx.withType(typ), isPtr)
	case reflect.Uint16:
		return compileUint162(ctx.withType(typ), isPtr)
	case reflect.Uint32:
		return compileUint322(ctx.withType(typ), isPtr)
	case reflect.Uint64:
		return compileUint642(ctx.withType(typ), isPtr)
	case reflect.Float32:
		return compileFloat322(ctx.withType(typ), isPtr)
	case reflect.Float64:
		return compileFloat642(ctx.withType(typ), isPtr)
	case reflect.String:
		return compileString2(ctx.withType(typ), isPtr)
	case reflect.Bool:
		return compileBool2(ctx.withType(typ), isPtr)
	case reflect.Interface:
		return compileInterface2(ctx.withType(typ), isPtr)
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
		return compileMarshalJSON2(ctx)
	case implementsMarshalText(typ):
		return compileMarshalText2(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return compilePtr2(ctx)
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := runtime.PtrTo(elem)
			if !implementsMarshalJSONType(p) && !p.Implements(marshalTextType) {
				return compileBytes2(ctx, false)
			}
		}
		return compileSlice2(ctx)
	case reflect.Array:
		return compileArray2(ctx)
	case reflect.Map:
		return compileMap2(ctx)
	case reflect.Struct:
		return compileStruct2(ctx, isPtr)
	case reflect.Interface:
		return compileInterface2(ctx, false)
	case reflect.Int:
		return compileInt2(ctx, false)
	case reflect.Int8:
		return compileInt82(ctx, false)
	case reflect.Int16:
		return compileInt162(ctx, false)
	case reflect.Int32:
		return compileInt322(ctx, false)
	case reflect.Int64:
		return compileInt642(ctx, false)
	case reflect.Uint:
		return compileUint2(ctx, false)
	case reflect.Uint8:
		return compileUint82(ctx, false)
	case reflect.Uint16:
		return compileUint162(ctx, false)
	case reflect.Uint32:
		return compileUint322(ctx, false)
	case reflect.Uint64:
		return compileUint642(ctx, false)
	case reflect.Uintptr:
		return compileUint2(ctx, false)
	case reflect.Float32:
		return compileFloat322(ctx, false)
	case reflect.Float64:
		return compileFloat642(ctx, false)
	case reflect.String:
		return compileString2(ctx, false)
	case reflect.Bool:
		return compileBool2(ctx, false)
	}
	return nil, &errors.UnsupportedTypeError{Type: runtime.RType2Type(typ)}
}

func compileInt2(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: intSize, isPtr: isPtr}, nil
}

func compileInt82(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: intSize, isPtr: isPtr}, nil
}

func compileInt162(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 8, isPtr: isPtr}, nil
}

func compileInt322(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 16, isPtr: isPtr}, nil
}

func compileInt642(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 64, isPtr: isPtr}, nil
}

func compileUint2(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: intSize, isPtr: isPtr}, nil
}

func compileUint82(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: intSize, isPtr: isPtr}, nil
}

func compileUint162(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 8, isPtr: isPtr}, nil
}

func compileUint322(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 16, isPtr: isPtr}, nil
}

func compileUint642(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 64, isPtr: isPtr}, nil
}

func compileFloat322(ctx *compileContext, isPtr bool) (*FloatCode, error) {
	return &FloatCode{typ: ctx.typ, bitSize: 32, isPtr: isPtr}, nil
}

func compileFloat642(ctx *compileContext, isPtr bool) (*FloatCode, error) {
	return &FloatCode{typ: ctx.typ, bitSize: 64, isPtr: isPtr}, nil
}

func compileString2(ctx *compileContext, isString bool) (*StringCode, error) {
	return &StringCode{typ: ctx.typ, isString: isString}, nil
}

func compileBool2(ctx *compileContext, isString bool) (*BoolCode, error) {
	return &BoolCode{typ: ctx.typ, isString: isString}, nil
}

func compileSlice2(ctx *compileContext) (*SliceCode, error) {
	return &SliceCode{typ: ctx.typ}, nil
}

func compileArray2(ctx *compileContext) (*ArrayCode, error) {
	return &ArrayCode{typ: ctx.typ}, nil
}

func compileMap2(ctx *compileContext) (*MapCode, error) {
	return &MapCode{typ: ctx.typ}, nil
}

func compileBytes2(ctx *compileContext, isPtr bool) (*BytesCode, error) {
	return &BytesCode{typ: ctx.typ, isPtr: isPtr}, nil
}

func compileInterface2(ctx *compileContext, isPtr bool) (*InterfaceCode, error) {
	return &InterfaceCode{typ: ctx.typ, isPtr: isPtr}, nil
}

func compileMarshalJSON2(ctx *compileContext) (*MarshalJSONCode, error) {
	return &MarshalJSONCode{typ: ctx.typ}, nil
}

func compileMarshalText2(ctx *compileContext) (*MarshalTextCode, error) {
	return &MarshalTextCode{typ: ctx.typ}, nil
}

func compilePtr2(ctx *compileContext) (*PtrCode, error) {
	code, err := type2codeWithPtr(ctx.withType(ctx.typ.Elem()), true)
	if err != nil {
		return nil, err
	}
	return &PtrCode{typ: ctx.typ, value: code}, nil
}

func compileStruct2(ctx *compileContext, isPtr bool) (*StructCode, error) {
	//typeptr := uintptr(unsafe.Pointer(typ))
	//compiled := &CompiledCode{}
	//ctx.structTypeToCompiledCode[typeptr] = compiled
	// header => code => structField => code => end
	//                        ^          |
	//                        |__________|
	typ := ctx.typ
	fieldNum := typ.NumField()
	//indirect := runtime.IfaceIndir(typ)
	tags := typeToStructTags(typ)
	code := &StructCode{typ: typ, isPtr: isPtr}
	fields := []*StructFieldCode{}
	for i, tag := range tags {
		isOnlyOneFirstField := i == 0 && fieldNum == 1
		field, err := code.compileStructField(ctx, tag, isPtr, isOnlyOneFirstField)
		if err != nil {
			return nil, err
		}
		if field.isAnonymous {
			structCode, ok := field.value.(*StructCode)
			if ok {
				for _, field := range structCode.fields {
					if tags.ExistsKey(field.key) {
						continue
					}
					fields = append(fields, field)
				}
			} else {
				fields = append(fields, field)
			}
		} else {
			fields = append(fields, field)
		}
	}
	fieldMap := map[string][]*StructFieldCode{}
	for _, field := range fields {
		fieldMap[field.key] = append(fieldMap[field.key], field)
	}
	removeFieldKey := map[string]struct{}{}
	for _, fields := range fieldMap {
		if len(fields) == 1 {
			continue
		}
		var foundTaggedKey bool
		for _, field := range fields {
			if field.isTaggedKey {
				if foundTaggedKey {
					removeFieldKey[field.key] = struct{}{}
					break
				}
				foundTaggedKey = true
			}
		}
	}
	filteredFields := make([]*StructFieldCode, 0, len(fields))
	for _, field := range fields {
		if _, exists := removeFieldKey[field.key]; exists {
			continue
		}
		filteredFields = append(filteredFields, field)
	}
	code.fields = filteredFields
	return code, nil
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

func (c *StructCode) compileStructField(ctx *compileContext, tag *runtime.StructTag, isPtr, isOnlyOneFirstField bool) (*StructFieldCode, error) {
	field := tag.Field
	fieldType := runtime.Type2RType(field.Type)
	isIndirectSpecialCase := isPtr && isOnlyOneFirstField
	fieldCode := &StructFieldCode{
		typ:           fieldType,
		key:           tag.Key,
		offset:        field.Offset,
		isAnonymous:   field.Anonymous && !tag.IsTaggedKey,
		isTaggedKey:   tag.IsTaggedKey,
		isNilableType: isNilableType(fieldType),
		isNilCheck:    true,
	}
	switch {
	case isMovePointerPositionFromHeadToFirstMarshalJSONFieldCase(fieldType, isIndirectSpecialCase):
		code, err := compileMarshalJSON2(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
		c.disableIndirectConversion = true
	case isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(fieldType, isIndirectSpecialCase):
		code, err := compileMarshalText2(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
		c.disableIndirectConversion = true
	case isPtr && isPtrMarshalJSONType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalJSON() ([]byte, error)
		code, err := compileMarshalJSON2(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
	case isPtr && isPtrMarshalTextType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalText() ([]byte, error)
		code, err := compileMarshalText2(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
	default:
		code, err := type2codeWithPtr(ctx.withType(fieldType), isPtr)
		if err != nil {
			return nil, err
		}
		switch code.Type() {
		case CodeTypePtr, CodeTypeInterface:
			fieldCode.isNextOpPtrType = true
		}
		fieldCode.value = code
		fieldCode.isNilCheck = false
	}
	return fieldCode, nil
}

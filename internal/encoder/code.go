package encoder

import (
	"reflect"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

type Code interface {
	Type() CodeType2
	ToOpcode() []*Opcode
	Optimize() error
}

type CodeType2 int

const (
	CodeTypeInterface CodeType2 = iota
	CodeTypePtr
)

type IntCode struct {
	typ      *runtime.Type
	bitSize  uint8
	isString bool
}

func (c *IntCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *IntCode) Optimize() error { return nil }

func newIntCode(typ *runtime.Type, bitSize uint8, isString bool) *IntCode {
	return &IntCode{typ: typ, bitSize: bitSize, isString: isString}
}

type UintCode struct {
	typ      *runtime.Type
	bitSize  uint8
	isString bool
}

func (c *UintCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *UintCode) Optimize() error { return nil }

func newUintCode(typ *runtime.Type, bitSize uint8, isString bool) *UintCode {
	return &UintCode{typ: typ, bitSize: bitSize, isString: isString}
}

type FloatCode struct {
	typ      *runtime.Type
	bitSize  uint8
	isString bool
}

func (c *FloatCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *FloatCode) Optimize() error { return nil }

func newFloatCode(typ *runtime.Type, bitSize uint8, isString bool) *FloatCode {
	return &FloatCode{typ: typ, bitSize: bitSize, isString: isString}
}

type StringCode struct {
	typ      *runtime.Type
	isString bool
}

func (c *StringCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *StringCode) Optimize() error { return nil }

func newStringCode(typ *runtime.Type, isString bool) *StringCode {
	return &StringCode{typ: typ, isString: isString}
}

type BoolCode struct {
	typ      *runtime.Type
	isString bool
}

func (c *BoolCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *BoolCode) Optimize() error { return nil }

func newBoolCode(typ *runtime.Type, isString bool) *BoolCode {
	return &BoolCode{typ: typ, isString: isString}
}

type SliceCode struct {
	typ *runtime.Type
}

func (c *SliceCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *SliceCode) Optimize() error { return nil }

func newSliceCode(typ *runtime.Type) *SliceCode {
	return &SliceCode{typ: typ}
}

type ArrayCode struct {
	typ *runtime.Type
}

func (c *ArrayCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *ArrayCode) Optimize() error { return nil }

func newArrayCode(typ *runtime.Type) *ArrayCode {
	return &ArrayCode{typ: typ}
}

type MapCode struct {
	typ *runtime.Type
}

func (c *MapCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *MapCode) Optimize() error { return nil }

func newMapCode(typ *runtime.Type) *MapCode {
	return &MapCode{typ: typ}
}

type BytesCode struct {
	typ *runtime.Type
}

func (c *BytesCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *BytesCode) Optimize() error { return nil }

func newBytesCode(typ *runtime.Type) *BytesCode {
	return &BytesCode{typ: typ}
}

type StructCode struct {
	typ                       *runtime.Type
	isPtr                     bool
	fields                    []*StructFieldCode
	disableIndirectConversion bool
}

func (c *StructCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *StructCode) Optimize() error { return nil }

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

func newStructCode(typ *runtime.Type, isPtr bool) (*StructCode, error) {
	//typeptr := uintptr(unsafe.Pointer(typ))
	//compiled := &CompiledCode{}
	//ctx.structTypeToCompiledCode[typeptr] = compiled
	// header => code => structField => code => end
	//                        ^          |
	//                        |__________|
	fieldNum := typ.NumField()
	indirect := runtime.IfaceIndir(typ)
	tags := typeToStructTags(typ)
	fields := []*StructFieldCode{}
	for i, tag := range tags {
		isOnlyOneFirstField := i == 0 && fieldNum == 1
		field, err := newStructFieldCode(tag, isPtr, isOnlyOneFirstField)
		if err != nil {
			return nil, err
		}
		if field.isAnonymous {
			fields = append(fields, field.toInlineCode()...)
		} else {
			fields = append(fields, field)
		}
	}
	return &StructCode{typ: typ, isPtr: isPtr, fields: fields}
}

func newStructFieldCode(tag *runtime.StructTag, isPtr, isOnlyOneFirstField bool) (Code, error) {
	field := tag.Field
	fieldType := runtime.Type2RType(field.Type)
	isIndirectSpecialCase := isPtr && isOnlyOneFirstField
	fieldCode := &StructFieldCode{
		typ:           fieldType,
		key:           tag.Key,
		offset:        field.Offset,
		isAnonymous:   field.Anonymous,
		isTaggedKey:   tag.IsTaggedKey,
		isNilableType: isNilableType(fieldType),
		isNilCheck:    true,
	}
	switch {
	case isMovePointerPositionFromHeadToFirstMarshalJSONFieldCase(isIndirectSpecialCase, fieldType):
		code, err := newMarshalJSONCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
	case isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(isIndirectSpecialCase, fieldType):
		code, err := newMarshalTextCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
	case isPtr && isPtrMarshalJSONType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalJSON() ([]byte, error)
		code, err := newMarshalJSONCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
	case isPtr && isPtrMarshalTextType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalText() ([]byte, error)
		code, err := newMarshalTextCode(fieldType)
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
	default:
		code, err := type2codeWithPtr(fieldType, isPtr)
		if err != nil {
			return nil, err
		}
		switch code.Type() {
		case PtrCodeType, InterfaceCodeType:
			fieldCode.isNextOpPtrType = true
		}
		fieldCode.value = code
		fieldCode.isNilCheck = false
	}
	return fieldCode, nil
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

type InterfaceCode struct {
	typ *runtime.Type
}

func (c *InterfaceCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *InterfaceCode) Optimize() error { return nil }

func newIfaceCode(typ *runtime.Type) *InterfaceCode {
	return &InterfaceCode{typ: typ}
}

type PtrCode struct {
	typ   *runtime.Type
	value Code
}

func (c *PtrCode) ToOpcode() []*Opcode {
	return []*Opcode{}
}

func (c *PtrCode) Optimize() error { return nil }

func newPtrCode(typ *runtime.Type, value Code) *PtrCode {
	return &PtrCode{typ: typ, value: value}
}

func type2code(typ *runtime.Type) (Code, error) {
	switch {
	case implementsMarshalJSON(typ):
		//return compileMarshalJSON(ctx)
	case implementsMarshalText(typ):
		//return compileMarshalText(ctx)
	}

	isPtr := false
	orgType := typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		isPtr = true
	}
	switch {
	case implementsMarshalJSON(typ):
		//return compileMarshalJSON(ctx)
	case implementsMarshalText(typ):
		//return compileMarshalText(ctx)
	}
	switch typ.Kind() {
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := runtime.PtrTo(elem)
			if !implementsMarshalJSONType(p) && !p.Implements(marshalTextType) {
				code := newBytesCode(typ)
				if isPtr {
					return newPtrCode(orgType, code), nil
				}
				return code, nil
			}
		}
		return newSliceCode(typ), nil
	case reflect.Map:
		code := newMapCode(typ)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Struct:
		return newStructCode(typ, isPtr), nil
	case reflect.Int:
		code := newIntCode(typ, intSize, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Int8:
		code := newIntCode(typ, 8, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Int16:
		code := newIntCode(typ, 16, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Int32:
		code := newIntCode(typ, 32, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Int64:
		code := newIntCode(typ, 64, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Uint, reflect.Uintptr:
		code := newUintCode(typ, intSize, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Uint8:
		code := newUintCode(typ, 8, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Uint16:
		code := newUintCode(typ, 16, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Uint32:
		code := newUintCode(typ, 32, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Uint64:
		code := newUintCode(typ, 64, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Float32:
		code := newFloatCode(typ, 32, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Float64:
		code := newFloatCode(typ, 64, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.String:
		code := newStringCode(typ, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Bool:
		code := newBoolCode(typ, false)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	case reflect.Interface:
		code := newIfaceCode(typ)
		if isPtr {
			return newPtrCode(orgType, code), nil
		}
		return code, nil
	default:
		if isPtr && typ.Implements(marshalTextType) {
			typ = orgType
		}
		code, err := type2codeWithPtr(typ, isPtr)
		if err != nil {
			return nil, err
		}
		return code, nil
	}
}

func type2codeWithPtr(typ *runtime.Type, isPtr bool) (Code, error) {
	switch {
	case implementsMarshalJSON(typ):
		//return compileMarshalJSON(ctx)
	case implementsMarshalText(typ):
		//return compileMarshalText(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		code, err := type2codeWithPtr(typ.Elem(), false)
		if err != nil {
			return nil, err
		}
		return newPtrCode(typ, code), nil
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			p := runtime.PtrTo(elem)
			if !implementsMarshalJSONType(p) && !p.Implements(marshalTextType) {
				return newBytesCode(typ), nil
			}
		}
		return newSliceCode(typ), nil
	case reflect.Array:
		return newArrayCode(typ), nil
	case reflect.Map:
		return newMapCode(typ), nil
	case reflect.Struct:
		return newStructCode(typ, isPtr), nil
	case reflect.Interface:
		return newIfaceCode(typ), nil
	case reflect.Int:
		return newIntCode(typ, intSize, false), nil
	case reflect.Int8:
		return newIntCode(typ, 8, false), nil
	case reflect.Int16:
		return newIntCode(typ, 16, false), nil
	case reflect.Int32:
		return newIntCode(typ, 32, false), nil
	case reflect.Int64:
		return newIntCode(typ, 64, false), nil
	case reflect.Uint:
		return newUintCode(typ, intSize, false), nil
	case reflect.Uint8:
		return newUintCode(typ, 8, false), nil
	case reflect.Uint16:
		return newUintCode(typ, 16, false), nil
	case reflect.Uint32:
		return newUintCode(typ, 32, false), nil
	case reflect.Uint64:
		return newUintCode(typ, 64, false), nil
	case reflect.Uintptr:
		return newUintCode(typ, intSize, false), nil
	case reflect.Float32:
		return newFloatCode(typ, 32, false), nil
	case reflect.Float64:
		return newFloatCode(typ, 64, false), nil
	case reflect.String:
		return newStringCode(typ, false), nil
	case reflect.Bool:
		return newBoolCode(typ, false), nil
	}
	return nil, &errors.UnsupportedTypeError{Type: runtime.RType2Type(typ)}
}

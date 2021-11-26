package encoder

import (
	"fmt"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

type Code interface {
	Type() CodeType2
	ToOpcode(*compileContext) Opcodes
}

type AnonymousCode interface {
	ToAnonymousOpcode(*compileContext) Opcodes
}

type Opcodes []*Opcode

func (o Opcodes) First() *Opcode {
	if len(o) == 0 {
		return nil
	}
	return o[0]
}

func (o Opcodes) Last() *Opcode {
	if len(o) == 0 {
		return nil
	}
	return o[len(o)-1]
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
	CodeTypeRecursive
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

func (c *IntCode) ToOpcode(ctx *compileContext) Opcodes {
	var code *Opcode
	switch {
	case c.isPtr:
		code = newOpCode(ctx, OpIntPtr)
	case c.isString:
		code = newOpCode(ctx, OpIntString)
	default:
		code = newOpCode(ctx, OpInt)
	}
	code.NumBitSize = c.bitSize
	ctx.incIndex()
	return Opcodes{code}
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

func (c *UintCode) ToOpcode(ctx *compileContext) Opcodes {
	var code *Opcode
	switch {
	case c.isPtr:
		code = newOpCode(ctx, OpUintPtr)
	case c.isString:
		code = newOpCode(ctx, OpUintString)
	default:
		code = newOpCode(ctx, OpUint)
	}
	code.NumBitSize = c.bitSize
	ctx.incIndex()
	return Opcodes{code}
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

func (c *FloatCode) ToOpcode(ctx *compileContext) Opcodes {
	var code *Opcode
	switch {
	case c.isPtr:
		switch c.bitSize {
		case 32:
			code = newOpCode(ctx, OpFloat32Ptr)
		default:
			code = newOpCode(ctx, OpFloat64Ptr)
		}
	default:
		switch c.bitSize {
		case 32:
			code = newOpCode(ctx, OpFloat32)
		default:
			code = newOpCode(ctx, OpFloat64)
		}
	}
	ctx.incIndex()
	return Opcodes{code}
}

type StringCode struct {
	typ      *runtime.Type
	isString bool
	isPtr    bool
}

func (c *StringCode) Type() CodeType2 {
	return CodeTypeString
}

func (c *StringCode) ToOpcode(ctx *compileContext) Opcodes {
	isJsonNumberType := c.typ == runtime.Type2RType(jsonNumberType)
	var code *Opcode
	if c.isPtr {
		if isJsonNumberType {
			code = newOpCode(ctx, OpNumberPtr)
		} else {
			code = newOpCode(ctx, OpStringPtr)
		}
	} else {
		if isJsonNumberType {
			code = newOpCode(ctx, OpNumber)
		} else {
			code = newOpCode(ctx, OpString)
		}
	}
	ctx.incIndex()
	return Opcodes{code}
}

type BoolCode struct {
	typ      *runtime.Type
	isString bool
	isPtr    bool
}

func (c *BoolCode) Type() CodeType2 {
	return CodeTypeBool
}

func (c *BoolCode) ToOpcode(ctx *compileContext) Opcodes {
	var code *Opcode
	switch {
	case c.isPtr:
		code = newOpCode(ctx, OpBoolPtr)
	default:
		code = newOpCode(ctx, OpBool)
	}
	ctx.incIndex()
	return Opcodes{code}
}

type BytesCode struct {
	typ   *runtime.Type
	isPtr bool
}

func (c *BytesCode) Type() CodeType2 {
	return CodeTypeBytes
}

func (c *BytesCode) ToOpcode(ctx *compileContext) Opcodes {
	var code *Opcode
	switch {
	case c.isPtr:
		code = newOpCode(ctx, OpBytesPtr)
	default:
		code = newOpCode(ctx, OpBytes)
	}
	ctx.incIndex()
	return Opcodes{code}
}

type SliceCode struct {
	typ   *runtime.Type
	value Code
}

func (c *SliceCode) Type() CodeType2 {
	return CodeTypeSlice
}

func (c *SliceCode) ToOpcode(ctx *compileContext) Opcodes {
	// header => opcode => elem => end
	//             ^        |
	//             |________|
	size := c.typ.Elem().Size()
	header := newSliceHeaderCode(ctx)
	ctx.incIndex()
	codes := c.value.ToOpcode(ctx.incIndent())
	codes.First().Flags |= IndirectFlags
	elemCode := newSliceElemCode(ctx.withType(c.typ.Elem()), header, size)
	ctx.incIndex()
	end := newOpCode(ctx, OpSliceEnd)
	ctx.incIndex()
	header.End = end
	header.Next = codes.First()
	codes.Last().Next = elemCode
	elemCode.Next = codes.First()
	elemCode.End = end
	return append(append(Opcodes{header}, codes...), elemCode, end)
}

type ArrayCode struct {
	typ   *runtime.Type
	value Code
}

func (c *ArrayCode) Type() CodeType2 {
	return CodeTypeArray
}

func (c *ArrayCode) ToOpcode(ctx *compileContext) Opcodes {
	// header => opcode => elem => end
	//             ^        |
	//             |________|
	elem := c.typ.Elem()
	alen := c.typ.Len()
	size := elem.Size()

	header := newArrayHeaderCode(ctx, alen)
	ctx.incIndex()

	codes := c.value.ToOpcode(ctx.incIndent())
	codes.First().Flags |= IndirectFlags

	elemCode := newArrayElemCode(ctx.withType(elem), header, alen, size)
	ctx.incIndex()

	end := newOpCode(ctx, OpArrayEnd)
	ctx.incIndex()

	header.End = end
	header.Next = codes.First()
	codes.Last().Next = elemCode
	elemCode.Next = codes.First()
	elemCode.End = end

	return append(append(Opcodes{header}, codes...), elemCode, end)
}

type MapCode struct {
	typ   *runtime.Type
	key   Code
	value Code
}

func (c *MapCode) Type() CodeType2 {
	return CodeTypeMap
}

func (c *MapCode) ToOpcode(ctx *compileContext) Opcodes {
	// header => code => value => code => key => code => value => code => end
	//                                     ^                       |
	//                                     |_______________________|
	header := newMapHeaderCode(ctx)
	ctx.incIndex()

	keyCodes := c.key.ToOpcode(ctx)

	value := newMapValueCode(ctx, header)
	ctx.incIndex()
	valueCodes := c.value.ToOpcode(ctx.incIndent())
	valueCodes.First().Flags |= IndirectFlags

	key := newMapKeyCode(ctx, header)
	ctx.incIndex()

	end := newMapEndCode(ctx, header)
	ctx.incIndex()

	header.Next = keyCodes.First()
	keyCodes.Last().Next = value
	value.Next = valueCodes.First()
	valueCodes.Last().Next = key
	key.Next = keyCodes.First()

	header.End = end
	key.End = end
	value.End = end
	return append(append(append(append(append(Opcodes{header}, keyCodes...), value), valueCodes...), key), end)
}

type StructCode struct {
	typ                       *runtime.Type
	isPtr                     bool
	fields                    []*StructFieldCode
	disableIndirectConversion bool
	isIndirect                bool
	isRecursive               bool
	recursiveCodes            Opcodes
}

func (c *StructCode) Type() CodeType2 {
	return CodeTypeStruct
}

func (c *StructCode) ToOpcode(ctx *compileContext) Opcodes {
	// header => code => structField => code => end
	//                        ^          |
	//                        |__________|
	if c.isRecursive {
		recursive := newRecursiveCode(ctx, &CompiledCode{})
		recursive.Type = c.typ
		ctx.incIndex()
		*ctx.recursiveCodes = append(*ctx.recursiveCodes, recursive)
		return Opcodes{recursive}
	}
	codes := Opcodes{}
	var prevField *Opcode
	ctx = ctx.incIndent()
	for idx, field := range c.fields {
		isFirstField := idx == 0
		isEndField := idx == len(c.fields)-1
		fieldCodes := field.ToOpcode(ctx, isFirstField, isEndField)
		for _, code := range fieldCodes {
			if c.isIndirect {
				code.Flags |= IndirectFlags
			}
		}
		if len(codes) > 0 {
			codes.Last().Next = fieldCodes.First()
			fieldCodes.First().Idx = codes.First().Idx
		}
		if prevField != nil {
			prevField.NextField = fieldCodes.First()
		}
		if isEndField {
			if len(codes) > 0 {
				codes.First().End = fieldCodes.Last()
			} else if field.isAnonymous {
				fieldCodes.First().End = fieldCodes.Last()
				//fieldCodes.First().Next.End = fieldCodes.Last()
				fieldCode := fieldCodes.First().Next
				for fieldCode.NextField != nil {
					fieldCode = fieldCode.NextField
				}
				// link curLastField => endField
				fieldCode.NextField = fieldCodes.Last()
			} else {
				fieldCodes.First().End = fieldCodes.Last()
			}
			codes = append(codes, fieldCodes...)
			break
		}
		if field.isAnonymous {
			// fieldCodes.First() is StructHead operation.
			// StructHead's next operation is truely head operation.
			fieldCode := fieldCodes.First().Next
			for fieldCode.NextField != nil {
				fieldCode = fieldCode.NextField
			}
			prevField = fieldCode
		} else {
			fieldCode := fieldCodes.First()
			for fieldCode.NextField != nil {
				fieldCode = fieldCode.NextField
			}
			prevField = fieldCode
		}
		codes = append(codes, fieldCodes...)
	}
	if len(codes) == 0 {
		head := &Opcode{
			Op:         OpStructHead,
			Idx:        opcodeOffset(ctx.ptrIndex),
			Type:       c.typ,
			DisplayIdx: ctx.opcodeIndex,
			Indent:     ctx.indent,
		}
		ctx.incOpcodeIndex()
		end := &Opcode{
			Op:         OpStructEnd,
			Idx:        opcodeOffset(ctx.ptrIndex),
			DisplayIdx: ctx.opcodeIndex,
			Indent:     ctx.indent,
		}
		head.NextField = end
		head.Next = end
		head.End = end
		end.PrevField = head
		codes = append(codes, head, end)
		ctx.incIndex()
	}
	ctx = ctx.decIndent()
	ctx.structTypeToCodes[uintptr(unsafe.Pointer(c.typ))] = codes
	return codes
}

func (c *StructCode) ToAnonymousOpcode(ctx *compileContext) Opcodes {
	// header => code => structField => code => end
	//                        ^          |
	//                        |__________|
	if c.isRecursive {
		recursive := newRecursiveCode(ctx, &CompiledCode{})
		recursive.Type = c.typ
		ctx.incIndex()
		*ctx.recursiveCodes = append(*ctx.recursiveCodes, recursive)
		return Opcodes{recursive}
	}
	codes := Opcodes{}
	var prevField *Opcode
	for idx, field := range c.fields {
		isFirstField := idx == 0
		isEndField := idx == len(c.fields)-1
		fieldCodes := field.ToAnonymousOpcode(ctx, isFirstField, isEndField)
		for _, code := range fieldCodes {
			if c.isIndirect {
				code.Flags |= IndirectFlags
			}
		}
		if len(codes) > 0 {
			codes.Last().Next = fieldCodes.First()
			fieldCodes.First().Idx = codes.First().Idx
		}
		if prevField != nil {
			prevField.NextField = fieldCodes.First()
		}
		if isEndField {
			if len(codes) > 0 {
				codes.First().End = fieldCodes.Last()
			} else {
				fieldCodes.First().End = fieldCodes.Last()
			}
		}
		prevField = fieldCodes.First()
		codes = append(codes, fieldCodes...)
	}
	return codes
}

func (c *StructCode) removeFieldsByTags(tags runtime.StructTags) {
	fields := make([]*StructFieldCode, 0, len(c.fields))
	for _, field := range c.fields {
		if field.isAnonymous {
			structCode := field.getAnonymousStruct()
			if structCode != nil && !structCode.isRecursive {
				structCode.removeFieldsByTags(tags)
				if len(structCode.fields) > 0 {
					fields = append(fields, field)
				}
				continue
			}
		}
		if tags.ExistsKey(field.key) {
			continue
		}
		fields = append(fields, field)
	}
	c.fields = fields
}

func (c *StructCode) enableIndirect() {
	if c.isIndirect {
		return
	}
	c.isIndirect = true
	if len(c.fields) == 0 {
		return
	}
	structCode := c.fields[0].getStruct()
	if structCode == nil {
		return
	}
	structCode.enableIndirect()
}

type StructFieldCode struct {
	typ                *runtime.Type
	key                string
	tag                *runtime.StructTag
	value              Code
	offset             uintptr
	isAnonymous        bool
	isTaggedKey        bool
	isNilableType      bool
	isNilCheck         bool
	isAddrForMarshaler bool
	isNextOpPtrType    bool
}

func (c *StructFieldCode) getStruct() *StructCode {
	value := c.value
	ptr, ok := value.(*PtrCode)
	if ok {
		value = ptr.value
	}
	structCode, ok := value.(*StructCode)
	if ok {
		return structCode
	}
	return nil
}

func (c *StructFieldCode) getAnonymousStruct() *StructCode {
	if !c.isAnonymous {
		return nil
	}
	return c.getStruct()
}

func (c *StructFieldCode) ToOpcode(ctx *compileContext, isFirstField, isEndField bool) Opcodes {
	var key string
	if ctx.escapeKey {
		rctx := &RuntimeContext{Option: &Option{Flag: HTMLEscapeOption}}
		key = fmt.Sprintf(`%s:`, string(AppendString(rctx, []byte{}, c.key)))
	} else {
		key = fmt.Sprintf(`"%s":`, c.key)
	}
	flags := c.flags()
	if c.isAnonymous {
		flags |= AnonymousKeyFlags
	}
	field := &Opcode{
		Idx:        opcodeOffset(ctx.ptrIndex),
		Flags:      flags,
		Key:        key,
		Offset:     uint32(c.offset),
		Type:       c.typ,
		DisplayIdx: ctx.opcodeIndex,
		Indent:     ctx.indent,
		DisplayKey: c.key,
	}
	ctx.incIndex()
	var codes Opcodes
	if c.isAnonymous {
		anonymCode, ok := c.value.(AnonymousCode)
		if ok {
			codes = anonymCode.ToAnonymousOpcode(ctx.withType(c.typ))
		} else {
			codes = c.value.ToOpcode(ctx.withType(c.typ))
		}
	} else {
		codes = c.value.ToOpcode(ctx.withType(c.typ))
	}
	if isFirstField {
		op := optimizeStructHeader(codes.First(), c.tag)
		field.Op = op
		field.NumBitSize = codes.First().NumBitSize
		field.PtrNum = codes.First().PtrNum
		fieldCodes := Opcodes{field}
		if op.IsMultipleOpHead() {
			field.Next = codes.First()
			fieldCodes = append(fieldCodes, codes...)
		} else {
			ctx.decIndex()
		}
		if isEndField {
			end := &Opcode{
				Op:         OpStructEnd,
				Idx:        opcodeOffset(ctx.ptrIndex),
				DisplayIdx: ctx.opcodeIndex,
				Indent:     ctx.indent,
			}
			fieldCodes.Last().Next = end
			fieldCodes.First().NextField = end
			fieldCodes = append(fieldCodes, end)
			ctx.incIndex()
		}
		return fieldCodes
	}
	op := optimizeStructField(codes.First(), c.tag)
	field.Op = op
	field.NumBitSize = codes.First().NumBitSize
	field.PtrNum = codes.First().PtrNum

	fieldCodes := Opcodes{field}
	if op.IsMultipleOpField() {
		field.Next = codes.First()
		fieldCodes = append(fieldCodes, codes...)
	} else {
		// optimize codes
		ctx.decIndex()
	}
	if isEndField {
		if isEnableStructEndOptimizationType(c.value.Type()) {
			field.Op = field.Op.FieldToEnd()
		} else {
			end := &Opcode{
				Op:         OpStructEnd,
				Idx:        opcodeOffset(ctx.ptrIndex),
				DisplayIdx: ctx.opcodeIndex,
				Indent:     ctx.indent,
			}
			fieldCodes.Last().Next = end
			fieldCodes.First().NextField = end
			fieldCodes = append(fieldCodes, end)
			ctx.incIndex()
		}
	}
	return fieldCodes
}

func (c *StructFieldCode) flags() OpFlags {
	var flags OpFlags
	if c.isTaggedKey {
		flags |= IsTaggedKeyFlags
	}
	if c.isNilableType {
		flags |= IsNilableTypeFlags
	}
	if c.isNilCheck {
		flags |= NilCheckFlags
	}
	if c.isAddrForMarshaler {
		flags |= AddrForMarshalerFlags
	}
	if c.isNextOpPtrType {
		flags |= IsNextOpPtrTypeFlags
	}
	return flags
}

func (c *StructFieldCode) ToAnonymousOpcode(ctx *compileContext, isFirstField, isEndField bool) Opcodes {
	var key string
	if ctx.escapeKey {
		rctx := &RuntimeContext{Option: &Option{Flag: HTMLEscapeOption}}
		key = fmt.Sprintf(`%s:`, string(AppendString(rctx, []byte{}, c.key)))
	} else {
		key = fmt.Sprintf(`"%s":`, c.key)
	}
	flags := c.flags()
	flags |= AnonymousHeadFlags
	if c.isAnonymous {
		flags |= AnonymousKeyFlags
	}
	field := &Opcode{
		Idx:        opcodeOffset(ctx.ptrIndex),
		Flags:      flags,
		Key:        key,
		Offset:     uint32(c.offset),
		Type:       c.typ,
		DisplayIdx: ctx.opcodeIndex,
		Indent:     ctx.indent,
		DisplayKey: c.key,
	}
	ctx.incIndex()
	var codes Opcodes
	if c.isAnonymous {
		anonymCode, ok := c.value.(AnonymousCode)
		if ok {
			codes = anonymCode.ToAnonymousOpcode(ctx.withType(c.typ))
		} else {
			codes = c.value.ToOpcode(ctx.withType(c.typ))
		}
	} else {
		codes = c.value.ToOpcode(ctx.withType(c.typ))
	}
	if isFirstField {
		op := optimizeStructHeader(codes.First(), c.tag)
		field.Op = op
		field.NumBitSize = codes.First().NumBitSize
		field.PtrNum = codes.First().PtrNum
		fieldCodes := Opcodes{field}
		if op.IsMultipleOpHead() {
			field.Next = codes.First()
			fieldCodes = append(fieldCodes, codes...)
		} else {
			ctx.decIndex()
		}
		return fieldCodes
	}
	op := optimizeStructField(codes.First(), c.tag)
	field.Op = op
	field.NumBitSize = codes.First().NumBitSize
	field.PtrNum = codes.First().PtrNum

	fieldCodes := Opcodes{field}
	if op.IsMultipleOpField() {
		field.Next = codes.First()
		fieldCodes = append(fieldCodes, codes...)
	} else {
		// optimize codes
		ctx.decIndex()
	}
	return fieldCodes
}

func isEnableStructEndOptimizationType(typ CodeType2) bool {
	switch typ {
	case CodeTypeInt, CodeTypeUint, CodeTypeFloat, CodeTypeString, CodeTypeBool:
		return true
	default:
		return false
	}
}

type InterfaceCode struct {
	typ   *runtime.Type
	isPtr bool
}

func (c *InterfaceCode) Type() CodeType2 {
	return CodeTypeInterface
}

func (c *InterfaceCode) ToOpcode(ctx *compileContext) Opcodes {
	var code *Opcode
	switch {
	case c.isPtr:
		code = newOpCode(ctx.withType(c.typ), OpInterfacePtr)
	default:
		code = newOpCode(ctx.withType(c.typ), OpInterface)
	}
	ctx.incIndex()
	return Opcodes{code}
}

type MarshalJSONCode struct {
	typ *runtime.Type
}

func (c *MarshalJSONCode) Type() CodeType2 {
	return CodeTypeMarshalJSON
}

func (c *MarshalJSONCode) ToOpcode(ctx *compileContext) Opcodes {
	code := newOpCode(ctx.withType(c.typ), OpMarshalJSON)
	typ := c.typ
	if isPtrMarshalJSONType(typ) {
		code.Flags |= AddrForMarshalerFlags
	}
	if typ.Implements(marshalJSONContextType) || runtime.PtrTo(typ).Implements(marshalJSONContextType) {
		code.Flags |= MarshalerContextFlags
	}
	if isNilableType(typ) {
		code.Flags |= IsNilableTypeFlags
	} else {
		code.Flags &= ^IsNilableTypeFlags
	}
	ctx.incIndex()
	return Opcodes{code}
}

type MarshalTextCode struct {
	typ *runtime.Type
}

func (c *MarshalTextCode) Type() CodeType2 {
	return CodeTypeMarshalText
}

func (c *MarshalTextCode) ToOpcode(ctx *compileContext) Opcodes {
	code := newOpCode(ctx.withType(c.typ), OpMarshalText)
	typ := c.typ
	if !typ.Implements(marshalTextType) && runtime.PtrTo(typ).Implements(marshalTextType) {
		code.Flags |= AddrForMarshalerFlags
	}
	if isNilableType(typ) {
		code.Flags |= IsNilableTypeFlags
	} else {
		code.Flags &= ^IsNilableTypeFlags
	}
	ctx.incIndex()
	return Opcodes{code}
}

type PtrCode struct {
	typ    *runtime.Type
	value  Code
	ptrNum uint8
}

func (c *PtrCode) Type() CodeType2 {
	return CodeTypePtr
}

func (c *PtrCode) ToOpcode(ctx *compileContext) Opcodes {
	codes := c.value.ToOpcode(ctx.withType(c.typ.Elem()))
	codes.First().Op = convertPtrOp(codes.First())
	codes.First().PtrNum = c.ptrNum
	return codes
}

func (c *PtrCode) ToAnonymousOpcode(ctx *compileContext) Opcodes {
	var codes Opcodes
	anonymCode, ok := c.value.(AnonymousCode)
	if ok {
		codes = anonymCode.ToAnonymousOpcode(ctx.withType(c.typ.Elem()))
	} else {
		codes = c.value.ToOpcode(ctx.withType(c.typ.Elem()))
	}
	codes.First().Op = convertPtrOp(codes.First())
	codes.First().PtrNum = c.ptrNum
	return codes
}

func (c *StructCode) compileStructField(ctx *compileContext, tag *runtime.StructTag, isPtr, isOnlyOneFirstField bool) (*StructFieldCode, error) {
	field := tag.Field
	fieldType := runtime.Type2RType(field.Type)
	isIndirectSpecialCase := isPtr && isOnlyOneFirstField
	fieldCode := &StructFieldCode{
		typ:           fieldType,
		key:           tag.Key,
		tag:           tag,
		offset:        field.Offset,
		isAnonymous:   field.Anonymous && !tag.IsTaggedKey,
		isTaggedKey:   tag.IsTaggedKey,
		isNilableType: isNilableType(fieldType),
		isNilCheck:    true,
	}
	switch {
	case isMovePointerPositionFromHeadToFirstMarshalJSONFieldCase(fieldType, isIndirectSpecialCase):
		code, err := compileMarshalJSON(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
		c.isIndirect = false
		c.disableIndirectConversion = true
	case isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(fieldType, isIndirectSpecialCase):
		code, err := compileMarshalText(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
		c.isIndirect = false
		c.disableIndirectConversion = true
	case isPtr && isPtrMarshalJSONType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalJSON() ([]byte, error)
		code, err := compileMarshalJSON(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
	case isPtr && isPtrMarshalTextType(fieldType):
		// *struct{ field T }
		// func (*T) MarshalText() ([]byte, error)
		code, err := compileMarshalText(ctx.withType(fieldType))
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
	}
	return fieldCode, nil
}

func isAssignableIndirect(fieldCode *StructFieldCode, isPtr bool) bool {
	if isPtr {
		return false
	}
	codeType := fieldCode.value.Type()
	if codeType == CodeTypeMarshalJSON {
		return false
	}
	if codeType == CodeTypeMarshalText {
		return false
	}
	return true
}

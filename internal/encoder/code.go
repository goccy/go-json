package encoder

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
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
		end := &Opcode{
			Op:         OpStructEnd,
			Idx:        opcodeOffset(ctx.ptrIndex),
			DisplayIdx: ctx.opcodeIndex,
			Indent:     ctx.indent,
		}
		head := &Opcode{
			Op:         OpStructHead,
			Idx:        opcodeOffset(ctx.ptrIndex),
			NextField:  end,
			Type:       c.typ,
			DisplayIdx: ctx.opcodeIndex,
			Indent:     ctx.indent,
		}
		codes = append(codes, head, end)
		end.PrevField = head
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

func linkRecursiveCode2(ctx *compileContext) {
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

		lastCode.Idx = beforeLastCode.Idx + uintptrSize
		lastCode.ElemIdx = lastCode.Idx + uintptrSize
		lastCode.Length = lastCode.Idx + 2*uintptrSize
		code.End.Next.Op = OpRecursiveEnd

		// extend length to alloc slot for elemIdx + length
		totalLength := uintptr(recursive.TotalLength()) + 3
		nextTotalLength := uintptr(codes.First().TotalLength()) + 3

		compiled.CurLen = totalLength
		compiled.NextLen = nextTotalLength
		compiled.Linked = true
	}
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
		codes = c.value.(AnonymousCode).ToAnonymousOpcode(ctx.withType(c.typ))
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
	if isEndField && !c.isAnonymous {
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
	flags |= AnonymousKeyFlags
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
		codes = c.value.(AnonymousCode).ToAnonymousOpcode(ctx.withType(c.typ))
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
		code = newOpCode(ctx, OpInterfacePtr)
	default:
		code = newOpCode(ctx, OpInterface)
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
	codes := c.value.(AnonymousCode).ToAnonymousOpcode(ctx.withType(c.typ.Elem()))
	codes.First().Op = convertPtrOp(codes.First())
	codes.First().PtrNum = c.ptrNum
	return codes
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
	return &IntCode{typ: ctx.typ, bitSize: 8, isPtr: isPtr}, nil
}

func compileInt162(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 16, isPtr: isPtr}, nil
}

func compileInt322(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 32, isPtr: isPtr}, nil
}

func compileInt642(ctx *compileContext, isPtr bool) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 64, isPtr: isPtr}, nil
}

func compileUint2(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: intSize, isPtr: isPtr}, nil
}

func compileUint82(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 8, isPtr: isPtr}, nil
}

func compileUint162(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 16, isPtr: isPtr}, nil
}

func compileUint322(ctx *compileContext, isPtr bool) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 32, isPtr: isPtr}, nil
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

func compileString2(ctx *compileContext, isPtr bool) (*StringCode, error) {
	return &StringCode{typ: ctx.typ, isPtr: isPtr}, nil
}

func compileBool2(ctx *compileContext, isPtr bool) (*BoolCode, error) {
	return &BoolCode{typ: ctx.typ, isPtr: isPtr}, nil
}

func compileIntString2(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: intSize, isString: true}, nil
}

func compileInt8String2(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 8, isString: true}, nil
}

func compileInt16String2(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 16, isString: true}, nil
}

func compileInt32String2(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 32, isString: true}, nil
}

func compileInt64String2(ctx *compileContext) (*IntCode, error) {
	return &IntCode{typ: ctx.typ, bitSize: 64, isString: true}, nil
}

func compileUintString2(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: intSize, isString: true}, nil
}

func compileUint8String2(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 8, isString: true}, nil
}

func compileUint16String2(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 16, isString: true}, nil
}

func compileUint32String2(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 32, isString: true}, nil
}

func compileUint64String2(ctx *compileContext) (*UintCode, error) {
	return &UintCode{typ: ctx.typ, bitSize: 64, isString: true}, nil
}

func compileSlice2(ctx *compileContext) (*SliceCode, error) {
	elem := ctx.typ.Elem()
	code, err := compileListElem2(ctx.withType(elem))
	if err != nil {
		return nil, err
	}
	if code.Type() == CodeTypeStruct {
		structCode := code.(*StructCode)
		structCode.enableIndirect()
	}
	return &SliceCode{typ: ctx.typ, value: code}, nil
}

func compileArray2(ctx *compileContext) (*ArrayCode, error) {
	typ := ctx.typ
	elem := typ.Elem()
	code, err := compileListElem2(ctx.withType(elem))
	if err != nil {
		return nil, err
	}
	if code.Type() == CodeTypeStruct {
		structCode := code.(*StructCode)
		structCode.enableIndirect()
	}
	return &ArrayCode{typ: ctx.typ, value: code}, nil
}

func compileMap2(ctx *compileContext) (*MapCode, error) {
	typ := ctx.typ
	keyCode, err := compileMapKey(ctx.withType(typ.Key()))
	if err != nil {
		return nil, err
	}
	valueCode, err := compileMapValue2(ctx.withType(typ.Elem()))
	if err != nil {
		return nil, err
	}
	if valueCode.Type() == CodeTypeStruct {
		structCode := valueCode.(*StructCode)
		structCode.enableIndirect()
	}
	return &MapCode{typ: ctx.typ, key: keyCode, value: valueCode}, nil
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
	ptr, ok := code.(*PtrCode)
	if ok {
		return &PtrCode{typ: ctx.typ, value: ptr.value, ptrNum: ptr.ptrNum + 1}, nil
	}
	return &PtrCode{typ: ctx.typ, value: code, ptrNum: 1}, nil
}

func compileListElem2(ctx *compileContext) (Code, error) {
	typ := ctx.typ
	switch {
	case isPtrMarshalJSONType(typ):
		return compileMarshalJSON2(ctx)
	case !typ.Implements(marshalTextType) && runtime.PtrTo(typ).Implements(marshalTextType):
		return compileMarshalText2(ctx)
	case typ.Kind() == reflect.Map:
		return compilePtr2(ctx.withType(runtime.PtrTo(typ)))
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
		return compileMarshalJSON2(ctx)
	case implementsMarshalText(typ):
		return compileMarshalText2(ctx)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return compilePtr2(ctx)
	case reflect.String:
		return compileString2(ctx, false)
	case reflect.Int:
		return compileIntString2(ctx)
	case reflect.Int8:
		return compileInt8String2(ctx)
	case reflect.Int16:
		return compileInt16String2(ctx)
	case reflect.Int32:
		return compileInt32String2(ctx)
	case reflect.Int64:
		return compileInt64String2(ctx)
	case reflect.Uint:
		return compileUintString2(ctx)
	case reflect.Uint8:
		return compileUint8String2(ctx)
	case reflect.Uint16:
		return compileUint16String2(ctx)
	case reflect.Uint32:
		return compileUint32String2(ctx)
	case reflect.Uint64:
		return compileUint64String2(ctx)
	case reflect.Uintptr:
		return compileUintString2(ctx)
	}
	return nil, &errors.UnsupportedTypeError{Type: runtime.RType2Type(typ)}
}

func compileMapValue2(ctx *compileContext) (Code, error) {
	switch ctx.typ.Kind() {
	case reflect.Map:
		return compilePtr2(ctx.withType(runtime.PtrTo(ctx.typ)))
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

func compileStruct2(ctx *compileContext, isPtr bool) (*StructCode, error) {
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
			}
			if isAssignableIndirect(field, isPtr) {
				if indirect {
					structCode.isIndirect = true
				} else {
					structCode.isIndirect = false
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
			structCode := field.getAnonymousStruct()
			if structCode != nil && !structCode.isRecursive {
				for k, v := range getFieldMap(structCode.fields) {
					fieldMap[k] = append(fieldMap[k], v...)
				}
				continue
			}
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
		code, err := compileMarshalJSON2(ctx.withType(fieldType))
		if err != nil {
			return nil, err
		}
		fieldCode.value = code
		fieldCode.isAddrForMarshaler = true
		fieldCode.isNilCheck = false
		c.isIndirect = false
		c.disableIndirectConversion = true
	case isMovePointerPositionFromHeadToFirstMarshalTextFieldCase(fieldType, isIndirectSpecialCase):
		code, err := compileMarshalText2(ctx.withType(fieldType))
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

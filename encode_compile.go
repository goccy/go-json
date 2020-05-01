package json

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"golang.org/x/xerrors"
)

func (e *Encoder) compile(typ *rtype) (*opcode, error) {
	switch typ.Kind() {
	case reflect.Ptr:
		return e.compilePtr(typ)
	case reflect.Slice:
		return e.compileSlice(typ)
	case reflect.Array:
		return e.compileArray(typ)
	case reflect.Map:
		return e.compileMap(typ)
	case reflect.Struct:
		return e.compileStruct(typ)
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
	return nil, xerrors.Errorf("failed to encode type %s: %w", typ.String(), ErrUnsupportedType)
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
	default:
		return newOpCode(opPtr, typ, code)
	}
	return code
}

func (e *Encoder) compilePtr(typ *rtype) (*opcode, error) {
	code, err := e.compile(typ.Elem())
	if err != nil {
		return nil, err
	}
	return e.optimizeStructFieldPtrHead(typ, code), nil
}

func (e *Encoder) compileInt(typ *rtype) (*opcode, error) {
	return newOpCode(opInt, typ, newEndOp()), nil
}

func (e *Encoder) compileInt8(typ *rtype) (*opcode, error) {
	return newOpCode(opInt8, typ, newEndOp()), nil
}

func (e *Encoder) compileInt16(typ *rtype) (*opcode, error) {
	return newOpCode(opInt16, typ, newEndOp()), nil
}

func (e *Encoder) compileInt32(typ *rtype) (*opcode, error) {
	return newOpCode(opInt32, typ, newEndOp()), nil
}

func (e *Encoder) compileInt64(typ *rtype) (*opcode, error) {
	return newOpCode(opInt64, typ, newEndOp()), nil
}

func (e *Encoder) compileUint(typ *rtype) (*opcode, error) {
	return newOpCode(opUint, typ, newEndOp()), nil
}

func (e *Encoder) compileUint8(typ *rtype) (*opcode, error) {
	return newOpCode(opUint8, typ, newEndOp()), nil
}

func (e *Encoder) compileUint16(typ *rtype) (*opcode, error) {
	return newOpCode(opUint16, typ, newEndOp()), nil
}

func (e *Encoder) compileUint32(typ *rtype) (*opcode, error) {
	return newOpCode(opUint32, typ, newEndOp()), nil
}

func (e *Encoder) compileUint64(typ *rtype) (*opcode, error) {
	return newOpCode(opUint64, typ, newEndOp()), nil
}

func (e *Encoder) compileFloat32(typ *rtype) (*opcode, error) {
	return newOpCode(opFloat32, typ, newEndOp()), nil
}

func (e *Encoder) compileFloat64(typ *rtype) (*opcode, error) {
	return newOpCode(opFloat64, typ, newEndOp()), nil
}

func (e *Encoder) compileString(typ *rtype) (*opcode, error) {
	return newOpCode(opString, typ, newEndOp()), nil
}

func (e *Encoder) compileBool(typ *rtype) (*opcode, error) {
	return newOpCode(opBool, typ, newEndOp()), nil
}

func (e *Encoder) compileInterface(typ *rtype) (*opcode, error) {
	return newOpCode(opInterface, typ, newEndOp()), nil
}

func (e *Encoder) compileSlice(typ *rtype) (*opcode, error) {
	elem := typ.Elem()
	size := elem.Size()
	code, err := e.compile(elem)
	if err != nil {
		return nil, err
	}

	// header => opcode => elem => end
	//             ^        |
	//             |________|

	header := newSliceHeaderCode()
	elemCode := &sliceElemCode{opcodeHeader: &opcodeHeader{op: opSliceElem}, size: size}
	end := newOpCode(opSliceEnd, nil, newEndOp())

	header.elem = elemCode
	header.end = end
	header.next = code
	code.beforeLastCode().next = (*opcode)(unsafe.Pointer(elemCode))
	elemCode.next = code
	elemCode.end = end
	return (*opcode)(unsafe.Pointer(header)), nil
}

func (e *Encoder) compileArray(typ *rtype) (*opcode, error) {
	elem := typ.Elem()
	alen := typ.Len()
	size := elem.Size()
	code, err := e.compile(elem)
	if err != nil {
		return nil, err
	}
	// header => opcode => elem => end
	//             ^        |
	//             |________|

	header := newArrayHeaderCode(alen)
	elemCode := &arrayElemCode{
		opcodeHeader: &opcodeHeader{
			op: opArrayElem,
		},
		len:  uintptr(alen),
		size: size,
	}
	end := newOpCode(opArrayEnd, nil, newEndOp())

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

func (e *Encoder) compileMap(typ *rtype) (*opcode, error) {
	// header => code => value => code => key => code => value => code => end
	//                                     ^                       |
	//                                     |_______________________|
	keyType := typ.Key()
	keyCode, err := e.compile(keyType)
	if err != nil {
		return nil, err
	}
	valueType := typ.Elem()
	valueCode, err := e.compile(valueType)
	if err != nil {
		return nil, err
	}
	header := newMapHeaderCode(typ)
	key := newMapKeyCode()
	value := newMapValueCode()
	header.key = key
	header.value = value
	end := newOpCode(opMapEnd, nil, newEndOp())

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

func (e *Encoder) compileStruct(typ *rtype) (*opcode, error) {
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
		valueCode, err := e.compile(fieldType)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf(`"%s":`, keyName)
		fieldCode := &structFieldCode{
			opcodeHeader: &opcodeHeader{
				typ:  fieldType,
				next: valueCode,
			},
			key:    []byte(key),
			offset: field.Offset,
		}
		if fieldIdx == 0 {
			head = fieldCode
			code = (*opcode)(unsafe.Pointer(fieldCode))
			prevField = fieldCode
			if isOmitEmpty {
				fieldCode.op = opStructFieldHeadOmitEmpty
				switch valueCode.op {
				case opInt:
					fieldCode.op = opStructFieldHeadIntOmitEmpty
				case opInt8:
					fieldCode.op = opStructFieldHeadInt8OmitEmpty
				case opInt16:
					fieldCode.op = opStructFieldHeadInt16OmitEmpty
				case opInt32:
					fieldCode.op = opStructFieldHeadInt32OmitEmpty
				case opInt64:
					fieldCode.op = opStructFieldHeadInt64OmitEmpty
				case opUint:
					fieldCode.op = opStructFieldHeadUintOmitEmpty
				case opUint8:
					fieldCode.op = opStructFieldHeadUint8OmitEmpty
				case opUint16:
					fieldCode.op = opStructFieldHeadUint16OmitEmpty
				case opUint32:
					fieldCode.op = opStructFieldHeadUint32OmitEmpty
				case opUint64:
					fieldCode.op = opStructFieldHeadUint64OmitEmpty
				case opFloat32:
					fieldCode.op = opStructFieldHeadFloat32OmitEmpty
				case opFloat64:
					fieldCode.op = opStructFieldHeadFloat64OmitEmpty
				case opString:
					fieldCode.op = opStructFieldHeadStringOmitEmpty
				case opBool:
					fieldCode.op = opStructFieldHeadBoolOmitEmpty
				default:
					code = valueCode.beforeLastCode()
				}
			} else {
				fieldCode.op = opStructFieldHead
				switch valueCode.op {
				case opInt:
					fieldCode.op = opStructFieldHeadInt
				case opInt8:
					fieldCode.op = opStructFieldHeadInt8
				case opInt16:
					fieldCode.op = opStructFieldHeadInt16
				case opInt32:
					fieldCode.op = opStructFieldHeadInt32
				case opInt64:
					fieldCode.op = opStructFieldHeadInt64
				case opUint:
					fieldCode.op = opStructFieldHeadUint
				case opUint8:
					fieldCode.op = opStructFieldHeadUint8
				case opUint16:
					fieldCode.op = opStructFieldHeadUint16
				case opUint32:
					fieldCode.op = opStructFieldHeadUint32
				case opUint64:
					fieldCode.op = opStructFieldHeadUint64
				case opFloat32:
					fieldCode.op = opStructFieldHeadFloat32
				case opFloat64:
					fieldCode.op = opStructFieldHeadFloat64
				case opString:
					fieldCode.op = opStructFieldHeadString
				case opBool:
					fieldCode.op = opStructFieldHeadBool
				default:
					code = valueCode.beforeLastCode()
				}
			}
		} else {
			fieldCode.op = opStructField
			code.next = (*opcode)(unsafe.Pointer(fieldCode))
			prevField.nextField = (*opcode)(unsafe.Pointer(fieldCode))
			prevField = fieldCode
			code = (*opcode)(unsafe.Pointer(fieldCode))
			if isOmitEmpty {
				fieldCode.op = opStructFieldOmitEmpty
				switch valueCode.op {
				case opInt:
					fieldCode.op = opStructFieldIntOmitEmpty
				case opInt8:
					fieldCode.op = opStructFieldInt8OmitEmpty
				case opInt16:
					fieldCode.op = opStructFieldInt16OmitEmpty
				case opInt32:
					fieldCode.op = opStructFieldInt32OmitEmpty
				case opInt64:
					fieldCode.op = opStructFieldInt64OmitEmpty
				case opUint:
					fieldCode.op = opStructFieldUintOmitEmpty
				case opUint8:
					fieldCode.op = opStructFieldUint8OmitEmpty
				case opUint16:
					fieldCode.op = opStructFieldUint16OmitEmpty
				case opUint32:
					fieldCode.op = opStructFieldUint32OmitEmpty
				case opUint64:
					fieldCode.op = opStructFieldUint64OmitEmpty
				case opFloat32:
					fieldCode.op = opStructFieldFloat32OmitEmpty
				case opFloat64:
					fieldCode.op = opStructFieldFloat64OmitEmpty
				case opString:
					fieldCode.op = opStructFieldStringOmitEmpty
				case opBool:
					fieldCode.op = opStructFieldBoolOmitEmpty
				default:
					code = valueCode.beforeLastCode()
				}
			} else {
				switch valueCode.op {
				case opInt:
					fieldCode.op = opStructFieldInt
				case opInt8:
					fieldCode.op = opStructFieldInt8
				case opInt16:
					fieldCode.op = opStructFieldInt16
				case opInt32:
					fieldCode.op = opStructFieldInt32
				case opInt64:
					fieldCode.op = opStructFieldInt64
				case opUint:
					fieldCode.op = opStructFieldUint
				case opUint8:
					fieldCode.op = opStructFieldUint8
				case opUint16:
					fieldCode.op = opStructFieldUint16
				case opUint32:
					fieldCode.op = opStructFieldUint32
				case opUint64:
					fieldCode.op = opStructFieldUint64
				case opFloat32:
					fieldCode.op = opStructFieldFloat32
				case opFloat64:
					fieldCode.op = opStructFieldFloat64
				case opString:
					fieldCode.op = opStructFieldString
				case opBool:
					fieldCode.op = opStructFieldBool
				default:
					code = valueCode.beforeLastCode()
				}
			}
		}
		fieldIdx++
	}

	structEndCode := newOpCode(opStructEnd, nil, nil)

	if prevField != nil && prevField.nextField == nil {
		prevField.nextField = structEndCode
	}

	// no struct field
	if head == nil {
		head = &structFieldCode{
			opcodeHeader: &opcodeHeader{
				op:  opStructFieldHead,
				typ: typ,
			},
			nextField: structEndCode,
		}
		code = (*opcode)(unsafe.Pointer(head))
	}
	head.end = structEndCode
	code.next = structEndCode
	structEndCode.next = newEndOp()
	return (*opcode)(unsafe.Pointer(head)), nil
}

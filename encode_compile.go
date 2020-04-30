package json

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"golang.org/x/xerrors"
)

func (e *Encoder) compileOp(typ *rtype) (*opcode, error) {
	switch typ.Kind() {
	case reflect.Ptr:
		return e.compilePtrOp(typ)
	case reflect.Slice:
		return e.compileSliceOp(typ)
	case reflect.Array:
		return e.compileArrayOp(typ)
	case reflect.Map:
		return e.compileMapOp(typ)
	case reflect.Struct:
		return e.compileStructOp(typ)
	case reflect.Int:
		return e.compileIntOp(typ)
	case reflect.Int8:
		return e.compileInt8Op(typ)
	case reflect.Int16:
		return e.compileInt16Op(typ)
	case reflect.Int32:
		return e.compileInt32Op(typ)
	case reflect.Int64:
		return e.compileInt64Op(typ)
	case reflect.Uint:
		return e.compileUintOp(typ)
	case reflect.Uint8:
		return e.compileUint8Op(typ)
	case reflect.Uint16:
		return e.compileUint16Op(typ)
	case reflect.Uint32:
		return e.compileUint32Op(typ)
	case reflect.Uint64:
		return e.compileUint64Op(typ)
	case reflect.Uintptr:
		return e.compileUintOp(typ)
	case reflect.Float32:
		return e.compileFloat32Op(typ)
	case reflect.Float64:
		return e.compileFloat64Op(typ)
	case reflect.String:
		return e.compileStringOp(typ)
	case reflect.Bool:
		return e.compileBoolOp(typ)
	case reflect.Interface:
		return e.compileInterfaceOp(typ)
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
	default:
		return newOpCode(opPtr, typ, code)
	}
	return code
}

func (e *Encoder) compilePtrOp(typ *rtype) (*opcode, error) {
	code, err := e.compileOp(typ.Elem())
	if err != nil {
		return nil, err
	}
	return e.optimizeStructFieldPtrHead(typ, code), nil
}

func (e *Encoder) compileIntOp(typ *rtype) (*opcode, error) {
	return newOpCode(opInt, typ, newEndOp()), nil
}

func (e *Encoder) compileInt8Op(typ *rtype) (*opcode, error) {
	return newOpCode(opInt8, typ, newEndOp()), nil
}

func (e *Encoder) compileInt16Op(typ *rtype) (*opcode, error) {
	return newOpCode(opInt16, typ, newEndOp()), nil
}

func (e *Encoder) compileInt32Op(typ *rtype) (*opcode, error) {
	return newOpCode(opInt32, typ, newEndOp()), nil
}

func (e *Encoder) compileInt64Op(typ *rtype) (*opcode, error) {
	return newOpCode(opInt64, typ, newEndOp()), nil
}

func (e *Encoder) compileUintOp(typ *rtype) (*opcode, error) {
	return newOpCode(opUint, typ, newEndOp()), nil
}

func (e *Encoder) compileUint8Op(typ *rtype) (*opcode, error) {
	return newOpCode(opUint8, typ, newEndOp()), nil
}

func (e *Encoder) compileUint16Op(typ *rtype) (*opcode, error) {
	return newOpCode(opUint16, typ, newEndOp()), nil
}

func (e *Encoder) compileUint32Op(typ *rtype) (*opcode, error) {
	return newOpCode(opUint32, typ, newEndOp()), nil
}

func (e *Encoder) compileUint64Op(typ *rtype) (*opcode, error) {
	return newOpCode(opUint64, typ, newEndOp()), nil
}

func (e *Encoder) compileFloat32Op(typ *rtype) (*opcode, error) {
	return newOpCode(opFloat32, typ, newEndOp()), nil
}

func (e *Encoder) compileFloat64Op(typ *rtype) (*opcode, error) {
	return newOpCode(opFloat64, typ, newEndOp()), nil
}

func (e *Encoder) compileStringOp(typ *rtype) (*opcode, error) {
	return newOpCode(opString, typ, newEndOp()), nil
}

func (e *Encoder) compileBoolOp(typ *rtype) (*opcode, error) {
	return newOpCode(opBool, typ, newEndOp()), nil
}

func (e *Encoder) compileInterfaceOp(typ *rtype) (*opcode, error) {
	return newOpCode(opInterface, typ, newEndOp()), nil
}

func (e *Encoder) compileSliceOp(typ *rtype) (*opcode, error) {
	elem := typ.Elem()
	size := elem.Size()
	code, err := e.compileOp(elem)
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

func (e *Encoder) compileArrayOp(typ *rtype) (*opcode, error) {
	elem := typ.Elem()
	alen := typ.Len()
	size := elem.Size()
	code, err := e.compileOp(elem)
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

func (e *Encoder) compileMapOp(typ *rtype) (*opcode, error) {
	// header => code => value => code => key => code => value => code => end
	//                                     ^                       |
	//                                     |_______________________|
	keyType := typ.Key()
	keyCode, err := e.compileOp(keyType)
	if err != nil {
		return nil, err
	}
	valueType := typ.Elem()
	valueCode, err := e.compileOp(valueType)
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

func (e *Encoder) compileStructOp(typ *rtype) (*opcode, error) {
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
		fieldType := type2rtype(field.Type)
		valueCode, err := e.compileOp(fieldType)
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
			fieldCode.op = opStructFieldHead
			head = fieldCode
			code = (*opcode)(unsafe.Pointer(fieldCode))
			prevField = fieldCode
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
		} else {
			fieldCode.op = opStructField
			code.next = (*opcode)(unsafe.Pointer(fieldCode))
			prevField.nextField = (*opcode)(unsafe.Pointer(fieldCode))
			prevField = fieldCode
			code = (*opcode)(unsafe.Pointer(fieldCode))
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
		prevField.nextField = newEndOp()
		fieldIdx++
	}
	structEndCode := newOpCode(opStructEnd, nil, nil)
	head.end = structEndCode
	code.next = structEndCode
	structEndCode.next = newEndOp()
	return (*opcode)(unsafe.Pointer(head)), nil
}

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
		return nil, errCompileSlowPath
	}
	return nil, xerrors.Errorf("failed to encode type %s: %w", typ.String(), ErrUnsupportedType)
}

func (e *Encoder) compilePtrOp(typ *rtype) (*opcode, error) {
	elem := typ.Elem()
	code, err := e.compileOp(elem)
	if err != nil {
		return nil, err
	}
	return &opcode{
		opcodeHeader: &opcodeHeader{
			op:   opPtr,
			typ:  typ,
			next: code,
		},
	}, nil
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

func (e *Encoder) compileSliceOp(typ *rtype) (*opcode, error) {
	elem := typ.Elem()
	size := elem.Size()
	code, err := e.compileOp(elem)
	if err != nil {
		return nil, err
	}

	// header => firstElem => opcode => elem => end
	//                          ^        |
	//                          |________|

	header := &opcode{opcodeHeader: &opcodeHeader{op: opSliceHead}}
	firstElem := &sliceElemCode{opcodeHeader: &opcodeHeader{op: opSliceElemFirst}}
	elemCode := &sliceElemCode{opcodeHeader: &opcodeHeader{op: opSliceElem}, size: size}
	end := &opcode{opcodeHeader: &opcodeHeader{op: opSliceEnd}}

	header.next = (*opcode)(unsafe.Pointer(firstElem))
	firstElem.next = code
	firstElem.elem = elemCode
	code.beforeLastCode().next = (*opcode)(unsafe.Pointer(elemCode))
	elemCode.next = code
	elemCode.end = end
	end.next = &opcode{opcodeHeader: &opcodeHeader{op: opEnd}}
	return (*opcode)(unsafe.Pointer(header)), nil
}

func (e *Encoder) compileStructOp(typ *rtype) (*opcode, error) {
	// header => firstField => structField => end
	//                          ^        |
	//                          |________|
	fieldNum := typ.NumField()
	fieldIdx := 0
	header := &opcode{opcodeHeader: &opcodeHeader{op: opStructHead}}
	code := header
	var prevField *structFieldCode
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
		if fieldIdx == 0 {
			fieldCode := &structFieldCode{
				opcodeHeader: &opcodeHeader{
					op:   opStructFieldFirst,
					typ:  fieldType,
					next: valueCode,
				},
				key:    key,
				offset: field.Offset,
			}
			code.next = (*opcode)(unsafe.Pointer(fieldCode))
			prevField = fieldCode
			if valueCode.op == opInt {
				fieldCode.op = opStructFieldFirstInt
				code = (*opcode)(unsafe.Pointer(fieldCode))
			} else if valueCode.op == opString {
				fieldCode.op = opStructFieldFirstString
				code = (*opcode)(unsafe.Pointer(fieldCode))
			} else {
				code = valueCode.beforeLastCode()
			}
		} else {
			fieldCode := &structFieldCode{
				opcodeHeader: &opcodeHeader{
					op:   opStructField,
					typ:  fieldType,
					next: valueCode,
				},
				key:    key,
				offset: field.Offset,
			}
			code.next = (*opcode)(unsafe.Pointer(fieldCode))
			prevField.nextField = (*opcode)(unsafe.Pointer(fieldCode))
			prevField = fieldCode
			if valueCode.op == opInt {
				fieldCode.op = opStructFieldInt
				code = (*opcode)(unsafe.Pointer(fieldCode))
			} else if valueCode.op == opString {
				fieldCode.op = opStructFieldString
				code = (*opcode)(unsafe.Pointer(fieldCode))
			} else {
				code = valueCode.beforeLastCode()
			}
		}
		prevField.nextField = &opcode{opcodeHeader: &opcodeHeader{op: opEnd}}
		fieldIdx++
	}
	structEndCode := &opcode{opcodeHeader: &opcodeHeader{op: opStructEnd}}
	code.next = structEndCode
	structEndCode.next = &opcode{opcodeHeader: &opcodeHeader{op: opEnd}}
	return header, nil
}

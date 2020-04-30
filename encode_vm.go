package json

import (
	"reflect"
	"unsafe"
)

func (e *Encoder) run(code *opcode) error {
	for {
		switch code.op {
		case opPtr:
			ptr := code.ptr
			code = code.next
			code.ptr = e.ptrToPtr(ptr)
		case opInt:
			e.encodeInt(e.ptrToInt(code.ptr))
			code = code.next
		case opInt8:
			e.encodeInt8(e.ptrToInt8(code.ptr))
			code = code.next
		case opInt16:
			e.encodeInt16(e.ptrToInt16(code.ptr))
			code = code.next
		case opInt32:
			e.encodeInt32(e.ptrToInt32(code.ptr))
			code = code.next
		case opInt64:
			e.encodeInt64(e.ptrToInt64(code.ptr))
			code = code.next
		case opUint:
			e.encodeUint(e.ptrToUint(code.ptr))
			code = code.next
		case opUint8:
			e.encodeUint8(e.ptrToUint8(code.ptr))
			code = code.next
		case opUint16:
			e.encodeUint16(e.ptrToUint16(code.ptr))
			code = code.next
		case opUint32:
			e.encodeUint32(e.ptrToUint32(code.ptr))
			code = code.next
		case opUint64:
			e.encodeUint64(e.ptrToUint64(code.ptr))
			code = code.next
		case opFloat32:
			e.encodeFloat32(e.ptrToFloat32(code.ptr))
			code = code.next
		case opFloat64:
			e.encodeFloat64(e.ptrToFloat64(code.ptr))
			code = code.next
		case opString:
			e.encodeEscapedString(e.ptrToString(code.ptr))
			code = code.next
		case opBool:
			e.encodeBool(e.ptrToBool(code.ptr))
			code = code.next
		case opInterface:
			ptr := code.ptr
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(ptr),
			}))
			vv := reflect.ValueOf(v).Interface()
			header := (*interfaceHeader)(unsafe.Pointer(&vv))
			typ := header.typ
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			c, err := e.compile(typ)
			if err != nil {
				return err
			}
			c.ptr = uintptr(header.ptr)
			c.beforeLastCode().next = code.next
			code = c
		case opSliceHead:
			p := code.ptr
			headerCode := code.toSliceHeaderCode()
			if p == 0 {
				e.encodeString("null")
				code = headerCode.end.next
			} else {
				e.encodeByte('[')
				header := (*reflect.SliceHeader)(unsafe.Pointer(p))
				headerCode.elem.set(header)
				if header.Len > 0 {
					code = code.next
					code.ptr = header.Data
				} else {
					e.encodeByte(']')
					code = headerCode.end.next
				}
			}
		case opSliceElem:
			c := code.toSliceElemCode()
			c.idx++
			if c.idx < c.len {
				e.encodeByte(',')
				code = code.next
				code.ptr = c.data + c.idx*c.size
			} else {
				e.encodeByte(']')
				code = c.end.next
			}
		case opArrayHead:
			p := code.ptr
			headerCode := code.toArrayHeaderCode()
			if p == 0 {
				e.encodeString("null")
				code = headerCode.end.next
			} else {
				e.encodeByte('[')
				if headerCode.len > 0 {
					code = code.next
					code.ptr = p
					headerCode.elem.ptr = p
				} else {
					e.encodeByte(']')
					code = headerCode.end.next
				}
			}
		case opArrayElem:
			c := code.toArrayElemCode()
			c.idx++
			if c.idx < c.len {
				e.encodeByte(',')
				code = code.next
				code.ptr = c.ptr + c.idx*c.size
			} else {
				e.encodeByte(']')
				code = c.end.next
			}
		case opMapHead:
			ptr := code.ptr
			mapHeadCode := code.toMapHeadCode()
			if ptr == 0 {
				e.encodeString("null")
				code = mapHeadCode.end.next
			} else {
				e.encodeByte('{')
				mlen := maplen(unsafe.Pointer(ptr))
				if mlen > 0 {
					iter := mapiterinit(code.typ, unsafe.Pointer(ptr))
					mapHeadCode.key.set(mlen, iter)
					mapHeadCode.value.set(iter)
					key := mapiterkey(iter)
					code.next.ptr = uintptr(key)
					code = code.next
				} else {
					e.encodeByte('}')
					code = mapHeadCode.end.next
				}
			}
		case opMapKey:
			c := code.toMapKeyCode()
			c.idx++
			if c.idx < c.len {
				e.encodeByte(',')
				key := mapiterkey(c.iter)
				c.next.ptr = uintptr(key)
				code = c.next
			} else {
				e.encodeByte('}')
				code = c.end.next
			}
		case opMapValue:
			e.encodeByte(':')
			c := code.toMapValueCode()
			value := mapitervalue(c.iter)
			c.next.ptr = uintptr(value)
			mapiternext(c.iter)
			code = c.next
		case opStructFieldPtrHead:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHead:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				code = field.next
				code.ptr = field.ptr + field.offset
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadInt:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt(e.ptrToInt(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt8:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt8(e.ptrToInt8(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt16:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt16(e.ptrToInt16(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt32(e.ptrToInt32(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt64(e.ptrToInt64(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint(e.ptrToUint(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint8:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint8(e.ptrToUint8(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint16:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint16(e.ptrToUint16(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint32(e.ptrToUint32(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint64(e.ptrToUint64(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadFloat32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadFloat32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeFloat32(e.ptrToFloat32(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadFloat64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadFloat64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeFloat64(e.ptrToFloat64(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadString:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadString:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeEscapedString(e.ptrToString(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructFieldPtrHeadBool:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadBool:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeBool(e.ptrToBool(field.ptr + field.offset))
				field.nextField.ptr = field.ptr
				code = field.next
			}
		case opStructField:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			e.encodeBytes(c.key)
			code = code.next
			code.ptr = c.ptr + c.offset
			c.nextField.ptr = c.ptr
		case opStructFieldInt:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt(e.ptrToInt(c.ptr + c.offset))
			code = code.next
		case opStructFieldInt8:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt8(e.ptrToInt8(c.ptr + c.offset))
			code = code.next
		case opStructFieldInt16:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt16(e.ptrToInt16(c.ptr + c.offset))
			code = code.next
		case opStructFieldInt32:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt32(e.ptrToInt32(c.ptr + c.offset))
			code = code.next
		case opStructFieldInt64:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt64(e.ptrToInt64(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint(e.ptrToUint(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint8:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint8(e.ptrToUint8(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint16:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint16(e.ptrToUint16(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint32:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint32(e.ptrToUint32(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint64:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint64(e.ptrToUint64(c.ptr + c.offset))
			code = code.next
		case opStructFieldFloat32:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeFloat32(e.ptrToFloat32(c.ptr + c.offset))
			code = code.next
		case opStructFieldFloat64:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeFloat64(e.ptrToFloat64(c.ptr + c.offset))
			code = code.next
		case opStructFieldString:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeEscapedString(e.ptrToString(c.ptr + c.offset))
			code = code.next
		case opStructFieldBool:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeBool(e.ptrToBool(c.ptr + c.offset))
			code = code.next
		case opStructEnd:
			e.encodeByte('}')
			code = code.next
		case opEnd:
			goto END
		}
	}
END:
	return nil
}

func (e *Encoder) ptrToPtr(p uintptr) uintptr     { return *(*uintptr)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt(p uintptr) int         { return *(*int)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt8(p uintptr) int8       { return *(*int8)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt16(p uintptr) int16     { return *(*int16)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt32(p uintptr) int32     { return *(*int32)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt64(p uintptr) int64     { return *(*int64)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint(p uintptr) uint       { return *(*uint)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint8(p uintptr) uint8     { return *(*uint8)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint16(p uintptr) uint16   { return *(*uint16)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint32(p uintptr) uint32   { return *(*uint32)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint64(p uintptr) uint64   { return *(*uint64)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToFloat32(p uintptr) float32 { return *(*float32)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToFloat64(p uintptr) float64 { return *(*float64)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToBool(p uintptr) bool       { return *(*bool)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToByte(p uintptr) byte       { return *(*byte)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToBytes(p uintptr) []byte    { return *(*[]byte)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToString(p uintptr) string   { return *(*string)(unsafe.Pointer(p)) }

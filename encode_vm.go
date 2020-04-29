package json

import (
	"reflect"
	"unsafe"
)

func (e *Encoder) run(code *opcode) error {
	//fmt.Println("================")
	//fmt.Println(code.dump())
	//fmt.Println("================")
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
		case opSliceHead:
			p := code.ptr
			if p == 0 {
				e.encodeString("null")
				code = code.next.toSliceElemCode().end
			} else {
				e.encodeByte('[')
				header := (*reflect.SliceHeader)(unsafe.Pointer(p))
				firstElem := code.next.toSliceElemCode()
				firstElem.set(header)
				firstElem.elem.set(header)
				code = code.next
			}
		case opSliceElemFirst:
			c := code.toSliceElemCode()
			if c.len > 0 {
				code = code.next
				code.ptr = c.data
			} else {
				code = c.end
			}
		case opSliceElem:
			c := code.toSliceElemCode()
			c.idx++
			if c.idx < c.len {
				e.encodeByte(',')
				code = code.next
				code.ptr = c.data + c.idx*c.size
			} else {
				code = c.end
			}
		case opSliceEnd:
			e.encodeByte(']')
			code = code.next
		case opStructHead:
			ptr := code.ptr
			if ptr == 0 {
				e.encodeString("null")
				code = code.toStructHeaderCode().end
			} else {
				e.encodeByte('{')
				code = code.next
				code.ptr = ptr
			}
		case opStructFieldFirst:
			c := code.toStructFieldCode()
			e.encodeString(c.key)
			code = code.next
			code.ptr = c.ptr + c.offset
			c.nextField.ptr = c.ptr
		case opStructFieldFirstInt:
			c := code.toStructFieldCode()
			e.encodeString(c.key)
			c.nextField.ptr = c.ptr
			e.encodeInt(e.ptrToInt(c.ptr + c.offset))
			code = code.next
		case opStructFieldFirstString:
			c := code.toStructFieldCode()
			e.encodeString(c.key)
			c.nextField.ptr = c.ptr
			e.encodeEscapedString(e.ptrToString(c.ptr + c.offset))
			code = code.next
		case opStructField:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			e.encodeString(c.key)
			code = code.next
			code.ptr = c.ptr + c.offset
			c.nextField.ptr = c.ptr
		case opStructFieldInt:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeString(c.key)
			e.encodeInt(e.ptrToInt(c.ptr + c.offset))
			code = code.next
		case opStructFieldString:
			e.encodeByte(',')
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeString(c.key)
			e.encodeEscapedString(e.ptrToString(c.ptr + c.offset))
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

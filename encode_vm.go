package json

import (
	"bytes"
	"encoding"
	"math"
	"reflect"
	"strconv"
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
			v := e.ptrToFloat64(code.ptr)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return &UnsupportedValueError{
					Value: reflect.ValueOf(v),
					Str:   strconv.FormatFloat(v, 'g', -1, 64),
				}
			}
			e.encodeFloat64(v)
			code = code.next
		case opString:
			e.encodeString(e.ptrToString(code.ptr))
			code = code.next
		case opBool:
			e.encodeBool(e.ptrToBool(code.ptr))
			code = code.next
		case opInterface:
			ifaceCode := code.toInterfaceCode()
			ptr := ifaceCode.ptr
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: ifaceCode.typ,
				ptr: unsafe.Pointer(ptr),
			}))
			rv := reflect.ValueOf(v)
			if rv.IsNil() {
				e.encodeNull()
				code = ifaceCode.next
				break
			}
			vv := rv.Interface()
			header := (*interfaceHeader)(unsafe.Pointer(&vv))
			typ := header.typ
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			e.indent = ifaceCode.indent
			var c *opcode
			if typ.Kind() == reflect.Map {
				code, err := e.compileMap(typ, false, ifaceCode.root, e.enabledIndent)
				if err != nil {
					return err
				}
				c = code
			} else {
				code, err := e.compile(typ, ifaceCode.root, e.enabledIndent)
				if err != nil {
					return err
				}
				c = code
			}
			c.ptr = uintptr(header.ptr)
			c.beforeLastCode().next = code.next
			code = c
		case opMarshalJSON:
			ptr := code.ptr
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(ptr),
			}))
			b, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return &MarshalerError{
					Type: rtype2type(code.typ),
					Err:  err,
				}
			}
			var buf bytes.Buffer
			if err := compact(&buf, b, true); err != nil {
				return err
			}
			e.encodeBytes(buf.Bytes())
			code = code.next
			code.ptr = ptr
		case opMarshalText:
			ptr := code.ptr
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(ptr),
			}))
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return &MarshalerError{
					Type: rtype2type(code.typ),
					Err:  err,
				}
			}
			e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
			code = code.next
			code.ptr = ptr
		case opSliceHead:
			p := code.ptr
			headerCode := code.toSliceHeaderCode()
			if p == 0 {
				e.encodeNull()
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
		case opSliceHeadIndent:
			p := code.ptr
			headerCode := code.toSliceHeaderCode()
			if p == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = headerCode.end.next
			} else {
				header := (*reflect.SliceHeader)(unsafe.Pointer(p))
				headerCode.elem.set(header)
				if header.Len > 0 {
					e.encodeBytes([]byte{'[', '\n'})
					e.encodeIndent(code.indent + 1)
					code = code.next
					code.ptr = header.Data
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{'[', ']', '\n'})
					code = headerCode.end.next
				}
			}
		case opRootSliceHeadIndent:
			p := code.ptr
			headerCode := code.toSliceHeaderCode()
			if p == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = headerCode.end.next
			} else {
				header := (*reflect.SliceHeader)(unsafe.Pointer(p))
				headerCode.elem.set(header)
				if header.Len > 0 {
					e.encodeBytes([]byte{'[', '\n'})
					e.encodeIndent(code.indent + 1)
					code = code.next
					code.ptr = header.Data
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{'[', ']'})
					code = headerCode.end.next
				}
			}
		case opSliceElemIndent:
			c := code.toSliceElemCode()
			c.idx++
			if c.idx < c.len {
				e.encodeBytes([]byte{',', '\n'})
				e.encodeIndent(code.indent + 1)
				code = code.next
				code.ptr = c.data + c.idx*c.size
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{']', '\n'})
				code = c.end.next
			}
		case opRootSliceElemIndent:
			c := code.toSliceElemCode()
			c.idx++
			if c.idx < c.len {
				e.encodeBytes([]byte{',', '\n'})
				e.encodeIndent(code.indent + 1)
				code = code.next
				code.ptr = c.data + c.idx*c.size
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{']'})
				code = c.end.next
			}
		case opArrayHead:
			p := code.ptr
			headerCode := code.toArrayHeaderCode()
			if p == 0 {
				e.encodeNull()
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
		case opArrayHeadIndent:
			p := code.ptr
			headerCode := code.toArrayHeaderCode()
			if p == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = headerCode.end.next
			} else {
				e.encodeBytes([]byte{'[', '\n'})
				if headerCode.len > 0 {
					e.encodeIndent(code.indent + 1)
					code = code.next
					code.ptr = p
					headerCode.elem.ptr = p
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{']', '\n'})
					code = headerCode.end.next
				}
			}
		case opArrayElemIndent:
			c := code.toArrayElemCode()
			c.idx++
			if c.idx < c.len {
				e.encodeBytes([]byte{',', '\n'})
				e.encodeIndent(code.indent + 1)
				code = code.next
				code.ptr = c.ptr + c.idx*c.size
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{']', '\n'})
				code = c.end.next
			}
		case opMapHead:
			ptr := code.ptr
			mapHeadCode := code.toMapHeadCode()
			if ptr == 0 {
				e.encodeNull()
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
		case opMapHeadLoad:
			ptr := code.ptr
			mapHeadCode := code.toMapHeadCode()
			if ptr == 0 {
				e.encodeNull()
				code = mapHeadCode.end.next
			} else {
				// load pointer
				ptr = uintptr(*(*unsafe.Pointer)(unsafe.Pointer(ptr)))
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
		case opMapHeadIndent:
			ptr := code.ptr
			mapHeadCode := code.toMapHeadCode()
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = mapHeadCode.end.next
			} else {
				mlen := maplen(unsafe.Pointer(ptr))
				if mlen > 0 {
					e.encodeBytes([]byte{'{', '\n'})
					iter := mapiterinit(code.typ, unsafe.Pointer(ptr))
					mapHeadCode.key.set(mlen, iter)
					mapHeadCode.value.set(iter)
					key := mapiterkey(iter)
					code.next.ptr = uintptr(key)
					code = code.next
					e.encodeIndent(code.indent)
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{'{', '}', '\n'})
					code = mapHeadCode.end.next
				}
			}
		case opMapHeadLoadIndent:
			ptr := code.ptr
			mapHeadCode := code.toMapHeadCode()
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = mapHeadCode.end.next
			} else {
				// load pointer
				ptr = uintptr(*(*unsafe.Pointer)(unsafe.Pointer(ptr)))
				mlen := maplen(unsafe.Pointer(ptr))
				if mlen > 0 {
					e.encodeBytes([]byte{'{', '\n'})
					iter := mapiterinit(code.typ, unsafe.Pointer(ptr))
					mapHeadCode.key.set(mlen, iter)
					mapHeadCode.value.set(iter)
					key := mapiterkey(iter)
					code.next.ptr = uintptr(key)
					code = code.next
					e.encodeIndent(code.indent)
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{'{', '}', '\n'})
					code = mapHeadCode.end.next
				}
			}
		case opRootMapHeadIndent:
			ptr := code.ptr
			mapHeadCode := code.toMapHeadCode()
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = mapHeadCode.end.next
			} else {
				mlen := maplen(unsafe.Pointer(ptr))
				if mlen > 0 {
					e.encodeBytes([]byte{'{', '\n'})
					iter := mapiterinit(code.typ, unsafe.Pointer(ptr))
					mapHeadCode.key.set(mlen, iter)
					mapHeadCode.value.set(iter)
					key := mapiterkey(iter)
					code.next.ptr = uintptr(key)
					code = code.next
					e.encodeIndent(code.indent)
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{'{', '}'})
					code = mapHeadCode.end.next
				}
			}
		case opMapKeyIndent:
			c := code.toMapKeyCode()
			c.idx++
			if c.idx < c.len {
				e.encodeBytes([]byte{',', '\n'})
				e.encodeIndent(code.indent)
				key := mapiterkey(c.iter)
				c.next.ptr = uintptr(key)
				code = c.next
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent - 1)
				e.encodeBytes([]byte{'}', '\n'})
				code = c.end.next
			}
		case opRootMapKeyIndent:
			c := code.toMapKeyCode()
			c.idx++
			if c.idx < c.len {
				e.encodeBytes([]byte{',', '\n'})
				e.encodeIndent(code.indent)
				key := mapiterkey(c.iter)
				c.next.ptr = uintptr(key)
				code = c.next
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent - 1)
				e.encodeBytes([]byte{'}'})
				code = c.end.next
			}
		case opMapValueIndent:
			e.encodeBytes([]byte{':', ' '})
			c := code.toMapValueCode()
			value := mapitervalue(c.iter)
			c.next.ptr = uintptr(value)
			mapiternext(c.iter)
			code = c.next
		case opStructFieldRecursive:
			recursive := code.toRecursiveCode()
			if err := e.run(newRecursiveCode(recursive)); err != nil {
				return err
			}
			code = recursive.next
		case opStructFieldPtrHead:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHead:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				if !field.anonymousKey {
					e.encodeBytes(field.key)
				}
				code = field.next
				code.ptr = ptr
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadInt:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt(e.ptrToInt(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadInt:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt(e.ptrToInt(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt8:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt8(e.ptrToInt8(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt8:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadInt8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt8(e.ptrToInt8(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt16:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt16(e.ptrToInt16(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt16:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadInt16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt16(e.ptrToInt16(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt32(e.ptrToInt32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadInt32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt32(e.ptrToInt32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt64(e.ptrToInt64(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadInt64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt64(e.ptrToInt64(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint(e.ptrToUint(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadUint:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint(e.ptrToUint(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint8:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint8(e.ptrToUint8(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint8:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadUint8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint8(e.ptrToUint8(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint16:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint16(e.ptrToUint16(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint16:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadUint16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint16(e.ptrToUint16(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint32(e.ptrToUint32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadUint32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint32(e.ptrToUint32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint64(e.ptrToUint64(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadUint64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint64(e.ptrToUint64(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadFloat32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadFloat32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeFloat32(e.ptrToFloat32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadFloat32:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadFloat32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeFloat32(e.ptrToFloat32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadFloat64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadFloat64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeFloat64(v)
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadFloat64:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadFloat64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				e.encodeBytes(field.key)
				e.encodeFloat64(v)
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadString:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadString:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(e.ptrToString(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadString:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadString:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeString(e.ptrToString(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadBool:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadBool:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeBool(e.ptrToBool(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadBool:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldAnonymousHeadBool:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeBool(e.ptrToBool(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadIndent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				code = field.next
				code.ptr = ptr
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadIntIndent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadIntIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeInt(e.ptrToInt(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt8Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt8Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeInt8(e.ptrToInt8(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt16Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt16Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeInt16(e.ptrToInt16(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt32Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt32Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeInt32(e.ptrToInt32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadInt64Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadInt64Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeInt64(e.ptrToInt64(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUintIndent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUintIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeUint(e.ptrToUint(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint8Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint8Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeUint8(e.ptrToUint8(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint16Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint16Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeUint16(e.ptrToUint16(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint32Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint32Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeUint32(e.ptrToUint32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadUint64Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadUint64Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeUint64(e.ptrToUint64(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadFloat32Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadFloat32Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeFloat32(e.ptrToFloat32(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadFloat64Indent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadFloat64Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeFloat64(v)
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadStringIndent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadStringIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(e.ptrToString(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadBoolIndent:
			code.ptr = e.ptrToPtr(code.ptr)
			fallthrough
		case opStructFieldHeadBoolIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeBool(e.ptrToBool(ptr))
				field.nextField.ptr = ptr
				code = field.next
			}
		case opStructFieldPtrHeadOmitEmpty:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmpty:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				p := ptr + field.offset
				if p == 0 || *(*uintptr)(unsafe.Pointer(p)) == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					code = field.next
					code.ptr = p
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmpty:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmpty:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				p := ptr + field.offset
				if p == 0 || *(*uintptr)(unsafe.Pointer(p)) == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					code = field.next
					code.ptr = p
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToInt(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToInt(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt8:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToInt8(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt8(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt8:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToInt8(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt8(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt16:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToInt16(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt16(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt16:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToInt16(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt16(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt32:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToInt32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt32(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt32:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToInt32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt32(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt64:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToInt64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt64(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt64:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToInt64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeInt64(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToUint(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToUint(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint8:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToUint8(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint8(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint8:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint8:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToUint8(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint8(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint16:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToUint16(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint16(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint16:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint16:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToUint16(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint16(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint32:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToUint32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint32(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint32:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToUint32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint32(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint64:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToUint64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint64(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint64:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToUint64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeUint64(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyFloat32:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToFloat32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeFloat32(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat32:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat32:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToFloat32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeFloat32(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyFloat64:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToFloat64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return &UnsupportedValueError{
							Value: reflect.ValueOf(v),
							Str:   strconv.FormatFloat(v, 'g', -1, 64),
						}
					}
					e.encodeBytes(field.key)
					e.encodeFloat64(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat64:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat64:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToFloat64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return &UnsupportedValueError{
							Value: reflect.ValueOf(v),
							Str:   strconv.FormatFloat(v, 'g', -1, 64),
						}
					}
					e.encodeBytes(field.key)
					e.encodeFloat64(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyString:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyString:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToString(ptr + field.offset)
				if v == "" {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeString(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyString:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyString:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToString(ptr + field.offset)
				if v == "" {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeString(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyBool:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBool:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToBool(ptr + field.offset)
				if !v {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeBool(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyBool:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBool:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToBool(ptr + field.offset)
				if !v {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					e.encodeBool(v)
					code = field.next
				}
				field.nextField.ptr = ptr
			}
		case opStructFieldPtrHeadOmitEmptyIndent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				p := ptr + field.offset
				if p == 0 || *(*uintptr)(unsafe.Pointer(p)) == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					code = field.next
					code.ptr = p
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyIntIndent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyIntIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToInt(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeInt(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt8Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt8Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToInt8(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeInt8(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt16Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt16Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToInt16(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeInt16(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt32Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt32Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToInt32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeInt32(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyInt64Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt64Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToInt64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeInt64(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyUintIndent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUintIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToUint(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeUint(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint8Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint8Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToUint8(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeUint8(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint16Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint16Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToUint16(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeUint16(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint32Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint32Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToUint32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeUint32(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyUint64Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint64Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToUint64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeUint64(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyFloat32Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToFloat32(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeFloat32(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyFloat64Indent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64Indent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToFloat64(ptr + field.offset)
				if v == 0 {
					code = field.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return &UnsupportedValueError{
							Value: reflect.ValueOf(v),
							Str:   strconv.FormatFloat(v, 'g', -1, 64),
						}
					}
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeFloat64(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyStringIndent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyStringIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToString(ptr + field.offset)
				if v == "" {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeString(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructFieldPtrHeadOmitEmptyBoolIndent:
			if code.ptr != 0 {
				code.ptr = e.ptrToPtr(code.ptr)
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBoolIndent:
			field := code.toStructFieldCode()
			ptr := field.ptr
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToBool(ptr + field.offset)
				if !v {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					e.encodeBool(v)
					code = field.next
				}
				field.nextField.ptr = field.ptr
			}
		case opStructField:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			e.encodeBytes(c.key)
			code = code.next
			code.ptr = c.ptr + c.offset
			c.nextField.ptr = c.ptr
		case opStructFieldInt:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt(e.ptrToInt(c.ptr + c.offset))
			code = code.next
		case opStructFieldInt8:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt8(e.ptrToInt8(c.ptr + c.offset))
			code = code.next
		case opStructFieldInt16:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt16(e.ptrToInt16(c.ptr + c.offset))
			code = code.next
		case opStructFieldInt32:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt32(e.ptrToInt32(c.ptr + c.offset))
			code = code.next
		case opStructFieldInt64:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeInt64(e.ptrToInt64(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint(e.ptrToUint(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint8:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint8(e.ptrToUint8(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint16:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint16(e.ptrToUint16(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint32:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint32(e.ptrToUint32(c.ptr + c.offset))
			code = code.next
		case opStructFieldUint64:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeUint64(e.ptrToUint64(c.ptr + c.offset))
			code = code.next
		case opStructFieldFloat32:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeFloat32(e.ptrToFloat32(c.ptr + c.offset))
			code = code.next
		case opStructFieldFloat64:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			v := e.ptrToFloat64(c.ptr + c.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return &UnsupportedValueError{
					Value: reflect.ValueOf(v),
					Str:   strconv.FormatFloat(v, 'g', -1, 64),
				}
			}
			e.encodeFloat64(v)
			code = code.next
		case opStructFieldString:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeString(e.ptrToString(c.ptr + c.offset))
			code = code.next
		case opStructFieldBool:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			c.nextField.ptr = c.ptr
			e.encodeBytes(c.key)
			e.encodeBool(e.ptrToBool(c.ptr + c.offset))
			code = code.next

		case opStructFieldIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			code = code.next
			code.ptr = c.ptr + c.offset
			c.nextField.ptr = c.ptr
		case opStructFieldIntIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeInt(e.ptrToInt(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldInt8Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeInt8(e.ptrToInt8(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldInt16Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeInt16(e.ptrToInt16(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldInt32Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeInt32(e.ptrToInt32(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldInt64Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeInt64(e.ptrToInt64(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldUintIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeUint(e.ptrToUint(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldUint8Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeUint8(e.ptrToUint8(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldUint16Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeUint16(e.ptrToUint16(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldUint32Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeUint32(e.ptrToUint32(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldUint64Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeUint64(e.ptrToUint64(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldFloat32Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeFloat32(e.ptrToFloat32(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldFloat64Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			v := e.ptrToFloat64(c.ptr + c.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return &UnsupportedValueError{
					Value: reflect.ValueOf(v),
					Str:   strconv.FormatFloat(v, 'g', -1, 64),
				}
			}
			e.encodeFloat64(v)
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldStringIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(e.ptrToString(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldBoolIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeBool(e.ptrToBool(c.ptr + c.offset))
			code = code.next
			c.nextField.ptr = c.ptr
		case opStructFieldOmitEmpty:
			c := code.toStructFieldCode()
			p := c.ptr + c.offset
			if p == 0 || *(*uintptr)(unsafe.Pointer(p)) == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				code = code.next
				code.ptr = p
			}
			c.nextField.ptr = c.ptr
		case opStructFieldOmitEmptyInt:
			c := code.toStructFieldCode()
			v := e.ptrToInt(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyInt8:
			c := code.toStructFieldCode()
			v := e.ptrToInt8(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt8(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyInt16:
			c := code.toStructFieldCode()
			v := e.ptrToInt16(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt16(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyInt32:
			c := code.toStructFieldCode()
			v := e.ptrToInt32(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt32(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyInt64:
			c := code.toStructFieldCode()
			v := e.ptrToInt64(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt64(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint:
			c := code.toStructFieldCode()
			v := e.ptrToUint(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint8:
			c := code.toStructFieldCode()
			v := e.ptrToUint8(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint8(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint16:
			c := code.toStructFieldCode()
			v := e.ptrToUint16(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint16(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint32:
			c := code.toStructFieldCode()
			v := e.ptrToUint32(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint32(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint64:
			c := code.toStructFieldCode()
			v := e.ptrToUint64(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint64(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyFloat32:
			c := code.toStructFieldCode()
			v := e.ptrToFloat32(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeFloat32(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyFloat64:
			c := code.toStructFieldCode()
			v := e.ptrToFloat64(c.ptr + c.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeFloat64(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyString:
			c := code.toStructFieldCode()
			v := e.ptrToString(c.ptr + c.offset)
			if v != "" {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeString(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyBool:
			c := code.toStructFieldCode()
			v := e.ptrToBool(c.ptr + c.offset)
			if v {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeBool(v)
			}
			code = code.next
			code.ptr = c.ptr

		case opStructFieldOmitEmptyIndent:
			c := code.toStructFieldCode()
			p := c.ptr + c.offset
			if p == 0 || *(*uintptr)(unsafe.Pointer(p)) == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				code = code.next
				code.ptr = p
			}
			c.nextField.ptr = c.ptr
		case opStructFieldOmitEmptyIntIndent:
			c := code.toStructFieldCode()
			v := e.ptrToInt(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyInt8Indent:
			c := code.toStructFieldCode()
			v := e.ptrToInt8(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt8(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyInt16Indent:
			c := code.toStructFieldCode()
			v := e.ptrToInt16(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt16(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyInt32Indent:
			c := code.toStructFieldCode()
			v := e.ptrToInt32(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt32(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyInt64Indent:
			c := code.toStructFieldCode()
			v := e.ptrToInt64(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt64(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUintIndent:
			c := code.toStructFieldCode()
			v := e.ptrToUint(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint8Indent:
			c := code.toStructFieldCode()
			v := e.ptrToUint8(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint8(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint16Indent:
			c := code.toStructFieldCode()
			v := e.ptrToUint16(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint16(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint32Indent:
			c := code.toStructFieldCode()
			v := e.ptrToUint32(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint32(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyUint64Indent:
			c := code.toStructFieldCode()
			v := e.ptrToUint64(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint64(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyFloat32Indent:
			c := code.toStructFieldCode()
			v := e.ptrToFloat32(c.ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeFloat32(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyFloat64Indent:
			c := code.toStructFieldCode()
			v := e.ptrToFloat64(c.ptr + c.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeFloat64(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyStringIndent:
			c := code.toStructFieldCode()
			v := e.ptrToString(c.ptr + c.offset)
			if v != "" {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeString(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructFieldOmitEmptyBoolIndent:
			c := code.toStructFieldCode()
			v := e.ptrToBool(c.ptr + c.offset)
			if v {
				if e.buf[len(e.buf)-2] != '{' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeBool(v)
			}
			code = code.next
			code.ptr = c.ptr
		case opStructEnd:
			e.encodeByte('}')
			code = code.next
		case opStructAnonymousEnd:
			code = code.next
		case opStructEndIndent:
			e.encodeByte('\n')
			e.encodeIndent(code.indent)
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

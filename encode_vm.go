package json

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"
)

func load(base uintptr, idx uintptr) uintptr {
	return *(*uintptr)(unsafe.Pointer(base + idx))
}

func store(base uintptr, idx uintptr, p uintptr) {
	*(*uintptr)(unsafe.Pointer(base + idx)) = p
}

func (e *Encoder) run(ctx *encodeRuntimeContext, seenPtr map[uintptr]struct{}, code *opcode) error {
	ctxptr := ctx.ptr()
	for {
		switch code.op {
		case opPtr:
			ptr := load(ctxptr, code.idx)
			code = code.next
			store(ctxptr, code.idx, e.ptrToPtr(ptr))
		case opInt:
			e.encodeInt(e.ptrToInt(load(ctxptr, code.idx)))
			code = code.next
		case opInt8:
			e.encodeInt8(e.ptrToInt8(load(ctxptr, code.idx)))
			code = code.next
		case opInt16:
			e.encodeInt16(e.ptrToInt16(load(ctxptr, code.idx)))
			code = code.next
		case opInt32:
			e.encodeInt32(e.ptrToInt32(load(ctxptr, code.idx)))
			code = code.next
		case opInt64:
			e.encodeInt64(e.ptrToInt64(load(ctxptr, code.idx)))
			code = code.next
		case opUint:
			e.encodeUint(e.ptrToUint(load(ctxptr, code.idx)))
			code = code.next
		case opUint8:
			e.encodeUint8(e.ptrToUint8(load(ctxptr, code.idx)))
			code = code.next
		case opUint16:
			e.encodeUint16(e.ptrToUint16(load(ctxptr, code.idx)))
			code = code.next
		case opUint32:
			e.encodeUint32(e.ptrToUint32(load(ctxptr, code.idx)))
			code = code.next
		case opUint64:
			e.encodeUint64(e.ptrToUint64(load(ctxptr, code.idx)))
			code = code.next
		case opFloat32:
			e.encodeFloat32(e.ptrToFloat32(load(ctxptr, code.idx)))
			code = code.next
		case opFloat64:
			v := e.ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return &UnsupportedValueError{
					Value: reflect.ValueOf(v),
					Str:   strconv.FormatFloat(v, 'g', -1, 64),
				}
			}
			e.encodeFloat64(v)
			code = code.next
		case opString:
			e.encodeString(e.ptrToString(load(ctxptr, code.idx)))
			code = code.next
		case opBool:
			e.encodeBool(e.ptrToBool(load(ctxptr, code.idx)))
			code = code.next
		case opBytes:
			ptr := load(ctxptr, code.idx)
			header := (*reflect.SliceHeader)(unsafe.Pointer(ptr))
			if ptr == 0 || header.Data == 0 {
				e.encodeNull()
			} else {
				b := e.ptrToBytes(ptr)
				encodedLen := base64.StdEncoding.EncodedLen(len(b))
				e.encodeByte('"')
				buf := make([]byte, encodedLen)
				base64.StdEncoding.Encode(buf, b)
				e.encodeBytes(buf)
				e.encodeByte('"')
			}
			code = code.next
		case opInterface:
			ifaceCode := code.toInterfaceCode()
			ptr := load(ctxptr, ifaceCode.idx)
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: ifaceCode.typ,
				ptr: unsafe.Pointer(ptr),
			}))
			if _, exists := seenPtr[ptr]; exists {
				return &UnsupportedValueError{
					Value: reflect.ValueOf(v),
					Str:   fmt.Sprintf("encountered a cycle via %s", code.typ),
				}
			}
			seenPtr[ptr] = struct{}{}
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
			var c *opcode
			if typ.Kind() == reflect.Map {
				code, err := e.compileMap(&encodeCompileContext{
					typ:        typ,
					root:       ifaceCode.root,
					withIndent: e.enabledIndent,
					indent:     ifaceCode.indent,
				}, false)
				if err != nil {
					return err
				}
				c = code
			} else {
				code, err := e.compile(&encodeCompileContext{
					typ:        typ,
					root:       ifaceCode.root,
					withIndent: e.enabledIndent,
					indent:     ifaceCode.indent,
				})
				if err != nil {
					return err
				}
				c = code
			}
			ctx := &encodeRuntimeContext{
				ptrs: make([]uintptr, c.length()),
			}
			ctx.init(uintptr(header.ptr))
			if err := e.run(ctx, seenPtr, c); err != nil {
				return err
			}
			code = ifaceCode.next
		case opMarshalJSON:
			ptr := load(ctxptr, code.idx)
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
			if len(b) == 0 {
				return errUnexpectedEndOfJSON(
					fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
					0,
				)
			}
			var buf bytes.Buffer
			if e.enabledIndent {
				if err := encodeWithIndent(
					&buf,
					b,
					string(e.prefix)+string(bytes.Repeat(e.indentStr, code.indent)),
					string(e.indentStr),
				); err != nil {
					return err
				}
			} else {
				if err := compact(&buf, b, true); err != nil {
					return err
				}
			}
			e.encodeBytes(buf.Bytes())
			code = code.next
		case opMarshalText:
			ptr := load(ctxptr, code.idx)
			isPtr := code.typ.Kind() == reflect.Ptr
			p := unsafe.Pointer(ptr)
			if p == nil {
				e.encodeNull()
			} else if isPtr && *(*unsafe.Pointer)(p) == nil {
				e.encodeBytes([]byte{'"', '"'})
			} else {
				if isPtr && code.typ.Elem().Implements(marshalTextType) {
					p = *(*unsafe.Pointer)(p)
				}
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: p,
				}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
			}
			code = code.next
		case opSliceHead:
			p := load(ctxptr, code.idx)
			headerCode := code.toSliceHeaderCode()
			header := (*reflect.SliceHeader)(unsafe.Pointer(p))
			if p == 0 || header.Data == 0 {
				e.encodeNull()
				code = headerCode.end.next
			} else {
				e.encodeByte('[')
				headerCode.elem.set(header)
				if header.Len > 0 {
					code = code.next
					store(ctxptr, code.idx, header.Data)
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
				store(ctxptr, code.idx, c.data+c.idx*c.size)
			} else {
				e.encodeByte(']')
				code = c.end.next
			}
		case opSliceHeadIndent:
			p := load(ctxptr, code.idx)
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
					store(ctxptr, code.idx, header.Data)
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{'[', ']'})
					code = headerCode.end.next
				}
			}
		case opRootSliceHeadIndent:
			p := load(ctxptr, code.idx)
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
					store(ctxptr, code.idx, header.Data)
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
				store(ctxptr, code.idx, c.data+c.idx*c.size)
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent)
				e.encodeByte(']')
				code = c.end.next
			}
		case opRootSliceElemIndent:
			c := code.toSliceElemCode()
			c.idx++
			if c.idx < c.len {
				e.encodeBytes([]byte{',', '\n'})
				e.encodeIndent(code.indent + 1)
				code = code.next
				store(ctxptr, code.idx, c.data+c.idx*c.size)
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent)
				e.encodeByte(']')
				code = c.end.next
			}
		case opArrayHead:
			p := load(ctxptr, code.idx)
			headerCode := code.toArrayHeaderCode()
			if p == 0 {
				e.encodeNull()
				code = headerCode.end.next
			} else {
				e.encodeByte('[')
				if headerCode.len > 0 {
					code = code.next
					store(ctxptr, code.idx, p)
					store(ctxptr, headerCode.elem.opcodeHeader.idx, p)
				} else {
					e.encodeByte(']')
					code = headerCode.end.next
				}
			}
		case opArrayElem:
			c := code.toArrayElemCode()
			c.idx++
			p := load(ctxptr, c.opcodeHeader.idx)
			if c.idx < c.len {
				e.encodeByte(',')
				code = code.next
				store(ctxptr, code.idx, p+c.idx*c.size)
			} else {
				e.encodeByte(']')
				code = c.end.next
			}
		case opArrayHeadIndent:
			p := load(ctxptr, code.idx)
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
					store(ctxptr, code.idx, p)
					store(ctxptr, headerCode.elem.opcodeHeader.idx, p)
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{']', '\n'})
					code = headerCode.end.next
				}
			}
		case opArrayElemIndent:
			c := code.toArrayElemCode()
			c.idx++
			p := load(ctxptr, c.opcodeHeader.idx)
			if c.idx < c.len {
				e.encodeBytes([]byte{',', '\n'})
				e.encodeIndent(code.indent + 1)
				code = code.next
				store(ctxptr, code.idx, p+c.idx*c.size)
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{']', '\n'})
				code = c.end.next
			}
		case opMapHead:
			ptr := load(ctxptr, code.idx)
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
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					e.encodeByte('}')
					code = mapHeadCode.end.next
				}
			}
		case opMapHeadLoad:
			ptr := load(ctxptr, code.idx)
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
					store(ctxptr, code.next.idx, uintptr(key))
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
				store(ctxptr, c.next.idx, uintptr(key))
				code = c.next
			} else {
				e.encodeByte('}')
				code = c.end.next
			}
		case opMapValue:
			e.encodeByte(':')
			c := code.toMapValueCode()
			value := mapitervalue(c.iter)
			store(ctxptr, c.next.idx, uintptr(value))
			mapiternext(c.iter)
			code = c.next
		case opMapHeadIndent:
			ptr := load(ctxptr, code.idx)
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
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
					e.encodeIndent(code.indent)
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{'{', '}'})
					code = mapHeadCode.end.next
				}
			}
		case opMapHeadLoadIndent:
			ptr := load(ctxptr, code.idx)
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
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
					e.encodeIndent(code.indent)
				} else {
					e.encodeIndent(code.indent)
					e.encodeBytes([]byte{'{', '}'})
					code = mapHeadCode.end.next
				}
			}
		case opRootMapHeadIndent:
			ptr := load(ctxptr, code.idx)
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
					store(ctxptr, code.next.idx, uintptr(key))
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
				store(ctxptr, c.next.idx, uintptr(key))
				code = c.next
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent - 1)
				e.encodeByte('}')
				code = c.end.next
			}
		case opRootMapKeyIndent:
			c := code.toMapKeyCode()
			c.idx++
			if c.idx < c.len {
				e.encodeBytes([]byte{',', '\n'})
				e.encodeIndent(code.indent)
				key := mapiterkey(c.iter)
				store(ctxptr, c.next.idx, uintptr(key))
				code = c.next
			} else {
				e.encodeByte('\n')
				e.encodeIndent(code.indent - 1)
				e.encodeByte('}')
				code = c.end.next
			}
		case opMapValueIndent:
			e.encodeBytes([]byte{':', ' '})
			c := code.toMapValueCode()
			value := mapitervalue(c.iter)
			store(ctxptr, c.next.idx, uintptr(value))
			mapiternext(c.iter)
			code = c.next
		case opStructFieldRecursive:
			recursive := code.toRecursiveCode()
			ptr := load(ctxptr, code.idx)
			if recursive.seenPtr != 0 && recursive.seenPtr == ptr {
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: unsafe.Pointer(ptr),
				}))
				return &UnsupportedValueError{
					Value: reflect.ValueOf(v),
					Str:   fmt.Sprintf("encountered a cycle via %s", code.typ),
				}
			}
			recursive.seenPtr = ptr
			recursiveCode := newRecursiveCode(recursive)
			ctx := &encodeRuntimeContext{
				ptrs: make([]uintptr, recursiveCode.length()),
			}
			ctx.init(ptr)
			if err := e.run(ctx, seenPtr, recursiveCode); err != nil {
				return err
			}
			code = recursive.next
		case opStructFieldPtrHead:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHead:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHead {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end.next
			} else {
				e.encodeByte('{')
				if !field.anonymousKey {
					e.encodeBytes(field.key)
				}
				code = field.next
				store(ctxptr, code.idx, ptr+field.offset)
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldAnonymousHead:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				code = field.next
				store(ctxptr, code.idx, ptr)
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadInt:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt(e.ptrToInt(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt(e.ptrToInt(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadInt8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt8 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt8(e.ptrToInt8(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt8(e.ptrToInt8(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadInt16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt16 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt16(e.ptrToInt16(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt16(e.ptrToInt16(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadInt32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt32 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt32(e.ptrToInt32(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt32(e.ptrToInt32(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadInt64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt64 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeInt64(e.ptrToInt64(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadInt64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeInt64(e.ptrToInt64(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint(e.ptrToUint(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint(e.ptrToUint(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint8 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint8(e.ptrToUint8(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, field.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint8(e.ptrToUint8(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint16 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint16(e.ptrToUint16(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint16(e.ptrToUint16(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint32 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint32(e.ptrToUint32(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint32(e.ptrToUint32(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint64 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeUint64(e.ptrToUint64(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadUint64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeUint64(e.ptrToUint64(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadFloat32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadFloat32 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeFloat32(e.ptrToFloat32(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadFloat32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeFloat32(e.ptrToFloat32(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadFloat64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadFloat64 {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				v := e.ptrToFloat64(ptr + field.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeFloat64(v)
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadFloat64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				v := e.ptrToFloat64(ptr + field.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				e.encodeBytes(field.key)
				e.encodeFloat64(v)
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadString:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadString:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadString {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(e.ptrToString(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadString:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadString:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeString(e.ptrToString(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadBool:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadBool:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadBool {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeBool(e.ptrToBool(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadBool:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadBool:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				e.encodeBool(e.ptrToBool(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadBytes:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadBytes:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadBytes {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				s := base64.StdEncoding.EncodeToString(e.ptrToBytes(ptr + field.offset))
				e.encodeByte('"')
				e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
				e.encodeByte('"')
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadBytes:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadBytes:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				s := base64.StdEncoding.EncodeToString(e.ptrToBytes(ptr + field.offset))
				e.encodeByte('"')
				e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
				e.encodeByte('"')
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadArray:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadArray:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx) + c.offset
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadArray {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'[', ']'})
				}
				code = c.end
			} else {
				e.encodeByte('{')
				if !c.anonymousKey {
					e.encodeBytes(c.key)
				}
				code = c.next
				store(ctxptr, code.idx, ptr)
				store(ctxptr, c.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadArray:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadArray:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx) + c.offset
			if ptr == 0 {
				code = c.end
			} else {
				e.encodeBytes(c.key)
				store(ctxptr, code.idx, ptr)
				store(ctxptr, c.nextField.idx, ptr)
				code = c.next
			}
		case opStructFieldPtrHeadSlice:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadSlice:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			p := ptr + c.offset
			if p == 0 {
				if code.op == opStructFieldPtrHeadSlice {
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'[', ']'})
				}
				code = c.end
			} else {
				e.encodeByte('{')
				if !c.anonymousKey {
					e.encodeBytes(c.key)
				}
				code = c.next
				store(ctxptr, code.idx, p)
				store(ctxptr, c.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadSlice:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadSlice:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			p := ptr + c.offset
			if p == 0 {
				code = c.end
			} else {
				e.encodeBytes(c.key)
				store(ctxptr, code.idx, p)
				store(ctxptr, c.nextField.idx, ptr)
				code = c.next
			}
		case opStructFieldPtrHeadMarshalJSON:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadMarshalJSON:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: unsafe.Pointer(ptr + field.offset),
				}))
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					e.encodeNull()
					code = field.end
					break
				}
				b, err := rv.Interface().(Marshaler).MarshalJSON()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				if len(b) == 0 {
					return errUnexpectedEndOfJSON(
						fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
						0,
					)
				}
				var buf bytes.Buffer
				if err := compact(&buf, b, true); err != nil {
					return err
				}
				e.encodeBytes(buf.Bytes())
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadMarshalJSON:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadMarshalJSON:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: unsafe.Pointer(ptr + field.offset),
				}))
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					e.encodeNull()
					code = field.end
					break
				}
				b, err := rv.Interface().(Marshaler).MarshalJSON()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				if len(b) == 0 {
					return errUnexpectedEndOfJSON(
						fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
						0,
					)
				}
				var buf bytes.Buffer
				if err := compact(&buf, b, true); err != nil {
					return err
				}
				e.encodeBytes(buf.Bytes())
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadMarshalText:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadMarshalText:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: unsafe.Pointer(ptr + field.offset),
				}))
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					e.encodeNull()
					code = field.end
					break
				}
				bytes, err := rv.Interface().(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrAnonymousHeadMarshalText:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadMarshalText:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end
			} else {
				e.encodeBytes(field.key)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: unsafe.Pointer(ptr + field.offset),
				}))
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					e.encodeNull()
					code = field.end
					break
				}
				bytes, err := rv.Interface().(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else if field.next == field.end {
				// not exists fields
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '}'})
				code = field.next
				store(ctxptr, code.idx, ptr)
				store(ctxptr, field.nextField.idx, ptr)
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				code = field.next
				store(ctxptr, code.idx, ptr)
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadIntIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadIntIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadIntIndent {
					e.encodeIndent(code.indent)
					e.encodeNull()
				} else {
					e.encodeBytes([]byte{'{', '}'})
				}
				code = field.end
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeInt(e.ptrToInt(ptr + field.offset))
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadInt8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt8Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadInt16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt16Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadInt32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadInt64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUintIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUintIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint8Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint16Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadUint64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadFloat32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadFloat64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadStringIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadStringIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadBoolIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadBoolIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadBytesIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadBytesIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				s := base64.StdEncoding.EncodeToString(e.ptrToBytes(ptr))
				e.encodeByte('"')
				e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
				e.encodeByte('"')
				store(ctxptr, field.nextField.idx, ptr)
				code = field.next
			}
		case opStructFieldPtrHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmpty:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
					store(ctxptr, code.idx, p)
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmpty:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				p := ptr + field.offset
				if p == 0 || *(*uintptr)(unsafe.Pointer(p)) == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					code = field.next
					store(ctxptr, code.idx, p)
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyString:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyString:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBool:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBool:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBytes:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToBytes(ptr + field.offset)
				if len(v) == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					s := base64.StdEncoding.EncodeToString(v)
					e.encodeByte('"')
					e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
					e.encodeByte('"')
					code = field.next
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBytes:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToBytes(ptr + field.offset)
				if len(v) == 0 {
					code = field.nextField
				} else {
					e.encodeBytes(field.key)
					s := base64.StdEncoding.EncodeToString(v)
					e.encodeByte('"')
					e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
					e.encodeByte('"')
					code = field.next
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalJSON:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				p := unsafe.Pointer(ptr + field.offset)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = field.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					b, err := v.(Marshaler).MarshalJSON()
					if err != nil {
						return &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					if len(b) == 0 {
						if isPtr {
							return errUnexpectedEndOfJSON(
								fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
								0,
							)
						}
						code = field.nextField
					} else {
						var buf bytes.Buffer
						if err := compact(&buf, b, true); err != nil {
							return err
						}
						e.encodeBytes(field.key)
						e.encodeBytes(buf.Bytes())
						code = field.next
					}
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMarshalJSON:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				p := unsafe.Pointer(ptr + field.offset)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = field.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					b, err := v.(Marshaler).MarshalJSON()
					if err != nil {
						return &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					if len(b) == 0 {
						if isPtr {
							return errUnexpectedEndOfJSON(
								fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
								0,
							)
						}
						code = field.nextField
					} else {
						var buf bytes.Buffer
						if err := compact(&buf, b, true); err != nil {
							return err
						}
						e.encodeBytes(field.key)
						e.encodeBytes(buf.Bytes())
						code = field.next
					}
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalText:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				p := unsafe.Pointer(ptr + field.offset)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = field.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bytes, err := v.(encoding.TextMarshaler).MarshalText()
					if err != nil {
						return &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					e.encodeBytes(field.key)
					e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
					code = field.next
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMarshalText:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				p := unsafe.Pointer(ptr + field.offset)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = field.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bytes, err := v.(encoding.TextMarshaler).MarshalText()
					if err != nil {
						return &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					e.encodeBytes(field.key)
					e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
					code = field.next
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
					store(ctxptr, code.idx, p)
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyIntIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt8Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt16Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUintIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint8Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint16Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyStringIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBoolIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
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
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadOmitEmptyBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBytesIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeIndent(code.indent)
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToBytes(ptr + field.offset)
				if len(v) == 0 {
					code = field.nextField
				} else {
					e.encodeIndent(code.indent + 1)
					e.encodeBytes(field.key)
					e.encodeByte(' ')
					s := base64.StdEncoding.EncodeToString(v)
					e.encodeByte('"')
					e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
					e.encodeByte('"')
					code = field.next
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTag:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				p := ptr + field.offset
				e.encodeBytes(field.key)
				code = field.next
				store(ctxptr, code.idx, p)
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTag:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				code = field.next
				store(ctxptr, code.idx, ptr+field.offset)
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt8(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt8(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt16(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt16(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt64(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToInt64(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint8(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint8:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint8(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint16(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint16:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint16(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint64(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToUint64(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToFloat32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat32:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToFloat32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				v := e.ptrToFloat64(ptr + field.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(v))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat64:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				v := e.ptrToFloat64(ptr + field.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(v))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagString:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(strconv.Quote(e.ptrToString(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagString:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(strconv.Quote(e.ptrToString(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBool:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToBool(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagBool:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				e.encodeString(fmt.Sprint(e.ptrToBool(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBytes:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				e.encodeBytes(field.key)
				s := base64.StdEncoding.EncodeToString(
					e.ptrToBytes(ptr + field.offset),
				)
				e.encodeByte('"')
				e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
				e.encodeByte('"')
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagBytes:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				e.encodeBytes(field.key)
				s := base64.StdEncoding.EncodeToString(
					e.ptrToBytes(ptr + field.offset),
				)
				e.encodeByte('"')
				e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
				e.encodeByte('"')
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagMarshalJSON:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				p := unsafe.Pointer(ptr + field.offset)
				isPtr := code.typ.Kind() == reflect.Ptr
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				b, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				if len(b) == 0 {
					if isPtr {
						return errUnexpectedEndOfJSON(
							fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
							0,
						)
					}
					e.encodeBytes(field.key)
					e.encodeBytes([]byte{'"', '"'})
					code = field.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, b, true); err != nil {
						return err
					}
					e.encodeString(buf.String())
					code = field.next
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagMarshalJSON:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				p := unsafe.Pointer(ptr + field.offset)
				isPtr := code.typ.Kind() == reflect.Ptr
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				b, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				if len(b) == 0 {
					if isPtr {
						return errUnexpectedEndOfJSON(
							fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
							0,
						)
					}
					e.encodeBytes(field.key)
					e.encodeBytes([]byte{'"', '"'})
					code = field.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, b, true); err != nil {
						return err
					}
					e.encodeBytes(field.key)
					e.encodeString(buf.String())
					code = field.next
				}
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagMarshalText:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeByte('{')
				p := unsafe.Pointer(ptr + field.offset)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				e.encodeBytes(field.key)
				e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagMarshalText:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = field.end.next
			} else {
				p := unsafe.Pointer(ptr + field.offset)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				e.encodeBytes(field.key)
				e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				p := ptr + field.offset
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				code = field.next
				store(ctxptr, code.idx, p)
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagIntIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToInt(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt8Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToInt8(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt16Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToInt16(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToInt32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToInt64(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUintIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToUint(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint8Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToUint8(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint16Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToUint16(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToUint32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToUint64(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat32Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToFloat32(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat64Indent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				v := e.ptrToFloat64(ptr + field.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(v))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagStringIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(strconv.Quote(e.ptrToString(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBoolIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				e.encodeString(fmt.Sprint(e.ptrToBool(ptr + field.offset)))
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructFieldPtrHeadStringTagBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBytesIndent:
			field := code.toStructFieldCode()
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				e.encodeIndent(code.indent)
				e.encodeNull()
				code = field.end.next
			} else {
				e.encodeBytes([]byte{'{', '\n'})
				e.encodeIndent(code.indent + 1)
				e.encodeBytes(field.key)
				e.encodeByte(' ')
				s := base64.StdEncoding.EncodeToString(
					e.ptrToBytes(ptr + field.offset),
				)
				e.encodeByte('"')
				e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
				e.encodeByte('"')
				code = field.next
				store(ctxptr, field.nextField.idx, ptr)
			}
		case opStructField:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			if !c.anonymousKey {
				e.encodeBytes(c.key)
			}
			code = code.next
			ptr := load(ctxptr, c.idx)
			store(ctxptr, code.idx, ptr+c.offset)
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldInt:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeInt(e.ptrToInt(ptr + c.offset))
			code = code.next
		case opStructFieldInt8:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeInt8(e.ptrToInt8(ptr + c.offset))
			code = code.next
		case opStructFieldInt16:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeInt16(e.ptrToInt16(ptr + c.offset))
			code = code.next
		case opStructFieldInt32:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeInt32(e.ptrToInt32(ptr + c.offset))
			code = code.next
		case opStructFieldInt64:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeInt64(e.ptrToInt64(ptr + c.offset))
			code = code.next
		case opStructFieldUint:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeUint(e.ptrToUint(ptr + c.offset))
			code = code.next
		case opStructFieldUint8:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeUint8(e.ptrToUint8(ptr + c.offset))
			code = code.next
		case opStructFieldUint16:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeUint16(e.ptrToUint16(ptr + c.offset))
			code = code.next
		case opStructFieldUint32:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeUint32(e.ptrToUint32(ptr + c.offset))
			code = code.next
		case opStructFieldUint64:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeUint64(e.ptrToUint64(ptr + c.offset))
			code = code.next
		case opStructFieldFloat32:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeFloat32(e.ptrToFloat32(ptr + c.offset))
			code = code.next
		case opStructFieldFloat64:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			v := e.ptrToFloat64(ptr + c.offset)
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
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeString(e.ptrToString(ptr + c.offset))
			code = code.next
		case opStructFieldBool:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			e.encodeBool(e.ptrToBool(ptr + c.offset))
			code = code.next
		case opStructFieldBytes:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			s := base64.StdEncoding.EncodeToString(e.ptrToBytes(ptr + c.offset))
			e.encodeByte('"')
			e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
			e.encodeByte('"')
			code = code.next
		case opStructFieldMarshalJSON:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
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
		case opStructFieldMarshalText:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			e.encodeBytes(c.key)
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
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
		case opStructFieldArray:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			code = code.next
			store(ctxptr, code.idx, ptr+c.offset)
			e.encodeBytes(c.key)
		case opStructFieldSlice:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			store(ctxptr, c.nextField.idx, ptr)
			code = code.next
			store(ctxptr, code.idx, ptr+c.offset)
			e.encodeBytes(c.key)
		case opStructFieldMap:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			e.encodeBytes(c.key)
			code = code.next
			store(ctxptr, code.idx, ptr+c.offset)
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldMapLoad:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			e.encodeBytes(c.key)
			code = code.next
			store(ctxptr, code.idx, ptr+c.offset)
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldStruct:
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			e.encodeBytes(c.key)
			code = code.next
			store(ctxptr, code.idx, ptr+c.offset)
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			code = code.next
			ptr := load(ctxptr, c.idx)
			store(ctxptr, code.idx, ptr+c.offset)
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldIntIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeInt(e.ptrToInt(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldInt8Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeInt8(e.ptrToInt8(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldInt16Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeInt16(e.ptrToInt16(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldInt32Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeInt32(e.ptrToInt32(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldInt64Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeInt64(e.ptrToInt64(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldUintIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeUint(e.ptrToUint(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldUint8Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeUint8(e.ptrToUint8(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldUint16Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeUint16(e.ptrToUint16(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldUint32Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeUint32(e.ptrToUint32(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldUint64Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeUint64(e.ptrToUint64(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldFloat32Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeFloat32(e.ptrToFloat32(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldFloat64Indent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			v := e.ptrToFloat64(ptr + c.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return &UnsupportedValueError{
					Value: reflect.ValueOf(v),
					Str:   strconv.FormatFloat(v, 'g', -1, 64),
				}
			}
			e.encodeFloat64(v)
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldStringIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeString(e.ptrToString(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldBoolIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			e.encodeBool(e.ptrToBool(ptr + c.offset))
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldBytesIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			s := base64.StdEncoding.EncodeToString(e.ptrToBytes(ptr + c.offset))
			e.encodeByte('"')
			e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
			e.encodeByte('"')
			code = code.next
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldMarshalJSONIndent:
			c := code.toStructFieldCode()
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeByte(',')
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
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
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldArrayIndent:
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			c := code.toStructFieldCode()
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			header := (*reflect.SliceHeader)(unsafe.Pointer(p))
			if p == 0 || header.Data == 0 {
				e.encodeNull()
				code = c.nextField
			} else {
				code = code.next
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldSliceIndent:
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			c := code.toStructFieldCode()
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			header := (*reflect.SliceHeader)(unsafe.Pointer(p))
			if p == 0 || header.Data == 0 {
				e.encodeNull()
				code = c.nextField
			} else {
				code = code.next
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldMapIndent:
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			c := code.toStructFieldCode()
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 {
				e.encodeNull()
				code = c.nextField
			} else {
				mlen := maplen(unsafe.Pointer(p))
				if mlen == 0 {
					e.encodeBytes([]byte{'{', '}'})
					mapCode := code.next
					mapHeadCode := mapCode.toMapHeadCode()
					code = mapHeadCode.end.next
				} else {
					code = code.next
				}
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldMapLoadIndent:
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			c := code.toStructFieldCode()
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 {
				e.encodeNull()
				code = c.nextField
			} else {
				p = uintptr(*(*unsafe.Pointer)(unsafe.Pointer(p)))
				mlen := maplen(unsafe.Pointer(p))
				if mlen == 0 {
					e.encodeBytes([]byte{'{', '}'})
					code = c.nextField
				} else {
					code = code.next
				}
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldStructIndent:
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			if p == 0 {
				e.encodeBytes([]byte{'{', '}'})
				code = c.nextField
			} else {
				headCode := c.next.toStructFieldCode()
				if headCode.next == headCode.end {
					// not exists fields
					e.encodeBytes([]byte{'{', '}'})
					code = c.nextField
				} else {
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmpty:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 || *(*uintptr)(unsafe.Pointer(p)) == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				code = code.next
				store(ctxptr, code.idx, p)
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptyInt:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyInt8:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt8(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt8(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyInt16:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt16(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt16(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyInt32:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt32(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt32(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyInt64:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt64(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeInt64(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint8:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint8(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint8(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint16:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint16(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint16(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint32:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint32(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint32(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint64:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint64(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeUint64(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyFloat32:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToFloat32(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeFloat32(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyFloat64:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToFloat64(ptr + c.offset)
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
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyString:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToString(ptr + c.offset)
			if v != "" {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeString(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyBool:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToBool(ptr + c.offset)
			if v {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				e.encodeBool(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyBytes:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToBytes(ptr + c.offset)
			if len(v) > 0 {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				e.encodeBytes(c.key)
				s := base64.StdEncoding.EncodeToString(v)
				e.encodeByte('"')
				e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
				e.encodeByte('"')
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyMarshalJSON:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
			}))
			if v != nil {
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
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyMarshalText:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
			}))
			if v != nil {
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: unsafe.Pointer(p),
				}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				e.encodeString(*(*string)(unsafe.Pointer(&bytes)))
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyArray:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			header := (*reflect.SliceHeader)(unsafe.Pointer(p))
			if p == 0 || header.Data == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				code = code.next
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptySlice:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			header := (*reflect.SliceHeader)(unsafe.Pointer(p))
			if p == 0 || header.Data == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-1] != '{' {
					e.encodeByte(',')
				}
				code = code.next
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptyMap:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 {
				code = c.nextField
			} else {
				mlen := maplen(unsafe.Pointer(p))
				if mlen == 0 {
					code = c.nextField
				} else {
					if e.buf[len(e.buf)-1] != '{' {
						e.encodeByte(',')
					}
					code = code.next
				}
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptyMapLoad:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 {
				code = c.nextField
			} else {
				p = uintptr(*(*unsafe.Pointer)(unsafe.Pointer(p)))
				mlen := maplen(unsafe.Pointer(p))
				if mlen == 0 {
					code = c.nextField
				} else {
					if e.buf[len(e.buf)-1] != '{' {
						e.encodeByte(',')
					}
					code = code.next
				}
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptyIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 || *(*uintptr)(unsafe.Pointer(p)) == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptyIntIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyInt8Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt8(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt8(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyInt16Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt16(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt16(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyInt32Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt32(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt32(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyInt64Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToInt64(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeInt64(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUintIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint8Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint8(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint8(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint16Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint16(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint16(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint32Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint32(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint32(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyUint64Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToUint64(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeUint64(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyFloat32Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToFloat32(ptr + c.offset)
			if v != 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeFloat32(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyFloat64Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToFloat64(ptr + c.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return &UnsupportedValueError{
						Value: reflect.ValueOf(v),
						Str:   strconv.FormatFloat(v, 'g', -1, 64),
					}
				}
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeFloat64(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyStringIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToString(ptr + c.offset)
			if v != "" {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeString(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyBoolIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToBool(ptr + c.offset)
			if v {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				e.encodeBool(v)
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyBytesIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToBytes(ptr + c.offset)
			if len(v) > 0 {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				s := base64.StdEncoding.EncodeToString(v)
				e.encodeByte('"')
				e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
				e.encodeByte('"')
			}
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmptyArrayIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			header := (*reflect.SliceHeader)(unsafe.Pointer(p))
			if p == 0 || header.Data == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				code = code.next
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptySliceIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			header := (*reflect.SliceHeader)(unsafe.Pointer(p))
			if p == 0 || header.Data == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				code = code.next
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptyMapIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 {
				code = c.nextField
			} else {
				mlen := maplen(unsafe.Pointer(p))
				if mlen == 0 {
					code = c.nextField
				} else {
					if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
						e.encodeBytes([]byte{',', '\n'})
					}
					e.encodeIndent(c.indent)
					e.encodeBytes(c.key)
					e.encodeByte(' ')
					code = code.next
				}
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptyMapLoadIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 {
				code = c.nextField
			} else {
				p = uintptr(*(*unsafe.Pointer)(unsafe.Pointer(p)))
				mlen := maplen(unsafe.Pointer(p))
				if mlen == 0 {
					code = c.nextField
				} else {
					if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
						e.encodeBytes([]byte{',', '\n'})
					}
					e.encodeIndent(c.indent)
					e.encodeBytes(c.key)
					e.encodeByte(' ')
					code = code.next
				}
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldOmitEmptyStructIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if p == 0 {
				code = c.nextField
			} else {
				if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
					e.encodeBytes([]byte{',', '\n'})
				}
				e.encodeIndent(c.indent)
				e.encodeBytes(c.key)
				e.encodeByte(' ')
				headCode := c.next.toStructFieldCode()
				if headCode.next == headCode.end {
					// not exists fields
					e.encodeBytes([]byte{'{', '}'})
					code = c.nextField
				} else {
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldStringTag:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			code = code.next
			store(ctxptr, code.idx, p)
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldStringTagInt:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToInt(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagInt8:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToInt8(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagInt16:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToInt16(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagInt32:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToInt32(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagInt64:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToInt64(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToUint(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint8:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToUint8(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint16:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToUint16(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint32:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToUint32(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint64:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToUint64(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagFloat32:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToFloat32(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagFloat64:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToFloat64(ptr + c.offset)
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
			e.encodeString(fmt.Sprint(v))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagString:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(strconv.Quote(e.ptrToString(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagBool:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			e.encodeString(fmt.Sprint(e.ptrToBool(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagBytes:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToBytes(ptr + c.offset)
			if e.buf[len(e.buf)-1] != '{' {
				e.encodeByte(',')
			}
			e.encodeBytes(c.key)
			s := base64.StdEncoding.EncodeToString(v)
			e.encodeByte('"')
			e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
			e.encodeByte('"')
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagMarshalJSON:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
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
			e.encodeString(buf.String())
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagMarshalText:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
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
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			p := ptr + c.offset
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			code = code.next
			store(ctxptr, code.idx, p)
			store(ctxptr, c.nextField.idx, ptr)
		case opStructFieldStringTagIntIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToInt(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagInt8Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToInt8(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagInt16Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToInt16(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagInt32Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToInt32(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagInt64Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToInt64(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUintIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToUint(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint8Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToUint8(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint16Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToUint16(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint32Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToUint32(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagUint64Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToUint64(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagFloat32Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToFloat32(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagFloat64Indent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			v := e.ptrToFloat64(ptr + c.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return &UnsupportedValueError{
					Value: reflect.ValueOf(v),
					Str:   strconv.FormatFloat(v, 'g', -1, 64),
				}
			}
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(v))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagStringIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			var b bytes.Buffer
			enc := NewEncoder(&b)
			enc.encodeString(e.ptrToString(ptr + c.offset))
			e.encodeString(string(enc.buf))
			enc.release()
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagBoolIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			e.encodeString(fmt.Sprint(e.ptrToBool(ptr + c.offset)))
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagBytesIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			s := base64.StdEncoding.EncodeToString(
				e.ptrToBytes(ptr + c.offset),
			)
			e.encodeByte('"')
			e.encodeBytes(*(*[]byte)(unsafe.Pointer(&s)))
			e.encodeByte('"')
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagMarshalJSONIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
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
			e.encodeString(buf.String())
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldStringTagMarshalTextIndent:
			c := code.toStructFieldCode()
			ptr := load(ctxptr, c.idx)
			if e.buf[len(e.buf)-2] != '{' || e.buf[len(e.buf)-1] == '}' {
				e.encodeBytes([]byte{',', '\n'})
			}
			e.encodeIndent(c.indent)
			e.encodeBytes(c.key)
			e.encodeByte(' ')
			p := ptr + c.offset
			v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
				typ: code.typ,
				ptr: unsafe.Pointer(p),
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
			store(ctxptr, c.nextField.idx, ptr)
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

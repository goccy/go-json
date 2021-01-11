package json

import (
	"bytes"
	"encoding"
	"fmt"
	"math"
	"reflect"
	"sort"
	"unsafe"
)

func (e *Encoder) runEscapedIndent(ctx *encodeRuntimeContext, b []byte, code *opcode) ([]byte, error) {
	recursiveLevel := 0
	var seenPtr map[uintptr]struct{}
	ptrOffset := uintptr(0)
	ctxptr := ctx.ptr()

	for {
		switch code.op {
		default:
			return nil, fmt.Errorf("failed to handle opcode. doesn't implement %s", code.op)
		case opPtr:
			ptr := load(ctxptr, code.idx)
			code = code.next
			store(ctxptr, code.idx, e.ptrToPtr(ptr))
		case opInt:
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt8:
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt16:
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt32:
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt64:
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opUint:
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint8:
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint16:
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint32:
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint64:
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opIntString:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt8String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt16String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt32String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt64String:
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUintString:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint8String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint16String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint32String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint64String:
			b = append(b, '"')
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opFloat32:
			b = encodeFloat32(b, e.ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opFloat64:
			v := e.ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opEscapedString:
			b = encodeEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBool:
			b = encodeBool(b, e.ptrToBool(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBytes:
			ptr := load(ctxptr, code.idx)
			slice := e.ptrToSlice(ptr)
			if ptr == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, e.ptrToBytes(ptr))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opInterface:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.next
				break
			}
			if seenPtr == nil {
				seenPtr = map[uintptr]struct{}{}
			}
			if _, exists := seenPtr[ptr]; exists {
				return nil, errUnsupportedValue(code, ptr)
			}
			seenPtr[ptr] = struct{}{}
			v := e.ptrToInterface(code, ptr)
			ctx.keepRefs = append(ctx.keepRefs, unsafe.Pointer(&v))
			rv := reflect.ValueOf(v)
			if rv.IsNil() {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.next
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
					typ:                      typ,
					root:                     code.root,
					indent:                   code.indent,
					structTypeToCompiledCode: map[uintptr]*compiledCode{},
				}, false)
				if err != nil {
					return nil, err
				}
				c = code
			} else {
				code, err := e.compile(&encodeCompileContext{
					typ:                      typ,
					root:                     code.root,
					indent:                   code.indent,
					structTypeToCompiledCode: map[uintptr]*compiledCode{},
				})
				if err != nil {
					return nil, err
				}
				c = code
			}
			c = toEscaped(c)
			beforeLastCode := c.beforeLastCode()
			lastCode := beforeLastCode.next
			lastCode.idx = beforeLastCode.idx + uintptrSize
			totalLength := uintptr(code.totalLength())
			nextTotalLength := uintptr(c.totalLength())
			curlen := uintptr(len(ctx.ptrs))
			offsetNum := ptrOffset / uintptrSize
			oldOffset := ptrOffset
			ptrOffset += totalLength * uintptrSize

			newLen := offsetNum + totalLength + nextTotalLength
			if curlen < newLen {
				ctx.ptrs = append(ctx.ptrs, make([]uintptr, newLen-curlen)...)
			}
			ctxptr = ctx.ptr() + ptrOffset // assign new ctxptr

			store(ctxptr, 0, uintptr(header.ptr))
			store(ctxptr, lastCode.idx, oldOffset)

			// link lastCode ( opInterfaceEnd ) => code.next
			lastCode.op = opInterfaceEnd
			lastCode.next = code.next

			code = c
			recursiveLevel++
		case opInterfaceEnd:
			recursiveLevel--
			// restore ctxptr
			offset := load(ctxptr, code.idx)
			ctxptr = ctx.ptr() + offset
			ptrOffset = offset
			code = code.next
		case opMarshalJSON:
			ptr := load(ctxptr, code.idx)
			v := e.ptrToInterface(code, ptr)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			if len(bb) == 0 {
				return nil, errUnexpectedEndOfJSON(
					fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
					0,
				)
			}
			var buf bytes.Buffer
			if err := encodeWithIndent(
				&buf,
				bb,
				string(e.prefix)+string(bytes.Repeat(e.indentStr, code.indent)),
				string(e.indentStr),
			); err != nil {
				return nil, err
			}
			b = append(b, buf.Bytes()...)
			b = encodeIndentComma(b)
			code = code.next
		case opMarshalText:
			ptr := load(ctxptr, code.idx)
			isPtr := code.typ.Kind() == reflect.Ptr
			p := e.ptrToUnsafePtr(ptr)
			if p == nil {
				b = encodeNull(b)
				b = encodeIndentComma(b)
			} else if isPtr && *(*unsafe.Pointer)(p) == nil {
				b = append(b, '"', '"', ',', '\n')
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
					return nil, errMarshaler(code, err)
				}
				b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opSliceHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				slice := e.ptrToSlice(p)
				store(ctxptr, code.elemIdx, 0)
				store(ctxptr, code.length, uintptr(slice.len))
				store(ctxptr, code.idx, uintptr(slice.data))
				if slice.len > 0 {
					b = append(b, '[', '\n')
					b = e.encodeIndent(b, code.indent+1)
					code = code.next
					store(ctxptr, code.idx, uintptr(slice.data))
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '[', ']', '\n')
					code = code.end.next
				}
			}
		case opRootSliceHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				slice := e.ptrToSlice(p)
				store(ctxptr, code.elemIdx, 0)
				store(ctxptr, code.length, uintptr(slice.len))
				store(ctxptr, code.idx, uintptr(slice.data))
				if slice.len > 0 {
					b = append(b, '[', '\n')
					b = e.encodeIndent(b, code.indent+1)
					code = code.next
					store(ctxptr, code.idx, uintptr(slice.data))
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '[', ']', ',', '\n')
					code = code.end.next
				}
			}
		case opSliceElem:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if idx < length {
				b = e.encodeIndent(b, code.indent+1)
				store(ctxptr, code.elemIdx, idx)
				data := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, data+idx*size)
			} else {
				b = b[:len(b)-2]
				b = append(b, '\n')
				b = e.encodeIndent(b, code.indent)
				b = append(b, ']', ',', '\n')
				code = code.end.next
			}
		case opRootSliceElem:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if idx < length {
				b = e.encodeIndent(b, code.indent+1)
				store(ctxptr, code.elemIdx, idx)
				code = code.next
				data := load(ctxptr, code.headIdx)
				store(ctxptr, code.idx, data+idx*code.size)
			} else {
				b = append(b, '\n')
				b = e.encodeIndent(b, code.indent)
				b = append(b, ']')
				code = code.end.next
			}
		case opArrayHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				if code.length > 0 {
					b = append(b, '[', '\n')
					b = e.encodeIndent(b, code.indent+1)
					store(ctxptr, code.elemIdx, 0)
					code = code.next
					store(ctxptr, code.idx, p)
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '[', ']', ',', '\n')
					code = code.end.next
				}
			}
		case opArrayElem:
			idx := load(ctxptr, code.elemIdx)
			idx++
			if idx < code.length {
				b = e.encodeIndent(b, code.indent+1)
				store(ctxptr, code.elemIdx, idx)
				p := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, p+idx*size)
			} else {
				b = b[:len(b)-2]
				b = append(b, '\n')
				b = e.encodeIndent(b, code.indent)
				b = append(b, ']', ',', '\n')
				code = code.end.next
			}
		case opMapHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				uptr := e.ptrToUnsafePtr(ptr)
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{', '\n')
					iter := mapiterinit(code.typ, uptr)
					ctx.keepRefs = append(ctx.keepRefs, iter)
					store(ctxptr, code.elemIdx, 0)
					store(ctxptr, code.length, uintptr(mlen))
					store(ctxptr, code.mapIter, uintptr(iter))

					if !e.unorderedMap {
						pos := make([]int, 0, mlen)
						pos = append(pos, len(b))
						posPtr := unsafe.Pointer(&pos)
						ctx.keepRefs = append(ctx.keepRefs, posPtr)
						store(ctxptr, code.end.mapPos, uintptr(posPtr))
					} else {
						b = e.encodeIndent(b, code.next.indent)
					}

					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '{', '}', ',', '\n')
					code = code.end.next
				}
			}
		case opMapHeadLoad:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				code = code.end.next
			} else {
				// load pointer
				ptr = e.ptrToPtr(ptr)
				uptr := e.ptrToUnsafePtr(ptr)
				if uintptr(uptr) == 0 {
					b = e.encodeIndent(b, code.indent)
					b = encodeNull(b)
					b = encodeIndentComma(b)
					code = code.end.next
					break
				}
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{', '\n')
					iter := mapiterinit(code.typ, uptr)
					ctx.keepRefs = append(ctx.keepRefs, iter)
					store(ctxptr, code.elemIdx, 0)
					store(ctxptr, code.length, uintptr(mlen))
					store(ctxptr, code.mapIter, uintptr(iter))
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))

					if !e.unorderedMap {
						pos := make([]int, 0, mlen)
						pos = append(pos, len(b))
						posPtr := unsafe.Pointer(&pos)
						ctx.keepRefs = append(ctx.keepRefs, posPtr)
						store(ctxptr, code.end.mapPos, uintptr(posPtr))
					} else {
						b = e.encodeIndent(b, code.next.indent)
					}

					code = code.next
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '{', '}', ',', '\n')
					code = code.end.next
				}
			}
		case opMapKey:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if e.unorderedMap {
				if idx < length {
					b = e.encodeIndent(b, code.indent)
					store(ctxptr, code.elemIdx, idx)
					ptr := load(ctxptr, code.mapIter)
					iter := e.ptrToUnsafePtr(ptr)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					last := len(b) - 1
					b[last] = '\n'
					b = e.encodeIndent(b, code.indent-1)
					b = append(b, '}', ',', '\n')
					code = code.end.next
				}
			} else {
				ptr := load(ctxptr, code.end.mapPos)
				posPtr := (*[]int)(*(*unsafe.Pointer)(unsafe.Pointer(&ptr)))
				*posPtr = append(*posPtr, len(b))
				if idx < length {
					ptr := load(ctxptr, code.mapIter)
					iter := e.ptrToUnsafePtr(ptr)
					store(ctxptr, code.elemIdx, idx)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					code = code.end
				}
			}
		case opMapValue:
			if e.unorderedMap {
				b = append(b, ':', ' ')
			} else {
				ptr := load(ctxptr, code.end.mapPos)
				posPtr := (*[]int)(*(*unsafe.Pointer)(unsafe.Pointer(&ptr)))
				*posPtr = append(*posPtr, len(b))
			}
			ptr := load(ctxptr, code.mapIter)
			iter := e.ptrToUnsafePtr(ptr)
			value := mapitervalue(iter)
			store(ctxptr, code.next.idx, uintptr(value))
			mapiternext(iter)
			code = code.next
		case opMapEnd:
			// this operation only used by sorted map
			length := int(load(ctxptr, code.length))
			type mapKV struct {
				key   string
				value string
			}
			kvs := make([]mapKV, 0, length)
			ptr := load(ctxptr, code.mapPos)
			pos := *(*[]int)(*(*unsafe.Pointer)(unsafe.Pointer(&ptr)))
			for i := 0; i < length; i++ {
				startKey := pos[i*2]
				startValue := pos[i*2+1]
				var endValue int
				if i+1 < length {
					endValue = pos[i*2+2]
				} else {
					endValue = len(b)
				}
				kvs = append(kvs, mapKV{
					key:   string(b[startKey:startValue]),
					value: string(b[startValue:endValue]),
				})
			}
			sort.Slice(kvs, func(i, j int) bool {
				return kvs[i].key < kvs[j].key
			})
			buf := b[pos[0]:]
			buf = buf[:0]
			for _, kv := range kvs {
				buf = append(buf, e.prefix...)
				buf = append(buf, bytes.Repeat(e.indentStr, code.indent+1)...)

				buf = append(buf, []byte(kv.key)...)
				buf[len(buf)-2] = ':'
				buf[len(buf)-1] = ' '
				buf = append(buf, []byte(kv.value)...)
			}
			buf = buf[:len(buf)-2]
			buf = append(buf, '\n')
			buf = append(buf, e.prefix...)
			buf = append(buf, bytes.Repeat(e.indentStr, code.indent)...)
			buf = append(buf, '}', ',', '\n')
			b = b[:pos[0]]
			b = append(b, buf...)
			code = code.next
		case opStructEscapedFieldPtrHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else if code.next == code.end {
				// not exists fields
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '}', ',', '\n')
				code = code.end.next
				store(ctxptr, code.idx, ptr)
			} else {
				b = append(b, '{', '\n')
				if !code.anonymousKey {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
				}
				p := ptr + code.offset
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructEscapedFieldHeadOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			if !code.anonymousKey {
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
			}
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldPtrHeadInt:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadIntOnly, opStructEscapedFieldHeadIntOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadIntPtr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadIntPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadIntPtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadIntNPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadIntOnly, opStructEscapedFieldAnonymousHeadIntOnly:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadIntPtr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadIntPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadIntPtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt8(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt8Only, opStructEscapedFieldHeadInt8Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt8(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt8Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt8Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt8PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt8PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadInt8NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt8Only, opStructEscapedFieldAnonymousHeadInt8Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt8Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt8Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt8PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt8PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt16(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt16Only, opStructEscapedFieldHeadInt16Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt16(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt16Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt16Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt16PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt16PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadInt16NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt16Only, opStructEscapedFieldAnonymousHeadInt16Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt16Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt16Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt16PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt16PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt32(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt32Only, opStructEscapedFieldHeadInt32Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt32(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt32PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadInt32NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt32Only, opStructEscapedFieldAnonymousHeadInt32Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt32PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, e.ptrToInt64(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt64Only, opStructEscapedFieldHeadInt64Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, e.ptrToInt64(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, e.ptrToInt64(p+code.offset))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt64PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadInt64NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, e.ptrToInt64(p+code.offset))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt64Only, opStructEscapedFieldAnonymousHeadInt64Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt64PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUintOnly, opStructEscapedFieldHeadUintOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUintPtr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUintPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUintPtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadUintNPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUintOnly, opStructEscapedFieldAnonymousHeadUintOnly:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUintPtr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUintPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUintPtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint8(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint8Only, opStructEscapedFieldHeadUint8Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint8(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint8Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint8Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint8PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint8PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadUint8NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint8Only, opStructEscapedFieldAnonymousHeadUint8Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint8Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint8Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint8PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint8PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint16(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint16Only, opStructEscapedFieldHeadUint16Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint16(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint16Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint16Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint16PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint16PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadUint16NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint16Only, opStructEscapedFieldAnonymousHeadUint16Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint16Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint16Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint16PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint16PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint32(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint32Only, opStructEscapedFieldHeadUint32Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint32(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint32PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadUint32NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint32Only, opStructEscapedFieldAnonymousHeadUint32Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint32PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, e.ptrToUint64(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint64Only, opStructEscapedFieldHeadUint64Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, e.ptrToUint64(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, e.ptrToUint64(p+code.offset))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint64PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadUint64NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, e.ptrToUint64(p+code.offset))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint64Only, opStructEscapedFieldAnonymousHeadUint64Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint64PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadFloat32Only, opStructEscapedFieldHeadFloat32Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = encodeFloat32(b, e.ptrToFloat32(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = encodeFloat32(b, e.ptrToFloat32(p))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadFloat32PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadFloat32NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = encodeFloat32(b, e.ptrToFloat32(p))
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadFloat32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadFloat32Only, opStructEscapedFieldAnonymousHeadFloat32Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadFloat32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadFloat32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat32PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadFloat64Only, opStructEscapedFieldHeadFloat64Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			v := e.ptrToFloat64(p)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					v := e.ptrToFloat64(p)
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = encodeFloat64(b, v)
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadFloat64PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := e.ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldHeadFloat64NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = e.ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					v := e.ptrToFloat64(p)
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = encodeFloat64(b, v)
				}
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadFloat64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadFloat64Only, opStructEscapedFieldAnonymousHeadFloat64Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadFloat64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := e.ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadFloat64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat64PtrOnly:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := e.ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadEscapedString:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeEscapedString(b, e.ptrToString(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadBool:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeBool(b, e.ptrToBool(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadBytes:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeByteSlice(b, e.ptrToBytes(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				p := ptr + code.offset
				if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendInt(b, int64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendInt(b, int64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendInt(b, int64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendInt(b, int64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendInt(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendUint(b, uint64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendUint(b, uint64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendUint(b, uint64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendUint(b, uint64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = appendUint(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToFloat32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = encodeFloat32(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToFloat64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = encodeFloat64(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToString(ptr + code.offset)
				if v == "" {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = encodeEscapedString(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToBool(ptr + code.offset)
				if !v {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = encodeBool(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToBytes(ptr + code.offset)
				if len(v) == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					b = encodeByteSlice(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				p := ptr + code.offset
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructEscapedFieldPtrHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = encodeFloat64(b, v)
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				s := e.ptrToString(ptr + code.offset)
				b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ', '"')
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedField:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldInt:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldInt8:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldInt16:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldInt32:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldInt64:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint8:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint16:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint32:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint64:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldFloat32:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldFloat64:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldEscapedString:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldBool:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldBytes:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldMarshalJSON:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = append(b, buf.Bytes()...)
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldArray:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			array := e.ptrToSlice(p)
			if p == 0 || uintptr(array.data) == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructEscapedFieldSlice:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructEscapedFieldMap:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				b = encodeNull(b)
				code = code.nextField
			} else {
				p = e.ptrToPtr(p)
				mlen := maplen(e.ptrToUnsafePtr(p))
				if mlen == 0 {
					b = append(b, '{', '}', ',', '\n')
					mapCode := code.next
					code = mapCode.end.next
				} else {
					code = code.next
				}
			}
		case opStructEscapedFieldMapLoad:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				b = encodeNull(b)
				code = code.nextField
			} else {
				p = e.ptrToPtr(p)
				mlen := maplen(e.ptrToUnsafePtr(p))
				if mlen == 0 {
					b = append(b, '{', '}', ',', '\n')
					code = code.nextField
				} else {
					code = code.next
				}
			}
		case opStructEscapedFieldStruct:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			if p == 0 {
				b = append(b, '{', '}', ',', '\n')
				code = code.nextField
			} else {
				headCode := code.next
				if headCode.next == headCode.end {
					// not exists fields
					b = append(b, '{', '}', ',', '\n')
					code = code.nextField
				} else {
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructEscapedFieldOmitEmpty:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 || **(**uintptr)(unsafe.Pointer(&p)) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructEscapedFieldOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat32(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyEscapedString:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeEscapedString(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeBool(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeByteSlice(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyArray:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			array := e.ptrToSlice(p)
			if p == 0 || uintptr(array.data) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				code = code.next
			}
		case opStructEscapedFieldOmitEmptySlice:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				code = code.next
			}
		case opStructEscapedFieldOmitEmptyMap:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				mlen := maplen(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
				if mlen == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					code = code.next
				}
			}
		case opStructEscapedFieldOmitEmptyMapLoad:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				mlen := maplen(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
				if mlen == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, code.escapedKey...)
					b = append(b, ' ')
					code = code.next
				}
			}
		case opStructEscapedFieldOmitEmptyStruct:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				headCode := code.next
				if headCode.next == headCode.end {
					// not exists fields
					b = append(b, '{', '}', ',', '\n')
					code = code.nextField
				} else {
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructEscapedFieldStringTag:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt8:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt16:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint8:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint16:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			s := e.ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = encodeEscapedString(b, buf.String())
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructAnonymousEnd:
			code = code.next
		case opStructEscapedEnd:
			last := len(b) - 1
			if b[last-1] == '{' {
				b[last] = '}'
				b = encodeIndentComma(b)
				code = code.next
				break
			}
			if b[last] == '\n' {
				// to remove ',' and '\n' characters
				b = b[:len(b)-2]
			}
			b = append(b, '\n')
			b = e.encodeIndent(b, code.indent)
			b = append(b, '}')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedEndInt:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndIntPtr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p)))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt8:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt8Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p)))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt16:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt16Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p)))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt32:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt32Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p)))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt64:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt64Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUintPtr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p)))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint8:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint8Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p)))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint16:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint16Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p)))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint32:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint32Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p)))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint64:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint64Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndFloat32:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndFloat32Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndFloat64:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndFloat64Ptr:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := e.ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndEscapedString:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndBool:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndBytes:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndMarshalJSON:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = append(b, buf.Bytes()...)
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyInt8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyInt16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyInt32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyInt64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyUint8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyUint16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyUint32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyUint64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendUint(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat32(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyEscapedString:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeEscapedString(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeBool(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeByteSlice(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt8:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt16:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint8:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint16:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			s := e.ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = encodeEscapedString(b, buf.String())
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opEnd:
			goto END
		}
	}
END:
	return b, nil
}

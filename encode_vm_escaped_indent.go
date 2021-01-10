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
		case opPtrIndent:
			ptr := load(ctxptr, code.idx)
			code = code.next
			store(ctxptr, code.idx, e.ptrToPtr(ptr))
		case opIntIndent:
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt8Indent:
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt16Indent:
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt32Indent:
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt64Indent:
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opUintIndent:
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint8Indent:
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint16Indent:
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint32Indent:
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint64Indent:
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opIntStringIndent:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt8StringIndent:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt16StringIndent:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt32StringIndent:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt64StringIndent:
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUintStringIndent:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint8StringIndent:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint16StringIndent:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint32StringIndent:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint64StringIndent:
			b = append(b, '"')
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opFloat32Indent:
			b = encodeFloat32(b, e.ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opFloat64Indent:
			v := e.ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opEscapedStringIndent:
			b = encodeEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBoolIndent:
			b = encodeBool(b, e.ptrToBool(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBytesIndent:
			ptr := load(ctxptr, code.idx)
			slice := e.ptrToSlice(ptr)
			if ptr == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, e.ptrToBytes(ptr))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opInterfaceIndent:
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
			c = toIndent(c)
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
			lastCode.op = opInterfaceEndIndent
			lastCode.next = code.next

			code = c
			recursiveLevel++
		case opInterfaceEndIndent:
			recursiveLevel--
			// restore ctxptr
			offset := load(ctxptr, code.idx)
			ctxptr = ctx.ptr() + offset
			ptrOffset = offset
			code = code.next
		case opMarshalJSONIndent:
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
		case opMarshalTextIndent:
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
		case opSliceHeadIndent:
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
		case opRootSliceHeadIndent:
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
		case opSliceElemIndent:
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
		case opRootSliceElemIndent:
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
		case opArrayHeadIndent:
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
		case opArrayElemIndent:
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
		case opMapHeadIndent:
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
		case opMapHeadLoadIndent:
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
		case opMapKeyIndent:
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
		case opMapValueIndent:
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
		case opMapEndIndent:
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
		case opStructEscapedFieldPtrHeadIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadIndent:
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
		case opStructEscapedFieldHeadOnlyIndent:
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
		case opStructEscapedFieldPtrHeadIntIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadIntIndent:
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
		case opStructEscapedFieldPtrHeadIntOnlyIndent, opStructEscapedFieldHeadIntOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadIntPtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadIntPtrIndent:
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
		case opStructEscapedFieldPtrHeadIntPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadIntPtrOnlyIndent:
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
		case opStructEscapedFieldHeadIntNPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadIntIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadIntIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadIntOnlyIndent, opStructEscapedFieldAnonymousHeadIntOnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadIntPtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadIntPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadIntPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadIntPtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadInt8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt8Indent:
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
		case opStructEscapedFieldPtrHeadInt8OnlyIndent, opStructEscapedFieldHeadInt8OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt8(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt8PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt8PtrIndent:
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
		case opStructEscapedFieldPtrHeadInt8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt8PtrOnlyIndent:
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
		case opStructEscapedFieldHeadInt8NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt8Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt8OnlyIndent, opStructEscapedFieldAnonymousHeadInt8OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt8PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt8PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt8PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadInt16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt16Indent:
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
		case opStructEscapedFieldPtrHeadInt16OnlyIndent, opStructEscapedFieldHeadInt16OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt16(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt16PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt16PtrIndent:
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
		case opStructEscapedFieldPtrHeadInt16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt16PtrOnlyIndent:
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
		case opStructEscapedFieldHeadInt16NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt16Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt16OnlyIndent, opStructEscapedFieldAnonymousHeadInt16OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt16PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt16PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt16PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadInt32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt32Indent:
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
		case opStructEscapedFieldPtrHeadInt32OnlyIndent, opStructEscapedFieldHeadInt32OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt32(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt32PtrIndent:
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
		case opStructEscapedFieldPtrHeadInt32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt32PtrOnlyIndent:
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
		case opStructEscapedFieldHeadInt32NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt32Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt32OnlyIndent, opStructEscapedFieldAnonymousHeadInt32OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt32PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt32PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadInt64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt64Indent:
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
		case opStructEscapedFieldPtrHeadInt64OnlyIndent, opStructEscapedFieldHeadInt64OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendInt(b, e.ptrToInt64(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt64PtrIndent:
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
		case opStructEscapedFieldPtrHeadInt64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt64PtrOnlyIndent:
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
		case opStructEscapedFieldHeadInt64NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt64Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt64OnlyIndent, opStructEscapedFieldAnonymousHeadInt64OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt64PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadInt64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt64PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadUintIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUintIndent:
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
		case opStructEscapedFieldPtrHeadUintOnlyIndent, opStructEscapedFieldHeadUintOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUintPtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUintPtrIndent:
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
		case opStructEscapedFieldPtrHeadUintPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUintPtrOnlyIndent:
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
		case opStructEscapedFieldHeadUintNPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUintIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUintIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUintOnlyIndent, opStructEscapedFieldAnonymousHeadUintOnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUintPtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUintPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUintPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUintPtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadUint8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint8Indent:
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
		case opStructEscapedFieldPtrHeadUint8OnlyIndent, opStructEscapedFieldHeadUint8OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint8(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint8PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint8PtrIndent:
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
		case opStructEscapedFieldPtrHeadUint8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint8PtrOnlyIndent:
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
		case opStructEscapedFieldHeadUint8NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint8Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint8OnlyIndent, opStructEscapedFieldAnonymousHeadUint8OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint8PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint8PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint8PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadUint16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint16Indent:
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
		case opStructEscapedFieldPtrHeadUint16OnlyIndent, opStructEscapedFieldHeadUint16OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint16(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint16PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint16PtrIndent:
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
		case opStructEscapedFieldPtrHeadUint16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint16PtrOnlyIndent:
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
		case opStructEscapedFieldHeadUint16NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint16Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint16OnlyIndent, opStructEscapedFieldAnonymousHeadUint16OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint16PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint16PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint16PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadUint32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint32Indent:
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
		case opStructEscapedFieldPtrHeadUint32OnlyIndent, opStructEscapedFieldHeadUint32OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint32(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint32PtrIndent:
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
		case opStructEscapedFieldPtrHeadUint32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint32PtrOnlyIndent:
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
		case opStructEscapedFieldHeadUint32NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint32Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint32OnlyIndent, opStructEscapedFieldAnonymousHeadUint32OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint32PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint32PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadUint64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint64Indent:
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
		case opStructEscapedFieldPtrHeadUint64OnlyIndent, opStructEscapedFieldHeadUint64OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = appendUint(b, e.ptrToUint64(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint64PtrIndent:
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
		case opStructEscapedFieldPtrHeadUint64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint64PtrOnlyIndent:
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
		case opStructEscapedFieldHeadUint64NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint64Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint64OnlyIndent, opStructEscapedFieldAnonymousHeadUint64OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint64PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadUint64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint64PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadFloat32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat32Indent:
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
		case opStructEscapedFieldPtrHeadFloat32OnlyIndent, opStructEscapedFieldHeadFloat32OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = encodeFloat32(b, e.ptrToFloat32(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat32PtrIndent:
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
		case opStructEscapedFieldPtrHeadFloat32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadFloat32PtrOnlyIndent:
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
		case opStructEscapedFieldHeadFloat32NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadFloat32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat32Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadFloat32OnlyIndent, opStructEscapedFieldAnonymousHeadFloat32OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadFloat32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat32PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadFloat32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat32PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadFloat64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat64Indent:
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
		case opStructEscapedFieldPtrHeadFloat64OnlyIndent, opStructEscapedFieldHeadFloat64OnlyIndent:
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
		case opStructEscapedFieldPtrHeadFloat64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat64PtrIndent:
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
		case opStructEscapedFieldPtrHeadFloat64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadFloat64PtrOnlyIndent:
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
		case opStructEscapedFieldHeadFloat64NPtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadFloat64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat64Indent:
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
		case opStructEscapedFieldPtrAnonymousHeadFloat64OnlyIndent, opStructEscapedFieldAnonymousHeadFloat64OnlyIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadFloat64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat64PtrIndent:
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
		case opStructEscapedFieldPtrAnonymousHeadFloat64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat64PtrOnlyIndent:
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
		case opStructEscapedFieldPtrHeadEscapedStringIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadEscapedStringIndent:
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
		case opStructEscapedFieldPtrHeadBoolIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadBoolIndent:
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
		case opStructEscapedFieldPtrHeadBytesIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadBytesIndent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyIndent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyIntIndent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt8Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt16Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt32Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyInt64Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUintIndent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint8Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint16Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint32Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyUint64Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyFloat32Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyFloat64Indent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyEscapedStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyEscapedStringIndent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyBoolIndent:
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
		case opStructEscapedFieldPtrHeadOmitEmptyBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyBytesIndent:
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
		case opStructEscapedFieldPtrHeadStringTagIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagIndent:
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
		case opStructEscapedFieldPtrHeadStringTagIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagIntIndent:
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
		case opStructEscapedFieldPtrHeadStringTagInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt8Indent:
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
		case opStructEscapedFieldPtrHeadStringTagInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt16Indent:
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
		case opStructEscapedFieldPtrHeadStringTagInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt32Indent:
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
		case opStructEscapedFieldPtrHeadStringTagInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt64Indent:
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
		case opStructEscapedFieldPtrHeadStringTagUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUintIndent:
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
		case opStructEscapedFieldPtrHeadStringTagUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint8Indent:
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
		case opStructEscapedFieldPtrHeadStringTagUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint16Indent:
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
		case opStructEscapedFieldPtrHeadStringTagUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint32Indent:
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
		case opStructEscapedFieldPtrHeadStringTagUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagUint64Indent:
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
		case opStructEscapedFieldPtrHeadStringTagFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagFloat32Indent:
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
		case opStructEscapedFieldPtrHeadStringTagFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagFloat64Indent:
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
		case opStructEscapedFieldPtrHeadStringTagEscapedStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagEscapedStringIndent:
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
		case opStructEscapedFieldPtrHeadStringTagBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagBoolIndent:
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
		case opStructEscapedFieldPtrHeadStringTagBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagBytesIndent:
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
		case opStructEscapedFieldIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldIntIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldInt8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldInt16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldInt32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldInt64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUintIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldUint64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldFloat32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldFloat64Indent:
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
		case opStructEscapedFieldEscapedStringIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldBoolIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldBytesIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldMarshalJSONIndent:
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
		case opStructEscapedFieldArrayIndent:
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
		case opStructEscapedFieldSliceIndent:
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
		case opStructEscapedFieldMapIndent:
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
		case opStructEscapedFieldMapLoadIndent:
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
		case opStructEscapedFieldStructIndent:
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
		case opStructEscapedFieldOmitEmptyIndent:
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
		case opStructEscapedFieldOmitEmptyIntIndent:
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
		case opStructEscapedFieldOmitEmptyInt8Indent:
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
		case opStructEscapedFieldOmitEmptyInt16Indent:
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
		case opStructEscapedFieldOmitEmptyInt32Indent:
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
		case opStructEscapedFieldOmitEmptyInt64Indent:
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
		case opStructEscapedFieldOmitEmptyUintIndent:
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
		case opStructEscapedFieldOmitEmptyUint8Indent:
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
		case opStructEscapedFieldOmitEmptyUint16Indent:
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
		case opStructEscapedFieldOmitEmptyUint32Indent:
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
		case opStructEscapedFieldOmitEmptyUint64Indent:
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
		case opStructEscapedFieldOmitEmptyFloat32Indent:
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
		case opStructEscapedFieldOmitEmptyFloat64Indent:
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
		case opStructEscapedFieldOmitEmptyEscapedStringIndent:
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
		case opStructEscapedFieldOmitEmptyBoolIndent:
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
		case opStructEscapedFieldOmitEmptyBytesIndent:
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
		case opStructEscapedFieldOmitEmptyArrayIndent:
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
		case opStructEscapedFieldOmitEmptySliceIndent:
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
		case opStructEscapedFieldOmitEmptyMapIndent:
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
		case opStructEscapedFieldOmitEmptyMapLoadIndent:
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
		case opStructEscapedFieldOmitEmptyStructIndent:
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
		case opStructEscapedFieldStringTagIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldStringTagIntIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUintIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagFloat64Indent:
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
		case opStructEscapedFieldStringTagEscapedStringIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			s := e.ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEscapedFieldStringTagMarshalJSONIndent:
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
		case opStructEscapedFieldStringTagMarshalTextIndent:
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
		case opStructAnonymousEndIndent:
			code = code.next
		case opStructEscapedEndIndent:
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
		case opStructEscapedEndIntIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndIntPtrIndent:
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
		case opStructEscapedEndInt8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt8PtrIndent:
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
		case opStructEscapedEndInt16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt16PtrIndent:
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
		case opStructEscapedEndInt32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt32PtrIndent:
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
		case opStructEscapedEndInt64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndInt64PtrIndent:
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
		case opStructEscapedEndUintIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUintPtrIndent:
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
		case opStructEscapedEndUint8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint8PtrIndent:
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
		case opStructEscapedEndUint16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint16PtrIndent:
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
		case opStructEscapedEndUint32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint32PtrIndent:
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
		case opStructEscapedEndUint64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndUint64PtrIndent:
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
		case opStructEscapedEndFloat32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndFloat32PtrIndent:
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
		case opStructEscapedEndFloat64Indent:
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
		case opStructEscapedEndFloat64PtrIndent:
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
		case opStructEscapedEndEscapedStringIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndBoolIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndBytesIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndMarshalJSONIndent:
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
		case opStructEscapedEndOmitEmptyIntIndent:
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
		case opStructEscapedEndOmitEmptyInt8Indent:
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
		case opStructEscapedEndOmitEmptyInt16Indent:
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
		case opStructEscapedEndOmitEmptyInt32Indent:
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
		case opStructEscapedEndOmitEmptyInt64Indent:
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
		case opStructEscapedEndOmitEmptyUintIndent:
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
		case opStructEscapedEndOmitEmptyUint8Indent:
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
		case opStructEscapedEndOmitEmptyUint16Indent:
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
		case opStructEscapedEndOmitEmptyUint32Indent:
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
		case opStructEscapedEndOmitEmptyUint64Indent:
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
		case opStructEscapedEndOmitEmptyFloat32Indent:
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
		case opStructEscapedEndOmitEmptyFloat64Indent:
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
		case opStructEscapedEndOmitEmptyEscapedStringIndent:
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
		case opStructEscapedEndOmitEmptyBoolIndent:
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
		case opStructEscapedEndOmitEmptyBytesIndent:
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
		case opStructEscapedEndStringTagIntIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUintIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagFloat64Indent:
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
		case opStructEscapedEndStringTagEscapedStringIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			s := e.ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ', '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndStringTagMarshalJSONIndent:
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
		case opStructEscapedEndStringTagMarshalTextIndent:
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

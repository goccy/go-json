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

func (e *Encoder) runIndent(ctx *encodeRuntimeContext, b []byte, code *opcode) ([]byte, error) {
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
		case opStringIndent:
			b = encodeNoEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
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
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructFieldPtrHeadIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
				}
				p := ptr + code.offset
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldHeadOnlyIndent:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			if !code.anonymousKey {
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
			}
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadIntIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadIntOnlyIndent, opStructFieldHeadIntOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadIntPtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadIntPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadIntPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadIntPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadIntNPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadIntIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadIntOnlyIndent, opStructFieldAnonymousHeadIntOnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadIntPtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadIntPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadIntPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadIntPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadInt8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt8(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt8OnlyIndent, opStructFieldHeadInt8OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt8(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadInt8PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt8PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadInt8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadInt8NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadInt8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt8OnlyIndent, opStructFieldAnonymousHeadInt8OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt8PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt8PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadInt8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadInt8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadInt16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt16(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt16OnlyIndent, opStructFieldHeadInt16OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt16(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadInt16PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt16PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadInt16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadInt16NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadInt16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt16OnlyIndent, opStructFieldAnonymousHeadInt16OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt16PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt16PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadInt16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadInt16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadInt32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt32(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt32OnlyIndent, opStructFieldHeadInt32OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendInt(b, int64(e.ptrToInt32(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadInt32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt32PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadInt32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadInt32NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadInt32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt32OnlyIndent, opStructFieldAnonymousHeadInt32OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt32PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadInt32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadInt32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadInt64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, e.ptrToInt64(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt64OnlyIndent, opStructFieldHeadInt64OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendInt(b, e.ptrToInt64(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadInt64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt64PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadInt64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadInt64NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadInt64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt64OnlyIndent, opStructFieldAnonymousHeadInt64OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt64PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadInt64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadInt64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUintIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUintOnlyIndent, opStructFieldHeadUintOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUintPtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUintPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadUintPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUintPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadUintNPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadUintIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUintOnlyIndent, opStructFieldAnonymousHeadUintOnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUintPtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUintPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadUintPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadUintPtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUint8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint8(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint8OnlyIndent, opStructFieldHeadUint8OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint8(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUint8PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint8PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadUint8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadUint8NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadUint8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint8OnlyIndent, opStructFieldAnonymousHeadUint8OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint8PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint8PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadUint8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadUint8PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUint16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint16(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint16OnlyIndent, opStructFieldHeadUint16OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint16(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUint16PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint16PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadUint16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadUint16NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadUint16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint16OnlyIndent, opStructFieldAnonymousHeadUint16OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint16PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint16PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadUint16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadUint16PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUint32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint32(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint32OnlyIndent, opStructFieldHeadUint32OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendUint(b, uint64(e.ptrToUint32(p)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUint32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint32PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadUint32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadUint32NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadUint32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint32OnlyIndent, opStructFieldAnonymousHeadUint32OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint32PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadUint32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadUint32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUint64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, e.ptrToUint64(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint64OnlyIndent, opStructFieldHeadUint64OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendUint(b, e.ptrToUint64(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUint64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint64PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadUint64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadUint64NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadUint64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint64OnlyIndent, opStructFieldAnonymousHeadUint64OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint64PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadUint64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadUint64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadFloat32OnlyIndent, opStructFieldHeadFloat32OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeFloat32(b, e.ptrToFloat32(p))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat32PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadFloat32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadFloat32NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadFloat32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadFloat32OnlyIndent, opStructFieldAnonymousHeadFloat32OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadFloat32PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat32PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadFloat32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadFloat32PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat64Indent:
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
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadFloat64OnlyIndent, opStructFieldHeadFloat64OnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			v := e.ptrToFloat64(p)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat64PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadFloat64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = append(b, '{', '\n')
			b = e.encodeIndent(b, code.indent+1)
			b = append(b, code.key...)
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
		case opStructFieldHeadFloat64NPtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadFloat64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadFloat64OnlyIndent, opStructFieldAnonymousHeadFloat64OnlyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadFloat64PtrIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat64PtrIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadFloat64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadFloat64PtrOnlyIndent:
			p := load(ctxptr, code.idx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadStringIndent:
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
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, e.ptrToString(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadBoolIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadBoolIndent:
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
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, e.ptrToBool(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadBytesIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadBytesIndent:
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
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, e.ptrToBytes(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadOmitEmptyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructFieldPtrHeadOmitEmptyIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyIntIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendInt(b, int64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt8Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendInt(b, int64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt16Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendInt(b, int64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt32Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendInt(b, int64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt64Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendInt(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUintIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendUint(b, uint64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint8Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendUint(b, uint64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint16Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendUint(b, uint64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint32Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendUint(b, uint64(v))
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint64Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = appendUint(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = encodeFloat32(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64Indent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = encodeFloat64(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyStringIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = encodeNoEscapedString(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBoolIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = encodeBool(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBytesIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					b = encodeByteSlice(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadStringTagIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagIndent:
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
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadStringTagIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat64Indent:
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
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = encodeFloat64(b, v)
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				s := e.ptrToString(ptr + code.offset)
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ', '"')
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldIntIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldInt8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldInt16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldInt32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldInt64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUintIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldFloat32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldFloat64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeNoEscapedString(b, e.ptrToString(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldBoolIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldBytesIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldMarshalJSONIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldArrayIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldSliceIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldMapIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldMapLoadIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldStructIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldOmitEmptyIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 || **(**uintptr)(unsafe.Pointer(&p)) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldOmitEmptyIntIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUintIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyFloat64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyStringIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyArrayIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			array := e.ptrToSlice(p)
			if p == 0 || uintptr(array.data) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
			}
		case opStructFieldOmitEmptySliceIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
			}
		case opStructFieldOmitEmptyMapIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					code = code.next
				}
			}
		case opStructFieldOmitEmptyMapLoadIndent:
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
					b = append(b, code.key...)
					b = append(b, ' ')
					code = code.next
				}
			}
		case opStructFieldOmitEmptyStructIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
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
		case opStructFieldStringTagIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldStringTagIntIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUintIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagFloat64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagStringIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			s := e.ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagMarshalJSONIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructFieldStringTagMarshalTextIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructAnonymousEndIndent:
			code = code.next
		case opStructEndIndent:
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
		case opStructEndIntIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndIntPtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndInt8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndInt8PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndInt16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndInt16PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndInt32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndInt32PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndInt64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndInt64PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndUintIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndUintPtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndUint8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndUint8PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndUint16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndUint16PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndUint32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndUint32PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndUint64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndUint64PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndFloat32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndFloat32PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndFloat64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndFloat64PtrIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndStringIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeNoEscapedString(b, e.ptrToString(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndBoolIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndBytesIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndMarshalJSONIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
		case opStructEndOmitEmptyIntIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyUintIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, uint64(v))
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyFloat64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyStringIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = e.encodeIndent(b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, v)
			}
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagIntIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagUintIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagFloat64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagStringIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			s := e.ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagMarshalJSONIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
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
			b = encodeNoEscapedString(b, buf.String())
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEndStringTagMarshalTextIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opEnd:
			goto END
		}
	}
END:
	return b, nil
}

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

func (e *Encoder) runEscaped(ctx *encodeRuntimeContext, b []byte, code *opcode) ([]byte, error) {
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
			b = encodeComma(b)
			code = code.next
		case opInt8:
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opInt16:
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opInt32:
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opInt64:
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opUint:
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opUint8:
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opUint16:
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opUint32:
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opUint64:
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opIntString:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opInt8String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opInt16String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opInt32String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opInt64String:
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUintString:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUint8String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUint16String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUint32String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUint64String:
			b = append(b, '"')
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opFloat32:
			b = encodeFloat32(b, e.ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opFloat64:
			v := e.ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opString:
			b = encodeNoEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opEscapedString:
			b = encodeEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opBool:
			b = encodeBool(b, e.ptrToBool(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opBytes:
			ptr := load(ctxptr, code.idx)
			slice := e.ptrToSlice(ptr)
			if ptr == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, e.ptrToBytes(ptr))
			}
			b = encodeComma(b)
			code = code.next
		case opInterface:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
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
				b = encodeComma(b)
				code = code.next
				break
			}
			vv := rv.Interface()
			header := (*interfaceHeader)(unsafe.Pointer(&vv))
			if header.typ.Kind() == reflect.Ptr {
				if rv.Elem().IsNil() {
					b = encodeNull(b)
					b = encodeComma(b)
					code = code.next
					break
				}
			}
			c, err := e.compileHead(&encodeCompileContext{
				typ:                      header.typ,
				root:                     code.root,
				indent:                   code.indent,
				structTypeToCompiledCode: map[uintptr]*compiledCode{},
			})
			if err != nil {
				return nil, err
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
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = append(append(b, buf.Bytes()...), ',')
			code = code.next
		case opMarshalText:
			ptr := load(ctxptr, code.idx)
			isPtr := code.typ.Kind() == reflect.Ptr
			p := e.ptrToUnsafePtr(ptr)
			if p == nil || isPtr && *(*unsafe.Pointer)(p) == nil {
				b = append(b, '"', '"', ',')
			} else {
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: p,
				}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				if e.enabledHTMLEscape {
					b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				} else {
					b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				}
				b = encodeComma(b)
			}
			code = code.next
		case opSliceHead:
			p := load(ctxptr, code.idx)
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				store(ctxptr, code.elemIdx, 0)
				store(ctxptr, code.length, uintptr(slice.len))
				store(ctxptr, code.idx, uintptr(slice.data))
				if slice.len > 0 {
					b = append(b, '[')
					code = code.next
					store(ctxptr, code.idx, uintptr(slice.data))
				} else {
					b = append(b, '[', ']', ',')
					code = code.end.next
				}
			}
		case opSliceElem:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if idx < length {
				store(ctxptr, code.elemIdx, idx)
				data := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, data+idx*size)
			} else {
				last := len(b) - 1
				b[last] = ']'
				b = encodeComma(b)
				code = code.end.next
			}
		case opArrayHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				if code.length > 0 {
					b = append(b, '[')
					store(ctxptr, code.elemIdx, 0)
					code = code.next
					store(ctxptr, code.idx, p)
				} else {
					b = append(b, '[', ']', ',')
					code = code.end.next
				}
			}
		case opArrayElem:
			idx := load(ctxptr, code.elemIdx)
			idx++
			if idx < code.length {
				store(ctxptr, code.elemIdx, idx)
				p := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, p+idx*size)
			} else {
				last := len(b) - 1
				b[last] = ']'
				b = encodeComma(b)
				code = code.end.next
			}
		case opMapHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				uptr := e.ptrToUnsafePtr(ptr)
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{')
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
					}
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					b = append(b, '{', '}', ',')
					code = code.end.next
				}
			}
		case opMapHeadLoad:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				// load pointer
				ptr = e.ptrToPtr(ptr)
				uptr := e.ptrToUnsafePtr(ptr)
				if ptr == 0 {
					b = encodeNull(b)
					b = encodeComma(b)
					code = code.end.next
					break
				}
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{')
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
					}
					code = code.next
				} else {
					b = append(b, '{', '}', ',')
					code = code.end.next
				}
			}
		case opMapKey:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if e.unorderedMap {
				if idx < length {
					ptr := load(ctxptr, code.mapIter)
					iter := e.ptrToUnsafePtr(ptr)
					store(ctxptr, code.elemIdx, idx)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					last := len(b) - 1
					b[last] = '}'
					b = encodeComma(b)
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
				last := len(b) - 1
				b[last] = ':'
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
			// this operation only used by sorted map.
			length := int(load(ctxptr, code.length))
			type mapKV struct {
				key   string
				value string
			}
			kvs := make([]mapKV, 0, length)
			ptr := load(ctxptr, code.mapPos)
			posPtr := e.ptrToUnsafePtr(ptr)
			pos := *(*[]int)(posPtr)
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
				buf = append(buf, []byte(kv.key)...)
				buf[len(buf)-1] = ':'
				buf = append(buf, []byte(kv.value)...)
			}
			buf[len(buf)-1] = '}'
			buf = append(buf, ',')
			b = b[:pos[0]]
			b = append(b, buf...)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadRecursive:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadRecursive:
			fallthrough
		case opStructEscapedFieldRecursive:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				if recursiveLevel > startDetectingCyclesAfter {
					if _, exists := seenPtr[ptr]; exists {
						return nil, errUnsupportedValue(code, ptr)
					}
				}
			}
			if seenPtr == nil {
				seenPtr = map[uintptr]struct{}{}
			}
			seenPtr[ptr] = struct{}{}
			c := toEscaped(code.jmp.code)
			c.end.next = newEndOp(&encodeCompileContext{})
			c.op = c.op.ptrHeadToHead()

			beforeLastCode := c.end
			lastCode := beforeLastCode.next

			lastCode.idx = beforeLastCode.idx + uintptrSize
			lastCode.elemIdx = lastCode.idx + uintptrSize

			// extend length to alloc slot for elemIdx
			totalLength := uintptr(code.totalLength() + 1)
			nextTotalLength := uintptr(c.totalLength() + 1)

			curlen := uintptr(len(ctx.ptrs))
			offsetNum := ptrOffset / uintptrSize
			oldOffset := ptrOffset
			ptrOffset += totalLength * uintptrSize

			newLen := offsetNum + totalLength + nextTotalLength
			if curlen < newLen {
				ctx.ptrs = append(ctx.ptrs, make([]uintptr, newLen-curlen)...)
			}
			ctxptr = ctx.ptr() + ptrOffset // assign new ctxptr

			store(ctxptr, c.idx, ptr)
			store(ctxptr, lastCode.idx, oldOffset)
			store(ctxptr, lastCode.elemIdx, uintptr(unsafe.Pointer(code.next)))

			// link lastCode ( opStructFieldRecursiveEnd ) => code.next
			lastCode.op = opStructFieldRecursiveEnd
			code = c
			recursiveLevel++
		case opStructFieldRecursiveEnd:
			recursiveLevel--

			// restore ctxptr
			offset := load(ctxptr, code.idx)
			ptr := load(ctxptr, code.elemIdx)
			code = (*opcode)(e.ptrToUnsafePtr(ptr))
			ctxptr = ctx.ptr() + offset
			ptrOffset = offset
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = append(b, code.escapedKey...)
				}
				p := ptr + code.offset
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructEscapedFieldHeadOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{')
			if !code.anonymousKey {
				b = append(b, code.escapedKey...)
			}
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldAnonymousHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				code = code.next
				store(ctxptr, code.idx, ptr)
			}
		case opStructEscapedFieldPtrHeadInt:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadIntOnly, opStructEscapedFieldHeadIntOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadIntPtr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadIntPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadIntPtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldHeadIntNPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
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
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadIntOnly, opStructEscapedFieldAnonymousHeadIntOnly:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt8Only, opStructEscapedFieldHeadInt8Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt8(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt8Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt8Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt8PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt8PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt8Only, opStructEscapedFieldAnonymousHeadInt8Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt16Only, opStructEscapedFieldHeadInt16Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt16(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt16Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt16Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt16PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt16PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt16Only, opStructEscapedFieldAnonymousHeadInt16Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt32Only, opStructEscapedFieldHeadInt32Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt32(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt32PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt32Only, opStructEscapedFieldAnonymousHeadInt32Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt64Only, opStructEscapedFieldHeadInt64Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendInt(b, e.ptrToInt64(p))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, e.ptrToInt64(p+code.offset))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt64PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadInt64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadInt64Only, opStructEscapedFieldAnonymousHeadInt64Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUintOnly, opStructEscapedFieldHeadUintOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUintPtr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUintPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUintPtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldHeadUintNPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
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
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUintOnly, opStructEscapedFieldAnonymousHeadUintOnly:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint8Only, opStructEscapedFieldHeadUint8Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint8(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint8Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint8Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint8PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint8PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint8Only, opStructEscapedFieldAnonymousHeadUint8Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint16Only, opStructEscapedFieldHeadUint16Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint16(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint16Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint16Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint16PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint16PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint16Only, opStructEscapedFieldAnonymousHeadUint16Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint32Only, opStructEscapedFieldHeadUint32Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint32(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint32PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint32Only, opStructEscapedFieldAnonymousHeadUint32Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint64Only, opStructEscapedFieldHeadUint64Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendUint(b, e.ptrToUint64(p))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, e.ptrToUint64(p+code.offset))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadUint64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint64PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadUint64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadUint64Only, opStructEscapedFieldAnonymousHeadUint64Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadFloat32Only, opStructEscapedFieldHeadFloat32Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = encodeFloat32(b, e.ptrToFloat32(p))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat32Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = encodeFloat32(b, e.ptrToFloat32(p+code.offset))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat32PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadFloat32PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadFloat32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadFloat32Only, opStructEscapedFieldAnonymousHeadFloat32Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p+code.offset))
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadFloat64Only, opStructEscapedFieldHeadFloat64Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			v := e.ptrToFloat64(p)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat64Ptr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					v := e.ptrToFloat64(p + code.offset)
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = encodeFloat64(b, v)
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadFloat64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadFloat64PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := e.ptrToFloat64(p + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadFloat64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.escapedKey...)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadFloat64Only, opStructEscapedFieldAnonymousHeadFloat64Only:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := e.ptrToFloat64(p + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeComma(b)
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
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := e.ptrToFloat64(p + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadEscapedString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadEscapedStringOnly, opStructEscapedFieldHeadEscapedStringOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, e.ptrToString(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadEscapedStringPtr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadEscapedStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = encodeEscapedString(b, e.ptrToString(p+code.offset))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadEscapedStringPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadEscapedStringPtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, e.ptrToString(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadEscapedString:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadEscapedStringOnly, opStructEscapedFieldAnonymousHeadEscapedStringOnly:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadEscapedStringPtr:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadEscapedStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.escapedKey...)
			p = e.ptrToPtr(p)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, e.ptrToString(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadEscapedStringPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldAnonymousHeadEscapedStringPtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, e.ptrToString(p+code.offset))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadBool {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadBoolOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructEscapedFieldHeadBoolOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.key...)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadBool:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadBytes {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadBytes:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadArray:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadArray:
			ptr := load(ctxptr, code.idx) + code.offset
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadArray {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '[', ']', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = append(b, code.escapedKey...)
				}
				code = code.next
				store(ctxptr, code.idx, ptr)
			}
		case opStructEscapedFieldPtrAnonymousHeadArray:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadArray:
			ptr := load(ctxptr, code.idx) + code.offset
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				store(ctxptr, code.idx, ptr)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadSlice:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadSlice:
			ptr := load(ctxptr, code.idx)
			p := ptr + code.offset
			if p == 0 {
				if code.op == opStructEscapedFieldPtrHeadSlice {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '[', ']', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = append(b, code.escapedKey...)
				}
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructEscapedFieldPtrAnonymousHeadSlice:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadSlice:
			ptr := load(ctxptr, code.idx)
			p := ptr + code.offset
			if p == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				store(ctxptr, code.idx, p)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				ptr += code.offset
				v := e.ptrToInterface(code, ptr)
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					b = encodeNull(b)
					code = code.end
					break
				}
				bb, err := rv.Interface().(Marshaler).MarshalJSON()
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
				if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
					return nil, err
				}
				b = append(b, buf.Bytes()...)
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadMarshalJSON:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				ptr += code.offset
				v := e.ptrToInterface(code, ptr)
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					b = encodeNull(b)
					code = code.end.next
					break
				}
				bb, err := rv.Interface().(Marshaler).MarshalJSON()
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
				if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
					return nil, err
				}
				b = append(b, buf.Bytes()...)
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				ptr += code.offset
				v := e.ptrToInterface(code, ptr)
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					b = encodeNull(b)
					b = encodeComma(b)
					code = code.end
					break
				}
				bytes, err := rv.Interface().(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadMarshalText:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldAnonymousHeadMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				ptr += code.offset
				v := e.ptrToInterface(code, ptr)
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					b = encodeNull(b)
					b = encodeComma(b)
					code = code.end.next
					break
				}
				bytes, err := rv.Interface().(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				p := ptr + code.offset
				if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				p := ptr + code.offset
				if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyIntOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyIntOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{')
			v := e.ptrToInt(ptr + code.offset)
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendInt(b, v)
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, uint64(v))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, uint64(v))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, uint64(v))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, uint64(v))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, uint64(v))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, uint64(v))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, uint64(v))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, uint64(v))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = appendUint(b, v)
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToFloat32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = encodeFloat32(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = encodeFloat32(b, v)
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToFloat64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = append(b, code.escapedKey...)
					b = encodeFloat64(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = append(b, code.escapedKey...)
					b = encodeFloat64(b, v)
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToString(ptr + code.offset)
				if v == "" {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = encodeEscapedString(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToString(ptr + code.offset)
				if v == "" {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = encodeEscapedString(b, v)
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToBool(ptr + code.offset)
				if !v {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = encodeBool(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToBool(ptr + code.offset)
				if !v {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = encodeBool(b, v)
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToBytes(ptr + code.offset)
				if len(v) == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = encodeByteSlice(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToBytes(ptr + code.offset)
				if len(v) == 0 {
					code = code.nextField
				} else {
					b = append(b, code.escapedKey...)
					b = encodeByteSlice(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = code.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bb, err := v.(Marshaler).MarshalJSON()
					if err != nil {
						return nil, &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					if len(bb) == 0 {
						if isPtr {
							return nil, errUnexpectedEndOfJSON(
								fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
								0,
							)
						}
						code = code.nextField
					} else {
						var buf bytes.Buffer
						if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
							return nil, err
						}
						b = append(b, code.escapedKey...)
						b = append(b, buf.Bytes()...)
						b = encodeComma(b)
						code = code.next
					}
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = code.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bb, err := v.(Marshaler).MarshalJSON()
					if err != nil {
						return nil, &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					if len(bb) == 0 {
						if isPtr {
							return nil, errUnexpectedEndOfJSON(
								fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
								0,
							)
						}
						code = code.nextField
					} else {
						var buf bytes.Buffer
						if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
							return nil, err
						}
						b = append(b, code.escapedKey...)
						b = append(b, buf.Bytes()...)
						b = encodeComma(b)
						code = code.next
					}
				}
			}
		case opStructEscapedFieldPtrHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = code.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bytes, err := v.(encoding.TextMarshaler).MarshalText()
					if err != nil {
						return nil, &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					b = append(b, code.escapedKey...)
					b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = code.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bytes, err := v.(encoding.TextMarshaler).MarshalText()
					if err != nil {
						return nil, &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					b = append(b, code.escapedKey...)
					b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
					b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				p := ptr + code.offset
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, ptr+code.offset)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagInt64Only:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt64Only:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadStringTagInt64PtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagInt64PtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, e.ptrToInt64(p+code.offset))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = encodeFloat64(b, v)
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = encodeFloat64(b, v)
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				s := e.ptrToString(ptr + code.offset)
				b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagEscapedString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				s := string(encodeEscapedString([]byte{}, e.ptrToString(ptr+code.offset)))
				b = encodeEscapedString(b, s)
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagBoolOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagBoolOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrAnonymousHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = append(b, '"')
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
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
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				if len(bb) == 0 {
					if isPtr {
						return nil, errUnexpectedEndOfJSON(
							fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
							0,
						)
					}
					b = append(b, code.escapedKey...)
					b = append(b, '"', '"')
					b = encodeComma(b)
					code = code.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
						return nil, err
					}
					b = append(b, code.escapedKey...)
					b = encodeEscapedString(b, buf.String())
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				if len(bb) == 0 {
					if isPtr {
						return nil, errUnexpectedEndOfJSON(
							fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
							0,
						)
					}
					b = append(b, code.escapedKey...)
					b = append(b, '"', '"')
					b = encodeComma(b)
					code = code.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
						return nil, err
					}
					b = append(b, code.escapedKey...)
					b = encodeEscapedString(b, buf.String())
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructEscapedFieldPtrHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrAnonymousHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedField:
			if !code.anonymousKey {
				b = append(b, code.escapedKey...)
			}
			ptr := load(ctxptr, code.headIdx) + code.offset
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructEscapedFieldIntPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldIntNPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			for i := 0; i < code.ptrNum-1; i++ {
				if p == 0 {
					break
				}
				p = e.ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt8Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt16Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt32Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt64Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUintPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint8Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint16Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint32Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint64Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldFloat32Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldFloat64Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			v := e.ptrToFloat64(p)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldFloat64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldEscapedStringPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, e.ptrToString(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldBoolPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, e.ptrToBool(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldBytes:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
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
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldArray:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldSlice:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldMap:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldMapLoad:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldStruct:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldOmitEmpty:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 || **(**uintptr)(unsafe.Pointer(&p)) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructEscapedFieldOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyUint64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.escapedKey...)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyEscapedString:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = append(b, code.escapedKey...)
				b = encodeBool(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			if v != nil {
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				var buf bytes.Buffer
				if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
					return nil, err
				}
				b = append(b, code.escapedKey...)
				b = append(b, buf.Bytes()...)
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			if v != nil {
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyArray:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			array := e.ptrToSlice(p)
			if p == 0 || uintptr(array.data) == 0 {
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructEscapedFieldOmitEmptySlice:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				code = code.nextField
			} else {
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
					code = code.next
				}
			}
		case opStructEscapedFieldStringTag:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = append(b, code.escapedKey...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint64(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			s := e.ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagMarshalJSON:
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
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, buf.String())
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedEnd:
			last := len(b) - 1
			if b[last] == ',' {
				b[last] = '}'
			} else {
				b = append(b, '}')
			}
			b = encodeComma(b)
			code = code.next
		case opStructAnonymousEnd:
			code = code.next
		case opStructEscapedEndIntPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndIntNPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			for i := 0; i < code.ptrNum-1; i++ {
				if p == 0 {
					break
				}
				p = e.ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt8Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt16Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt32Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt64Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUintPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint8Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint16Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint32Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint64Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndFloat32Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndFloat64Ptr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
				b = appendStructEnd(b)
				code = code.next
				break
			}
			v := e.ptrToFloat64(p)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndFloat64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndEscapedStringPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, e.ptrToString(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndBoolPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, e.ptrToBool(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndBytes:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
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
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyInt8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyInt16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyInt32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyInt64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, v)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(v))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyUint8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(v))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyUint16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(v))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyUint32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(v))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyUint64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, v)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, v)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.escapedKey...)
				b = encodeFloat64(b, v)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyEscapedString:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, v)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = append(b, code.escapedKey...)
				b = encodeBool(b, v)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, v)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			if v != nil {
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				var buf bytes.Buffer
				if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
					return nil, err
				}
				b = append(b, code.escapedKey...)
				b = append(b, buf.Bytes()...)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndOmitEmptyMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			if v != nil {
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint64(ptr+code.offset)))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			s := e.ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagMarshalJSON:
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
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, buf.String())
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = appendStructEnd(b)
			code = code.next
		case opEnd:
			goto END
		}
	}
END:
	return b, nil
}

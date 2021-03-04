package json

import (
	"bytes"
	"encoding"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"unsafe"
)

const startDetectingCyclesAfter = 1000

func load(base uintptr, idx uintptr) uintptr {
	addr := base + idx
	return **(**uintptr)(unsafe.Pointer(&addr))
}

func store(base uintptr, idx uintptr, p uintptr) {
	addr := base + idx
	**(**uintptr)(unsafe.Pointer(&addr)) = p
}

func ptrToUint64(p uintptr) uint64      { return **(**uint64)(unsafe.Pointer(&p)) }
func ptrToFloat32(p uintptr) float32    { return **(**float32)(unsafe.Pointer(&p)) }
func ptrToFloat64(p uintptr) float64    { return **(**float64)(unsafe.Pointer(&p)) }
func ptrToBool(p uintptr) bool          { return **(**bool)(unsafe.Pointer(&p)) }
func ptrToBytes(p uintptr) []byte       { return **(**[]byte)(unsafe.Pointer(&p)) }
func ptrToString(p uintptr) string      { return **(**string)(unsafe.Pointer(&p)) }
func ptrToSlice(p uintptr) *sliceHeader { return *(**sliceHeader)(unsafe.Pointer(&p)) }
func ptrToPtr(p uintptr) uintptr {
	return uintptr(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
}
func ptrToUnsafePtr(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}
func ptrToInterface(code *opcode, p uintptr) interface{} {
	return *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: code.typ,
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&p)),
	}))
}

func errUnsupportedValue(code *opcode, ptr uintptr) *UnsupportedValueError {
	v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: code.typ,
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&ptr)),
	}))
	return &UnsupportedValueError{
		Value: reflect.ValueOf(v),
		Str:   fmt.Sprintf("encountered a cycle via %s", code.typ),
	}
}

func errUnsupportedFloat(v float64) *UnsupportedValueError {
	return &UnsupportedValueError{
		Value: reflect.ValueOf(v),
		Str:   strconv.FormatFloat(v, 'g', -1, 64),
	}
}

func errMarshaler(code *opcode, err error) *MarshalerError {
	return &MarshalerError{
		Type: rtype2type(code.typ),
		Err:  err,
	}
}

func encodeRun(ctx *encodeRuntimeContext, b []byte, codeSet *opcodeSet, opt EncodeOption) ([]byte, error) {
	recursiveLevel := 0
	ptrOffset := uintptr(0)
	ctxptr := ctx.ptr()
	code := codeSet.code

	for {
		switch code.op {
		default:
			return nil, fmt.Errorf("encoder: opcode %s has not been implemented", code.op)
		case opPtr:
			ptr := load(ctxptr, code.idx)
			code = code.next
			store(ctxptr, code.idx, ptrToPtr(ptr))
		case opInt:
			b = appendInt(b, ptrToUint64(load(ctxptr, code.idx)), code)
			b = encodeComma(b)
			code = code.next
		case opUint:
			b = appendUint(b, ptrToUint64(load(ctxptr, code.idx)), code)
			b = encodeComma(b)
			code = code.next
		case opIntString:
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(load(ctxptr, code.idx)), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUintString:
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(load(ctxptr, code.idx)), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opFloat32:
			b = encodeFloat32(b, ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opFloat64:
			v := ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opString:
			b = encodeNoEscapedString(b, ptrToString(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opBool:
			b = encodeBool(b, ptrToBool(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opBytes:
			ptr := load(ctxptr, code.idx)
			slice := ptrToSlice(ptr)
			if ptr == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(ptr))
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
			for _, seen := range ctx.seenPtr {
				if ptr == seen {
					return nil, errUnsupportedValue(code, ptr)
				}
			}
			ctx.seenPtr = append(ctx.seenPtr, ptr)
			iface := (*interfaceHeader)(ptrToUnsafePtr(ptr))
			if iface == nil || iface.ptr == nil {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			ctx.keepRefs = append(ctx.keepRefs, unsafe.Pointer(iface))
			ifaceCodeSet, err := encodeCompileToGetCodeSet(uintptr(unsafe.Pointer(iface.typ)))
			if err != nil {
				return nil, err
			}

			totalLength := uintptr(codeSet.codeLength)
			nextTotalLength := uintptr(ifaceCodeSet.codeLength)

			curlen := uintptr(len(ctx.ptrs))
			offsetNum := ptrOffset / uintptrSize

			newLen := offsetNum + totalLength + nextTotalLength
			if curlen < newLen {
				ctx.ptrs = append(ctx.ptrs, make([]uintptr, newLen-curlen)...)
			}
			oldPtrs := ctx.ptrs

			newPtrs := ctx.ptrs[(ptrOffset+totalLength*uintptrSize)/uintptrSize:]
			newPtrs[0] = uintptr(iface.ptr)

			ctx.ptrs = newPtrs

			bb, err := encodeRun(ctx, b, ifaceCodeSet, opt)
			if err != nil {
				return nil, err
			}

			ctx.ptrs = oldPtrs
			ctxptr = ctx.ptr()
			ctx.seenPtr = ctx.seenPtr[:len(ctx.seenPtr)-1]

			b = bb
			code = code.next
		case opMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			v := ptrToInterface(code, ptr)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			runtime.KeepAlive(v)
			if len(bb) == 0 {
				return nil, errUnexpectedEndOfJSON(
					fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
					0,
				)
			}
			buf := bytes.NewBuffer(b)
			//TODO: we should validate buffer with `compact`
			if err := compact(buf, bb, false); err != nil {
				return nil, err
			}
			b = buf.Bytes()
			b = encodeComma(b)
			code = code.next
		case opMarshalText:
			ptr := load(ctxptr, code.idx)
			isPtr := code.typ.Kind() == reflect.Ptr
			p := ptrToUnsafePtr(ptr)
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
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
			}
			code = code.next
		case opSliceHead:
			p := load(ctxptr, code.idx)
			slice := ptrToSlice(p)
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
				uptr := ptrToUnsafePtr(ptr)
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{')
					iter := mapiterinit(code.typ, uptr)
					ctx.keepRefs = append(ctx.keepRefs, iter)
					store(ctxptr, code.elemIdx, 0)
					store(ctxptr, code.length, uintptr(mlen))
					store(ctxptr, code.mapIter, uintptr(iter))
					if (opt & EncodeOptionUnorderedMap) == 0 {
						mapCtx := newMapContext(mlen)
						mapCtx.pos = append(mapCtx.pos, len(b))
						ctx.keepRefs = append(ctx.keepRefs, unsafe.Pointer(mapCtx))
						store(ctxptr, code.end.mapPos, uintptr(unsafe.Pointer(mapCtx)))
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
				ptr = ptrToPtr(ptr)
				uptr := ptrToUnsafePtr(ptr)
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
					if (opt & EncodeOptionUnorderedMap) == 0 {
						mapCtx := newMapContext(mlen)
						mapCtx.pos = append(mapCtx.pos, len(b))
						ctx.keepRefs = append(ctx.keepRefs, unsafe.Pointer(mapCtx))
						store(ctxptr, code.end.mapPos, uintptr(unsafe.Pointer(mapCtx)))
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
			if (opt & EncodeOptionUnorderedMap) != 0 {
				if idx < length {
					ptr := load(ctxptr, code.mapIter)
					iter := ptrToUnsafePtr(ptr)
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
				mapCtx := (*encodeMapContext)(ptrToUnsafePtr(ptr))
				mapCtx.pos = append(mapCtx.pos, len(b))
				if idx < length {
					ptr := load(ctxptr, code.mapIter)
					iter := ptrToUnsafePtr(ptr)
					store(ctxptr, code.elemIdx, idx)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					code = code.end
				}
			}
		case opMapValue:
			if (opt & EncodeOptionUnorderedMap) != 0 {
				last := len(b) - 1
				b[last] = ':'
			} else {
				ptr := load(ctxptr, code.end.mapPos)
				mapCtx := (*encodeMapContext)(ptrToUnsafePtr(ptr))
				mapCtx.pos = append(mapCtx.pos, len(b))
			}
			ptr := load(ctxptr, code.mapIter)
			iter := ptrToUnsafePtr(ptr)
			value := mapitervalue(iter)
			store(ctxptr, code.next.idx, uintptr(value))
			mapiternext(iter)
			code = code.next
		case opMapEnd:
			// this operation only used by sorted map.
			length := int(load(ctxptr, code.length))
			ptr := load(ctxptr, code.mapPos)
			mapCtx := (*encodeMapContext)(ptrToUnsafePtr(ptr))
			pos := mapCtx.pos
			for i := 0; i < length; i++ {
				startKey := pos[i*2]
				startValue := pos[i*2+1]
				var endValue int
				if i+1 < length {
					endValue = pos[i*2+2]
				} else {
					endValue = len(b)
				}
				mapCtx.slice.items = append(mapCtx.slice.items, mapItem{
					key:   b[startKey:startValue],
					value: b[startValue:endValue],
				})
			}
			sort.Sort(mapCtx.slice)
			buf := mapCtx.buf
			for _, item := range mapCtx.slice.items {
				buf = append(buf, item.key...)
				buf[len(buf)-1] = ':'
				buf = append(buf, item.value...)
			}
			buf[len(buf)-1] = '}'
			buf = append(buf, ',')
			b = b[:pos[0]]
			b = append(b, buf...)
			mapCtx.buf = buf
			releaseMapContext(mapCtx)
			code = code.next
		case opStructFieldPtrAnonymousHeadRecursive:
			store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadRecursive:
			fallthrough
		case opStructFieldRecursive:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				if recursiveLevel > startDetectingCyclesAfter {
					for _, seen := range ctx.seenPtr {
						if ptr == seen {
							return nil, errUnsupportedValue(code, ptr)
						}
					}
				}
			}
			ctx.seenPtr = append(ctx.seenPtr, ptr)
			c := code.jmp.code
			curlen := uintptr(len(ctx.ptrs))
			offsetNum := ptrOffset / uintptrSize
			oldOffset := ptrOffset
			ptrOffset += code.jmp.curLen * uintptrSize

			newLen := offsetNum + code.jmp.curLen + code.jmp.nextLen
			if curlen < newLen {
				ctx.ptrs = append(ctx.ptrs, make([]uintptr, newLen-curlen)...)
			}
			ctxptr = ctx.ptr() + ptrOffset // assign new ctxptr

			store(ctxptr, c.idx, ptr)
			store(ctxptr, c.end.next.idx, oldOffset)
			store(ctxptr, c.end.next.elemIdx, uintptr(unsafe.Pointer(code.next)))
			code = c
			recursiveLevel++
		case opStructFieldRecursiveEnd:
			recursiveLevel--

			// restore ctxptr
			offset := load(ctxptr, code.idx)
			ctx.seenPtr = ctx.seenPtr[:len(ctx.seenPtr)-1]

			codePtr := load(ctxptr, code.elemIdx)
			code = (*opcode)(ptrToUnsafePtr(codePtr))
			ctxptr = ctx.ptr() + offset
			ptrOffset = offset
		case opStructFieldPtrHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHead:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if !code.anonymousKey {
				b = append(b, code.key...)
			}
			p += code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadOmitEmpty:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmpty:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			p += code.offset
			if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadStringTag:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTag:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			p += code.offset
			b = append(b, code.key...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrAnonymousHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHead:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			p += code.offset
			b = append(b, code.key...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrAnonymousHeadOmitEmpty:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmpty:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			p += code.offset
			if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrAnonymousHeadStringTag:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringTag:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			p += code.offset
			b = append(b, code.key...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadInt:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = appendInt(b, ptrToUint64(p+code.offset), code)
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyInt:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			u64 := ptrToUint64(p + code.offset)
			v := u64 & code.mask
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = appendInt(b, u64, code)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(p+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.key...)
				b = appendInt(b, ptrToUint64(p), code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldHeadIntNPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, ptrToUint64(p), code)
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadInt:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = appendInt(b, ptrToUint64(p+code.offset), code)
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyInt:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			u64 := ptrToUint64(p + code.offset)
			v := u64 & code.mask
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = appendInt(b, u64, code)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagInt:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(p+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = appendInt(b, ptrToUint64(p), code)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringTagIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadUint:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = appendUint(b, ptrToUint64(p+code.offset), code)
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyUint:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			u64 := ptrToUint64(p + code.offset)
			v := u64 & code.mask
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = appendUint(b, u64, code)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(p+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.key...)
				b = appendUint(b, ptrToUint64(p), code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldHeadUintNPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendUint(b, ptrToUint64(p), code)
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadUint:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = appendUint(b, ptrToUint64(p+code.offset), code)
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyUint:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			u64 := ptrToUint64(p + code.offset)
			v := u64 & code.mask
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = appendUint(b, u64, code)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagUint:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(p+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = appendUint(b, ptrToUint64(p), code)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringTagUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat32:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = encodeFloat32(b, ptrToFloat32(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyFloat32:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			v := ptrToFloat32(p + code.offset)
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeFloat32(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagFloat32:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat32(b, ptrToFloat32(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, ptrToFloat32(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.key...)
				b = encodeFloat32(b, ptrToFloat32(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldHeadFloat32NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = encodeFloat32(b, ptrToFloat32(p+code.offset))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadFloat32:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = encodeFloat32(b, ptrToFloat32(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat32:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			v := ptrToFloat32(p + code.offset)
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeFloat32(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagFloat32:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat32(b, ptrToFloat32(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, ptrToFloat32(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeFloat32(b, ptrToFloat32(p))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat64:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyFloat64:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			v := ptrToFloat64(p + code.offset)
			if v == 0 {
				code = code.nextField
			} else {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.key...)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagFloat64:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.key...)
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldHeadFloat64NPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					v := ptrToFloat64(p)
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = encodeFloat64(b, v)
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadFloat64:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.key...)
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat64:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			v := ptrToFloat64(p + code.offset)
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagFloat64:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = append(b, '"')
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadString:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, ptrToString(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyString:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			v := ptrToString(p + code.offset)
			if v == "" {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagString:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			s := ptrToString(p + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, ptrToString(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, ptrToString(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, ptrToString(p))))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldHeadStringNPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = encodeNoEscapedString(b, ptrToString(p+code.offset))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadString:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, ptrToString(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyString:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			v := ptrToString(p + code.offset)
			if v == "" {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagString:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, ptrToString(p+code.offset))))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, ptrToString(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, ptrToString(p))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringTagStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, ptrToString(p))))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadBool:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = encodeBool(b, ptrToBool(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyBool:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			v := ptrToBool(p + code.offset)
			if v {
				b = append(b, code.key...)
				b = encodeBool(b, v)
				b = encodeComma(b)
				code = code.next
			} else {
				code = code.nextField
			}
		case opStructFieldPtrHeadStringTagBool:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeBool(b, ptrToBool(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, ptrToBool(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.key...)
				b = encodeBool(b, ptrToBool(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldHeadBoolNPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = ptrToPtr(p)
				}
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = encodeBool(b, ptrToBool(p+code.offset))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadBool:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = encodeBool(b, ptrToBool(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyBool:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			v := ptrToBool(p + code.offset)
			if v {
				b = append(b, code.key...)
				b = encodeBool(b, v)
				b = encodeComma(b)
				code = code.next
			} else {
				code = code.nextField
			}
		case opStructFieldPtrAnonymousHeadStringTagBool:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeBool(b, ptrToBool(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, ptrToBool(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeBool(b, ptrToBool(p))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringTagBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadBytes:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = encodeByteSlice(b, ptrToBytes(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyBytes:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			v := ptrToBytes(p + code.offset)
			if v == nil {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeByteSlice(b, ptrToBytes(p))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagBytes:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeByteSlice(b, ptrToBytes(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.key...)
				b = encodeByteSlice(b, ptrToBytes(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeByteSlice(b, ptrToBytes(p))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadBytes:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = encodeByteSlice(b, ptrToBytes(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyBytes:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			v := ptrToBytes(p + code.offset)
			if v == nil {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeByteSlice(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagBytes:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeByteSlice(b, ptrToBytes(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				b = encodeByteSlice(b, ptrToBytes(p))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadStringTagBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeByteSlice(b, ptrToBytes(p))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadArray, opStructFieldPtrHeadStringTagArray,
			opStructFieldPtrHeadSlice, opStructFieldPtrHeadStringTagSlice:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadArray, opStructFieldHeadStringTagArray,
			opStructFieldHeadSlice, opStructFieldHeadStringTagSlice:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			p += code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadOmitEmptyArray:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyArray:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			p += code.offset
			b = append(b, code.key...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadOmitEmptySlice:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptySlice:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			p += code.offset
			slice := ptrToSlice(p)
			if slice.data == nil {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadArrayPtr, opStructFieldPtrHeadStringTagArrayPtr,
			opStructFieldPtrHeadSlicePtr, opStructFieldPtrHeadStringTagSlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadArrayPtr, opStructFieldHeadStringTagArrayPtr,
			opStructFieldHeadSlicePtr, opStructFieldHeadStringTagSlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.nextField
			} else {
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadOmitEmptyArrayPtr, opStructFieldPtrHeadOmitEmptySlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyArrayPtr, opStructFieldHeadOmitEmptySlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrAnonymousHeadArray, opStructFieldPtrAnonymousHeadStringTagArray,
			opStructFieldPtrAnonymousHeadSlice, opStructFieldPtrAnonymousHeadStringTagSlice:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadArray, opStructFieldAnonymousHeadStringTagArray,
			opStructFieldAnonymousHeadSlice, opStructFieldAnonymousHeadStringTagSlice:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			p += code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrAnonymousHeadOmitEmptyArray:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyArray:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrAnonymousHeadOmitEmptySlice:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptySlice:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			slice := ptrToSlice(p + code.offset)
			if slice.data == nil {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrAnonymousHeadArrayPtr, opStructFieldPtrAnonymousHeadStringTagArrayPtr,
			opStructFieldPtrAnonymousHeadSlicePtr, opStructFieldPtrAnonymousHeadStringTagSlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadArrayPtr, opStructFieldAnonymousHeadStringTagArrayPtr,
			opStructFieldAnonymousHeadSlicePtr, opStructFieldAnonymousHeadStringTagSlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.nextField
			} else {
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyArrayPtr,
			opStructFieldPtrAnonymousHeadOmitEmptySlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyArrayPtr,
			opStructFieldAnonymousHeadOmitEmptySlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMap, opStructFieldPtrHeadStringTagMap:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMap, opStructFieldHeadStringTagMap:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if p != 0 && code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadOmitEmptyMap:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyMap:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if p != 0 && code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if maplen(ptrToUnsafePtr(p)) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMapPtr, opStructFieldPtrHeadStringTagMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMapPtr, opStructFieldHeadStringTagMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.nextField
				break
			}
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.nextField
			} else {
				if code.indirect {
					p = ptrToPtr(p)
				}
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadOmitEmptyMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			if p == 0 {
				code = code.nextField
				break
			}
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				code = code.nextField
			} else {
				if code.indirect {
					p = ptrToPtr(p)
				}
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrAnonymousHeadMap, opStructFieldPtrAnonymousHeadStringTagMap:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadMap, opStructFieldAnonymousHeadStringTagMap:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if p != 0 && code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrAnonymousHeadOmitEmptyMap:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMap:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			if p != 0 && code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if maplen(ptrToUnsafePtr(p)) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrAnonymousHeadMapPtr, opStructFieldPtrAnonymousHeadStringTagMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadMapPtr, opStructFieldAnonymousHeadStringTagMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			if p != 0 {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 && code.indirect {
				p = ptrToPtr(p)
			}
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrAnonymousHeadOmitEmptyMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			if p == 0 {
				code = code.end.next
				break
			}
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				code = code.nextField
			} else {
				if code.indirect {
					p = ptrToPtr(p)
				}
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			bb, err := encodeMarshalJSON(b, ptrToInterface(code, p+code.offset))
			if err != nil {
				return nil, err
			}
			b = bb
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			p := ptrToUnsafePtr(ptr + code.offset)
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
					b = append(b, code.key...)
					buf := bytes.NewBuffer(b)
					//TODO: we should validate buffer with `compact`
					if err := compact(buf, bb, false); err != nil {
						return nil, err
					}
					b = buf.Bytes()
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			bb, err := encodeMarshalJSON(b, ptrToInterface(code, p+code.offset))
			if err != nil {
				return nil, err
			}
			b = bb
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			b = append(b, '{')
			b = append(b, code.key...)
			bb, err := encodeMarshalText(b, ptrToInterface(code, p+code.offset))
			if err != nil {
				return nil, err
			}
			b = bb
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldAnonymousHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				code = code.end.next
				break
			}
			b = append(b, code.key...)
			bb, err := encodeMarshalText(b, ptrToInterface(code, p+code.offset))
			if err != nil {
				return nil, err
			}
			b = bb
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := ptrToUnsafePtr(ptr)
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
						b = append(b, code.key...)
						buf := bytes.NewBuffer(b)
						//TODO: we should validate buffer with `compact`
						if err := compact(buf, bb, false); err != nil {
							return nil, err
						}
						b = buf.Bytes()
						b = encodeComma(b)
						code = code.next
					}
				}
			}
		case opStructFieldPtrHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := ptrToUnsafePtr(ptr)
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
					b = append(b, code.key...)
					b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := ptrToUnsafePtr(ptr)
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
					b = append(b, code.key...)
					b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := ptrToUnsafePtr(ptr)
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
					b = append(b, code.key...)
					b = append(b, '"', '"')
					b = encodeComma(b)
					code = code.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, bb, false); err != nil {
						return nil, err
					}
					b = append(b, code.key...)
					b = encodeNoEscapedString(b, buf.String())
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := ptrToUnsafePtr(ptr)
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
					b = append(b, code.key...)
					b = append(b, '"', '"')
					b = encodeComma(b)
					code = code.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, bb, false); err != nil {
						return nil, err
					}
					b = append(b, code.key...)
					b = encodeNoEscapedString(b, buf.String())
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := ptrToUnsafePtr(ptr)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := ptrToUnsafePtr(ptr)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructField:
			if !code.anonymousKey {
				b = append(b, code.key...)
			}
			ptr := load(ctxptr, code.headIdx) + code.offset
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmpty:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 || **(**uintptr)(unsafe.Pointer(&p)) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldStringTag:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = append(b, code.key...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = append(b, code.key...)
				b = appendInt(b, u64, code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldIntPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyIntPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = appendInt(b, ptrToUint64(p), code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagIntPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = append(b, code.key...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldIntNPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			for i := 0; i < code.ptrNum-1; i++ {
				if p == 0 {
					break
				}
				p = ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = append(b, code.key...)
				b = appendUint(b, u64, code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldUintPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyUintPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = appendUint(b, ptrToUint64(p), code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagUintPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = append(b, code.key...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = encodeFloat32(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, ptrToFloat32(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat32Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = encodeFloat32(b, ptrToFloat32(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat32Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = append(b, code.key...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			v := ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.key...)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat64Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			v := ptrToFloat64(p)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat64Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat64Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = append(b, code.key...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, '"')
				b = encodeFloat64(b, v)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, ptrToString(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToString(ptr + code.offset)
			if v != "" {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			s := ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, ptrToString(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyStringPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, ptrToString(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagStringPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, ptrToString(p))))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBool(ptr + code.offset)
			if v {
				b = append(b, code.key...)
				b = encodeBool(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldBoolPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, ptrToBool(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyBoolPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = encodeBool(b, ptrToBool(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagBoolPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldBytes:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeByteSlice(b, ptrToBytes(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = append(b, code.key...)
				b = encodeByteSlice(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			b = append(b, code.key...)
			b = encodeByteSlice(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldBytesPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyBytesPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = encodeByteSlice(b, ptrToBytes(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagBytesPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			buf := bytes.NewBuffer(b)
			//TODO: we should validate buffer with `compact`
			if err := compact(buf, bb, false); err != nil {
				return nil, err
			}
			b = buf.Bytes()
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, false); err != nil {
				return nil, err
			}
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, buf.String())
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if code.typ.Kind() == reflect.Ptr && code.typ.Elem().Implements(marshalJSONType) {
				p = ptrToPtr(p)
			}
			v := ptrToInterface(code, p)
			if v != nil && p != 0 {
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.key...)
				buf := bytes.NewBuffer(b)
				//TODO: we should validate buffer with `compact`
				if err := compact(buf, bb, false); err != nil {
					return nil, err
				}
				b = buf.Bytes()
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			if v != nil {
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldArray, opStructFieldStringTagArray:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyArray:
			p := load(ctxptr, code.headIdx)
			p += code.offset
			b = append(b, code.key...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldArrayPtr, opStructFieldStringTagArrayPtr:
			b = append(b, code.key...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyArrayPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldSlice, opStructFieldStringTagSlice:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptySlice:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldSlicePtr, opStructFieldStringTagSlicePtr:
			b = append(b, code.key...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptySlicePtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldMap, opStructFieldStringTagMap:
			b = append(b, code.key...)
			p := load(ctxptr, code.headIdx)
			if p != 0 {
				p = ptrToPtr(p + code.offset)
			}
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyMap:
			p := load(ctxptr, code.headIdx)
			if p == 0 {
				code = code.nextField
				break
			}
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldMapPtr, opStructFieldStringTagMapPtr:
			b = append(b, code.key...)
			p := load(ctxptr, code.headIdx)
			if p != 0 {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				p = ptrToPtr(p)
			}
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyMapPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				p = ptrToPtr(p)
			}
			if p != 0 {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldMapLoad:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyMapLoad:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				mlen := maplen(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
				if mlen == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructFieldStruct:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEnd:
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
		case opStructEndInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = append(b, code.key...)
				b = appendInt(b, u64, code)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndIntPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyIntPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = appendInt(b, ptrToUint64(p), code)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagIntPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndIntNPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			for i := 0; i < code.ptrNum-1; i++ {
				if p == 0 {
					break
				}
				p = ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = append(b, code.key...)
				b = appendUint(b, u64, code)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUintPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyUintPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = appendUint(b, ptrToUint64(p), code)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagUintPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUintNPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			for i := 0; i < code.ptrNum-1; i++ {
				if p == 0 {
					break
				}
				p = ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = encodeFloat32(b, v)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, ptrToFloat32(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyFloat32Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = encodeFloat32(b, ptrToFloat32(p))
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagFloat32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat32NPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			for i := 0; i < code.ptrNum-1; i++ {
				if p == 0 {
					break
				}
				p = ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, ptrToFloat32(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			v := ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.key...)
				b = encodeFloat64(b, v)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat64Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
				b = appendStructEnd(b)
				code = code.next
				break
			}
			v := ptrToFloat64(p)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyFloat64Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagFloat64Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = append(b, '"')
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat64NPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			for i := 0; i < code.ptrNum-1; i++ {
				if p == 0 {
					break
				}
				p = ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, ptrToString(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToString(ptr + code.offset)
			if v != "" {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, v)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			s := ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndStringPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, ptrToString(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyStringPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, ptrToString(p))
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagStringPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := ptrToString(p)
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, v)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndStringNPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			for i := 0; i < code.ptrNum-1; i++ {
				if p == 0 {
					break
				}
				p = ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, ptrToString(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBool(ptr + code.offset)
			if v {
				b = append(b, code.key...)
				b = encodeBool(b, v)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndBoolPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, ptrToBool(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyBoolPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.key...)
				b = encodeBool(b, ptrToBool(p))
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagBoolPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndBytes:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeByteSlice(b, ptrToBytes(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = append(b, code.key...)
				b = encodeByteSlice(b, v)
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			b = append(b, code.key...)
			b = encodeByteSlice(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			buf := bytes.NewBuffer(b)
			//TODO: we should validate buffer with `compact`
			if err := compact(buf, bb, false); err != nil {
				return nil, err
			}
			b = buf.Bytes()
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			if v != nil && (code.typ.Kind() != reflect.Ptr || ptrToPtr(p) != 0) {
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.key...)
				buf := bytes.NewBuffer(b)
				//TODO: we should validate buffer with `compact`
				if err := compact(buf, bb, false); err != nil {
					return nil, err
				}
				b = buf.Bytes()
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, false); err != nil {
				return nil, err
			}
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, buf.String())
			b = appendStructEnd(b)
			code = code.next
		case opStructEndMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			if v != nil {
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = appendStructEnd(b)
			} else {
				last := len(b) - 1
				if b[last] == ',' {
					b[last] = '}'
					b = encodeComma(b)
				} else {
					b = appendStructEnd(b)
				}
			}
			code = code.next
		case opStructEndStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = appendStructEnd(b)
			code = code.next
		case opEnd:
			goto END
		}
	}
END:
	return b, nil
}

package json

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

func encodeRunEscaped(ctx *encodeRuntimeContext, b []byte, codeSet *opcodeSet, opt EncodeOption) ([]byte, error) {
	recursiveLevel := 0
	ptrOffset := uintptr(0)
	ctxptr := ctx.ptr()
	code := codeSet.code

	for {
		switch code.op {
		default:
			return nil, fmt.Errorf("encoder (escaped): opcode %s has not been implemented", code.op)
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
			b = encodeEscapedString(b, ptrToString(load(ctxptr, code.idx)))
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
		case opNumber:
			bb, err := encodeNumber(b, ptrToNumber(load(ctxptr, code.idx)))
			if err != nil {
				return nil, err
			}
			b = encodeComma(bb)
			code = code.next
		case opInterfacePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
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

			bb, err := encodeRunEscaped(ctx, b, ifaceCodeSet, opt)
			if err != nil {
				return nil, err
			}

			ctx.ptrs = oldPtrs
			ctxptr = ctx.ptr()
			ctx.seenPtr = ctx.seenPtr[:len(ctx.seenPtr)-1]

			b = bb
			code = code.next
		case opMarshalJSONPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
			if err != nil {
				return nil, err
			}
			b = encodeComma(bb)
			code = code.next
		case opMarshalTextPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = append(b, `""`...)
				b = encodeComma(b)
				code = code.next
				break
			}
			bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
			if err != nil {
				return nil, err
			}
			b = encodeComma(bb)
			code = code.next
		case opSlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opSlice:
			p := load(ctxptr, code.idx)
			slice := ptrToSlice(p)
			if p == 0 || slice.data == nil {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
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
		case opArray:
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
		case opMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opMap:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			uptr := ptrToUnsafePtr(p)
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
		case opStructFieldRecursivePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHead:
			p := load(ctxptr, code.idx)
			if p == 0 && (code.indirect || code.next.op == opStructEnd) {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if !code.anonymousKey {
				b = append(b, code.escapedKey...)
			}
			p += code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadOmitEmpty:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmpty:
			p := load(ctxptr, code.idx)
			if p == 0 && (code.indirect || code.next.op == opStructEnd) {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			p += code.offset
			if p == 0 || (strings.Contains(code.next.op.String(), "Ptr") && ptrToPtr(p) == 0) {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadStringTag:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTag:
			p := load(ctxptr, code.idx)
			if p == 0 && (code.indirect || code.next.op == opStructEnd) {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			p += code.offset
			b = append(b, code.escapedKey...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadInt:
			if code.indirect {
				p := load(ctxptr, code.idx)
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = ptrToPtr(p)
				}
				store(ctxptr, code.idx, p)
			}
			fallthrough
		case opStructFieldHeadInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			u64 := ptrToUint64(p + code.offset)
			v := u64 & code.mask
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(p+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, ptrToUint64(p), code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			u64 := ptrToUint64(p + code.offset)
			v := u64 & code.mask
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(p+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, ptrToUint64(p), code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			v := ptrToFloat32(p + code.offset)
			if v == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat32(b, ptrToFloat32(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, ptrToFloat32(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			v := ptrToFloat64(p + code.offset)
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
		case opStructFieldPtrHeadStringTagFloat64:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				p := load(ctxptr, code.idx)
				for i := 0; i < code.ptrNum; i++ {
					if p == 0 {
						break
					}
					p = ptrToPtr(p)
				}
				store(ctxptr, code.idx, p)
			}
			fallthrough
		case opStructFieldHeadString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, ptrToString(p+code.offset))
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			v := ptrToString(p + code.offset)
			if v == "" {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, v)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			s := ptrToString(p + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, ptrToString(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, ptrToString(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, ptrToString(p))))
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			v := ptrToBool(p + code.offset)
			if v {
				b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeBool(b, ptrToBool(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.escapedKey...)
				b = encodeBool(b, ptrToBool(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			v := ptrToBytes(p + code.offset)
			if v == nil {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeByteSlice(b, ptrToBytes(p+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, ptrToBytes(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
		case opStructFieldHeadNumber:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = encodeComma(bb)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyNumber:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyNumber:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			v := ptrToNumber(p + code.offset)
			if v == "" {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				bb, err := encodeNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = encodeComma(bb)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagNumber:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadStringTagNumber:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.escapedKey...)
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = encodeComma(bb)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadStringTagNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = append(bb, '"')
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			p += code.offset
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			p += code.offset
			slice := ptrToSlice(p)
			if slice.data == nil {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadArrayPtr, opStructFieldPtrHeadStringTagArrayPtr,
			opStructFieldPtrHeadSlicePtr, opStructFieldPtrHeadStringTagSlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadArrayPtr, opStructFieldHeadStringTagArrayPtr,
			opStructFieldHeadSlicePtr, opStructFieldHeadStringTagSlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyArrayPtr, opStructFieldHeadOmitEmptySlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMap, opStructFieldPtrHeadStringTagMap:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMap, opStructFieldHeadStringTagMap:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if p != 0 && code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadOmitEmptyMap:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyMap:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if p != 0 && code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if maplen(ptrToUnsafePtr(p)) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMapPtr, opStructFieldPtrHeadStringTagMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMapPtr, opStructFieldHeadStringTagMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
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
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(p))
			}
			fallthrough
		case opStructFieldHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadMarshalJSON {
					p = ptrToPtr(p + code.offset)
				}
			}
			if code.nilcheck && p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadStringTagMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(p))
			}
			fallthrough
		case opStructFieldHeadStringTagMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadStringTagMarshalJSON {
					p = ptrToPtr(p + code.offset)
				}
			}
			if code.nilcheck && p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(p))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadOmitEmptyMarshalJSON {
					p = ptrToPtr(p + code.offset)
				}
			}
			iface := ptrToInterface(code, p)
			if code.nilcheck && encodeIsNilForMarshaler(iface) {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				bb, err := encodeMarshalJSON(code, b, iface, true)
				if err != nil {
					return nil, err
				}
				b = bb
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadMarshalJSONPtr, opStructFieldPtrHeadStringTagMarshalJSONPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMarshalJSONPtr, opStructFieldHeadStringTagMarshalJSONPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalJSONPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalJSONPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(p))
			}
			fallthrough
		case opStructFieldHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadMarshalText {
					p = ptrToPtr(p + code.offset)
				}
			}
			if code.nilcheck && p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadStringTagMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(p))
			}
			fallthrough
		case opStructFieldHeadStringTagMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadStringTagMarshalText {
					p = ptrToPtr(p + code.offset)
				}
			}
			if code.nilcheck && p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(p))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadOmitEmptyMarshalText {
					p = ptrToPtr(p + code.offset)
				}
			}
			if p == 0 && code.nilcheck {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadMarshalTextPtr, opStructFieldPtrHeadStringTagMarshalTextPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMarshalTextPtr, opStructFieldHeadStringTagMarshalTextPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			b = append(b, code.escapedKey...)
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalTextPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, ptrToPtr(p))
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalTextPtr:
			p := load(ctxptr, code.idx)
			if p == 0 && code.indirect {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if !code.anonymousHead {
				b = append(b, '{')
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
				b = encodeComma(b)
				code = code.next
			}
		case opStructField:
			if !code.anonymousKey {
				b = append(b, code.escapedKey...)
			}
			ptr := load(ctxptr, code.headIdx) + code.offset
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldOmitEmpty:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 || (strings.Contains(code.next.op.String(), "Ptr") && ptrToPtr(p) == 0) {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldStringTag:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = append(b, code.escapedKey...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, u64, code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldIntPtr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
				b = appendInt(b, ptrToUint64(p), code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagIntPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = append(b, code.escapedKey...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendUint(b, u64, code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldUintPtr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
				b = appendUint(b, ptrToUint64(p), code)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagUintPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat32Ptr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, ptrToFloat32(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat32Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat64Ptr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, ptrToString(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToString(ptr + code.offset)
			if v != "" {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			s := ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, ptrToString(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyStringPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, ptrToString(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagStringPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, ptrToString(p))))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBool(ptr + code.offset)
			if v {
				b = append(b, code.escapedKey...)
				b = encodeBool(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldBoolPtr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
				b = encodeBool(b, ptrToBool(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagBoolPtr:
			b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, ptrToBytes(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldBytesPtr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
				b = encodeByteSlice(b, ptrToBytes(p))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldStringTagBytesPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldNumber:
			p := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = encodeComma(bb)
			code = code.next
		case opStructFieldOmitEmptyNumber:
			p := load(ctxptr, code.headIdx)
			v := ptrToNumber(p + code.offset)
			if v != "" {
				b = append(b, code.escapedKey...)
				bb, err := encodeNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = encodeComma(bb)
			}
			code = code.next
		case opStructFieldStringTagNumber:
			p := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldNumberPtr:
			b = append(b, code.escapedKey...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyNumberPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = append(b, code.escapedKey...)
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = encodeComma(bb)
			}
			code = code.next
		case opStructFieldStringTagNumberPtr:
			b = append(b, code.escapedKey...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = append(bb, '"')
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldMarshalJSON, opStructFieldStringTagMarshalJSON:
			p := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			p += code.offset
			if code.typ.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.nilcheck {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyMarshalJSON:
			p := load(ctxptr, code.headIdx)
			p += code.offset
			if code.typ.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.nilcheck {
				code = code.nextField
				break
			}
			b = append(b, code.escapedKey...)
			bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
			if err != nil {
				return nil, err
			}
			b = encodeComma(bb)
			code = code.next
		case opStructFieldMarshalJSONPtr, opStructFieldStringTagMarshalJSONPtr:
			p := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyMarshalJSONPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = append(b, code.escapedKey...)
				bb, err := encodeMarshalJSON(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = encodeComma(bb)
			}
			code = code.next
		case opStructFieldMarshalText, opStructFieldStringTagMarshalText:
			p := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			p += code.offset
			if code.typ.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.nilcheck {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyMarshalText:
			p := load(ctxptr, code.headIdx)
			p += code.offset
			if code.typ.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.nilcheck {
				code = code.nextField
				break
			}
			b = append(b, code.escapedKey...)
			bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
			if err != nil {
				return nil, err
			}
			b = encodeComma(bb)
			code = code.next
		case opStructFieldMarshalTextPtr, opStructFieldStringTagMarshalTextPtr:
			p := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldOmitEmptyMarshalTextPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = append(b, code.escapedKey...)
				bb, err := encodeMarshalText(code, b, ptrToInterface(code, p), true)
				if err != nil {
					return nil, err
				}
				b = encodeComma(bb)
			}
			code = code.next
		case opStructFieldArray, opStructFieldStringTagArray:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyArray:
			p := load(ctxptr, code.headIdx)
			p += code.offset
			b = append(b, code.escapedKey...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldArrayPtr, opStructFieldStringTagArrayPtr:
			b = append(b, code.escapedKey...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyArrayPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldSlice, opStructFieldStringTagSlice:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldSlicePtr, opStructFieldStringTagSlicePtr:
			b = append(b, code.escapedKey...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptySlicePtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldMap, opStructFieldStringTagMap:
			b = append(b, code.escapedKey...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyMap:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p == 0 || maplen(ptrToUnsafePtr(p)) == 0 {
				code = code.nextField
			} else {
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldMapPtr, opStructFieldStringTagMapPtr:
			b = append(b, code.escapedKey...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
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
				b = append(b, code.escapedKey...)
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldStruct:
			b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndIntPtr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
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
		case opStructEndUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUintPtr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
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
		case opStructEndFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat32Ptr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
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
		case opStructEndFloat64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat64Ptr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
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
		case opStructEndString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, ptrToString(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToString(ptr + code.offset)
			if v != "" {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, v)
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
			b = append(b, code.escapedKey...)
			s := ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndStringPtr:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeEscapedString(b, ptrToString(p))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyStringPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, ptrToString(p))
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
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := ptrToString(p)
				b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, v)))
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBool(ptr + code.offset)
			if v {
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndBoolPtr:
			b = append(b, code.escapedKey...)
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
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, ptrToBytes(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = append(b, code.escapedKey...)
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
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndNumber:
			p := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = appendStructEnd(bb)
			code = code.next
		case opStructEndOmitEmptyNumber:
			p := load(ctxptr, code.headIdx)
			v := ptrToNumber(p + code.offset)
			if v != "" {
				b = append(b, code.escapedKey...)
				bb, err := encodeNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = appendStructEnd(bb)
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
		case opStructEndStringTagNumber:
			p := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = appendStructEnd(b)
			code = code.next
		case opStructEndNumberPtr:
			b = append(b, code.escapedKey...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendStructEnd(b)
			code = code.next
		case opStructEndOmitEmptyNumberPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = append(b, code.escapedKey...)
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = appendStructEnd(bb)
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
		case opStructEndStringTagNumberPtr:
			b = append(b, code.escapedKey...)
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = append(bb, '"')
			}
			b = appendStructEnd(b)
			code = code.next
		case opEnd:
			goto END
		}
	}
END:
	return b, nil
}

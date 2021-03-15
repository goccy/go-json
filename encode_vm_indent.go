package json

import (
	"bytes"
	"encoding"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"unsafe"
)

func encodeRunIndent(ctx *encodeRuntimeContext, b []byte, codeSet *opcodeSet, opt EncodeOption) ([]byte, error) {
	recursiveLevel := 0
	ptrOffset := uintptr(0)
	ctxptr := ctx.ptr()
	code := codeSet.code

	for {
		switch code.op {
		default:
			return nil, fmt.Errorf("encoder (indent): opcode %s has not been implemented", code.op)
		case opPtr:
			ptr := load(ctxptr, code.idx)
			code = code.next
			store(ctxptr, code.idx, ptrToPtr(ptr))
		case opInt:
			b = appendInt(b, ptrToUint64(load(ctxptr, code.idx)), code)
			b = encodeIndentComma(b)
			code = code.next
		case opUint:
			b = appendUint(b, ptrToUint64(load(ctxptr, code.idx)), code)
			b = encodeIndentComma(b)
			code = code.next
		case opIntString:
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(load(ctxptr, code.idx)), code)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUintString:
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(load(ctxptr, code.idx)), code)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opFloat32:
			b = encodeFloat32(b, ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opFloat64:
			v := ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opString:
			b = encodeNoEscapedString(b, ptrToString(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBool:
			b = encodeBool(b, ptrToBool(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBytes:
			ptr := load(ctxptr, code.idx)
			slice := ptrToSlice(ptr)
			if ptr == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(ptr))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opNumber:
			bb, err := encodeNumber(b, ptrToNumber(load(ctxptr, code.idx)))
			if err != nil {
				return nil, err
			}
			b = encodeIndentComma(bb)
			code = code.next
		case opInterface:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
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
				b = encodeIndentComma(b)
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

			oldBaseIndent := ctx.baseIndent
			ctx.baseIndent = code.indent
			bb, err := encodeRunIndent(ctx, b, ifaceCodeSet, opt)
			if err != nil {
				return nil, err
			}
			ctx.baseIndent = oldBaseIndent

			ctx.ptrs = oldPtrs
			ctxptr = ctx.ptr()
			ctx.seenPtr = ctx.seenPtr[:len(ctx.seenPtr)-1]

			b = bb
			code = code.next
		case opMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
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
			var compactBuf bytes.Buffer
			if err := compact(&compactBuf, bb, false); err != nil {
				return nil, err
			}
			var indentBuf bytes.Buffer
			if err := encodeWithIndent(
				&indentBuf,
				compactBuf.Bytes(),
				string(ctx.prefix)+strings.Repeat(string(ctx.indentStr), ctx.baseIndent+code.indent),
				string(ctx.indentStr),
			); err != nil {
				return nil, err
			}
			b = append(b, indentBuf.Bytes()...)
			b = encodeIndentComma(b)
			code = code.next
		case opMarshalText:
			ptr := load(ctxptr, code.idx)
			isPtr := code.typ.Kind() == reflect.Ptr
			p := ptrToUnsafePtr(ptr)
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
		case opSlice:
			p := load(ctxptr, code.idx)
			slice := ptrToSlice(p)
			if p == 0 || slice.data == nil {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				store(ctxptr, code.elemIdx, 0)
				store(ctxptr, code.length, uintptr(slice.len))
				store(ctxptr, code.idx, uintptr(slice.data))
				if slice.len > 0 {
					b = append(b, '[', '\n')
					b = appendIndent(ctx, b, code.indent+1)
					code = code.next
					store(ctxptr, code.idx, uintptr(slice.data))
				} else {
					b = appendIndent(ctx, b, code.indent)
					b = append(b, '[', ']', ',', '\n')
					code = code.end.next
				}
			}
		case opSliceElem:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if idx < length {
				b = appendIndent(ctx, b, code.indent+1)
				store(ctxptr, code.elemIdx, idx)
				data := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, data+idx*size)
			} else {
				b = b[:len(b)-2]
				b = append(b, '\n')
				b = appendIndent(ctx, b, code.indent)
				b = append(b, ']', ',', '\n')
				code = code.end.next
			}
		case opArray:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				if code.length > 0 {
					b = append(b, '[', '\n')
					b = appendIndent(ctx, b, code.indent+1)
					store(ctxptr, code.elemIdx, 0)
					code = code.next
					store(ctxptr, code.idx, p)
				} else {
					b = appendIndent(ctx, b, code.indent)
					b = append(b, '[', ']', ',', '\n')
					code = code.end.next
				}
			}
		case opArrayElem:
			idx := load(ctxptr, code.elemIdx)
			idx++
			if idx < code.length {
				b = appendIndent(ctx, b, code.indent+1)
				store(ctxptr, code.elemIdx, idx)
				p := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, p+idx*size)
			} else {
				b = b[:len(b)-2]
				b = append(b, '\n')
				b = appendIndent(ctx, b, code.indent)
				b = append(b, ']', ',', '\n')
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
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				uptr := ptrToUnsafePtr(ptr)
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{', '\n')
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
					} else {
						b = appendIndent(ctx, b, code.next.indent)
					}

					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					b = append(b, '{', '}', ',', '\n')
					code = code.end.next
				}
			}
		case opMapKey:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if (opt & EncodeOptionUnorderedMap) != 0 {
				if idx < length {
					b = appendIndent(ctx, b, code.indent)
					store(ctxptr, code.elemIdx, idx)
					ptr := load(ctxptr, code.mapIter)
					iter := ptrToUnsafePtr(ptr)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					last := len(b) - 1
					b[last] = '\n'
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}', ',', '\n')
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
				b = append(b, ':', ' ')
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
			// this operation only used by sorted map
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
				buf = append(buf, ctx.prefix...)
				buf = append(buf, bytes.Repeat(ctx.indentStr, ctx.baseIndent+code.indent+1)...)
				buf = append(buf, item.key...)
				buf[len(buf)-2] = ':'
				buf[len(buf)-1] = ' '
				buf = append(buf, item.value...)
			}
			buf = buf[:len(buf)-2]
			buf = append(buf, '\n')
			buf = append(buf, ctx.prefix...)
			buf = append(buf, bytes.Repeat(ctx.indentStr, ctx.baseIndent+code.indent)...)
			buf = append(buf, '}', ',', '\n')

			b = b[:pos[0]]
			b = append(b, buf...)
			mapCtx.buf = buf
			releaseMapContext(mapCtx)
			code = code.next
		case opStructFieldPtrHeadRecursive:
			store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
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
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if !code.anonymousKey && len(code.key) > 0 {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
			}
			p += code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldPtrHeadOmitEmpty:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			p += code.offset
			if p == 0 || (strings.Contains(code.next.op.String(), "Ptr") && ptrToPtr(p) == 0) {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadStringTag:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			p += code.offset
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendInt(b, ptrToUint64(p+code.offset), code)
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			u64 := ptrToUint64(p + code.offset)
			v := u64 & code.mask
			if v == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, u64, code)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, ptrToUint64(p+code.offset), code)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, ptrToUint64(p), code)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = appendUint(b, ptrToUint64(p+code.offset), code)
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			u64 := ptrToUint64(p + code.offset)
			v := u64 & code.mask
			if v == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, u64, code)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, ptrToUint64(p+code.offset), code)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, ptrToUint64(p), code)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagUintPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeFloat32(b, ptrToFloat32(p+code.offset))
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToFloat32(p + code.offset)
			if v == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, v)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, ptrToFloat32(p+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, ptrToFloat32(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			if v == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			v := ptrToFloat64(p + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagFloat64Ptr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadString:
			if code.indirect {
				store(ctxptr, code.idx, ptrToPtr(load(ctxptr, code.idx)))
			}
			fallthrough
		case opStructFieldHeadString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeNoEscapedString(b, ptrToString(p+code.offset))
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToString(p + code.offset)
			if v == "" {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, v)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			v := ptrToString(p + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, v)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, ptrToString(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, ptrToString(p))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagStringPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, ptrToString(p))))
			}
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeBool(b, ptrToBool(p+code.offset))
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToBool(p + code.offset)
			if v {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, v)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeBool(b, ptrToBool(p+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, ptrToBool(p+code.offset))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, ptrToBool(p))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagBoolPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeByteSlice(b, ptrToBytes(p+code.offset))
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToBytes(p + code.offset)
			if len(v) == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, v)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeByteSlice(b, ptrToBytes(p+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, ptrToBytes(p))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagBytesPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldHeadNumber:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = encodeIndentComma(bb)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToNumber(p + code.offset)
			if v == "" {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = encodeIndentComma(bb)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = append(b, '"')
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p != 0 {
				b = append(b, code.key...)
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = encodeIndentComma(bb)
			}
			code = code.next
		case opStructFieldPtrHeadStringTagNumberPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			p += code.offset
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			p += code.offset
			slice := ptrToSlice(p)
			if slice.data == nil {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadArrayPtr, opStructFieldPtrHeadStringTagArrayPtr,
			opStructFieldPtrHeadSlicePtr, opStructFieldPtrHeadStringTagSlicePtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMap, opStructFieldPtrHeadStringTagMap:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
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
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if p != 0 && code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if maplen(ptrToUnsafePtr(p)) == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMapPtr, opStructFieldPtrHeadStringTagMapPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.nextField
				break
			}
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
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
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadMarshalJSON {
					p = ptrToPtr(p + code.offset)
				}
			}
			if code.nilcheck && p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadStringTagMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadStringTagMarshalJSON {
					p = ptrToPtr(p + code.offset)
				}
			}
			if code.nilcheck && p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadOmitEmptyMarshalJSON {
					p = ptrToPtr(p + code.offset)
				}
			}
			if p == 0 && code.nilcheck {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
				if err != nil {
					return nil, err
				}
				b = bb
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadMarshalJSONPtr, opStructFieldPtrHeadStringTagMarshalJSONPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalJSONPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
				if err != nil {
					return nil, err
				}
				b = bb
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadMarshalText {
					p = ptrToPtr(p + code.offset)
				}
			}
			if code.nilcheck && p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadStringTagMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadStringTagMarshalText {
					p = ptrToPtr(p + code.offset)
				}
			}
			if code.nilcheck && p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if code.typ.Kind() == reflect.Ptr {
				if code.indirect || code.op == opStructFieldPtrHeadOmitEmptyMarshalText {
					p = ptrToPtr(p + code.offset)
				}
			}
			if p == 0 && code.nilcheck {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadMarshalTextPtr, opStructFieldPtrHeadStringTagMarshalTextPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.indent+1)
			b = append(b, code.key...)
			b = append(b, ' ')
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldPtrHeadOmitEmptyMarshalTextPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				if !code.anonymousHead {
					b = encodeNull(b)
					b = encodeIndentComma(b)
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
					b = encodeIndentComma(b)
				}
				code = code.end.next
				break
			}
			if code.indirect {
				p = ptrToPtr(p + code.offset)
			}
			if !code.anonymousHead {
				b = append(b, '{', '\n')
			}
			if p == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructField:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmpty:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 || (strings.Contains(code.next.op.String(), "Ptr") && ptrToPtr(p) == 0) {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldStringTag:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldInt:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, u64, code)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldIntPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyIntPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, ptrToUint64(p), code)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagIntPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, u64, code)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUintPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyUintPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, ptrToUint64(p), code)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagUintPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldFloat32:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldFloat32Ptr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, ptrToFloat32(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat32Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat32Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldFloat64:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldFloat64Ptr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyFloat64Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagFloat64Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldString:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeNoEscapedString(b, ptrToString(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToString(ptr + code.offset)
			if v != "" {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagString:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			s := ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, ptrToString(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyStringPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, ptrToString(p))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagStringPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, ptrToString(p))))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldBool:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBool(ptr + code.offset)
			if v {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldBoolPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, ptrToBool(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyBoolPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, ptrToBool(p))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagBoolPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldBytes:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, ptrToBytes(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeByteSlice(b, ptrToBytes(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldBytesPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyBytesPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, ptrToBytes(p))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldStringTagBytesPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, ptrToBytes(p))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldNumber:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := load(ctxptr, code.headIdx)
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = encodeIndentComma(bb)
			code = code.next
		case opStructFieldOmitEmptyNumber:
			p := load(ctxptr, code.headIdx)
			v := ptrToNumber(p + code.offset)
			if v != "" {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = encodeIndentComma(bb)
			}
			code = code.next
		case opStructFieldStringTagNumber:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = append(b, '"')
			p := load(ctxptr, code.headIdx)
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldNumberPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyNumberPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = encodeIndentComma(bb)
			}
			code = code.next
		case opStructFieldStringTagNumberPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldMarshalJSON, opStructFieldStringTagMarshalJSON:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := load(ctxptr, code.headIdx)
			p += code.offset
			if code.typ.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.nilcheck {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
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
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
			if err != nil {
				return nil, err
			}
			b = encodeIndentComma(bb)
			code = code.next
		case opStructFieldMarshalJSONPtr, opStructFieldStringTagMarshalJSONPtr:
			p := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyMarshalJSONPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeMarshalJSONIndent(ctx, code, b, ptrToInterface(code, p), code.indent, false)
				if err != nil {
					return nil, err
				}
				b = encodeIndentComma(bb)
			}
			code = code.next
		case opStructFieldMarshalText, opStructFieldStringTagMarshalText:
			p := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p += code.offset
			if code.typ.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.nilcheck {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
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
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
			if err != nil {
				return nil, err
			}
			b = encodeIndentComma(bb)
			code = code.next
		case opStructFieldMarshalTextPtr, opStructFieldStringTagMarshalTextPtr:
			p := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p = ptrToPtr(p + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldOmitEmptyMarshalTextPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeMarshalTextIndent(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = encodeIndentComma(bb)
			}
			code = code.next
		case opStructFieldArray, opStructFieldStringTagArray:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := load(ctxptr, code.headIdx)
			p += code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyArray:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := load(ctxptr, code.headIdx)
			p += code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldArrayPtr, opStructFieldStringTagArrayPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptyArrayPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldSlice, opStructFieldStringTagSlice:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := load(ctxptr, code.headIdx)
			p += code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptySlice:
			p := load(ctxptr, code.headIdx)
			p += code.offset
			slice := ptrToSlice(p)
			if slice.data == nil {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldSlicePtr, opStructFieldStringTagSlicePtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldOmitEmptySlicePtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldMap, opStructFieldStringTagMap:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
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
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldMapPtr, opStructFieldStringTagMapPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
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
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			} else {
				code = code.nextField
			}
		case opStructFieldStruct:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = appendIndent(ctx, b, code.indent)
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
		case opStructFieldOmitEmptyStruct:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				b = appendIndent(ctx, b, code.indent)
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
		case opStructAnonymousEnd:
			code = code.next
		case opStructEnd:
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
			b = appendIndent(ctx, b, code.indent)
			b = append(b, '}')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEndInt:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, u64, code)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndIntPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyIntPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, ptrToUint64(p), code)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagIntPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndUint:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			u64 := ptrToUint64(ptr + code.offset)
			v := u64 & code.mask
			if v != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, u64, code)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, ptrToUint64(ptr+code.offset), code)
			b = append(b, '"')
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndUintPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyUintPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendUint(b, ptrToUint64(p), code)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagUintPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndFloat32:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, v)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, ptrToFloat32(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndFloat32Ptr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, ptrToFloat32(p))
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyFloat32Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagFloat32Ptr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndFloat64:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if v != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat64(b, v)
			b = append(b, '"')
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndFloat64Ptr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyFloat64Ptr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = encodeFloat64(b, v)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagFloat64Ptr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndString:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeNoEscapedString(b, ptrToString(ptr+code.offset))
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToString(ptr + code.offset)
			if v != "" {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, v)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagString:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			s := ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndStringPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, ptrToString(p))
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyStringPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeNoEscapedString(b, ptrToString(p))
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagStringPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, ptrToString(p))))
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndBool:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBool(ptr + code.offset)
			if v {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, v)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeBool(b, ptrToBool(ptr+code.offset))
			b = append(b, '"')
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndBoolPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, ptrToBool(p))
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyBoolPtr:
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeBool(b, ptrToBool(p))
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagBoolPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = append(b, '"')
				b = encodeBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndBytes:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, ptrToBytes(ptr+code.offset))
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeByteSlice(b, v)
				b = appendStructEndIndent(ctx, b, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeByteSlice(b, ptrToBytes(ptr+code.offset))
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndNumber:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			p := load(ctxptr, code.headIdx)
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = appendStructEndIndent(ctx, bb, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyNumber:
			p := load(ctxptr, code.headIdx)
			v := ptrToNumber(p + code.offset)
			if v != "" {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = appendStructEndIndent(ctx, bb, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagNumber:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = append(b, '"')
			p := load(ctxptr, code.headIdx)
			bb, err := encodeNumber(b, ptrToNumber(p+code.offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndNumberPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opStructEndOmitEmptyNumberPtr:
			p := load(ctxptr, code.headIdx)
			p = ptrToPtr(p + code.offset)
			if p != 0 {
				b = appendIndent(ctx, b, code.indent)
				b = append(b, code.key...)
				b = append(b, ' ')
				bb, err := encodeNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = appendStructEndIndent(ctx, bb, code.indent-1)
			} else {
				last := len(b) - 1
				if b[last-1] == '{' {
					b[last] = '}'
				} else {
					if b[last] == '\n' {
						// to remove ',' and '\n' characters
						b = b[:len(b)-2]
					}
					b = append(b, '\n')
					b = appendIndent(ctx, b, code.indent-1)
					b = append(b, '}')
				}
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructEndStringTagNumberPtr:
			b = appendIndent(ctx, b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
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
			b = appendStructEndIndent(ctx, b, code.indent-1)
			code = code.next
		case opEnd:
			goto END
		}
	}
END:
	return b, nil
}

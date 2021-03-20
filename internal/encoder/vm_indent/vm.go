package vm_indent

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/runtime"

	// HACK: compile order
	// `vm`, `vm_escaped`, `vm_indent`, `vm_escaped_indent` packages uses a lot of memory to compile,
	// so forcibly make dependencies and avoid compiling in concurrent.
	// dependency order: vm => vm_escaped => vm_indent => vm_escaped_indent
	_ "github.com/goccy/go-json/internal/encoder/vm_escaped_indent"
)

const uintptrSize = 4 << (^uintptr(0) >> 63)

var (
	appendInt           = encoder.AppendInt
	appendUint          = encoder.AppendUint
	appendFloat32       = encoder.AppendFloat32
	appendFloat64       = encoder.AppendFloat64
	appendString        = encoder.AppendString
	appendByteSlice     = encoder.AppendByteSlice
	appendNumber        = encoder.AppendNumber
	appendMarshalJSON   = encoder.AppendMarshalJSONIndent
	appendMarshalText   = encoder.AppendMarshalTextIndent
	errUnsupportedValue = encoder.ErrUnsupportedValue
	errUnsupportedFloat = encoder.ErrUnsupportedFloat
	mapiterinit         = encoder.MapIterInit
	mapiterkey          = encoder.MapIterKey
	mapitervalue        = encoder.MapIterValue
	mapiternext         = encoder.MapIterNext
	maplen              = encoder.MapLen
)

type emptyInterface struct {
	typ *runtime.Type
	ptr unsafe.Pointer
}

func Run(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, opt encoder.Option) ([]byte, error) {
	recursiveLevel := 0
	ptrOffset := uintptr(0)
	ctxptr := ctx.Ptr()
	code := codeSet.Code

	for {
		switch code.Op {
		default:
			return nil, fmt.Errorf("encoder (indent): opcode %s has not been implemented", code.Op)
		case encoder.OpPtr:
			p := load(ctxptr, code.Idx)
			code = code.Next
			store(ctxptr, code.Idx, ptrToPtr(p))
		case encoder.OpIntPtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpInt:
			b = appendInt(b, ptrToUint64(load(ctxptr, code.Idx)), code)
			b = appendComma(b)
			code = code.Next
		case encoder.OpUintPtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpUint:
			b = appendUint(b, ptrToUint64(load(ctxptr, code.Idx)), code)
			b = appendComma(b)
			code = code.Next
		case encoder.OpIntString:
			b = append(b, '"')
			b = appendInt(b, ptrToUint64(load(ctxptr, code.Idx)), code)
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpUintString:
			b = append(b, '"')
			b = appendUint(b, ptrToUint64(load(ctxptr, code.Idx)), code)
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpFloat32Ptr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpFloat32:
			b = appendFloat32(b, ptrToFloat32(load(ctxptr, code.Idx)))
			b = appendComma(b)
			code = code.Next
		case encoder.OpFloat64Ptr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpFloat64:
			v := ptrToFloat64(load(ctxptr, code.Idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = appendFloat64(b, v)
			b = appendComma(b)
			code = code.Next
		case encoder.OpStringPtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpString:
			b = appendString(b, ptrToString(load(ctxptr, code.Idx)))
			b = appendComma(b)
			code = code.Next
		case encoder.OpBoolPtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpBool:
			b = appendBool(b, ptrToBool(load(ctxptr, code.Idx)))
			b = appendComma(b)
			code = code.Next
		case encoder.OpBytesPtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpBytes:
			b = appendByteSlice(b, ptrToBytes(load(ctxptr, code.Idx)))
			b = appendComma(b)
			code = code.Next
		case encoder.OpNumberPtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpNumber:
			bb, err := appendNumber(b, ptrToNumber(load(ctxptr, code.Idx)))
			if err != nil {
				return nil, err
			}
			b = appendComma(bb)
			code = code.Next
		case encoder.OpInterfacePtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpInterface:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			for _, seen := range ctx.SeenPtr {
				if p == seen {
					return nil, errUnsupportedValue(code, p)
				}
			}
			ctx.SeenPtr = append(ctx.SeenPtr, p)
			iface := (*emptyInterface)(ptrToUnsafePtr(p))
			if iface.ptr == nil {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			ctx.KeepRefs = append(ctx.KeepRefs, unsafe.Pointer(iface))
			ifaceCodeSet, err := encoder.CompileToGetCodeSet(uintptr(unsafe.Pointer(iface.typ)))
			if err != nil {
				return nil, err
			}

			totalLength := uintptr(codeSet.CodeLength)
			nextTotalLength := uintptr(ifaceCodeSet.CodeLength)

			curlen := uintptr(len(ctx.Ptrs))
			offsetNum := ptrOffset / uintptrSize

			newLen := offsetNum + totalLength + nextTotalLength
			if curlen < newLen {
				ctx.Ptrs = append(ctx.Ptrs, make([]uintptr, newLen-curlen)...)
			}
			oldPtrs := ctx.Ptrs

			newPtrs := ctx.Ptrs[(ptrOffset+totalLength*uintptrSize)/uintptrSize:]
			newPtrs[0] = uintptr(iface.ptr)

			ctx.Ptrs = newPtrs

			oldBaseIndent := ctx.BaseIndent
			ctx.BaseIndent = code.Indent
			bb, err := Run(ctx, b, ifaceCodeSet, opt)
			if err != nil {
				return nil, err
			}
			ctx.BaseIndent = oldBaseIndent

			ctx.Ptrs = oldPtrs
			ctxptr = ctx.Ptr()
			ctx.SeenPtr = ctx.SeenPtr[:len(ctx.SeenPtr)-1]

			b = bb
			code = code.Next
		case encoder.OpMarshalJSONPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, ptrToPtr(p))
			fallthrough
		case encoder.OpMarshalJSON:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			if code.Type.Kind() == reflect.Ptr && code.Indirect {
				p = ptrToPtr(p)
			}
			bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent, false)
			if err != nil {
				return nil, err
			}
			b = appendComma(bb)
			code = code.Next
		case encoder.OpMarshalTextPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, ptrToPtr(p))
			fallthrough
		case encoder.OpMarshalText:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				b = append(b, `""`...)
				b = appendComma(b)
				code = code.Next
				break
			}
			if code.Type.Kind() == reflect.Ptr && code.Indirect {
				p = ptrToPtr(p)
			}
			bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
			if err != nil {
				return nil, err
			}
			b = appendComma(bb)
			code = code.Next
		case encoder.OpSlicePtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpSlice:
			p := load(ctxptr, code.Idx)
			slice := ptrToSlice(p)
			if p == 0 || slice.Data == nil {
				b = appendNull(b)
				b = appendComma(b)
				code = code.End.Next
				break
			}
			store(ctxptr, code.ElemIdx, 0)
			store(ctxptr, code.Length, uintptr(slice.Len))
			store(ctxptr, code.Idx, uintptr(slice.Data))
			if slice.Len > 0 {
				b = append(b, '[', '\n')
				b = appendIndent(ctx, b, code.Indent+1)
				code = code.Next
				store(ctxptr, code.Idx, uintptr(slice.Data))
			} else {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, '[', ']', ',', '\n')
				code = code.End.Next
			}
		case encoder.OpSliceElem:
			idx := load(ctxptr, code.ElemIdx)
			length := load(ctxptr, code.Length)
			idx++
			if idx < length {
				b = appendIndent(ctx, b, code.Indent+1)
				store(ctxptr, code.ElemIdx, idx)
				data := load(ctxptr, code.HeadIdx)
				size := code.Size
				code = code.Next
				store(ctxptr, code.Idx, data+idx*size)
			} else {
				b = b[:len(b)-2]
				b = append(b, '\n')
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, ']', ',', '\n')
				code = code.End.Next
			}
		case encoder.OpArrayPtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpArray:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.End.Next
				break
			}
			if code.Length > 0 {
				b = append(b, '[', '\n')
				b = appendIndent(ctx, b, code.Indent+1)
				store(ctxptr, code.ElemIdx, 0)
				code = code.Next
				store(ctxptr, code.Idx, p)
			} else {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, '[', ']', ',', '\n')
				code = code.End.Next
			}
		case encoder.OpArrayElem:
			idx := load(ctxptr, code.ElemIdx)
			idx++
			if idx < code.Length {
				b = appendIndent(ctx, b, code.Indent+1)
				store(ctxptr, code.ElemIdx, idx)
				p := load(ctxptr, code.HeadIdx)
				size := code.Size
				code = code.Next
				store(ctxptr, code.Idx, p+idx*size)
			} else {
				b = b[:len(b)-2]
				b = append(b, '\n')
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, ']', ',', '\n')
				code = code.End.Next
			}
		case encoder.OpMapPtr:
			p := loadNPtr(ctxptr, code.Idx, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, p)
			fallthrough
		case encoder.OpMap:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.End.Next
				break
			}
			uptr := ptrToUnsafePtr(p)
			mlen := maplen(uptr)
			if mlen <= 0 {
				b = append(b, '{', '}', ',', '\n')
				code = code.End.Next
				break
			}
			b = append(b, '{', '\n')
			iter := mapiterinit(code.Type, uptr)
			ctx.KeepRefs = append(ctx.KeepRefs, iter)
			store(ctxptr, code.ElemIdx, 0)
			store(ctxptr, code.Length, uintptr(mlen))
			store(ctxptr, code.MapIter, uintptr(iter))
			if (opt & encoder.UnorderedMapOption) == 0 {
				mapCtx := encoder.NewMapContext(mlen)
				mapCtx.Pos = append(mapCtx.Pos, len(b))
				ctx.KeepRefs = append(ctx.KeepRefs, unsafe.Pointer(mapCtx))
				store(ctxptr, code.End.MapPos, uintptr(unsafe.Pointer(mapCtx)))
			} else {
				b = appendIndent(ctx, b, code.Next.Indent)
			}
			key := mapiterkey(iter)
			store(ctxptr, code.Next.Idx, uintptr(key))
			code = code.Next
		case encoder.OpMapKey:
			idx := load(ctxptr, code.ElemIdx)
			length := load(ctxptr, code.Length)
			idx++
			if (opt & encoder.UnorderedMapOption) != 0 {
				if idx < length {
					b = appendIndent(ctx, b, code.Indent)
					store(ctxptr, code.ElemIdx, idx)
					ptr := load(ctxptr, code.MapIter)
					iter := ptrToUnsafePtr(ptr)
					key := mapiterkey(iter)
					store(ctxptr, code.Next.Idx, uintptr(key))
					code = code.Next
				} else {
					last := len(b) - 1
					b[last] = '\n'
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}', ',', '\n')
					code = code.End.Next
				}
			} else {
				ptr := load(ctxptr, code.End.MapPos)
				mapCtx := (*encoder.MapContext)(ptrToUnsafePtr(ptr))
				mapCtx.Pos = append(mapCtx.Pos, len(b))
				if idx < length {
					ptr := load(ctxptr, code.MapIter)
					iter := ptrToUnsafePtr(ptr)
					store(ctxptr, code.ElemIdx, idx)
					key := mapiterkey(iter)
					store(ctxptr, code.Next.Idx, uintptr(key))
					code = code.Next
				} else {
					code = code.End
				}
			}
		case encoder.OpMapValue:
			if (opt & encoder.UnorderedMapOption) != 0 {
				b = append(b, ':', ' ')
			} else {
				ptr := load(ctxptr, code.End.MapPos)
				mapCtx := (*encoder.MapContext)(ptrToUnsafePtr(ptr))
				mapCtx.Pos = append(mapCtx.Pos, len(b))
			}
			ptr := load(ctxptr, code.MapIter)
			iter := ptrToUnsafePtr(ptr)
			value := mapitervalue(iter)
			store(ctxptr, code.Next.Idx, uintptr(value))
			mapiternext(iter)
			code = code.Next
		case encoder.OpMapEnd:
			// this operation only used by sorted map
			length := int(load(ctxptr, code.Length))
			ptr := load(ctxptr, code.MapPos)
			mapCtx := (*encoder.MapContext)(ptrToUnsafePtr(ptr))
			pos := mapCtx.Pos
			for i := 0; i < length; i++ {
				startKey := pos[i*2]
				startValue := pos[i*2+1]
				var endValue int
				if i+1 < length {
					endValue = pos[i*2+2]
				} else {
					endValue = len(b)
				}
				mapCtx.Slice.Items = append(mapCtx.Slice.Items, encoder.MapItem{
					Key:   b[startKey:startValue],
					Value: b[startValue:endValue],
				})
			}
			sort.Sort(mapCtx.Slice)
			buf := mapCtx.Buf
			for _, item := range mapCtx.Slice.Items {
				buf = append(buf, ctx.Prefix...)
				buf = append(buf, bytes.Repeat(ctx.IndentStr, ctx.BaseIndent+code.Indent+1)...)
				buf = append(buf, item.Key...)
				buf[len(buf)-2] = ':'
				buf[len(buf)-1] = ' '
				buf = append(buf, item.Value...)
			}
			buf = buf[:len(buf)-2]
			buf = append(buf, '\n')
			buf = append(buf, ctx.Prefix...)
			buf = append(buf, bytes.Repeat(ctx.IndentStr, ctx.BaseIndent+code.Indent)...)
			buf = append(buf, '}', ',', '\n')

			b = b[:pos[0]]
			b = append(b, buf...)
			mapCtx.Buf = buf
			encoder.ReleaseMapContext(mapCtx)
			code = code.Next
		case encoder.OpRecursivePtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				code = code.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpRecursive:
			ptr := load(ctxptr, code.Idx)
			if ptr != 0 {
				if recursiveLevel > encoder.StartDetectingCyclesAfter {
					for _, seen := range ctx.SeenPtr {
						if ptr == seen {
							return nil, errUnsupportedValue(code, ptr)
						}
					}
				}
			}
			ctx.SeenPtr = append(ctx.SeenPtr, ptr)
			c := code.Jmp.Code
			curlen := uintptr(len(ctx.Ptrs))
			offsetNum := ptrOffset / uintptrSize
			oldOffset := ptrOffset
			ptrOffset += code.Jmp.CurLen * uintptrSize

			newLen := offsetNum + code.Jmp.CurLen + code.Jmp.NextLen
			if curlen < newLen {
				ctx.Ptrs = append(ctx.Ptrs, make([]uintptr, newLen-curlen)...)
			}
			ctxptr = ctx.Ptr() + ptrOffset // assign new ctxptr

			store(ctxptr, c.Idx, ptr)
			store(ctxptr, c.End.Next.Idx, oldOffset)
			store(ctxptr, c.End.Next.ElemIdx, uintptr(unsafe.Pointer(code.Next)))
			code = c
			recursiveLevel++
		case encoder.OpRecursiveEnd:
			recursiveLevel--

			// restore ctxptr
			offset := load(ctxptr, code.Idx)
			ctx.SeenPtr = ctx.SeenPtr[:len(ctx.SeenPtr)-1]

			codePtr := load(ctxptr, code.ElemIdx)
			code = (*encoder.Opcode)(ptrToUnsafePtr(codePtr))
			ctxptr = ctx.Ptr() + offset
			ptrOffset = offset
		case encoder.OpStructPtrHead:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHead:
			p := load(ctxptr, code.Idx)
			if p == 0 && (code.Indirect || code.Next.Op == encoder.OpStructEnd) {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if !code.AnonymousKey && len(code.Key) > 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
			}
			p += code.Offset
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructPtrHeadOmitEmpty:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmpty:
			p := load(ctxptr, code.Idx)
			if p == 0 && (code.Indirect || code.Next.Op == encoder.OpStructEnd) {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			p += code.Offset
			if p == 0 || (ptrToPtr(p) == 0 && code.IsNextOpPtrType) {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructPtrHeadStringTag:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTag:
			p := load(ctxptr, code.Idx)
			if p == 0 && (code.Indirect || code.Next.Op == encoder.OpStructEnd) {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			p += code.Offset
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructPtrHeadInt:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadInt:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendInt(b, ptrToUint64(p+code.Offset), code)
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyInt:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyInt:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			u64 := ptrToUint64(p + code.Offset)
			v := u64 & code.Mask
			if v == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendInt(b, u64, code)
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadStringTagInt:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagInt:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendInt(b, ptrToUint64(p+code.Offset), code)
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadIntPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadIntPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyIntPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyIntPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendInt(b, ptrToUint64(p), code)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructPtrHeadStringTagIntPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTagIntPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadUint:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadUint:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendUint(b, ptrToUint64(p+code.Offset), code)
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyUint:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyUint:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			u64 := ptrToUint64(p + code.Offset)
			v := u64 & code.Mask
			if v == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendUint(b, u64, code)
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadStringTagUint:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagUint:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendUint(b, ptrToUint64(p+code.Offset), code)
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadUintPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadUintPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyUintPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyUintPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendUint(b, ptrToUint64(p), code)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructPtrHeadStringTagUintPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTagUintPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadFloat32:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadFloat32:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendFloat32(b, ptrToFloat32(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyFloat32:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyFloat32:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToFloat32(p + code.Offset)
			if v == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendFloat32(b, v)
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadStringTagFloat32:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagFloat32:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendFloat32(b, ptrToFloat32(p+code.Offset))
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadFloat32Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadFloat32Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendFloat32(b, ptrToFloat32(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendFloat32(b, ptrToFloat32(p))
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructPtrHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTagFloat32Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadFloat64:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadFloat64:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			v := ptrToFloat64(p + code.Offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendFloat64(b, v)
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyFloat64:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyFloat64:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToFloat64(p + code.Offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			if v == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendFloat64(b, v)
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadStringTagFloat64:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagFloat64:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			v := ptrToFloat64(p + code.Offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = appendFloat64(b, v)
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadFloat64Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadFloat64Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructPtrHeadStringTagFloat64Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTagFloat64Ptr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadString:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadString:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendString(b, ptrToString(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyString:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyString:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToString(p + code.Offset)
			if v == "" {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendString(b, v)
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadStringTagString:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagString:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			v := ptrToString(p + code.Offset)
			b = appendString(b, string(appendString([]byte{}, v)))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadStringPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendString(b, ptrToString(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyStringPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendString(b, ptrToString(p))
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructPtrHeadStringTagStringPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTagStringPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendString(b, string(appendString([]byte{}, ptrToString(p))))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadBool:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadBool:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendBool(b, ptrToBool(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyBool:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyBool:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToBool(p + code.Offset)
			if v {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendBool(b, v)
				b = appendComma(b)
				code = code.Next
			} else {
				code = code.NextField
			}
		case encoder.OpStructPtrHeadStringTagBool:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagBool:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendBool(b, ptrToBool(p+code.Offset))
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadBoolPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadBoolPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendBool(b, ptrToBool(p+code.Offset))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyBoolPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyBoolPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendBool(b, ptrToBool(p))
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructPtrHeadStringTagBoolPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTagBoolPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadBytes:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadBytes:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendByteSlice(b, ptrToBytes(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyBytes:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyBytes:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToBytes(p + code.Offset)
			if len(v) == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendByteSlice(b, v)
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadStringTagBytes:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagBytes:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendByteSlice(b, ptrToBytes(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadBytesPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadBytesPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendByteSlice(b, ptrToBytes(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyBytesPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyBytesPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendByteSlice(b, ptrToBytes(p))
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructPtrHeadStringTagBytesPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTagBytesPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToPtr(p + code.Offset)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendByteSlice(b, ptrToBytes(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadNumber:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadNumber:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			bb, err := appendNumber(b, ptrToNumber(p+code.Offset))
			if err != nil {
				return nil, err
			}
			b = appendComma(bb)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyNumber:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyNumber:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			v := ptrToNumber(p + code.Offset)
			if v == "" {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = appendComma(bb)
				code = code.Next
			}
		case encoder.OpStructPtrHeadStringTagNumber:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagNumber:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = append(b, '"')
			bb, err := appendNumber(b, ptrToNumber(p+code.Offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadNumberPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadNumberPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyNumberPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyNumberPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = appendComma(bb)
			}
			code = code.Next
		case encoder.OpStructPtrHeadStringTagNumberPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadStringTagNumberPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = append(bb, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadArray, encoder.OpStructPtrHeadStringTagArray,
			encoder.OpStructPtrHeadSlice, encoder.OpStructPtrHeadStringTagSlice:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadArray, encoder.OpStructHeadStringTagArray,
			encoder.OpStructHeadSlice, encoder.OpStructHeadStringTagSlice:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p += code.Offset
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructPtrHeadOmitEmptyArray:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyArray:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			p += code.Offset
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructPtrHeadOmitEmptySlice:
			if code.Indirect {
				p := load(ctxptr, code.Idx)
				if p == 0 {
					if !code.AnonymousHead {
						b = appendNull(b)
						b = appendComma(b)
					}
					code = code.End.Next
					break
				}
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptySlice:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			p += code.Offset
			slice := ptrToSlice(p)
			if slice.Data == nil {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructPtrHeadArrayPtr, encoder.OpStructPtrHeadStringTagArrayPtr,
			encoder.OpStructPtrHeadSlicePtr, encoder.OpStructPtrHeadStringTagSlicePtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadArrayPtr, encoder.OpStructHeadStringTagArrayPtr,
			encoder.OpStructHeadSlicePtr, encoder.OpStructHeadStringTagSlicePtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.NextField
			} else {
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructPtrHeadOmitEmptyArrayPtr, encoder.OpStructPtrHeadOmitEmptySlicePtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyArrayPtr, encoder.OpStructHeadOmitEmptySlicePtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructPtrHeadMap, encoder.OpStructPtrHeadStringTagMap:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadMap, encoder.OpStructHeadStringTagMap:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if p != 0 && code.Indirect {
				p = ptrToPtr(p + code.Offset)
			}
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructPtrHeadOmitEmptyMap:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyMap:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if p != 0 && code.Indirect {
				p = ptrToPtr(p + code.Offset)
			}
			if maplen(ptrToUnsafePtr(p)) == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructPtrHeadMapPtr, encoder.OpStructPtrHeadStringTagMapPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadMapPtr, encoder.OpStructHeadStringTagMapPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.NextField
				break
			}
			p = ptrToPtr(p + code.Offset)
			if p == 0 {
				b = appendNull(b)
				b = appendComma(b)
				code = code.NextField
			} else {
				if code.Indirect {
					p = ptrToNPtr(p, code.PtrNum)
				}
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructPtrHeadOmitEmptyMapPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyMapPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if p == 0 {
				code = code.NextField
				break
			}
			p = ptrToPtr(p + code.Offset)
			if p == 0 {
				code = code.NextField
			} else {
				if code.Indirect {
					p = ptrToNPtr(p, code.PtrNum)
				}
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructPtrHeadMarshalJSON:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if code.Indirect {
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadMarshalJSON:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Type.Kind() == reflect.Ptr {
				if code.Indirect || code.Op == encoder.OpStructPtrHeadMarshalJSON {
					p = ptrToPtr(p + code.Offset)
				}
			}
			if p == 0 && code.Nilcheck {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent+1, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadStringTagMarshalJSON:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if code.Indirect {
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagMarshalJSON:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Type.Kind() == reflect.Ptr {
				if code.Indirect || code.Op == encoder.OpStructPtrHeadStringTagMarshalJSON {
					p = ptrToPtr(p + code.Offset)
				}
			}
			if p == 0 && code.Nilcheck {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent+1, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyMarshalJSON:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if code.Indirect {
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyMarshalJSON:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Type.Kind() == reflect.Ptr {
				if code.Indirect || code.Op == encoder.OpStructPtrHeadOmitEmptyMarshalJSON {
					p = ptrToPtr(p + code.Offset)
				}
			}
			iface := ptrToInterface(code, p)
			if code.Nilcheck && encoder.IsNilForMarshaler(iface) {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendMarshalJSON(ctx, code, b, iface, code.Indent+1, false)
				if err != nil {
					return nil, err
				}
				b = bb
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadMarshalJSONPtr, encoder.OpStructPtrHeadStringTagMarshalJSONPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadMarshalJSONPtr, encoder.OpStructHeadStringTagMarshalJSONPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent+1, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyMarshalJSONPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyMarshalJSONPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if p == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent+1, false)
				if err != nil {
					return nil, err
				}
				b = bb
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadMarshalText:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if code.Indirect {
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadMarshalText:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Type.Kind() == reflect.Ptr {
				if code.Indirect || code.Op == encoder.OpStructPtrHeadMarshalText {
					p = ptrToPtr(p + code.Offset)
				}
			}
			if p == 0 && code.Nilcheck {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadStringTagMarshalText:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if code.Indirect {
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadStringTagMarshalText:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Type.Kind() == reflect.Ptr {
				if code.Indirect || code.Op == encoder.OpStructPtrHeadStringTagMarshalText {
					p = ptrToPtr(p + code.Offset)
				}
			}
			if p == 0 && code.Nilcheck {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyMarshalText:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if code.Indirect {
				store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			}
			fallthrough
		case encoder.OpStructHeadOmitEmptyMarshalText:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if code.Type.Kind() == reflect.Ptr {
				if code.Indirect || code.Op == encoder.OpStructPtrHeadOmitEmptyMarshalText {
					p = ptrToPtr(p + code.Offset)
				}
			}
			if p == 0 && code.Nilcheck {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructPtrHeadMarshalTextPtr, encoder.OpStructPtrHeadStringTagMarshalTextPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadMarshalTextPtr, encoder.OpStructHeadStringTagMarshalTextPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			b = appendIndent(ctx, b, code.Indent+1)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if p == 0 {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructPtrHeadOmitEmptyMarshalTextPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			store(ctxptr, code.Idx, ptrToNPtr(p, code.PtrNum))
			fallthrough
		case encoder.OpStructHeadOmitEmptyMarshalTextPtr:
			p := load(ctxptr, code.Idx)
			if p == 0 && code.Indirect {
				if !code.AnonymousHead {
					b = appendNull(b)
					b = appendComma(b)
				}
				code = code.End.Next
				break
			}
			if code.Indirect {
				p = ptrToNPtr(p+code.Offset, code.PtrNum)
			}
			if !code.AnonymousHead {
				b = append(b, '{', '\n')
			}
			if p == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent+1)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
				b = appendComma(b)
				code = code.Next
			}
		case encoder.OpStructField:
			if !code.AnonymousKey {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
			}
			p := load(ctxptr, code.HeadIdx) + code.Offset
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldOmitEmpty:
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			if ptrToPtr(p) == 0 && code.IsNextOpPtrType {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructFieldStringTag:
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldInt:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendInt(b, ptrToUint64(p+code.Offset), code)
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyInt:
			p := load(ctxptr, code.HeadIdx)
			u64 := ptrToUint64(p + code.Offset)
			v := u64 & code.Mask
			if v != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendInt(b, u64, code)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagInt:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendInt(b, ptrToUint64(p+code.Offset), code)
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldIntPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyIntPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendInt(b, ptrToUint64(p), code)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagIntPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldUint:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendUint(b, ptrToUint64(p+code.Offset), code)
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyUint:
			p := load(ctxptr, code.HeadIdx)
			u64 := ptrToUint64(p + code.Offset)
			v := u64 & code.Mask
			if v != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendUint(b, u64, code)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagUint:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendUint(b, ptrToUint64(p+code.Offset), code)
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldUintPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyUintPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendUint(b, ptrToUint64(p), code)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagUintPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldFloat32:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendFloat32(b, ptrToFloat32(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyFloat32:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToFloat32(p + code.Offset)
			if v != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendFloat32(b, v)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagFloat32:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendFloat32(b, ptrToFloat32(p+code.Offset))
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldFloat32Ptr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendFloat32(b, ptrToFloat32(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendFloat32(b, ptrToFloat32(p))
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagFloat32Ptr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldFloat64:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			v := ptrToFloat64(p + code.Offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = appendFloat64(b, v)
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyFloat64:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToFloat64(p + code.Offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendFloat64(b, v)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagFloat64:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToFloat64(p + code.Offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendFloat64(b, v)
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldFloat64Ptr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagFloat64Ptr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldString:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendString(b, ptrToString(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyString:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToString(p + code.Offset)
			if v != "" {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendString(b, v)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagString:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			s := ptrToString(p + code.Offset)
			b = appendString(b, string(appendString([]byte{}, s)))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldStringPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendString(b, ptrToString(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyStringPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendString(b, ptrToString(p))
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagStringPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendString(b, string(appendString([]byte{}, ptrToString(p))))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldBool:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendBool(b, ptrToBool(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyBool:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToBool(p + code.Offset)
			if v {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendBool(b, v)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagBool:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendBool(b, ptrToBool(p+code.Offset))
			b = append(b, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldBoolPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendBool(b, ptrToBool(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyBoolPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendBool(b, ptrToBool(p))
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagBoolPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldBytes:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendByteSlice(b, ptrToBytes(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyBytes:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToBytes(p + code.Offset)
			if len(v) > 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendByteSlice(b, v)
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagBytes:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendByteSlice(b, ptrToBytes(p+code.Offset))
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldBytesPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendByteSlice(b, ptrToBytes(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyBytesPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendByteSlice(b, ptrToBytes(p))
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagBytesPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendByteSlice(b, ptrToBytes(p))
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldNumber:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			bb, err := appendNumber(b, ptrToNumber(p+code.Offset))
			if err != nil {
				return nil, err
			}
			b = appendComma(bb)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyNumber:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToNumber(p + code.Offset)
			if v != "" {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = appendComma(bb)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagNumber:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = append(b, '"')
			p := load(ctxptr, code.HeadIdx)
			bb, err := appendNumber(b, ptrToNumber(p+code.Offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldNumberPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyNumberPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = appendComma(bb)
			}
			code = code.Next
		case encoder.OpStructFieldStringTagNumberPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = append(bb, '"')
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldMarshalJSON, encoder.OpStructFieldStringTagMarshalJSON:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			if code.Type.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.Nilcheck {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent+1, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyMarshalJSON:
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			if code.Type.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.Nilcheck {
				code = code.NextField
				break
			}
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent+1, false)
			if err != nil {
				return nil, err
			}
			b = appendComma(bb)
			code = code.Next
		case encoder.OpStructFieldMarshalJSONPtr, encoder.OpStructFieldStringTagMarshalJSONPtr:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent+1, false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyMarshalJSONPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendMarshalJSON(ctx, code, b, ptrToInterface(code, p), code.Indent+1, false)
				if err != nil {
					return nil, err
				}
				b = appendComma(bb)
			}
			code = code.Next
		case encoder.OpStructFieldMarshalText, encoder.OpStructFieldStringTagMarshalText:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p += code.Offset
			if code.Type.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.Nilcheck {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyMarshalText:
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			if code.Type.Kind() == reflect.Ptr {
				p = ptrToPtr(p)
			}
			if p == 0 && code.Nilcheck {
				code = code.NextField
				break
			}
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
			if err != nil {
				return nil, err
			}
			b = appendComma(bb)
			code = code.Next
		case encoder.OpStructFieldMarshalTextPtr, encoder.OpStructFieldStringTagMarshalTextPtr:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructFieldOmitEmptyMarshalTextPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendMarshalText(code, b, ptrToInterface(code, p), false)
				if err != nil {
					return nil, err
				}
				b = appendComma(bb)
			}
			code = code.Next
		case encoder.OpStructFieldArray, encoder.OpStructFieldStringTagArray:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldOmitEmptyArray:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldArrayPtr, encoder.OpStructFieldStringTagArrayPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldOmitEmptyArrayPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			} else {
				code = code.NextField
			}
		case encoder.OpStructFieldSlice, encoder.OpStructFieldStringTagSlice:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldOmitEmptySlice:
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			slice := ptrToSlice(p)
			if slice.Data == nil {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructFieldSlicePtr, encoder.OpStructFieldStringTagSlicePtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldOmitEmptySlicePtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			} else {
				code = code.NextField
			}
		case encoder.OpStructFieldMap, encoder.OpStructFieldStringTagMap:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToPtr(p + code.Offset)
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldOmitEmptyMap:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToPtr(p + code.Offset)
			if p == 0 || maplen(ptrToUnsafePtr(p)) == 0 {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructFieldMapPtr, encoder.OpStructFieldStringTagMapPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToPtr(p + code.Offset)
			if p != 0 {
				p = ptrToNPtr(p, code.PtrNum)
			}
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldOmitEmptyMapPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToPtr(p + code.Offset)
			if p != 0 {
				p = ptrToNPtr(p, code.PtrNum)
			}
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			} else {
				code = code.NextField
			}
		case encoder.OpStructFieldStruct, encoder.OpStructFieldStringTagStruct:
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			code = code.Next
			store(ctxptr, code.Idx, p)
		case encoder.OpStructFieldOmitEmptyStruct:
			p := load(ctxptr, code.HeadIdx)
			p += code.Offset
			if ptrToPtr(p) == 0 && code.IsNextOpPtrType {
				code = code.NextField
			} else {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				code = code.Next
				store(ctxptr, code.Idx, p)
			}
		case encoder.OpStructAnonymousEnd:
			code = code.Next
		case encoder.OpStructEnd:
			last := len(b) - 1
			if b[last-1] == '{' {
				b[last] = '}'
				b = appendComma(b)
				code = code.Next
				break
			}
			if b[last] == '\n' {
				// to remove ',' and '\n' characters
				b = b[:len(b)-2]
			}
			b = append(b, '\n')
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, '}')
			b = appendComma(b)
			code = code.Next
		case encoder.OpStructEndInt:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendInt(b, ptrToUint64(p+code.Offset), code)
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyInt:
			p := load(ctxptr, code.HeadIdx)
			u64 := ptrToUint64(p + code.Offset)
			v := u64 & code.Mask
			if v != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendInt(b, u64, code)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagInt:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendInt(b, ptrToUint64(p+code.Offset), code)
			b = append(b, '"')
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndIntPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendInt(b, ptrToUint64(p), code)
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyIntPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendInt(b, ptrToUint64(p), code)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagIntPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendInt(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndUint:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendUint(b, ptrToUint64(p+code.Offset), code)
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyUint:
			p := load(ctxptr, code.HeadIdx)
			u64 := ptrToUint64(p + code.Offset)
			v := u64 & code.Mask
			if v != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendUint(b, u64, code)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagUint:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendUint(b, ptrToUint64(p+code.Offset), code)
			b = append(b, '"')
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndUintPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendUint(b, ptrToUint64(p), code)
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyUintPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendUint(b, ptrToUint64(p), code)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagUintPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendUint(b, ptrToUint64(p), code)
				b = append(b, '"')
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndFloat32:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendFloat32(b, ptrToFloat32(p+code.Offset))
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyFloat32:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToFloat32(p + code.Offset)
			if v != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendFloat32(b, v)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagFloat32:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendFloat32(b, ptrToFloat32(p+code.Offset))
			b = append(b, '"')
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndFloat32Ptr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendFloat32(b, ptrToFloat32(p))
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyFloat32Ptr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendFloat32(b, ptrToFloat32(p))
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagFloat32Ptr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendFloat32(b, ptrToFloat32(p))
				b = append(b, '"')
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndFloat64:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			v := ptrToFloat64(p + code.Offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = appendFloat64(b, v)
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyFloat64:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToFloat64(p + code.Offset)
			if v != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagFloat64:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToFloat64(p + code.Offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendFloat64(b, v)
			b = append(b, '"')
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndFloat64Ptr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyFloat64Ptr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagFloat64Ptr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				v := ptrToFloat64(p)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = appendFloat64(b, v)
				b = append(b, '"')
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndString:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendString(b, ptrToString(p+code.Offset))
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyString:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToString(p + code.Offset)
			if v != "" {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendString(b, v)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagString:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			s := ptrToString(p + code.Offset)
			b = appendString(b, string(appendString([]byte{}, s)))
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndStringPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendString(b, ptrToString(p))
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyStringPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendString(b, ptrToString(p))
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagStringPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendString(b, string(appendString([]byte{}, ptrToString(p))))
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndBool:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendBool(b, ptrToBool(p+code.Offset))
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyBool:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToBool(p + code.Offset)
			if v {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendBool(b, v)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagBool:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ', '"')
			b = appendBool(b, ptrToBool(p+code.Offset))
			b = append(b, '"')
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndBoolPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendBool(b, ptrToBool(p))
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyBoolPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendBool(b, ptrToBool(p))
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagBoolPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				b = appendBool(b, ptrToBool(p))
				b = append(b, '"')
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndBytes:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			b = appendByteSlice(b, ptrToBytes(p+code.Offset))
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyBytes:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToBytes(p + code.Offset)
			if len(v) > 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendByteSlice(b, v)
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagBytes:
			p := load(ctxptr, code.HeadIdx)
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = appendByteSlice(b, ptrToBytes(p+code.Offset))
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndBytesPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendByteSlice(b, ptrToBytes(p))
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyBytesPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				b = appendByteSlice(b, ptrToBytes(p))
				b = appendStructEnd(ctx, b, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagBytesPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = appendByteSlice(b, ptrToBytes(p))
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndNumber:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			bb, err := appendNumber(b, ptrToNumber(p+code.Offset))
			if err != nil {
				return nil, err
			}
			b = appendStructEnd(ctx, bb, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyNumber:
			p := load(ctxptr, code.HeadIdx)
			v := ptrToNumber(p + code.Offset)
			if v != "" {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendNumber(b, v)
				if err != nil {
					return nil, err
				}
				b = appendStructEnd(ctx, bb, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagNumber:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			b = append(b, '"')
			p := load(ctxptr, code.HeadIdx)
			bb, err := appendNumber(b, ptrToNumber(p+code.Offset))
			if err != nil {
				return nil, err
			}
			b = append(bb, '"')
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndNumberPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = bb
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpStructEndOmitEmptyNumberPtr:
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p != 0 {
				b = appendIndent(ctx, b, code.Indent)
				b = append(b, code.Key...)
				b = append(b, ' ')
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = appendStructEnd(ctx, bb, code.Indent-1)
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
					b = appendIndent(ctx, b, code.Indent-1)
					b = append(b, '}')
				}
				b = appendComma(b)
			}
			code = code.Next
		case encoder.OpStructEndStringTagNumberPtr:
			b = appendIndent(ctx, b, code.Indent)
			b = append(b, code.Key...)
			b = append(b, ' ')
			p := load(ctxptr, code.HeadIdx)
			p = ptrToNPtr(p+code.Offset, code.PtrNum)
			if p == 0 {
				b = appendNull(b)
			} else {
				b = append(b, '"')
				bb, err := appendNumber(b, ptrToNumber(p))
				if err != nil {
					return nil, err
				}
				b = append(bb, '"')
			}
			b = appendStructEnd(ctx, b, code.Indent-1)
			code = code.Next
		case encoder.OpEnd:
			goto END
		}
	}
END:
	return b, nil
}

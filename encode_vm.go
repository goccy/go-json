package json

import (
	"bytes"
	"encoding"
	"fmt"
	"math"
	"reflect"
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

func (e *Encoder) run(ctx *encodeRuntimeContext, b []byte, code *opcode) ([]byte, error) {
	recursiveLevel := 0
	var seenPtr map[uintptr]struct{}
	ptrOffset := uintptr(0)
	ctxptr := ctx.ptr()

	for {
		switch code.op {
		default:
			return nil, fmt.Errorf("failed to handle opcode. doesn't implement %s", code.op)
		case opPtr, opPtrIndent:
			ptr := load(ctxptr, code.idx)
			code = code.next
			store(ctxptr, code.idx, e.ptrToPtr(ptr))
		case opInt:
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opIntIndent:
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt8:
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opInt8Indent:
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt16:
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opInt16Indent:
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt32:
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opInt32Indent:
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opInt64:
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opInt64Indent:
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opUint:
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opUintIndent:
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint8:
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opUint8Indent:
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint16:
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opUint16Indent:
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint32:
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = encodeComma(b)
			code = code.next
		case opUint32Indent:
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = encodeIndentComma(b)
			code = code.next
		case opUint64:
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opUint64Indent:
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opIntString:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opIntStringIndent:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt8String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opInt8StringIndent:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt16String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opInt16StringIndent:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt32String:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opInt32StringIndent:
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opInt64String:
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opInt64StringIndent:
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUintString:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUintStringIndent:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint8String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUint8StringIndent:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint16String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUint16StringIndent:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint32String:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUint32StringIndent:
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(load(ctxptr, code.idx))))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opUint64String:
			b = append(b, '"')
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opUint64StringIndent:
			b = append(b, '"')
			b = appendUint(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opFloat32:
			b = encodeFloat32(b, e.ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opFloat32Indent:
			b = encodeFloat32(b, e.ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opFloat64:
			v := e.ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opFloat64Indent:
			v := e.ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opString:
			b = encodeNoEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opEscapedString:
			b = encodeEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opStringIndent:
			b = encodeNoEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opEscapedStringIndent:
			b = encodeEscapedString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBool:
			b = encodeBool(b, e.ptrToBool(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opBoolIndent:
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
			b = encodeComma(b)
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
			if e.enabledIndent {
				c = toIndent(c)
			}
			if e.enabledHTMLEscape {
				c = toEscaped(c)
			}
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
			if e.enabledIndent {
				c = toIndent(c)
			}
			if e.enabledHTMLEscape {
				c = toEscaped(c)
			}
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
		case opInterfaceEnd, opInterfaceEndIndent:
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
				if e.enabledHTMLEscape {
					b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				} else {
					b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
				}
				b = encodeIndentComma(b)
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
		case opStructFieldPtrAnonymousHeadRecursive, opStructEscapedFieldPtrAnonymousHeadRecursive:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadRecursive, opStructEscapedFieldAnonymousHeadRecursive:
			fallthrough
		case opStructFieldRecursive, opStructEscapedFieldRecursive:
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
			c := code.jmp.code
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
		case opStructFieldPtrHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = append(b, code.key...)
				}
				p := ptr + code.offset
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldHeadOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{')
			if !code.anonymousKey {
				b = append(b, code.key...)
			}
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
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
		case opStructFieldAnonymousHead, opStructEscapedFieldAnonymousHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				code = code.next
				store(ctxptr, code.idx, ptr)
			}
		case opStructFieldPtrHeadInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadIntOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructFieldHeadIntOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt(p)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadInt {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadIntOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructEscapedFieldHeadIntOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt(p)))
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
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				p = e.ptrToPtr(p)
				if p == 0 {
					b = encodeNull(b)
				} else {
					b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
				}
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadIntPtrOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadIntPtrOnly:
			p := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.key...)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p+code.offset)))
			}
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldPtrHeadIntPtr:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
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
		case opStructFieldPtrAnonymousHeadInt:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadInt8:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt8 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt8:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadInt8 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadInt16:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt16 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt16:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadInt16 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadInt32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadInt32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadInt64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadInt64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadUint {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadUint8:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint8 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint8:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadUint8 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadUint16:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint16 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint16:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadUint16 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadUint32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadUint32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadUint64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadUint64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadFloat32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadFloat32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadFloat32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadFloat64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, '{')
				b = append(b, code.key...)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructEscapedFieldHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadFloat64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
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
		case opStructFieldPtrAnonymousHeadFloat64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.key...)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadString {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, e.ptrToString(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructFieldHeadStringOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.key...)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
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
				if code.op == opStructEscapedFieldPtrHeadEscapedString {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.escapedKey...)
				b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadEscapedStringOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructEscapedFieldHeadEscapedStringOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrAnonymousHeadString:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, e.ptrToString(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
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

		case opStructEscapedFieldPtrAnonymousHeadEscapedStringOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				code = code.end.next
				break
			}
			fallthrough
		case opStructEscapedFieldAnonymousHeadEscapedStringOnly:
			p := load(ctxptr, code.idx)
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, e.ptrToString(p+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrHeadBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadBool {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadBoolOnly:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			fallthrough
		case opStructFieldHeadBoolOnly:
			ptr := load(ctxptr, code.idx)
			b = append(b, '{')
			b = append(b, code.key...)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
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
		case opStructFieldPtrAnonymousHeadBool:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadBytes {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
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
		case opStructFieldPtrAnonymousHeadBytes:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadArray:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadArray:
			ptr := load(ctxptr, code.idx) + code.offset
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadArray {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '[', ']', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = append(b, code.key...)
				}
				code = code.next
				store(ctxptr, code.idx, ptr)
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
		case opStructFieldPtrAnonymousHeadArray:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadArray:
			ptr := load(ctxptr, code.idx) + code.offset
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				store(ctxptr, code.idx, ptr)
				code = code.next
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
		case opStructFieldPtrHeadSlice:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadSlice:
			ptr := load(ctxptr, code.idx)
			p := ptr + code.offset
			if p == 0 {
				if code.op == opStructFieldPtrHeadSlice {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '[', ']', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = append(b, code.key...)
				}
				code = code.next
				store(ctxptr, code.idx, p)
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
		case opStructFieldPtrAnonymousHeadSlice:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadSlice:
			ptr := load(ctxptr, code.idx)
			p := ptr + code.offset
			if p == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				store(ctxptr, code.idx, p)
				code = code.next
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
		case opStructFieldPtrHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadMarshalJSON:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
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
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructFieldPtrAnonymousHeadMarshalText:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructFieldPtrHeadIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
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
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, ptr)
			}
		case opStructEscapedFieldPtrHeadIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
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
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, ptr)
			}
		case opStructFieldPtrHeadIntIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadIntIndent {
					b = e.encodeIndent(b, code.indent)
					b = encodeNull(b)
					b = encodeIndentComma(b)
				} else {
					b = append(b, '{', '}', ',', '\n')
				}
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
		case opStructEscapedFieldPtrHeadIntIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructEscapedFieldPtrHeadIntIndent {
					b = e.encodeIndent(b, code.indent)
					b = encodeNull(b)
					b = encodeIndentComma(b)
				} else {
					b = append(b, '{', '}', ',', '\n')
				}
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
		case opStructFieldPtrHeadInt8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = encodeIndentComma(b)
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt8(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = encodeIndentComma(b)
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt8(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
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
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt16(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
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
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = appendInt(b, int64(e.ptrToInt16(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt32Indent:
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
				b = appendInt(b, int64(e.ptrToInt32(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt32Indent:
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
				b = appendInt(b, int64(e.ptrToInt32(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt64Indent:
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
				b = appendInt(b, int64(e.ptrToInt64(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadInt64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadInt64Indent:
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
				b = appendInt(b, int64(e.ptrToInt64(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUintIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUintIndent:
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
				b = appendUint(b, uint64(e.ptrToUint(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUintIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUintIndent:
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
				b = appendUint(b, uint64(e.ptrToUint(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint8Indent:
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
				b = appendUint(b, uint64(e.ptrToUint8(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint8Indent:
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
				b = appendUint(b, uint64(e.ptrToUint8(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint16Indent:
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
				b = appendUint(b, uint64(e.ptrToUint16(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint16Indent:
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
				b = appendUint(b, uint64(e.ptrToUint16(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint32Indent:
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
				b = appendUint(b, uint64(e.ptrToUint32(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint32Indent:
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
				b = appendUint(b, uint64(e.ptrToUint32(ptr)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint64Indent:
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
				b = appendUint(b, e.ptrToUint64(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadUint64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadUint64Indent:
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
				b = appendUint(b, e.ptrToUint64(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadFloat32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat32Indent:
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
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadFloat32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat32Indent:
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
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadFloat64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.key...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructEscapedFieldPtrHeadFloat64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructEscapedFieldHeadFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = append(b, code.escapedKey...)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmpty:
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
					b = append(b, code.key...)
					code = code.next
					store(ctxptr, code.idx, p)
				}
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
		case opStructFieldPtrAnonymousHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				p := ptr + code.offset
				if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt:
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
					b = append(b, code.key...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
					code = code.next
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
		case opStructFieldPtrAnonymousHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
					b = appendInt(b, int64(v))
					b = encodeComma(b)
					code = code.next
				}
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
		case opStructFieldPtrHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt8:
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
					b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt16:
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
					b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt32:
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
					b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt64:
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
					b = append(b, code.key...)
					b = appendInt(b, v)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint:
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
					b = append(b, code.key...)
					b = appendUint(b, uint64(v))
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
		case opStructFieldPtrAnonymousHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint8:
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
					b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint16:
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
					b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint32:
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
					b = append(b, code.key...)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint64:
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
					b = append(b, code.key...)
					b = appendUint(b, v)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32:
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
					b = append(b, code.key...)
					b = encodeFloat32(b, v)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64:
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
					b = append(b, code.key...)
					b = encodeFloat64(b, v)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat64:
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
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyString:
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
					b = append(b, code.key...)
					b = encodeNoEscapedString(b, v)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToString(ptr + code.offset)
				if v == "" {
					code = code.nextField
				} else {
					b = append(b, code.key...)
					b = encodeNoEscapedString(b, v)
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
		case opStructFieldPtrHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBool:
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
					b = append(b, code.key...)
					b = encodeBool(b, v)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToBool(ptr + code.offset)
				if !v {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBytes:
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
					b = append(b, code.key...)
					b = encodeByteSlice(b, v)
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
		case opStructFieldPtrAnonymousHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToBytes(ptr + code.offset)
				if len(v) == 0 {
					code = code.nextField
				} else {
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalJSON:
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
						b = append(b, code.key...)
						b = append(b, buf.Bytes()...)
						b = encodeComma(b)
						code = code.next
					}
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
		case opStructFieldPtrAnonymousHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMarshalJSON:
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
						b = append(b, code.key...)
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
		case opStructFieldPtrHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
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
					b = append(b, code.key...)
					b = encodeEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
					b = encodeComma(b)
					code = code.next
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
		case opStructFieldPtrAnonymousHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMarshalText:
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
					b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				p := ptr + code.offset
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, p)
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
		case opStructFieldPtrAnonymousHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				code = code.next
				store(ctxptr, code.idx, ptr+code.offset)
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
		case opStructFieldPtrHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
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
		case opStructFieldPtrAnonymousHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
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
		case opStructFieldPtrAnonymousHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
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
		case opStructFieldPtrAnonymousHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
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
		case opStructFieldPtrAnonymousHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
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

		case opStructFieldPtrAnonymousHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendInt(b, e.ptrToInt64(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
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
		case opStructFieldPtrAnonymousHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
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
		case opStructFieldPtrAnonymousHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
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
		case opStructFieldPtrAnonymousHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
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
		case opStructFieldPtrAnonymousHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = appendUint(b, e.ptrToUint64(ptr+code.offset))
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
		case opStructFieldPtrAnonymousHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
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
		case opStructFieldPtrAnonymousHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat64:
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
				b = append(b, code.key...)
				b = append(b, '"')
				b = encodeFloat64(b, v)
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
		case opStructFieldPtrAnonymousHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				s := e.ptrToString(ptr + code.offset)
				b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
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
		case opStructFieldPtrAnonymousHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				s := string(encodeNoEscapedString([]byte{}, (e.ptrToString(ptr + code.offset))))
				b = encodeNoEscapedString(b, s)
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
		case opStructFieldPtrHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = append(b, '"')
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = append(b, '"')
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
		case opStructFieldPtrAnonymousHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
				b = append(b, '"')
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = append(b, '"')
				b = encodeComma(b)
				code = code.next
			}
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
		case opStructFieldPtrHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = append(b, code.key...)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
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
		case opStructFieldPtrAnonymousHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = append(b, code.key...)
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
		case opStructFieldPtrHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
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
					b = append(b, code.key...)
					b = append(b, '"', '"')
					b = encodeComma(b)
					code = code.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
						return nil, err
					}
					b = encodeNoEscapedString(b, buf.String())
					b = encodeComma(b)
					code = code.next
				}
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
		case opStructFieldPtrAnonymousHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagMarshalJSON:
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
					b = append(b, code.key...)
					b = append(b, '"', '"')
					b = encodeComma(b)
					code = code.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
						return nil, err
					}
					b = append(b, code.key...)
					b = encodeNoEscapedString(b, buf.String())
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
		case opStructFieldPtrHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
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
				p := e.ptrToUnsafePtr(ptr)
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
		case opStructFieldPtrAnonymousHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagMarshalText:
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
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructField:
			if !code.anonymousKey {
				b = append(b, code.key...)
			}
			ptr := load(ctxptr, code.headIdx) + code.offset
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructEscapedField:
			if !code.anonymousKey {
				b = append(b, code.escapedKey...)
			}
			ptr := load(ctxptr, code.headIdx) + code.offset
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldIntPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p)))
			}
			b = encodeComma(b)
			code = code.next
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
		case opStructFieldIntNPtr:
			b = append(b, code.key...)
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
		case opStructFieldInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt8Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p)))
			}
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
		case opStructFieldInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt16Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p)))
			}
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
		case opStructFieldInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p)))
			}
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
		case opStructFieldInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt64Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p))
			}
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
		case opStructFieldInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldUintPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p)))
			}
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
		case opStructFieldUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint8Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p)))
			}
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
		case opStructFieldUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint16Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p)))
			}
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
		case opStructFieldUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p)))
			}
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
		case opStructFieldUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint64Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p))
			}
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
		case opStructFieldUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
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
		case opStructFieldFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat64Ptr:
			b = append(b, code.key...)
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
		case opStructFieldFloat64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			v := e.ptrToFloat64(ptr + code.offset)
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
		case opStructFieldStringPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, e.ptrToString(p))
			}
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
		case opStructFieldString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, e.ptrToString(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldBoolPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, e.ptrToBool(p))
			}
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
		case opStructFieldBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldBytes:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldBytes:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
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
		case opStructFieldMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructFieldArray:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldArray:
			b = append(b, code.escapedKey...)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldSlice:
			b = append(b, code.key...)
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
		case opStructFieldMap:
			b = append(b, code.key...)
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
		case opStructFieldMapLoad:
			b = append(b, code.key...)
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
		case opStructFieldStruct:
			b = append(b, code.key...)
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
		case opStructFieldIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldIntIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldInt8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldInt16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldInt32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldInt64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldUintIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldUint8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldUint16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldUint32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldUint64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldFloat32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructFieldStringIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeNoEscapedString(b, e.ptrToString(ptr+code.offset))
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
		case opStructFieldBoolIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
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
		case opStructFieldBytesIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
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
		case opStructFieldOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = appendInt(b, int64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructEscapedFieldOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = append(b, code.escapedKey...)
				b = appendInt(b, int64(v))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructFieldOmitEmptyInt16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructFieldOmitEmptyInt32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructFieldOmitEmptyInt64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = appendInt(b, v)
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
		case opStructFieldOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = appendUint(b, uint64(v))
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
		case opStructFieldOmitEmptyUint8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructFieldOmitEmptyUint16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructFieldOmitEmptyUint32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructFieldOmitEmptyUint64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = appendUint(b, v)
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
		case opStructFieldOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = encodeFloat32(b, v)
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
		case opStructFieldOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.key...)
				b = encodeFloat64(b, v)
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
		case opStructFieldOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, v)
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
		case opStructFieldOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = append(b, code.key...)
				b = encodeBool(b, v)
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
		case opStructFieldOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = append(b, code.key...)
				b = encodeByteSlice(b, v)
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
		case opStructFieldOmitEmptyMarshalJSON:
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
				b = append(b, code.key...)
				b = append(b, buf.Bytes()...)
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
		case opStructFieldOmitEmptyMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
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
		case opStructFieldOmitEmptyArray, opStructEscapedFieldOmitEmptyArray:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			array := e.ptrToSlice(p)
			if p == 0 || uintptr(array.data) == 0 {
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructFieldOmitEmptySlice, opStructEscapedFieldOmitEmptySlice:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructFieldOmitEmptyMap, opStructEscapedFieldOmitEmptyMap:
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
		case opStructFieldOmitEmptyMapLoad, opStructEscapedFieldOmitEmptyMapLoad:
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
		case opStructFieldStringTag:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = append(b, code.key...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldStringTag:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = append(b, code.escapedKey...)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
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
		case opStructFieldStringTagInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
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
		case opStructFieldStringTagInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
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
		case opStructFieldStringTagInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
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
		case opStructFieldStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
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
		case opStructFieldStringTagUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
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
		case opStructFieldStringTagUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
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
		case opStructFieldStringTagUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
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
		case opStructFieldStringTagUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint64(ptr+code.offset)))
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
		case opStructFieldStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
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
		case opStructFieldStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
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
		case opStructFieldStringTagString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			s := e.ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			s := e.ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
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
		case opStructFieldStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			b = append(b, code.key...)
			b = encodeByteSlice(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructEscapedFieldStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagMarshalJSON:
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
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, buf.String())
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
		case opStructFieldStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructFieldStringTagIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructEscapedFieldStringTagIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagIntIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagUintIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructEscapedFieldStringTagFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
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
		case opStructFieldStringTagStringIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			s := e.ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
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
		case opStructFieldStringTagBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
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
		case opStructFieldStringTagBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
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
		case opStructEnd, opStructEscapedEnd:
			last := len(b) - 1
			if b[last] == ',' {
				b[last] = '}'
			} else {
				b = append(b, '}')
			}
			b = encodeComma(b)
			code = code.next
		case opStructAnonymousEnd, opStructAnonymousEndIndent:
			code = code.next
		case opStructEndIndent, opStructEscapedEndIndent:
			last := len(b) - 1
			if b[last] == '\n' {
				// to remove ',' and '\n' characters
				b = b[:len(b)-2]
			}
			b = append(b, '\n')
			b = e.encodeIndent(b, code.indent)
			b = append(b, '}')
			b = encodeIndentComma(b)
			code = code.next
		case opStructEndIntPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt(p)))
			}
			b = appendStructEnd(b)
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
		case opStructEndIntNPtr:
			b = append(b, code.key...)
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
		case opStructEndInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndInt8Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt8(p)))
			}
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
		case opStructEndInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndInt16Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt16(p)))
			}
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
		case opStructEndInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndInt32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, int64(e.ptrToInt32(p)))
			}
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
		case opStructEndInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndInt64Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendInt(b, e.ptrToInt64(p))
			}
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
		case opStructEndInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUintPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint(p)))
			}
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
		case opStructEndUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUint8Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint8(p)))
			}
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
		case opStructEndUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUint16Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint16(p)))
			}
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
		case opStructEndUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUint32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, uint64(e.ptrToUint32(p)))
			}
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
		case opStructEndUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndUint64Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = appendUint(b, e.ptrToUint64(p))
			}
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
		case opStructEndUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat32Ptr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
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
		case opStructEndFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndFloat64Ptr:
			b = append(b, code.key...)
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
		case opStructEndFloat64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			v := e.ptrToFloat64(ptr + code.offset)
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
		case opStructEndStringPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeNoEscapedString(b, e.ptrToString(p))
			}
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
		case opStructEndString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, e.ptrToString(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeEscapedString(b, e.ptrToString(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndBoolPtr:
			b = append(b, code.key...)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, e.ptrToBool(p))
			}
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
		case opStructEndBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndBytes:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndBytes:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
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
		case opStructEndMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructEndIntIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = e.appendStructEndIndent(b, code.indent-1)
			code = code.next
		case opStructEscapedEndIntIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
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
		case opStructEscapedEndInt8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
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
		case opStructEscapedEndInt16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
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
		case opStructEscapedEndInt32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
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
		case opStructEscapedEndInt64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
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
		case opStructEscapedEndUintIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
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
		case opStructEscapedEndUint8Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
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
		case opStructEscapedEndUint16Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
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
		case opStructEscapedEndUint32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
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
		case opStructEscapedEndUint64Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
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
		case opStructEscapedEndFloat32Indent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.escapedKey...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
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
		case opStructEndStringIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeNoEscapedString(b, e.ptrToString(ptr+code.offset))
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
		case opStructEndBoolIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
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
		case opStructEndBytesIndent:
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
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
		case opStructEndOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = appendInt(b, int64(v))
			}
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
		case opStructEndOmitEmptyInt8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructEndOmitEmptyInt16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructEndOmitEmptyInt32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructEndOmitEmptyInt64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = appendInt(b, v)
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
		case opStructEndOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = appendUint(b, uint64(v))
				b = appendStructEnd(b)
			}
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
		case opStructEndOmitEmptyUint8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructEndOmitEmptyUint16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructEndOmitEmptyUint32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
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
		case opStructEndOmitEmptyUint64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = appendUint(b, v)
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
		case opStructEndOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = append(b, code.key...)
				b = encodeFloat32(b, v)
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
		case opStructEndOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, code.key...)
				b = encodeFloat64(b, v)
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
		case opStructEndOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, v)
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
		case opStructEndOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = append(b, code.key...)
				b = encodeBool(b, v)
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
		case opStructEndOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = append(b, code.key...)
				b = encodeByteSlice(b, v)
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
		case opStructEndOmitEmptyMarshalJSON:
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
				b = append(b, code.key...)
				b = append(b, buf.Bytes()...)
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
		case opStructEndOmitEmptyMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			if v != nil {
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = append(b, code.key...)
				b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructEndStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
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
		case opStructEndStringTagInt8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
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
		case opStructEndStringTagInt16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
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
		case opStructEndStringTagInt32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
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
		case opStructEndStringTagInt64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
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
		case opStructEndStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
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
		case opStructEndStringTagUint8:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
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
		case opStructEndStringTagUint16:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
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
		case opStructEndStringTagUint32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
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
		case opStructEndStringTagUint64:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = appendUint(b, uint64(e.ptrToUint64(ptr+code.offset)))
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
		case opStructEndStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
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
		case opStructEndStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeFloat64(b, v)
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
		case opStructEndStringTagString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			s := e.ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagEscapedString:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.escapedKey...)
			s := e.ptrToString(ptr + code.offset)
			b = encodeEscapedString(b, string(encodeEscapedString([]byte{}, s)))
			b = appendStructEnd(b)
			code = code.next
		case opStructEndStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = append(b, code.key...)
			b = append(b, '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
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
		case opStructEndStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			b = append(b, code.key...)
			b = encodeByteSlice(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEscapedEndStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			b = append(b, code.escapedKey...)
			b = encodeByteSlice(b, v)
			b = appendStructEnd(b)
			code = code.next
		case opStructEndStringTagMarshalJSON:
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
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, buf.String())
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
		case opStructEndStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = append(b, code.key...)
			b = encodeNoEscapedString(b, *(*string)(unsafe.Pointer(&bytes)))
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
		case opStructEndStringTagIntIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt(ptr+code.offset)))
			b = append(b, '"')
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
		case opStructEndStringTagInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt8(ptr+code.offset)))
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
		case opStructEndStringTagInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt16(ptr+code.offset)))
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
		case opStructEndStringTagInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, int64(e.ptrToInt32(ptr+code.offset)))
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
		case opStructEndStringTagInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendInt(b, e.ptrToInt64(ptr+code.offset))
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
		case opStructEndStringTagUintIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint(ptr+code.offset)))
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
		case opStructEndStringTagUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint8(ptr+code.offset)))
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
		case opStructEndStringTagUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint16(ptr+code.offset)))
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
		case opStructEndStringTagUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, uint64(e.ptrToUint32(ptr+code.offset)))
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
		case opStructEndStringTagUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = appendUint(b, e.ptrToUint64(ptr+code.offset))
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
		case opStructEndStringTagFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
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
		case opStructEndStringTagStringIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			s := e.ptrToString(ptr + code.offset)
			b = encodeNoEscapedString(b, string(encodeNoEscapedString([]byte{}, s)))
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
		case opStructEndStringTagBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ', '"')
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = append(b, '"')
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
		case opStructEndStringTagBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = append(b, code.key...)
			b = append(b, ' ')
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
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
			b = encodeEscapedString(b, buf.String())
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

func (e *Encoder) ptrToInt(p uintptr) int            { return **(**int)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToInt8(p uintptr) int8          { return **(**int8)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToInt16(p uintptr) int16        { return **(**int16)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToInt32(p uintptr) int32        { return **(**int32)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToInt64(p uintptr) int64        { return **(**int64)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint(p uintptr) uint          { return **(**uint)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint8(p uintptr) uint8        { return **(**uint8)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint16(p uintptr) uint16      { return **(**uint16)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint32(p uintptr) uint32      { return **(**uint32)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint64(p uintptr) uint64      { return **(**uint64)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToFloat32(p uintptr) float32    { return **(**float32)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToFloat64(p uintptr) float64    { return **(**float64)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToBool(p uintptr) bool          { return **(**bool)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToByte(p uintptr) byte          { return **(**byte)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToBytes(p uintptr) []byte       { return **(**[]byte)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToString(p uintptr) string      { return **(**string)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToSlice(p uintptr) *sliceHeader { return *(**sliceHeader)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToPtr(p uintptr) uintptr {
	return uintptr(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
}
func (e *Encoder) ptrToUnsafePtr(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}
func (e *Encoder) ptrToInterface(code *opcode, p uintptr) interface{} {
	return *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: code.typ,
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&p)),
	}))
}

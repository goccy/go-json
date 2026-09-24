package decoder

import (
	"bytes"
	"encoding"
	"encoding/json"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

type interfaceDecoder struct {
	typ           reflect.Type
	structName    string
	fieldName     string
	sliceDecoder  *sliceDecoder
	mapDecoder    *mapDecoder
	floatDecoder  *floatDecoder
	numberDecoder *numberDecoder
	stringDecoder *stringDecoder
}

func newEmptyInterfaceDecoder(structName, fieldName string) *interfaceDecoder {
	ifaceDecoder := &interfaceDecoder{
		typ:        emptyInterfaceType,
		structName: structName,
		fieldName:  fieldName,
		floatDecoder: newFloatDecoder(structName, fieldName, func(p unsafe.Pointer, v float64) {
			*(*any)(p) = v
		}),
		numberDecoder: newNumberDecoder(structName, fieldName, func(p unsafe.Pointer, v json.Number) {
			*(*any)(p) = v
		}),
		stringDecoder: newStringDecoder(structName, fieldName),
	}
	ifaceDecoder.sliceDecoder = newSliceDecoder(
		ifaceDecoder,
		emptyInterfaceType,
		emptyInterfaceType.Size(),
		structName, fieldName,
	)
	ifaceDecoder.mapDecoder = newMapDecoder(
		interfaceMapType,
		stringType,
		ifaceDecoder.stringDecoder,
		interfaceMapType.Elem(),
		ifaceDecoder,
		structName,
		fieldName,
	)
	return ifaceDecoder
}

func newInterfaceDecoder(typ reflect.Type, structName, fieldName string) *interfaceDecoder {
	emptyIfaceDecoder := newEmptyInterfaceDecoder(structName, fieldName)
	stringDecoder := newStringDecoder(structName, fieldName)
	return &interfaceDecoder{
		typ:        typ,
		structName: structName,
		fieldName:  fieldName,
		sliceDecoder: newSliceDecoder(
			emptyIfaceDecoder,
			emptyInterfaceType,
			emptyInterfaceType.Size(),
			structName, fieldName,
		),
		mapDecoder: newMapDecoder(
			interfaceMapType,
			stringType,
			stringDecoder,
			interfaceMapType.Elem(),
			emptyIfaceDecoder,
			structName,
			fieldName,
		),
		floatDecoder: newFloatDecoder(structName, fieldName, func(p unsafe.Pointer, v float64) {
			*(*any)(p) = v
		}),
		numberDecoder: newNumberDecoder(structName, fieldName, func(p unsafe.Pointer, v json.Number) {
			*(*any)(p) = v
		}),
		stringDecoder: stringDecoder,
	}
}

var (
	emptyInterfaceType = reflect.TypeOf((*any)(nil)).Elem()
	EmptyInterfaceType = emptyInterfaceType
	interfaceMapType   = reflect.TypeOf((*map[string]any)(nil)).Elem()
	stringType         = reflect.TypeOf("")
)

func decodeUnmarshaler(buf []byte, cursor, depth int64, unmarshaler json.Unmarshaler) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor, depth)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	dst := make([]byte, len(src))
	copy(dst, src)

	if err := unmarshaler.UnmarshalJSON(dst); err != nil {
		return 0, err
	}
	return end, nil
}

func decodeUnmarshalerContext(ctx *RuntimeContext, buf []byte, cursor, depth int64, unmarshaler unmarshalerContext) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor, depth)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	dst := make([]byte, len(src))
	copy(dst, src)

	if err := unmarshaler.UnmarshalJSON(ctx.Option.Context, dst); err != nil {
		return 0, err
	}
	return end, nil
}

func decodeTextUnmarshaler(buf []byte, cursor, depth int64, unmarshaler encoding.TextUnmarshaler, p unsafe.Pointer) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor, depth)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	if bytes.Equal(src, nullbytes) {
		*(*unsafe.Pointer)(p) = nil
		return end, nil
	}
	if s, ok := unquoteBytes(src); ok {
		src = s
	}
	if err := unmarshaler.UnmarshalText(src); err != nil {
		return 0, err
	}
	return end, nil
}

type emptyInterface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

func (d *interfaceDecoder) errUnmarshalType(typ reflect.Type, offset int64) *errors.UnmarshalTypeError {
	return &errors.UnmarshalTypeError{
		Value:  typ.String(),
		Type:   typ,
		Offset: offset,
		Struct: d.structName,
		Field:  d.fieldName,
	}
}

func (d *interfaceDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	runtimeInterfaceValue := *(*any)(unsafe.Pointer(&emptyInterface{
		typ: runtime.TypePtr(d.typ),
		ptr: p,
	}))
	rv := reflect.ValueOf(runtimeInterfaceValue)
	if rv.NumMethod() > 0 && rv.CanInterface() {
		if u, ok := rv.Interface().(unmarshalerContext); ok {
			return decodeUnmarshalerContext(ctx, buf, cursor, depth, u)
		}
		if u, ok := rv.Interface().(json.Unmarshaler); ok {
			return decodeUnmarshaler(buf, cursor, depth, u)
		}
		if u, ok := rv.Interface().(encoding.TextUnmarshaler); ok {
			return decodeTextUnmarshaler(buf, cursor, depth, u, p)
		}
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] == 'n' {
			if err := validateNull(buf, cursor); err != nil {
				return 0, err
			}
			cursor += 4
			**(**any)(unsafe.Pointer(&p)) = nil
			return cursor, nil
		}
		return 0, d.errUnmarshalType(rv.Type(), cursor)
	}

	iface := rv.Interface()
	ifaceHeader := (*emptyInterface)(unsafe.Pointer(&iface))
	typ := reflect.TypeOf(iface)
	if ifaceHeader.ptr == nil || d.typ == typ || typ == nil {
		// concrete type is empty interface
		return d.decodeEmptyInterface(ctx, cursor, depth, p)
	}
	if typ.Kind() == reflect.Ptr && typ.Elem() == d.typ || typ.Kind() != reflect.Ptr {
		return d.decodeEmptyInterface(ctx, cursor, depth, p)
	}
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == 'n' {
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 4
		**(**any)(unsafe.Pointer(&p)) = nil
		return cursor, nil
	}
	decoder, err := ctx.DecoderOf(runtime.TypePtr(typ))
	if err != nil {
		return 0, err
	}
	return decoder.Decode(ctx, cursor, depth, ifaceHeader.ptr)
}

func (d *interfaceDecoder) decodeEmptyInterface(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case '{':
		depth++
		if depth > maxDecodeNestingDepth {
			return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
		}
		m := map[string]any{}
		cursor, err := decodeStringAnyMap(ctx, d, m, cursor, depth)
		if err != nil {
			return 0, err
		}
		**(**any)(unsafe.Pointer(&p)) = m
		return cursor, nil
	case '[':
		return d.decodeAnySlice(ctx, cursor, depth, p)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if (ctx.Option.Flags & UseNumberOption) != 0 {
			return d.numberDecoder.Decode(ctx, cursor, depth, p)
		}
		num, c, err := d.floatDecoder.decodeByte(buf, cursor)
		if err != nil {
			return 0, err
		}
		if !validEndNumberChar[buf[c]] {
			return 0, errors.ErrUnexpectedEndOfJSON("float", c)
		}
		f, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&num)), 64)
		if err != nil {
			return 0, errors.ErrSyntax(err.Error(), c)
		}
		**(**any)(unsafe.Pointer(&p)) = ctx.boxFloat(f)
		return c, nil
	case '"':
		s, c, _, err := d.stringDecoder.decodeString(ctx, cursor)
		if err != nil {
			return 0, err
		}
		**(**any)(unsafe.Pointer(&p)) = ctx.boxString(s)
		return c, nil
	case 't':
		if err := validateTrue(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 4
		**(**any)(unsafe.Pointer(&p)) = true
		return cursor, nil
	case 'f':
		if err := validateFalse(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 5
		**(**any)(unsafe.Pointer(&p)) = false
		return cursor, nil
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 4
		**(**any)(unsafe.Pointer(&p)) = nil
		return cursor, nil
	}
	return cursor, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
}

// decodeStringAnyMap decodes the object at cursor into m, a map[string]interface{}, by the assignment of Go:
// the values are decoded as interface{} by d. depth counts the object already.
func decodeStringAnyMap(ctx *RuntimeContext, d *interfaceDecoder, m map[string]any, cursor, depth int64) (int64, error) {
	buf := ctx.Buf
	cursor++ // '{'
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == '}' {
		return cursor + 1, nil
	}
	for {
		key, c, ok, err := d.stringDecoder.decodeString(ctx, cursor)
		if err != nil {
			return 0, err
		}
		if !ok {
			// null is not a key
			return 0, errors.ErrSyntax("invalid character 'n' looking for beginning of object key string", skipWhiteSpace(buf, cursor)+1)
		}
		cursor = skipWhiteSpace(buf, c)
		if buf[cursor] != ':' {
			return 0, errors.ErrExpected("colon after object key", cursor)
		}
		cursor++
		c, err = d.decodeEmptyInterface(ctx, cursor, depth, unsafe.Pointer(&ctx.slot))
		if err != nil {
			ctx.slot = nil
			return 0, err
		}
		m[key] = ctx.slot
		ctx.slot = nil
		cursor = skipWhiteSpace(buf, c)
		switch buf[cursor] {
		case '}':
			return cursor + 1, nil
		case ',':
			cursor++
		default:
			return 0, errors.ErrExpected("comma after object value", cursor)
		}
	}
}

// decodeAnySlice decodes the array at cursor as a []interface{}: the elements are pushed to the stack
// of the context, and copied into a slice of their number at the end of the array.
func (d *interfaceDecoder) decodeAnySlice(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}
	cursor++ // '['
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == ']' {
		**(**any)(unsafe.Pointer(&p)) = []any{}
		return cursor + 1, nil
	}
	base := len(ctx.anyStack)
	for {
		c, err := d.decodeEmptyInterface(ctx, cursor, depth, unsafe.Pointer(&ctx.slot))
		if err != nil {
			ctx.slot = nil
			ctx.popAny(base)
			return 0, err
		}
		ctx.anyStack = append(ctx.anyStack, ctx.slot)
		ctx.slot = nil
		cursor = skipWhiteSpace(buf, c)
		switch buf[cursor] {
		case ']':
			elems := make([]any, len(ctx.anyStack)-base)
			copy(elems, ctx.anyStack[base:])
			ctx.popAny(base)
			**(**any)(unsafe.Pointer(&p)) = elems
			return cursor + 1, nil
		case ',':
			cursor++
		default:
			ctx.popAny(base)
			return 0, errors.ErrInvalidCharacter(buf[cursor], "slice", cursor)
		}
	}
}

func NewPathDecoder() Decoder {
	ifaceDecoder := &interfaceDecoder{
		typ:        emptyInterfaceType,
		structName: "",
		fieldName:  "",
		floatDecoder: newFloatDecoder("", "", func(p unsafe.Pointer, v float64) {
			*(*any)(p) = v
		}),
		numberDecoder: newNumberDecoder("", "", func(p unsafe.Pointer, v json.Number) {
			*(*any)(p) = v
		}),
		stringDecoder: newStringDecoder("", ""),
	}
	ifaceDecoder.sliceDecoder = newSliceDecoder(
		ifaceDecoder,
		emptyInterfaceType,
		emptyInterfaceType.Size(),
		"", "",
	)
	ifaceDecoder.mapDecoder = newMapDecoder(
		interfaceMapType,
		stringType,
		ifaceDecoder.stringDecoder,
		interfaceMapType.Elem(),
		ifaceDecoder,
		"", "",
	)
	return ifaceDecoder
}

var (
	truebytes  = []byte("true")
	falsebytes = []byte("false")
)

func (d *interfaceDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case '{':
		return d.mapDecoder.DecodePath(ctx, cursor, depth)
	case '[':
		return d.sliceDecoder.DecodePath(ctx, cursor, depth)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.floatDecoder.DecodePath(ctx, cursor, depth)
	case '"':
		return d.stringDecoder.DecodePath(ctx, cursor, depth)
	case 't':
		if err := validateTrue(buf, cursor); err != nil {
			return nil, 0, err
		}
		cursor += 4
		return [][]byte{truebytes}, cursor, nil
	case 'f':
		if err := validateFalse(buf, cursor); err != nil {
			return nil, 0, err
		}
		cursor += 5
		return [][]byte{falsebytes}, cursor, nil
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return nil, 0, err
		}
		cursor += 4
		return [][]byte{nullbytes}, cursor, nil
	}
	return nil, cursor, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
}

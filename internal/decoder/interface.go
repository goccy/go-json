package decoder

import (
	"bytes"
	"encoding"
	"encoding/json"
	"reflect"
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
	decoder, err := CompileToGetDecoder(runtime.TypePtr(typ))
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
		var v map[string]any
		ptr := unsafe.Pointer(&v)
		cursor, err := d.mapDecoder.Decode(ctx, cursor, depth, ptr)
		if err != nil {
			return 0, err
		}
		**(**any)(unsafe.Pointer(&p)) = v
		return cursor, nil
	case '[':
		var v []any
		ptr := unsafe.Pointer(&v)
		cursor, err := d.sliceDecoder.Decode(ctx, cursor, depth, ptr)
		if err != nil {
			return 0, err
		}
		**(**any)(unsafe.Pointer(&p)) = v
		return cursor, nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if (ctx.Option.Flags & UseNumberOption) != 0 {
			return d.numberDecoder.Decode(ctx, cursor, depth, p)
		}
		return d.floatDecoder.Decode(ctx, cursor, depth, p)
	case '"':
		var v string
		ptr := unsafe.Pointer(&v)
		cursor, err := d.stringDecoder.Decode(ctx, cursor, depth, ptr)
		if err != nil {
			return 0, err
		}
		**(**any)(unsafe.Pointer(&p)) = v
		return cursor, nil
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

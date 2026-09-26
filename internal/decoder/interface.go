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
	typ reflect.Type
	// hasMethods is set for an interface type which has methods, whose value may implement the unmarshalers: the
	// value of an interface{} is read without reflect ( see Decode ).
	hasMethods bool
	// values are how the values of the types which the interface value holds are decoded, by the type of an
	// interface{}, or by the itab of an interface type with methods ( see valueDecoderOf ).
	values        runtime.TypeCache[ifaceValueDecoder]
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
		hasMethods: typ.NumMethod() > 0,
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
	if d.hasMethods {
		return d.decodeWithMethods(ctx, cursor, depth, p)
	}
	// p is an interface{}: a nil one, the most common, is decoded as it is, and the value of another one may be
	// decoded into what it points to
	h := (*emptyInterface)(p)
	if h.ptr == nil {
		return d.decodeEmptyInterface(ctx, cursor, depth, p)
	}
	v := d.values.Load(uintptr(h.typ))
	if v == nil {
		var err error
		if v, err = d.valueDecoderOf(ctx, h.typ, *(*any)(p)); err != nil {
			return 0, err
		}
	}
	return d.decodeValue(ctx, cursor, depth, p, v, h.ptr)
}

// ifaceValueKind is how the value which an interface value holds is decoded.
type ifaceValueKind uint8

const (
	// ifaceValueReplaced is a value which is replaced by the value decoded as the one of an interface{}.
	ifaceValueReplaced ifaceValueKind = iota
	// ifaceValuePointee is a pointer to a value which is not an interface value: the value is decoded into what it
	// points to.
	ifaceValuePointee
	// ifaceValueUnmarshalerContext, ifaceValueUnmarshaler and ifaceValueTextUnmarshaler are values of an interface
	// type with methods which implement the unmarshalers.
	ifaceValueUnmarshalerContext
	ifaceValueUnmarshaler
	ifaceValueTextUnmarshaler
	// ifaceValueOther is a value of an interface type with methods which implements no unmarshaler: it can be
	// decoded from null only.
	ifaceValueOther
)

// ifaceValueDecoder is how a value of a type which an interface value holds is decoded, which depends on the type
// only: it is found once for a type.
type ifaceValueDecoder struct {
	kind ifaceValueKind
	// typ is the type of the value.
	typ unsafe.Pointer
	// dec is the decoder of a pointer of ifaceValuePointee.
	dec Decoder
}

// valueDecoderOf finds how the value iface, of the type typ, which the interface value holds is decoded, and
// keeps it for the type by key: the type of an interface{}, or the itab of an interface type with methods.
func (d *interfaceDecoder) valueDecoderOf(ctx *RuntimeContext, key unsafe.Pointer, iface any) (*ifaceValueDecoder, error) {
	h := (*emptyInterface)(unsafe.Pointer(&iface))
	v := &ifaceValueDecoder{typ: h.typ}
	typ := reflect.TypeOf(iface)
	switch {
	case d.hasMethods:
		switch iface.(type) {
		case unmarshalerContext:
			v.kind = ifaceValueUnmarshalerContext
		case json.Unmarshaler:
			v.kind = ifaceValueUnmarshaler
		case encoding.TextUnmarshaler:
			v.kind = ifaceValueTextUnmarshaler
		default:
			v.kind = ifaceValueOther
		}
	case typ.Kind() == reflect.Ptr && typ.Elem() != d.typ:
		dec, err := ctx.DecoderOf(h.typ)
		if err != nil {
			return nil, err
		}
		v.kind, v.dec = ifaceValuePointee, dec
	default:
		v.kind = ifaceValueReplaced
	}
	return d.values.Store(uintptr(key), v), nil
}

// decodeValue decodes the value at cursor into the interface value at p, which holds a value of the type of v,
// whose data word is ptr.
func (d *interfaceDecoder) decodeValue(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer, v *ifaceValueDecoder, ptr unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	switch v.kind {
	case ifaceValueReplaced:
		return d.decodeEmptyInterface(ctx, cursor, depth, p)
	case ifaceValuePointee:
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] == 'n' {
			if err := validateNull(buf, cursor); err != nil {
				return 0, err
			}
			cursor += 4
			**(**any)(unsafe.Pointer(&p)) = nil
			return cursor, nil
		}
		return v.dec.Decode(ctx, cursor, depth, ptr)
	}
	// a value of an interface type with methods: the value as an interface{}, whose type and data word are the ones
	// the interface value holds
	value := *(*any)(unsafe.Pointer(&emptyInterface{typ: v.typ, ptr: ptr}))
	switch v.kind {
	case ifaceValueUnmarshalerContext:
		return decodeUnmarshalerContext(ctx, buf, cursor, depth, value.(unmarshalerContext))
	case ifaceValueUnmarshaler:
		return decodeUnmarshaler(buf, cursor, depth, value.(json.Unmarshaler))
	case ifaceValueTextUnmarshaler:
		return decodeTextUnmarshaler(buf, cursor, depth, value.(encoding.TextUnmarshaler), p)
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
	// the error reports the interface type
	return 0, d.errUnmarshalType(d.typ, cursor)
}

// nonEmptyInterface is the layout of a value of an interface type with methods: the itab of the type it holds,
// which is nil for a nil value, and the data word.
type nonEmptyInterface struct {
	itab unsafe.Pointer
	ptr  unsafe.Pointer
}

// decodeWithMethods decodes the value of an interface type which has methods, whose value may implement the
// unmarshalers. How a value is decoded is kept by its itab, which is of one type for the interface type of the
// decoder.
func (d *interfaceDecoder) decodeWithMethods(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	h := (*nonEmptyInterface)(p)
	if h.itab == nil {
		// a nil value implements no unmarshaler: it can be decoded from null only
		buf := ctx.Buf
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] == 'n' {
			if err := validateNull(buf, cursor); err != nil {
				return 0, err
			}
			return cursor + 4, nil
		}
		return 0, d.errUnmarshalType(d.typ, cursor)
	}
	v := d.values.Load(uintptr(h.itab))
	if v == nil {
		var err error
		if v, err = d.valueDecoderOf(ctx, h.itab, reflect.NewAt(d.typ, p).Elem().Interface()); err != nil {
			return 0, err
		}
	}
	return d.decodeValue(ctx, cursor, depth, p, v, h.ptr)
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
		m, cursor, err := decodeNewStringAnyMap(ctx, d, cursor, depth)
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
		if f, next, ok := parseFloatFast(buf, cursor); ok && validEndNumberChar[buf[next]] {
			**(**any)(unsafe.Pointer(&p)) = ctx.boxFloat(f)
			return next, nil
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

// decodeNewStringAnyMap decodes the object at cursor into a new map[string]interface{}. The entries are pushed
// to the stacks of the context, and put into a map of their number at the end of the object: a map filled
// entry by entry grows and moves its entries several times on the way, which took more time than the
// entries themselves. A key which is repeated is put later, so that the last one wins as in a map filled in
// order.
func decodeNewStringAnyMap(ctx *RuntimeContext, d *interfaceDecoder, cursor, depth int64) (map[string]any, int64, error) {
	buf := ctx.Buf
	cursor++ // '{'
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == '}' {
		return map[string]any{}, cursor + 1, nil
	}
	base, keyBase := len(ctx.anyStack), len(ctx.keyStack)
	fail := func(err error) (map[string]any, int64, error) {
		ctx.slot = nil
		ctx.popAny(base)
		ctx.popKeys(keyBase)
		return nil, 0, err
	}
	for {
		key, c, ok, err := d.stringDecoder.decodeString(ctx, cursor)
		if err != nil {
			return fail(err)
		}
		if !ok {
			// null is not a key
			return fail(errors.ErrSyntax("invalid character 'n' looking for beginning of object key string", skipWhiteSpace(buf, cursor)+1))
		}
		cursor = skipWhiteSpace(buf, c)
		if buf[cursor] != ':' {
			return fail(errors.ErrExpected("colon after object key", cursor))
		}
		cursor++
		c, err = d.decodeEmptyInterface(ctx, cursor, depth, unsafe.Pointer(&ctx.slot))
		if err != nil {
			return fail(err)
		}
		// the key and the value are pushed together, after the value, whose own entries are popped already.
		ctx.keyStack = append(ctx.keyStack, key)
		ctx.anyStack = append(ctx.anyStack, ctx.slot)
		ctx.slot = nil
		cursor = skipWhiteSpace(buf, c)
		switch buf[cursor] {
		case '}':
			keys, values := ctx.keyStack[keyBase:], ctx.anyStack[base:]
			m := make(map[string]any, len(keys))
			for i, key := range keys {
				m[key] = values[i]
			}
			ctx.popAny(base)
			ctx.popKeys(keyBase)
			return m, cursor + 1, nil
		case ',':
			cursor++
		default:
			return fail(errors.ErrExpected("comma after object value", cursor))
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

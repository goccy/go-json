package decoder

import (
	"bytes"
	"encoding/base64"
	stderrors "errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonnum"
	"github.com/goccy/go-json/internal/runtime"
)

// The decoders of the semantics of encoding/json/v2, which the v2 json package decodes by: a value is decoded by the
// decoder of v1 of its type where v2 decodes it as v1 does, and by a decoder of its own where v2 differs. The
// decoding stops at the first error. An error of a value is an errors.UnmarshalError, which skipTypeError returns
// instead of recording it, and which the decoders here make of the type errors which the decoders of v1 record
// ( see v2Error ). Any other error, as a syntax error, the v2 json package finds again in the input, as
// encoding/json/v2 reports it: the decoders here only find the text which v2 rejects and v1 decodes, the strings of
// invalid UTF-8 and the objects which have a name twice ( see errV2Text ).

var (
	// errV2Text is the error of the text of a value which v2 rejects: a string of invalid UTF-8, or an object which
	// has a name twice.
	errV2Text = stderrors.New("json: invalid text of encoding/json/v2")
	// errNilInterface is the error of a nil value of an interface type with methods, which no value can be
	// decoded into.
	errNilInterface = stderrors.New("cannot derive concrete type for nil interface with finite type set")
	// errV2DuplicateName is the error of an object of an interface{} which has a name twice, which
	// decodeNewStringAnyMap returns with the semantics of v2: they reject it, or decode the value of the name into
	// the value of its first one ( see v2InterfaceDecoder ).
	errV2DuplicateName = stderrors.New("json: duplicate name of encoding/json/v2")
	// errNoDefaultRepresentation is the error of a time.Duration, which v2 has no representation of without the
	// `format` option.
	errNoDefaultRepresentation = stderrors.New("no default representation")
	errArrayUnderflow          = stderrors.New("too few array elements")
	errArrayOverflow           = stderrors.New("too many array elements")
)

var timeDurationType = reflect.TypeOf(time.Duration(0))

// cachedV2Decoders are the decoders of v2, without and with StringifyNumbers.
var cachedV2Decoders [2]runtime.TypeCache[Decoder]

// CompileToGetDecoderV2 returns the decoder of the type for the semantics of encoding/json/v2, with the numbers in
// JSON strings if stringifyNumbers is set.
func CompileToGetDecoderV2(typ unsafe.Pointer, stringifyNumbers bool) (Decoder, error) {
	cache := &cachedV2Decoders[0]
	if stringifyNumbers {
		cache = &cachedV2Decoders[1]
	}
	typeptr := uintptr(typ)
	if dec := cache.Load(typeptr); dec != nil {
		return *dec, nil
	}
	c := newCompiler(true)
	c.stringifyNumbers = stringifyNumbers
	dec, err := compileHead(runtime.TypeOfPtr(typ), c)
	if err != nil {
		return nil, err
	}
	return *cache.Store(typeptr, &dec), nil
}

// DecoderOfV2 returns the decoder of the semantics of encoding/json/v2 of the type, for the options and the
// functions of unmarshaling of ctx.
func (ctx *RuntimeContext) DecoderOfV2(typ unsafe.Pointer) (Decoder, error) {
	if funcs := ctx.v2Funcs(); funcs != nil {
		return funcs.decoderOf(typ, ctx.Option.V2&V2StringifyNumbers != 0)
	}
	return CompileToGetDecoderV2(typ, ctx.Option.V2&V2StringifyNumbers != 0)
}

// compileV2 returns the decoder of a value of typ by the semantics of v2.
func (c *compiler) compileV2(typ reflect.Type, structName, fieldName string) (Decoder, error) {
	if dec, exists := c.decoders[uintptr(runtime.TypePtr(typ))]; exists {
		return dec, nil
	}
	if funcs := c.funcs.funcsOf(typ); len(funcs) > 0 {
		d := &v2FuncDecoder{typ: typ, funcs: funcs}
		return d, c.compileNextV2(d, &d.next, typ, methodUnmarshalJSONFrom, structName, fieldName)
	}
	return c.compileDefaultV2(typ, methodUnmarshalJSONFrom, structName, fieldName)
}

// compileNextV2 compiles into next the decoder of typ which dec leaves a value to, by the methods of typ from first
// on, or by its kind ( see compileDefaultV2 ): dec stays the decoder of typ in c.decoders, so that the values of typ
// within the value are decoded by it too. Only a type which may have values of itself is put there: a decoder of a
// kind of v1 looks itself up there ( see compileKind ).
func (c *compiler) compileNextV2(dec Decoder, next *Decoder, typ reflect.Type, first int, structName, fieldName string) error {
	key := uintptr(runtime.TypePtr(typ))
	switch typ.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Pointer:
		if _, exists := c.decoders[key]; !exists {
			c.decoders[key] = dec
		}
	}
	d, err := c.compileDefaultV2(typ, first, structName, fieldName)
	*next = d
	return err
}

// compileDefaultV2 returns the decoder of a value of typ without the functions of unmarshaling: by the time types,
// by the first of its methods of unmarshaling from first on, or by its kind. A method which may decline a value
// leaves it to the decoder of the methods after it.
func (c *compiler) compileDefaultV2(typ reflect.Type, first int, structName, fieldName string) (Decoder, error) {
	switch typ {
	case timeDurationType:
		return &v2InvalidDecoder{typ: typ, err: errNoDefaultRepresentation}, nil
	case timeType:
		if c.stringTag {
			return &v2InvalidDecoder{typ: typ, err: errInvalidStringTag}, nil
		}
		return v2TimeDecoder{}, nil
	}
	switch v2MethodOf(typ, first) {
	case methodUnmarshalJSONFrom:
		d := &v2UnmarshalFromDecoder{typ: typ}
		return d, c.compileNextV2(d, &d.next, typ, methodUnmarshalJSON, structName, fieldName)
	case methodUnmarshalJSON:
		return &v2UnmarshalJSONDecoder{typ: typ, ptrType: runtime.TypePtr(reflect.PointerTo(typ))}, nil
	case methodUnmarshalText:
		return &v2UnmarshalTextDecoder{typ: typ, ptrType: runtime.TypePtr(reflect.PointerTo(typ)), zero: newZeroer(typ)}, nil
	}
	if c.stringTag && typ.Kind() != reflect.Ptr && !isNumberKind(typ.Kind()) {
		// the option applies to a number, through pointers
		return &v2InvalidDecoder{typ: typ, err: errInvalidStringTag}, nil
	}
	switch kind := typ.Kind(); kind {
	case reflect.Ptr:
		return compilePtr(typ, structName, fieldName, c)
	case reflect.Slice:
		if isV2Bytes(typ) {
			return newV2BytesDecoder(typ), nil
		}
		return compileSlice(typ, structName, fieldName, c)
	case reflect.Array:
		if isV2Bytes(typ) {
			return newV2BytesDecoder(typ), nil
		}
		return c.compileArrayV2(typ, structName, fieldName)
	case reflect.Interface:
		return c.newV2InterfaceDecoder(typ), nil
	case reflect.String:
		d := newStringDecoder(structName, fieldName)
		d.typ = typ
		return v2StringDecoder{d}, nil
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		if c.stringifyNumbers && kind != reflect.Bool {
			return newV2StringNumberDecoder(typ), nil
		}
		dec, err := compileKind(typ, structName, fieldName, c)
		if err != nil {
			return nil, err
		}
		return &v2ScalarDecoder{dec: dec, size: typ.Size()}, nil
	case reflect.Struct:
		return c.compileStructV2(typ)
	case reflect.Map:
		return c.compileMapV2(typ, structName, fieldName)
	}
	// a complex number, a channel, a function or an unsafe.Pointer, which has no representation
	return &v2InvalidDecoder{typ: typ}, nil
}

// isV2Bytes reports whether typ is a []byte or a [N]byte, whose value is a base64 string: a slice or an array of a
// named byte type is a JSON array of its values.
func isV2Bytes(typ reflect.Type) bool {
	elem := typ.Elem()
	return elem.Kind() == reflect.Uint8 && elem.PkgPath() == ""
}

// V2TypeError returns the type error which a decoder of v1 recorded, as the error of v2 ( see v2Error ), or nil if
// none did.
func (ctx *RuntimeContext) V2TypeError() error {
	if ctx.typeError == nil {
		return nil
	}
	return ctx.v2Error()
}

// v2Error returns the type error which a decoder of v1 recorded, as the error of v2, which stops the decoding: the
// error of a number which is not a value of its type reports the number. The type error is cleared.
func (ctx *RuntimeContext) v2Error() error {
	p := ctx.typeError
	ctx.typeError = nil
	if p.plain != nil {
		return p.plain
	}
	e := &errors.UnmarshalError{GoType: p.typ, Start: p.start, End: p.end, Err: p.err}
	if p.literal {
		e.HasValue = true
		e.Err = numberCause(ctx.Buf[p.start:p.end], p.typ)
	}
	return e
}

// numberCause returns the cause of the error of the number num, which is not a value of typ, as strconv returns it:
// strconv.ErrSyntax for a number which is not an integer, or negative, of an integer type, and strconv.ErrRange for a
// number out of the range of the type.
func numberCause(num []byte, typ reflect.Type) error {
	switch typ.Kind() {
	case reflect.Float32, reflect.Float64:
		return strconv.ErrRange
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if num[0] == '-' {
			return strconv.ErrSyntax
		}
	}
	if bytes.ContainsAny(num, ".eE") {
		return strconv.ErrSyntax
	}
	return strconv.ErrRange
}

// validV2Text reports whether the strings of text, a JSON text valid by the grammar, have the text which v2 allows:
// valid UTF-8, and no escape of a surrogate which is not a pair.
func validV2Text(text []byte) bool {
	return utf8.Valid(text) && !hasLoneSurrogate(text)
}

// hasLoneSurrogate reports whether the escapes of b, a part of a JSON text valid by the grammar, have a surrogate
// which is not a pair: a high surrogate which no low surrogate follows, or a low surrogate which no high surrogate
// precedes.
func hasLoneSurrogate(b []byte) bool {
	for {
		// a backslash is in a string only, where it starts an escape; an escape of \u has six bytes.
		i := bytes.IndexByte(b, '\\')
		if i < 0 {
			return false
		}
		if b[i+1] != 'u' {
			b = b[i+2:]
			continue
		}
		code := hexValue(b[i+2 : i+6])
		b = b[i+6:]
		switch {
		case code >= 0xdc00 && code < 0xe000:
			return true
		case code >= 0xd800 && code < 0xdc00:
			if len(b) < 6 || b[0] != '\\' || b[1] != 'u' {
				return true
			}
			if low := hexValue(b[2:6]); low < 0xdc00 || low >= 0xe000 {
				return true
			}
			b = b[6:]
		}
	}
}

// hexValue returns the number of the four hexadecimal digits of b.
func hexValue(b []byte) rune {
	var r rune
	for _, c := range b {
		switch {
		case c <= '9':
			c -= '0'
		case c <= 'F':
			c -= 'A' - 10
		default:
			c -= 'a' - 10
		}
		r = r<<4 | rune(c)
	}
	return r
}

// isNumberKind reports whether a value of the kind is a JSON number.
func isNumberKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// v2ScalarDecoder is the decoder of a bool or a number of v2: null is its zero value, and any other value is decoded
// by the decoder of v1, whose type error stops the decoding.
type v2ScalarDecoder struct {
	dec  Decoder
	size uintptr
}

func (d *v2ScalarDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == 'n' {
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		// a bool or a number has no pointer
		clear(unsafe.Slice((*byte)(p), d.size))
		return cursor + 4, nil
	}
	next, err := d.dec.Decode(ctx, cursor, depth, p)
	if err == nil && ctx.typeError != nil {
		return 0, ctx.v2Error()
	}
	return next, err
}

func (d *v2ScalarDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// v2StringDecoder is the decoder of a string of v2: null is "", and a string must be valid UTF-8, without an escape
// of a surrogate which is not a pair, unless V2AllowInvalidUTF8. The string is decoded as stringDecoder decodes it,
// which replaces such text by U+FFFD: a string which has it is checked again from the input.
type v2StringDecoder struct{ *stringDecoder }

func (d v2StringDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	var (
		s    string
		next int64
		ok   bool
		err  error
	)
	buf := ctx.Buf
	start := skipWhiteSpace(buf, cursor)
	if buf[start] == '"' {
		s, next, ok, err = d.decodeStringField(ctx, start+1)
	} else {
		s, next, ok, err = d.decodeStringValue(ctx, start)
	}
	if err != nil {
		return 0, err
	}
	if ok && ctx.Option.V2&V2AllowInvalidUTF8 == 0 && strings.IndexByte(s, 0xef) >= 0 && !validV2Text(ctx.Buf[start:next]) {
		return 0, errV2Text
	}
	**(**string)(unsafe.Pointer(&p)) = s
	return next, nil
}

// v2StringNumberDecoder is the decoder of a number of v2 with StringifyNumbers: the number is in a JSON string, by
// the grammar of the JSON numbers, and null is its zero value.
type v2StringNumberDecoder struct {
	typ  reflect.Type
	kind reflect.Kind
	str  *stringDecoder
}

func newV2StringNumberDecoder(typ reflect.Type) *v2StringNumberDecoder {
	return &v2StringNumberDecoder{typ: typ, kind: typ.Kind(), str: newStringDecoder("", "")}
}

func (d *v2StringNumberDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		clear(unsafe.Slice((*byte)(p), d.typ.Size()))
		return cursor + 4, nil
	case '"':
	default:
		return ctx.skipTypeError(cursor, depth, d.typ)
	}
	num, next, err := d.str.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	if !jsonnum.IsValid(num) {
		return 0, &errors.UnmarshalError{GoType: d.typ, Start: cursor, End: next, HasValue: true, Err: strconv.ErrSyntax}
	}
	if err := setNumber(p, d.kind, num); err != nil {
		return 0, &errors.UnmarshalError{GoType: d.typ, Start: cursor, End: next, HasValue: true, Err: err}
	}
	return next, nil
}

func (d *v2StringNumberDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// setNumber sets the number at p, of the numeric kind, to num, a JSON number, and returns the cause of the error of
// a number which is not a value of the kind, as strconv returns it. A float out of its range is set to ±Inf.
func setNumber(p unsafe.Pointer, kind reflect.Kind, num []byte) error {
	s := unsafe.String(unsafe.SliceData(num), len(num))
	var err error
	switch kind {
	case reflect.Float32, reflect.Float64:
		bits := 64
		if kind == reflect.Float32 {
			bits = 32
		}
		var f float64
		f, err = strconv.ParseFloat(s, bits)
		if kind == reflect.Float32 {
			*(*float32)(p) = float32(f)
		} else {
			*(*float64)(p) = f
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		if i, err = strconv.ParseInt(s, 10, intBits(kind)); err == nil {
			switch kind {
			case reflect.Int:
				*(*int)(p) = int(i)
			case reflect.Int8:
				*(*int8)(p) = int8(i)
			case reflect.Int16:
				*(*int16)(p) = int16(i)
			case reflect.Int32:
				*(*int32)(p) = int32(i)
			default:
				*(*int64)(p) = i
			}
		}
	default:
		var u uint64
		if u, err = strconv.ParseUint(s, 10, intBits(kind)); err == nil {
			switch kind {
			case reflect.Uint:
				*(*uint)(p) = uint(u)
			case reflect.Uint8:
				*(*uint8)(p) = uint8(u)
			case reflect.Uint16:
				*(*uint16)(p) = uint16(u)
			case reflect.Uint32:
				*(*uint32)(p) = uint32(u)
			case reflect.Uintptr:
				*(*uintptr)(p) = uintptr(u)
			default:
				*(*uint64)(p) = u
			}
		}
	}
	if err != nil {
		return stderrors.Unwrap(err)
	}
	return nil
}

// intBits returns the bits of an integer of the kind.
func intBits(kind reflect.Kind) int {
	switch kind {
	case reflect.Int8, reflect.Uint8:
		return 8
	case reflect.Int16, reflect.Uint16:
		return 16
	case reflect.Int32, reflect.Uint32:
		return 32
	case reflect.Int, reflect.Uint, reflect.Uintptr:
		return strconv.IntSize
	}
	return 64
}

// v2InvalidDecoder is the decoder of a type which no JSON value is decoded into by v2, null included: its error,
// whose cause is err, is the one of the type, before the value is read.
type v2InvalidDecoder struct {
	typ reflect.Type
	err error
}

func (d *v2InvalidDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	cursor = skipWhiteSpace(ctx.Buf, cursor)
	return 0, &errors.UnmarshalError{GoType: d.typ, Start: cursor, NoKind: true, Err: d.err}
}

func (d *v2InvalidDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// zeroer sets a value of a type to its zero value.
type zeroer struct {
	ptrType unsafe.Pointer
	size    uintptr
	// pointers is whether the type has pointers, whose zero value is set with the write barriers of the GC.
	pointers bool
}

func newZeroer(typ reflect.Type) zeroer {
	return zeroer{ptrType: ptrTypeOf(typ), size: typ.Size(), pointers: hasPointers(typ)}
}

func (z zeroer) zero(p unsafe.Pointer) {
	if z.pointers {
		valueAt(z.ptrType, p).SetZero()
		return
	}
	clear(unsafe.Slice((*byte)(p), z.size))
}

// hasPointers reports whether a value of typ has a pointer.
func hasPointers(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return false
	case reflect.Array:
		return typ.Len() > 0 && hasPointers(typ.Elem())
	case reflect.Struct:
		for i := range typ.NumField() {
			if hasPointers(typ.Field(i).Type) {
				return true
			}
		}
		return false
	}
	return true
}

// mergesInto reports whether a value of typ is decoded by v2 into the value it has, which a value of v2 must not be
// when it is an element of a slice: a pointer, which is decoded into what it points to, a value of an interface,
// a slice, a map, a struct, or an array of them.
func mergesInto(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Struct:
		return true
	case reflect.Array:
		return mergesInto(typ.Elem())
	}
	return false
}

// v2SliceDecoder is the decoder of a slice of v2: the elements are decoded as sliceDecoder decodes them, into the
// array of the slice, whose elements are set to their zero values before, as v2 decodes an element into its zero
// value. An element which the decoding overwrites anyway, as a number or a string is, is not set.
type v2SliceDecoder struct {
	*sliceDecoder
	zero bool
}

func newV2SliceDecoder(d *sliceDecoder) *v2SliceDecoder {
	return &v2SliceDecoder{sliceDecoder: d, zero: mergesInto(d.elemType)}
}

func (d *v2SliceDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	cursor = skipWhiteSpace(ctx.Buf, cursor)
	if d.zero && ctx.Buf[cursor] == '[' && (*sliceHeader)(p).cap > 0 {
		s := valueAt(d.slicePtrType, p)
		s.Slice(0, s.Cap()).Clear()
	}
	return d.sliceDecoder.Decode(ctx, cursor, depth, p)
}

// compileArrayV2 returns the decoder of the array type typ, which is not of bytes.
func (c *compiler) compileArrayV2(typ reflect.Type, structName, fieldName string) (Decoder, error) {
	d := &v2ArrayDecoder{typ: typ, array: newZeroer(typ), elem: newZeroer(typ.Elem()), alen: typ.Len()}
	dec, err := compileElem(typ, d, structName, fieldName, c)
	if err != nil {
		return nil, err
	}
	d.valueDecoder = dec
	return d, nil
}

// v2ArrayDecoder is the decoder of an array of v2: an element is decoded into its zero value, and the JSON array
// must have as many elements as the array. null is the zero value.
type v2ArrayDecoder struct {
	typ          reflect.Type
	valueDecoder Decoder
	// array and elem set the array and an element to their zero values.
	array zeroer
	elem  zeroer
	alen  int
}

func (d *v2ArrayDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		d.array.zero(p)
		return cursor + 4, nil
	case '[':
	default:
		return ctx.skipTypeError(cursor, depth, d.typ)
	}
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}
	cursor = skipWhiteSpace(buf, cursor+1)
	idx := 0
	var lengthErr error
	if buf[cursor] != ']' {
		for {
			if idx < d.alen {
				ep := unsafe.Add(p, uintptr(idx)*d.elem.size)
				d.elem.zero(ep)
				c, err := d.valueDecoder.Decode(ctx, cursor, depth, ep)
				if err != nil {
					return 0, err
				}
				cursor = c
			} else {
				c, err := skipValue(buf, cursor, depth)
				if err != nil {
					return 0, err
				}
				cursor = c
				lengthErr = errArrayOverflow
			}
			idx++
			cursor = skipWhiteSpace(buf, cursor)
			if buf[cursor] == ']' {
				break
			}
			if buf[cursor] != ',' {
				return 0, errors.ErrInvalidCharacter(buf[cursor], "array", cursor)
			}
			cursor = skipWhiteSpace(buf, cursor+1)
		}
	}
	for ; idx < d.alen; idx++ {
		d.elem.zero(unsafe.Add(p, uintptr(idx)*d.elem.size))
		lengthErr = errArrayUnderflow
	}
	if lengthErr != nil {
		// the error of the length is at the end of the array, which tells it
		return 0, &errors.UnmarshalError{GoType: d.typ, Start: cursor, Err: lengthErr}
	}
	return cursor + 1, nil
}

func (d *v2ArrayDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// v2BytesDecoder is the decoder of a []byte or a [N]byte of v2, whose bytes are a base64 string, by the standard
// encoding with padding: the array of a []byte is reused, and the bytes of a [N]byte must be N. null is the zero
// value.
type v2BytesDecoder struct {
	typ reflect.Type
	// alen is the length of a [N]byte, or -1 for a []byte.
	alen int
	str  *stringDecoder
}

func newV2BytesDecoder(typ reflect.Type) *v2BytesDecoder {
	d := &v2BytesDecoder{typ: typ, alen: -1, str: newStringDecoder("", "")}
	if typ.Kind() == reflect.Array {
		d.alen = typ.Len()
	}
	return d
}

func (d *v2BytesDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		if d.alen < 0 {
			*(*[]byte)(p) = nil
		} else {
			clear(unsafe.Slice((*byte)(p), d.alen))
		}
		return cursor + 4, nil
	case '"':
	default:
		return ctx.skipTypeError(cursor, depth, d.typ)
	}
	text, end, err := d.str.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	var dst []byte
	if d.alen < 0 {
		dst = (*(*[]byte)(p))[:0]
	} else {
		dst = unsafe.Slice((*byte)(p), d.alen)[:0]
	}
	b, err := base64.StdEncoding.AppendDecode(dst, text)
	if err == nil && len(text) != base64.StdEncoding.EncodedLen(len(b)) {
		// the decoding skips a line feed and a carriage return, which RFC 4648 doesn't allow
		i := bytes.IndexAny(text, "\r\n")
		err = fmt.Errorf("illegal character %s at offset %d", strconv.QuoteRune(rune(text[i])), i)
	}
	if err == nil && d.alen >= 0 {
		dst = dst[:d.alen]
		clear(dst[copy(dst, b):])
		if len(b) != d.alen {
			err = fmt.Errorf("decoded length of %d mismatches array length of %d", len(b), d.alen)
		}
	}
	if err != nil {
		return 0, &errors.UnmarshalError{GoType: d.typ, Start: start, Err: err}
	}
	if d.alen < 0 {
		if b == nil {
			b = []byte{}
		}
		*(*[]byte)(p) = b
	}
	return end, nil
}

func (d *v2BytesDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

var anySliceType = reflect.TypeOf([]any(nil))

// v2InterfaceDecoder is the decoder of an interface value of v2. null sets it to nil. A nil value of an interface type
// without methods is set to the value of the JSON value, decoded as v1 decodes it, and a nil value of an interface type
// with methods can't be decoded into. A value which an interface value holds is decoded into: a copy of it is
// decoded by its own decoder, which the interface value is set to, but a map[string]interface{}, which an object is
// merged into, and a []interface{} ( see decodeMerging ).
type v2InterfaceDecoder struct {
	typ     reflect.Type
	ptrType unsafe.Pointer
	empty   bool
	fast    *interfaceDecoder
	// fromAny is whether a function of unmarshaling takes an interface{} or a type of the values of a nil
	// interface{}: a value is then decoded by the decoders of the functions, as a value of its type ( see
	// decodeByKind ), and never by decodeMerging, which calls no function.
	fromAny bool
}

// anyValueTypes are the types of the values of a nil interface{}, by the first byte of the JSON value.
var anyValueTypes = func() [256]reflect.Type {
	var types [256]reflect.Type
	types['t'], types['f'] = boolType, boolType
	types['"'] = stringType
	types['-'] = float64Type
	for c := '0'; c <= '9'; c++ {
		types[c] = float64Type
	}
	types['{'] = interfaceMapType
	types['['] = anySliceType
	return types
}()

var boolType = reflect.TypeOf(false)

func (c *compiler) newV2InterfaceDecoder(typ reflect.Type) *v2InterfaceDecoder {
	d := &v2InterfaceDecoder{
		typ: typ, ptrType: ptrTypeOf(typ), empty: typ.NumMethod() == 0, fast: newEmptyInterfaceDecoder("", ""),
	}
	for _, t := range []reflect.Type{emptyInterfaceType, boolType, stringType, float64Type, interfaceMapType, anySliceType} {
		d.fromAny = d.fromAny || len(c.funcs.funcsOf(t)) > 0
	}
	return d
}

func (d *v2InterfaceDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == 'n' {
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		*(*[2]unsafe.Pointer)(p) = [2]unsafe.Pointer{}
		return cursor + 4, nil
	}
	// the first word of an interface value is its type, or its itab, which is nil for a nil value
	if (*[2]unsafe.Pointer)(p)[0] != nil {
		return d.decodeHeld(ctx, cursor, depth, p)
	}
	if !d.empty {
		return 0, &errors.UnmarshalError{GoType: d.typ, Start: cursor, NoKind: true, Err: errNilInterface}
	}
	if d.fromAny {
		return d.decodeByKind(ctx, cursor, depth, p)
	}
	next, err := d.fast.decodeEmptyInterface(ctx, cursor, depth, p)
	if (err != nil || ctx.typeError != nil) && ctx.Option.V2&V2AllowDuplicateNames != 0 {
		if c := buf[cursor]; c == '{' || c == '[' {
			// The value may have a name twice, whose second value is decoded into the first one, which v1 doesn't
			// do, and whose error may come before the one v1 found: the value is decoded again, by the merging of
			// v2, which meets its errors in the order of the input.
			ctx.typeError = nil
			*(*[2]unsafe.Pointer)(p) = [2]unsafe.Pointer{}
			return d.decodeMerging(ctx, cursor, depth, p)
		}
	}
	if ctx.typeError != nil {
		// a number out of the range of a float64, which v1 records and decodes on from: it comes before any other
		// error of the value
		return 0, ctx.v2Error()
	}
	if err != nil {
		return 0, err
	}
	if ctx.Option.V2&V2AllowInvalidUTF8 == 0 && !validV2Text(buf[cursor:next]) {
		return 0, errV2Text
	}
	return next, nil
}

// decodeByKind decodes the value at cursor into the nil interface value at p as a value of the type of its kind,
// by the decoder of the functions of unmarshaling, which the interface value is set to.
func (d *v2InterfaceDecoder) decodeByKind(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	typ := anyValueTypes[ctx.Buf[cursor]]
	if typ == nil {
		return 0, syntaxErrorAt(ctx.Buf, cursor, whereValue)
	}
	v := reflect.New(typ)
	dec, err := ctx.DecoderOfV2(runtime.TypePtr(v.Type()))
	if err != nil {
		return 0, err
	}
	next, err := dec.Decode(ctx, cursor, depth, v.UnsafePointer())
	valueAt(d.ptrType, p).Set(v.Elem())
	return next, err
}

// decodeHeld decodes the value at cursor into the value which the interface value at p holds.
func (d *v2InterfaceDecoder) decodeHeld(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	iface := valueAt(d.ptrType, p)
	held := iface.Elem()
	if !d.fromAny && (held.Type() == interfaceMapType || held.Type() == anySliceType) {
		return d.decodeMerging(ctx, cursor, depth, p)
	}
	v := reflect.New(held.Type())
	v.Elem().Set(held)
	dec, err := ctx.DecoderOfV2(runtime.TypePtr(v.Type()))
	if err != nil {
		return 0, err
	}
	next, err := dec.Decode(ctx, cursor, depth, v.UnsafePointer())
	iface.Set(v.Elem())
	return next, err
}

// decodeMerging decodes the object or the array at cursor into the interface{} at p, which holds nothing, a
// map[string]interface{}, which the object is merged into, or a []interface{}, whose array is reused: the value of a
// name the map has is decoded into the value of the name, as the value of a name which the object has twice, which
// V2AllowDuplicateNames lets it have, is decoded into the value of the first one.
func (d *v2InterfaceDecoder) decodeMerging(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	held := *(*any)(p)
	switch buf[cursor] {
	case '{':
		m, ok := held.(map[string]any)
		if held != nil && !ok {
			return ctx.skipTypeError(cursor, depth, reflect.TypeOf(held))
		}
		if m == nil {
			m = map[string]any{}
			*(*any)(p) = m
		}
		return d.decodeObjectInto(ctx, cursor, depth, m)
	case '[':
		s, ok := held.([]any)
		if held != nil && !ok {
			return ctx.skipTypeError(cursor, depth, reflect.TypeOf(held))
		}
		next, err := d.decodeArrayInto(ctx, cursor, depth, &s)
		*(*any)(p) = s
		return next, err
	}
	if held != nil {
		return ctx.skipTypeError(cursor, depth, reflect.TypeOf(held))
	}
	return d.Decode(ctx, cursor, depth, p)
}

// decodeElem decodes the value at cursor of an object or an array which decodeMerging decodes, into the interface{}
// at p, which holds the value of its name in the map, or nothing: an object or an array into nothing is decoded by
// decodeMerging too, so that the input is read once, whatever names it has twice.
func (d *v2InterfaceDecoder) decodeElem(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	cursor = skipWhiteSpace(ctx.Buf, cursor)
	if c := ctx.Buf[cursor]; (c == '{' || c == '[') && *(*any)(p) == nil {
		return d.decodeMerging(ctx, cursor, depth, p)
	}
	return d.Decode(ctx, cursor, depth, p)
}

// decodeObjectInto decodes the object at cursor into m, merging the value of a name into the value of the name
// which m has.
func (d *v2InterfaceDecoder) decodeObjectInto(ctx *RuntimeContext, cursor, depth int64, m map[string]any) (int64, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}
	cursor = skipWhiteSpace(buf, cursor+1)
	if buf[cursor] == '}' {
		return cursor + 1, nil
	}
	var names nameSet
	for {
		if buf[cursor] != '"' {
			return 0, syntaxErrorAt(buf, cursor, whereKey)
		}
		key, c, _, err := d.fast.stringDecoder.decodeString(ctx, cursor)
		if err != nil {
			return 0, err
		}
		if ctx.Option.V2&V2AllowInvalidUTF8 == 0 && strings.IndexByte(key, 0xef) >= 0 && !validV2Text(buf[cursor:c]) {
			return 0, errV2Text
		}
		if !names.add(key) && ctx.Option.V2&V2AllowDuplicateNames == 0 {
			return 0, errV2Text
		}
		cursor = skipWhiteSpace(buf, c)
		if buf[cursor] != ':' {
			return 0, syntaxErrorAt(buf, cursor, whereAfterKey)
		}
		value := m[key]
		c, err = d.decodeElem(ctx, cursor+1, depth, unsafe.Pointer(&value))
		m[key] = value
		if err != nil {
			return 0, err
		}
		cursor = skipWhiteSpace(buf, c)
		switch buf[cursor] {
		case '}':
			return cursor + 1, nil
		case ',':
			cursor = skipWhiteSpace(buf, cursor+1)
		default:
			return 0, syntaxErrorAt(buf, cursor, whereAfterPair)
		}
	}
}

// decodeArrayInto decodes the array at cursor into *s from its first element, which is set to nil before.
func (d *v2InterfaceDecoder) decodeArrayInto(ctx *RuntimeContext, cursor, depth int64, s *[]any) (int64, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}
	*s = (*s)[:0]
	if *s == nil {
		*s = []any{}
	}
	cursor = skipWhiteSpace(buf, cursor+1)
	if buf[cursor] == ']' {
		return cursor + 1, nil
	}
	for {
		*s = append(*s, nil)
		c, err := d.decodeElem(ctx, cursor, depth, unsafe.Pointer(&(*s)[len(*s)-1]))
		if err != nil {
			return 0, err
		}
		cursor = skipWhiteSpace(buf, c)
		switch buf[cursor] {
		case ']':
			return cursor + 1, nil
		case ',':
			cursor++
		default:
			return 0, syntaxErrorAt(buf, cursor, whereAfterElement)
		}
	}
}

func (d *v2InterfaceDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// maxListedNames is the number of the names of an object which a nameSet compares a name with: past it, they are
// kept in a map, so that an object of many names is not quadratic in its names.
const maxListedNames = 32

// nameSet is the names of an object, which is checked for a name it has twice.
type nameSet struct {
	list []string
	set  map[string]struct{}
}

// add adds the name, and reports whether the set didn't have it.
func (s *nameSet) add(name string) bool {
	if s.set != nil {
		if _, exists := s.set[name]; exists {
			return false
		}
		s.set[name] = struct{}{}
		return true
	}
	for _, n := range s.list {
		if n == name {
			return false
		}
	}
	s.list = append(s.list, name)
	if len(s.list) > maxListedNames {
		s.set = make(map[string]struct{}, len(s.list))
		for _, n := range s.list {
			s.set[n] = struct{}{}
		}
	}
	return true
}

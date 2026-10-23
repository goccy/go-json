package json

import (
	"bytes"
	"errors"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/decoder"
	"github.com/goccy/go-json/internal/encoder/run"
	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/options"
	iruntime "github.com/goccy/go-json/internal/runtime"
	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
)

// errNonNilReference is the error of an output which is not a non-nil pointer.
var errNonNilReference = errors.New("value must be passed as a non-nil pointer reference")

// Unmarshal decodes a []byte input into a Go value according to the provided unmarshal and decode options (while
// ignoring marshal or encode options). The input must be a single JSON value with optional whitespace
// interspersed. The output must be a non-nil pointer.
//
// Type-specific unmarshal functions and methods take precedence over the default representation of a value.
// Functions or methods that operate on *T are only called when decoding a value of type T (by taking its address)
// or a non-nil value of *T. Unmarshal ensures that a value is always addressable (by boxing it on the heap if
// necessary) so that these functions and methods can be consistently called. If a value must be shallow copied to
// call a pointer-receiver Unmarshaler, UnmarshalerFrom, or encoding.TextUnmarshaler method, then any mutations
// performed by the method are shallow copied back into the destination value.
//
// The input is decoded into the output according to the following rules:
//
//   - If any type-specific functions in a WithUnmarshalers option match the value type, then those functions are
//     called to decode the JSON value. If all applicable functions return errors.ErrUnsupported, then the input is
//     decoded according to subsequent rules.
//
//   - If the value type implements UnmarshalerFrom, then the UnmarshalJSONFrom method is called to decode the JSON
//     value. If the method returns errors.ErrUnsupported, then the input is decoded according to subsequent rules.
//
//   - If the value type implements Unmarshaler, then the UnmarshalJSON method is called to decode the JSON value.
//
//   - If the value type implements encoding.TextUnmarshaler, then the input is decoded as a JSON string and the
//     UnmarshalText method is called with the decoded string value. This fails with a SemanticError if the input
//     is not a JSON string.
//
//   - Otherwise, the JSON value is decoded according to the value's type as described in detail below.
//
// The input is decoded into the output as follows:
//
//   - A Go boolean is decoded from a JSON boolean (e.g., true or false).
//
//   - A Go string is decoded from a JSON string.
//
//   - A Go []byte or [N]byte is decoded from a JSON string containing a binary value using Base 64 Encoding per
//     RFC 4648, section 4. When decoding into a non-nil []byte, the slice length is reset to zero and the decoded
//     input is appended to it. When decoding into a [N]byte, the input must decode to exactly N bytes, otherwise
//     it fails with a SemanticError.
//
//   - A Go integer is decoded from a JSON number. It must be decoded from a JSON string containing a JSON number
//     if StringifyNumbers is specified or decoding a JSON object name. It fails with a SemanticError if the JSON
//     number has a fractional or exponent component. It also fails if it overflows the representation of the Go
//     integer type.
//
//   - A Go float is decoded from a JSON number. It must be decoded from a JSON string containing a JSON number if
//     StringifyNumbers is specified or decoding a JSON object name. It fails if it overflows the representation of
//     the Go float type.
//
//   - A Go map is decoded from a JSON object, where each JSON object name and value pair is recursively decoded as
//     the Go map key and value. Maps are not cleared. If the Go map is nil, then a new map is allocated to decode
//     into. If the decoded key matches an existing Go map entry, the entry value is reused by decoding the JSON
//     object value into it.
//
//   - A Go struct is decoded from a JSON object. See the “JSON Representation of Go structs” section in the
//     package-level documentation for more details.
//
//   - A Go slice is decoded from a JSON array, where each JSON element is recursively decoded and appended to the
//     Go slice. Before appending into a Go slice, a new slice is allocated if it is nil, otherwise the slice length
//     is reset to zero.
//
//   - A Go array is decoded from a JSON array, where each JSON array element is recursively decoded as each
//     corresponding Go array element. Each Go array element is zeroed before decoding into it. It fails with a
//     SemanticError if the JSON array does not contain the exact same number of elements as the Go array.
//
//   - A Go pointer is decoded based on the JSON kind and underlying Go type. If the input is a JSON null, then
//     this stores a nil pointer. Otherwise, it allocates a new underlying value if the pointer is nil, and
//     recursively JSON decodes into the underlying value.
//
//   - A Go interface is decoded based on the JSON kind and underlying Go type. If the input is a JSON null, then
//     this stores a nil interface value. Otherwise, a nil interface value of an empty interface type is
//     initialized with a zero Go bool, string, float64, map[string]any, or []any if the input is a JSON boolean,
//     string, number, object, or array, respectively. If the interface value is still nil, then this fails with a
//     SemanticError since decoding could not determine an appropriate Go type to decode into. Otherwise an
//     underlying value exists and it recursively decodes the JSON input into it.
//
//   - A Go time.Time is decoded from a JSON string containing the time formatted in RFC 3339 with nanosecond
//     precision.
//
//   - A Go time.Duration currently has no default representation and results in a SemanticError.
//
//   - A JSON null is decoded into any of the other values above as its zero value.
//
//   - All other Go types (e.g., complex numbers, channels, and functions) have no default representation and
//     result in a SemanticError.
func Unmarshal(in []byte, out any, opts ...Options) error {
	e := (*emptyInterface)(unsafe.Pointer(&out))
	ctx := decoder.TakeRuntimeContext()
	err := unmarshalInput(ctx, ctx.SetInput(in), e.typ, e.ptr, opts)
	ctx.Option.V2 = 0
	decoder.ReleaseRuntimeContext(ctx)
	// the decoder refers to the value by its address only.
	runtime.KeepAlive(out)
	return err
}

// UnmarshalOf decodes a []byte input into the value out points to, as Unmarshal does.
//
// Unmarshal takes its output as an interface value, and the decoder refers to the value by its address, so the
// value always escapes to the heap. UnmarshalOf takes the pointer by its type and never gives it to the decoder: it
// copies *out to a value in the heap which is reused, decodes into that value, and copies the result back to *out,
// also after an error. So out may point to a variable on the stack of the caller, and a value which needs no
// allocation of its own (no pointer, slice or map to fill) is decoded without an allocation: the strings are copied
// into a buffer shared by the calls.
//
// A value is copied twice, so a large value is faster with Unmarshal when it is in the heap anyway. As the result
// is a copy, an UnmarshalJSONFrom, UnmarshalJSON or UnmarshalText method of a type in the value must not keep the
// address of its receiver, which is the address of the copy.
func UnmarshalOf[T any](in []byte, out *T, opts ...Options) error {
	ptrType := reflect.TypeFor[*T]()
	if out == nil {
		return &SemanticError{action: "unmarshal", GoType: ptrType, Err: errNonNilReference}
	}
	ctx := decoder.TakeRuntimeContext()
	var p unsafe.Pointer
	if unsafe.Sizeof(*out) > run.MaxReusedValueSize {
		// every context of the pool would keep such a value: it is left to the GC.
		p = reflect.New(ptrType.Elem()).UnsafePointer()
	} else {
		p = ctx.TakeValue(ptrType.Elem(), iruntime.TypePtr(ptrType.Elem()))
	}
	*(*T)(p) = *out
	err := unmarshalInput(ctx, ctx.SetInput(in), iruntime.TypePtr(ptrType), p, opts)
	ctx.Option.V2 = 0
	*out = *(*T)(p)
	// the pool must not keep what the value refers to alive.
	var zero T
	*(*T)(p) = zero
	decoder.ReleaseRuntimeContext(ctx)
	return err
}

// emptyInterface is the layout of a value of interface{}.
type emptyInterface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

// configPool holds the configs of the options of the calls which have options, which would escape on the stack:
// they are applied by the calls of their interface.
var configPool = sync.Pool{New: func() any { return new(options.Config) }}

// noOptions is the config of a call without options.
var noOptions options.Config

// unmarshalInput decodes buf, the input of ctx followed by a nul byte, into the value at p, of the pointer type typ,
// by the options.
func unmarshalInput(ctx *decoder.RuntimeContext, buf []byte, typ, p unsafe.Pointer, opts []Options) error {
	cfg := &noOptions
	flags := decoder.V2Semantics
	if len(opts) > 0 {
		cfg = configPool.Get().(*options.Config)
		cfg.Apply(opts)
		flags = unmarshalFlags(cfg)
	}
	ctx.Option.Flags = 0
	ctx.Option.V2 = flags
	call := ctx.V2Call()
	call.Config = cfg
	if u, _ := cfg.Unmarshalers.(*Unmarshalers); u != nil {
		call.Funcs = u.funcs
	}
	err := decodeInput(ctx, buf, typ, p, cfg)
	call.Funcs, call.Config = nil, nil
	if st := (*fromState)(call.From); st != nil && st.att.In != nil {
		st.release()
	}
	if cfg != &noOptions {
		// the config may hold the functions of the caller
		*cfg = options.Config{}
		configPool.Put(cfg)
	}
	return err
}

// unmarshalFlags returns the options of the decoder for the options of a call.
func unmarshalFlags(c *options.Config) decoder.V2Flags {
	flags := decoder.V2Semantics
	if c.Has(options.AllowInvalidUTF8) {
		flags |= decoder.V2AllowInvalidUTF8
	}
	if c.Has(options.AllowDuplicateNames) {
		flags |= decoder.V2AllowDuplicateNames
	}
	if c.Has(options.StringifyNumbers) {
		flags |= decoder.V2StringifyNumbers
	}
	if c.Has(options.MatchCaseInsensitiveNames) {
		flags |= decoder.V2MatchCaseInsensitiveNames
	}
	if c.Has(options.RejectUnknownMembers) {
		flags |= decoder.V2RejectUnknownMembers
	}
	return flags
}

// decodeInput decodes buf, the input of ctx followed by a nul byte, into the value at p, of the pointer type typ,
// by the options of cfg, which the options of ctx are made of.
func decodeInput(ctx *decoder.RuntimeContext, buf []byte, typ, p unsafe.Pointer, cfg *options.Config) error {
	if typ == nil {
		return &SemanticError{action: "unmarshal", Err: errNonNilReference}
	}
	dec, err := ctx.DecoderOfV2(typ)
	if err != nil {
		if _, ok := err.(*ierrors.InvalidUnmarshalError); ok {
			// not a pointer
			return &SemanticError{action: "unmarshal", GoType: iruntime.TypeOfPtr(typ), Err: errNonNilReference}
		}
		return err
	}
	if p == nil {
		return &SemanticError{action: "unmarshal", GoType: iruntime.TypeOfPtr(typ), Err: errNonNilReference}
	}
	cursor, err := dec.Decode(ctx, 0, 0, p)
	if terr := ctx.V2TypeError(); terr != nil {
		// a type error which a decoder of v1 recorded, before any error it returned
		err = terr
	}
	if err == nil && !endsAt(buf, cursor) {
		// a value after the value, which is found by unmarshalError
		err = errAfterValue
	}
	if err != nil {
		ctx.DiscardTypeError()
		return unmarshalError(buf[:len(buf)-1], err, cfg)
	}
	return nil
}

// errAfterValue is the error of a value of the input which is followed by more than white space.
var errAfterValue = errors.New("value after the top-level value")

// endsAt reports whether only white space follows cursor in buf, the input followed by a nul byte.
func endsAt(buf []byte, cursor int64) bool {
	for ; ; cursor++ {
		switch buf[cursor] {
		case ' ', '\t', '\r', '\n':
			continue
		}
		return cursor == int64(len(buf))-1
	}
}

// unmarshalError returns the error err of the decoding of in, the input, as encoding/json/v2 reports it. A syntax
// error is found again in the input, by jsontext, which reports it as encoding/json/v2 does: the decoder finds a
// syntax error, or the text which v2 rejects, but not where encoding/json/v2 reports it. An error of a value is
// reported at its place, which jsontext finds, unless the input has a syntax error before, which encoding/json/v2
// would meet first.
func unmarshalError(in []byte, err error, cfg *options.Config) error {
	switch err.(type) {
	case *SemanticError, *jsontext.SyntacticError:
		// the error of a method or a function which read a jsontext.Decoder, at its place ( see unmarshalFrom )
		return err
	}
	e, ok := err.(*ierrors.UnmarshalError)
	if !ok {
		if _, serr := textcoder.UnmarshalPlace(in, cfg, -1, textcoder.ReadToken); serr != nil {
			return serr
		}
		// the decoder rejects no input which jsontext accepts
		return &SemanticError{action: "unmarshal", Err: err}
	}
	read := textcoder.ReadToken
	switch {
	case e.NoKind:
		read = textcoder.ReadNothing
	case e.Method != ierrors.UnmarshalByKind || e.GoType != nil && readsWhole(e.GoType):
		read = textcoder.ReadValue
	}
	ptr, serr := textcoder.UnmarshalPlace(in, cfg, e.Start, read)
	if serr != nil {
		return serr
	}
	if e.Err == ierrors.ErrDuplicateName {
		// a name of an object of a struct or of a map, which is the one of an earlier member by its field or its
		// key: a syntax error of the name.
		return &jsontext.SyntacticError{ByteOffset: e.Start, JSONPointer: jsontext.Pointer(ptr), Err: e.Err}
	}
	se := &SemanticError{
		action: "unmarshal", ByteOffset: e.Start, JSONPointer: jsontext.Pointer(ptr), GoType: e.GoType,
		Err: e.Err,
	}
	if !e.NoKind {
		se.JSONKind = jsontext.Value(in[e.Start:]).Kind()
	}
	if e.HasValue {
		se.JSONValue = jsontext.Value(bytes.Clone(in[e.Start:e.End]))
	}
	return methodError(se, e)
}

// methodError returns se, the error of e, an error of the value, as encoding/json/v2 reports it when a method or a
// function of unmarshaling returned it: a SemanticError of UnmarshalJSON or of a function of UnmarshalFunc is
// within the value, and a SemanticError or a SyntacticError of UnmarshalText is reported as it is.
func methodError(se *SemanticError, e *ierrors.UnmarshalError) error {
	switch e.Method {
	case ierrors.UnmarshalJSONMethod:
		se.Err = wrapErrUnsupported(se.Err, "UnmarshalJSON method")
		if inner, ok := se.Err.(*SemanticError); ok {
			collapsed := *inner
			collapsed.ByteOffset += se.ByteOffset
			collapsed.JSONPointer = se.JSONPointer + collapsed.JSONPointer
			return &collapsed
		}
	case ierrors.UnmarshalTextMethod:
		se.Err = wrapErrUnsupported(se.Err, "UnmarshalText method")
		switch se.Err.(type) {
		case *SemanticError, *jsontext.SyntacticError:
			return se.Err
		}
	}
	return se
}

// readsWhole reports whether encoding/json/v2 reads a value whole before it decodes it into a value of typ, whose
// syntax error comes first: a string, a number, and the bytes of a base64 string. A value of another type is read by
// its first token, which is the whole value but for an object or an array.
func readsWhole(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice, reflect.Array:
		return typ.Elem().Kind() == reflect.Uint8 && typ.Elem().PkgPath() == ""
	}
	return false
}

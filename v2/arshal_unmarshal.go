package json

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"runtime"
	"slices"
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
	buf := ctx.SetInput(in)
	cfg := &noOptions
	if len(opts) > 0 {
		cfg = beginCallOf(ctx, opts)
		// a call of its own, which ends here, so that a call without options, which most calls are, doesn't test the
		// option
		if cfg.Has(options.ReportErrorsWithLegacySemantics) {
			err := unmarshalWithV1Errors(ctx, buf, nil, e.typ, e.ptr, cfg)
			endCall(ctx)
			ctx.Option.V2 = 0
			decoder.ReleaseRuntimeContext(ctx)
			runtime.KeepAlive(out)
			return err
		}
	} else {
		setPlainCall(ctx)
	}
	// the value is decoded as decodeInput decodes it, here, as v1's Unmarshal decodes its value itself
	var dec decoder.Decoder
	var err error
	if e.typ != nil && e.ptr != nil {
		dec, err = ctx.DecoderOfV2(e.typ)
	}
	if dec == nil {
		err = outputError(ctx, e.typ, err)
	} else {
		var cursor int64
		if cursor, err = dec.Decode(ctx, 0, 0, e.ptr); err != nil || ctx.HasTypeError() || !endsAt(buf, cursor) {
			err = inputError(ctx, buf, cursor, err, cfg)
		}
	}
	endCall(ctx)
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
// allocation of its own (no pointer, slice, map or string to fill) is decoded without an allocation: a string is a
// copy of its own, as Unmarshal makes it.
//
// A value is copied twice, so a large value is faster with Unmarshal when it is in the heap anyway. As the result
// is a copy, an UnmarshalJSONFrom, UnmarshalJSON or UnmarshalText method of a type in the value must not keep the
// address of its receiver, which is the address of the copy.
func UnmarshalOf[T any](in []byte, out *T, opts ...Options) error {
	ptrType := reflect.TypeFor[*T]()
	if out == nil {
		return nilOutputError(ptrType, opts)
	}
	ctx := decoder.TakeRuntimeContext()
	p := reusedValue(ctx, out)
	buf := ctx.SetInput(in)
	cfg := &noOptions
	if len(opts) > 0 {
		cfg = beginCallOf(ctx, opts)
	} else {
		setPlainCall(ctx)
	}
	var err error
	if cfg.Has(options.ReportErrorsWithLegacySemantics) {
		err = unmarshalWithV1Errors(ctx, buf, nil, iruntime.TypePtr(ptrType), p, cfg)
	} else {
		err = decodeInput(ctx, buf, iruntime.TypePtr(ptrType), p, cfg)
	}
	endCall(ctx)
	ctx.Option.V2 = 0
	returnValue(p, out)
	decoder.ReleaseRuntimeContext(ctx)
	return err
}

// reusedValue returns a value of ctx which is reused, set to *out, which UnmarshalOf decodes into.
func reusedValue[T any](ctx *decoder.RuntimeContext, out *T) unsafe.Pointer {
	var p unsafe.Pointer
	if unsafe.Sizeof(*out) > run.MaxReusedValueSize {
		// every context of the pool would keep such a value: it is left to the GC.
		p = reflect.New(reflect.TypeFor[T]()).UnsafePointer()
	} else {
		p = ctx.TakeValue(reflect.TypeFor[T](), iruntime.TypePtr(reflect.TypeFor[T]()))
	}
	*(*T)(p) = *out
	return p
}

// returnValue sets *out to the value at p, which reusedValue returned, which is set to its zero value: the pool must
// not keep what the value refers to alive.
func returnValue[T any](p unsafe.Pointer, out *T) {
	*out = *(*T)(p)
	var zero T
	*(*T)(p) = zero
}

// emptyInterface is the layout of a value of interface{}.
type emptyInterface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

// noOptions is the config of a call without options.
var noOptions options.Config

// beginCallOf sets ctx for a call by the options, and returns their config, which endCall drops if it holds more than
// flags ( see endCall ): the config of the call of ctx, which would escape on the stack, as the options are applied by
// the calls of their interface. The options are applied as Config.Apply applies them, and the call is set as setCall
// sets it, here, without their calls. A call without options is set by setPlainCall, which the compiler inlines into
// the callers.
func beginCallOf(ctx *decoder.RuntimeContext, opts []Options) *options.Config {
	call := ctx.V2Call()
	cfg := &call.Options
	// the flags of the call before, which endCall leaves, as they refer to nothing
	cfg.Set, cfg.Value = 0, 0
	for _, o := range opts {
		switch o := o.(type) {
		case options.Bool:
			o.ApplyTo(cfg)
		case nil:
		default:
			o.ApplyTo(cfg)
		}
	}
	ctx.Option.Flags = decoderFlags(cfg)
	decoder.SetV2Flags(ctx, unmarshalFlags(cfg))
	call.Config, call.Funcs = cfg, nil
	if u, _ := cfg.Unmarshalers.(*Unmarshalers); u != nil {
		call.Funcs = u.funcs
	}
	return cfg
}

// setCall sets ctx for a call by the options of cfg, which clearCall drops from ctx.
func setCall(ctx *decoder.RuntimeContext, cfg *options.Config) {
	ctx.Option.Flags = decoderFlags(cfg)
	decoder.SetV2Flags(ctx, unmarshalFlags(cfg))
	call := ctx.V2Call()
	call.Config, call.Funcs = cfg, nil
	if u, _ := cfg.Unmarshalers.(*Unmarshalers); u != nil {
		call.Funcs = u.funcs
	}
}

// setPlainCall is setCall of a call without options.
func setPlainCall(ctx *decoder.RuntimeContext) {
	ctx.Option.Flags = 0
	decoder.SetV2Flags(ctx, decoder.V2Semantics)
	call := ctx.V2Call()
	call.Config, call.Funcs = &noOptions, nil
}

// endCall drops what beginCallOf or setPlainCall set in ctx for a call which refers to something of the caller or of
// the input: the decoder of the methods which read a jsontext.Decoder, and the options of the call if they hold more
// than flags, which may hold the functions of the caller, are dropped by endCallOf, so that the compiler inlines the
// rest. The config of the call, its options or noOptions, and the flags of its options refer to nothing: the next call
// sets them.
func endCall(ctx *decoder.RuntimeContext) {
	// the state of the methods or the options which hold more than flags, tested by one comparison, so that the cost
	// of the function is within the budget of the inliner on the platforms of 64 bits
	if call := ctx.CurrentV2Call(); uintptr(call.From)|uintptr(call.Options.Set&configHolds) != 0 {
		endCallOf(call)
	}
}

// configHolds are the options of a config which hold more than flags: strings, and the functions of the caller.
const configHolds = options.IndentSet | options.PrefixSet | options.MarshalersSet | options.UnmarshalersSet

// configHolds fits in a uintptr of 32 bits, which endCall tests it as: the conversion of a constant which overflows
// doesn't compile.
const _ = uint32(configHolds)

// endCallOf is endCall of a call which has a state of the methods which read a jsontext.Decoder, or options which
// hold more than flags.
func endCallOf(call *decoder.V2Call) {
	call.Funcs, call.Config = nil, nil
	if call.From != nil {
		releaseFrom(call)
	}
	if call.Options.Set&configHolds != 0 {
		call.Options = options.Config{}
	}
}

// clearCall drops what setCall set in ctx for a call.
func clearCall(ctx *decoder.RuntimeContext) {
	call := ctx.V2Call()
	call.Funcs, call.Config = nil, nil
	if call.From != nil {
		releaseFrom(call)
	}
}

// releaseFrom releases the decoder of the methods of the call, which has a state of them, if it was attached in the
// call.
func releaseFrom(call *decoder.V2Call) {
	if st := (*fromState)(call.From); st.att.In != nil {
		st.release()
	}
}

// decoderFlags returns the options of the decoders of v1 which the options of a call set: UseNumberOption, by which
// the decoder of an empty interface decodes a number as a json.Number, for UnmarshalAnyWithRawNumber.
func decoderFlags(c *options.Config) decoder.OptionFlags {
	if c.Value&options.UnmarshalAnyWithRawNumber != 0 {
		return decoder.UseNumberOption
	}
	return 0
}

// unmarshalFlags returns the options of the decoder for the options of a call.
func unmarshalFlags(c *options.Config) decoder.V2Flags {
	// the value is read once, so that the function stays small enough to be inlined into its callers
	v := c.Value
	// the options of encoding/json which change the unmarshaling, copied at once, as their flags are in the same order
	flags := decoder.V2Semantics |
		decoder.V2Flags(v&v1UnmarshalOptions/options.ParseBytesWithLooseRFC4648)*decoder.V2ParseBytesWithLooseRFC4648
	if v&options.AllowInvalidUTF8 != 0 {
		flags |= decoder.V2AllowInvalidUTF8
	}
	if v&options.AllowDuplicateNames != 0 {
		flags |= decoder.V2AllowDuplicateNames
	}
	if v&options.StringifyNumbers != 0 {
		flags |= decoder.V2StringifyNumbers
	}
	if v&options.MatchCaseInsensitiveNames != 0 {
		flags |= decoder.V2MatchCaseInsensitiveNames
	}
	if v&options.RejectUnknownMembers != 0 {
		flags |= decoder.V2RejectUnknownMembers
	}
	return flags
}

// v1UnmarshalOptions are the options of encoding/json which change the unmarshaling: all but
// OmitEmptyWithLegacySemantics, which are adjacent properties, from ParseBytesWithLooseRFC4648, in the order of their
// flags in the decoder.
const v1UnmarshalOptions = options.V1OptionFlags &^ options.OmitEmptyWithLegacySemantics

// the last of the options is as far from the first as its flag is, as both are made by iota in the same order: the
// length of the array type is 0 then, and it is another type, or a negative constant, otherwise.
var _ [0]struct{} = [options.StringifyWithLegacySemantics/options.ParseBytesWithLooseRFC4648 -
	options.Flags(decoder.V2StringifyWithLegacySemantics/decoder.V2ParseBytesWithLooseRFC4648)]struct{}{}

// decodeInput decodes buf, the input of ctx followed by a nul byte, into the value at p, of the pointer type typ,
// by the options of cfg, which the options of ctx are made of.
func decodeInput(ctx *decoder.RuntimeContext, buf []byte, typ, p unsafe.Pointer, cfg *options.Config) error {
	// the decoder of the output, as outputDecoder finds it
	var dec decoder.Decoder
	var err error
	if typ != nil && p != nil {
		dec, err = ctx.DecoderOfV2(typ)
	}
	if dec == nil {
		return outputError(ctx, typ, err)
	}
	cursor, err := dec.Decode(ctx, 0, 0, p)
	if err != nil || ctx.HasTypeError() || !endsAt(buf, cursor) {
		return inputError(ctx, buf, cursor, err, cfg)
	}
	return nil
}

// inputError returns the error of the decoding of buf, the input of ctx followed by a nul byte, which ended at cursor
// with err, a type error which a decoder of v1 recorded, or a value after the value, by the options of cfg.
func inputError(ctx *decoder.RuntimeContext, buf []byte, cursor int64, err error, cfg *options.Config) error {
	if terr := ctx.V2TypeError(); terr != nil {
		// a type error which a decoder of v1 recorded, before any error it returned
		err = terr
	}
	if err == nil && !endsAt(buf, cursor) {
		// a value after the value, which is found by unmarshalError
		err = errAfterValue
	}
	if err == nil {
		return nil
	}
	ctx.DiscardTypeError()
	return unmarshalError(&errorInput{buf: buf[:len(buf)-1]}, err, cfg)
}

// unmarshalWithV1Errors decodes buf, the input of ctx followed by a nul byte, or, if s is not nil, the whole input of
// the stream s, into the value at p, of the pointer type typ, by the options of cfg, which have
// ReportErrorsWithLegacySemantics, as encoding/json decodes it: an output which is not a non-nil pointer is an
// InvalidUnmarshalError, before the input is read; the input is checked first, so that a syntax error changes
// nothing, and is a SyntaxError, and an error of the reader is returned as it is; the decoding goes on after the error
// of a value ( see decoder.RuntimeContext.RecordV1Error ), and the first one is reported, as a type error of
// encoding/json.
func unmarshalWithV1Errors(ctx *decoder.RuntimeContext, buf []byte, s *decoder.Stream, typ, p unsafe.Pointer, cfg *options.Config) error {
	dec, err := outputDecoder(ctx, typ, p)
	if err != nil {
		return v1OutputError(err)
	}
	if s != nil {
		var err error
		if buf, err = s.ReadAllV2(); err != nil {
			return err
		}
		// decoded where it was read, without a copy into the context, which keeps none of it after the call
		decoder.SetBuf(ctx, buf)
		defer decoder.SetBuf(ctx, nil)
	}
	if err := v1InputError(buf[:len(buf)-1], cfg); err != nil {
		return err
	}
	return decodeWithV1Errors(ctx, dec, 0, typ, p, nil)
}

// decodeWithV1Errors decodes the input of ctx, a valid JSON value at the offset base of the whole input, by dec into
// the value at p, of the pointer type typ, as unmarshalWithV1Errors decodes it after its checks. in is the decoder
// which read the value, in the values which it is in, which the paths of the errors start with, or nil.
func decodeWithV1Errors(ctx *decoder.RuntimeContext, dec decoder.Decoder, base int64, typ, p unsafe.Pointer, in *jsontext.Decoder) error {
	if _, err := dec.Decode(ctx, 0, 0, p); err != nil && !ctx.RecordV1Error(err) {
		ctx.DiscardTypeError()
		return v1FatalError(err, base)
	}
	var outer []string
	if in != nil && ctx.HasTypeError() {
		outer = slices.Collect(in.StackPointer().Tokens())
	}
	// the offsets of the error are the ones of the whole input
	return ctx.TypeError(iruntime.TypeOfPtr(typ), 0, -base, outer)
}

// nilOutputError returns the error of a nil pointer out of the pointer type typ of UnmarshalOf or UnmarshalReadOf, by
// the options: an InvalidUnmarshalError with ReportErrorsWithLegacySemantics, as Unmarshal reports it.
func nilOutputError(typ reflect.Type, opts []Options) error {
	var cfg options.Config
	cfg.Apply(opts)
	if cfg.Has(options.ReportErrorsWithLegacySemantics) {
		return &ierrors.InvalidUnmarshalError{Type: typ}
	}
	return &SemanticError{action: "unmarshal", GoType: typ, Err: errNonNilReference}
}

// v1OutputError returns err, the error of an output which outputDecoder finds no decoder for, as encoding/json
// reports it: an InvalidUnmarshalError of an output which is not a non-nil pointer.
func v1OutputError(err error) error {
	if se, ok := err.(*SemanticError); ok && se.Err == errNonNilReference {
		return &ierrors.InvalidUnmarshalError{Type: se.GoType}
	}
	return err
}

// v1InputError returns the syntax error of in, the input of a call which is a single JSON value, by the options of
// cfg, as encoding/json reports it before it decodes the value, or nil: the value is read whole, as encoding/json/v2
// checks it with the option, and the input after it, as the call reads it.
func v1InputError(in []byte, cfg *options.Config) error {
	if jsontext.Value(in).IsValid(cfg) {
		return nil
	}
	// the decoder reads a bytes.Buffer in place, whole, as a value whose error is at the end of a read is reported
	_, err := jsontext.NewDecoder(bytes.NewBuffer(in), cfg).ReadValue()
	if err == nil || err == io.EOF {
		// the end of the input, or a value after the value
		_, err = textcoder.UnmarshalPlace(in, cfg, -1, textcoder.ReadToken)
	}
	return v1SyntaxError(err)
}

// v1FatalError returns err, an error which the decoding stops at with ReportErrorsWithLegacySemantics, of the input
// at the offset base of the whole input, as encoding/json reports it: a name which another member has, which
// encoding/json/v2 reports as a syntax error at the name, a syntax error or an error of the reader which a method or a
// function returned, of which the error of the reader is returned as it is, also within a token, which jsontext
// reports as the syntax error of the token.
func v1FatalError(err error, base int64) error {
	if e, ok := err.(*ierrors.UnmarshalError); ok {
		if e.Err == ierrors.ErrDuplicateName {
			return ierrors.ErrSyntax(e.Err.Error(), base+e.Start)
		}
		err = e.Err
	}
	if serr, ok := err.(*jsontext.SyntacticError); ok && textcoder.IsIOError(serr.Err) {
		err = serr.Err
	}
	if textcoder.IsIOError(err) {
		return errors.Unwrap(err)
	}
	return v1SyntaxError(err)
}

// outputDecoder returns the decoder of the output, the value at p of the pointer type typ, which must be a non-nil
// pointer.
func outputDecoder(ctx *decoder.RuntimeContext, typ, p unsafe.Pointer) (decoder.Decoder, error) {
	var dec decoder.Decoder
	var err error
	if typ != nil && p != nil {
		dec, err = ctx.DecoderOfV2(typ)
	}
	if dec == nil {
		return nil, outputError(ctx, typ, err)
	}
	return dec, nil
}

// outputError returns the error of an output which outputDecoder finds no decoder for: typ is nil, err is the error
// of the decoder of typ, or the value is nil, which is found after the decoder, as the decoder may be an error too.
func outputError(ctx *decoder.RuntimeContext, typ unsafe.Pointer, err error) error {
	if typ == nil {
		return &SemanticError{action: "unmarshal", Err: errNonNilReference}
	}
	if err == nil {
		// a nil value, whose type may have no decoder
		_, err = ctx.DecoderOfV2(typ)
	}
	if err != nil {
		if _, ok := err.(*ierrors.InvalidUnmarshalError); !ok {
			return err
		}
		// not a pointer
	}
	return &SemanticError{action: "unmarshal", GoType: iruntime.TypeOfPtr(typ), Err: errNonNilReference}
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

// errorInput is the input where the errors of a decoding are found again: the whole input of Unmarshal, or a part of
// the buffer of a stream, which is read from its place by att ( see decoder.StreamPlace ). The offsets of the errors
// of the decoder are in buf, whose first byte is at the offset base of the input.
type errorInput struct {
	buf  []byte
	base int64
	att  *textcoder.DecoderAttachment
}

// place reads the input as encoding/json/v2 reads it, up to the value at stop in buf, or the first syntax error with
// stop -1, as textcoder.UnmarshalPlace does.
func (in *errorInput) place(cfg *options.Config, stop int64, read textcoder.Read) (string, error) {
	if in.att == nil {
		return textcoder.UnmarshalPlace(in.buf, cfg, stop, read)
	}
	if stop >= 0 {
		stop += in.base
	}
	return textcoder.UnmarshalPlaceAt(in.att, stop, read)
}

// unmarshalError returns the error err of the decoding of in as encoding/json/v2 reports it. A syntax error is found
// again in the input, by jsontext, which reports it as encoding/json/v2 does: the decoder finds a syntax error, or the
// text which v2 rejects, but not where encoding/json/v2 reports it. An error of a value is reported at its place,
// which jsontext finds, unless the input has a syntax error before, which encoding/json/v2 would meet first.
func unmarshalError(in *errorInput, err error, cfg *options.Config) error {
	switch err.(type) {
	case *SemanticError, *jsontext.SyntacticError:
		// the error of a method or a function which read a jsontext.Decoder, at its place ( see unmarshalFrom )
		return err
	}
	e, ok := err.(*ierrors.UnmarshalError)
	if !ok {
		if _, serr := in.place(cfg, -1, textcoder.ReadToken); serr != nil {
			return serr
		}
		// the decoder rejects no input which jsontext accepts
		return &SemanticError{action: "unmarshal", Err: err}
	}
	read := textcoder.ReadToken
	switch {
	case e.NoKind:
		read = textcoder.ReadNothing
	case e.Method != ierrors.UnmarshalByKind || e.GoType != nil && readsWhole(e.GoType, cfg) && !decoder.IsArrayLengthError(e.Err):
		// but the length of a JSON array read into an array of bytes element by element ( see
		// FormatBytesWithLegacySemantics ), at the end of the JSON array, a token
		read = textcoder.ReadValue
	}
	ptr, serr := in.place(cfg, e.Start, read)
	if serr != nil {
		return serr
	}
	if e.Err == ierrors.ErrDuplicateName {
		// a name of an object of a struct or of a map, which is the one of an earlier member by its field or its
		// key: a syntax error of the name.
		return &jsontext.SyntacticError{ByteOffset: in.base + e.Start, JSONPointer: jsontext.Pointer(ptr), Err: e.Err}
	}
	se := &SemanticError{
		action: "unmarshal", ByteOffset: in.base + e.Start, JSONPointer: jsontext.Pointer(ptr), GoType: e.GoType,
		Err: e.Err,
	}
	if !e.NoKind {
		se.JSONKind = jsontext.Value(in.buf[e.Start:]).Kind()
	}
	if e.HasValue {
		se.JSONValue = jsontext.Value(bytes.Clone(in.buf[e.Start:e.End]))
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

// readsWhole reports whether encoding/json/v2 reads a value whole before it decodes it into a value of typ by the
// options of cfg, whose syntax error comes first: a string, a number, and the bytes of a base64 string. A value of
// another type is read by its first token, which is the whole value but for an object or an array.
func readsWhole(typ reflect.Type, cfg *options.Config) bool {
	switch typ.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice, reflect.Array:
		return decoder.IsV2Bytes(typ, unmarshalFlags(cfg))
	}
	return false
}

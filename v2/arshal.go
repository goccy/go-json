package json

import (
	"errors"
	"io"
	"reflect"
	"slices"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/encoder/run"
	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
)

// Marshal serializes a Go value as a []byte according to the provided marshal and encode options (while ignoring
// unmarshal or decode options). It does not terminate the output with a newline.
//
// Type-specific marshal functions and methods take precedence over the default representation of a value.
// Functions or methods that operate on *T are only called when encoding a value of type T (by taking its
// address) or a non-nil value of *T. Marshal ensures that a value is always addressable (by boxing it on the heap
// if necessary) so that these functions and methods can be consistently called. For performance, it is
// recommended that Marshal be passed a non-nil pointer to the value.
//
// The input value is encoded as JSON according the following rules:
//
//   - If any type-specific functions in a WithMarshalers option match the value type, then those functions are
//     called to encode the value. If all applicable functions return errors.ErrUnsupported, then the value is
//     encoded according to subsequent rules.
//
//   - If the value type implements MarshalerTo, then the MarshalJSONTo method is called to encode the value.
//
//   - If the value type implements Marshaler, then the MarshalJSON method is called to encode the value.
//
//   - If the value type implements encoding.TextAppender, then the AppendText method is called to encode the
//     value and subsequently encode its result as a JSON string.
//
//   - If the value type implements encoding.TextMarshaler, then the MarshalText method is called to encode the
//     value and subsequently encode its result as a JSON string.
//
//   - Otherwise, the value is encoded according to the value's type as described in detail below.
//
// Most Go types have a default JSON representation as follows:
//
//   - A Go boolean is encoded as a JSON boolean (e.g., true or false).
//
//   - A Go string is encoded as a JSON string.
//
//   - A Go []byte or [N]byte is encoded as a JSON string containing a binary value using Base 64 Encoding per
//     RFC 4648, section 4.
//
//   - A Go integer is encoded as a JSON number without fractions or exponents. If StringifyNumbers is specified
//     or encoding a JSON object name, then the JSON number is encoded within a JSON string.
//
//   - A Go float is encoded as a JSON number. If StringifyNumbers is specified or encoding a JSON object name,
//     then the JSON number is encoded within a JSON string. Encoding a NaN or ±Inf results in a SemanticError.
//
//   - A Go map is encoded as a JSON object, where each Go map key and value is recursively encoded as a name and
//     value pair in the JSON object. The Go map key must encode as a JSON string, otherwise this results in a
//     SemanticError. The Go map is traversed in a non-deterministic order. For deterministic encoding, consider
//     using the Deterministic option. By default, a nil map is encoded as an empty JSON object, unless the
//     FormatNilMapAsNull option is specified.
//
//   - A Go struct is encoded as a JSON object. See the “JSON Representation of Go structs” section in the
//     package-level documentation for more details.
//
//   - A Go slice is encoded as a JSON array, where each Go slice element is recursively JSON-encoded as the
//     elements of the JSON array. By default, a nil slice is encoded as an empty JSON array, unless the
//     FormatNilSliceAsNull option is specified.
//
//   - A Go array is encoded as a JSON array, where each Go array element is recursively JSON-encoded as the
//     elements of the JSON array. The JSON array length is always identical to the Go array length.
//
//   - A Go pointer is encoded as a JSON null if nil, otherwise it is the recursively JSON-encoded representation
//     of the underlying value.
//
//   - A Go interface is encoded as a JSON null if nil, otherwise it is the recursively JSON-encoded
//     representation of the underlying value.
//
//   - A Go time.Time is encoded as a JSON string containing the timestamp formatted in RFC 3339 with nanosecond
//     precision.
//
//   - A Go time.Duration currently has no default representation and results in a SemanticError.
//
//   - All other Go types (e.g., complex numbers, channels, and functions) have no default representation and
//     result in a SemanticError.
//
// JSON cannot represent cyclic data structures and Marshal does not handle them.
func Marshal(in any, opts ...Options) ([]byte, error) {
	ctx := encoder.TakeRuntimeContext()
	st := takeCallState(ctx, opts)
	st.setPlace(nil, textcoder.Level{}, true)
	buf, _, err := marshal(ctx, st, in, false, nil, 0)
	if err == nil && st.cfg.Value&formatFlags != 0 {
		buf, err = format(buf, &st.cfg)
	}
	releaseCallState(st)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctx)
		return nil, err
	}
	out := make([]byte, len(buf))
	copy(out, buf)
	encoder.ReleaseRuntimeContext(ctx)
	return out, nil
}

// MarshalWrite serializes a Go value into an io.Writer according to the provided marshal and encode options
// (while ignoring unmarshal or decode options). It does not terminate the output with a newline. See Marshal for
// details about the conversion of a Go value into JSON.
func MarshalWrite(out io.Writer, in any, opts ...Options) error {
	ctx := encoder.TakeRuntimeContext()
	st := takeCallState(ctx, opts)
	st.setPlace(nil, textcoder.Level{}, true)
	buf, _, err := marshal(ctx, st, in, false, nil, 0)
	if err == nil && st.cfg.Value&formatFlags != 0 {
		buf, err = format(buf, &st.cfg)
	}
	releaseCallState(st)
	if err == nil {
		_, err = out.Write(buf)
	}
	encoder.ReleaseRuntimeContext(ctx)
	return err
}

// MarshalOf returns the JSON encoding of in, as Marshal does.
//
// Marshal takes its argument as an interface value, which a value that is not a pointer, a map or a channel is
// copied to the heap for. MarshalOf takes the value by its type, and copies it to a value in the heap which is
// reused, so that it encodes the value without an allocation other than the one of the result. See MarshalWriteOf
// for the encoding without the result.
func MarshalOf[T any](in T, opts ...Options) ([]byte, error) {
	ctx := encoder.TakeRuntimeContext()
	st := takeCallState(ctx, opts)
	st.setPlace(nil, textcoder.Level{}, true)
	buf, _, err := marshal(ctx, st, nil, isNilPointer(&in), func(ctx *encoder.RuntimeContext) ([]byte, error) {
		return run.EncodeOf(ctx, &in)
	}, 0)
	if err == nil && st.cfg.Value&formatFlags != 0 {
		buf, err = format(buf, &st.cfg)
	}
	releaseCallState(st)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctx)
		return nil, err
	}
	out := make([]byte, len(buf))
	copy(out, buf)
	encoder.ReleaseRuntimeContext(ctx)
	return out, nil
}

// MarshalWriteOf writes the JSON encoding of in to out, as MarshalWrite does. It takes the value by its type, as
// MarshalOf does: the encoding is written from a buffer which is reused, without the copy of the result, so that a
// value is written without an allocation, unless the writer or a method of the value makes one.
func MarshalWriteOf[T any](out io.Writer, in T, opts ...Options) error {
	ctx := encoder.TakeRuntimeContext()
	st := takeCallState(ctx, opts)
	st.setPlace(nil, textcoder.Level{}, true)
	buf, _, err := marshal(ctx, st, nil, isNilPointer(&in), func(ctx *encoder.RuntimeContext) ([]byte, error) {
		return run.EncodeOf(ctx, &in)
	}, 0)
	if err == nil && st.cfg.Value&formatFlags != 0 {
		buf, err = format(buf, &st.cfg)
	}
	releaseCallState(st)
	if err == nil {
		_, err = out.Write(buf)
	}
	encoder.ReleaseRuntimeContext(ctx)
	return err
}

// isNilPointer reports whether the value at v is a nil pointer, without an interface value of it, which would
// move it to the heap.
func isNilPointer[T any](v *T) bool {
	return reflect.TypeFor[T]().Kind() == reflect.Pointer && *(*unsafe.Pointer)(unsafe.Pointer(v)) == nil
}

// MarshalEncode serializes a Go value into a jsontext.Encoder according to the provided marshal or encode options
// (while ignoring unmarshal or decode options). The options provided take precedence over options already
// applied on the jsontext.Encoder and only apply for the duration of the marshal call.
//
// See Marshal for details about the conversion of a Go value into JSON.
func MarshalEncode(out *jsontext.Encoder, in any, opts ...Options) error {
	ctx := encoder.TakeRuntimeContext()
	st := takeCallState(ctx, nil)
	out.Options().ApplyTo(&st.cfg)
	st.orig = st.cfg
	st.cfg.Apply(opts)
	inner, top, base := textcoder.Place(out)
	if len(opts) > 0 {
		if err := optionsChange(&st.orig, &st.cfg, inner); err != nil {
			releaseCallState(st)
			encoder.ReleaseRuntimeContext(ctx)
			levels, _ := textcoder.Position(out, nil)
			return &SemanticError{action: "marshal", ByteOffset: base, JSONPointer: pointerOf(levels, +1), GoType: reflect.TypeOf(in), Err: err}
		}
	}
	st.setPlace(out, inner, top)
	buf, opened, err := marshal(ctx, st, in, false, nil, base)
	if err != nil {
		if opened && !st.cfg.Has(options.AllowDuplicateNames) {
			// the value stopped in an object or an array of its own, which the encoder can't go on with.
			textcoder.Invalidate(out)
		}
		releaseCallState(st)
		encoder.ReleaseRuntimeContext(ctx)
		return err
	}
	// the encoder writes the value by the options of the call, which it was made by.
	restore := textcoder.Configure(out, &st.cfg)
	err = out.WriteValue(buf)
	restore()
	releaseCallState(st)
	encoder.ReleaseRuntimeContext(ctx)
	return err
}

var (
	errChangingDuplicateNames = errors.New("cannot change duplicate name checks after a JSON object has already begun processing")
	errChangingInvalidUTF8    = errors.New("cannot change UTF-8 checks after a JSON object has already begun processing")
	errChangingWhitespace     = errors.New("cannot change whitespace formatting within a MarshalEncode call")
)

// optionsChange returns the error of the options c of MarshalEncode which change the options orig of the encoder
// in a way it can't follow at its place, whose innermost level is inner: the checks of a name of an object which
// is being written, and the white space.
func optionsChange(orig, c *options.Config, inner textcoder.Level) error {
	if inner.Object && inner.Count%2 == 0 {
		switch {
		case orig.Has(options.AllowDuplicateNames) != c.Has(options.AllowDuplicateNames):
			return errChangingDuplicateNames
		case orig.Has(options.AllowInvalidUTF8) != c.Has(options.AllowInvalidUTF8):
			return errChangingInvalidUTF8
		}
	}
	const ws = options.Multiline | options.SpaceAfterColon | options.SpaceAfterComma
	if orig.Value&ws != c.Value&ws || orig.Indent != c.Indent || orig.Prefix != c.Prefix {
		return errChangingWhitespace
	}
	return nil
}

// marshal encodes in by the options of st into the buffer of ctx, at the place of the output which the state is
// set to ( see callState.setPlace ), whose offset is base. It returns the output without the comma
// which the engine writes after a value, or the error at its place in the whole output and whether the value
// stopped in an object or an array it opened.
//
// The value is in, or, if encode is set, the value which encode encodes, which is a nil pointer if nilPointer is
// set ( see MarshalOf ).
func marshal(ctx *encoder.RuntimeContext, st *callState, in any, nilPointer bool, encode func(*encoder.RuntimeContext) ([]byte, error), base int64) ([]byte, bool, error) {
	// the options of the engine: the default ones, which most calls have, without a call. ctx.RewriteFrom is not
	// reset: the levels of the output are read from its start after the state is reset ( see tracker.sync ).
	if st.cfg.Set == 0 {
		ctx.Option.Flag = defaultFlags
		ctx.CheckNames = false
	} else if st.configure(ctx, in, nilPointer) {
		// a nil pointer is null, which no function is called for.
		return []byte("null"), false, nil
	}
	st.base, st.opened, st.attached = base, false, false
	st.tracked.stale = true
	// the state is kept with ctx ( see takeCallState ), or is the one of the caller of a nested call: it is left
	// in the options, which the encodings of v1 don't read. The functions of the caller are not kept. A pointer is
	// written only if it changes, as a write of it costs a write barrier while the GC marks.
	if ctx.Option.V2 != any(st) {
		ctx.Option.V2 = st
	}
	var buf []byte
	var err error
	if encode != nil {
		buf, err = encode(ctx)
	} else {
		buf, err = run.Encode(ctx, in)
	}
	if invalidOut, invalid := ctx.InvalidUTF8Output(); invalid {
		// the first error, before which the encoding went on ( see encoder.RejectInvalidUTF8Option ).
		buf, err = invalidOut, &ierrors.TextError{Err: ierrors.ErrInvalidUTF8}
	}
	if ctx.Option.Funcs != nil {
		ctx.Option.Funcs = nil
	}
	if err != nil {
		return nil, len(levelsOf(buf, nil)) > 1, st.marshalError(buf, err)
	}
	buf = buf[:len(buf)-1]
	if ctx.CheckNames {
		// a name with invalid UTF-8 may be the same as another one after U+FFFD is written for its bytes.
		v := jsontext.Value(slices.Clone(buf))
		if err := v.Compact(jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(false)); err != nil {
			return nil, true, err
		}
	}
	return buf, false, nil
}

// configure sets the options of the engine for the options of the call, which are not the default ones, before the
// state is reset. It reports whether in is a nil pointer, or the value is one if nilPointer is set, which the
// functions of marshaling are not called for.
func (st *callState) configure(ctx *encoder.RuntimeContext, in any, nilPointer bool) bool {
	c := &st.cfg
	ctx.Option.Flag = optionFlags(c)
	// two names may be the same after U+FFFD is written for their invalid bytes.
	ctx.CheckNames = c.Has(options.AllowInvalidUTF8) && !c.Has(options.AllowDuplicateNames)
	if m, _ := c.Marshalers.(*Marshalers); m != nil && m.funcs != nil {
		if v := reflect.ValueOf(in); nilPointer || v.Kind() == reflect.Pointer && v.IsNil() {
			return true
		}
		ctx.Option.Flag |= encoder.MarshalFuncsOption
		ctx.Option.Funcs = m.funcs
	}
	return false
}

// format formats the output of the encoder as the options want it, which the encoder doesn't write by itself.
// It is called only for the options of formatFlags.
func format(buf []byte, c *options.Config) ([]byte, error) {
	// formatted by jsontext, which validates it as well.
	v := jsontext.Value(buf)
	if err := v.Format(c); err != nil {
		return nil, err
	}
	return v, nil
}

// formatFlags are the options of the format of the output, which the engine doesn't write by itself.
const formatFlags = options.Multiline | options.SpaceAfterColon | options.SpaceAfterComma |
	options.PreserveRawStrings | options.CanonicalizeRawInts | options.CanonicalizeRawFloats |
	options.ReorderRawObjects

// defaultFlags are the options of the engine for the default options.
var defaultFlags = optionFlags(&options.Config{})

// optionFlags returns the options of the engine for the options of c.
func optionFlags(c *options.Config) encoder.OptionFlag {
	flags := encoder.V2Option | encoder.TextEscapeOption
	if c.Has(options.EscapeForHTML) {
		flags |= encoder.HTMLEscapeOption
	}
	if c.Has(options.EscapeForJS) {
		flags |= encoder.NormalizeUTF8Option
	}
	if !c.Has(options.Deterministic) {
		flags |= encoder.UnorderedMapOption
	}
	if c.Has(options.FormatNilSliceAsNull) {
		flags |= encoder.FormatNilSliceAsNullOption
	}
	if c.Has(options.FormatNilMapAsNull) {
		flags |= encoder.FormatNilMapAsNullOption
	}
	if c.Has(options.StringifyNumbers) {
		flags |= encoder.StringifyNumbersOption
	}
	if c.Has(options.OmitZeroStructFields) {
		flags |= encoder.OmitZeroStructFieldsOption
	}
	if !c.Has(options.AllowInvalidUTF8) {
		flags |= encoder.RejectInvalidUTF8Option
	}
	if c.Has(options.AllowDuplicateNames) {
		flags |= encoder.AllowDuplicateNamesOption
	}
	if c.Has(options.MatchCaseInsensitiveNames) {
		flags |= encoder.MatchCaseInsensitiveNamesOption
	}
	return flags
}

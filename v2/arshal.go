package json

import (
	"errors"
	"io"
	"math"
	"reflect"
	"slices"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/encoder/run"
	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonstring"
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
	buf, _, err := marshal(ctx, st, in, nil, 0)
	if err == nil {
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
	buf, _, err := marshal(ctx, st, in, nil, 0)
	if err == nil {
		buf, err = format(buf, &st.cfg)
	}
	releaseCallState(st)
	if err == nil {
		_, err = out.Write(buf)
	}
	encoder.ReleaseRuntimeContext(ctx)
	return err
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
	levels, base := textcoder.Position(out, nil)
	if len(opts) > 0 {
		if err := optionsChange(&st.orig, &st.cfg, levels); err != nil {
			releaseCallState(st)
			encoder.ReleaseRuntimeContext(ctx)
			ptr := pointerOf(levels, +1)
			return &SemanticError{action: "marshal", ByteOffset: base, JSONPointer: ptr, GoType: reflect.TypeOf(in), Err: err}
		}
	}
	buf, opened, err := marshal(ctx, st, in, levels, base)
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
// in a way it can't follow at its place, which levels are of: the checks of a name of an object which is being
// written, and the white space.
func optionsChange(orig, c *options.Config, levels []textcoder.Level) error {
	if top := levels[len(levels)-1]; top.Object && top.Count%2 == 0 {
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

// marshal encodes in by the options of st into the buffer of ctx, at the place of the output which outer and base
// are of, or at the top level of an output of its own if outer is nil. It returns the output without the comma
// which the engine writes after a value, or the error at its place in the whole output and whether the value
// stopped in an object or an array it opened.
func marshal(ctx *encoder.RuntimeContext, st *callState, in any, outer []textcoder.Level, base int64) ([]byte, bool, error) {
	c := &st.cfg
	if c.Set == 0 {
		ctx.Option.Flag = defaultFlags
	} else {
		ctx.Option.Flag = optionFlags(c)
	}
	ctx.RewriteFrom = math.MaxInt
	// two names may be the same after U+FFFD is written for their invalid bytes.
	ctx.CheckNames = c.Has(options.AllowInvalidUTF8) && !c.Has(options.AllowDuplicateNames)
	st.reset(outer, base)
	// the state is kept with ctx ( see takeCallState ), or is the one of the caller of a nested call: it is left
	// in the options, which the encodings of v1 don't read. The functions of the caller are not kept.
	ctx.Option.V2 = st
	if m, _ := c.Marshalers.(*Marshalers); m != nil && m.funcs != nil {
		if v := reflect.ValueOf(in); v.Kind() == reflect.Pointer && v.IsNil() {
			// a nil pointer is null, which no function is called for.
			return []byte("null"), false, nil
		}
		ctx.Option.Flag |= encoder.MarshalFuncsOption
		ctx.Option.Funcs = m.funcs
	}
	buf, err := encode(ctx, in)
	ctx.Option.Funcs = nil
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

// format formats the output of the encoder as the options want it, which the encoder doesn't write by itself.
func format(buf []byte, c *options.Config) ([]byte, error) {
	if c.Value&formatFlags == 0 {
		return buf, nil
	}
	// formatted by jsontext, which validates it as well.
	v := jsontext.Value(buf)
	if err := v.Format(c); err != nil {
		return nil, err
	}
	return v, nil
}

// encode encodes in by ctx. A string of invalid UTF-8 is an error, which the escaper of the strings reports by a
// panic ( see jsonstring.InvalidUTF8 ): the strings of valid UTF-8 cost no check of what it reports.
func encode(ctx *encoder.RuntimeContext, in any) ([]byte, error) {
	r := encodeResult{ctx: ctx}
	r.encode(in)
	return r.buf, r.err
}

// encodeResult is the result of encode, which is set after a panic of invalid UTF-8 as well.
type encodeResult struct {
	ctx *encoder.RuntimeContext
	buf []byte
	err error
}

func (r *encodeResult) encode(in any) {
	defer r.recoverInvalidUTF8()
	r.buf, r.err = run.Encode(r.ctx, in)
}

func (r *encodeResult) recoverInvalidUTF8() {
	if v := recover(); v != nil {
		e, ok := v.(*jsonstring.InvalidUTF8)
		if !ok {
			panic(v)
		}
		r.buf, r.err = e.Out, &ierrors.TextError{Err: ierrors.ErrInvalidUTF8}
		r.ctx.Abandon()
	}
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

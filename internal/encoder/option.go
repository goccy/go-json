package encoder

import (
	"context"
	"io"
)

type OptionFlag uint32

const (
	// The options of the escaper of the strings are the lowest bits: they are its index ( see StringEscaper ).
	//
	// NormalizeUTF8Option replaces invalid UTF-8 with U+FFFD and escapes U+2028 and U+2029, as encoding/json
	// does; for the v2 semantics, it escapes U+2028 and U+2029 ( EscapeForJS of jsontext ).
	NormalizeUTF8Option OptionFlag = 1 << iota
	HTMLEscapeOption
	// TextEscapeOption escapes the strings as jsontext does, for the v2 semantics.
	TextEscapeOption
	// RejectInvalidUTF8Option, with TextEscapeOption, fails a string of invalid UTF-8, which is written with
	// U+FFFD otherwise: the escaper records the output before the string ( see RuntimeContext.InvalidUTF8Output ),
	// which the v2 json package reports.
	RejectInvalidUTF8Option
	IndentOption
	UnorderedMapOption
	DebugOption
	ColorizeOption
	ContextOption
	FieldQueryOption
	// OptimizeFieldOrderOption lets the encoder order the fields of a struct as it encodes them fastest:
	// the fields of the same kind together, and a recursive field last. The opcodes of a type are compiled
	// for it apart from the ones for the order of the struct.
	OptimizeFieldOrderOption
	// V2Option is the semantics of encoding/json/v2, which the v2 json package encodes by: the opcodes of a type
	// are compiled for it apart from the ones of the v1 semantics.
	V2Option
	// StringifyNumbersOption encodes a number as a JSON string, for the v2 semantics: the opcodes of the numbers
	// are compiled for it.
	StringifyNumbersOption
	// OmitZeroStructFieldsOption omits every field of a struct which is zero, for the v2 semantics: the opcodes of
	// the fields are compiled for it.
	OmitZeroStructFieldsOption
	// The options of encoding/json/v2 which the opcodes of the v2 semantics look at as they encode: the ones which
	// may differ from a call to another for the same opcodes.
	//
	// FormatNilSliceAsNullOption encodes a nil slice as null, not [].
	FormatNilSliceAsNullOption
	// FormatNilMapAsNullOption encodes a nil map as null, not {}.
	FormatNilMapAsNullOption
	// AllowDuplicateNamesOption lets a JSON object have the same name twice, which is an error otherwise.
	AllowDuplicateNamesOption
	// MatchCaseInsensitiveNamesOption matches the names of an embedded fallback with the names of the fields case
	// insensitively, unless a field has the option case:strict.
	MatchCaseInsensitiveNamesOption
	// MarshalFuncsOption is the functions of WithMarshalers of encoding/json/v2 ( Option.Funcs ), whose opcodes
	// are compiled for them apart from the ones of the tables shared by every goroutine.
	MarshalFuncsOption
)

// UncachedOption are the options whose opcodes are not the ones of the tables shared by every goroutine, nor
// of the recent opcodes of a context: a context which may filter the fields, and the functions of marshaling.
const UncachedOption = ContextOption | MarshalFuncsOption

// NilMapIsEmpty is whether a nil map is written as an empty JSON object, as encoding/json/v2 does unless
// FormatNilMapAsNull, or as null.
func NilMapIsEmpty(ctx *RuntimeContext) bool {
	return ctx.Option.Flag&(V2Option|FormatNilMapAsNullOption) == V2Option
}

// NilSliceIsEmpty is whether a nil slice is written as an empty JSON array, as encoding/json/v2 does unless
// FormatNilSliceAsNull, or as null.
func NilSliceIsEmpty(ctx *RuntimeContext) bool {
	return ctx.Option.Flag&(V2Option|FormatNilSliceAsNullOption) == V2Option
}

type Option struct {
	Flag OptionFlag
	// V2 is the state of the call of the v2 json package, which its hooks take ( see V2Hooks ).
	V2 any
	// Funcs are the functions of marshaling of the call, of MarshalFuncsOption.
	Funcs       *MarshalFuncs
	ColorScheme *ColorScheme
	Context     context.Context
	DebugOut    io.Writer
	DebugDOTOut io.WriteCloser
}

type EncodeFormat struct {
	Header string
	Footer string
}

type EncodeFormatScheme struct {
	Int       EncodeFormat
	Uint      EncodeFormat
	Float     EncodeFormat
	Bool      EncodeFormat
	String    EncodeFormat
	Binary    EncodeFormat
	ObjectKey EncodeFormat
	Null      EncodeFormat
}

type (
	ColorScheme = EncodeFormatScheme
	ColorFormat = EncodeFormat
)

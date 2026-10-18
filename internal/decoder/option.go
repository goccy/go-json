package decoder

import "context"

type OptionFlags uint8

const (
	FirstWinOption OptionFlags = 1 << iota
	ContextOption
	PathOption
	// UseNumberOption decodes a number into interface{} as json.Number instead of float64.
	UseNumberOption
	// DisallowUnknownFieldsOption makes an object key which matches no field of the struct an error.
	DisallowUnknownFieldsOption
	// NoCopyStringOption makes the decoded strings refer to the input instead of a copy of their bytes.
	NoCopyStringOption
)

// V2Flags are the options of a decoding by the semantics of encoding/json/v2, which the v2 json package decodes by
// ( see CompileToGetDecoderV2 ). They are zero for the decodings of v1.
type V2Flags uint8

const (
	// V2Semantics is set for every decoding of v2.
	V2Semantics V2Flags = 1 << iota
	// V2AllowInvalidUTF8 lets the strings have invalid UTF-8, which is replaced by U+FFFD.
	V2AllowInvalidUTF8
	// V2AllowDuplicateNames lets an object have a name more than once.
	V2AllowDuplicateNames
	// V2StringifyNumbers decodes the numbers of the numeric types from JSON strings, by decoders of their own.
	V2StringifyNumbers
)

type Option struct {
	Flags OptionFlags
	// V2 is in the room which Flags leaves before Context, so that Option keeps its size.
	V2      V2Flags
	Context context.Context
	Path    *Path
}

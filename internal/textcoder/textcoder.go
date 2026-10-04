// Package textcoder lets the v2 json package give a jsontext.Encoder to a method or a function of marshaling,
// which writes after the output of the encoder of the v2 semantics: jsontext sets these functions, which take its
// encoder as an any, since this package can't import jsontext.
package textcoder

import "github.com/goccy/go-json/internal/options"

// Level is a level of the output: the top level, or an open object or array.
type Level struct {
	Object bool
	// Count is the number of the tokens of the level written before the value of the next level, which is open,
	// or, at the innermost level, before the next token: the elements of an array, the names and the values of an
	// object.
	Count int64
	// Names are the names of an object, unquoted, in their order: the last one is of the value which is open or
	// next, if Count is odd.
	Names [][]byte
}

var (
	// Attach makes enc, a *jsontext.Encoder, write after out, which ends at the place of the output which levels
	// describe, the top level first, by the options. base is the offset of out in the whole output.
	Attach func(enc any, out []byte, base int64, levels []Level, opts options.Options)
	// Detach returns the output of enc, which Attach set up, and its depth and the count of the tokens of its
	// innermost level.
	Detach func(enc any) (out []byte, depth int, count int64)
	// Position returns the levels of enc, a *jsontext.Encoder, the top level first, and the offset in its output
	// where its next value starts: after the delimiter and the white space before it.
	Position func(enc any, levels []Level) ([]Level, int64)
	// Configure makes enc write by the options until restore is called, which gives it back its own.
	Configure func(enc any, opts options.Options) (restore func())
	// Invalidate makes enc fail every write until a Reset: MarshalEncode stopped within an object or an array it
	// opened.
	Invalidate func(enc any)
)

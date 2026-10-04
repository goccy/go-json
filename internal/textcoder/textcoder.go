// Package textcoder lets the v2 json package give a jsontext.Encoder to a method or a function of marshaling,
// which writes after the output of the encoder of the v2 semantics: jsontext sets these functions, which take its
// encoder as an any, since this package can't import jsontext.
package textcoder

import (
	"unsafe"

	"github.com/goccy/go-json/internal/options"
)

// Level is a level of the output: the top level, or an open object or array.
type Level struct {
	Object bool
	// Count is the number of the tokens of the level written before the value of the next level, which is open,
	// or, at the innermost level, before the next token: the elements of an array, the names and the values of an
	// object.
	Count int64
	// Name is the last name of an object, unquoted: of the value which is open or next, if Count is odd. The names
	// before it are not kept: a method writes a single value, which is checked for duplicate names in itself, and
	// the place of a value is the last name of each object.
	Name []byte
}

// Outer gives the levels of the place where an encoder was attached, the top level first ( see Attach ): they are
// found only when the encoder needs them, for a JSON pointer or the stack, which most writes don't.
type Outer interface {
	Levels() []Level
}

var (
	// Attach makes enc, a *jsontext.Encoder, write after out by the options, at a place of the output whose
	// innermost level is inner: the top level if top is set, or an open object or array, whose name, count and
	// the levels around it outer gives when enc needs them. enc is the pointer itself, without the check of a type
	// assertion, which the calls for every value of a method don't take. inner has the kind of the level, and a count of the
	// parity of the place, a value or a name, and which is 0 only if the level has no token yet. base is the
	// offset of out in the whole output, or, if out is empty, the offset of the value after its delimiter: the
	// delimiter which enc writes before the value, whose length skip is, is not a part of the output then, which
	// the caller writes to another encoder. same tells that opts and outer are the ones enc was attached with the
	// last time. It returns skip, and the depth and the count of the innermost level of enc, which Detach returns
	// after the writes, at the same depth.
	Attach func(enc unsafe.Pointer, out []byte, base int64, inner Level, top bool, outer Outer, opts *options.Config, same bool) (skip int, depth int, count int64)
	// Detach returns the output of enc, which Attach set up, and its depth and the count of the tokens of its
	// innermost level.
	Detach func(enc unsafe.Pointer) (out []byte, depth int, count int64)
	// Position returns the levels of enc, a *jsontext.Encoder, the top level first, and the offset in its output
	// where its next value starts: after the delimiter and the white space before it.
	Position func(enc any, levels []Level) ([]Level, int64)
	// Place returns the innermost level of enc, whose count may be the one which Attach was given ( see Attach ),
	// whether it is the top level, and the offset in its output where its next value starts.
	Place func(enc any) (inner Level, top bool, offset int64)
	// Configure makes enc write by the options until restore is called, which gives it back its own.
	Configure func(enc any, opts options.Options) (restore func())
	// Invalidate makes enc fail every write until a Reset: MarshalEncode stopped within an object or an array it
	// opened.
	Invalidate func(enc any)
)

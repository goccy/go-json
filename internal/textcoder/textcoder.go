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

// Attachment is the place of the output where an encoder is attached ( see Attach ), and, after Attach and Detach,
// where it is.
type Attachment struct {
	// Out is the output before the place, at the offset Base of the whole output, or, if Out is empty, Base is
	// the offset of the value after its delimiter: the delimiter which the encoder writes before the value, whose
	// length Attach sets in Skip, is not a part of the output then, which the caller writes to another encoder.
	Out  []byte
	Base int64
	// Object and Count are the innermost level of the place, an open object or array, or the top level if Top is
	// set: Count has the parity of the place, a value or a name, and is 0 only if the level has no token yet. The
	// levels around it, and its name and count, Outer gives when the encoder needs them.
	Object bool
	Top    bool
	Count  int64
	Outer  Outer
	// Opts are the options of the encoder, which are the ones it was attached with the last time if Same is set.
	Opts *options.Config
	Same bool
	// Skip, and Depth and Count of the innermost level, which Detach sets after the writes, Attach sets.
	Skip  int
	Depth int
	// Failed is the error of a call which failed at the encoder, which Detach sets ( see Fail ).
	Failed error
}

// Outer gives the levels of the place where an encoder was attached, the top level first ( see Attach ): they are
// found only when the encoder needs them, for a JSON pointer or the stack, which most writes don't.
type Outer interface {
	Levels() []Level
}

var (
	// Attach makes enc, a *jsontext.Encoder, write after the output of the place p by its options ( see
	// Attachment ).
	// enc is the pointer itself, without the check of a type assertion, which the calls for every value of a
	// method don't take; the place is given by a pointer, as its fields don't fit the registers of a call.
	Attach func(enc unsafe.Pointer, p *Attachment)
	// Detach sets the output of enc, which Attach set up, and its depth and the count of the tokens of its
	// innermost level in p.
	Detach func(enc unsafe.Pointer, p *Attachment)
	// Position returns the levels of enc, a *jsontext.Encoder, the top level first, and the offset in its output
	// where its next value starts: after the delimiter and the white space before it.
	Position func(enc any, levels []Level) ([]Level, int64)
	// Place returns the innermost level of enc, whose count may be the one which Attach was given ( see Attach ),
	// whether it is the top level, the offset in its output where its next value starts, and the Outer which it
	// was attached with, if it was.
	Place func(enc any) (inner Level, top bool, offset int64, outer Outer)
	// Configure makes enc write by the options until restore is called, which gives it back its own.
	Configure func(enc any, opts options.Options) (restore func())
	// Invalidate makes enc fail every write until a Reset: MarshalEncode stopped within an object or an array it
	// opened.
	Invalidate func(enc any)
	// Fail records err in enc, an encoder which Attach attached, the first error of a call of MarshalEncode which
	// failed at it: the value which enc was given the place of fails with it, whatever its method returns.
	Fail func(enc any, err error)
)

package errors

import (
	"errors"
	"reflect"
)

// SemanticError is an error of encoding a value of GoType by the semantics of encoding/json/v2, which the encoder
// returns with the output written before it: the v2 json package reports it as its SemanticError, at the place of
// that output.
type SemanticError struct {
	GoType reflect.Type
	Err    error
}

func (e *SemanticError) Error() string {
	s := "json: cannot marshal"
	if e.GoType != nil {
		s += " from Go " + e.GoType.String()
	}
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

func (e *SemanticError) Unwrap() error { return e.Err }

// ErrCycle reports a value which is a value of itself, which JSON can't represent.
var ErrCycle = errors.New("encountered a cycle")

// The errors of the JSON text, which jsontext reports by these very errors.
var (
	ErrDuplicateName = errors.New("duplicate object member name")
	ErrNonStringName = errors.New("object member name must be a string")
)

// TextError is an error of the JSON text which the encoder of the semantics of encoding/json/v2 would write, which
// it returns with the output written before it: the v2 json package reports it as jsontext.SyntacticError at the
// place of that output. Err is ErrInvalidUTF8, ErrDuplicateName or ErrNonStringName.
type TextError struct {
	Err error
	// Name is the duplicate name, of ErrDuplicateName, which the pointer of the error points to.
	Name string
	// GoType is the type of the value which made the error, if the v2 json package reports it in a SemanticError,
	// as it does for a name which is not a string.
	GoType reflect.Type
	// Out is the output where the error is, if it is not the output which the encoder returns with it: a name of
	// an object which was written before.
	Out []byte
	// AtDelim is whether the offset of the error is the one of the comma which ends Out, before the name: the
	// place where encoding/json/v2 reports a name of an embedded fallback.
	AtDelim bool
}

func (e *TextError) Error() string { return "jsontext: " + e.Err.Error() }

func (e *TextError) Unwrap() error { return e.Err }

// OutputError is an error of the output Out, which the encoder wrote in a context of its own ( a value written by a
// default representation, or the value of an embedded fallback ): the v2 json package reports Err at the place of
// Out instead of the output which the encoder returns with it.
type OutputError struct {
	Out []byte
	Err error
}

func (e *OutputError) Error() string { return e.Err.Error() }

func (e *OutputError) Unwrap() error { return e.Err }

// MethodError is an error which a method or a function of marshaling returned, for the semantics of
// encoding/json/v2, which the encoder returns with the output written before the call: the v2 json package
// reports it at the place of that output, for the type of the value. How it reports an error which is itself one
// of its SemanticErrors depends on the method ( see MethodKind ).
type MethodError struct {
	GoType reflect.Type
	Err    error
	Kind   MethodKind
	// Out is the output where the error is, of MethodJSONTo, which wrote to the encoder: the output before the
	// call if it is nil. Where is where its pointer points: the next value (+1), the last value (-1), or the
	// level the method stopped in (0).
	Out   []byte
	Where int8
}

// MethodKind is the kind of the method or function which returned a MethodError.
type MethodKind uint8

const (
	// MethodJSON is MarshalJSON or a function of MarshalFunc: the place of a SemanticError it returns is within
	// the value it returns, after the place of the call.
	MethodJSON MethodKind = iota
	// MethodText is MarshalText or AppendText: a SemanticError it returns is reported as it is.
	MethodText
	// MethodJSONTo is MarshalJSONTo or a function of MarshalToFunc, which writes to the encoder: a SemanticError
	// it returns is given the place of the output where it stopped, if it has none.
	MethodJSONTo
)

func (e *MethodError) Error() string {
	return (&SemanticError{GoType: e.GoType, Err: e.Err}).Error()
}

func (e *MethodError) Unwrap() error { return e.Err }

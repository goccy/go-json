// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stdtest_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/jsontext"
	. "github.com/goccy/go-json/v2"
)

// CaseName is a test case name with the position where the case is declared.
type CaseName struct {
	Name  string
	Where CasePos
}

// CasePos is the position of a test case declaration.
type CasePos struct {
	file string
	line int
}

func (p CasePos) String() string { return fmt.Sprintf("%s:%d", p.file, p.line) }

// Name returns a CaseName recording the position of its caller.
func Name(s string) CaseName {
	_, file, line, _ := runtime.Caller(1)
	return CaseName{Name: s, Where: CasePos{file: filepath.Base(file), line: line}}
}

func len64[Bytes ~[]byte | ~string](in Bytes) int64 {
	return int64(len(in))
}

// stackPosition returns the depth of e and the number of the tokens of its innermost object or array.
func stackPosition(e *jsontext.Encoder) (int, int64) {
	_, n := e.StackIndex(e.StackDepth())
	return e.StackDepth(), n
}

func wrapErrUnsupported(err error, what string) error {
	if errors.Is(err, errors.ErrUnsupported) {
		return errorText(what + " may not return errors.ErrUnsupported")
	}
	return err
}

// ptr returns a pointer to v: new(v) of an expression needs Go 1.26.
func ptr[T any](v T) *T {
	return &v
}

// startDetectingCyclesAfter is the depth of the values after which a pointer, a map or a slice which is a value
// of itself is reported.
const startDetectingCyclesAfter = 1000

// errorText is an expected error which is compared by its Error text.
// It stands for errors whose types are unexported.
type errorText string

func (e errorText) Error() string { return string(e) }

var (
	errInvalidUTF8     = errorText("invalid UTF-8")
	errCycle           = errorText("encountered a cycle")
	errNonNilReference = errorText("value must be passed as a non-nil pointer reference")
	errNilInterface    = errorText("cannot derive concrete type for nil interface with finite type set")

	errInvalidStringTag    = errorText("invalid use of `string` tag option")
	errNoExportedFields    = errorText("Go struct has no exported fields")
	errNonSingularValue    = errorText("must read or write exactly one value")
	errRawEmbedNotObject   = errorText("embedded raw value must be a JSON object")
	errUnsupportedMutation = errorText("unsupported calls must not read or write any tokens")
)

func newInvalidCharacterError(prefix, where string, offset int64, pointer jsontext.Pointer) error {
	r, _ := utf8.DecodeRuneInString(prefix)
	var what string
	switch {
	case r == '\'':
		what = `'\''`
	case r == '"':
		what = `'"'`
	case r < ' ' || r == utf8.RuneError || !utf8.ValidRune(r):
		what = strings.Trim(fmt.Sprintf("%q", string(r)), `"`)
		what = "'" + what + "'"
	default:
		what = "'" + string(r) + "'"
	}
	return &jsontext.SyntacticError{ByteOffset: offset, JSONPointer: pointer, Err: errorText("invalid character " + what + " " + where)}
}

func newNonStringNameError(offset int64, pointer jsontext.Pointer) error {
	return &jsontext.SyntacticError{ByteOffset: offset, JSONPointer: pointer, Err: jsontext.ErrNonStringName}
}

func newInvalidUTF8Error(offset int64, pointer jsontext.Pointer) error {
	return &jsontext.SyntacticError{ByteOffset: offset, JSONPointer: pointer, Err: errInvalidUTF8}
}

func newDuplicateNameError(ptr jsontext.Pointer, quotedName []byte, offset int64) error {
	if quotedName != nil {
		name, _ := jsontext.AppendUnquote(nil, quotedName)
		ptr = ptr.AppendToken(string(name))
	}
	return &jsontext.SyntacticError{ByteOffset: offset, JSONPointer: ptr, Err: jsontext.ErrDuplicateName}
}

func newParseTimeError(layout, value, layoutElem, valueElem, message string) error {
	return &time.ParseError{Layout: layout, Value: value, LayoutElem: layoutElem, ValueElem: valueElem, Message: message}
}

// semanticError is an expected *SemanticError, whose action can't be set outside of the json package: it is
// compared by the text of the error instead ( see equalError ).
type semanticError struct {
	action string
	SemanticError
}

func (e *semanticError) Error() string { return e.SemanticError.Error() }

func EM(err error) *semanticError {
	return &semanticError{action: "marshal", SemanticError: SemanticError{Err: err}}
}

func EU(err error) *semanticError {
	return &semanticError{action: "unmarshal", SemanticError: SemanticError{Err: err}}
}

func (e *semanticError) withVal(val string) *semanticError {
	e.JSONValue = jsontext.Value(val)
	return e
}

func (e *semanticError) withPos(prefix string, pointer jsontext.Pointer) *semanticError {
	e.ByteOffset = int64(len(prefix))
	e.JSONPointer = pointer
	return e
}

func (e *semanticError) withType(k jsontext.Kind, t reflect.Type) *semanticError {
	e.JSONKind = k
	e.GoType = t
	return e
}

// equalError reports whether got is the error which want expects.
func equalError(got, want error) bool {
	switch want := want.(type) {
	case nil:
		return got == nil
	case *semanticError:
		g, ok := got.(*SemanticError)
		if !ok || !strings.HasPrefix(g.Error(), "json: cannot "+want.action) {
			return false
		}
		return g.ByteOffset == want.ByteOffset && g.JSONPointer == want.JSONPointer && g.JSONKind == want.JSONKind &&
			string(g.JSONValue) == string(want.JSONValue) && g.GoType == want.GoType && equalError(g.Err, want.Err)
	case *jsontext.SyntacticError:
		g, ok := got.(*jsontext.SyntacticError)
		if !ok {
			return false
		}
		return g.ByteOffset == want.ByteOffset && g.JSONPointer == want.JSONPointer && equalError(g.Err, want.Err)
	case errorText:
		return got != nil && got.Error() == string(want)
	default:
		if got == want {
			return true
		}
		return got != nil && reflect.TypeOf(got) == reflect.TypeOf(want) && got.Error() == want.Error()
	}
}

// unsupported is an option which this package doesn't have: an option of the v1 semantics, which the root package
// is to have, or the experimental option of the format tag, which encoding/json/v2 of Go 1.27 doesn't support
// either. A case which uses it is skipped.
type unsupported string

func (unsupported) ApplyTo(*options.Config) {}

func skipUnsupported(t *testing.T, opts []Options) {
	t.Helper()
	for _, o := range opts {
		if u, ok := o.(unsupported); ok {
			t.Skipf("unsupported option %s", string(u))
		}
	}
}

// ignoredFormat stands for the options of a format which a type ignores without the format tag: it has no effect.
type ignoredFormat struct{}

func (ignoredFormat) ApplyTo(*options.Config) {}

var _ = errors.New

var (
	anyType      = reflect.TypeFor[any]()
	boolType     = reflect.TypeFor[bool]()
	stringType   = reflect.TypeFor[string]()
	float64Type  = reflect.TypeFor[float64]()
	bytesType    = reflect.TypeFor[[]byte]()
	timeTimeType = reflect.TypeFor[time.Time]()
)

var (
	errChangingDuplicateNames = errorText("cannot change duplicate name checks after a JSON object has already begun processing")
	errChangingInvalidUTF8    = errorText("cannot change UTF-8 checks after a JSON object has already begun processing")
	errChangingWhitespace     = errorText("cannot change whitespace formatting within a MarshalEncode call")
)

// isSemanticErr reports whether err is a SemanticError of the error want, which is unexported by the json package.
func isSemanticErr(err error, want errorText) bool {
	var serr *SemanticError
	return errors.As(err, &serr) && serr.Err != nil && serr.Err.Error() == string(want)
}

// isZeroer is a type with the IsZero method, which the omitzero option calls.
type isZeroer interface {
	IsZero() bool
}

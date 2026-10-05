package json

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/jsontext"
)

// The text of a SemanticError is made of the action, the kind of the JSON value, the JSON value, the Go type,
// the place and the underlying error, each of which is left out if it is not set.
func TestSemanticErrorText(t *testing.T) {
	type wide struct{ A1, A2, A3, A4, A5, A6, A7, A8, A9, A10, A11, A12 string }
	type wideUnexported struct {
		A1, A2, A3, A4, A5, A6, A7, A8, A9, A10, A11, A12 string
		b                                                 int
	}
	syntax := &jsontext.SyntacticError{JSONPointer: "/a/b", ByteOffset: 40, Err: ierrors.ErrInvalidUTF8}
	for _, c := range []struct {
		err  *SemanticError
		want string
	}{
		{&SemanticError{}, `json: cannot handle`},
		{&SemanticError{action: "marshal"}, `json: cannot marshal`},
		{&SemanticError{action: "marshal", JSONKind: '['}, `json: cannot marshal JSON array`},
		{&SemanticError{action: "marshal", JSONKind: '?'}, `json: cannot marshal`},
		{&SemanticError{action: "unmarshal", JSONKind: '0', JSONValue: jsontext.Value(`-5`), GoType: reflect.TypeFor[uint8]()},
			`json: cannot unmarshal JSON number -5 into Go uint8`},
		{&SemanticError{action: "marshal", GoType: reflect.TypeFor[chan int]()}, `json: cannot marshal from Go chan int`},
		{&SemanticError{JSONKind: 't', GoType: reflect.TypeFor[string]()}, `json: cannot handle JSON boolean with Go string`},
		{&SemanticError{GoType: reflect.TypeFor[struct {
			A1, A2, A3, A4, A5, A6, A7, A8, A9, A10, A11, A12 string
		}]()}, `json: cannot handle Go struct`},
		{&SemanticError{GoType: reflect.TypeFor[struct{ wide }]()}, `json: cannot handle Go struct { json.wide }`},
		{&SemanticError{GoType: reflect.TypeFor[struct {
			A1, A2, A3, A4, A5, A6, A7, A8, A9, A10, A11, A12 string
			b                                                 int
		}]()}, `json: cannot handle Go v2.struct`},
		{&SemanticError{GoType: reflect.TypeFor[wideUnexported]()}, `json: cannot handle Go json.wideUnexported`},
		{&SemanticError{action: "marshal", ByteOffset: 7}, `json: cannot marshal after offset 7`},
		{&SemanticError{action: "marshal", ByteOffset: 7, JSONPointer: "/x/0"}, `json: cannot marshal within "/x/0"`},
		{&SemanticError{action: "marshal", Err: errors.New("boom")}, `json: cannot marshal: boom`},
		{&SemanticError{action: "unmarshal", JSONPointer: "/x/y", GoType: reflect.TypeFor[struct{ X int }](), Err: ErrUnknownName},
			`json: cannot unmarshal into Go struct { X int }: unknown object member name "y" within "/x"`},
		{&SemanticError{action: "unmarshal", JSONPointer: "/y", Err: ErrUnknownName}, `json: cannot unmarshal: unknown object member name "y"`},
		// the place of a wrapped SyntacticError is shown once.
		{&SemanticError{JSONPointer: "/a", ByteOffset: 10, Err: syntax}, `json: cannot handle: invalid UTF-8 within "/a/b" after offset 40`},
		{&SemanticError{JSONPointer: "/c", ByteOffset: 10, Err: syntax}, `json: cannot handle within "/c": invalid UTF-8 within "/a/b" after offset 40`},
		{&SemanticError{ByteOffset: 50, Err: syntax}, `json: cannot handle after offset 50: invalid UTF-8 within "/a/b" after offset 40`},
		{&SemanticError{ByteOffset: 30, Err: syntax}, `json: cannot handle: invalid UTF-8 within "/a/b" after offset 40`},
		// a long pointer is shortened.
		{&SemanticError{JSONPointer: jsontext.Pointer(strings.Repeat("/abc", 40))},
			`json: cannot handle within "/abc/abc/abc/abc/abc/abc/abc/abc/abc/abc/abc/abc/…/abc/abc/abc/abc/abc/abc/abc/abc/abc/abc/abc/abc"`},
	} {
		if got := c.err.Error(); got != c.want {
			t.Errorf("%#v.Error():\ngot  %s\nwant %s", c.err, got, c.want)
		}
	}
	if err := (&SemanticError{Err: ErrUnknownName}); !errors.Is(err, ErrUnknownName) {
		t.Error("errors.Is(SemanticError, ErrUnknownName) = false")
	}
}

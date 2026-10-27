package json_test

import (
	"errors"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

type maxDepthList struct{ Next *maxDepthList }

// maxDepthNestedList is a list whose values nest two more levels in a field.
type maxDepthNestedList struct {
	X    [][]int
	Next *maxDepthNestedList
}

type maxDepthTree struct {
	Kid *maxDepthTree
	N   int
}

// maxDepthValues are values nested n levels deep: through interface values, a list and a recursive type which is not
// a list.
var maxDepthValues = map[string]func(n int) any{
	"interface values": func(n int) any {
		var v any = 1
		for range n {
			v = []any{v}
		}
		return v
	},
	// a value of a marshaler, which writes no level of its own.
	"interface values of time.Time": func(n int) any {
		var v any = time.Time{}
		for range n {
			v = []any{v}
		}
		return v
	},
	// values which nest levels of their own in the frame of an interface value.
	"interface values of [][]int": func(n int) any {
		var v any = [][]int{{1}}
		for range n - 2 {
			v = []any{v}
		}
		return v
	},
	"list of [][]int": func(n int) any {
		var l *maxDepthNestedList
		for range n - 2 {
			l = &maxDepthNestedList{X: [][]int{{1}}, Next: l}
		}
		return l
	},
	"list": func(n int) any {
		var l *maxDepthList
		for range n {
			l = &maxDepthList{l}
		}
		return l
	},
	"tree": func(n int) any {
		var t *maxDepthTree
		for range n {
			t = &maxDepthTree{Kid: t}
		}
		return t
	},
}

// maxDepthDeclined is an array of a struct of a slice of the array, which nests three levels, every one of which a
// function declines: it is written by its default representation, in a context of its own.
type maxDepthDeclined [1]struct {
	V    int
	Next []maxDepthDeclined
}

// maxDepthPtrMarshaler is a struct whose MarshalJSON has a pointer receiver.
type maxDepthPtrMarshaler struct{ X int }

func (*maxDepthPtrMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"called"`), nil }

// maxDepthAddressable is maxDepthDeclined of a field of maxDepthPtrMarshaler, whose values in the slices are written
// by their default representations, addressable, with CallMethodsWithLegacySemantics.
type maxDepthAddressable [1]struct {
	V    maxDepthPtrMarshaler
	Next []maxDepthAddressable
}

// A value written by default representations, in contexts of their own, has the max depth of the others: 3332 levels
// of three are 9999 deep, and the innermost value one more, as encoding/json/v2 writes them, and 3333 are deeper.
func TestMarshalMaxDepthOfDefaultRepresentations(t *testing.T) {
	declines := json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, any) error { return errors.ErrUnsupported }))
	legacy := gojson.CallMethodsWithLegacySemantics(true)
	for name, c := range map[string]struct {
		of   func(n int) any
		opts json.Options
	}{
		"declined": {func(n int) any {
			var v maxDepthDeclined
			for range n {
				v = maxDepthDeclined{{Next: []maxDepthDeclined{v}}}
			}
			return v
		}, declines},
		"addressable": {func(n int) any {
			var v maxDepthAddressable
			for range n {
				v = maxDepthAddressable{{Next: []maxDepthAddressable{v}}}
			}
			return v
		}, legacy},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := json.Marshal(c.of(3332), c.opts); err != nil {
				t.Errorf("3332 levels: %v", err)
			}
			_, err := json.Marshal(c.of(3333), c.opts)
			var serr *jsontext.SyntacticError
			if !errors.As(err, &serr) {
				t.Errorf("3333 levels: got %v, want the error of the max depth", err)
			}
		})
	}
}

// A value nested deeper than 10000 objects and arrays fails, as encoding/json/v2 fails for it.
func TestMarshalMaxDepth(t *testing.T) {
	for name, of := range maxDepthValues {
		t.Run(name, func(t *testing.T) {
			if _, err := json.Marshal(of(10000)); err != nil {
				t.Errorf("10000 levels: %v", err)
			}
			_, err := json.Marshal(of(10001))
			var serr *jsontext.SyntacticError
			if !errors.As(err, &serr) {
				t.Errorf("10001 levels: got %v, want the error of the max depth", err)
			}
		})
	}
}

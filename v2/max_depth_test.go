package json_test

import (
	"errors"
	"testing"
	"time"

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

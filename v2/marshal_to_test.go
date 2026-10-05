package json

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
)

// trackedValue writes its value by MarshalJSONTo, and records the pointer which the encoder gives it.
type trackedValue struct {
	V        any
	pointers *[]jsontext.Pointer
}

func (v trackedValue) MarshalJSONTo(enc *jsontext.Encoder) error {
	*v.pointers = append(*v.pointers, enc.StackPointer())
	return MarshalEncode(enc, v.V)
}

type trackedOmit struct {
	A trackedValue  `json:"a,omitempty"`
	B *trackedValue `json:"b,omitempty"`
	C []any         `json:"c,omitempty"`
	D int           `json:"d"`
}

// The levels which the encoder follows as its output grows are the ones of the whole output, also after a part
// of it is written again: a member which omitempty takes back, the entries of a map which are sorted, and the
// comma of an embedded fallback.
func TestTrackedLevels(t *testing.T) {
	checkTracked = func(out []byte, levels []textcoder.Level) {
		want := levelsOf(out, nil)
		if normalized(levels) != normalized(want) {
			t.Fatalf("levels of %s:\ngot  %+v\nwant %+v", out, levels, want)
		}
	}
	defer func() { checkTracked = nil }()

	r := rand.New(rand.NewPCG(1, 2))
	var pointers []jsontext.Pointer
	var value func(depth int) any
	value = func(depth int) any {
		n := r.IntN(8)
		if depth > 4 {
			n = 0
		}
		switch n {
		case 0:
			return trackedValue{V: r.IntN(100), pointers: &pointers}
		case 1:
			return trackedValue{V: strings.Repeat("x", r.IntN(3000)), pointers: &pointers}
		case 2:
			list := make([]any, r.IntN(5))
			for i := range list {
				list[i] = value(depth + 1)
			}
			return list
		case 3:
			m := map[int]any{}
			for i := range r.IntN(5) {
				m[i*7] = value(depth + 1)
			}
			return m
		case 4:
			return trackedOmit{
				A: trackedValue{V: map[string]any{}, pointers: &pointers},
				B: &trackedValue{V: "", pointers: &pointers},
				C: []any{value(depth + 1)},
				D: r.IntN(10),
			}
		case 5:
			return struct {
				X any            `json:"x"`
				M map[string]any `json:",embed"`
			}{X: value(depth + 1), M: map[string]any{"m": value(depth + 1)}}
		case 6:
			return trackedValue{V: []any{value(depth + 1), value(depth + 1)}, pointers: &pointers}
		default:
			return map[string]any{"k": value(depth + 1), "l": []any{value(depth + 1)}}
		}
	}
	for i := range 200 {
		v := value(0)
		for _, opts := range [][]Options{nil, {Deterministic(true)}} {
			pointers = pointers[:0]
			if _, err := Marshal(v, opts...); err != nil {
				t.Fatalf("%d: Marshal: %v", i, err)
			}
		}
	}
}

// normalized is the levels as text, which compares the names of the objects by their bytes.
func normalized(levels []textcoder.Level) string {
	var b strings.Builder
	for _, l := range levels {
		fmt.Fprintf(&b, "{%v %d %q}", l.Object, l.Count, l.Name)
	}
	return b.String()
}

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

type cycleDetectionHolder struct {
	V any
}

type cycleDetectionNode struct {
	Next *cycleDetectionNode
	Any  any
}

// nestedInterfaceValue wraps the value by the slices of an interface value.
func nestedInterfaceValue(depth int, v any) any {
	for i := 0; i < depth; i++ {
		v = []any{v}
	}
	return v
}

// The values are checked for a cycle after the nesting gets deep.
// A value which is referred to twice is not a cycle, whatever it holds.
func TestEncodeSharedValueNestedDeeply(t *testing.T) {
	const depth = 1100
	shared := &cycleDetectionHolder{V: (*int)(nil)}
	for _, test := range []struct {
		name string
		v    any
	}{
		{"typed nil in interface", []*cycleDetectionHolder{shared, shared}},
		{"value in interface", []any{shared, shared}},
		{"recursive type", []*cycleDetectionNode{{Any: shared}, {Any: shared}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			v := nestedInterfaceValue(depth, test.v)
			expected, err := stdjson.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(expected, got) {
				t.Fatalf("expected %d bytes but got %d bytes", len(expected), len(got))
			}
		})
	}
}

func TestEncodeCycleNestedDeeply(t *testing.T) {
	t.Run("interface", func(t *testing.T) {
		cycle := []any{nil}
		cycle[0] = cycle
		for _, depth := range []int{0, 1100} {
			if _, err := json.Marshal(nestedInterfaceValue(depth, cycle)); err == nil {
				t.Fatalf("depth %d: expected an error", depth)
			}
		}
	})
	t.Run("recursive type", func(t *testing.T) {
		node := &cycleDetectionNode{}
		node.Next = &cycleDetectionNode{Next: node}
		for _, depth := range []int{0, 1100} {
			if _, err := json.Marshal(nestedInterfaceValue(depth, node)); err == nil {
				t.Fatalf("depth %d: expected an error", depth)
			}
		}
	})
}

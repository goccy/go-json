package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// A recursive struct is encoded by jumping to its own opcodes with a new frame of the pointers.
// The frame must not overlap the one of the caller, whatever the struct has before the recursive field.

type recursiveMapNode struct {
	Meta     map[string]any     `json:"meta"`
	Children []recursiveMapNode `json:"children"`
}

type recursiveLeadingFieldsNode struct {
	EvaluationPath string                       `json:"evaluationPath"`
	SchemaLocation string                       `json:"schemaLocation"`
	Annotations    map[string]any               `json:"annotations"`
	Details        []recursiveLeadingFieldsNode `json:"details"`
}

type recursiveSliceOfInterfaceNode struct {
	Alerts [][]any
	Items  []recursiveSliceOfInterfaceNode
}

type recursiveInterfaceSliceNode struct {
	Alerts []any
	Items  []recursiveSliceOfInterfaceNode
}

type recursiveInterfaceNode struct {
	Value    any
	Children []recursiveInterfaceNode
}

type recursivePtrNode struct {
	Meta map[string]any
	Next *recursivePtrNode
}

func TestEncodeRecursiveStructWithInterface(t *testing.T) {
	for _, test := range []struct {
		name string
		v    any
	}{
		{"map of interface in nested node", []recursiveMapNode{{Children: []recursiveMapNode{{Meta: map[string]any{"a": 1}}}}}},
		{"map of interface in every node", recursiveMapNode{
			Meta: map[string]any{"a": 1},
			Children: []recursiveMapNode{
				{Meta: map[string]any{"b": "x"}, Children: []recursiveMapNode{{Meta: map[string]any{"c": true}}}},
				{Meta: map[string]any{"d": []any{1, "y"}}},
			},
		}},
		{"leading fields and map of interface", recursiveLeadingFieldsNode{
			Details: []recursiveLeadingFieldsNode{{Annotations: map[string]any{"description": "d"}}},
		}},
		{"slice of slice of interface", recursiveSliceOfInterfaceNode{
			Items: []recursiveSliceOfInterfaceNode{{Alerts: [][]any{{2}}}},
		}},
		{"slice of interface", recursiveInterfaceSliceNode{
			Alerts: []any{1},
			Items:  []recursiveSliceOfInterfaceNode{{Alerts: [][]any{{2, "a"}}}},
		}},
		{"interface holding the recursive struct", recursiveInterfaceNode{
			Value:    recursiveInterfaceNode{Value: 1},
			Children: []recursiveInterfaceNode{{Value: map[string]any{"k": recursiveInterfaceNode{Value: "v"}}}},
		}},
		{"pointer recursion with map of interface", &recursivePtrNode{
			Meta: map[string]any{"a": 1},
			Next: &recursivePtrNode{Meta: map[string]any{"b": map[string]any{"c": 2}}, Next: &recursivePtrNode{}},
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			expected, err := stdjson.Marshal(test.v)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				got, err := json.Marshal(test.v)
				if err != nil {
					t.Fatalf("call %d: %v", i, err)
				}
				if string(got) != string(expected) {
					t.Fatalf("call %d: expected %s but got %s", i, expected, got)
				}
			}
		})
	}
}

// A recursive struct which is embedded, or which is referred to from an embedded struct.
// The opcodes of an embedded struct are only the fields, and the ones of a struct have the braces and
// the check of nil, so a recursive code must jump to the right ones.

type recursiveEmbeddedInner struct {
	Inner *recursiveEmbeddedInner `json:"inner"`
}

type recursiveEmbeddedByPointer struct {
	*recursiveEmbeddedInner
}

type recursiveEmbeddedByValue struct {
	recursiveEmbeddedInner
	Name string `json:"name"`
}

type recursiveCustomer struct {
	IDs []*recursiveIDMapping
}

type recursiveIDMapping struct {
	Customer *recursiveCustomer
}

type recursiveCustomerHolder struct {
	recursiveCustomer
	MyField string
}

// the recursive struct is embedded in the struct which it refers to.
type recursiveLabel struct {
	Parts []*recursiveResLabel `json:"parts"`
}

type recursiveResLabel struct {
	ID int `json:"id"`
	recursiveLabel
}

func TestEncodeEmbeddedRecursiveStruct(t *testing.T) {
	for _, test := range []struct {
		name string
		v    any
	}{
		{"nil embedded pointer", &recursiveEmbeddedByPointer{}},
		{"embedded pointer", &recursiveEmbeddedByPointer{&recursiveEmbeddedInner{}}},
		{"embedded pointer with recursion", &recursiveEmbeddedByPointer{&recursiveEmbeddedInner{&recursiveEmbeddedInner{&recursiveEmbeddedInner{}}}}},
		{"embedded value", recursiveEmbeddedByValue{Name: "n"}},
		{"embedded value with recursion", recursiveEmbeddedByValue{recursiveEmbeddedInner{&recursiveEmbeddedInner{}}, "n"}},
		{"embedded struct referring to itself through a slice", []recursiveCustomerHolder{{MyField: "111"}}},
		{"populated embedded struct referring to itself", []recursiveCustomerHolder{{
			recursiveCustomer: recursiveCustomer{IDs: []*recursiveIDMapping{{}, {Customer: &recursiveCustomer{IDs: []*recursiveIDMapping{{}}}}}},
			MyField:           "111",
		}}},
		{"struct embedded in the struct it refers to", &recursiveLabel{Parts: []*recursiveResLabel{{ID: 1}}}},
		{"nested struct embedded in the struct it refers to", &recursiveLabel{Parts: []*recursiveResLabel{
			{ID: 1, recursiveLabel: recursiveLabel{Parts: []*recursiveResLabel{{ID: 2}, nil}}},
		}}},
		{"start from the embedding struct", &recursiveResLabel{ID: 1, recursiveLabel: recursiveLabel{Parts: []*recursiveResLabel{{ID: 2}}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			expected, err := stdjson.Marshal(test.v)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				got, err := json.Marshal(test.v)
				if err != nil {
					t.Fatalf("call %d: %v", i, err)
				}
				if string(got) != string(expected) {
					t.Fatalf("call %d: expected %s but got %s", i, expected, got)
				}
			}
		})
	}
}

// The frame of the pointers of a recursive code must not overlap the frame it is jumped from.
// Otherwise the indent to restore, which is kept at the end of a frame, is overwritten by a pointer,
// and gigabytes of the indent are written.

type recursiveIndentNode struct {
	Next *recursiveIndentNode `json:"next"`
}

type recursiveIndentTree struct {
	Name     string                `json:"name"`
	Children []recursiveIndentTree `json:"children"`
}

type recursiveIndentMutualA struct {
	B *recursiveIndentMutualB `json:"b"`
}

type recursiveIndentMutualB struct {
	Value int                     `json:"value"`
	A     *recursiveIndentMutualA `json:"a"`
}

func TestEncodeDeepRecursiveStructWithIndent(t *testing.T) {
	const maxDepth = 6
	values := map[string][]any{}
	for depth := 1; depth <= maxDepth; depth++ {
		node := &recursiveIndentNode{}
		tree := recursiveIndentTree{Name: "leaf"}
		mutual := &recursiveIndentMutualA{}
		for i := 1; i < depth; i++ {
			node = &recursiveIndentNode{Next: node}
			tree = recursiveIndentTree{Name: "node", Children: []recursiveIndentTree{tree, {Name: "sibling"}}}
			mutual = &recursiveIndentMutualA{B: &recursiveIndentMutualB{Value: i, A: mutual}}
		}
		values["pointer"] = append(values["pointer"], node)
		values["slice"] = append(values["slice"], tree)
		values["mutual"] = append(values["mutual"], mutual)
		values["in interface"] = append(values["in interface"], []any{node, map[string]any{"k": tree}})
	}
	for name, vs := range values {
		t.Run(name, func(t *testing.T) {
			for i, v := range vs {
				depth := i + 1
				expected, err := stdjson.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				got, err := json.Marshal(v)
				if err != nil {
					t.Fatalf("depth %d: %v", depth, err)
				}
				if string(got) != string(expected) {
					t.Fatalf("depth %d: expected %s but got %s", depth, expected, got)
				}
				expectedIndent, err := stdjson.MarshalIndent(v, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				gotIndent, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					t.Fatalf("depth %d ( indent ): %v", depth, err)
				}
				if string(gotIndent) != string(expectedIndent) {
					// the broken output can be gigabytes, so it is not printed.
					t.Fatalf("depth %d ( indent ): expected %d bytes but got %d bytes", depth, len(expectedIndent), len(gotIndent))
				}
			}
		})
	}
}

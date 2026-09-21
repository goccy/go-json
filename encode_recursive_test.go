package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// A recursive struct is encoded by jumping to its own opcodes with a new frame of the pointers.
// The frame must not overlap the one of the caller, whatever the struct has before the recursive field.

type recursiveMapNode struct {
	Meta     map[string]interface{} `json:"meta"`
	Children []recursiveMapNode     `json:"children"`
}

type recursiveLeadingFieldsNode struct {
	EvaluationPath string                       `json:"evaluationPath"`
	SchemaLocation string                       `json:"schemaLocation"`
	Annotations    map[string]interface{}       `json:"annotations"`
	Details        []recursiveLeadingFieldsNode `json:"details"`
}

type recursiveSliceOfInterfaceNode struct {
	Alerts [][]interface{}
	Items  []recursiveSliceOfInterfaceNode
}

type recursiveInterfaceSliceNode struct {
	Alerts []interface{}
	Items  []recursiveSliceOfInterfaceNode
}

type recursiveInterfaceNode struct {
	Value    interface{}
	Children []recursiveInterfaceNode
}

type recursivePtrNode struct {
	Meta map[string]interface{}
	Next *recursivePtrNode
}

func TestEncodeRecursiveStructWithInterface(t *testing.T) {
	for _, test := range []struct {
		name string
		v    interface{}
	}{
		{"map of interface in nested node", []recursiveMapNode{{Children: []recursiveMapNode{{Meta: map[string]interface{}{"a": 1}}}}}},
		{"map of interface in every node", recursiveMapNode{
			Meta: map[string]interface{}{"a": 1},
			Children: []recursiveMapNode{
				{Meta: map[string]interface{}{"b": "x"}, Children: []recursiveMapNode{{Meta: map[string]interface{}{"c": true}}}},
				{Meta: map[string]interface{}{"d": []interface{}{1, "y"}}},
			},
		}},
		{"leading fields and map of interface", recursiveLeadingFieldsNode{
			Details: []recursiveLeadingFieldsNode{{Annotations: map[string]interface{}{"description": "d"}}},
		}},
		{"slice of slice of interface", recursiveSliceOfInterfaceNode{
			Items: []recursiveSliceOfInterfaceNode{{Alerts: [][]interface{}{{2}}}},
		}},
		{"slice of interface", recursiveInterfaceSliceNode{
			Alerts: []interface{}{1},
			Items:  []recursiveSliceOfInterfaceNode{{Alerts: [][]interface{}{{2, "a"}}}},
		}},
		{"interface holding the recursive struct", recursiveInterfaceNode{
			Value:    recursiveInterfaceNode{Value: 1},
			Children: []recursiveInterfaceNode{{Value: map[string]interface{}{"k": recursiveInterfaceNode{Value: "v"}}}},
		}},
		{"pointer recursion with map of interface", &recursivePtrNode{
			Meta: map[string]interface{}{"a": 1},
			Next: &recursivePtrNode{Meta: map[string]interface{}{"b": map[string]interface{}{"c": 2}}, Next: &recursivePtrNode{}},
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
		v    interface{}
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

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

type nestedFrameNode struct {
	Name  string
	Tags  []string
	Any   interface{}       `json:",omitempty"`
	Next  *nestedFrameNode  `json:",omitempty"`
	Kids  []nestedFrameNode `json:",omitempty"`
	Attrs map[string]*nestedFrameNode
}

func nestedFrameList(depth int) *nestedFrameNode {
	var head *nestedFrameNode
	for i := 0; i < depth; i++ {
		head = &nestedFrameNode{
			Name: fmt.Sprint("node", i),
			Tags: []string{"a", "b"},
			Next: head,
		}
	}
	return head
}

func nestedFrameTree(depth int) nestedFrameNode {
	node := nestedFrameNode{Name: fmt.Sprint("tree", depth), Any: depth}
	if depth > 0 {
		node.Kids = []nestedFrameNode{nestedFrameTree(depth - 1), nestedFrameTree(depth - 1)}
		node.Attrs = map[string]*nestedFrameNode{"attr": {Name: "leaf", Any: []interface{}{depth, "x"}}}
	}
	return node
}

func nestedFrameMap(depth int) interface{} {
	var v interface{} = "leaf"
	for i := 0; i < depth; i++ {
		v = map[string]interface{}{"v": v, "list": []interface{}{i, nil, "s"}}
	}
	return v
}

// The VM has a fixed number of slots for the frames of the interface values and of the recursive types,
// and it goes on with new slots when a frame doesn't fit. The values here are nested deeply enough to do it
// many times, at every position of a frame in the slots.
func TestEncodeNestedFrames(t *testing.T) {
	for _, test := range []struct {
		name string
		v    interface{}
	}{
		{"recursive list", nestedFrameList(300)},
		{"recursive tree", nestedFrameTree(8)},
		{"map of interface", nestedFrameMap(200)},
		{"recursive in interface", nestedInterfaceValue(50, nestedFrameList(50))},
		{"interface in recursive", &nestedFrameNode{Name: "root", Any: nestedFrameMap(40), Next: &nestedFrameNode{Any: nestedFrameTree(4)}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for depth := 0; depth < 12; depth++ {
				// the depth shifts the position of the frames in the slots.
				v := nestedInterfaceValue(depth, test.v)
				t.Run("Marshal", func(t *testing.T) {
					expected, err := stdjson.Marshal(v)
					if err != nil {
						t.Fatal(err)
					}
					got, err := json.Marshal(v)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(expected, got) {
						t.Fatalf("depth %d: expected %d bytes but got %d bytes", depth, len(expected), len(got))
					}
				})
				t.Run("MarshalIndent", func(t *testing.T) {
					expected, err := stdjson.MarshalIndent(v, "", "  ")
					if err != nil {
						t.Fatal(err)
					}
					got, err := json.MarshalIndent(v, "", "  ")
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(expected, got) {
						t.Fatalf("depth %d: expected %d bytes but got %d bytes", depth, len(expected), len(got))
					}
				})
			}
		})
	}
}

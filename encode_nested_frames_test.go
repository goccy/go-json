package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"runtime"
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

// The VM allocates the frames of the interface values and of the recursive types one after another in its slots,
// which grow when the next frame doesn't fit. The values here are nested deeply enough to make them grow many
// times, with the frames at various positions in the slots.
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
				// the depth shifts the positions of the frames in the slots.
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

type nestedTypeLeaf struct {
	M map[string][]int
}

type nestedTypeLevel4 struct{ X []nestedTypeLeaf }
type nestedTypeLevel3 struct{ X []nestedTypeLevel4 }
type nestedTypeLevel2 struct{ X []nestedTypeLevel3 }
type nestedTypeLevel1 struct{ X []nestedTypeLevel2 }
type nestedTypeLevel0 struct{ X []nestedTypeLevel1 }

// The length of a frame depends on how deep the values are nested in a type, and it has no limit.
func TestEncodeDeeplyNestedType(t *testing.T) {
	slices := [][][][][][][][][][][][][][][][]int{{{{{{{{{{{{{{{{1, 2}, {3}}}}}}}}}}}}}}}}
	structs := []nestedTypeLevel0{{X: []nestedTypeLevel1{{X: []nestedTypeLevel2{{X: []nestedTypeLevel3{{X: []nestedTypeLevel4{
		{X: []nestedTypeLeaf{{M: map[string][]int{"a": {1, 2}, "b": nil}}, {}}},
		{},
	}}}}}}}}}
	for _, v := range []interface{}{slices, structs} {
		expected, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expected, got) {
			t.Fatalf("expected %s but got %s", expected, got)
		}
	}
}

type gcDuringMapValue int

func (v gcDuringMapValue) MarshalJSON() ([]byte, error) {
	runtime.GC()
	return []byte(fmt.Sprint(int(v))), nil
}

// The VM refers to the context of a map only by a slot, which the GC doesn't see.
// The context has to stay alive while the map, and the maps in it, are encoded.
func TestEncodeMapWhileGCRuns(t *testing.T) {
	v := map[string]map[string]gcDuringMapValue{
		"a": {"x": 1, "y": 2},
		"b": {"z": 3},
		"c": {},
	}
	expected := `{"a":{"x":1,"y":2},"b":{"z":3},"c":{}}`
	for i := 0; i < 20; i++ {
		got, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != expected {
			t.Fatalf("expected %s but got %s", expected, got)
		}
	}
}

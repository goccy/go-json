package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// A value of a recursive type which is the last field of a value of the same type is encoded in the frame of
// that value, without a frame of its own: a list of such values must encode as encoding/json does whatever
// its length, with an indent, with other values in the frame ( an interface value, another recursive type,
// a slice of the type ) and after a cycle.

type tailNode struct {
	V    int       `json:"v"`
	I    any       `json:"i,omitempty"`
	Next *tailNode `json:"next"`
}

type tailNodeOmitEmpty struct {
	V    int                `json:"v"`
	Next *tailNodeOmitEmpty `json:"next,omitempty"`
}

type tailTree struct {
	Children []*tailTree `json:"children,omitempty"`
	Next     *tailTree   `json:"next"`
}

type tailOther struct {
	V    int        `json:"v"`
	Next *tailOther `json:"next"`
}

type tailMixed struct {
	Other *tailOther `json:"other"`
	Next  *tailMixed `json:"next"`
}

// the recursive field which is not the last is not a tail.
type tailFirst struct {
	Next *tailFirst `json:"next"`
	V    int        `json:"v"`
}

func newTailList(n int) *tailNode {
	var head *tailNode
	for i := n; i > 0; i-- {
		head = &tailNode{V: i, Next: head}
		if i%3 == 0 {
			head.I = &tailNode{V: -i}
		}
	}
	return head
}

func TestEncodeTailRecursion(t *testing.T) {
	deep := newTailList(1500) // more than the level cycles are detected from
	var omit *tailNodeOmitEmpty
	for i := 3; i > 0; i-- {
		omit = &tailNodeOmitEmpty{V: i, Next: omit}
	}
	var first *tailFirst
	for i := 3; i > 0; i-- {
		first = &tailFirst{V: i, Next: first}
	}
	values := []any{
		(*tailNode)(nil),
		&tailNode{},
		newTailList(1),
		newTailList(2),
		newTailList(3),
		newTailList(10),
		deep,
		[]*tailNode{newTailList(2), nil, newTailList(1)},
		map[string]*tailNode{"a": newTailList(2)},
		struct {
			L *tailNode `json:"l"`
			N int       `json:"n"`
		}{newTailList(2), 1},
		omit,
		&tailTree{Children: []*tailTree{{Next: &tailTree{}}, {}}, Next: &tailTree{Children: []*tailTree{{}}}},
		&tailMixed{Other: &tailOther{1, &tailOther{2, nil}}, Next: &tailMixed{Other: &tailOther{3, nil}}},
		first,
		// a value with a tail after a tail: the interface value holds a list of its own.
		&tailNode{V: 1, I: newTailList(3), Next: &tailNode{V: 2, I: newTailList(2)}},
	}
	for _, v := range values {
		expected, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T:\n got %.300s\nwant %.300s", v, got, expected)
		}
		expected, _ = stdjson.MarshalIndent(v, "", "  ")
		got, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T with indent:\n got %.500s\nwant %.500s", v, got, expected)
		}
	}
	// a cycle is an error, and the context is left usable.
	cycle := newTailList(3)
	cycle.Next.Next.Next = cycle
	if _, err := json.Marshal(cycle); err == nil {
		t.Fatal("a cycle must be an error")
	}
	if _, err := json.Marshal(newTailList(3)); err != nil {
		t.Fatal(err)
	}
}

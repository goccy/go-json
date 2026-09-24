//go:build go1.24

package json_test

import (
	"runtime"
	"strings"
	"testing"
	"weak"

	json "github.com/goccy/go-json"
)

// The decoders keep buffers and contexts in pools between the calls. Nothing in them may refer to
// what a call decoded: a decoded string refers to the copy of the input, so a pointer left behind
// keeps the whole input of a previous call alive, and a program which decodes large inputs at a high
// rate keeps as many of them alive as its pools hold objects.

type retentionNode struct {
	Name string
	Kids []*retentionNode
}

func retentionInput() []byte {
	var b strings.Builder
	b.WriteString(`{"Name":"root","Kids":[`)
	for i := 0; i < 64; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"Name":"` + strings.Repeat("n", 64) + `","Kids":[{"Name":"leaf","Kids":[]},{"Name":"leaf","Kids":[]}]}`)
	}
	b.WriteString(`]}`)
	return []byte(b.String())
}

func decodedIsCollected(t *testing.T, decode func() any) {
	t.Helper()
	var refs []weak.Pointer[retentionNode]
	for i := 0; i < 8; i++ {
		switch v := decode().(type) {
		case *retentionNode:
			refs = append(refs, weak.Make(v), weak.Make(v.Kids[0]), weak.Make(v.Kids[63].Kids[1]))
		case map[string]any:
			// the values of interface{} are checked through the node decoded beside them
			refs = append(refs, weak.Make(v["node"].(*retentionNode)))
		}
	}
	// A pool keeps its objects over one collection ( its victim cache ): what they refer to is alive
	// through that collection, so one collection tells whether a pooled object refers to a decoded value.
	runtime.GC()
	for i, r := range refs {
		if r.Value() != nil {
			t.Fatalf("a value decoded by a previous call is still alive ( %d )", i)
		}
	}
}

func TestDecodeKeepsNothingOfPreviousCall(t *testing.T) {
	input := retentionInput()
	t.Run("Unmarshal", func(t *testing.T) {
		decodedIsCollected(t, func() any {
			v := new(retentionNode)
			if err := json.Unmarshal(input, v); err != nil {
				t.Fatal(err)
			}
			return v
		})
	})
	t.Run("Decoder", func(t *testing.T) {
		decodedIsCollected(t, func() any {
			v := new(retentionNode)
			if err := json.NewDecoder(strings.NewReader(string(input))).Decode(v); err != nil {
				t.Fatal(err)
			}
			return v
		})
	})
}

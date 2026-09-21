package json_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

type lifetimeNode struct {
	Name  string
	Next  *lifetimeNode
	Any   interface{}
	Map   map[string]*lifetimeNode
	Slice []*lifetimeNode
}

// The encoder refers to the values being encoded from its context, which is pooled.
// The context must not keep them alive after the encoding.
func TestEncodedValueIsNotKeptAlive(t *testing.T) {
	for _, test := range []struct {
		name    string
		marshal func(interface{}) ([]byte, error)
	}{
		{"Marshal", json.Marshal},
		{"MarshalIndent", func(v interface{}) ([]byte, error) { return json.MarshalIndent(v, "", "  ") }},
		{"Error", func(v interface{}) ([]byte, error) {
			// the encoding stops at the last value, so the context still refers to the values.
			values := v.(*lifetimeNode).Slice
			values[len(values)-1].Any = make(chan int)
			if _, err := json.Marshal(v); err == nil {
				t.Error("expected an error")
			}
			return nil, nil
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			const num = 4
			collected := make(chan struct{}, num)
			func() {
				leaf := &lifetimeNode{Name: "leaf"}
				// a finalizer of a value is not called until the values referring to it are freed,
				// so the leaf has no finalizer.
				values := []*lifetimeNode{
					{Name: "next", Next: leaf},
					{Name: "any", Any: leaf},
					{Name: "map", Map: map[string]*lifetimeNode{"a": leaf}},
					{Name: "slice", Slice: []*lifetimeNode{leaf}},
				}
				for _, v := range values {
					runtime.SetFinalizer(v, func(*lifetimeNode) { collected <- struct{}{} })
				}
				if _, err := test.marshal(&lifetimeNode{Name: "root", Slice: values}); err != nil {
					t.Fatal(err)
				}
			}()
			// Only one GC: the pool releases what it has after a few, whatever it holds.
			runtime.GC()
			for i := 0; i < num; i++ {
				select {
				case <-collected:
				case <-time.After(10 * time.Second):
					t.Fatalf("%d of %d values are kept alive after the encoding", num-i, num)
				}
			}
		})
	}
}

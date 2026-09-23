package json_test

import (
	stdjson "encoding/json"
	"sync"
	"testing"

	"github.com/goccy/go-json"
)

// The stack of a goroutine is copied when it grows, and only the pointers are updated.
// The encoder must not read the value being encoded from the stack which has been freed.

type stackGrowingMarshaler struct{}

//go:noinline
func growStack(n int) int {
	var pad [256]int
	pad[n%256] = n
	if n == 0 {
		return pad[0]
	}
	return growStack(n-1) + pad[n%256]
}

// MarshalJSON grows the stack while the value which has it is being encoded,
// and makes the other goroutines reuse and overwrite the stack which has been freed.
func (stackGrowingMarshaler) MarshalJSON() ([]byte, error) {
	growStack(200)
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			growStack(50)
		}()
	}
	wg.Wait()
	return []byte("1"), nil
}

type stackCopiedValue struct {
	G stackGrowingMarshaler
	A string
	B int
	C string
}

func TestEncodeWhileStackIsCopied(t *testing.T) {
	expected, err := stdjson.Marshal(stackCopiedValue{A: "aaaaaaaaaaaaaaaa", B: 123456789, C: "cccccccccccccccc"})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		marshal func(any) ([]byte, error)
	}{
		{"Marshal", json.Marshal},
		{"MarshalNoEscape", json.MarshalNoEscape},
		{"MarshalIndent", func(v any) ([]byte, error) { return json.MarshalIndent(v, "", "") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "MarshalIndent" {
				compact, err := stdjson.MarshalIndent(stackCopiedValue{A: "aaaaaaaaaaaaaaaa", B: 123456789, C: "cccccccccccccccc"}, "", "")
				if err != nil {
					t.Fatal(err)
				}
				expected = compact
			}
			var (
				wg    sync.WaitGroup
				mu    sync.Mutex
				wrong []string
			)
			for i := 0; i < 200; i++ {
				wg.Add(1)
				// a new goroutine starts with a small stack.
				go func() {
					defer wg.Done()
					v := stackCopiedValue{A: "aaaaaaaaaaaaaaaa", B: 123456789, C: "cccccccccccccccc"}
					got, err := test.marshal(&v)
					if err != nil || string(got) != string(expected) {
						mu.Lock()
						wrong = append(wrong, string(got))
						mu.Unlock()
					}
				}()
			}
			wg.Wait()
			if len(wrong) != 0 {
				t.Fatalf("%d / 200 results are wrong. expected %s but got %s", len(wrong), expected, wrong[0])
			}
		})
	}
}

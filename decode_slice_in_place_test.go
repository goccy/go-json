package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

func TestDecodeSliceLengths(t *testing.T) {
	// An array of any length is decoded into a slice of any length and capacity, whose existing elements are
	// decoded into, as encoding/json decodes it: the elements are decoded into the slice, and the ones of a long
	// array into a buffer first.
	type elem struct {
		A int
		B string
		C []int
	}
	for n := 0; n <= 40; n++ {
		var b strings.Builder
		b.WriteByte('[')
		for i := 0; i < n; i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, `{"A":%d,"C":[%d]}`, i, i)
		}
		b.WriteByte(']')
		doc := []byte(b.String())
		for _, prev := range []struct{ len, cap int }{{0, 0}, {3, 3}, {3, 20}, {20, 20}, {0, 40}, {40, 64}} {
			newSlice := func() []elem {
				s := make([]elem, prev.len, prev.cap)
				for i := range s[:prev.cap] {
					s[:prev.cap][i] = elem{A: -1, B: fmt.Sprint("old", i), C: []int{-1}}
				}
				return s
			}
			want, got := newSlice(), newSlice()
			if err := stdjson.Unmarshal(doc, &want); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(doc, &got); err != nil {
				t.Fatalf("%d elements into len %d cap %d: %v", n, prev.len, prev.cap, err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%d elements into len %d cap %d:\n got %+v\nwant %+v", n, prev.len, prev.cap, got, want)
			}
		}
	}
}

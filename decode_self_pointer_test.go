package json_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

// selfPointer is a pointer which points to its own type, and cyclePointer one of two pointer types which point
// to each other: no value but null can be decoded into them.
type (
	selfPointer   *selfPointer
	cyclePointer  *cyclePointer2
	cyclePointer2 *cyclePointer
)

// A pointer type which points to pointers only, endlessly, takes null, and any other value is a type error, at the
// root, in a field and as the key of a map. encoding/json doesn't end on them.
func TestDecodeSelfPointerType(t *testing.T) {
	tests := []struct {
		in      string
		value   func() any
		wantErr bool
	}{
		{`null`, func() any { return new(selfPointer) }, false},
		{`1`, func() any { return new(selfPointer) }, true},
		{`null`, func() any { return new(cyclePointer) }, false},
		{`{}`, func() any { return new(cyclePointer) }, true},
		{`{"F":null,"G":2}`, func() any {
			return new(struct {
				F selfPointer
				G int
			})
		}, false},
		{`{"F":1,"G":2}`, func() any {
			return new(struct {
				F selfPointer
				G int
			})
		}, true},
		{`{"F":[1],"G":2}`, func() any {
			return new(struct {
				F *cyclePointer2
				G int
			})
		}, true},
		{`{"5":1}`, func() any { return new(map[selfPointer]int) }, true},
		{`{"5":1}`, func() any { return new(map[*cyclePointer]int) }, true},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%T%s", test.value(), test.in), func(t *testing.T) {
			for _, decode := range []func(v any) error{
				func(v any) error { return json.Unmarshal([]byte(test.in), v) },
				func(v any) error { return json.NewDecoder(bytes.NewReader([]byte(test.in))).Decode(v) },
			} {
				err := decode(test.value())
				if test.wantErr && mapKeyKindError(err) != "type" {
					t.Errorf("got error %v, want a type error", err)
				}
				if !test.wantErr && err != nil {
					t.Errorf("got error %v", err)
				}
			}
		})
	}
}

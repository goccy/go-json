//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// A key of pointers to an interface{} is named by the dynamic value at the end of its pointers and interfaces, where
// encoding/json is built on encoding/json/v2 and encodes the keys of interface types ( see
// TestEncodeMapInterfaceKeys ): a number is a string, where it was written as a number, a float or a struct is not
// a name, where they were written as values, which is not JSON, and a key of an interface which holds a pointer to
// an interface is named, where it failed.
func TestEncodeMapKeyPointersToInterfaces(t *testing.T) {
	str, num, text := any("s"), any(5), any(&pointerTextKey{"q"})
	float, object := any(1.5), any(struct{ A int }{1})
	strp := &str
	deep := new(any)
	last := deep
	for range 3000 {
		next := new(any)
		*last = next
		last = next
	}
	*last = "end"
	for name, v := range map[string]any{
		"a string":                      map[*any]int{&str: 1},
		"a number":                      map[*any]int{&num: 1},
		"a float":                       map[*any]int{&float: 1},
		"a struct":                      map[*any]int{&object: 1},
		"a value of MarshalText":        map[*any]int{&text: 1},
		"two pointers":                  map[**any]int{&strp: 1},
		"in a key of an interface":      map[any]int{&text: 1},
		"through many interfaces":       map[*any]int{deep: 1},
		"through many, in an interface": map[any]int{deep: 1},
	} {
		t.Run(name, func(t *testing.T) {
			want, wantErr := stdjson.Marshal(v)
			got, err := json.Marshal(v)
			if string(got) != string(want) || (err == nil) != (wantErr == nil) {
				t.Errorf("got %s %v, want %s %v", got, err, want, wantErr)
			}
		})
	}
	// colored as a string, as a key of MarshalText is.
	t.Run("colored", func(t *testing.T) {
		assertColoredKey(t, map[*any]int{&str: 1}, "s")
		assertColoredKey(t, map[**any]int{&strp: 1}, "s")
		assertColoredKey(t, map[*any]int{&text: 1}, "pq")
		assertColoredKey(t, map[*any]int{&num: 1}, "5")
	})
}

// The pointers and the interfaces of a key are followed one by one, not by a call each: a key of 200000 of them is
// named, where the stack would overflow.
func TestEncodeMapKeyOfManyPointers(t *testing.T) {
	deep := new(any)
	last := deep
	for range 200000 {
		next := new(any)
		*last = next
		last = next
	}
	*last = "end"
	for name, v := range map[string]any{
		"a key of pointers":        map[*any]int{deep: 1},
		"in a key of an interface": map[any]int{deep: 1},
	} {
		t.Run(name, func(t *testing.T) {
			if got, err := json.Marshal(v); err != nil || string(got) != `{"end":1}` {
				t.Errorf("got %s %v, want {\"end\":1}", got, err)
			}
		})
	}
}

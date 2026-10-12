//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// A key of pointers to an interface{} is named by the dynamic value at the end of its pointers and interfaces, where
// encoding/json is built on encoding/json/v2 and encodes the keys of interface types ( see
// TestEncodeMapInterfaceKeys ): a number is a string, where it was written as a number, and the names of several
// panicked when they were sorted, a float or a struct is not a name, where they were written as values, which is not
// JSON, and a key of an interface which holds a pointer to an interface is named, where it failed.
func TestEncodeMapKeyPointersToInterfaces(t *testing.T) {
	str, num, num2, text := any("s"), any(5), any(6), any(&pointerTextKey{"q"})
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
		"numbers, sorted by names":      map[*any]int{&num: 1, &num2: 2},
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

type keyByte byte

type textKeyByte byte

func (*textKeyByte) MarshalText() ([]byte, error) { return []byte("b"), nil }

type textKeyBytes []byte

func (b textKeyBytes) MarshalText() ([]byte, error) { return append([]byte("t"), b...), nil }

// A key which leads to a slice of bytes is named by its base64, as encoding/json writes the slice as a value, which
// is a string: a nil slice, which is null, and an array of bytes or a slice of bytes which have a method of
// marshaling, which are arrays, are not names, and a slice of bytes which has MarshalText is its text.
func TestEncodeMapKeyPointersToBytes(t *testing.T) {
	for name, x := range map[string]any{
		"bytes":                  []byte("ab"),
		"no bytes":               []byte{},
		"nil bytes":              []byte(nil),
		"named bytes":            stdjson.RawMessage(`"r"`),
		"bytes of a named byte":  []keyByte{1, 2},
		"bytes with MarshalText": []textKeyByte{1},
		"a slice of MarshalText": textKeyBytes("ab"),
		"an array of bytes":      [2]byte{1, 2},
	} {
		for by, v := range map[string]any{
			"a pointer":    map[*any]int{&x: 1},
			"an interface": map[any]int{&x: 1},
		} {
			t.Run(name+" by "+by, func(t *testing.T) {
				want, wantErr := stdjson.Marshal(v)
				got, err := json.Marshal(v)
				if string(got) != string(want) || (err == nil) != (wantErr == nil) {
					t.Errorf("got %s %v, want %s %v", got, err, want, wantErr)
				}
			})
		}
	}
	t.Run("colored", func(t *testing.T) {
		x := any([]byte("ab"))
		assertColoredKey(t, map[*any]int{&x: 1}, "YWI=")
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

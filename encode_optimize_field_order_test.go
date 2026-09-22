package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"reflect"
	"testing"

	"github.com/goccy/go-json"
)

// OptimizeFieldOrder lets the encoder put the fields of the same kind together and a recursive field last: the
// keys must be the same as without the option, only in another order; the values must decode back to what was
// encoded; and the opcodes of a type with and without the option must not be mixed up.

type orderedFields struct {
	Next *orderedFields `json:"next,omitempty"`
	A    int            `json:"a"`
	S1   string         `json:"s1"`
	B    int            `json:"b"`
	O    int            `json:"o,omitempty"`
	F    float64        `json:"f"`
	S2   string         `json:"s2"`
	T    bool           `json:"t"`
	C    int            `json:"c"`
	Sub  struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"sub"`
	F2 float64 `json:"f2"`
	P  *int    `json:"p"`
	U  uint8   `json:"u"`
}

func newOrderedFields(n int) *orderedFields {
	var head *orderedFields
	for i := n; i > 0; i-- {
		head = &orderedFields{Next: head, A: i, S1: "s1", B: 2, O: i % 2, F: 1.5, S2: "s2", T: true, C: 3, F2: 2.5, U: 4}
		head.Sub.X = 1
	}
	return head
}

func keysOf(t *testing.T, data []byte) []string {
	t.Helper()
	dec := stdjson.NewDecoder(bytes.NewReader(data))
	if tok, err := dec.Token(); err != nil || tok != stdjson.Delim('{') {
		t.Fatalf("not an object: %s", data)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, tok.(string))
		var v stdjson.RawMessage
		if err := dec.Decode(&v); err != nil {
			t.Fatal(err)
		}
	}
	return keys
}

func TestEncodeOptimizeFieldOrder(t *testing.T) {
	v := newOrderedFields(3)
	plain, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := stdjson.Marshal(v)
	if string(plain) != string(expected) {
		t.Fatalf("without the option:\n got %s\nwant %s", plain, expected)
	}
	optimized, err := json.MarshalWithOption(v, json.OptimizeFieldOrder())
	if err != nil {
		t.Fatal(err)
	}
	// the ints together where the first int is, the strings where the first string is, the floats too, and
	// the recursive field last: the others in their order.
	wantKeys := []string{"a", "b", "c", "s1", "s2", "o", "f", "f2", "t", "sub", "p", "u", "next"}
	if keys := keysOf(t, optimized); !reflect.DeepEqual(keys, wantKeys) {
		t.Fatalf("keys with the option: %v, want %v\n%s", keys, wantKeys, optimized)
	}
	var decoded orderedFields
	if err := stdjson.Unmarshal(optimized, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&decoded, v) {
		t.Fatalf("the values with the option decode to %+v, want %+v", decoded, *v)
	}
	// the option doesn't change the opcodes without it, and the other way round: the two are cached apart.
	for i := 0; i < 3; i++ {
		got, _ := json.Marshal(v)
		if string(got) != string(plain) {
			t.Fatalf("without the option after it:\n got %s\nwant %s", got, plain)
		}
		got, _ = json.MarshalWithOption(v, json.OptimizeFieldOrder())
		if string(got) != string(optimized) {
			t.Fatalf("with the option after without:\n got %s\nwant %s", got, optimized)
		}
	}
	// with an indent, and in an interface value, a slice and a map.
	indented, err := json.MarshalIndentWithOption(v, "", " ", json.OptimizeFieldOrder())
	if err != nil {
		t.Fatal(err)
	}
	if keys := keysOf(t, indented); !reflect.DeepEqual(keys, wantKeys) {
		t.Fatalf("keys with the option and an indent: %v", keys)
	}
	nested, err := json.MarshalWithOption(map[string]interface{}{"k": []interface{}{v}}, json.OptimizeFieldOrder())
	if err != nil {
		t.Fatal(err)
	}
	var decodedNested map[string][]*orderedFields
	if err := stdjson.Unmarshal(nested, &decodedNested); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decodedNested["k"][0], v) {
		t.Fatalf("the values in an interface value with the option decode to %+v", decodedNested["k"][0])
	}
	// a long list: encoded in one frame with the recursive field last.
	long := newOrderedFields(1200)
	optimized, err = json.MarshalWithOption(long, json.OptimizeFieldOrder())
	if err != nil {
		t.Fatal(err)
	}
	var decodedLong orderedFields
	if err := stdjson.Unmarshal(optimized, &decodedLong); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&decodedLong, long) {
		t.Fatal("the long list with the option doesn't decode to what was encoded")
	}
	// two fields of the struct's own type: neither is moved.
	type twoRecursive struct {
		Left  *twoRecursive `json:"left"`
		A     int           `json:"a"`
		Right *twoRecursive `json:"right"`
	}
	two, err := json.MarshalWithOption(&twoRecursive{A: 1, Right: &twoRecursive{A: 2}}, json.OptimizeFieldOrder())
	if err != nil {
		t.Fatal(err)
	}
	if keys := keysOf(t, two); !reflect.DeepEqual(keys, []string{"left", "a", "right"}) {
		t.Fatalf("keys with two recursive fields: %v", keys)
	}
}

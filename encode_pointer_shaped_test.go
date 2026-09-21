package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

// A value whose type consists of a single pointer ( a struct of one pointer field, an array of one pointer,
// a map, ... ) is stored in an interface value as the pointer itself, not as a pointer to the value.
// The encoder must produce the same result as encoding/json however such a value is passed.

type pointerShapedItem struct {
	A string `json:"a"`
}

// the pointee starts with a nested struct value.
type pointerShapedDetail struct {
	I pointerShapedItem `json:"i"`
}

type pointerShapedBody struct {
	Payload *pointerShapedDetail `json:"p"`
}

// the pointee starts with a pointer field and has more fields.
type pointerShapedPtrFields struct {
	Foo *string `json:"foo"`
	Bar *string `json:"bar"`
}

type pointerShapedPtrFieldsHolder struct {
	B *pointerShapedPtrFields `json:"b"`
}

type pointerShapedMixedFields struct {
	ID   *string `json:"id"`
	Name string  `json:"name"`
}

type pointerShapedMixedFieldsHolder struct {
	Bot *pointerShapedMixedFields `json:"bot"`
}

// the only field is an embedded pointer to a struct which embeds another struct.
type pointerShapedDBObject struct {
	ID int64 `json:"id,omitempty"`
}

type pointerShapedEmbeddedItem struct {
	pointerShapedDBObject
}

type pointerShapedEmbeddedHolder struct {
	*pointerShapedEmbeddedItem
}

// omitempty on the pointer: whether the pointee is zero must not matter.
type pointerShapedChild struct {
	Value int `json:"v"`
}

type pointerShapedOmitEmptyHolder struct {
	Child *pointerShapedChild `json:"c,omitempty"`
}

type pointerShapedStringChild struct {
	Value string `json:"v"`
}

type pointerShapedOmitEmptyStringHolder struct {
	Child *pointerShapedStringChild `json:"c,omitempty"`
}

// recursive type passed by value.
type pointerShapedRecursive struct {
	Next *pointerShapedRecursive `json:"next"`
}

type pointerShapedMapHolder struct {
	M map[string]int `json:"m"`
}

type pointerShapedNested struct {
	Holder pointerShapedBody `json:"holder"`
}

// a map type which has MarshalJSON with a value receiver.
type pointerShapedSet map[string]struct{}

func (s pointerShapedSet) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"len=%d"`, len(s))), nil
}

type pointerShapedSetHolder struct {
	S pointerShapedSet `json:"s"`
}

// MarshalJSON with a pointer receiver.
type pointerShapedPtrMarshaler struct {
	A int
}

func (m *pointerShapedPtrMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`"ptr receiver"`), nil
}

func TestEncodePointerShapedValue(t *testing.T) {
	str := "hello"
	num := 1
	for _, test := range []struct {
		name string
		v    interface{}
		// skipIndent is set for a recursive type: the indent of a recursive type held by an interface value
		// is wrong whether the value is pointer-shaped or not, which is not what this test is about.
		skipIndent bool
	}{
		{name: "struct of pointer to struct starting with struct", v: pointerShapedBody{Payload: &pointerShapedDetail{I: pointerShapedItem{A: "a"}}}},
		{name: "struct of nil pointer", v: pointerShapedBody{}},
		{name: "struct of pointer to struct starting with pointer", v: pointerShapedPtrFieldsHolder{B: &pointerShapedPtrFields{Foo: &str}}},
		{name: "struct of pointer to struct starting with nil pointer", v: pointerShapedPtrFieldsHolder{B: &pointerShapedPtrFields{Bar: &str}}},
		{name: "struct of pointer to struct of pointer and string", v: pointerShapedMixedFieldsHolder{Bot: &pointerShapedMixedFields{ID: &str, Name: "name"}}},
		{name: "struct of embedded pointer", v: pointerShapedEmbeddedHolder{&pointerShapedEmbeddedItem{pointerShapedDBObject{ID: 1}}}},
		{name: "struct of nil embedded pointer", v: pointerShapedEmbeddedHolder{}},
		{name: "omitempty pointer to zero int struct", v: pointerShapedOmitEmptyHolder{Child: &pointerShapedChild{}}},
		{name: "omitempty pointer to non-zero int struct", v: pointerShapedOmitEmptyHolder{Child: &pointerShapedChild{Value: 1}}},
		{name: "omitempty nil pointer", v: pointerShapedOmitEmptyHolder{}},
		{name: "omitempty pointer to empty string struct", v: pointerShapedOmitEmptyStringHolder{Child: &pointerShapedStringChild{}}},
		{name: "recursive struct", v: pointerShapedRecursive{Next: &pointerShapedRecursive{}}, skipIndent: true},
		{name: "deeper recursive struct", v: pointerShapedRecursive{Next: &pointerShapedRecursive{Next: &pointerShapedRecursive{}}}, skipIndent: true},
		{name: "struct of map", v: pointerShapedMapHolder{M: map[string]int{"a": 1}}},
		{name: "struct of nil map", v: pointerShapedMapHolder{}},
		{name: "struct of pointer-shaped struct", v: pointerShapedNested{Holder: pointerShapedBody{Payload: &pointerShapedDetail{I: pointerShapedItem{A: "a"}}}}},
		{name: "array of pointer", v: [1]*int{&num}},
		{name: "array of nil pointer", v: [1]*int{nil}},
		{name: "array of pointer-shaped struct", v: [1]pointerShapedBody{{Payload: &pointerShapedDetail{}}}},
		{name: "map of marshaler map", v: map[string]pointerShapedSet{"foo": {"a": struct{}{}}}},
		{name: "map of empty marshaler map", v: map[string]pointerShapedSet{"foo": {}}},
		{name: "int key map of marshaler map", v: map[int]pointerShapedSet{1: {"a": struct{}{}}}},
		{name: "map of map of marshaler map", v: map[string]map[string]pointerShapedSet{"x": {"foo": {"a": struct{}{}}}}},
		{name: "marshaler map", v: pointerShapedSet{"a": struct{}{}}},
		{name: "struct of marshaler map", v: pointerShapedSetHolder{S: pointerShapedSet{"a": struct{}{}}}},
		{name: "value of pointer receiver marshaler", v: pointerShapedPtrMarshaler{A: 1}},
		{name: "pointer to pointer receiver marshaler", v: &pointerShapedPtrMarshaler{A: 1}},
		{name: "pointer to pointer to pointer receiver marshaler", v: func() **pointerShapedPtrMarshaler {
			m := &pointerShapedPtrMarshaler{A: 1}
			return &m
		}()},
		{name: "nil pointer to pointer receiver marshaler", v: (*pointerShapedPtrMarshaler)(nil)},
	} {
		t.Run(test.name, func(t *testing.T) {
			// the same value must be encoded in the same way wherever it is placed.
			for _, placement := range []struct {
				name string
				v    interface{}
			}{
				{"value", test.v},
				{"pointer to interface", &test.v},
				{"slice of interface", []interface{}{test.v}},
				{"map of interface", map[string]interface{}{"k": test.v}},
			} {
				expected, err := stdjson.Marshal(placement.v)
				if err != nil {
					t.Fatal(err)
				}
				// the second call uses the cached opcode.
				for i := 0; i < 2; i++ {
					got, err := json.Marshal(placement.v)
					if err != nil {
						t.Fatalf("%s ( call %d ): %v", placement.name, i, err)
					}
					if string(got) != string(expected) {
						t.Fatalf("%s ( call %d ): expected %s but got %s", placement.name, i, expected, got)
					}
				}
				if test.skipIndent {
					continue
				}
				gotIndent, err := json.MarshalIndent(placement.v, "", "  ")
				if err != nil {
					t.Fatalf("%s ( indent ): %v", placement.name, err)
				}
				expectedIndent, err := stdjson.MarshalIndent(placement.v, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				if string(gotIndent) != string(expectedIndent) {
					t.Fatalf("%s ( indent ): expected %s but got %s", placement.name, expectedIndent, gotIndent)
				}
			}
		})
	}
}

// addressRecorder records the receiver with which MarshalJSON is called.
type addressRecorder struct {
	A, B int
}

var recordedAddresses []*addressRecorder

func (r *addressRecorder) MarshalJSON() ([]byte, error) {
	recordedAddresses = append(recordedAddresses, r)
	return []byte("1"), nil
}

type addressRecorderHolder struct {
	R addressRecorder
}

// A marshaler with a pointer receiver is called with the address of the original value, not of a copy,
// as encoding/json does.
func TestEncodePointerReceiverMarshalerAddress(t *testing.T) {
	value := &addressRecorder{}
	elems := []addressRecorder{{}, {}}
	holder := &addressRecorderHolder{}
	for _, test := range []struct {
		name     string
		v        interface{}
		expected []*addressRecorder
	}{
		{"pointer", value, []*addressRecorder{value}},
		{"elements of slice", elems, []*addressRecorder{&elems[0], &elems[1]}},
		{"field of pointer to struct", holder, []*addressRecorder{&holder.R}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, marshal := range []func(interface{}) ([]byte, error){
				json.Marshal,
				func(v interface{}) ([]byte, error) { return json.MarshalIndent(v, "", "  ") },
			} {
				recordedAddresses = nil
				if _, err := marshal(test.v); err != nil {
					t.Fatal(err)
				}
				if len(recordedAddresses) != len(test.expected) {
					t.Fatalf("expected %d calls but got %d", len(test.expected), len(recordedAddresses))
				}
				for i, addr := range recordedAddresses {
					if addr != test.expected[i] {
						t.Fatalf("call %d: the marshaler was called with a copy %p, not with the original value %p", i, addr, test.expected[i])
					}
				}
			}
		})
	}
}

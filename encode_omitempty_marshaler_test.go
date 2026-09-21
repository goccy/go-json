package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

// omitempty decides that a field is empty by the kind of its value before the marshaler is called,
// also for a marshaler with a pointer receiver, which is called with the address of the field.

type omitEmptyPtrReceiverString string

func (s *omitEmptyPtrReceiverString) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", "json:"+string(*s))), nil
}

type omitEmptyPtrReceiverInt int

func (i *omitEmptyPtrReceiverInt) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", int(*i)+100)), nil
}

type omitEmptyPtrReceiverStruct struct {
	V int
}

func (s *omitEmptyPtrReceiverStruct) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"v":%d}`, s.V)), nil
}

type omitEmptyPtrReceiverFields struct {
	S  omitEmptyPtrReceiverString `json:"s,omitempty"`
	I  omitEmptyPtrReceiverInt    `json:"i,omitempty"`
	St omitEmptyPtrReceiverStruct `json:"st,omitempty"`
	// without omitempty the marshaler is always called.
	S2 omitEmptyPtrReceiverString `json:"s2"`
}

type omitEmptyPtrReceiverSingleField struct {
	S omitEmptyPtrReceiverString `json:"s,omitempty"`
}

func TestEncodeOmitEmptyWithPointerReceiverMarshaler(t *testing.T) {
	for _, test := range []struct {
		name string
		v    interface{}
	}{
		{"pointer to zero fields", &omitEmptyPtrReceiverFields{}},
		{"pointer to fields", &omitEmptyPtrReceiverFields{S: "a", I: 1, St: omitEmptyPtrReceiverStruct{V: 1}, S2: "c"}},
		{"zero fields", omitEmptyPtrReceiverFields{}},
		{"fields", omitEmptyPtrReceiverFields{S: "a", I: 1, St: omitEmptyPtrReceiverStruct{V: 1}, S2: "c"}},
		{"pointer to zero single field", &omitEmptyPtrReceiverSingleField{}},
		{"pointer to single field", &omitEmptyPtrReceiverSingleField{S: "a"}},
		{"zero single field", omitEmptyPtrReceiverSingleField{}},
		{"slice of zero and non-zero", []omitEmptyPtrReceiverSingleField{{}, {S: "a"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			expected, err := stdjson.Marshal(test.v)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				got, err := json.Marshal(test.v)
				if err != nil {
					t.Fatalf("call %d: %v", i, err)
				}
				if string(got) != string(expected) {
					t.Fatalf("call %d: expected %s but got %s", i, expected, got)
				}
			}
		})
	}
}

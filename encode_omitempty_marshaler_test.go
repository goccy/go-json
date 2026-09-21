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

// omitempty for the types which have MarshalText, with a value receiver and with a pointer receiver.

type omitEmptyTextString string

func (s omitEmptyTextString) MarshalText() ([]byte, error) {
	return []byte("text:" + string(s)), nil
}

type omitEmptyTextInt int

func (i omitEmptyTextInt) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("int:%d", int(i))), nil
}

type omitEmptyTextStruct struct {
	V int
}

func (s omitEmptyTextStruct) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("struct:%d", s.V)), nil
}

type omitEmptyTextSlice []string

func (s omitEmptyTextSlice) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("slice:%d", len(s))), nil
}

type omitEmptyTextMap map[string]int

func (m omitEmptyTextMap) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("map:%d", len(m))), nil
}

type omitEmptyPtrReceiverTextString string

func (s *omitEmptyPtrReceiverTextString) MarshalText() ([]byte, error) {
	return []byte("ptrtext:" + string(*s)), nil
}

type omitEmptyTextFields struct {
	S  omitEmptyTextString            `json:"s,omitempty"`
	I  omitEmptyTextInt               `json:"i,omitempty"`
	St omitEmptyTextStruct            `json:"st,omitempty"`
	Sl omitEmptyTextSlice             `json:"sl,omitempty"`
	M  omitEmptyTextMap               `json:"m,omitempty"`
	P  *omitEmptyTextString           `json:"p,omitempty"`
	PR omitEmptyPtrReceiverTextString `json:"pr,omitempty"`
	// without omitempty the marshaler is always called.
	S2 omitEmptyTextString `json:"s2"`
}

// the field with omitempty is the first and the only field, which has opcodes of its own.
type omitEmptyTextSingleField struct {
	S omitEmptyTextString `json:"s,omitempty"`
}

type omitEmptyTextLastField struct {
	A int                 `json:"a"`
	S omitEmptyTextString `json:"s,omitempty"`
}

func TestEncodeOmitEmptyWithTextMarshaler(t *testing.T) {
	str := omitEmptyTextString("p")
	emptyStr := omitEmptyTextString("")
	for _, test := range []struct {
		name string
		v    interface{}
	}{
		{"zero fields", omitEmptyTextFields{}},
		{"pointer to zero fields", &omitEmptyTextFields{}},
		{"fields", omitEmptyTextFields{
			S: "a", I: 1, St: omitEmptyTextStruct{V: 1}, Sl: omitEmptyTextSlice{"x"}, M: omitEmptyTextMap{"k": 1}, P: &str, PR: "b", S2: "c",
		}},
		{"pointer to fields", &omitEmptyTextFields{
			S: "a", I: 1, St: omitEmptyTextStruct{V: 1}, Sl: omitEmptyTextSlice{"x"}, M: omitEmptyTextMap{"k": 1}, P: &str, PR: "b", S2: "c",
		}},
		{"empty but not nil", &omitEmptyTextFields{Sl: omitEmptyTextSlice{}, M: omitEmptyTextMap{}, P: &emptyStr}},
		{"zero single field", omitEmptyTextSingleField{}},
		{"pointer to zero single field", &omitEmptyTextSingleField{}},
		{"single field", omitEmptyTextSingleField{S: "a"}},
		{"pointer to single field", &omitEmptyTextSingleField{S: "a"}},
		{"zero last field", omitEmptyTextLastField{A: 1}},
		{"last field", &omitEmptyTextLastField{A: 1, S: "a"}},
		{"slice of zero and non-zero", []omitEmptyTextSingleField{{}, {S: "a"}, {}}},
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
			expectedIndent, err := stdjson.MarshalIndent(test.v, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			gotIndent, err := json.MarshalIndent(test.v, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if string(gotIndent) != string(expectedIndent) {
				t.Fatalf("indent: expected %s but got %s", expectedIndent, gotIndent)
			}
		})
	}
}

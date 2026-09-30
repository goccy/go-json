package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"math"
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
		v    any
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
		v    any
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

// omitempty decides that a field with a marshaler is empty by the kind of the field, as encoding/json does: a nil
// func or chan is not empty, -0 is, and the integer checked is the one of the size of the field, also for a field
// encoded by the generic field opcode ( a long key, or omitzero ).

type omitEmptyKindFunc func()

func (f omitEmptyKindFunc) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte(`"nil"`), nil
	}
	return []byte(`"func"`), nil
}

type omitEmptyKindPtrReceiverFunc func()

func (f *omitEmptyKindPtrReceiverFunc) MarshalJSON() ([]byte, error) {
	if *f == nil {
		return []byte(`"nil"`), nil
	}
	return []byte(`"func"`), nil
}

type omitEmptyKindTextFunc func()

func (f omitEmptyKindTextFunc) MarshalText() ([]byte, error) {
	if f == nil {
		return []byte("nil"), nil
	}
	return []byte("func"), nil
}

type omitEmptyKindChan chan int

func (c omitEmptyKindChan) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte(`"nil"`), nil
	}
	return []byte(`"chan"`), nil
}

type omitEmptyKindFloat float64

func (f omitEmptyKindFloat) MarshalJSON() ([]byte, error) {
	return []byte(`"float"`), nil
}

type omitEmptyKindInt8 int8

func (i omitEmptyKindInt8) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"int8:%d"`, int8(i))), nil
}

type omitEmptyKindPtrReceiverUint16 uint16

func (u *omitEmptyKindPtrReceiverUint16) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"uint16:%d"`, uint16(*u))), nil
}

type omitEmptyKindFields struct {
	F  omitEmptyKindFunc              `json:",omitempty"`
	PF omitEmptyKindPtrReceiverFunc   `json:",omitempty"`
	TF omitEmptyKindTextFunc          `json:",omitempty"`
	C  omitEmptyKindChan              `json:",omitempty"`
	Fl omitEmptyKindFloat             `json:",omitempty"`
	I8 omitEmptyKindInt8              `json:",omitempty"`
	U  omitEmptyKindPtrReceiverUint16 `json:",omitempty"`
	B  int8
	// the field of a long key and the one of omitzero are encoded by the generic field opcode.
	ThisIsTheKeyOfAFieldLongerThanAChunkOfTheEncoder omitEmptyKindInt8 `json:",omitempty"`
	Z                                                omitEmptyKindInt8 `json:",omitempty,omitzero"`
	B2                                               int8
}

func TestEncodeOmitEmptyByTheKindOfTheMarshalerField(t *testing.T) {
	negativeZero := omitEmptyKindFloat(math.Copysign(0, -1))
	for _, test := range []struct {
		name string
		v    any
	}{
		{"zero fields", &omitEmptyKindFields{B: -1, B2: -1}},
		{"non-zero fields", &omitEmptyKindFields{
			F: func() {}, PF: func() {}, TF: func() {}, C: make(omitEmptyKindChan), Fl: 1, I8: 1, U: 0x100, B: -1,
			ThisIsTheKeyOfAFieldLongerThanAChunkOfTheEncoder: 1, Z: 1, B2: -1,
		}},
		{"negative zero", &omitEmptyKindFields{Fl: negativeZero}},
		{"slice", []omitEmptyKindFields{{B: -1}, {I8: -1, U: 1}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			expected, err := stdjson.Marshal(test.v)
			if err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(test.v)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(expected) {
				t.Fatalf("expected %s but got %s", expected, got)
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

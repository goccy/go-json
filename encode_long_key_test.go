package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

// A key up to a chunk ( encoder.KeyChunkSize ) is written by the opcode of the field, a longer one by the
// generic field opcode with the value by its own opcode: both must encode as encoding/json does, for every kind
// of a value, with and without omitempty and the string option.

type longKeyMarshaler struct{ V int }

func (m longKeyMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"v":%d}`, m.V)), nil
}

type longKeyPtrMarshaler struct{ V int }

func (m *longKeyPtrMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"p":%d}`, m.V)), nil
}

type longKeyTextMarshaler struct{ V int }

func (m longKeyTextMarshaler) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("t%d", m.V)), nil
}

type longKeyNode struct {
	Next *longKeyNode `json:"a_key_of_a_recursive_field_which_is_longer_than_a_chunk,omitempty"`
	V    int          `json:"v"`
}

type longKeyEmbedded struct {
	EmbeddedFieldWithAKeyLongerThanAChunkOfTheVM int `json:"embedded_field_with_a_key_longer_than_a_chunk,omitempty"`
}

type longKeys struct {
	Int       int                  `json:"an_int_field_with_a_key_longer_than_a_chunk"`
	Int8      int8                 `json:"an_int8_field_with_a_key_longer_than_a_chunk"`
	Int16     int16                `json:"an_int16_field_with_a_key_longer_than_a_chunk"`
	Int32     int32                `json:"an_int32_field_with_a_key_longer_than_a_chunk"`
	Int64     int64                `json:"an_int64_field_with_a_key_longer_than_a_chunk"`
	Uint      uint                 `json:"a_uint_field_with_a_key_longer_than_a_chunk"`
	Uint8     uint8                `json:"a_uint8_field_with_a_key_longer_than_a_chunk"`
	Float32   float32              `json:"a_float32_field_with_a_key_longer_than_a_chunk"`
	Float64   float64              `json:"a_float64_field_with_a_key_longer_than_a_chunk"`
	Bool      bool                 `json:"a_bool_field_with_a_key_longer_than_a_chunk"`
	String    string               `json:"a_string_field_with_a_key_longer_than_a_chunk"`
	Bytes     []byte               `json:"a_bytes_field_with_a_key_longer_than_a_chunk"`
	Slice     []int                `json:"a_slice_field_with_a_key_longer_than_a_chunk"`
	Array     [2]int               `json:"an_array_field_with_a_key_longer_than_a_chunk"`
	Array0    [0]int               `json:"an_empty_array_field_with_a_key_longer_than_a_chunk"`
	Map       map[string]int       `json:"a_map_field_with_a_key_longer_than_a_chunk"`
	Interface interface{}          `json:"an_interface_field_with_a_key_longer_than_a_chunk"`
	Struct    longKeyEmbedded      `json:"a_struct_field_with_a_key_longer_than_a_chunk"`
	Ptr       *int                 `json:"a_pointer_field_with_a_key_longer_than_a_chunk"`
	PtrStruct *longKeyEmbedded     `json:"a_struct_pointer_field_with_a_key_longer_than_a_chunk"`
	Number    json.Number          `json:"a_number_field_with_a_key_longer_than_a_chunk"`
	Marshaler longKeyMarshaler     `json:"a_marshaler_field_with_a_key_longer_than_a_chunk"`
	PtrMarsh  longKeyPtrMarshaler  `json:"a_pointer_marshaler_field_with_a_key_longer_than_a_chunk"`
	PtrMarshP *longKeyPtrMarshaler `json:"a_pointer_marshaler_pointer_field_with_a_key_longer_than_a_chunk"`
	Text      longKeyTextMarshaler `json:"a_text_marshaler_field_with_a_key_longer_than_a_chunk"`
	Node      longKeyNode          `json:"a_recursive_field_with_a_key_longer_than_a_chunk"`
	Raw       json.RawMessage      `json:"a_raw_message_field_with_a_key_longer_than_a_chunk"`
	longKeyEmbedded
}

type longKeysOmitEmpty struct {
	Int       int                  `json:"an_int_field_with_a_key_longer_than_a_chunk,omitempty"`
	Int8      int8                 `json:"an_int8_field_with_a_key_longer_than_a_chunk,omitempty"`
	Int16     int16                `json:"an_int16_field_with_a_key_longer_than_a_chunk,omitempty"`
	Int32     int32                `json:"an_int32_field_with_a_key_longer_than_a_chunk,omitempty"`
	Int64     int64                `json:"an_int64_field_with_a_key_longer_than_a_chunk,omitempty"`
	Uint      uint                 `json:"a_uint_field_with_a_key_longer_than_a_chunk,omitempty"`
	Uint8     uint8                `json:"a_uint8_field_with_a_key_longer_than_a_chunk,omitempty"`
	Float32   float32              `json:"a_float32_field_with_a_key_longer_than_a_chunk,omitempty"`
	Float64   float64              `json:"a_float64_field_with_a_key_longer_than_a_chunk,omitempty"`
	Bool      bool                 `json:"a_bool_field_with_a_key_longer_than_a_chunk,omitempty"`
	String    string               `json:"a_string_field_with_a_key_longer_than_a_chunk,omitempty"`
	Bytes     []byte               `json:"a_bytes_field_with_a_key_longer_than_a_chunk,omitempty"`
	Slice     []int                `json:"a_slice_field_with_a_key_longer_than_a_chunk,omitempty"`
	Array     [2]int               `json:"an_array_field_with_a_key_longer_than_a_chunk,omitempty"`
	Array0    [0]int               `json:"an_empty_array_field_with_a_key_longer_than_a_chunk,omitempty"`
	Map       map[string]int       `json:"a_map_field_with_a_key_longer_than_a_chunk,omitempty"`
	Interface interface{}          `json:"an_interface_field_with_a_key_longer_than_a_chunk,omitempty"`
	Struct    longKeyEmbedded      `json:"a_struct_field_with_a_key_longer_than_a_chunk,omitempty"`
	Ptr       *int                 `json:"a_pointer_field_with_a_key_longer_than_a_chunk,omitempty"`
	PtrStruct *longKeyEmbedded     `json:"a_struct_pointer_field_with_a_key_longer_than_a_chunk,omitempty"`
	Number    json.Number          `json:"a_number_field_with_a_key_longer_than_a_chunk,omitempty"`
	Marshaler longKeyMarshaler     `json:"a_marshaler_field_with_a_key_longer_than_a_chunk,omitempty"`
	PtrMarsh  longKeyPtrMarshaler  `json:"a_pointer_marshaler_field_with_a_key_longer_than_a_chunk,omitempty"`
	PtrMarshP *longKeyPtrMarshaler `json:"a_pointer_marshaler_pointer_field_with_a_key_longer_than_a_chunk,omitempty"`
	Text      longKeyTextMarshaler `json:"a_text_marshaler_field_with_a_key_longer_than_a_chunk,omitempty"`
	Node      longKeyNode          `json:"a_recursive_field_with_a_key_longer_than_a_chunk,omitempty"`
	Raw       json.RawMessage      `json:"a_raw_message_field_with_a_key_longer_than_a_chunk,omitempty"`
	longKeyEmbedded
}

type longKeysString struct {
	Int     int      `json:"an_int_field_with_a_key_longer_than_a_chunk,string"`
	Uint    uint16   `json:"a_uint_field_with_a_key_longer_than_a_chunk,string"`
	Float32 float32  `json:"a_float32_field_with_a_key_longer_than_a_chunk,string"`
	Float64 float64  `json:"a_float64_field_with_a_key_longer_than_a_chunk,string"`
	Bool    bool     `json:"a_bool_field_with_a_key_longer_than_a_chunk,string"`
	String  string   `json:"a_string_field_with_a_key_longer_than_a_chunk,string"`
	Ptr     *int     `json:"a_pointer_field_with_a_key_longer_than_a_chunk,string"`
	PtrOmit *float64 `json:"a_pointer_field_with_a_key_longer_than_a_chunk_omitted,omitempty,string"`
	IntOmit int      `json:"an_int_field_with_a_key_longer_than_a_chunk_omitted,omitempty,string"`
}

// the keys around the size of a chunk, with the quotes and the colon: 31, 32 and 33 bytes.
type keysAroundChunk struct {
	A int `json:"a_key_of_28_characters_here_"`
	B int `json:"a_key_of_29_characters_here__"`
	C int `json:"a_key_of_30_characters_here___"`
	D int `json:"a_key_of_28_characters_omit_,omitempty"`
	E int `json:"a_key_of_29_characters_omit__,omitempty"`
	F int `json:"a_key_of_30_characters_omit___,omitempty"`
}

func TestEncodeLongKey(t *testing.T) {
	one := 1
	f := 1.5
	values := []interface{}{
		longKeys{},
		&longKeys{},
		longKeys{
			Int: 1, Int8: -2, Int16: 3, Int32: -4, Int64: 5, Uint: 6, Uint8: 7, Float32: 1.5, Float64: -2.5,
			Bool: true, String: "s", Bytes: []byte("b"), Slice: []int{1}, Array: [2]int{1, 2},
			Map: map[string]int{"k": 1}, Interface: "i", Struct: longKeyEmbedded{1}, Ptr: &one,
			PtrStruct: &longKeyEmbedded{2}, Number: "3", Marshaler: longKeyMarshaler{4},
			PtrMarsh: longKeyPtrMarshaler{5}, PtrMarshP: &longKeyPtrMarshaler{6}, Text: longKeyTextMarshaler{7},
			Node: longKeyNode{Next: &longKeyNode{V: 8}, V: 9}, Raw: json.RawMessage(`{"r":1}`),
			longKeyEmbedded: longKeyEmbedded{10},
		},
		longKeysOmitEmpty{},
		&longKeysOmitEmpty{},
		longKeysOmitEmpty{
			Int: 1, Int8: -2, Int16: 3, Int32: -4, Int64: 5, Uint: 6, Uint8: 7, Float32: 1.5, Float64: -2.5,
			Bool: true, String: "s", Bytes: []byte("b"), Slice: []int{1}, Array: [2]int{1, 2},
			Map: map[string]int{"k": 1}, Interface: "i", Struct: longKeyEmbedded{1}, Ptr: &one,
			PtrStruct: &longKeyEmbedded{2}, Number: "3", Marshaler: longKeyMarshaler{4},
			PtrMarsh: longKeyPtrMarshaler{5}, PtrMarshP: &longKeyPtrMarshaler{6}, Text: longKeyTextMarshaler{7},
			Node: longKeyNode{Next: &longKeyNode{V: 8}, V: 9}, Raw: json.RawMessage(`{"r":1}`),
			longKeyEmbedded: longKeyEmbedded{10},
		},
		// empty but not nil.
		longKeysOmitEmpty{Bytes: []byte{}, Slice: []int{}, Map: map[string]int{}, Interface: 0, Raw: json.RawMessage{}},
		longKeysString{},
		&longKeysString{},
		longKeysString{Int: -1, Uint: 2, Float32: 1.5, Float64: -2.5, Bool: true, String: "s", Ptr: &one, PtrOmit: &f, IntOmit: 3},
		keysAroundChunk{},
		keysAroundChunk{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6},
		[]longKeysOmitEmpty{{}, {Int: 1}},
		map[string]longKeysString{"k": {Int: 1}},
	}
	for _, v := range values {
		expected, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T:\n got %s\nwant %s", v, got, expected)
		}
		// with the indent and the escape of HTML too.
		expected, _ = stdjson.MarshalIndent(v, "", "  ")
		got, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T with indent:\n got %s\nwant %s", v, got, expected)
		}
	}
}

// a key which is longer than any chunk and is escaped: what a chunk can't hold.
func TestEncodeVeryLongKey(t *testing.T) {
	type veryLongKey struct {
		A int    `json:"a_key_which_is_much_longer_than_a_chunk_of_the_VM_and_has_a_<tag>_and_a_\"quote\"_in_it_to_escape,omitempty"`
		B string `json:"another_key_which_is_much_longer_than_a_chunk_of_the_VM_with_a_value_after_it_of_any_length"`
	}
	for _, v := range []interface{}{veryLongKey{}, veryLongKey{A: 1, B: strings.Repeat("x", 100)}} {
		expected, _ := stdjson.Marshal(v)
		got, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T:\n got %s\nwant %s", v, got, expected)
		}
	}
}

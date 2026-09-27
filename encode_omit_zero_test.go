package json_test

import (
	stdjson "encoding/json"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

// omitzero omits a field whose value is zero, as encoding/json decides it from Go 1.24: by the IsZero method
// of the type if it has one, and else by reflect.Value.IsZero. With an older Go, encoding/json ignores the
// option, and so does go-json: the cases compare with encoding/json of the Go the test runs with.

type omitZeroValueMethod struct{ N int }

func (v omitZeroValueMethod) IsZero() bool { return v.N < 0 }

type omitZeroPointerMethod struct{ N int }

func (v *omitZeroPointerMethod) IsZero() bool { return v.N < 0 }

type omitZeroer interface{ IsZero() bool }

type omitZeroMarshaler struct{ N int }

func (v *omitZeroMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"m"`), nil }

func (v *omitZeroMarshaler) IsZero() bool { return v.N < 0 }

type omitZeroTextMarshaler struct{ N int }

func (v *omitZeroTextMarshaler) MarshalText() ([]byte, error) { return []byte("t"), nil }

type omitZeroMarshalers struct {
	JSON    omitZeroMarshaler      `json:"json,omitzero"`
	JSONPtr *omitZeroMarshaler     `json:"json_ptr,omitzero"`
	Text    omitZeroTextMarshaler  `json:"text,omitzero"`
	TextPtr *omitZeroTextMarshaler `json:"text_ptr,omitzero"`
}

type omitZeroInner struct {
	A int
	B string
}

type omitZeroEmbedded struct {
	E int `json:"e,omitzero"`
}

type omitZeroScalars struct {
	Bool    bool           `json:"bool,omitzero"`
	Int     int            `json:"int,omitzero"`
	Int8    int8           `json:"int8,omitzero"`
	Int16   int16          `json:"int16,omitzero"`
	Int32   int32          `json:"int32,omitzero"`
	Uint64  uint64         `json:"uint64,omitzero"`
	Float32 float32        `json:"float32,omitzero"`
	Float64 float64        `json:"float64,omitzero"`
	String  string         `json:"string,omitzero"`
	Number  json.Number    `json:"number,omitzero"`
	Ptr     *int           `json:"ptr,omitzero"`
	Slice   []int          `json:"slice,omitzero"`
	Map     map[string]int `json:"map,omitzero"`
	Any     any            `json:"any,omitzero"`
	Array   [2]int         `json:"array,omitzero"`
	Empty   [0]int         `json:"empty,omitzero"`
	Struct  omitZeroInner  `json:"struct,omitzero"`
	Bytes   []byte         `json:"bytes,omitzero"`
	Quoted  int            `json:"quoted,omitzero,string"`
}

type omitZeroMethods struct {
	Time      time.Time              `json:"time,omitzero"`
	TimePtr   *time.Time             `json:"time_ptr,omitzero"`
	Value     omitZeroValueMethod    `json:"value,omitzero"`
	ValuePtr  *omitZeroValueMethod   `json:"value_ptr,omitzero"`
	Pointer   omitZeroPointerMethod  `json:"pointer,omitzero"`
	PointerP  *omitZeroPointerMethod `json:"pointer_ptr,omitzero"`
	Interface omitZeroer             `json:"interface,omitzero"`
}

type omitZeroBoth struct {
	Slice  []int          `json:"slice,omitempty,omitzero"`
	Map    map[string]int `json:"map,omitzero,omitempty"`
	Struct omitZeroInner  `json:"struct,omitempty,omitzero"`
	Int    int            `json:"int,omitempty,omitzero"`
}

type omitZeroLayout struct {
	First                                   int `json:"first,omitzero"`
	omitZeroEmbedded                        `json:",omitzero"`
	AKeyWhichIsLongerThanAChunkOfTheEncoder string           `json:"a_key_which_is_longer_than_a_chunk_of_the_encoder,omitzero"`
	Named                                   omitZeroEmbedded `json:"named,omitzero"`
	Last                                    string           `json:"last,omitzero"`
}

type omitZeroOnly struct {
	A int `json:"a,omitzero"`
}

func TestEncodeOmitZero(t *testing.T) {
	one := 1
	zero := 0
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	var zeroTime time.Time
	values := []any{
		omitZeroScalars{},
		&omitZeroScalars{},
		omitZeroScalars{
			Bool: true, Int: 1, Int8: -1, Int16: 2, Int32: 3, Uint64: 4, Float32: 1.5, Float64: 2.5, String: "s",
			Number: "1", Ptr: &one, Slice: []int{1}, Map: map[string]int{"a": 1}, Any: 1, Array: [2]int{0, 1},
			Struct: omitZeroInner{B: "b"}, Bytes: []byte("b"), Quoted: 7,
		},
		omitZeroScalars{
			Float32: float32(math.Copysign(0, -1)), Float64: math.Copysign(0, -1), Ptr: &zero,
			Slice: []int{}, Map: map[string]int{}, Any: (*int)(nil), Bytes: []byte{},
		},
		omitZeroScalars{Any: 0, Array: [2]int{}, Struct: omitZeroInner{}},
		omitZeroScalars{Array: [2]int{1}},
		omitZeroMethods{},
		omitZeroMethods{
			Time: now, TimePtr: &zeroTime, Value: omitZeroValueMethod{-1}, ValuePtr: &omitZeroValueMethod{-1},
			Pointer: omitZeroPointerMethod{-1}, PointerP: &omitZeroPointerMethod{-1},
			Interface: omitZeroValueMethod{-1},
		},
		omitZeroMethods{
			TimePtr: &now, Value: omitZeroValueMethod{0}, ValuePtr: &omitZeroValueMethod{1},
			Pointer: omitZeroPointerMethod{0}, PointerP: &omitZeroPointerMethod{1},
			Interface: &omitZeroPointerMethod{1},
		},
		omitZeroMethods{Interface: (*omitZeroPointerMethod)(nil)},
		omitZeroMethods{Interface: (*omitZeroValueMethod)(nil)},
		omitZeroBoth{},
		omitZeroBoth{Slice: []int{}, Map: map[string]int{}},
		omitZeroBoth{Slice: []int{1}, Map: map[string]int{"a": 1}, Struct: omitZeroInner{A: 1}, Int: 1},
		omitZeroLayout{},
		omitZeroLayout{
			First: 1, omitZeroEmbedded: omitZeroEmbedded{E: 2}, AKeyWhichIsLongerThanAChunkOfTheEncoder: "long",
			Named: omitZeroEmbedded{E: 3}, Last: "last",
		},
		omitZeroMarshalers{},
		&omitZeroMarshalers{},
		&omitZeroMarshalers{JSON: omitZeroMarshaler{-1}, JSONPtr: &omitZeroMarshaler{-1}, TextPtr: &omitZeroTextMarshaler{}},
		&omitZeroMarshalers{JSON: omitZeroMarshaler{1}, JSONPtr: &omitZeroMarshaler{1}, Text: omitZeroTextMarshaler{1}},
		omitZeroOnly{},
		omitZeroOnly{A: 1},
		&omitZeroOnly{},
		[]omitZeroOnly{{}, {A: 1}},
		map[string]omitZeroOnly{"z": {}, "o": {A: 1}},
		struct {
			Inner *omitZeroOnly `json:"inner,omitzero"`
		}{Inner: &omitZeroOnly{}},
		struct {
			Inner omitZeroOnly `json:"inner,omitzero"`
			Next  int          `json:"next"`
		}{},
	}
	for _, v := range values {
		want, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(want) {
			t.Errorf("%T: got %s, want %s", v, got, want)
		}
		wantIndent, err := stdjson.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		gotIndent, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(gotIndent) != string(wantIndent) {
			t.Errorf("%T: indent: got %s, want %s", v, gotIndent, wantIndent)
		}
		// the order of the fields may change, but not which of them are written.
		ordered, err := json.MarshalWithOption(v, json.OptimizeFieldOrder())
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		var gotFields, wantFields any
		if err := stdjson.Unmarshal(ordered, &gotFields); err != nil {
			t.Fatalf("%T: %v: %s", v, err, ordered)
		}
		if err := stdjson.Unmarshal(want, &wantFields); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotFields, wantFields) {
			t.Errorf("%T: OptimizeFieldOrder: got %s, want %s", v, ordered, want)
		}
	}
}

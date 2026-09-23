package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// Fields of the same kind in a row are encoded by the opcode of the run, which falls through the opcodes of
// the fields after the first: runs of every length, of every kind which has them, broken by a field of
// another kind or by omitempty, in embedded structs and at the end of a struct, must encode as encoding/json.

type fieldRunEmbedded struct {
	E1 int
	E2 int
	E3 int
	E4 int
}

type fieldRunPtrEmbedded struct {
	P1 string
	P2 string
}

type fieldRuns struct {
	I1 int
	I2 int
	I3 int
	I4 int
	I5 int
	S1 string
	S2 string
	S3 string
	B1 bool
	B2 bool
	B3 bool
	B4 bool
	F1 float64
	F2 float64
	F3 float64
	U1 uint
	U2 uint
	U3 uint16
	U4 uint8
	fieldRunEmbedded
	I6 int
	*fieldRunPtrEmbedded
	S4 string
	O1 int `json:",omitempty"`
	O2 int `json:",omitempty"`
	I7 int
	I8 int
	X1 int8
	X2 int64
	X3 int32
	M1 float32
	M2 float32
	N1 int `json:"n1,string"`
	N2 int `json:"n2,string"`
	N3 int
	N4 int
	N5 int
	N6 int
	N7 int
}

func TestEncodeFieldRuns(t *testing.T) {
	values := []any{
		fieldRuns{},
		&fieldRuns{},
		fieldRuns{
			I1: 1, I2: -2, I3: 3, I4: -4, I5: 5, S1: "a", S2: "b\"c", S3: "日本", B1: true, B3: true,
			F1: 1.5, F2: -2.5, F3: 1e21, U1: 1, U2: 2, U3: 3, U4: 4, fieldRunEmbedded: fieldRunEmbedded{1, 2, 3, 4},
			I6: 6, fieldRunPtrEmbedded: &fieldRunPtrEmbedded{"p1", "p2"}, S4: "s4", O1: 1, I7: 7, I8: 8,
			X1: -1, X2: 2, X3: -3, M1: 1.25, M2: -1.25, N1: 1, N2: 2, N3: 3, N4: 4, N5: 5, N6: 6, N7: 7,
		},
		[]fieldRuns{{I1: 1}, {S1: "x"}},
		struct {
			A int
			B int
		}{1, 2},
		struct {
			A string
			B string
			C string
			D string
			E string
			F string
			G string
		}{"a", "b", "c", "d", "e", "f", "g"},
		struct {
			A bool
			B bool
			C float64
			D float64
			E float64
			F float64
		}{true, false, 1, 2, 3, 4},
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

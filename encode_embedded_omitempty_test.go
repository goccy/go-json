package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// A struct embedded in an embedded struct. If a field with omitempty is empty, the encoder jumps to the next
// field, so the fields must be linked through the embedded structs in whatever position they are.

type embeddedLeaf struct {
	B string `json:"b"`
}

type embeddedOmitEmptyLeaf struct {
	Error string `json:"error,omitempty"`
}

type embeddedOmitEmptySliceLeaf struct {
	Domains []string `json:"domains,omitempty"`
}

type embeddedReason struct {
	Code string `json:"code"`
}

// the embedded struct comes first, and the last field is an empty omitempty field.
type embeddedThenOmitEmptySlice struct {
	embeddedLeaf
	A []string `json:"a,omitempty"`
}

type embeddedThenOmitEmptyPtr struct {
	embeddedLeaf
	R *embeddedReason `json:"reason,omitempty"`
}

type embeddedThenOmitEmptyMap struct {
	embeddedLeaf
	M map[string]int `json:"m,omitempty"`
}

type embeddedThenOmitEmptyInterface struct {
	embeddedLeaf
	I interface{} `json:"i,omitempty"`
}

// the embedded struct whose only field is an empty omitempty field comes first.
type omitEmptyEmbeddedThenField struct {
	embeddedOmitEmptyLeaf
	Data string `json:"data"`
}

type omitEmptySliceEmbeddedThenField struct {
	embeddedOmitEmptySliceLeaf
	ID int `json:"id,omitempty"`
}

// the embedded struct whose only field is an empty omitempty field comes last.
type fieldThenOmitEmptyEmbedded struct {
	ID int `json:"id,omitempty"`
	embeddedOmitEmptySliceLeaf
}

// embedded structs on both sides of a field.
type embeddedBothSides struct {
	embeddedOmitEmptyLeaf
	Name string `json:"name,omitempty"`
	embeddedOmitEmptySliceLeaf
}

// three levels of embedding.
type embeddedMiddle struct {
	fieldThenOmitEmptyEmbedded
	Tail string `json:"tail,omitempty"`
}

func TestEncodeEmbeddedStructWithOmitEmpty(t *testing.T) {
	reason := &embeddedReason{Code: "c"}
	for _, test := range []struct {
		name string
		v    interface{}
	}{
		{"empty slice after embedded", struct{ embeddedThenOmitEmptySlice }{}},
		{"slice after embedded", struct{ embeddedThenOmitEmptySlice }{embeddedThenOmitEmptySlice{A: []string{"x"}}}},
		{"nil pointer after embedded", struct{ embeddedThenOmitEmptyPtr }{}},
		{"pointer after embedded", struct{ embeddedThenOmitEmptyPtr }{embeddedThenOmitEmptyPtr{R: reason}}},
		{"nil map after embedded", struct{ embeddedThenOmitEmptyMap }{}},
		{"map after embedded", struct{ embeddedThenOmitEmptyMap }{embeddedThenOmitEmptyMap{M: map[string]int{"a": 1}}}},
		{"nil interface after embedded", struct{ embeddedThenOmitEmptyInterface }{}},
		{"interface after embedded", struct{ embeddedThenOmitEmptyInterface }{embeddedThenOmitEmptyInterface{I: 1}}},
		{"field after empty embedded", struct{ omitEmptyEmbeddedThenField }{omitEmptyEmbeddedThenField{Data: "hi"}}},
		{"field after embedded", struct{ omitEmptyEmbeddedThenField }{omitEmptyEmbeddedThenField{embeddedOmitEmptyLeaf{"e"}, "hi"}}},
		{"field after empty slice embedded", struct {
			omitEmptySliceEmbeddedThenField
		}{omitEmptySliceEmbeddedThenField{ID: 1}}},
		{"empty embedded after field", struct{ fieldThenOmitEmptyEmbedded }{fieldThenOmitEmptyEmbedded{ID: 1}}},
		{"empty embedded after empty field", struct{ fieldThenOmitEmptyEmbedded }{}},
		{"embedded after field", struct{ fieldThenOmitEmptyEmbedded }{fieldThenOmitEmptyEmbedded{1, embeddedOmitEmptySliceLeaf{[]string{"d"}}}}},
		{"empty embedded on both sides", struct{ embeddedBothSides }{embeddedBothSides{Name: "n"}}},
		{"all empty on both sides", struct{ embeddedBothSides }{}},
		{"three levels all empty", struct{ embeddedMiddle }{}},
		{"three levels with tail", struct{ embeddedMiddle }{embeddedMiddle{Tail: "t"}}},
		{"three levels with slice", struct{ embeddedMiddle }{embeddedMiddle{fieldThenOmitEmptyEmbedded: fieldThenOmitEmptyEmbedded{embeddedOmitEmptySliceLeaf: embeddedOmitEmptySliceLeaf{[]string{"d"}}}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, placement := range []struct {
				name string
				v    interface{}
			}{
				{"value", test.v},
				{"pointer to interface", &test.v},
				{"slice of interface", []interface{}{test.v}},
			} {
				expected, err := stdjson.Marshal(placement.v)
				if err != nil {
					t.Fatal(err)
				}
				got, err := json.Marshal(placement.v)
				if err != nil {
					t.Fatalf("%s: %v", placement.name, err)
				}
				if string(got) != string(expected) {
					t.Fatalf("%s: expected %s but got %s", placement.name, expected, got)
				}
				expectedIndent, err := stdjson.MarshalIndent(placement.v, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				gotIndent, err := json.MarshalIndent(placement.v, "", "  ")
				if err != nil {
					t.Fatalf("%s ( indent ): %v", placement.name, err)
				}
				if string(gotIndent) != string(expectedIndent) {
					t.Fatalf("%s ( indent ): expected %s but got %s", placement.name, expectedIndent, gotIndent)
				}
			}
		})
	}
}

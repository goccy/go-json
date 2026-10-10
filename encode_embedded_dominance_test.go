package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

type (
	dominanceV      struct{ V int }
	dominanceOtherV struct{ V int }
	dominanceDepth2 struct{ dominanceV }
	dominanceDepth3 struct{ dominanceDepth2 }
	dominanceTagged struct {
		V int `json:"V"`
	}
	dominanceTaggedDepth2 struct{ dominanceTagged }
	dominanceOtherTagged  struct {
		Z int `json:"V"`
	}
	dominancePointer struct{ *dominanceV }
	dominanceTwoV    struct {
		dominanceV
		dominanceOtherV
	}
	dominanceThirdV struct{ V int }
	dominanceOneV   struct{ dominanceThirdV }

	// a struct embedded in a struct which it refers to, whose code is compiled for the recursion.
	dominanceRecursive struct {
		B *dominanceEmbedsRecursive
		N int
	}
	dominanceEmbedsRecursive struct {
		*dominanceRecursive
		dominanceNamedField
	}
	dominanceNamedField struct {
		DominanceRecursive int `json:"dominanceRecursive"`
	}
)

// Of the fields of the same name in a struct and the structs embedded in it, only the least deep ones are written,
// and of those, the one with the name in its tag, or none if they are more than one, as encoding/json does.
func TestEncodeEmbeddedFieldDominance(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"depth 2 against depth 1", struct {
			dominanceDepth2
			dominanceOtherV
		}{dominanceDepth2{dominanceV{1}}, dominanceOtherV{2}}, `{"V":2}`},
		{"depth 3 against depth 1", struct {
			dominanceDepth3
			dominanceOtherV
		}{dominanceDepth3{dominanceDepth2{dominanceV{1}}}, dominanceOtherV{2}}, `{"V":2}`},
		{"depth 2 against depth 1 behind a pointer", struct {
			dominancePointer
			dominanceOtherV
		}{dominancePointer{&dominanceV{1}}, dominanceOtherV{2}}, `{"V":2}`},
		{"a tagged field at depth 2 against an untagged one", struct {
			dominanceTaggedDepth2
			dominanceDepth2
		}{dominanceTaggedDepth2{dominanceTagged{1}}, dominanceDepth2{dominanceV{2}}}, `{"V":1}`},
		{"a deeper tagged field against an untagged one", struct {
			dominanceDepth2
			dominanceTaggedDepth2
			dominanceOtherV
		}{dominanceDepth2{dominanceV{1}}, dominanceTaggedDepth2{dominanceTagged{2}}, dominanceOtherV{3}}, `{"V":3}`},
		{"a tagged field against an untagged one", struct {
			dominanceTagged
			dominanceOtherV
		}{dominanceTagged{1}, dominanceOtherV{2}}, `{"V":1}`},
		{"two untagged fields", struct {
			dominanceV
			dominanceOtherV
		}{dominanceV{1}, dominanceOtherV{2}}, `{}`},
		{"two tagged fields", struct {
			dominanceTagged
			dominanceOtherTagged
		}{dominanceTagged{1}, dominanceOtherTagged{2}}, `{}`},
		{"three untagged fields at depth 2 in two embedded structs", struct {
			dominanceTwoV
			dominanceOneV
		}{dominanceTwoV{dominanceV{1}, dominanceOtherV{2}}, dominanceOneV{dominanceThirdV{3}}}, `{}`},
		{"a recursive embedded struct, which writes no name of its own", dominanceRecursive{
			B: &dominanceEmbedsRecursive{&dominanceRecursive{N: 1}, dominanceNamedField{5}},
		}, `{"B":{"B":null,"N":1,"dominanceRecursive":5},"N":0}`},
		{"a field of the struct itself", struct {
			dominanceDepth2
			V int
		}{dominanceDepth2{dominanceV{1}}, 2}, `{"V":2}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.in)
			if err != nil || string(got) != tt.want {
				t.Errorf("got %s, %v, want %s", got, err, tt.want)
			}
		})
	}
}

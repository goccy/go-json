package json_test

import (
	"bytes"
	"io"
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

type (
	dominanceHiddenRecursive struct {
		B *dominanceEmbedsHidden
		dominanceV
	}
	// a recursive struct embedded in a struct which hides a field of it.
	dominanceEmbedsHidden struct {
		*dominanceHiddenRecursive
		V int
	}
	dominanceListRecursive struct {
		B *dominanceEmbedsList
		N int
	}
	dominanceEmbedsList struct {
		*dominanceListRecursive
		X int
	}
	// a recursive struct whose own field hides another of its own, deeper.
	dominanceOwnHidden struct {
		B *dominanceEmbedsOwnHidden
		V int
		dominanceDepth2
	}
	dominanceEmbedsOwnHidden struct {
		*dominanceOwnHidden
		U int
	}
	// a recursive struct whose fields are all hidden by the struct it is embedded in.
	dominanceAllHidden struct {
		B *dominanceEmbedsAllHidden
		V int
	}
	dominanceEmbedsAllHidden struct {
		*dominanceAllHidden
		B *dominanceAllHidden
		V int
	}
	// a recursive struct whose only field written is hidden, beside an array of no elements, which is never written.
	dominanceZeroArray struct {
		X *dominanceEmbedsZeroArray
		Z [0]int `json:",omitempty"`
	}
	dominanceEmbedsZeroArray struct {
		*dominanceZeroArray
		X int
	}
	// a recursive struct whose last field is omitted.
	dominanceLastOmitted struct {
		B *dominanceEmbedsLastOmitted
		V int `json:",omitzero"`
	}
	dominanceEmbedsLastOmitted struct {
		*dominanceLastOmitted
		X int
	}
)

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

// The debug output draws the opcodes of a recursive struct embedded with nothing to write, which jump to nothing.
func TestEncodeEmbeddedRecursiveStructWithNothingToWriteDebug(t *testing.T) {
	in := dominanceAllHidden{B: &dominanceEmbedsAllHidden{&dominanceAllHidden{V: 1}, nil, 2}}
	var dot bytes.Buffer
	got, err := json.MarshalWithOption(in, json.Debug(), json.DebugWith(io.Discard), json.DebugDOT(nopWriteCloser{&dot}))
	if err != nil || string(got) != `{"B":{"B":null,"V":2},"V":0}` {
		t.Errorf("got %s, %v", got, err)
	}
	if dot.Len() == 0 {
		t.Error("no graph is written")
	}
}

// The fields of a recursive struct embedded in another are hidden by the fields of that one, as the fields of any
// embedded struct, and are written at its level.
func TestEncodeEmbeddedRecursiveStructDominance(t *testing.T) {
	tests := []struct {
		name     string
		in       any
		want     string
		indented string
	}{
		{"a field hidden by the struct it is embedded in", dominanceHiddenRecursive{
			B: &dominanceEmbedsHidden{&dominanceHiddenRecursive{dominanceV: dominanceV{1}}, 2},
		}, `{"B":{"B":null,"V":2},"V":0}`, "{\n \"B\": {\n  \"B\": null,\n  \"V\": 2\n },\n \"V\": 0\n}"},
		{"fields at the level of the struct it is embedded in", dominanceListRecursive{
			B: &dominanceEmbedsList{&dominanceListRecursive{N: 1}, 2},
		}, `{"B":{"B":null,"N":1,"X":2},"N":0}`, "{\n \"B\": {\n  \"B\": null,\n  \"N\": 1,\n  \"X\": 2\n },\n \"N\": 0\n}"},
		{"a field which wins over another of the recursive struct", dominanceOwnHidden{
			B: &dominanceEmbedsOwnHidden{&dominanceOwnHidden{V: 4}, 5},
		}, `{"B":{"B":null,"V":4,"U":5},"V":0}`, "{\n \"B\": {\n  \"B\": null,\n  \"V\": 4,\n  \"U\": 5\n },\n \"V\": 0\n}"},
		{"all fields hidden by the struct it is embedded in", dominanceAllHidden{
			B: &dominanceEmbedsAllHidden{&dominanceAllHidden{V: 1}, nil, 2},
		}, `{"B":{"B":null,"V":2},"V":0}`, "{\n \"B\": {\n  \"B\": null,\n  \"V\": 2\n },\n \"V\": 0\n}"},
		{"nothing left to write", dominanceZeroArray{
			X: &dominanceEmbedsZeroArray{&dominanceZeroArray{}, 1},
		}, `{"X":{"X":1}}`, "{\n \"X\": {\n  \"X\": 1\n }\n}"},
		{"the last field omitted", dominanceLastOmitted{
			B: &dominanceEmbedsLastOmitted{&dominanceLastOmitted{}, 1},
		}, `{"B":{"B":null,"X":1}}`, "{\n \"B\": {\n  \"B\": null,\n  \"X\": 1\n }\n}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.in)
			if err != nil || string(got) != tt.want {
				t.Errorf("got %s, %v, want %s", got, err, tt.want)
			}
			got, err = json.MarshalIndent(tt.in, "", " ")
			if err != nil || string(got) != tt.indented {
				t.Errorf("indented: got %s, %v, want %s", got, err, tt.indented)
			}
		})
	}
}

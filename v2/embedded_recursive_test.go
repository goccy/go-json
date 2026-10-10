package json_test

import (
	"testing"

	json "github.com/goccy/go-json/v2"
)

type (
	embeddedRecursive struct {
		X    *embedsRecursive
		V, U int
	}
	// a recursive struct embedded in a struct which hides a member of it.
	embedsRecursive struct {
		*embeddedRecursive
		V int
	}
	embeddedRecursiveV      struct{ V int }
	embeddedRecursiveDepth2 struct{ embeddedRecursiveV }
	// a recursive struct whose own member hides another of its own, deeper.
	embeddedOwnHidden struct {
		B *embedsOwnHidden
		V int
		embeddedRecursiveDepth2
	}
	embedsOwnHidden struct {
		*embeddedOwnHidden
		U int
	}
	// a recursive struct whose only member written is hidden, beside an array of no elements, which is never written.
	embeddedZeroArray struct {
		X *embedsZeroArray
		Z [0]int `json:",omitzero"`
	}
	embedsZeroArray struct {
		*embeddedZeroArray
		X int
	}
	// a recursive struct whose last member is omitted, after a member hidden by the struct it is embedded in.
	embeddedLastOmitted struct {
		B *embedsLastOmitted
		V int    `json:",omitzero"`
		S string `json:",omitempty"`
	}
	embedsLastOmitted struct {
		*embeddedLastOmitted
		S string `json:",omitempty"`
	}
)

// The members of a recursive struct embedded in another are hidden by the members of that one, as encoding/json/v2
// chooses the members of an object.
func TestMarshalEmbeddedRecursiveStruct(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"a member hidden by the struct it is embedded in", embeddedRecursive{
			X: &embedsRecursive{&embeddedRecursive{V: 1, U: 2}, 3},
		}, `{"X":{"X":null,"U":2,"V":3},"V":0,"U":0}`},
		{"a member which wins over another of the recursive struct", embeddedOwnHidden{
			B: &embedsOwnHidden{&embeddedOwnHidden{V: 4}, 5},
		}, `{"B":{"B":null,"V":4,"U":5},"V":0}`},
		{"nothing left to write", embeddedZeroArray{
			X: &embedsZeroArray{&embeddedZeroArray{}, 1},
		}, `{"X":{"X":1}}`},
		{"the last member omitted", embeddedLastOmitted{
			B: &embedsLastOmitted{&embeddedLastOmitted{S: "x"}, ""},
		}, `{"B":{"B":null}}`},
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

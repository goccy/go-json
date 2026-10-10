package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

type (
	lastOmittedE struct{ A int }
	// an embedded struct as the first field, before a recursive struct embedded at the other end.
	lastOmittedFirstEmbedded struct {
		lastOmittedE
		B *lastOmittedEmbedsFirst
		W *int `json:",omitempty"`
	}
	lastOmittedEmbedsFirst struct {
		*lastOmittedFirstEmbedded
		X int
	}
	// a field in the middle whose value embeds a recursive struct.
	lastOmittedMiddle struct {
		B *lastOmittedEmbedsMiddle
		C *lastOmittedEmbedsMiddle
		W *int `json:",omitempty"`
	}
	lastOmittedEmbedsMiddle struct {
		*lastOmittedMiddle
		X int
	}
	// an embedded struct whose last field is omitted, without a recursive struct embedded.
	lastOmittedInner struct {
		Next *lastOmittedOuter
		W    *int `json:",omitempty"`
	}
	lastOmittedOuter struct {
		lastOmittedInner
		Z int
	}
)

// The fields after the last field of an embedded struct, which is omitted, are written, wherever the embedded struct
// is in the struct.
func TestEncodeEmbeddedStructWithLastFieldOmitted(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"an embedded struct first", lastOmittedFirstEmbedded{B: &lastOmittedEmbedsFirst{&lastOmittedFirstEmbedded{}, 1}},
			`{"A":0,"B":{"A":0,"B":null,"X":1}}`},
		{"a struct embedding another in the middle", lastOmittedMiddle{B: &lastOmittedEmbedsMiddle{&lastOmittedMiddle{}, 1}},
			`{"B":{"B":null,"C":null,"X":1},"C":null}`},
		{"an embedded struct followed by a field", lastOmittedOuter{Z: 1}, `{"Next":null,"Z":1}`},
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

package jsontext

import (
	"strings"
	"testing"
	"unsafe"
)

// The place of a value of a member of the name "" has the name, which is empty but not nil, as a level without a
// name has none: a new decoder, whose names have no array yet, keeps the name "" as no bytes ( see
// textcoder.Level ).
func TestDecoderPlaceOfAnEmptyName(t *testing.T) {
	d := NewDecoder(strings.NewReader(`{"":{"x":1}}`))
	for range 2 {
		if _, err := d.ReadToken(); err != nil {
			t.Fatal(err)
		}
	}
	levels := decoderPlace(unsafe.Pointer(d), nil)
	if len(levels) != 2 || levels[1].Name == nil || len(levels[1].Name) != 0 {
		t.Fatalf("got the levels %+v, want the name \"\" in the second one", levels)
	}
}

// The place of the value of a member of the name "" which a new encoder wrote, whose names have no array yet, has
// the name, as the place of a decoder has it.
func TestEncoderPositionOfAnEmptyName(t *testing.T) {
	enc := NewEncoder(new(strings.Builder))
	if err := enc.WriteToken(BeginObject); err != nil {
		t.Fatal(err)
	}
	if err := enc.WriteToken(String("")); err != nil {
		t.Fatal(err)
	}
	levels, _ := encoderPosition(enc, nil)
	if len(levels) != 2 || levels[1].Name == nil || len(levels[1].Name) != 0 {
		t.Fatalf("got the levels %+v, want the name \"\" in the second one", levels)
	}
}

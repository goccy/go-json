//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	"bytes"
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// placeOfEncoder records the pointer which its MarshalJSONTo method is given.
type placeOfEncoder struct{ p *string }

func (v placeOfEncoder) MarshalJSONTo(enc *jsontext.Encoder) error {
	*v.p = string(enc.StackPointer())
	return enc.WriteToken(jsontext.Null)
}

// stdPlaceOfEncoder is placeOfEncoder for encoding/json/v2.
type stdPlaceOfEncoder struct{ p *string }

func (v stdPlaceOfEncoder) MarshalJSONTo(enc *stdjsontext.Encoder) error {
	*v.p = string(enc.StackPointer())
	return enc.WriteToken(stdjsontext.Null)
}

// TestMarshalEncodeAtAMemberOfAnEmptyName checks the place of the error of MarshalEncode, and the pointer which a
// MarshalJSONTo method is given, at the value of a member of the name "" which a new encoder wrote, whose names have
// no array yet: the place has the name, as encoding/json/v2 has it.
func TestMarshalEncodeAtAMemberOfAnEmptyName(t *testing.T) {
	begin := func(enc *jsontext.Encoder, std *stdjsontext.Encoder) {
		t.Helper()
		for range 2 {
			if err := enc.WriteToken(jsontext.BeginObject); err != nil {
				t.Fatal(err)
			}
			if err := enc.WriteToken(jsontext.String("")); err != nil {
				t.Fatal(err)
			}
			if err := std.WriteToken(stdjsontext.BeginObject); err != nil {
				t.Fatal(err)
			}
			if err := std.WriteToken(stdjsontext.String("")); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, v := range []any{make(chan int), map[string]any{"a": make(chan int)}, map[string]any{"": make(chan int)}} {
		enc, std := jsontext.NewEncoder(new(bytes.Buffer)), stdjsontext.NewEncoder(new(bytes.Buffer))
		begin(enc, std)
		var se *json.SemanticError
		var stdSe *stdjson.SemanticError
		if err, stdErr := json.MarshalEncode(enc, v), stdjson.MarshalEncode(std, v); !errors.As(err, &se) || !errors.As(stdErr, &stdSe) {
			t.Fatalf("%v: got %v and %v, want SemanticErrors", v, err, stdErr)
		}
		if se.ByteOffset != stdSe.ByteOffset || string(se.JSONPointer) != string(stdSe.JSONPointer) {
			t.Fatalf("%v: got the place %d %q, want %d %q", v, se.ByteOffset, se.JSONPointer, stdSe.ByteOffset, stdSe.JSONPointer)
		}
	}
	var got, want string
	enc, std := jsontext.NewEncoder(new(bytes.Buffer)), stdjsontext.NewEncoder(new(bytes.Buffer))
	begin(enc, std)
	if err := json.MarshalEncode(enc, placeOfEncoder{&got}); err != nil {
		t.Fatal(err)
	}
	if err := stdjson.MarshalEncode(std, stdPlaceOfEncoder{&want}); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("the pointer of the method: got %q, want %q", got, want)
	}
}

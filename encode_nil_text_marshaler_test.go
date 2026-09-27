package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// A nil pointer to a TextMarshaler is null as a value, wherever the value is, and "" as the key of a map,
// as encoding/json writes them, whether MarshalText has a value receiver or a pointer receiver.

type nilTextValue struct{ S string }

func (v nilTextValue) MarshalText() ([]byte, error) { return []byte(v.S), nil }

type nilTextPointer struct{ S string }

func (v *nilTextPointer) MarshalText() ([]byte, error) { return []byte(v.S), nil }

func TestEncodeNilTextMarshaler(t *testing.T) {
	values := []any{
		[]*nilTextValue{nil, {"x"}},
		[]*nilTextPointer{nil, {"x"}},
		[2]*nilTextValue{},
		[2]*nilTextPointer{},
		[]**nilTextValue{nil},
		map[string]*nilTextValue{"a": nil, "b": {"x"}},
		map[string]*nilTextPointer{"a": nil, "b": {"x"}},
		map[*nilTextValue]int{nil: 1, {"x"}: 2},
		map[*nilTextPointer]int{nil: 1, {"x"}: 2},
		map[*nilTextValue]*nilTextValue{nil: nil},
		map[*nilTextPointer]*nilTextPointer{nil: nil},
		map[*nilTextValue][]*nilTextValue{nil: {nil}},
		map[string]any{"a": (*nilTextValue)(nil), "b": (*nilTextPointer)(nil)},
		struct {
			A *nilTextValue
			B *nilTextPointer
			C []*nilTextValue
			D map[string]*nilTextPointer
		}{C: []*nilTextValue{nil}, D: map[string]*nilTextPointer{"d": nil}},
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
	}
}

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"testing"

	"github.com/goccy/go-json"
)

type mapKeyNamed string

// The key of a map whose type is a plain string is written by the map opcodes themselves.
// Every kind of a key must be encoded as encoding/json does, with and without the escape of HTML.
func TestEncodeMapKeys(t *testing.T) {
	for _, test := range []struct {
		name string
		v    interface{}
	}{
		{"string", map[string]int{"b": 2, "a": 1, "": 0}},
		{"string to escape", map[string]int{`"q"`: 1, `back\slash`: 2, "\n": 3, "<html>&": 4, "日本語": 5, "\xff": 6}},
		{"named string", map[mapKeyNamed]int{"b": 2, "a": 1}},
		{"int", map[int]string{2: "b", -1: "a"}},
		{"bool value", map[string]bool{"t": true, "f": false}},
		{"nested", map[string]map[string]int{"x": {"b": 2, "a": 1}, "y": {}}},
		{"empty", map[string]int{}},
		{"nil", map[string]int(nil)},
		{"one", map[string][]int{"only": {1, 2}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, escape := range []bool{true, false} {
				var expected []byte
				var buf bytes.Buffer
				enc := stdjson.NewEncoder(&buf)
				enc.SetEscapeHTML(escape)
				if err := enc.Encode(test.v); err != nil {
					t.Fatal(err)
				}
				expected = bytes.TrimSuffix(buf.Bytes(), []byte("\n"))

				var opts []json.EncodeOptionFunc
				if !escape {
					opts = append(opts, json.DisableHTMLEscape())
				}
				got, err := json.MarshalWithOption(test.v, opts...)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(expected, got) {
					t.Fatalf("escape=%v: expected %s but got %s", escape, expected, got)
				}
				indented, err := json.MarshalIndentWithOption(test.v, "", "  ", opts...)
				if err != nil {
					t.Fatal(err)
				}
				buf.Reset()
				enc = stdjson.NewEncoder(&buf)
				enc.SetEscapeHTML(escape)
				enc.SetIndent("", "  ")
				if err := enc.Encode(test.v); err != nil {
					t.Fatal(err)
				}
				if expected := bytes.TrimSuffix(buf.Bytes(), []byte("\n")); !bytes.Equal(expected, indented) {
					t.Fatalf("escape=%v, indent: expected %s but got %s", escape, expected, indented)
				}
			}
		})
	}
}

type mapValueFails struct{}

func (mapValueFails) MarshalJSON() ([]byte, error) { return nil, errors.New("fails") }

// The elements of a map whose keys are sorted are encoded into a buffer of their own. An error in the middle
// of a map, also of a nested one, must not break the encoding which follows with the same pooled context.
func TestEncodeSortedMapAfterError(t *testing.T) {
	broken := map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2, "d": mapValueFails{}}, "e": 3}
	fine := map[string]interface{}{"x": map[string]int{"z": 1, "y": 2}, "w": []interface{}{map[string]int{"b": 1, "a": 2}}}
	expected, err := stdjson.Marshal(fine)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50; i++ {
		if _, err := json.Marshal(broken); err == nil {
			t.Fatal("expected an error")
		}
		got, err := json.Marshal(fine)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expected, got) {
			t.Fatalf("expected %s but got %s", expected, got)
		}
	}
}

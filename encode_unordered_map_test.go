package json_test

import (
	stdjson "encoding/json"
	"reflect"
	"regexp"
	"testing"

	"github.com/goccy/go-json"
)

// An unordered map with values of interface{} is written as it is read: the entries which hold a scalar first,
// and the others after them by their opcodes. Every kind of a value, and every option which changes how a
// value is written, must give the same object as encoding/json, in any order.

// the colors are escape sequences of the terminal around the tokens.
var ansiEscapes = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type unorderedMapMarshaler struct{ V int }

func (m unorderedMapMarshaler) MarshalJSON() ([]byte, error) { return []byte(`{"m":1}`), nil }

type unorderedMapKey string

func TestEncodeUnorderedMapOfInterfaces(t *testing.T) {
	var nilPtr *int
	one := 1
	scalars := map[string]interface{}{
		"int": 1, "int8": int8(-2), "uint": uint(3), "float": 1.5, "float32": float32(2.5), "string": "s<>& ",
		"bool": true, "nil": nil, "number": stdjson.Number("12.5"), "bytes": []byte("bytes"),
	}
	others := map[string]interface{}{
		"map": map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": nil}}, "slice": []interface{}{1, "s", nil},
		"struct": struct{ A int }{1}, "ptr": &one, "nilptr": nilPtr, "marshaler": unorderedMapMarshaler{1},
		"ptrmarshaler": &unorderedMapMarshaler{2}, "ints": []int{1, 2}, "strings": map[string]string{"k": "v"},
		"empty": map[string]interface{}{}, "emptyslice": []interface{}{},
	}
	both := map[string]interface{}{}
	for k, v := range scalars {
		both[k] = v
	}
	for k, v := range others {
		both[k] = v
	}
	values := []interface{}{
		scalars, others, both, map[string]interface{}{}, map[string]interface{}(nil),
		map[unorderedMapKey]interface{}{"a": 1, "b": []int{2}},
		[]map[string]interface{}{both, scalars},
		struct {
			M map[string]interface{}
			N map[string]interface{} `json:",omitempty"`
		}{M: both},
	}
	encoders := []struct {
		name   string
		encode func(v interface{}) ([]byte, error)
	}{
		{"unordered", func(v interface{}) ([]byte, error) { return json.MarshalWithOption(v, json.UnorderedMap()) }},
		{"unordered no escape", func(v interface{}) ([]byte, error) {
			return json.MarshalWithOption(v, json.UnorderedMap(), json.DisableHTMLEscape(), json.DisableNormalizeUTF8())
		}},
		{"unordered indent", func(v interface{}) ([]byte, error) {
			return json.MarshalIndentWithOption(v, "", "  ", json.UnorderedMap())
		}},
		{"unordered colorized", func(v interface{}) ([]byte, error) {
			return json.MarshalWithOption(v, json.UnorderedMap(), json.Colorize(json.DefaultColorScheme))
		}},
		{"unordered colorized indent", func(v interface{}) ([]byte, error) {
			return json.MarshalIndentWithOption(v, "", "  ", json.UnorderedMap(), json.Colorize(json.DefaultColorScheme))
		}},
	}
	for _, value := range values {
		expected, err := stdjson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var want interface{}
		if err := stdjson.Unmarshal(expected, &want); err != nil {
			t.Fatal(err)
		}
		for _, enc := range encoders {
			b, err := enc.encode(value)
			if err != nil {
				t.Fatalf("%s: %T: %v", enc.name, value, err)
			}
			var got interface{}
			if err := stdjson.Unmarshal(ansiEscapes.ReplaceAll(b, nil), &got); err != nil {
				t.Fatalf("%s: %T: %v\n%s", enc.name, value, err, b)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s: %T:\n got %s\nwant %s", enc.name, value, b, expected)
			}
		}
	}
}

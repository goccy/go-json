package json_test

import (
	stdjson "encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/goccy/go-json"
)

// An unordered map with keys of a string kind is written as it is read when its values are scalars, or values
// of interface{}: then the entries which hold a scalar come first, and the others after them by their opcodes.
// Every kind of a value, and every option which changes how a value is written, must give the same object as
// encoding/json, in any order.

// the colors are escape sequences of the terminal around the tokens.
var ansiEscapes = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type unorderedMapMarshaler struct{ V int }

func (m unorderedMapMarshaler) MarshalJSON() ([]byte, error) { return []byte(`{"m":1}`), nil }

type unorderedMapKey string

var unorderedMapEncoders = []struct {
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

// assertUnorderedMapsEqual encodes each value by every encoder of an unordered map and compares the object
// with the one of encoding/json.
func assertUnorderedMapsEqual(t *testing.T, values []interface{}) {
	t.Helper()
	for _, value := range values {
		expected, err := stdjson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var want interface{}
		if err := stdjson.Unmarshal(expected, &want); err != nil {
			t.Fatal(err)
		}
		for _, enc := range unorderedMapEncoders {
			b, err := enc.encode(value)
			if err != nil {
				t.Fatalf("%s: %T: %v", enc.name, value, err)
			}
			var got interface{}
			if err := stdjson.Unmarshal(ansiEscapes.ReplaceAll(b, nil), &got); err != nil {
				t.Fatalf("%s: %T: %v\n%s", enc.name, value, err, b)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s: %T:\n got %.300s\nwant %.300s", enc.name, value, b, expected)
			}
		}
	}
}

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
	assertUnorderedMapsEqual(t, values)
}

type unorderedMapString string

func TestEncodeUnorderedMapOfScalars(t *testing.T) {
	stringKey := func(i int) string { return "k" + strconv.Itoa(i*7919%1000) }
	one := 1
	values := []interface{}{
		map[string]string{}, map[string]string(nil), map[string]bool(nil),
	}
	for _, n := range []int{1, 2, 9, 100, 1000} {
		values = append(values,
			mapOf(n, stringKey, func(i int) string { return "v<>&\u2028" + strconv.Itoa(i) }),
			mapOf(n, stringKey, func(i int) unorderedMapString { return unorderedMapString(strconv.Itoa(i)) }),
			mapOf(n, stringKey, func(i int) bool { return i%2 == 0 }),
			mapOf(n, stringKey, func(i int) int8 { return int8(i) }),
			mapOf(n, stringKey, func(i int) int16 { return int16(-i) }),
			mapOf(n, stringKey, func(i int) int32 { return int32(i) }),
			mapOf(n, stringKey, func(i int) int { return -i * 1000000 }),
			mapOf(n, stringKey, func(i int) uint8 { return uint8(i) }),
			mapOf(n, stringKey, func(i int) uint64 { return uint64(i) << 40 }),
			mapOf(n, stringKey, func(i int) float32 { return float32(i) + 0.25 }),
			mapOf(n, stringKey, func(i int) float64 { return float64(i) * 1e-7 }),
			mapOf(n, stringKey, func(i int) []byte { return []byte(strconv.Itoa(i)) }),
			mapOf(n, stringKey, func(i int) stdjson.Number { return stdjson.Number(strconv.Itoa(i)) }),
			mapOf(n, func(i int) unorderedMapKey { return unorderedMapKey(stringKey(i)) }, func(i int) string { return "v" }),
			// the values which are not written as the map is read: pointers, structs, and keys with a marshaler.
			mapOf(n, stringKey, func(i int) *int { return &one }),
			mapOf(n, stringKey, func(i int) *string { return nil }),
			mapOf(n, stringKey, func(i int) struct{ A int } { return struct{ A int }{i} }),
			mapOf(n, stringKey, func(i int) unorderedMapMarshaler { return unorderedMapMarshaler{i} }),
			mapOf(n, func(i int) mapTextKey { return mapTextKey{i} }, func(i int) string { return "v" }),
		)
	}
	values = append(values,
		struct {
			M map[string]int
			N map[string]string `json:",omitempty"`
			E map[string]bool
		}{M: map[string]int{"a": 1}, E: map[string]bool{}},
		[]map[string]float64{{"a": 1.5}, {}},
		map[string]map[string]int{"a": {"x": 1, "y": 2}, "b": {}},
	)
	assertUnorderedMapsEqual(t, values)
}

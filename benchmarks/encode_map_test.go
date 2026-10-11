//go:build go1.25

package benchmark

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
	gojsontext "github.com/goccy/go-json/jsontext"
	gojsonv2 "github.com/goccy/go-json/v2"
)

// The benchmarks of maps under the conditions which take different paths of the encoder of maps:
//
//   - The order. A sorted map ( SortedMap: go-json by default, go-json/v2 with Deterministic( true ) ) is sorted by
//     its keys when they are strings, and the name of every other key ( an integer, or the text of a key's
//     MarshalText ) is written before the entries are sorted by it. An unordered map ( UnorderedMap: go-json with
//     UnorderedMap(), go-json/v2 by default ) is written in the order of the map: as it is read when the keys are
//     strings and the values scalars or of interface{}, and else through the entries read first. go-json/v2 checks
//     the names of an unordered map for duplicates when its keys are named by a method, as MarshalText: the names
//     of strings, integers and floats are different for every key.
//   - The keys: strings, a named string type, int, int8, uint, float64 ( go-json/v2 only: go-json rejects it as
//     encoding/json before Go 1.27 does ), keys of MarshalText, and keys whose names are escaped, which a sorted
//     map compares by their unescaped names. SortedMapNoHTMLEscape writes the escaped ones with DisableHTMLEscape(),
//     whose names are escaped by other code than the default's.
//   - The sizes: empty, 1, 10, 100 and 1000 entries, for the keys of strings and of integers with values of
//     integers: an empty map ends before its entries are read, and the cost of an entry and of the sort grows
//     with the size.
//   - The values: integers and strings ( scalars ), structs and maps ( written by their opcodes ), and interface{},
//     whose map of a JSON object is written as a scalar. The map of interface{} values of mixed kinds is measured by
//     Benchmark_Encode_MapInterface_*.
//   - Compact and indented ( SortedMapIndent: MarshalIndent of go-json, jsontext.Multiline( true ) and
//     jsontext.WithIndent of go-json/v2 ).
//
// Not every payload is measured under every condition, so that the encode benchmarks of the CI stay within their
// time ( see mapBench ): the sorted maps of go-json run every payload, and the other conditions the ones whose path
// they change. Empty maps and maps of one entry end or skip the sort by the same code in every condition, an
// unordered map writes values of strings by the same code as the ones of integers, and go-json/v2, which writes
// the entries by the code of go-json, runs the kinds of keys and of values which take code of their own, and the
// sizes 10, 100 and 1000. Encoder.Encode and MarshalNoEscape of go-json encode a map by the same code as
// Marshal, so they aren't measured apart.
//
// The libraries: GoJson is go-json, GoJsonV2 is go-json/v2, EncodingJson is encoding/json, which always sorts
// the keys of a map, and Sonic and SonicStd are sonic.ConfigDefault, which doesn't sort them nor escape HTML,
// and sonic.ConfigStd, which does both as encoding/json does. Not every pair does the same work for the names
// which are escaped ( EscapedString10, EscapedText10 ): GoJson escapes HTML and normalizes UTF-8 where Sonic and
// GoJsonV2 don't. GoJsonLikeSonic is go-json with the options which make it do what Sonic does ( see
// likeSonicOptions ), measured with the escaped names only: for the other names, which have no character
// escaped for HTML and are valid UTF-8, it does what GoJson does. So the unordered EscapedText10 of Sonic is to
// be compared with the one of GoJsonLikeSonic, and the sorted Escaped* of GoJsonV2 with the ones of
// SortedMapNoHTMLEscape_GoJson. Each sub-benchmark checks once, before its first
// round, that the library writes what encoding/json writes for the payload: the same bytes when the map is
// sorted, and the same JSON value, decoded, when it isn't.

// mapBench is a set of the conditions which a payload is measured under.
type mapBench uint8

const (
	// mapSorted is the sorted maps, compact, of go-json, encoding/json and sonic.
	mapSorted mapBench = 1 << iota
	// mapUnordered is the unordered maps of go-json and sonic.
	mapUnordered
	// mapIndent is the sorted maps, indented.
	mapIndent
	// mapNoHTMLEscape is the sorted maps of go-json with DisableHTMLEscape().
	mapNoHTMLEscape
	// mapLikeSonic is the unordered maps of go-json with the options of GoJsonLikeSonic.
	mapLikeSonic
	// mapV2 is the maps of go-json/v2, under the conditions above which it has ( sorted, unordered, indented ).
	mapV2
	// mapV2Only is a payload of go-json/v2 only.
	mapV2Only
)

// mapTextKey is a key written by its MarshalText.
type mapTextKey struct {
	Region int
	ID     int
}

func (k mapTextKey) MarshalText() ([]byte, error) {
	b := strconv.AppendInt(nil, int64(k.Region), 10)
	b = append(b, '-')
	return strconv.AppendInt(b, int64(k.ID), 10), nil
}

// mapEscapedTextKey is a key whose MarshalText writes a name which is escaped.
type mapEscapedTextKey int

func (k mapEscapedTextKey) MarshalText() ([]byte, error) {
	return []byte(mapEscapedName(int(k))), nil
}

// mapEscapedName is the name of the i-th key which is escaped: a quote, a backslash, the characters which
// encoding/json escapes for HTML, a control character and characters which aren't ASCII. The names begin with
// their numbers, so that they are in the same order by their escaped names as by their names.
func mapEscapedName(i int) string {
	return strconv.Itoa(1000+i) + "\"q\\<a&b>\x01é日本"
}

type mapNamedString string

// mapRecord is the value of a map of structs.
type mapRecord struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
	Tags  []int   `json:"tags"`
}

// mapKey returns the i-th key of a map: the keys are spread, so that their order in the map, their sorted order
// and the order of their numbers differ.
func mapKey(i int) int {
	return (i*7919)%10007 - 5000
}

func newMap[K comparable, V any](n int, key func(int) K, value func(int) V) map[K]V {
	m := make(map[K]V, n)
	for i := 0; i < n; i++ {
		m[key(i)] = value(i)
	}
	return m
}

func mapIntKey(i int) int         { return mapKey(i) }
func mapStringKey(i int) string   { return "key" + strconv.Itoa(mapKey(i)) }
func mapIntValue(i int) int       { return i * 3 }
func mapStringValue(i int) string { return "value " + strconv.Itoa(i) }
func mapRecordValue(i int) mapRecord {
	return mapRecord{ID: i, Name: "record " + strconv.Itoa(i), Score: float64(i) / 4, Tags: []int{i, i + 1}}
}

func mapAnyValue(i int) any {
	switch i % 4 {
	case 0:
		return i
	case 1:
		return "value " + strconv.Itoa(i)
	case 2:
		return i%3 == 0
	default:
		return []any{i, "x"}
	}
}

// mapFloatKey is the i-th key of a map of float64, and mapFloatName its name, as a JSON number is written.
func mapFloatKey(i int) float64 { return float64(mapKey(i)) / 4 }
func mapFloatName(i int) string { return strconv.FormatFloat(mapFloatKey(i), 'g', -1, 64) }

// mapPayload is a map measured under the conditions of bench. want is what encoding/json is given to write the
// expected output, the payload itself unless it is set.
type mapPayload struct {
	name  string
	value any
	want  any
	bench mapBench
}

const (
	mapAll       = mapSorted | mapUnordered | mapV2
	mapSortedAll = mapSorted | mapV2
)

var mapPayloads = []mapPayload{
	// the sizes, of the keys of strings and of integers.
	{name: "StringInt0", value: map[string]int{}, bench: mapSorted | mapUnordered},
	{name: "StringInt1", value: newMap(1, mapStringKey, mapIntValue), bench: mapSorted},
	{name: "StringInt10", value: newMap(10, mapStringKey, mapIntValue), bench: mapAll | mapIndent},
	{name: "StringInt100", value: newMap(100, mapStringKey, mapIntValue), bench: mapAll | mapIndent},
	{name: "StringInt1000", value: newMap(1000, mapStringKey, mapIntValue), bench: mapSorted | mapUnordered},
	{name: "IntInt0", value: map[int]int{}, bench: mapSorted | mapUnordered},
	{name: "IntInt1", value: newMap(1, mapIntKey, mapIntValue), bench: mapSorted},
	{name: "IntInt10", value: newMap(10, mapIntKey, mapIntValue), bench: mapAll | mapIndent},
	{name: "IntInt100", value: newMap(100, mapIntKey, mapIntValue), bench: mapAll | mapIndent},
	{name: "IntInt1000", value: newMap(1000, mapIntKey, mapIntValue), bench: mapAll},
	// the other keys.
	{name: "NamedStringInt10", value: newMap(10, func(i int) mapNamedString { return mapNamedString(mapStringKey(i)) }, mapIntValue), bench: mapSorted},
	{name: "Int8Int10", value: newMap(10, func(i int) int8 { return int8(mapKey(i)) }, mapIntValue), bench: mapSorted},
	{name: "UintInt10", value: newMap(10, func(i int) uint { return uint(mapKey(i) + 5000) }, mapIntValue), bench: mapSorted},
	{
		name:  "Float64Int10",
		value: newMap(10, mapFloatKey, mapIntValue),
		want:  newMap(10, mapFloatName, mapIntValue),
		bench: mapAll | mapV2Only,
	},
	{name: "TextString20", value: newMap(20, func(i int) mapTextKey { return mapTextKey{Region: i % 4, ID: mapKey(i)} }, mapStringValue), bench: mapAll | mapIndent},
	{name: "EscapedString10", value: newMap(10, mapEscapedName, mapIntValue), bench: mapSortedAll | mapNoHTMLEscape},
	{name: "EscapedText10", value: newMap(10, func(i int) mapEscapedTextKey { return mapEscapedTextKey(i) }, mapIntValue), bench: mapAll | mapIndent | mapNoHTMLEscape | mapLikeSonic},
	// the values.
	{name: "StringString100", value: newMap(100, mapStringKey, mapStringValue), bench: mapSorted},
	{name: "StringRecord20", value: newMap(20, mapStringKey, mapRecordValue), bench: mapAll | mapIndent},
	{name: "StringNested10x3", value: newMap(10, mapStringKey, func(int) map[string]int { return newMap(3, mapStringKey, mapIntValue) }), bench: mapSorted | mapUnordered},
	{name: "StringAny10", value: newMap(10, mapStringKey, mapAnyValue), bench: mapAll},
	{name: "IntRecord20", value: newMap(20, mapIntKey, mapRecordValue), bench: mapAll | mapIndent},
	{name: "IntNested10x3", value: newMap(10, mapIntKey, func(int) map[int]int { return newMap(3, mapIntKey, mapIntValue) }), bench: mapAll | mapIndent},
}

func init() {
	for _, p := range mapPayloads {
		if p.bench&mapV2Only == 0 {
			sonicTypes = append(sonicTypes, reflect.TypeOf(p.value))
		}
	}
}

// mapLibrary is how a library encodes a map under a condition.
type mapLibrary struct {
	// v2 is set for go-json/v2, which runs the payloads of mapV2 only, and the others the ones without mapV2Only.
	v2 bool
	// sorted is whether the output is sorted: it is compared with the bytes of encoding/json, and else with its
	// JSON value. indent is whether it is indented, and htmlEscape whether it escapes HTML, as encoding/json does.
	sorted, indent, htmlEscape bool
	marshal                    func(any) ([]byte, error)
}

// mapExpected returns what encoding/json writes for v, indented or not and with HTML escaped or not.
func mapExpected(v any, indent, htmlEscape bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(htmlEscape)
	if indent {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// sameJSONValue reports whether a and b are the same JSON value, whatever the order of their members.
func sameJSONValue(a, b []byte) bool {
	var va, vb any
	if json.Unmarshal(a, &va) != nil || json.Unmarshal(b, &vb) != nil {
		return false
	}
	return reflect.DeepEqual(va, vb)
}

// benchMap runs a sub-benchmark of the library for each payload of the condition bench.
func benchMap(b *testing.B, bench mapBench, lib mapLibrary) {
	for _, p := range mapPayloads {
		if p.bench&bench == 0 || (lib.v2 && p.bench&mapV2 == 0) || (!lib.v2 && p.bench&mapV2Only != 0) {
			continue
		}
		b.Run(p.name, func(b *testing.B) {
			want := p.want
			if want == nil {
				want = p.value
			}
			expected, err := mapExpected(want, lib.indent, lib.htmlEscape)
			if err != nil {
				b.Fatal(err)
			}
			got, err := lib.marshal(p.value)
			if err != nil {
				b.Fatal(err)
			}
			if lib.sorted && !bytes.Equal(got, expected) || !lib.sorted && !sameJSONValue(got, expected) {
				b.Fatalf("got %s, want %s", got, expected)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := lib.marshal(p.value); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func Benchmark_Encode_SortedMap_GoJson(b *testing.B) {
	benchMap(b, mapSorted, mapLibrary{sorted: true, htmlEscape: true, marshal: gojson.Marshal})
}

func Benchmark_Encode_SortedMap_GoJsonV2(b *testing.B) {
	benchMap(b, mapSorted, mapLibrary{v2: true, sorted: true, marshal: func(v any) ([]byte, error) {
		return gojsonv2.Marshal(v, gojsonv2.Deterministic(true))
	}})
}

func Benchmark_Encode_SortedMap_EncodingJson(b *testing.B) {
	benchMap(b, mapSorted, mapLibrary{sorted: true, htmlEscape: true, marshal: json.Marshal})
}

func Benchmark_Encode_SortedMap_SonicStd(b *testing.B) {
	pretouchSonic()
	benchMap(b, mapSorted, mapLibrary{sorted: true, htmlEscape: true, marshal: sonic.ConfigStd.Marshal})
}

func Benchmark_Encode_SortedMapNoHTMLEscape_GoJson(b *testing.B) {
	benchMap(b, mapNoHTMLEscape, mapLibrary{sorted: true, marshal: func(v any) ([]byte, error) {
		return gojson.MarshalWithOption(v, gojson.DisableHTMLEscape())
	}})
}

func Benchmark_Encode_UnorderedMap_GoJson(b *testing.B) {
	benchMap(b, mapUnordered, mapLibrary{marshal: func(v any) ([]byte, error) {
		return gojson.MarshalWithOption(v, gojson.UnorderedMap())
	}})
}

func Benchmark_Encode_UnorderedMap_GoJsonLikeSonic(b *testing.B) {
	benchMap(b, mapLikeSonic, mapLibrary{marshal: func(v any) ([]byte, error) {
		return gojson.MarshalWithOption(v, likeSonicOptions...)
	}})
}

func Benchmark_Encode_UnorderedMap_GoJsonV2(b *testing.B) {
	benchMap(b, mapUnordered, mapLibrary{v2: true, marshal: func(v any) ([]byte, error) {
		return gojsonv2.Marshal(v)
	}})
}

func Benchmark_Encode_UnorderedMap_Sonic(b *testing.B) {
	pretouchSonic()
	benchMap(b, mapUnordered, mapLibrary{marshal: sonic.ConfigDefault.Marshal})
}

func Benchmark_Encode_SortedMapIndent_GoJson(b *testing.B) {
	benchMap(b, mapIndent, mapLibrary{sorted: true, indent: true, htmlEscape: true, marshal: func(v any) ([]byte, error) {
		return gojson.MarshalIndent(v, "", "  ")
	}})
}

func Benchmark_Encode_SortedMapIndent_GoJsonV2(b *testing.B) {
	benchMap(b, mapIndent, mapLibrary{v2: true, sorted: true, indent: true, marshal: func(v any) ([]byte, error) {
		return gojsonv2.Marshal(v, gojsonv2.Deterministic(true), gojsontext.Multiline(true), gojsontext.WithIndent("  "))
	}})
}

func Benchmark_Encode_SortedMapIndent_EncodingJson(b *testing.B) {
	benchMap(b, mapIndent, mapLibrary{sorted: true, indent: true, htmlEscape: true, marshal: func(v any) ([]byte, error) {
		return json.MarshalIndent(v, "", "  ")
	}})
}

func Benchmark_Encode_SortedMapIndent_SonicStd(b *testing.B) {
	pretouchSonic()
	benchMap(b, mapIndent, mapLibrary{sorted: true, indent: true, htmlEscape: true, marshal: func(v any) ([]byte, error) {
		return sonic.ConfigStd.MarshalIndent(v, "", "  ")
	}})
}

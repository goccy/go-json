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

// The maps whose entries are written in the order of their names: encoding/json and go-json sort the names of every
// map, and go-json/v2 sorts them with Deterministic( true ). The name of a key which is not a string, an integer or
// the text of a key's MarshalText, is written before the entries are sorted by it, so these maps take another path
// than the maps of strings, which are here too, of the same sizes, to be compared with them.
//
// Each benchmark encodes every payload of sortedMapPayloads in a sub-benchmark of its name. The libraries:
//
//   - GoJson is Marshal or MarshalIndent of go-json, GoJsonV2 is Marshal of go-json/v2 with Deterministic( true ),
//     and with jsontext.Multiline( true ) and an indent of two spaces for the Indent benchmarks.
//   - EncodingJson is encoding/json.
//   - SonicStd is sonic.ConfigStd, which sorts the keys of a map as encoding/json does. sonic.ConfigDefault, the
//     Sonic of the other benchmarks, doesn't sort them, so it isn't compared here.
//
// Every library writes the same bytes for a payload, which is checked once before the first round of each
// sub-benchmark: the names have no character which encoding/json escapes for HTML and go-json/v2 doesn't.

// sortedMapTextKey is a key written by its MarshalText.
type sortedMapTextKey struct {
	Region int
	ID     int
}

func (k sortedMapTextKey) MarshalText() ([]byte, error) {
	b := strconv.AppendInt(nil, int64(k.Region), 10)
	b = append(b, '-')
	return strconv.AppendInt(b, int64(k.ID), 10), nil
}

// sortedMapRecord is the value of a map of records.
type sortedMapRecord struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
	Tags  []int   `json:"tags"`
}

// sortedMapKey returns the i-th key of a map: the keys are spread, so that their order in the map, their sorted
// order and the order of their numbers differ.
func sortedMapKey(i int) int {
	return (i*7919)%10007 - 5000
}

func newSortedMapInts(n int) map[int]int {
	m := make(map[int]int, n)
	for i := 0; i < n; i++ {
		m[sortedMapKey(i)] = i * 3
	}
	return m
}

func newSortedMapStrings(n int) map[string]int {
	m := make(map[string]int, n)
	for i := 0; i < n; i++ {
		m["key"+strconv.Itoa(sortedMapKey(i))] = i * 3
	}
	return m
}

func newSortedMapRecords(n int) map[int]sortedMapRecord {
	m := make(map[int]sortedMapRecord, n)
	for i := 0; i < n; i++ {
		m[sortedMapKey(i)] = sortedMapRecord{ID: i, Name: "record " + strconv.Itoa(i), Score: float64(i) / 4, Tags: []int{i, i + 1}}
	}
	return m
}

func newSortedMapNested(outer, inner int) map[int]map[int]int {
	m := make(map[int]map[int]int, outer)
	for i := 0; i < outer; i++ {
		m[sortedMapKey(i)] = newSortedMapInts(inner)
	}
	return m
}

func newSortedMapTexts(n int) map[sortedMapTextKey]string {
	m := make(map[sortedMapTextKey]string, n)
	for i := 0; i < n; i++ {
		m[sortedMapTextKey{Region: i % 4, ID: sortedMapKey(i)}] = "value " + strconv.Itoa(i)
	}
	return m
}

// sortedMapPayloads are the payloads of the benchmarks of sorted maps, by the names of their sub-benchmarks.
var sortedMapPayloads = []struct {
	name  string
	value any
}{
	{"Int10", newSortedMapInts(10)},
	{"Int100", newSortedMapInts(100)},
	{"Record20", newSortedMapRecords(20)},
	{"Nested10x3", newSortedMapNested(10, 3)},
	{"Text20", newSortedMapTexts(20)},
	{"String10", newSortedMapStrings(10)},
	{"String100", newSortedMapStrings(100)},
}

func init() {
	for _, p := range sortedMapPayloads {
		sonicTypes = append(sonicTypes, reflect.TypeOf(p.value))
	}
}

// benchSortedMap runs a sub-benchmark of marshal for each payload, after checking that it writes what
// encoding/json does ( want ).
func benchSortedMap(b *testing.B, want, marshal func(any) ([]byte, error)) {
	for _, p := range sortedMapPayloads {
		b.Run(p.name, func(b *testing.B) {
			expected, err := want(p.value)
			if err != nil {
				b.Fatal(err)
			}
			got, err := marshal(p.value)
			if err != nil {
				b.Fatal(err)
			}
			if !bytes.Equal(got, expected) {
				b.Fatalf("got %s, want %s", got, expected)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := marshal(p.value); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func marshalIndentEncodingJson(v any) ([]byte, error) { return json.MarshalIndent(v, "", "  ") }

func Benchmark_Encode_SortedMap_GoJson(b *testing.B) {
	benchSortedMap(b, json.Marshal, gojson.Marshal)
}

func Benchmark_Encode_SortedMap_GoJsonV2(b *testing.B) {
	benchSortedMap(b, json.Marshal, func(v any) ([]byte, error) {
		return gojsonv2.Marshal(v, gojsonv2.Deterministic(true))
	})
}

func Benchmark_Encode_SortedMap_EncodingJson(b *testing.B) {
	benchSortedMap(b, json.Marshal, json.Marshal)
}

func Benchmark_Encode_SortedMap_SonicStd(b *testing.B) {
	pretouchSonic()
	benchSortedMap(b, json.Marshal, sonic.ConfigStd.Marshal)
}

func Benchmark_Encode_SortedMapIndent_GoJson(b *testing.B) {
	benchSortedMap(b, marshalIndentEncodingJson, func(v any) ([]byte, error) {
		return gojson.MarshalIndent(v, "", "  ")
	})
}

func Benchmark_Encode_SortedMapIndent_GoJsonV2(b *testing.B) {
	benchSortedMap(b, marshalIndentEncodingJson, func(v any) ([]byte, error) {
		return gojsonv2.Marshal(v, gojsonv2.Deterministic(true), gojsontext.Multiline(true), gojsontext.WithIndent("  "))
	})
}

func Benchmark_Encode_SortedMapIndent_EncodingJson(b *testing.B) {
	benchSortedMap(b, marshalIndentEncodingJson, marshalIndentEncodingJson)
}

func Benchmark_Encode_SortedMapIndent_SonicStd(b *testing.B) {
	pretouchSonic()
	benchSortedMap(b, marshalIndentEncodingJson, func(v any) ([]byte, error) {
		return sonic.ConfigStd.MarshalIndent(v, "", "  ")
	})
}

package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/goccy/go-json"
)

// The entries of a map with keys of a string kind are read by ranging over the map as a map of the same
// layout, by the size of the value; a map with keys of another kind is read by reflect.MapIter. Every size of
// a value around the words, values with pointers, values too large to be in the map, keys of every kind, and
// maps large enough to have grown, must encode as encoding/json does.

type mapValueTwoWords struct {
	A int
	B string
}

type mapValue12 struct {
	A, B, C int32
}

type mapValue136 struct {
	A [17]int64
}

type mapValueWithPointers struct {
	P *int
	S []string
	M map[string]int
	I interface{}
}

type mapTextKey struct{ V int }

func (k mapTextKey) MarshalText() ([]byte, error) { return []byte(fmt.Sprintf("k%d", k.V)), nil }

func mapOf[K comparable, V any](n int, key func(int) K, value func(int) V) map[K]V {
	m := make(map[K]V, 0)
	for i := 0; i < n; i++ {
		m[key(i)] = value(i)
	}
	return m
}

func TestEncodeMapLayouts(t *testing.T) {
	one := 1
	stringKey := func(i int) string { return "k" + strconv.Itoa(i*7919%1000) }
	values := []interface{}{}
	for _, n := range []int{1, 2, 8, 9, 50, 1000} {
		values = append(values,
			mapOf(n, stringKey, func(i int) struct{} { return struct{}{} }),
			mapOf(n, stringKey, func(i int) bool { return i%2 == 0 }),
			mapOf(n, stringKey, func(i int) int8 { return int8(i) }),
			mapOf(n, stringKey, func(i int) int32 { return int32(-i) }),
			mapOf(n, stringKey, func(i int) int { return i }),
			mapOf(n, stringKey, func(i int) float64 { return float64(i) + 0.5 }),
			mapOf(n, stringKey, func(i int) mapValue12 { return mapValue12{int32(i), 2, 3} }),
			mapOf(n, stringKey, func(i int) string { return strconv.Itoa(i) }),
			mapOf(n, stringKey, func(i int) interface{} { return []interface{}{i, "s", nil}[i%3] }),
			mapOf(n, stringKey, func(i int) mapValueTwoWords { return mapValueTwoWords{i, "s"} }),
			mapOf(n, stringKey, func(i int) [3]int { return [3]int{i, i, i} }),
			mapOf(n, stringKey, func(i int) mapValueWithPointers {
				return mapValueWithPointers{P: &one, S: []string{"a"}, M: map[string]int{"x": i}, I: i}
			}),
			mapOf(n, stringKey, func(i int) [16]int64 { return [16]int64{15: int64(i)} }),
			mapOf(n, stringKey, func(i int) mapValue136 { return mapValue136{[17]int64{16: int64(i)}} }),
			mapOf(n, stringKey, func(i int) *int { return &one }),
			mapOf(n, stringKey, func(i int) map[string]int { return map[string]int{"n": i} }),
			mapOf(n, func(i int) mapKeyString { return mapKeyString(stringKey(i)) }, func(i int) int { return i }),
			mapOf(n, func(i int) int { return i - n/2 }, func(i int) string { return "v" }),
			mapOf(n, func(i int) int64 { return int64(i) * 1000000000 }, func(i int) int { return i }),
			mapOf(n, func(i int) uint8 { return uint8(i) }, func(i int) interface{} { return i }),
			mapOf(n, func(i int) mapTextKey { return mapTextKey{i} }, func(i int) mapValueTwoWords { return mapValueTwoWords{i, "s"} }),
		)
	}
	values = append(values,
		map[string]int{},
		map[string]int(nil),
		struct {
			M map[string]int `json:"m,omitempty"`
			N map[string]int `json:"n,omitempty"`
			E map[int]int    `json:"e,omitempty"`
		}{M: map[string]int{"a": 1}, N: map[string]int{}, E: map[int]int{}},
		[]map[string]interface{}{{"a": map[string]interface{}{"b": []interface{}{map[string]int{"c": 1}}}}},
	)
	for _, v := range values {
		expected, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T:\n got %.300s\nwant %.300s", v, got, expected)
		}
		expected, _ = stdjson.MarshalIndent(v, "", "  ")
		got, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T with indent:\n got %.300s\nwant %.300s", v, got, expected)
		}
	}
}

// the entries of an unordered map are indented as the entries of a sorted one.
func TestEncodeMapUnorderedIndent(t *testing.T) {
	v := map[string]map[string]int{"a": {"x": 1}}
	expected, _ := stdjson.MarshalIndent(v, "", "  ")
	got, err := json.MarshalIndentWithOption(v, "", "  ", json.UnorderedMap())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(expected) {
		t.Fatalf("\n got %s\nwant %s", got, expected)
	}
}

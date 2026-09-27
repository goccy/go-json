package json_test

import (
	stdjson "encoding/json"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

// The entries of a sorted map whose values are scalars are written by one call of the VM, up to an entry whose
// value is not a scalar, which its opcodes write. The values of every kind in any order, with indent and
// colors, must encode as encoding/json does.
func TestEncodeSortedMapScalarRuns(t *testing.T) {
	one := 1
	var nilInt *int
	values := []any{
		1, -2, uint8(3), 1.5, float32(2.5), "s", "<&>", true, false, nil, &one, nilInt,
		[]int{1}, []any{}, map[string]any{"z": 1, "a": "b"}, struct{ A int }{1}, &struct{ B string }{"c"},
		time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC), stdjson.RawMessage(`{"r":1}`), []byte("bytes"),
	}
	var maps []any
	for n := 1; n <= 3*len(values); n++ {
		m := map[string]any{}
		for i := 0; i < n; i++ {
			m["k"+strconv.Itoa(i*7919%1000)] = values[(i*31+n)%len(values)]
		}
		maps = append(maps, m, []any{m, map[string]any{"in": m}})
	}
	maps = append(maps,
		map[string]int{"b": 2, "a": 1, "c": 3},
		map[string]string{"b": "x", "a": "<y>"},
		map[string]bool{"t": true, "f": false},
		map[string][]byte{"b": []byte("x"), "a": nil},
		map[string]*int{"a": &one, "b": nil},
		map[string]any{},
	)
	for _, v := range maps {
		want, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(v)
		if err != nil || string(got) != string(want) {
			t.Fatalf("%v:\n got %s %v\nwant %s", v, got, err, want)
		}
		want, _ = stdjson.MarshalIndent(v, "", "  ")
		got, err = json.MarshalIndent(v, "", "  ")
		if err != nil || string(got) != string(want) {
			t.Fatalf("%v with indent:\n got %s %v\nwant %s", v, got, err, want)
		}
		// the colored VMs, with a scheme of no colors, write what the others write.
		noColor := json.Colorize(&json.ColorScheme{})
		want, _ = stdjson.Marshal(v)
		got, err = json.MarshalWithOption(v, noColor)
		if err != nil || string(got) != string(want) {
			t.Fatalf("%v colored:\n got %s %v\nwant %s", v, got, err, want)
		}
		want, _ = stdjson.MarshalIndent(v, "", "  ")
		got, err = json.MarshalIndentWithOption(v, "", "  ", noColor)
		if err != nil || string(got) != string(want) {
			t.Fatalf("%v colored with indent:\n got %s %v\nwant %s", v, got, err, want)
		}
	}
	// an error of a scalar is returned.
	for _, v := range []any{map[string]any{"a": 1, "b": math.NaN()}, map[string]float64{"a": math.Inf(1)}} {
		if _, err := json.Marshal(v); err == nil {
			t.Fatalf("%v: expected an error", v)
		}
	}
}

//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjson "encoding/json/v2"
	"errors"
	"fmt"
	"math"
	"net/netip"
	"testing"
	"time"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The tests here compare the v2 json package with encoding/json/v2, which it replaces: the same value and the same
// options must give the same output, or an error of the same type at the same place.

type diffInner struct {
	S string  `json:"s,omitempty"`
	N int     `json:",string"`
	F float64 `json:"f,omitzero"`
	P *int    `json:"p,omitempty"`
}

type diffEmbedded struct {
	E1 string
	E2 int `json:"e2"`
}

type diffText string

func (t diffText) MarshalText() ([]byte, error) { return []byte("text:" + string(t)), nil }

type diffJSON struct{ V int }

func (j diffJSON) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf(`{"v": %d}`, j.V)), nil }

type diffAll struct {
	diffEmbedded
	*diffInner
	Name     string            `json:"name"`
	Tags     []string          `json:"tags"`
	Bytes    []byte            `json:"bytes"`
	Array    [3]byte           `json:"array"`
	Map      map[string]int    `json:"map"`
	IntMap   map[int]string    `json:"intMap"`
	TextMap  map[diffText]bool `json:"textMap"`
	Any      any               `json:"any"`
	Ptr      *diffInner        `json:"ptr"`
	Time     time.Time         `json:"time"`
	Text     diffText          `json:"text"`
	JSON     diffJSON          `json:"json"`
	Addr     netip.Addr        `json:"addr"`
	Empty    struct{}          `json:"empty,omitempty"`
	Skipped  int               `json:"-"`
	Floats   []float32         `json:"floats"`
	Nested   [][]any           `json:"nested"`
	Unicode  string            `json:"unicode"`
	HTML     string            `json:"html"`
	Interfac fmt.Stringer      `json:"stringer"`
}

func diffValues() []any {
	n := 7
	return []any{
		nil, true, false, 0, -1, int8(-128), uint64(math.MaxUint64), 3.5, float32(0.1), 1e21, 1e-7, "", "a\"b\\c",
		"<html>&</html>", "  ", "日本語", []byte(nil), []byte{}, []byte("bytes"), [4]byte{1, 2, 3, 4},
		[]int(nil), []int{}, []int{1, 2}, map[string]int(nil), map[string]int{"b": 2, "a": 1}, map[int]bool{3: true, -1: false},
		map[float64]string{1.5: "x"}, []any{nil, 1, "x", []any{}, map[string]any{}}, &n, (*int)(nil),
		diffAll{}, &diffAll{
			diffEmbedded: diffEmbedded{"e", 1}, diffInner: &diffInner{S: "s", N: 5, F: 0.5, P: &n},
			Name: "name", Tags: []string{"a", "b"}, Bytes: []byte{0xff}, Array: [3]byte{9, 8, 7},
			Map: map[string]int{"z": 26, "y": 25}, IntMap: map[int]string{2: "two", 1: "one"},
			TextMap: map[diffText]bool{"k": true}, Any: map[string]any{"x": []any{1.5, "s", nil, true}},
			Ptr: &diffInner{}, Time: time.Date(2026, 10, 4, 1, 2, 3, 4, time.UTC), Text: "t", JSON: diffJSON{3},
			Addr: netip.MustParseAddr("192.0.2.1"), Floats: []float32{1.5, 0.1, 1e-8}, Nested: [][]any{{1}, nil, {}},
			Unicode: "é\u0000\u001f", HTML: "<a>&</a>", Interfac: time.Second,
		},
		struct {
			A int `json:"a,omitempty"`
			B any `json:"b,omitempty"`
			C *struct {
				D []int `json:"d,omitempty"`
			} `json:"c,omitempty"`
		}{B: []any{}, C: &struct {
			D []int `json:"d,omitempty"`
		}{}},
		struct{ X chan int }{}, map[string]chan int{}, []func(){nil}, time.Duration(5),
		struct {
			X time.Duration `json:",string"`
		}{}, struct {
			S string `json:",string"`
		}{}, math.NaN(), "\xff", map[string]int{"\xff": 1},
	}
}

func diffOptions() [][]json.Options {
	return [][]json.Options{
		nil,
		{json.Deterministic(true)},
		{json.StringifyNumbers(true), json.Deterministic(true)},
		{json.FormatNilSliceAsNull(true), json.FormatNilMapAsNull(true), json.Deterministic(true)},
		{json.OmitZeroStructFields(true), json.Deterministic(true)},
	}
}

func stdOptions(opts []json.Options) []stdjson.Options {
	var out []stdjson.Options
	for _, o := range opts {
		for _, f := range []struct {
			ours func(bool) json.Options
			std  func(bool) stdjson.Options
		}{
			{json.Deterministic, stdjson.Deterministic},
			{json.StringifyNumbers, stdjson.StringifyNumbers},
			{json.FormatNilSliceAsNull, stdjson.FormatNilSliceAsNull},
			{json.FormatNilMapAsNull, stdjson.FormatNilMapAsNull},
			{json.OmitZeroStructFields, stdjson.OmitZeroStructFields},
		} {
			if v, ok := json.GetOption(o, f.ours); ok {
				out = append(out, f.std(v))
			}
		}
	}
	return out
}

func TestMarshalSameAsStd(t *testing.T) {
	for i, v := range diffValues() {
		for _, opts := range diffOptions() {
			got, err := json.Marshal(v, opts...)
			want, wantErr := stdjson.Marshal(v, stdOptions(opts)...)
			name := fmt.Sprintf("%d:%T", i, v)
			if (err == nil) != (wantErr == nil) {
				t.Errorf("%s: Marshal error = %v; std %v", name, err, wantErr)
				continue
			}
			if err != nil {
				var serr *json.SemanticError
				var stdSerr *stdjson.SemanticError
				if errors.As(err, &serr) != errors.As(wantErr, &stdSerr) {
					t.Errorf("%s: Marshal error = %v; std %v", name, err, wantErr)
				} else if serr != nil && (serr.ByteOffset != stdSerr.ByteOffset || string(serr.JSONPointer) != string(stdSerr.JSONPointer) || serr.GoType != stdSerr.GoType) {
					t.Errorf("%s: Marshal error = %#v; std %#v", name, serr, stdSerr)
				}
				continue
			}
			// the members of a map are not sorted without Deterministic: they are compared in their sorted order.
			if deterministic, _ := json.GetOption(json.JoinOptions(opts...), json.Deterministic); !deterministic {
				canonical := func(b []byte) []byte {
					v := jsontext.Value(b)
					if err := v.Canonicalize(); err != nil {
						t.Fatalf("%s: Canonicalize(%s): %v", name, b, err)
					}
					return v
				}
				got, want = canonical(got), canonical(want)
			}
			if string(got) != string(want) {
				t.Errorf("%s: Marshal:\ngot  %s\nwant %s", name, got, want)
			}
		}
	}
}

//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsonv1 "encoding/json"
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

// The members of omitempty: written, then unwritten if empty, or known not to be empty by their types.
type diffOmitInner struct {
	S *string `json:",omitempty"`
	A any     `json:",omitempty"`
	N int
}

type diffOmitAll struct {
	S *string `json:",omitempty"`
	A any     `json:",omitempty"`
}

type diffOmit struct {
	diffOmitInner
	LongNameOfTheMember string
	NilPointer          *bool          `json:",omitempty"`
	Bool                *bool          `json:",omitempty"`
	Int                 *int           `json:",omitempty"`
	IntPointer          **int          `json:",omitempty"`
	Inner               *diffOmitInner `json:",omitempty"`
	All                 *diffOmitAll   `json:",omitempty"`
	AllValue            diffOmitAll    `json:",omitempty"`
	InnerValue          diffOmitInner  `json:",omitempty"`
	Stamp               *diffStamp     `json:",omitempty"`
	StampValue          diffStamp      `json:",omitempty"`
	Quoted              *string        `json:",omitempty,string"`
	StringPointer       **string       `json:",omitempty"`
	Last                *string        `json:",omitempty"`
}

// diffDupKey is a key of a map whose name is the one of other keys: the keys of the same parity have the same name.
type diffDupKey int

func (k diffDupKey) MarshalText() ([]byte, error) {
	return []byte([]string{"even", "odd \u00e9"}[k&1]), nil
}

// diffEmptyJSON is a value whose MarshalJSON writes an empty value, null, which omitempty omits.
type diffEmptyJSON struct{}

func (diffEmptyJSON) MarshalJSON() ([]byte, error) { return []byte("null"), nil }

// diffOmitSkip has fields of omitempty of each kind of pointer, which are omitted for nil or for an empty value,
// each after a member whose value is "": an omitted field must not take the member before it with it, which ends
// with an empty value as an empty member of the field would.
type diffOmitSkip struct {
	E0   string
	Int  **int `json:",omitempty"`
	E1   string
	Bool **bool `json:",omitempty"`
	E2   string
	Flt  **float64 `json:",omitempty"`
	E3   string
	Str  **string `json:",omitempty"`
	E4   string
	Quo  **string `json:",omitempty,string"`
	E5   string
	Byt  *[]byte `json:",omitempty"`
	E6   string
	JSN  *diffEmptyJSON `json:",omitempty"`
	E7   string
	Txt  *diffText `json:",omitempty"`
	E8   string
	Any  any `json:",omitempty"`
	E9   string
}

// diffRaw has raw values of omitempty: nil is omitted without a call, an empty one is an error of its method.
type diffRaw struct {
	Nil   stdjsonv1.RawMessage `json:",omitempty"`
	Empty stdjsonv1.RawMessage `json:",omitempty"`
	Value stdjsonv1.RawMessage `json:",omitempty"`
	Null  stdjsonv1.RawMessage `json:",omitempty"`
}

// diffStamp has MarshalJSON of time.Time, by embedding it.
type diffStamp struct{ time.Time }

func diffValues() []any {
	n := 7
	empty, nonEmpty := "", "x"
	emptyPointer, nonEmptyPointer := &empty, &nonEmpty
	var nilInt *int
	f := false
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
		diffOmit{}, &diffOmit{
			diffOmitInner: diffOmitInner{S: &empty, A: ""}, Bool: &f, Int: &n, IntPointer: &nilInt,
			Inner: &diffOmitInner{S: &empty, A: []any{}}, All: &diffOmitAll{S: &empty, A: map[string]any{}},
			AllValue: diffOmitAll{A: nil}, Last: &empty,
			Stamp: &diffStamp{time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)}, Quoted: &empty, StringPointer: &emptyPointer,
		},
		&diffOmit{Last: &nonEmpty, StringPointer: &nonEmptyPointer, Quoted: &nonEmpty},
		diffOmitSkip{}, diffOmitSkip{
			Int: &nilInt, Bool: new(*bool), Flt: new(*float64), Str: new(*string), Quo: new(*string), Byt: new([]byte),
			JSN: &diffEmptyJSON{}, Txt: new(diffText), Any: "\"",
		}, diffOmitSkip{Str: &emptyPointer, Quo: &emptyPointer, Byt: &[]byte{}, Any: ""},
		// the first duplicate name is the same in any order of the entries ( see
		// TestMarshalSortedDuplicateNamesSameAsStd for the others ).
		map[diffDupKey]int{1: 1}, map[diffDupKey]int{1: 1, 2: 2}, map[diffDupKey]int{1: 1, 3: 3},
		struct{ M map[diffDupKey]int }{map[diffDupKey]int{1: 1, 3: 3}},
		diffRaw{Value: stdjsonv1.RawMessage(`{"a":1}`), Null: stdjsonv1.RawMessage(`null`)},
		diffRaw{Empty: stdjsonv1.RawMessage{}}, diffRaw{Value: stdjsonv1.RawMessage(`{}`)},
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
			compareWithStd(t, fmt.Sprintf("%d:%T", i, v), v, opts)
		}
	}
}

// The names of the keys of a sorted map which have the same name: the first one which an entry before it has is
// reported, where the entries are sorted, which the maps of more than one name of the same keys depend on.
func TestMarshalSortedDuplicateNamesSameAsStd(t *testing.T) {
	for i, v := range []any{
		map[diffDupKey]int{0: 0, 1: 1, 2: 2, 3: 3}, map[diffDupKey]int{1: 1, 2: 2, 4: 4},
		struct{ M map[diffDupKey]int }{map[diffDupKey]int{0: 0, 1: 1, 3: 3}},
	} {
		for _, opts := range diffOptions() {
			if deterministic, _ := json.GetOption(json.JoinOptions(opts...), json.Deterministic); deterministic {
				compareWithStd(t, fmt.Sprintf("%d:%T", i, v), v, opts)
			}
		}
	}
}

// compareWithStd fails if Marshal of v with opts doesn't write what encoding/json/v2 writes, or fails at another
// place.
func compareWithStd(t *testing.T, name string, v any, opts []json.Options) {
	t.Helper()
	got, err := json.Marshal(v, opts...)
	want, wantErr := stdjson.Marshal(v, stdOptions(opts)...)
	if (err == nil) != (wantErr == nil) {
		t.Errorf("%s: Marshal error = %v; std %v", name, err, wantErr)
		return
	}
	if err != nil {
		var serr *json.SemanticError
		var stdSerr *stdjson.SemanticError
		if errors.As(err, &serr) != errors.As(wantErr, &stdSerr) {
			t.Errorf("%s: Marshal error = %v; std %v", name, err, wantErr)
		} else if serr != nil && (serr.ByteOffset != stdSerr.ByteOffset || string(serr.JSONPointer) != string(stdSerr.JSONPointer) || serr.GoType != stdSerr.GoType) {
			t.Errorf("%s: Marshal error = %#v; std %#v", name, serr, stdSerr)
		}
		return
	}
	// the members of a map are not sorted without Deterministic: they are compared in their sorted order.
	if deterministic, _ := json.GetOption(json.JoinOptions(opts...), json.Deterministic); !deterministic {
		canonical := func(b []byte) ([]byte, bool) {
			v := jsontext.Value(b)
			if err := v.Canonicalize(); err != nil {
				t.Errorf("%s: Canonicalize(%s): %v", name, b, err)
				return nil, false
			}
			return v, true
		}
		var gotOK, wantOK bool
		if got, gotOK = canonical(got); !gotOK {
			return
		}
		if want, wantOK = canonical(want); !wantOK {
			return
		}
	}
	if string(got) != string(want) {
		t.Errorf("%s: Marshal:\ngot  %s\nwant %s", name, got, want)
	}
}

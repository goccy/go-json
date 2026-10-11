package json_test

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

// The entries of a sorted map whose values are scalars or of interface{} are written directly by one call of the
// VM, up to an entry of interface{} whose value is not a scalar, which its opcodes write. The values of every kind in any order, with indent and
// colors, must encode as encoding/json does.
func TestEncodeSortedMapDirectEntries(t *testing.T) {
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

// sortedMapFailingValue fails to be marshaled when it is odd.
type sortedMapFailingValue int

func (v sortedMapFailingValue) MarshalJSON() ([]byte, error) {
	if v%2 != 0 {
		return nil, fmt.Errorf("odd value %d", int(v))
	}
	return fmt.Appendf(nil, "%d", int(v)), nil
}

type sortedMapTextKey int

func (k sortedMapTextKey) MarshalText() ([]byte, error) { return fmt.Appendf(nil, "t%03d", int(k)), nil }

// The values of a map whose keys are not strings are encoded in the order of the keys, as encoding/json encodes
// them: the error is the one of the first value which fails in that order. The error of the method is compared, as
// the text of the error which wraps it differs between the versions of encoding/json.
func TestEncodeSortedMapValueError(t *testing.T) {
	ints := map[int]sortedMapFailingValue{}
	texts := map[sortedMapTextKey]sortedMapFailingValue{}
	for i := range 40 {
		ints[i] = sortedMapFailingValue(i * 7 % 40)
		texts[sortedMapTextKey(i)] = sortedMapFailingValue(i * 7 % 40)
	}
	for _, v := range []any{ints, texts, struct{ A, B any }{0, ints}} {
		_, wantErr := stdjson.Marshal(v)
		want := errors.Unwrap(wantErr)
		_, err := json.Marshal(v)
		if got := errors.Unwrap(err); got == nil || want == nil || got.Error() != want.Error() {
			t.Errorf("%T: Marshal error = %v; want %v", v, err, wantErr)
		}
		_, err = json.MarshalIndent(v, "", " ")
		if got := errors.Unwrap(err); got == nil || got.Error() != want.Error() {
			t.Errorf("%T: MarshalIndent error = %v; want %v", v, err, wantErr)
		}
	}
}

// The entries of a sorted map whose keys are not strings, which are written after every key is encoded, are written
// as encoding/json writes them, by every VM: with indent and colors.
func TestEncodeSortedMapNonStringKeys(t *testing.T) {
	nested := map[int]any{}
	texts := map[sortedMapTextKey]any{}
	for i := range 20 {
		nested[i*7%20-10] = map[uint]any{uint(i): []any{i, "s"}, uint(i + 100): map[sortedMapTextKey]int{1: i}}
		texts[sortedMapTextKey(i)] = struct{ A, B int }{i, -i}
	}
	noColor := json.Colorize(&json.ColorScheme{})
	for _, v := range []any{nested, texts, map[int]int{1: 1}, struct{ M map[int8]string }{map[int8]string{15: "a", -2: "b", 3: "c"}}} {
		want, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		wantIndent, _ := stdjson.MarshalIndent(v, "", "  ")
		got, err := json.Marshal(v)
		if err != nil || string(got) != string(want) {
			t.Errorf("%T:\n got %s %v\nwant %s", v, got, err, want)
		}
		got, err = json.MarshalIndent(v, "", "  ")
		if err != nil || string(got) != string(wantIndent) {
			t.Errorf("%T with indent:\n got %s %v\nwant %s", v, got, err, wantIndent)
		}
		got, err = json.MarshalWithOption(v, noColor)
		if err != nil || string(got) != string(want) {
			t.Errorf("%T colored:\n got %s %v\nwant %s", v, got, err, want)
		}
		got, err = json.MarshalIndentWithOption(v, "", "  ", noColor)
		if err != nil || string(got) != string(wantIndent) {
			t.Errorf("%T colored with indent:\n got %s %v\nwant %s", v, got, err, wantIndent)
		}
	}
}

// sortedMapNameKey is a key named by its text, which may need escapes.
type sortedMapNameKey string

func (k sortedMapNameKey) MarshalText() ([]byte, error) { return []byte(k), nil }

// The keys named by texts are sorted by their names, as encoding/json sorts them, whatever the names have: escapes,
// or a space or a '!', which come before the quote that ends a name.
func TestEncodeSortedMapTextKeyNames(t *testing.T) {
	runes := []rune("ab5 !\"\\<\x01é")
	r := rand.New(rand.NewPCG(1, 2))
	noColor := json.Colorize(&json.ColorScheme{})
	for range 500 {
		m := map[sortedMapNameKey]int{}
		for i := range 1 + r.IntN(20) {
			var name []rune
			for range r.IntN(4) {
				name = append(name, runes[r.IntN(len(runes))])
			}
			m[sortedMapNameKey(name)] = i
		}
		want, err := stdjson.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(m)
		if err != nil || string(got) != string(want) {
			t.Fatalf("got %s %v\nwant %s", got, err, want)
		}
		got, err = json.MarshalWithOption(m, noColor)
		if err != nil || string(got) != string(want) {
			t.Fatalf("colored: got %s %v\nwant %s", got, err, want)
		}
		want, _ = stdjson.MarshalIndent(m, "", " ")
		got, err = json.MarshalIndent(m, "", " ")
		if err != nil || string(got) != string(want) {
			t.Fatalf("with indent: got %s %v\nwant %s", got, err, want)
		}
	}
}

// The keys are sorted by their names with any colors: an integer is colored within its quotes, which the codes of
// its color don't put out of the order of its name.
func TestEncodeSortedMapColoredKeys(t *testing.T) {
	format := func(name string) json.ColorFormat { return json.ColorFormat{Header: "<" + name + ">", Footer: "</" + name + ">"} }
	scheme := &json.ColorScheme{Int: format("i"), Uint: format("u"), String: format("s"), ObjectKey: format("k")}
	tags := regexp.MustCompile(`</?[a-z]>`)
	for _, v := range []any{
		map[int]int{1: 1, 10: 2, 100: 3, 2: 4, -1: 5, -10: 6},
		map[uint8]int{1: 1, 10: 2, 100: 3, 2: 4},
		map[sortedMapNameKey]int{"a": 1, "a!": 2, "a ": 3, "": 4, "b": 5},
	} {
		want, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.MarshalWithOption(v, json.Colorize(scheme))
		if err != nil || string(tags.ReplaceAll(got, nil)) != string(want) {
			t.Errorf("%T: got %s %v\nwant %s", v, got, err, want)
		}
		want, _ = stdjson.MarshalIndent(v, "", " ")
		got, err = json.MarshalIndentWithOption(v, "", " ", json.Colorize(scheme))
		if err != nil || string(tags.ReplaceAll(got, nil)) != string(want) {
			t.Errorf("%T with indent: got %s %v\nwant %s", v, got, err, want)
		}
	}
}

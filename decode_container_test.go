package json_test

import (
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// The decoders of arrays, slices, maps and interface{} allocate and assign through the reflect package.
// These tests fix what they do with the value which is decoded into.

func TestDecodeArrayZeroesTheRest(t *testing.T) {
	// A shorter JSON array zeroes the rest of the array, and nothing after it.
	var v struct {
		A [4]uint8
		B uint32
	}
	v.A = [4]uint8{9, 9, 9, 9}
	v.B = 0xdeadbeef
	if err := json.Unmarshal([]byte(`{"A":[1]}`), &v); err != nil {
		t.Fatal(err)
	}
	if v.A != [4]uint8{1, 0, 0, 0} || v.B != 0xdeadbeef {
		t.Fatalf("got %v %x", v.A, v.B)
	}
	type pair struct{ X, Y int }
	w := [2]pair{{1, 7}, {2, 7}}
	if err := json.Unmarshal([]byte(`[]`), &w); err != nil {
		t.Fatal(err)
	}
	if w != [2]pair{} {
		t.Fatalf("got %v", w)
	}
	s := [3]*int{new(int), new(int), new(int)}
	if err := json.Unmarshal([]byte(`[1]`), &s); err != nil {
		t.Fatal(err)
	}
	if *s[0] != 1 || s[1] != nil || s[2] != nil {
		t.Fatalf("got %v", s)
	}
}

func TestDecodeSliceIntoExisting(t *testing.T) {
	type elem struct{ A, B int }
	// The elements of the slice are decoded into its existing elements.
	v := []elem{{A: 1, B: 1}, {A: 2, B: 2}}
	if err := json.Unmarshal([]byte(`[{"A":10},{"B":20},{"A":30}]`), &v); err != nil {
		t.Fatal(err)
	}
	if want := []elem{{10, 1}, {2, 20}, {30, 0}}; !reflect.DeepEqual(v, want) {
		t.Fatalf("got %v, want %v", v, want)
	}
	// The array of the slice is reused when it is large enough; the elements after
	// its length are decoded into zero values.
	backing := make([]elem, 4)
	backing[1] = elem{A: 5, B: 5}
	v = backing[:1]
	if err := json.Unmarshal([]byte(`[{"A":1},{"A":2}]`), &v); err != nil {
		t.Fatal(err)
	}
	if want := []elem{{1, 0}, {2, 0}}; !reflect.DeepEqual(v, want) {
		t.Fatalf("got %v, want %v", v, want)
	}
	if &v[0] != &backing[0] {
		t.Fatal("the array of the slice is not reused")
	}
	// A shorter array truncates.
	v = []elem{{1, 1}, {2, 2}, {3, 3}}
	if err := json.Unmarshal([]byte(`[{"A":9}]`), &v); err != nil {
		t.Fatal(err)
	}
	if want := []elem{{9, 1}}; !reflect.DeepEqual(v, want) {
		t.Fatalf("got %v, want %v", v, want)
	}
	// An empty array is an empty slice, not nil; null is nil.
	var e []int
	if err := json.Unmarshal([]byte(`[]`), &e); err != nil {
		t.Fatal(err)
	}
	if e == nil || len(e) != 0 {
		t.Fatalf("got %#v", e)
	}
	if err := json.Unmarshal([]byte(`null`), &e); err != nil {
		t.Fatal(err)
	}
	if e != nil {
		t.Fatalf("got %#v", e)
	}
}

func TestDecodeSliceOfPointersIsNotShared(t *testing.T) {
	// Every element of a new slice gets a pointer of its own, also when the decoder reuses its buffers.
	for i := 0; i < 3; i++ {
		var v []*struct{ A int }
		if err := json.Unmarshal([]byte(`[{"A":1},{"A":2},{"A":3},{"A":4},{"A":5}]`), &v); err != nil {
			t.Fatal(err)
		}
		for j, e := range v {
			if e.A != j+1 {
				t.Fatalf("element %d: got %d", j, e.A)
			}
			for k := 0; k < j; k++ {
				if v[k] == e {
					t.Fatalf("elements %d and %d share a pointer", k, j)
				}
			}
		}
	}
}

func TestDecodeSliceAfterError(t *testing.T) {
	// An error in the middle of an array leaves nothing behind for the next decode.
	type elem struct {
		P *int
		S string
	}
	var v []elem
	if err := json.Unmarshal([]byte(`[{"P":1,"S":"a"},{"P":"x"}]`), &v); err == nil {
		t.Fatal("expected an error")
	}
	v = nil
	if err := json.Unmarshal([]byte(`[{"S":"b"},{"S":"c"}]`), &v); err != nil {
		t.Fatal(err)
	}
	if len(v) != 2 || v[0].P != nil || v[1].P != nil || v[0].S != "b" || v[1].S != "c" {
		t.Fatalf("got %+v", v)
	}
}

func TestDecodeMapValuesAreNotShared(t *testing.T) {
	// Every value of a map is decoded into a zero value of its own.
	var m map[string]*struct{ A, B int }
	if err := json.Unmarshal([]byte(`{"x":{"A":1},"y":{"B":2}}`), &m); err != nil {
		t.Fatal(err)
	}
	if m["x"] == m["y"] || m["x"].A != 1 || m["x"].B != 0 || m["y"].A != 0 || m["y"].B != 2 {
		t.Fatalf("got x=%+v y=%+v", m["x"], m["y"])
	}
	var s map[string]struct{ A, B int }
	if err := json.Unmarshal([]byte(`{"x":{"A":1},"y":{"B":2}}`), &s); err != nil {
		t.Fatal(err)
	}
	if s["x"].B != 0 || s["y"].A != 0 {
		t.Fatalf("got %+v", s)
	}
	var sl map[string][]int
	if err := json.Unmarshal([]byte(`{"x":[1,2,3],"y":[4]}`), &sl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sl, map[string][]int{"x": {1, 2, 3}, "y": {4}}) {
		t.Fatalf("got %v", sl)
	}
	// A map which exists is added to.
	im := map[int]string{1: "a"}
	if err := json.Unmarshal([]byte(`{"2":"b"}`), &im); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(im, map[int]string{1: "a", 2: "b"}) {
		t.Fatalf("got %v", im)
	}
}

func TestDecodeInterfaceNested(t *testing.T) {
	// The values nested in interface{} are decoded through a shared slot and stack: each of them
	// must end up where it belongs.
	in := `{"a":[1,"s",[true,null,{"b":[2,3]}],{}],"c":{"d":{"e":[[],[[4]]]}},"f":"g","h":1.5}`
	var got any
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"a": []any{1.0, "s", []any{true, nil, map[string]any{"b": []any{2.0, 3.0}}}, map[string]any{}},
		"c": map[string]any{"d": map[string]any{"e": []any{[]any{}, []any{[]any{4.0}}}}},
		"f": "g",
		"h": 1.5,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
	// The same, through a field of map[string]interface{} and of []interface{}.
	var typed struct {
		M map[string]any `json:"m"`
		S []any          `json:"s"`
	}
	typed.M = map[string]any{"old": true}
	if err := json.Unmarshal([]byte(`{"m":`+in+`,"s":[`+in+`,2]}`), &typed); err != nil {
		t.Fatal(err)
	}
	wantM := map[string]any{"old": true}
	for k, v := range want {
		wantM[k] = v
	}
	if !reflect.DeepEqual(typed.M, wantM) {
		t.Fatalf("got %#v", typed.M)
	}
	if !reflect.DeepEqual(typed.S, []any{want, 2.0}) {
		t.Fatalf("got %#v", typed.S)
	}
}

func TestDecodeInterfaceManyValues(t *testing.T) {
	// More numbers and strings than a slab holds, over several calls: every value keeps its own.
	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < 100; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"n":`)
		b.WriteString(strings.Repeat("1", i%9+1))
		b.WriteString(`,"s":"`)
		b.WriteString(strings.Repeat("x", i))
		b.WriteString(`"}`)
	}
	b.WriteString("]")
	var results [][]any
	for call := 0; call < 3; call++ {
		var v []any
		if err := json.Unmarshal([]byte(b.String()), &v); err != nil {
			t.Fatal(err)
		}
		results = append(results, v)
	}
	for _, v := range results {
		for i, e := range v {
			m := e.(map[string]any)
			var n float64
			for j := 0; j < i%9+1; j++ {
				n = n*10 + 1
			}
			if m["n"] != n || m["s"] != strings.Repeat("x", i) {
				t.Fatalf("element %d: got %v", i, m)
			}
		}
	}
}

func TestDecodeInterfaceAfterError(t *testing.T) {
	// An error in the middle of nested arrays leaves nothing on the stack for the next decode.
	var v any
	if err := json.Unmarshal([]byte(`[1,[2,[3,x]]]`), &v); err == nil {
		t.Fatal("expected an error")
	}
	if err := json.Unmarshal([]byte(`[[1],[2,3]]`), &v); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, []any{[]any{1.0}, []any{2.0, 3.0}}) {
		t.Fatalf("got %#v", v)
	}
	if err := json.Unmarshal([]byte(`{null:1}`), &v); err == nil {
		t.Fatal("expected an error for null as a key")
	}
}

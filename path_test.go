package json_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

func TestExtractPath(t *testing.T) {
	src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
	t.Run("$.a.b", func(t *testing.T) {
		path, err := json.CreatePath("$.a.b")
		if err != nil {
			t.Fatal(err)
		}
		contents, err := path.Extract(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) != 1 {
			t.Fatal("failed to extract")
		}
		if !bytes.Equal(contents[0], []byte("10")) {
			t.Fatal("failed to extract")
		}
	})
	t.Run("$.b", func(t *testing.T) {
		path, err := json.CreatePath("$.b")
		if err != nil {
			t.Fatal(err)
		}
		contents, err := path.Extract(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) != 1 {
			t.Fatal("failed to extract")
		}
		if !bytes.Equal(contents[0], []byte(`"text"`)) {
			t.Fatal("failed to extract")
		}
	})
	t.Run("$.a", func(t *testing.T) {
		path, err := json.CreatePath("$.a")
		if err != nil {
			t.Fatal(err)
		}
		contents, err := path.Extract(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) != 1 {
			t.Fatal("failed to extract")
		}
		if !bytes.Equal(contents[0], []byte(`{"b":10,"c":true}`)) {
			t.Fatal("failed to extract")
		}
	})
}

func TestUnmarshalPath(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
		t.Run("success", func(t *testing.T) {
			path, err := json.CreatePath("$.a.b")
			if err != nil {
				t.Fatal(err)
			}
			var v int
			if err := path.Unmarshal(src, &v); err != nil {
				t.Fatal(err)
			}
			if v != 10 {
				t.Fatal("failed to unmarshal path")
			}
		})
		t.Run("failure", func(t *testing.T) {
			path, err := json.CreatePath("$.a.c")
			if err != nil {
				t.Fatal(err)
			}
			var v map[string]any
			if err := path.Unmarshal(src, &v); err == nil {
				t.Fatal("expected error")
			}
		})
	})
	t.Run("bool", func(t *testing.T) {
		src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
		t.Run("success", func(t *testing.T) {
			path, err := json.CreatePath("$.a.c")
			if err != nil {
				t.Fatal(err)
			}
			var v bool
			if err := path.Unmarshal(src, &v); err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Fatal("failed to unmarshal path")
			}
		})
		t.Run("failure", func(t *testing.T) {
			path, err := json.CreatePath("$.a.b")
			if err != nil {
				t.Fatal(err)
			}
			var v bool
			if err := path.Unmarshal(src, &v); err == nil {
				t.Fatal("expected error")
			}
		})
	})
	t.Run("map", func(t *testing.T) {
		src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
		t.Run("success", func(t *testing.T) {
			path, err := json.CreatePath("$.a")
			if err != nil {
				t.Fatal(err)
			}
			var v map[string]any
			if err := path.Unmarshal(src, &v); err != nil {
				t.Fatal(err)
			}
			if len(v) != 2 {
				t.Fatal("failed to decode map")
			}
		})
	})

	t.Run("path with single quote selector", func(t *testing.T) {
		path, err := json.CreatePath("$['a.b'].c")
		if err != nil {
			t.Fatal(err)
		}

		var v string
		if err := path.Unmarshal([]byte(`{"a.b": {"c": "world"}}`), &v); err != nil {
			t.Fatal(err)
		}
		if v != "world" {
			t.Fatal("failed to unmarshal path")
		}
	})
	t.Run("path with double quote selector", func(t *testing.T) {
		path, err := json.CreatePath(`$."a.b".c`)
		if err != nil {
			t.Fatal(err)
		}

		var v string
		if err := path.Unmarshal([]byte(`{"a.b": {"c": "world"}}`), &v); err != nil {
			t.Fatal(err)
		}
		if v != "world" {
			t.Fatal("failed to unmarshal path")
		}
	})
}

func TestGetPath(t *testing.T) {
	t.Run("selector", func(t *testing.T) {
		var v any
		if err := json.Unmarshal([]byte(`{"a":{"b":10,"c":true},"b":"text"}`), &v); err != nil {
			t.Fatal(err)
		}
		path, err := json.CreatePath("$.a.b")
		if err != nil {
			t.Fatal(err)
		}
		var b int
		if err := path.Get(v, &b); err != nil {
			t.Fatal(err)
		}
		if b != 10 {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("index", func(t *testing.T) {
		var v any
		if err := json.Unmarshal([]byte(`{"a":[{"b":10,"c":true},{"b":"text"}]}`), &v); err != nil {
			t.Fatal(err)
		}
		path, err := json.CreatePath("$.a[0].b")
		if err != nil {
			t.Fatal(err)
		}
		var b int
		if err := path.Get(v, &b); err != nil {
			t.Fatal(err)
		}
		if b != 10 {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("indexAll", func(t *testing.T) {
		var v any
		if err := json.Unmarshal([]byte(`{"a":[{"b":1,"c":true},{"b":2},{"b":3}]}`), &v); err != nil {
			t.Fatal(err)
		}
		path, err := json.CreatePath("$.a[*].b")
		if err != nil {
			t.Fatal(err)
		}
		var b []int
		if err := path.Get(v, &b); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(b, []int{1, 2, 3}) {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("recursive", func(t *testing.T) {
		var v any
		if err := json.Unmarshal([]byte(`{"a":[{"b":1,"c":true},{"b":2},{"b":3}],"a2":{"b":4}}`), &v); err != nil {
			t.Fatal(err)
		}
		path, err := json.CreatePath("$..b")
		if err != nil {
			t.Fatal(err)
		}
		var b []int
		if err := path.Get(v, &b); err != nil {
			t.Fatal(err)
		}
		sort.Ints(b)
		if !reflect.DeepEqual(b, []int{1, 2, 3, 4}) {
			t.Fatalf("failed to decode by json.Get")
		}
	})
}

// A key of an object is a string: any other key is a syntax error for a path too, wherever the object is.
func TestPathKeyNotString(t *testing.T) {
	for _, src := range []string{`{null:1,"a":2}`, `{"a":2,null:3}`, `{5:1,"a":2}`, `{"a":2, true:3}`, `{"a":2,}`, `{"a":{null:7}}`, `{"b":{5:1},"a":2}`} {
		for _, p := range []string{"$.a", "$.a.b", "$..a"} {
			t.Run(p+src, func(t *testing.T) {
				path, err := json.CreatePath(p)
				if err != nil {
					t.Fatal(err)
				}
				var syntaxErr *json.SyntaxError
				if _, err := path.Extract([]byte(src)); !errors.As(err, &syntaxErr) {
					t.Errorf("Extract: got error %v, want a syntax error", err)
				}
				var v any
				if err := path.Unmarshal([]byte(src), &v); !errors.As(err, &syntaxErr) {
					t.Errorf("Unmarshal: got error %v, want a syntax error", err)
				}
			})
		}
	}
}

type PathInner struct {
	B int
	C *PathInner
}

type pathStruct struct {
	A   int
	P   *PathInner
	I   any
	Sl  []*PathInner
	M   map[string]*PathInner
	Arr [2]int
}

type pathHidden struct {
	A int
	b int
	S int `json:"-"`
}

type pathTagged struct {
	A int    `json:"a"`
	Y string `json:"b,omitempty"`
}

type pathHiddenInner struct {
	B int
}

type pathEmbedded struct {
	PathInner
	D int
}

type pathEmbeddedPtr struct {
	*PathInner
	D int
}

type pathEmbeddedHidden struct {
	*pathHiddenInner
	D int
}

type pathNode struct {
	V    int
	Next *pathNode
}

type pathInt int

type pathRecMap map[string]pathRecMap

type pathPtr *PathInner

type pathSelfPtr *pathSelfPtr

type pathPtrA *pathPtrB

type pathPtrB *pathPtrA

type pathDeep []pathDeep

type pathPtrMap map[string]*pathPtrMap

type pathPtrSlice []*pathPtrSlice

type PathEmbeddedBase struct {
	Next *pathEmbeddedNode `json:",omitempty"`
}

type pathEmbeddedNode struct {
	PathEmbeddedBase
	V int
}

type PathTaggedBase struct {
	Next *pathTaggedNode
}

type pathTaggedNode struct {
	PathTaggedBase `json:"b"`
	V              int
}

type pathKey string

var pathTestPaths = []string{
	`$.a`, `$.a.b`, `$.a[0]`, `$.a[1]`, `$.a[*]`, `$.a[-1]`, `$..b`, `$[0]`, `$[1]`, `$[-1]`, `$[*]`, `$[0].b`, `$[*].b`,
	`$..b.c`, `$..b[0]`, `$..b[*]`, `$.b`, `$.a.b.c`, `$..a..b`, `$[*][*]`, `$[*][0]`, `$[*]..b`, `$..a.b`, `$.a..b`,
	`$..c`, `$.a[*].b`, `$[*].b.c`, `$..b[0].c`, `$..b[*].c`, `$['a'].b`, `$.A`, `$.A.B`, `$..A`, `$..B`, `$.P.B`,
	`$.P.C.B`, `$.I.b`, `$.Sl[*].B`, `$.M.x.B`, `$.Arr[*]`, `$.S`, `$.PathInner.B`, `$..V`, `$..Next.V`, `$.x.x`,
}

var pathTestTexts = []string{
	`null`, `1`, `"s"`, `true`, `[]`, `{}`,
	`{"a":null}`, `{"a":1}`, `{"a":"s"}`, `{"a":[]}`, `{"a":{}}`,
	`{"a":{"b":null}}`, `{"a":{"b":1}}`, `{"b":{"c":1}}`, `{"b":{"b":1}}`, `{"b":[1,2]}`, `{"b":null}`,
	`{"a":[null,1]}`, `{"a":[{"b":1},{"b":null}]}`, `{"a":[[1],[null]]}`, `{"a":[{"b":[1,2]},{"b":3}]}`,
	`[null]`, `[1,null]`, `[[1],[null]]`, `[{"b":null}]`, `[{"b":{"c":2}}]`, `[{"b":1},{"c":{"b":2}}]`,
	`{"a":{"b":{"c":null}}}`, `{"x":{"b":[null]}}`, `{"b":"s","c":{"b":null}}`, `{"a":{"b":[1,{"b":2}]}}`,
	`{"b":{"c":[{"b":{"c":5}}]}}`, `{"a":{"a":{"b":1}},"b":{"a":{"b":2}}}`, `[[{"b":1}],[[{"b":2}]]]`,
	`[{"b":1},[2]]`, `{"a":[{"b":1},[2],3]}`, `[[1],{"b":2},[[3]]]`, `{"b":[[1],{"c":2}],"c":{"b":{"c":3}}}`,
	`[{"b":{"c":1}},{"b":[{"c":2}]}]`, `{"a":[[{"b":1}],{"b":2}]}`,
}

// pathTestSources are Go values with nil pointers, maps and interfaces, structs with fields a path does not select,
// embedded structs, named types and values which hold themselves.
func pathTestSources() []any {
	one := 1
	cyclicNode := &pathNode{V: 1}
	cyclicNode.Next = cyclicNode
	cyclicMap := map[string]any{"b": 1}
	cyclicMap["x"] = cyclicMap
	cyclicSlice := []any{1, nil}
	cyclicSlice[1] = cyclicSlice
	var selfInterface any
	selfInterface = &selfInterface
	var selfPtr pathSelfPtr
	selfPtr = &selfPtr
	return []any{
		nil, (*pathStruct)(nil), pathStruct{A: 1}, &pathStruct{},
		&pathStruct{P: &PathInner{B: 2}, I: map[string]any{"b": nil}, Sl: []*PathInner{nil, {B: 3}}, M: map[string]*PathInner{"x": nil}},
		map[string]*pathStruct{"x": nil}, map[string]any{"a": nil}, map[string]*int{"a": nil, "b": &one}, []*pathStruct{nil},
		[]any{nil, map[string]any{"b": nil}}, pathHidden{A: 1, b: 2, S: 3}, &pathHidden{}, pathTagged{A: 1, Y: "y"},
		pathEmbedded{PathInner: PathInner{B: 1}}, pathEmbeddedPtr{D: 1}, pathEmbeddedPtr{PathInner: &PathInner{B: 1}},
		pathEmbeddedHidden{D: 1}, map[int]int{1: 1}, map[pathInt]int{1: 1}, [2]int{1, 2}, map[string]chan int{"a": nil},
		map[string]func(){"a": nil}, &one, (*int)(nil), []int(nil), map[string]any(nil), map[string]pathPtr{"a": nil},
		map[string]any{"a": map[string]any{"B": 1, "C": nil, "D": 2, "PathInner": map[string]any{"B": 1}, "b": "x"}},
		map[string]any{"a": []any{1, "2", nil, true, 1.5}}, map[string]any{"a": &PathInner{B: 1}},
		map[string]pathInt{"a": 1}, cyclicNode, cyclicMap, cyclicSlice, selfInterface, selfPtr,
		map[any]any{"a": 3, 1: 2, 2.5: 1, true: 0, nil: 4}, map[pathKey]int{"a": 1, "b": 2},
	}
}

var pathTestDestinations = []func() any{
	func() any { return new(any) }, func() any { return new(int) }, func() any { return new(int8) },
	func() any { return new(uint) }, func() any { return new(string) }, func() any { return new(float32) },
	func() any { return new(bool) }, func() any { return new([]int) }, func() any { return new([]any) },
	func() any { return new([2]int) }, func() any { return new(map[string]any) }, func() any { return new(map[string]int) },
	func() any { return new(pathStruct) }, func() any { return new(PathInner) }, func() any { return new(pathHidden) },
	func() any { return new(pathEmbeddedPtr) }, func() any { return new(pathEmbeddedHidden) }, func() any { return new(*int) },
	func() any { return new(error) }, func() any { return new(fmt.Stringer) }, func() any { return new(chan int) },
	func() any { return new(pathInt) }, func() any { return new(time.Duration) }, func() any { return new(json.Number) },
	func() any { return new([]pathInt) }, func() any { return new(map[pathInt]pathInt) }, func() any { return new(pathRecMap) },
	func() any { return new(pathPtr) }, func() any { return new(pathNode) }, func() any { return new(complex128) },
	func() any { return new(pathSelfPtr) }, func() any { return new(pathPtrA) }, func() any { return new(pathDeep) },
	func() any { return nil }, func() any { return (*int)(nil) }, func() any { return 0 },
}

// pathNoPanic runs f, a call of a method of a Path, and reports a panic of it as a failure of the call which name
// describes.
func pathNoPanic(t *testing.T, name func() string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: panic: %v", name(), r)
		}
	}()
	f()
}

// Path.Get, Path.Unmarshal and Path.Extract return, with values or an error, for every path, source and destination:
// none panics in null, a nil pointer or interface, a struct, a named type or a value which holds itself, or for a
// destination which is nil or not a pointer.
func TestPathNoPanic(t *testing.T) {
	for _, p := range pathTestPaths {
		path, err := json.CreatePath(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, text := range pathTestTexts {
			var src any
			if err := json.Unmarshal([]byte(text), &src); err != nil {
				t.Fatal(err)
			}
			pathNoPanic(t, func() string { return fmt.Sprintf("Extract(%s) of %s", p, text) }, func() { _, _ = path.Extract([]byte(text)) })
			for _, newDst := range pathTestDestinations {
				pathNoPanic(t, func() string { return fmt.Sprintf("Get(%s) of %s into %T", p, text, newDst()) }, func() { _ = path.Get(src, newDst()) })
				pathNoPanic(t, func() string { return fmt.Sprintf("Unmarshal(%s) of %s into %T", p, text, newDst()) }, func() { _ = path.Unmarshal([]byte(text), newDst()) })
			}
		}
		for i, src := range pathTestSources() {
			for _, newDst := range pathTestDestinations {
				pathNoPanic(t, func() string { return fmt.Sprintf("Get(%s) of source %d, %T, into %T", p, i, src, newDst()) }, func() { _ = path.Get(src, newDst()) })
			}
		}
	}
}

// pathSelected returns as JSON texts the values in v, which Path.Get or Path.Unmarshal gave: v itself, or the
// elements of the slice of a path with .. or [*], sorted, as Get goes through the members of a map in no order.
func pathSelected(t *testing.T, v any, multi bool) []string {
	t.Helper()
	values := []any{v}
	if multi {
		values, _ = v.([]any)
	}
	texts := make([]string, len(values))
	for i, value := range values {
		b, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		texts[i] = string(b)
	}
	if multi {
		sort.Strings(texts)
	}
	return texts
}

// Path.Get selects in a value decoded from a text what Path.Unmarshal selects in the text: the same values, and none
// where Unmarshal selects none, or fails as the path does not fit the text. Get has an error there unless the path
// has .. or [*], which leave out the values in which the rest of the path selects nothing.
func TestPathGetSelectsAsUnmarshal(t *testing.T) {
	for _, p := range pathTestPaths {
		path, err := json.CreatePath(p)
		if err != nil {
			t.Fatal(err)
		}
		multi := strings.Contains(p, "..") || strings.Contains(p, "[*]")
		for _, text := range pathTestTexts {
			var src any
			if err := json.Unmarshal([]byte(text), &src); err != nil {
				t.Fatal(err)
			}
			var want []string
			var values []any
			if err := path.Unmarshal([]byte(text), &values); err == nil {
				want = pathSelected(t, values, true)
			}
			var got []string
			var v any
			if err := path.Get(src, &v); err == nil {
				got = pathSelected(t, v, multi)
			}
			if (len(got) != 0 || len(want) != 0) && !reflect.DeepEqual(got, want) {
				t.Errorf("%s of %s: Get selects %q, Unmarshal %q", p, text, got, want)
			}
		}
	}
}

// .. selects its name at any depth, in the value it selects too, and [*] every element, with the rest of the path
// applied to each value: one in which the rest selects nothing, or which it does not fit, adds nothing. Extract gives
// the values in the order of the text.
func TestPathRecursiveAndAll(t *testing.T) {
	for _, tc := range []struct {
		path, text string
		want       []string
	}{
		{`$..b`, `{"b":{"b":1}}`, []string{`{"b":1}`, `1`}},
		{`$..b`, `{"a":{"b":[1,2]},"c":[{"b":3}]}`, []string{`[1,2]`, `3`}},
		{`$..b`, `{"b":null,"a":{"b":{"c":1}}}`, []string{`null`, `{"c":1}`}},
		{`$..b`, `[1,"b",{"x":[{"b":true}]}]`, []string{`true`}},
		{`$..b.c`, `{"b":{"c":[1,2]},"x":{"b":{"c":3}},"y":{"b":[{"c":4}]}}`, []string{`[1,2]`, `3`}},
		{`$..b[0]`, `{"b":[1,2],"c":{"b":{"0":3}},"d":{"b":[[4]]}}`, []string{`1`, `[4]`}},
		{`$..b[*]`, `{"b":[1,{"b":[2]}]}`, []string{`1`, `{"b":[2]}`, `2`}},
		{`$..b[*].c`, `{"b":[[1],{"c":2}],"c":{"b":{"c":3}}}`, []string{`2`}},
		{`$..a..b`, `{"a":{"x":{"b":1}},"b":2}`, []string{`1`}},
		{`$[*]..b`, `[{"b":1},{"a":{"b":2}},3]`, []string{`1`, `2`}},
		{`$[*][*]`, `[[1,2],[3],{"a":4},5]`, []string{`1`, `2`, `3`}},
		{`$[*].b`, `[{"b":1},[2],{"c":{"b":3}},{"b":4}]`, []string{`1`, `4`}},
		{`$.a[*][0]`, `{"a":[[1],{"0":2},[[3]]]}`, []string{`1`, `[3]`}},
		{`$..c`, `{"a":1}`, nil},
		{`$..b`, `null`, nil},
	} {
		path, err := json.CreatePath(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		contents, err := path.Extract([]byte(tc.text))
		if err != nil {
			t.Errorf("Extract(%s) of %s: %v", tc.path, tc.text, err)
			continue
		}
		var extracted []string
		for _, content := range contents {
			extracted = append(extracted, string(content))
		}
		if !reflect.DeepEqual(extracted, tc.want) {
			t.Errorf("Extract(%s) of %s: got %q, want %q", tc.path, tc.text, extracted, tc.want)
		}
		var src any
		if err := json.Unmarshal([]byte(tc.text), &src); err != nil {
			t.Fatal(err)
		}
		var v any
		if err := path.Get(src, &v); err != nil {
			t.Errorf("Get(%s) of %s: %v", tc.path, tc.text, err)
			continue
		}
		want := append([]string{}, tc.want...)
		sort.Strings(want)
		if got := pathSelected(t, v, true); !reflect.DeepEqual(got, want) {
			t.Errorf("Get(%s) of %s: got %q, want %q", tc.path, tc.text, got, want)
		}
	}
}

// Path.Get selects in a struct by the names of its JSON object, in a struct embedded in it by the name of the field,
// and not in an unexported field or one tagged "-".
func TestPathGetStruct(t *testing.T) {
	src := &struct {
		ID     int `json:"id"`
		hidden int
		Skip   int `json:"-"`
		Next   *pathNode
		PathInner
	}{ID: 1, hidden: 2, Skip: 3, Next: &pathNode{V: 4, Next: &pathNode{V: 5}}, PathInner: PathInner{B: 6}}
	for p, want := range map[string]any{`$.id`: 1, `$.Next.Next.V`: 5, `$.PathInner.B`: 6, `$..V`: []any{4, 5}} {
		path, err := json.CreatePath(p)
		if err != nil {
			t.Fatal(err)
		}
		var got any
		if err := path.Get(src, &got); err != nil {
			t.Errorf("Get(%s): %v", p, err)
			continue
		}
		if values, ok := got.([]any); ok {
			sort.Slice(values, func(i, j int) bool { return values[i].(int) < values[j].(int) })
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Get(%s): got %#v, want %#v", p, got, want)
		}
	}
	for _, p := range []string{`$.hidden`, `$.Skip`, `$.ID`} {
		path, err := json.CreatePath(p)
		if err != nil {
			t.Fatal(err)
		}
		var got any
		if err := path.Get(src, &got); err == nil {
			t.Errorf("Get(%s): got %#v, want an error", p, got)
		}
	}
}

// Path.Get casts the value it selects to the destination, which may be of a named type: null to its zero value, and
// to a pointer by allocating what it points to. It returns an error for a destination which is not a pointer or is
// nil, for a value which cannot be cast to it, and for a value which holds itself.
func TestPathGetCast(t *testing.T) {
	get := func(p string, src any, dst any) error {
		t.Helper()
		path, err := json.CreatePath(p)
		if err != nil {
			t.Fatal(err)
		}
		return path.Get(src, dst)
	}
	var doc any
	if err := json.Unmarshal([]byte(`{"n":1500000000,"s":"7","l":[1,2],"m":{"k":3},"o":{"B":4,"C":{"B":5},"D":6},"z":null}`), &doc); err != nil {
		t.Fatal(err)
	}
	seven, five := 7, 5
	for _, tc := range []struct {
		path string
		dst  any
		want any
	}{
		{`$.n`, new(time.Duration), 1500 * time.Millisecond},
		{`$.n`, new(json.Number), json.Number("1.5e+09")},
		{`$.s`, new(pathInt), pathInt(7)},
		{`$.s`, new(*int), &seven},
		{`$.l`, new([]pathInt), []pathInt{1, 2}},
		{`$.l`, new([2]int), [2]int{1, 2}},
		{`$.m`, new(map[string]pathInt), map[string]pathInt{"k": 3}},
		{`$.o`, new(pathEmbeddedPtr), pathEmbeddedPtr{PathInner: &PathInner{B: 4, C: &PathInner{B: 5}}, D: 6}},
		{`$.o`, new(pathEmbeddedHidden), pathEmbeddedHidden{D: 6}},
		{`$.z`, &five, 0},
		{`$.z`, new([]int), []int(nil)},
		{`$.z`, new(*int), (*int)(nil)},
	} {
		if err := get(tc.path, doc, tc.dst); err != nil {
			t.Errorf("Get(%s) into %T: %v", tc.path, tc.dst, err)
			continue
		}
		if got := reflect.ValueOf(tc.dst).Elem().Interface(); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Get(%s) into %T: got %#v, want %#v", tc.path, tc.dst, got, tc.want)
		}
	}
	var ptr pathPtr
	if err := get(`$.P`, pathStruct{P: &PathInner{B: 1}}, &ptr); err != nil || ptr == nil || ptr.B != 1 {
		t.Errorf("Get($.P) into *pathPtr: got %#v, %v", ptr, err)
	}
	// pointer types which lead back to each other take a value of one of them
	var pa pathPtrA
	var pb pathPtrB
	pa, pb = &pb, &pa
	var toPA *pathPtrA
	if err := get(`$.a`, map[string]any{"a": pa}, &toPA); err != nil || toPA == nil || *toPA != pa {
		t.Errorf("Get($.a) of a pathPtrA into *pathPtrA: got %v, %v", toPA, err)
	}
	for _, dst := range []any{nil, 0, (*int)(nil)} {
		var invalid *json.InvalidUnmarshalError
		if err := get(`$.n`, doc, dst); !errors.As(err, &invalid) {
			t.Errorf("Get($.n) into %#v: got error %v, want an InvalidUnmarshalError", dst, err)
		}
	}
	for _, dst := range []any{new(error), new(chan int), new(fmt.Stringer), new(pathStruct), new(pathSelfPtr), new(pathPtrA)} {
		if err := get(`$.n`, doc, dst); err == nil {
			t.Errorf("Get($.n) into %T: got no error", dst)
		}
	}
	// a json.Number is a number of JSON
	for _, src := range []any{"abc", true, "", "1e", " 1"} {
		var n json.Number
		if err := get(`$.a`, map[string]any{"a": src}, &n); err == nil {
			t.Errorf("Get($.a) of %#v into json.Number: got %q, want an error", src, n)
		}
	}
	for src, want := range map[any]json.Number{"12": "12", 12: "12", -1.5: "-1.5"} {
		var n json.Number
		if err := get(`$.a`, map[string]any{"a": src}, &n); err != nil || n != want {
			t.Errorf("Get($.a) of %#v into json.Number: got %q, %v, want %q", src, n, err, want)
		}
	}
	node := &pathNode{V: 1}
	node.Next = node
	cyclicMap := map[string]any{"V": 1}
	cyclicMap["x"] = cyclicMap
	cyclicSlice := []any{map[string]any{"V": 1}, nil}
	cyclicSlice[1] = cyclicSlice
	var selfInterface any
	selfInterface = &selfInterface
	var selfPtr pathSelfPtr
	selfPtr = &selfPtr
	for _, src := range []any{node, cyclicMap, cyclicSlice, selfInterface, selfPtr} {
		var unsupported *json.UnsupportedValueError
		if err := get(`$..V`, src, new(any)); !errors.As(err, &unsupported) {
			t.Errorf("Get($..V) of %T which holds itself: got error %v, want an UnsupportedValueError", src, err)
		}
	}
	// a pointer which leads back to itself holds no value for any path
	for _, src := range []any{selfInterface, selfPtr} {
		for _, p := range []string{`$.V`, `$[0]`, `$[*]`} {
			var unsupported *json.UnsupportedValueError
			if err := get(p, src, new(any)); !errors.As(err, &unsupported) {
				t.Errorf("Get(%s) of %T which leads back to itself: got error %v, want an UnsupportedValueError", p, src, err)
			}
		}
	}
	selfMap := map[string]any{}
	selfMap["x"] = selfMap
	var unsupported *json.UnsupportedValueError
	if err := get(`$.x`, selfMap, new(pathRecMap)); !errors.As(err, &unsupported) {
		t.Errorf("Get($.x) of a map which holds itself into a recursive map: got error %v, want an UnsupportedValueError", err)
	}
}

// A Path selects the same values in a call after one which fails, and in calls which goroutines make at the same
// time: Extract and Unmarshal do not change it.
func TestPathReuse(t *testing.T) {
	path, err := json.CreatePath(`$.a[1].b`)
	if err != nil {
		t.Fatal(err)
	}
	text := []byte(`{"a":[{"b":0},{"b":1}],"b":2}`)
	extract := func() string {
		contents, err := path.Extract(text)
		if err != nil {
			return err.Error()
		}
		return string(bytes.Join(contents, []byte(",")))
	}
	for _, bad := range []string{`{"a":[{"b":0},{"b":[1,}]}`, `{"a":[0,{"b":0,]}`, `{"a":{"b":1}}`} {
		if _, err := path.Extract([]byte(bad)); err == nil {
			t.Fatalf("Extract of %s: got no error", bad)
		}
		var v []any
		_ = path.Unmarshal([]byte(bad), &v)
		if got := extract(); got != "1" {
			t.Fatalf("Extract after a failed call on %s: got %s, want 1", bad, got)
		}
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if got := extract(); got != "1" {
					t.Errorf("Extract at the same time as other goroutines: got %s, want 1", got)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// The cases of issue 774: Path.Get past JSON null, past a nil pointer and in a struct returns, with no panic, what the
// path selects, or an error where a path without .. or [*] selects nothing.
func TestIssue774PathGetPastNil(t *testing.T) {
	var doc any
	if err := json.Unmarshal([]byte(`{"a":null}`), &doc); err != nil {
		t.Fatal(err)
	}
	type S struct{ A int }
	for _, tc := range []struct {
		path    string
		src     any
		dst     any
		want    any
		wantErr bool
	}{
		{`$..b`, doc, new(any), []any(nil), false},
		{`$.a.b`, doc, new(any), nil, true},
		{`$.a[0]`, doc, new(any), nil, true},
		{`$.a[*]`, doc, new(any), nil, true},
		{`$.a`, doc, new(any), nil, false},
		{`$.A`, S{A: 1}, new(int), 1, false},
		{`$..A`, S{A: 1}, new([]int), []int{1}, false},
		{`$..A`, map[string]*S{"x": nil}, new([]int), []int{}, false},
		{`$..A`, map[string]*S{"x": {A: 2}}, new([]int), []int{2}, false},
	} {
		path, err := json.CreatePath(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		err = path.Get(tc.src, tc.dst)
		if tc.wantErr {
			if err == nil {
				t.Errorf("Get(%s) of %#v: got no error", tc.path, tc.src)
			}
			continue
		}
		if err != nil {
			t.Errorf("Get(%s) of %#v: %v", tc.path, tc.src, err)
			continue
		}
		if got := reflect.ValueOf(tc.dst).Elem().Interface(); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Get(%s) of %#v: got %#v, want %#v", tc.path, tc.src, got, tc.want)
		}
	}
}

// Path.Get and Path.Unmarshal go into a value as deep as encoding/json decodes and encodes, counting its objects and
// arrays and not the pointers between them, and return an error for a Go value nested deeper, instead of exhausting
// the stack. A chain of pointers, however long, is followed in a loop.
func TestPathDepth(t *testing.T) {
	const maxDepth = 10000 // of encoding/json, which the decoder decodes and the encoder encodes at most
	path := func(p string) *json.Path {
		t.Helper()
		path, err := json.CreatePath(p)
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	all, root, first, member, values := path(`$..x`), path(`$`), path(`$[0]`), path(`$.a`), path(`$..V`)

	arrays := []byte(strings.Repeat(`[`, maxDepth) + strings.Repeat(`]`, maxDepth))
	var src any
	if err := json.Unmarshal(arrays, &src); err != nil {
		t.Fatal(err)
	}
	var v any
	if err := all.Get(src, &v); err != nil {
		t.Errorf("Get($..x) of %d arrays: %v", maxDepth, err)
	}
	var deep pathDeep
	if err := root.Unmarshal(arrays, &deep); err != nil {
		t.Errorf("Unmarshal($) of %d arrays into a recursive slice: %v", maxDepth, err)
	}
	// a pointer and a map or a slice at each level
	var ptrMap pathPtrMap
	if err := json.Unmarshal([]byte(strings.Repeat(`{"x":`, maxDepth-1)+`{}`+strings.Repeat(`}`, maxDepth-1)), &ptrMap); err != nil {
		t.Fatal(err)
	}
	var selected []any
	if err := all.Get(ptrMap, &selected); err != nil || len(selected) != maxDepth-1 {
		t.Errorf("Get($..x) of %d maps of pointers: got %d values, %v, want %d", maxDepth, len(selected), err, maxDepth-1)
	}
	var ptrSlice pathPtrSlice
	if err := json.Unmarshal(arrays, &ptrSlice); err != nil {
		t.Fatal(err)
	}
	if err := all.Get(ptrSlice, &v); err != nil {
		t.Errorf("Get($..x) of %d slices of pointers: %v", maxDepth, err)
	}
	var list *pathNode
	for i := 0; i < maxDepth; i++ {
		list = &pathNode{V: i, Next: list}
	}
	if _, err := json.Marshal(list); err != nil {
		t.Fatal(err)
	}
	if err := values.Get(list, &selected); err != nil || len(selected) != maxDepth {
		t.Errorf("Get($..V) of a list of %d: got %d values, %v, want %d", maxDepth, len(selected), err, maxDepth)
	}
	// a struct embedded by its value is a part of the object it is in
	var embedded *pathEmbeddedNode
	for i := 0; i < maxDepth; i++ {
		embedded = &pathEmbeddedNode{PathEmbeddedBase: PathEmbeddedBase{Next: embedded}, V: i}
	}
	if _, err := json.Marshal(embedded); err != nil {
		t.Fatal(err)
	}
	if err := values.Get(embedded, &selected); err != nil || len(selected) != maxDepth {
		t.Errorf("Get($..V) of %d structs which embed the next: got %d values, %v, want %d", maxDepth, len(selected), err, maxDepth)
	}

	var deeper any
	for i := 0; i < 2*maxDepth; i++ {
		deeper = []any{deeper}
	}
	longer := list
	for i := 0; i < maxDepth; i++ {
		longer = &pathNode{V: i, Next: longer}
	}
	// a struct embedded with a name in its tag is a member of its own, a level, as Marshal writes it
	var tagged *pathTaggedNode
	for i := 0; i < maxDepth; i++ {
		tagged = &pathTaggedNode{PathTaggedBase: PathTaggedBase{Next: tagged}, V: i}
	}
	if _, err := json.Marshal(tagged); err == nil {
		t.Fatalf("Marshal of %d structs which embed the next with a name: got no error", maxDepth)
	}
	for name, get := range map[string]func() error{
		"Get($..x) of nested slices":                  func() error { return all.Get(deeper, &v) },
		"Get($[0]) of nested slices into a recursive": func() error { return first.Get(deeper, &deep) },
		"Get($..V) of a list":                         func() error { return values.Get(longer, &v) },
		"Get($..V) of structs embedded with a name":   func() error { return values.Get(tagged, &v) },
	} {
		var unsupported *json.UnsupportedValueError
		if err := get(); !errors.As(err, &unsupported) {
			t.Errorf("%s %d deep: got error %v, want an UnsupportedValueError", name, 2*maxDepth, err)
		}
	}

	var chain pathSelfPtr
	var held any = 7
	for i := 0; i < 10*maxDepth; i++ {
		next := chain
		chain = &next
		inner := held
		held = &inner
	}
	var got int
	if err := member.Get(map[string]any{"a": chain}, &got); err != nil || got != 0 {
		t.Errorf("Get($.a) of a chain of %d pointers to nil into an int: got %d, %v, want 0", 10*maxDepth, got, err)
	}
	if err := member.Get(map[string]any{"a": held}, &got); err != nil || got != 7 {
		t.Errorf("Get($.a) of a chain of %d pointers to 7 into an int: got %d, %v, want 7", 10*maxDepth, got, err)
	}
	var unsupported *json.UnsupportedValueError
	if err := member.Get(chain, &v); err == nil || errors.As(err, &unsupported) {
		t.Errorf("Get($.a) of a chain of %d pointers to nil: got error %v, want the one of null", 10*maxDepth, err)
	}
}

// Path.Get selects a member of a map by a key of a string kind, also when an interface holds it, and passes over
// the other keys, in whatever order the map has them.
func TestPathGetMapKeys(t *testing.T) {
	path, err := json.CreatePath(`$.a`)
	if err != nil {
		t.Fatal(err)
	}
	all, err := json.CreatePath(`$..a`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		var got int
		if err := path.Get(map[any]any{1: 2, "a": 3, 2.5: 1, true: 0, nil: 4}, &got); err != nil || got != 3 {
			t.Fatalf("Get($.a) of a map with keys of any type: got %d, %v, want 3", got, err)
		}
	}
	var got int
	if err := path.Get(map[pathKey]int{"a": 1, "b": 2}, &got); err != nil || got != 1 {
		t.Errorf("Get($.a) of a map with a named string key: got %d, %v, want 1", got, err)
	}
	var values []int
	if err := all.Get(map[string]any{"x": map[pathKey]int{"a": 1}, "y": map[any]int{"a": 2, 3: 4}}, &values); err != nil {
		t.Fatal(err)
	}
	sort.Ints(values)
	if !reflect.DeepEqual(values, []int{1, 2}) {
		t.Errorf("Get($..a): got %v, want [1 2]", values)
	}
}

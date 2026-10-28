//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsonv1 "encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	json "github.com/goccy/go-json/v2"
)

// The tests here compare the three ways to the semantics of encoding/json: encoding/json itself, which Go 1.27 makes
// of encoding/json/v2 with its DefaultOptionsV1 with GOEXPERIMENT=jsonv2, the root package of go-json, which replaces
// it, and the v2 json package with DefaultOptionsV1 of the root package, which must be the same as encoding/json. The
// root package keeps the differences which it has by design or which issues records ( rootDiffers ): it decodes as it
// reads the input, which a syntax error finds partly decoded, and names the type of the value in the MarshalerError of
// a method of a value receiver ( issues/encoder-error-texts.md ), as encoding/json of Go 1.27 without
// GOEXPERIMENT=jsonv2 does.

// rootDiffers are the known differences of the root package.
const (
	rootDecodesBeforeSyntaxError = "decodes before the syntax error"
	rootNamesValueReceiver       = "names the value of the receiver"
)

// v1Inner and v1Outer are structs of the tags and the kinds which the options of encoding/json change.
type v1Inner struct {
	A int    `json:"a,omitempty"`
	B string `json:",omitempty"`
}

type v1Outer struct {
	Name    string
	Num     int           `json:",string"`
	Flag    bool          `json:",string"`
	Text    string        `json:",string"`
	Bytes   []byte        `json:",omitempty"`
	Array   [2]byte       `json:"array"`
	Dur     time.Duration `json:"dur"`
	Inner   v1Inner       `json:"inner,omitempty"`
	PInner  *v1Inner      `json:"pinner,omitempty"`
	Map     map[string]int
	Slice   []int
	Snake_N int
	Any     any
}

// v1PtrJSON has a MarshalJSON and an UnmarshalJSON of its pointer, which encoding/json calls for an addressable value.
type v1PtrJSON struct{ X int }

func (v *v1PtrJSON) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf(`{"p":%d}`, v.X)), nil }

func (v *v1PtrJSON) UnmarshalJSON(b []byte) error {
	if string(b) == `"bad"` {
		return errors.New("v1PtrJSON failed")
	}
	v.X = len(b)
	return nil
}

type v1Methods struct {
	P  v1PtrJSON
	PP *v1PtrJSON
	L  []v1PtrJSON
}

// v1Text has a MarshalText which fails for 0.
type v1Text int

func (t v1Text) MarshalText() ([]byte, error) {
	if t == 0 {
		return nil, errors.New("v1Text failed")
	}
	return []byte(fmt.Sprint("t", int(t))), nil
}

// v1Value is a value of the marshaling of the comparison, and the known difference of the root package, if it has
// one.
type v1Value struct {
	v           any
	rootDiffers string
}

// v1Values are the values of the marshaling of the comparison.
func v1Values() []v1Value {
	n := 1
	values := []any{
		v1Outer{Name: "<a&b>", Num: 3, Flag: true, Text: "x", Array: [2]byte{1, 2}, Dur: 5, Map: map[string]int{"b": 1, "a": 2}},
		v1Outer{Bytes: []byte("hi"), PInner: &v1Inner{}, Slice: []int{}, Any: []any{1, "s", nil}},
		map[string]int{"z": 1, "a": 2, "m": 3}, map[int]string{2: "b", 10: "a"}, map[v1Text]int{1: 1, 2: 2},
		[]int(nil), map[string]int(nil), []byte(nil), []byte("bytes"), [3]byte{7, 8, 9},
		"  ", "\xff\xfe", "<script>", time.Duration(42), time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
		stdjsonv1.Number("12.5"), stdjsonv1.RawMessage(`{"r": 1}`), &n, (*int)(nil), struct{ a, B int }{1, 2},
		v1Methods{P: v1PtrJSON{1}, PP: &v1PtrJSON{2}, L: []v1PtrJSON{{3}}}, &v1Methods{P: v1PtrJSON{4}},
		[]v1PtrJSON{{5}}, map[string]v1PtrJSON{"k": {6}},
		[]v1Byte{1, 2}, [2]v1Byte{3, 4},
		math.NaN(), math.Inf(1), make(chan int), func() {}, complex(1, 2),
		struct {
			A int `json:"x"`
			B int `json:"x"`
			C int
		}{1, 2, 3},
	}
	var all []v1Value
	for _, v := range values {
		all = append(all, v1Value{v: v})
	}
	return append(all, v1Value{v1Text(0), rootNamesValueReceiver}, v1Value{map[v1Text]int{0: 1}, rootNamesValueReceiver})
}

// v1Case is an input and the type of the value it is decoded into, made by into for each way, which may set it first,
// and the known difference of the root package, if it has one.
type v1Case struct {
	name        string
	in          string
	into        func() any
	rootDiffers string
}

func v1Cases() []v1Case {
	newOuter := outOf(v1Outer{})
	return []v1Case{
		{"fields", `{"Name":"n","Num":"3","Flag":"true","Text":"\"x\"","Bytes":"aGk=","array":[1,2],"dur":5,"inner":{"a":1},"Map":{"a":1},"Slice":[1],"Snake_N":2,"Any":{"k":[1.5]}}`,
			newOuter, ""},
		{"folded names", `{"NAME":"n","name":"m","snake_n":1,"SnakeN":2}`, newOuter, ""},
		{"names twice", `{"Name":"a","Name":"b","Map":{"x":1,"x":2}}`, newOuter, ""},
		{"type errors", `{"Name":1,"Num":"x","Slice":[1,"a",3],"Map":{"a":"b","c":4},"dur":"5s"}`, newOuter, ""},
		{"merge", `{"Map":{"b":2},"Slice":[9],"inner":{"B":"y"},"pinner":{"a":7}}`, func() any {
			return &v1Outer{Map: map[string]int{"a": 1}, Slice: []int{1, 2, 3}, Inner: v1Inner{A: 5}, PInner: &v1Inner{B: "q"}}
		}, ""},
		{"array lengths", `[[1,2,3],[4]]`, outOf([2][2]int{}), ""},
		{"loose base64", `{"Bytes":"aG` + "\\r\\n" + `k="}`, newOuter, ""},
		{"loose time", `"2020-01-02T3:04:05Z"`, outOf(time.Time{}), ""},
		{"named bytes", `["AQI=","AQI="]`, outOf([2][]v1Byte{}), ""},
		{"bytes of an array", `[1,2]`, outOf([]byte(nil)), ""},
		{"invalid UTF-8", "{\"Name\":\"a\xffb\",\"\xff\":1}", newOuter, ""},
		{"null", `{"Name":null,"Num":null,"Map":null,"Slice":null,"pinner":null}`, func() any {
			return &v1Outer{Name: "keep", Num: 1, Map: map[string]int{"a": 1}, Slice: []int{1}, PInner: &v1Inner{}}
		}, ""},
		{"numbers into any", `[1, 1.5, 1e400, "s", true, null, {"a": [2]}]`, outOf([]any(nil)), ""},
		{"methods", `{"P":[1],"PP":"bad","L":[{}, "bad", 3]}`, outOf(v1Methods{}), ""},
		{"map keys", `{"1":"a","x":"b","3":"c"}`, outOf(map[int]string(nil)), ""},
		{"syntax error", `{"Name":"n",}`, newOuter, rootDecodesBeforeSyntaxError},
		{"syntax error in an array", `{"Name":"n","Slice":[1,]}`, newOuter, rootDecodesBeforeSyntaxError},
		{"missing value", `{"Name":"n","Num":}`, newOuter, rootDecodesBeforeSyntaxError},
		{"trailing data", `{"Name":"n"} x`, newOuter, rootDecodesBeforeSyntaxError},
		{"not a pointer", `1`, func() any { return 1 }, ""},
		{"unknown fields", `{"Unknown":{"deep":[1,2]},"Name":"n"}`, newOuter, ""},
		{"string option of a quoted null", `{"Num":"null","Flag":"null"}`, outOf(v1Outer{Num: 3, Flag: true}), ""},
	}
}

// encoding/json, the root package and the v2 json package with DefaultOptionsV1 marshal the same values to the same
// output, or to the same error.
func TestMarshalV1WaysMatchStd(t *testing.T) {
	v1 := gojson.DefaultOptionsV1()
	for i, c := range v1Values() {
		want, wantErr := stdjsonv1.Marshal(c.v)
		got, err := json.Marshal(c.v, v1)
		if g, w := describeV1Error(err), describeV1Error(wantErr); g != w || err == nil && string(got) != string(want) {
			t.Errorf("%d %T: v2 with DefaultOptionsV1:\n got %s %s\nwant %s %s", i, c.v, got, g, want, w)
		}
		if c.rootDiffers != "" {
			continue
		}
		got, err = gojson.Marshal(c.v)
		if g, w := describeV1Error(err), describeV1Error(wantErr); g != w || err == nil && string(got) != string(want) {
			t.Errorf("%d %T: root package:\n got %s %s\nwant %s %s", i, c.v, got, g, want, w)
		}
	}
}

// encoding/json, the root package and the v2 json package with DefaultOptionsV1 unmarshal the same inputs to the same
// values, or to the same errors.
func TestUnmarshalV1WaysMatchStd(t *testing.T) {
	v1 := gojson.DefaultOptionsV1()
	for _, c := range v1Cases() {
		theirs := c.into()
		wantErr := stdjsonv1.Unmarshal([]byte(c.in), theirs)
		ours := c.into()
		err := json.Unmarshal([]byte(c.in), ours, v1)
		if g, w := describeV1Error(err), describeV1Error(wantErr); g != w || !sameOutputs(ours, theirs) {
			t.Errorf("%s: v2 with DefaultOptionsV1:\n got %s %#v\nwant %s %#v", c.name, g, valueOf(ours), w, valueOf(theirs))
		}
		if c.rootDiffers != "" {
			continue
		}
		ours = c.into()
		err = gojson.Unmarshal([]byte(c.in), ours)
		if g, w := describeV1Error(err), describeV1Error(wantErr); g != w || !sameOutputs(ours, theirs) {
			t.Errorf("%s: root package:\n got %s %#v\nwant %s %#v", c.name, g, valueOf(ours), w, valueOf(theirs))
		}
	}
}

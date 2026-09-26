package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// The cases of the reported issues of the decoder: go-json decodes them as encoding/json of the running Go does.

// checkUnmarshal decodes doc into the values which newValue makes by go-json and by encoding/json, and compares
// the errors and the decoded values.
func checkUnmarshal(t *testing.T, doc string, newValue func() any) {
	t.Helper()
	want, got := newValue(), newValue()
	wantErr := describeTypeError(stdjson.Unmarshal([]byte(doc), want))
	gotErr := describeTypeError(json.Unmarshal([]byte(doc), got))
	if gotErr != wantErr {
		t.Errorf("%q:\n got %s\nwant %s", doc, gotErr, wantErr)
		return
	}
	if wantErr == "SyntaxError" {
		return
	}
	// a NaN is not equal to itself: the values are the same if they are printed the same
	if !reflect.DeepEqual(got, want) && describeValue(got) != describeValue(want) {
		t.Errorf("%q: decoded %s, want %s", doc, describeValue(got), describeValue(want))
	}
}

// describeValue prints the value which v points to, with the length of the strings, so that a string whose bytes
// are not what its length says is seen.
func describeValue(v any) string {
	return fmt.Sprintf("%#v", reflect.ValueOf(v).Elem().Interface())
}

// issue642Pair is a struct whose pointer implements encoding.TextUnmarshaler.
type issue642Pair struct{ A, B int }

func (p *issue642Pair) UnmarshalText(b []byte) error {
	p.A, p.B = len(b), -len(b)
	return nil
}

// issue642Name is a string whose pointer implements encoding.TextUnmarshaler.
type issue642Name string

func (n *issue642Name) UnmarshalText(b []byte) error {
	*n = issue642Name("text:" + string(b))
	return nil
}

// issue642List is a slice whose pointer implements encoding.TextUnmarshaler: null sets it to nil, as it does to
// any slice.
type issue642List []string

func (l *issue642List) UnmarshalText(b []byte) error {
	*l = issue642List{string(b)}
	return nil
}

// issue642Map is a map whose pointer implements encoding.TextUnmarshaler: null sets it to nil.
type issue642Map map[string]int

func (m *issue642Map) UnmarshalText(b []byte) error {
	*m = issue642Map{string(b): 1}
	return nil
}

func TestIssue642NullIntoTextUnmarshaler(t *testing.T) {
	// null is not given to UnmarshalText: a value of a type whose pointer implements encoding.TextUnmarshaler is
	// left as it is by null, but a slice, a map, a pointer or an interface value is set to nil.
	type fields struct {
		P  issue642Pair
		N  issue642Name
		L  issue642List
		M  issue642Map
		PP *issue642Pair
		PN *issue642Name
	}
	full := func() fields {
		pair, name := issue642Pair{7, 8}, issue642Name("keep")
		return fields{
			P: issue642Pair{7, 8}, N: "keep", L: issue642List{"keep"}, M: issue642Map{"keep": 1},
			PP: &pair, PN: &name,
		}
	}
	for _, doc := range []string{`null`, ` null `, `"x"`} {
		checkUnmarshal(t, doc, func() any { v := issue642Pair{7, 8}; return &v })
		checkUnmarshal(t, doc, func() any { v := issue642Name("keep"); return &v })
		checkUnmarshal(t, doc, func() any { v := issue642List{"keep"}; return &v })
		checkUnmarshal(t, doc, func() any { v := issue642Map{"keep": 1}; return &v })
	}
	for _, doc := range []string{
		`{"P":null,"N":null,"L":null,"M":null,"PP":null,"PN":null}`,
		`{"P":"x","N":"x","L":"x","M":"x","PP":"x","PN":"x"}`,
	} {
		checkUnmarshal(t, doc, func() any { v := full(); return &v })
	}
	for _, doc := range []string{`[null,"x",null]`} {
		checkUnmarshal(t, doc, func() any { return &[]issue642Pair{{7, 8}} })
		checkUnmarshal(t, doc, func() any { return &[]issue642Name{"keep"} })
		checkUnmarshal(t, doc, func() any { return &[3]issue642Name{"keep", "keep", "keep"} })
		checkUnmarshal(t, doc, func() any { return &[3]issue642Pair{{7, 8}, {7, 8}, {7, 8}} })
	}
	checkUnmarshal(t, `{"a":null,"b":"x"}`, func() any { return &map[string]issue642Name{"a": "keep"} })
	checkUnmarshal(t, `{"a":null,"b":"x"}`, func() any { return &map[string]issue642Pair{"a": {7, 8}} })
}

type decodeIssue450Embedded struct {
	Embedded string `json:"Embedded"`
	Other    int
}

type decodeIssue450Struct struct {
	decodeIssue450Embedded
	Name string
}

type decodeIssue450PtrStruct struct {
	*decodeIssue450Embedded
	Name string
}

type decodeIssue450Inner struct{ Inner string }

type decodeIssue450Outer struct{ decodeIssue450Inner }

// DecodeIssue450Label is embedded as a value which is not a struct: it is a field of its name.
type DecodeIssue450Label string

type decodeIssue450Labeled struct {
	DecodeIssue450Label
	Visible string
}

func TestIssue450EmbeddedFieldOfItsName(t *testing.T) {
	// A field promoted from an embedded struct is decoded although it has the name of the embedded struct, which
	// is no key of its own.
	for _, doc := range []string{
		`{"Embedded":"inside","Other":111,"Name":"outside"}`,
		`{"embedded":"inside"}`,
	} {
		checkUnmarshal(t, doc, func() any { return new(decodeIssue450Struct) })
		checkUnmarshal(t, doc, func() any { return new(decodeIssue450PtrStruct) })
	}
	checkUnmarshal(t, `{"Inner":"x"}`, func() any { return new(decodeIssue450Outer) })
	// an embedded value which is not a struct is a field of its name
	checkUnmarshal(t, `{"DecodeIssue450Label":"x","Visible":"v"}`, func() any { return new(decodeIssue450Labeled) })
}

func TestIssue555NumbersOutOfRange(t *testing.T) {
	// A number out of the range of a float64 decoded into an interface{} is a type error, after which the
	// decoding goes on, and a float field is set as encoding/json of the Go version sets it.
	huge := "1" + strings.Repeat("0", 400)
	for _, n := range []string{`1e400`, `-1e400`, huge, `1e39`, `-1e39`, `1e-400`} {
		for _, doc := range []string{n, `{"n":` + n + `,"m":1}`, `[` + n + `,1]`} {
			checkUnmarshal(t, doc, func() any { return new(any) })
			checkUnmarshal(t, doc, func() any { return new(map[string]any) })
			checkUnmarshal(t, doc, func() any { return new([]any) })
		}
		checkUnmarshal(t, `{"F":`+n+`,"X":1}`, func() any { return new(struct{ F, X float64 }) })
		checkUnmarshal(t, `{"F":`+n+`,"X":1}`, func() any { return new(struct{ F, X float32 }) })
		checkUnmarshal(t, `{"F":`+n+`,"X":1}`, func() any { return new(struct{ F, X any }) })
	}
}

func TestStringOptionNumbers(t *testing.T) {
	// A number in a string of the string option is read by strconv, as encoding/json does: it may have leading
	// zeros, and a value which fails is not stored.
	type fields struct {
		I  int     `json:",string"`
		U  uint8   `json:",string"`
		F  float64 `json:",string"`
		F3 float32 `json:",string"`
		P  *int    `json:",string"`
		X  int
	}
	for _, s := range []string{
		`01`, `00`, `-01`, `+1`, `1-`, ` 1`, `1 `, `1.`, `.5`, `1e5`, `1.e5`, `-Inf`, `+Inf`, `Inf`, `NaN`, `0x10`,
		`1_000`, `300`, `-1`, `1e400`, `1e39`, ``, `-`, `1.5`, `99999999999999999999`,
	} {
		for _, name := range []string{"I", "U", "F", "F3", "P"} {
			checkUnmarshal(t, `{"`+name+`":"`+s+`","X":7}`, func() any { return new(fields) })
		}
	}
}

func TestNumberMapKeys(t *testing.T) {
	// The keys of a map of numbers are read by strconv, as encoding/json reads them.
	for _, key := range []string{`01`, `+1`, `-1`, ` 1`, `1 `, `1.5`, `1e2`, `300`, `-Inf`, `0x10`, `1_000`, ``} {
		doc := `{"` + key + `":1,"2":2}`
		checkUnmarshal(t, doc, func() any { return new(map[int]int) })
		checkUnmarshal(t, doc, func() any { return new(map[uint8]int) })
		checkUnmarshal(t, doc, func() any { return new(map[float64]int) })
	}
}

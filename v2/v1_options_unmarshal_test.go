package json_test

import (
	stdjsonv1 "encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The options of encoding/json when unmarshaling, on every version of Go, with the values which encoding/json/v2 of
// Go 1.27 decodes ( see std_v1_options_diff_test.go, which compares with it ).

type v1UnmarshalCase struct {
	name     string
	in       string
	opts     []json.Options
	newValue func() any
	// want is the value which the pointer made by newValue points to after the decoding, or nil for an error
	want any
}

// outOf returns the function which makes a pointer to a new value of T set to v. v must refer to no memory, which
// the values of two calls would share: a value which does, a map, a slice or a pointer which is not nil, is made by a
// function of its own.
func outOf[T any](v T) func() any {
	return func() any {
		p := new(T)
		*p = v
		return p
	}
}

// streamBufferSize is the size of the buffer of the stream of UnmarshalRead ( initBufSize of internal/decoder ), the
// most bytes it reads at once: an object or an array which the buffer doesn't hold whole is decoded element by element.
const streamBufferSize = 512

// longerThanStreamBuffer is a string whose JSON string is longer than the buffer of a stream.
var longerThanStreamBuffer = strings.Repeat("x", streamBufferSize)

func runV1UnmarshalCases(t *testing.T, cases []v1UnmarshalCase) {
	t.Helper()
	for _, c := range cases {
		v := c.newValue()
		err := json.Unmarshal([]byte(c.in), v, c.opts...)
		got := reflect.ValueOf(v).Elem().Interface()
		switch {
		case c.want == nil && err == nil:
			t.Errorf("%s: got %#v, want an error", c.name, got)
		case c.want != nil && err != nil:
			t.Errorf("%s: %v", c.name, err)
		case c.want != nil && !reflect.DeepEqual(got, c.want):
			t.Errorf("%s: got %#v, want %#v", c.name, got, c.want)
		}
	}
}

func TestUnmarshalMatchCaseSensitiveDelimiter(t *testing.T) {
	type names struct {
		FooBar int `json:"foo_bar"`
		Baz    int `json:"baz,case:ignore"`
	}
	both := []json.Options{json.MatchCaseInsensitiveNames(true), gojson.MatchCaseSensitiveDelimiter(true)}
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"another case", `{"FOO_BAR":1}`, both, outOf(names{}), names{FooBar: 1}},
		{"without the underscore", `{"foobar":1}`, both, outOf(names{}), names{}},
		{"without the underscore insensitive", `{"foobar":1}`, both[:1], outOf(names{}), names{FooBar: 1}},
		{"case:ignore", `{"BAZ":2}`, both[1:], outOf(names{}), names{Baz: 2}},
	})
}

func TestUnmarshalStringifyWithLegacySemantics(t *testing.T) {
	stringify := []json.Options{gojson.StringifyWithLegacySemantics(true)}
	type values struct {
		B bool    `json:",string"`
		S string  `json:",string"`
		I int     `json:",string"`
		P *string `json:",string"`
	}
	s := "s"
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"values", `{"B":"true","S":"\"x\"","I":"+3"}`, stringify, outOf(values{}), values{B: true, S: "x", I: 3}},
		{"values by default", `{"B":"true"}`, nil, outOf(values{}), nil},
		{"quoted nulls", `{"B":"null","S":"null","I":"null","P":"null"}`, stringify, outOf(values{B: true, S: "x", I: 1, P: &s}),
			values{}},
		// encoding/json.Number, a string kind, is no string of the option: a JSON string of a JSON string of it fails
		{"number of a JSON string", `{"N":"\"1\""}`, stringify, outOf(struct {
			N stdjsonv1.Number `json:",string"`
		}{}), nil},
		// a quoted null of an escape is a quoted null too, unlike encoding/json/v2 for a pointer, which tells it by the
		// raw value ( see issues/decoder-v2-small-deviations.md )
		{"quoted null of an escape", `{"P":"nul\u006c"}`, stringify, outOf(values{P: &s}), values{}},
	})
}

func TestUnmarshalMergeWithLegacySemantics(t *testing.T) {
	merge := []json.Options{gojson.MergeWithLegacySemantics(true)}
	type pair struct{ X, Y int }
	type values struct {
		I  int
		S  string
		A  [2]pair
		Sl []pair
		M  map[string]pair
	}
	newValues := func() any {
		return &values{I: 1, S: "s", A: [2]pair{{1, 2}, {3, 4}}, Sl: append(make([]pair, 0, 2), pair{1, 2}), M: map[string]pair{"a": {1, 2}}}
	}
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"nulls", `{"I":null,"S":null}`, merge, newValues, values{I: 1, S: "s", A: [2]pair{{1, 2}, {3, 4}},
			Sl: []pair{{1, 2}}, M: map[string]pair{"a": {1, 2}}}},
		{"elements and values", `{"A":[{"X":9},{"Y":8}],"Sl":[{"X":9}],"M":{"a":{"X":9}}}`, merge, newValues, values{I: 1, S: "s",
			A: [2]pair{{9, 2}, {3, 8}}, Sl: []pair{{9, 2}}, M: map[string]pair{"a": {9, 0}}}},
		{"elements and values by default", `{"A":[{"X":9},{"Y":8}],"Sl":[{"X":9}],"M":{"a":{"X":9}}}`, nil, newValues, values{I: 1,
			S: "s", A: [2]pair{{9, 0}, {0, 8}}, Sl: []pair{{9, 0}}, M: map[string]pair{"a": {9, 2}}}},
	})
}

// v1KeyText is a string whose UnmarshalJSON and UnmarshalText tell which decodes it, of which
// CallMethodsWithLegacySemantics calls UnmarshalText only for a key.
type v1KeyText string

func (k *v1KeyText) UnmarshalJSON([]byte) error {
	*k = "json"
	return nil
}

func (k *v1KeyText) UnmarshalText([]byte) error {
	*k = "text"
	return nil
}

func TestUnmarshalCallMethodsWithLegacySemantics(t *testing.T) {
	callMethods := []json.Options{gojson.CallMethodsWithLegacySemantics(true)}
	newMap := outOf(map[v1KeyText]v1KeyText(nil))
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"key", `{"a":"b"}`, callMethods, newMap, map[v1KeyText]v1KeyText{"text": "json"}},
		{"key by default", `{"a":"b"}`, nil, newMap, map[v1KeyText]v1KeyText{"json": "json"}},
	})
}

// v1SyntaxJSON is an int whose UnmarshalJSON fails with a syntax error, which stops the decoding.
type v1SyntaxJSON int

func (*v1SyntaxJSON) UnmarshalJSON([]byte) error {
	return &jsontext.SyntacticError{Err: errV1Method}
}

// v1FailingJSON is an int whose UnmarshalJSON fails.
type v1FailingJSON int

var errV1Method = errors.New("method failed")

func (*v1FailingJSON) UnmarshalJSON([]byte) error { return errV1Method }

// With ReportErrorsWithLegacySemantics, Unmarshal, UnmarshalRead and UnmarshalDecode check the input first, decode on
// after the error of a value, and report the first one, as the errors of encoding/json.
func TestUnmarshalReportErrorsWithLegacySemantics(t *testing.T) {
	reportErrors := gojson.ReportErrorsWithLegacySemantics(true)
	type kinds struct {
		A, B int
		C    []int
	}
	const kindsIn = `{"A":"x","B":"y","C":[1,"z",3]}`
	var of kinds
	err := json.UnmarshalOf([]byte(kindsIn), &of, reportErrors)
	var ofErr *gojson.UnmarshalTypeError
	if !errors.As(err, &ofErr) || ofErr.Field != "A" || !reflect.DeepEqual(of.C, []int{1, 0, 3}) {
		t.Errorf("UnmarshalOf of kinds: got %v, %+v, want the error of A and C decoded", err, of)
	}
	var nilErr *gojson.InvalidUnmarshalError
	if err := json.UnmarshalOf([]byte(`1`), (*int)(nil), reportErrors); !errors.As(err, &nilErr) {
		t.Errorf("UnmarshalOf into nil: got %v, want an InvalidUnmarshalError", err)
	}
	if err := json.UnmarshalReadOf(strings.NewReader(`1`), (*int)(nil), reportErrors); !errors.As(err, &nilErr) {
		t.Errorf("UnmarshalReadOf into nil: got %v, want an InvalidUnmarshalError", err)
	}
	// the output is checked before the input is read, as encoding/json checks it: a reader which fails is not read
	for _, out := range []any{nil, 1, (*int)(nil)} {
		in := iotest.ErrReader(errors.New("read failed"))
		if err := json.UnmarshalRead(in, out, reportErrors); !errors.As(err, &nilErr) {
			t.Errorf("UnmarshalRead of a failing reader into %T: got %v, want an InvalidUnmarshalError", out, err)
		}
	}
	// the error of the reader within a token is returned as it is, as encoding/json returns it
	errRead := errors.New("read failed")
	var elems []int
	dec := jsontext.NewDecoder(io.MultiReader(strings.NewReader(`[1,2`), iotest.ErrReader(errRead)))
	if err := json.UnmarshalDecode(dec, &elems, reportErrors); err != errRead {
		t.Errorf("UnmarshalDecode of a reader which fails within a number: got %v, want %v", err, errRead)
	}
	for _, u := range unmarshalFuncs {
		unmarshal := func(in string, out any, opts ...json.Options) error {
			return u.unmarshal(in, out, append(opts, reportErrors)...)
		}
		var v kinds
		err := unmarshal(kindsIn, &v)
		var typeErr *gojson.UnmarshalTypeError
		// the offset of an UnmarshalTypeError is the one after the value, as encoding/json reports it
		if !errors.As(err, &typeErr) || typeErr.Value != "string" || typeErr.Offset != int64(len(`{"A":"x"`)) ||
			typeErr.Struct != "kinds" || typeErr.Field != "A" || !reflect.DeepEqual(v.C, []int{1, 0, 3}) {
			t.Errorf("%s of kinds: got %v, %+v, want the error of A and C decoded", u.name, err, v)
		}
		v = kinds{A: 9}
		var syntaxErr *gojson.SyntaxError
		if err := unmarshal(`{"A":1,`, &v); !errors.As(err, &syntaxErr) || v.A != 9 {
			t.Errorf("%s of a cut input: got %v, %+v, want a SyntaxError and the value unchanged", u.name, err, v)
		}
		var invalidErr *gojson.InvalidUnmarshalError
		if err := unmarshal(`1`, v); !errors.As(err, &invalidErr) {
			t.Errorf("%s into a value: got %v, want an InvalidUnmarshalError", u.name, err)
		}
		var withArray struct{ A [2]int }
		if err := unmarshal(`{"A":[1,2,3]}`, &withArray); !errors.As(err, &typeErr) || typeErr.Offset != int64(len(`{"A":[1,2,3]`)) ||
			typeErr.Field != "A" {
			t.Errorf("%s of an array too long: got %v, want the error of the array after it", u.name, err)
		}
		var methods struct {
			J v1FailingJSON
			N int
		}
		if err := unmarshal(`{"J":1,"N":3}`, &methods); err != errV1Method || methods.N != 3 {
			t.Errorf("%s of a method: got %v, %+v, want the error of the method and N decoded", u.name, err, methods)
		}
		var syntax struct {
			J v1SyntaxJSON
			N int
		}
		if err := unmarshal(`{"J":1,"N":3}`, &syntax); !errors.As(err, &syntaxErr) || syntax.N != 0 {
			t.Errorf("%s of a method of a syntax error: got %v, %+v, want a SyntaxError and N not decoded", u.name, err, syntax)
		}
		var unknown struct{ A int }
		if err := unmarshal(`{"B":1,"A":2}`, &unknown, json.RejectUnknownMembers(true)); err == nil ||
			err.Error() != `json: unknown field "B"` || unknown.A != 2 {
			t.Errorf("%s of an unknown member: got %v, %+v, want the error of the name and A decoded", u.name, err, unknown)
		}
		var withChan struct {
			C chan int
			N int
		}
		if err := unmarshal(`{"C":null,"N":2}`, &withChan); err != nil || withChan.N != 2 {
			t.Errorf("%s of a null channel: got %v, %+v", u.name, err, withChan)
		}
		var conflict reportConflict
		if err := unmarshal(`{"x":1,"C":2}`, &conflict); err != nil || conflict.C != 2 || conflict.A != 0 {
			t.Errorf("%s of conflicting names: got %v, %+v", u.name, err, conflict)
		}
		var keys map[int]int
		if err := unmarshal(`{"x":1,"2":2}`, &keys); !errors.As(err, &typeErr) || typeErr.Value != "number x" || keys[2] != 2 {
			t.Errorf("%s of a key: got %v, %v, want the error of the key and the next entry", u.name, err, keys)
		}
		// encoding/json/v2 panics for it with the option ( see issues/decoder-v2-small-deviations.md )
		embedded := struct {
			A int
			X jsontext.Value `json:",embed"`
			C int
		}{X: jsontext.Value("1")}
		if err := unmarshal(`{"A":1,"b":[2],"C":3}`, &embedded); !errors.As(err, &typeErr) || typeErr.Field != "b" ||
			embedded.C != 3 || string(embedded.X) != "1" {
			t.Errorf("%s of a member into an embedded value which is not an object: got %v, %+v, want its error and C decoded",
				u.name, err, embedded)
		}
	}
}

func TestUnmarshalArrayFromAnyLength(t *testing.T) {
	anyLength := []json.Options{gojson.UnmarshalArrayFromAnyLength(true)}
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"longer", `[1,2,3]`, anyLength, outOf([2]int{}), [2]int{1, 2}},
		{"longer by default", `[1,2,3]`, nil, outOf([2]int{}), nil},
		{"shorter", `[1]`, anyLength, outOf([3]int{7, 7, 7}), [3]int{1, 0, 0}},
		{"shorter by default", `[1]`, nil, outOf([3]int{}), nil},
		{"invalid UTF-8 past the length", "[1,2,\"\xff\"]", anyLength, outOf([2]int{}), nil},
		{"a name twice past the length", `[1,2,{"a":1,"a":2}]`, anyLength, outOf([2]int{}), nil},
		{"bytes longer", `"AQID"`, anyLength, outOf([2]byte{}), [2]byte{1, 2}},
		{"bytes shorter", `"AQID"`, anyLength, outOf([4]byte{7, 7, 7, 7}), [4]byte{1, 2, 3, 0}},
		{"bytes shorter by default", `"AQID"`, nil, outOf([4]byte{}), nil},
	})
}

func TestParseBytesWithLooseRFC4648(t *testing.T) {
	loose := []json.Options{gojson.ParseBytesWithLooseRFC4648(true)}
	in := `"AQ` + "\\r\\n" + `ID"`
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"line feed", in, loose, outOf([]byte(nil)), []byte{1, 2, 3}},
		{"line feed by default", in, nil, outOf([]byte(nil)), nil},
		{"line feed into an array", in, loose, outOf([3]byte{}), [3]byte{1, 2, 3}},
		{"wrong character", `"AQ*ID"`, loose, outOf([]byte(nil)), nil},
	})
}

func TestParseTimeWithLooseRFC3339(t *testing.T) {
	loose := []json.Options{gojson.ParseTimeWithLooseRFC3339(true)}
	oneDigit := time.Date(2006, 1, 2, 1, 4, 5, 0, time.UTC)
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"hour of one digit", `"2006-01-02T1:04:05Z"`, loose, outOf(time.Time{}), oneDigit},
		{"hour of one digit by default", `"2006-01-02T1:04:05Z"`, nil, outOf(time.Time{}), nil},
		{"key of an hour of one digit", `{"2006-01-02T1:04:05Z":1}`, loose, outOf(map[time.Time]int(nil)), map[time.Time]int{oneDigit: 1}},
		{"wrong text", `"2006-01-02"`, loose, outOf(time.Time{}), nil},
	})
}

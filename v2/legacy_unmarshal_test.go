package json_test

import (
	stdjsonv1 "encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The options of encoding/json when unmarshaling, on every version of Go, with the values which encoding/json/v2 of
// Go 1.27 decodes ( see std_legacy_diff_test.go, which compares with it ).

type legacyUnmarshalCase struct {
	name     string
	in       string
	opts     []json.Options
	newValue func() any
	// want is the value which the pointer made by newValue points to after the decoding, or nil for an error
	want any
}

func newOf[T any](v T) func() any {
	return func() any {
		p := new(T)
		*p = v
		return p
	}
}

func runLegacyUnmarshalCases(t *testing.T, cases []legacyUnmarshalCase) {
	t.Helper()
	for _, c := range cases {
		v := c.newValue()
		err := json.Unmarshal([]byte(c.in), v, c.opts...)
		switch {
		case c.want == nil && err == nil:
			t.Errorf("%s: got %#v, want an error", c.name, reflect.ValueOf(v).Elem().Interface())
		case c.want != nil && err != nil:
			t.Errorf("%s: %v", c.name, err)
		case c.want != nil && !reflect.DeepEqual(reflect.ValueOf(v).Elem().Interface(), c.want):
			t.Errorf("%s: got %#v, want %#v", c.name, reflect.ValueOf(v).Elem().Interface(), c.want)
		}
	}
}

func TestUnmarshalMatchCaseSensitiveDelimiter(t *testing.T) {
	type names struct {
		FooBar int `json:"foo_bar"`
		Baz    int `json:"baz,case:ignore"`
	}
	both := []json.Options{json.MatchCaseInsensitiveNames(true), gojson.MatchCaseSensitiveDelimiter(true)}
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"another case", `{"FOO_BAR":1}`, both, newOf(names{}), names{FooBar: 1}},
		{"without the underscore", `{"foobar":1}`, both, newOf(names{}), names{}},
		{"without the underscore insensitive", `{"foobar":1}`, both[:1], newOf(names{}), names{FooBar: 1}},
		{"case:ignore", `{"BAZ":2}`, both[1:], newOf(names{}), names{Baz: 2}},
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
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"values", `{"B":"true","S":"\"x\"","I":"+3"}`, stringify, newOf(values{}), values{B: true, S: "x", I: 3}},
		{"values by default", `{"B":"true"}`, nil, newOf(values{}), nil},
		{"quoted nulls", `{"B":"null","S":"null","I":"null","P":"null"}`, stringify, newOf(values{B: true, S: "x", I: 1, P: &s}),
			values{}},
		// a quoted null of an escape is a quoted null too, unlike encoding/json/v2, which tells it by the raw value
		// encoding/json.Number, a string kind, is no string of the option: a JSON string of a JSON string of it fails
		{"number of a JSON string", `{"N":"\"1\""}`, stringify, newOf(struct {
			N stdjsonv1.Number `json:",string"`
		}{}), nil},
		{"quoted null of an escape", `{"P":"nul\u006c"}`, stringify, newOf(values{P: &s}), values{}},
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
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"nulls", `{"I":null,"S":null}`, merge, newValues, values{I: 1, S: "s", A: [2]pair{{1, 2}, {3, 4}},
			Sl: []pair{{1, 2}}, M: map[string]pair{"a": {1, 2}}}},
		{"elements and values", `{"A":[{"X":9},{"Y":8}],"Sl":[{"X":9}],"M":{"a":{"X":9}}}`, merge, newValues, values{I: 1, S: "s",
			A: [2]pair{{9, 2}, {3, 8}}, Sl: []pair{{9, 2}}, M: map[string]pair{"a": {9, 0}}}},
		{"elements and values by default", `{"A":[{"X":9},{"Y":8}],"Sl":[{"X":9}],"M":{"a":{"X":9}}}`, nil, newValues, values{I: 1,
			S: "s", A: [2]pair{{9, 0}, {0, 8}}, Sl: []pair{{9, 0}}, M: map[string]pair{"a": {9, 2}}}},
	})
}

// legacyLitKey is a string whose UnmarshalJSON and UnmarshalText tell which decodes it.
type legacyLitKey string

func (k *legacyLitKey) UnmarshalJSON([]byte) error {
	*k = "json"
	return nil
}

func (k *legacyLitKey) UnmarshalText([]byte) error {
	*k = "text"
	return nil
}

func TestUnmarshalCallMethodsWithLegacySemantics(t *testing.T) {
	legacy := []json.Options{gojson.CallMethodsWithLegacySemantics(true)}
	newMap := func() any { return new(map[legacyLitKey]legacyLitKey) }
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"key", `{"a":"b"}`, legacy, newMap, map[legacyLitKey]legacyLitKey{"text": "json"}},
		{"key by default", `{"a":"b"}`, nil, newMap, map[legacyLitKey]legacyLitKey{"json": "json"}},
	})
}

// legacyLitSyntaxJSON is an int whose UnmarshalJSON fails with a syntax error, which stops the decoding.
type legacyLitSyntaxJSON int

func (*legacyLitSyntaxJSON) UnmarshalJSON([]byte) error {
	return &jsontext.SyntacticError{Err: errLegacyLitMethod}
}

// legacyLitFailingJSON is an int whose UnmarshalJSON fails.
type legacyLitFailingJSON int

var errLegacyLitMethod = errors.New("method failed")

func (*legacyLitFailingJSON) UnmarshalJSON([]byte) error { return errLegacyLitMethod }

// With ReportErrorsWithLegacySemantics, Unmarshal, UnmarshalRead and UnmarshalDecode check the input first, decode on
// after the error of a value, and report the first one, as the errors of encoding/json.
func TestUnmarshalReportErrorsWithLegacySemantics(t *testing.T) {
	legacy := gojson.ReportErrorsWithLegacySemantics(true)
	type kinds struct {
		A, B int
		C    []int
	}
	unmarshals := map[string]func(in string, out any, opts ...json.Options) error{
		"Unmarshal": func(in string, out any, opts ...json.Options) error { return json.Unmarshal([]byte(in), out, opts...) },
		"UnmarshalRead": func(in string, out any, opts ...json.Options) error {
			return json.UnmarshalRead(strings.NewReader(in), out, opts...)
		},
		"UnmarshalDecode": func(in string, out any, opts ...json.Options) error {
			return json.UnmarshalDecode(jsontext.NewDecoder(strings.NewReader(in)), out, opts...)
		},
	}
	var of kinds
	err := json.UnmarshalOf([]byte(`{"A":"x","B":"y","C":[1,"z",3]}`), &of, legacy)
	var ofErr *gojson.UnmarshalTypeError
	if !errors.As(err, &ofErr) || ofErr.Field != "A" || !reflect.DeepEqual(of.C, []int{1, 0, 3}) {
		t.Errorf("UnmarshalOf of kinds: got %v, %+v, want the error of A and C decoded", err, of)
	}
	var nilErr *gojson.InvalidUnmarshalError
	if err := json.UnmarshalOf([]byte(`1`), (*int)(nil), legacy); !errors.As(err, &nilErr) {
		t.Errorf("UnmarshalOf into nil: got %v, want an InvalidUnmarshalError", err)
	}
	if err := json.UnmarshalReadOf(strings.NewReader(`1`), (*int)(nil), legacy); !errors.As(err, &nilErr) {
		t.Errorf("UnmarshalReadOf into nil: got %v, want an InvalidUnmarshalError", err)
	}
	for name, unmarshal := range unmarshals {
		var v kinds
		err := unmarshal(`{"A":"x","B":"y","C":[1,"z",3]}`, &v, legacy)
		var typeErr *gojson.UnmarshalTypeError
		if !errors.As(err, &typeErr) || typeErr.Value != "string" || typeErr.Offset != 8 || typeErr.Struct != "kinds" ||
			typeErr.Field != "A" || !reflect.DeepEqual(v.C, []int{1, 0, 3}) {
			t.Errorf("%s of kinds: got %v, %+v, want the error of A and C decoded", name, err, v)
		}
		v = kinds{A: 9}
		var syntaxErr *gojson.SyntaxError
		if err := unmarshal(`{"A":1,`, &v, legacy); !errors.As(err, &syntaxErr) || v.A != 9 {
			t.Errorf("%s of a cut input: got %v, %+v, want a SyntaxError and the value unchanged", name, err, v)
		}
		var invalidErr *gojson.InvalidUnmarshalError
		if err := unmarshal(`1`, v, legacy); !errors.As(err, &invalidErr) {
			t.Errorf("%s into a value: got %v, want an InvalidUnmarshalError", name, err)
		}
		var withArray struct{ A [2]int }
		if err := unmarshal(`{"A":[1,2,3]}`, &withArray, legacy); !errors.As(err, &typeErr) || typeErr.Offset != 12 ||
			typeErr.Field != "A" {
			t.Errorf("%s of an array too long: got %v, want the error of the array after it", name, err)
		}
		var methods struct {
			J legacyLitFailingJSON
			N int
		}
		if err := unmarshal(`{"J":1,"N":3}`, &methods, legacy); err != errLegacyLitMethod || methods.N != 3 {
			t.Errorf("%s of a method: got %v, %+v, want the error of the method and N decoded", name, err, methods)
		}
		var syntax struct {
			J legacyLitSyntaxJSON
			N int
		}
		if err := unmarshal(`{"J":1,"N":3}`, &syntax, legacy); !errors.As(err, &syntaxErr) || syntax.N != 0 {
			t.Errorf("%s of a method of a syntax error: got %v, %+v, want a SyntaxError and N not decoded", name, err, syntax)
		}
		var unknown struct{ A int }
		if err := unmarshal(`{"B":1,"A":2}`, &unknown, legacy, json.RejectUnknownMembers(true)); err == nil ||
			err.Error() != `json: unknown field "B"` || unknown.A != 2 {
			t.Errorf("%s of an unknown member: got %v, %+v, want the error of the name and A decoded", name, err, unknown)
		}
		var withChan struct {
			C chan int
			N int
		}
		if err := unmarshal(`{"C":null,"N":2}`, &withChan, legacy); err != nil || withChan.N != 2 {
			t.Errorf("%s of a null channel: got %v, %+v", name, err, withChan)
		}
		var conflict struct {
			A int `json:"x"`
			B int `json:"x"`
			C int
		}
		if err := unmarshal(`{"x":1,"C":2}`, &conflict, legacy); err != nil || conflict.C != 2 || conflict.A != 0 {
			t.Errorf("%s of conflicting names: got %v, %+v", name, err, conflict)
		}
		var keys map[int]int
		if err := unmarshal(`{"x":1,"2":2}`, &keys, legacy); !errors.As(err, &typeErr) || typeErr.Value != "number x" ||
			keys[2] != 2 {
			t.Errorf("%s of a key: got %v, %v, want the error of the key and the next entry", name, err, keys)
		}
		// encoding/json/v2 panics for it with the option ( see issues/decoder-v2-small-deviations.md )
		embedded := struct {
			A int
			X jsontext.Value `json:",embed"`
			C int
		}{X: jsontext.Value("1")}
		if err := unmarshal(`{"A":1,"b":[2],"C":3}`, &embedded, legacy); !errors.As(err, &typeErr) || typeErr.Field != "b" ||
			embedded.C != 3 || string(embedded.X) != "1" {
			t.Errorf("%s of a member into an embedded value which is not an object: got %v, %+v, want its error and C decoded", name, err, embedded)
		}
	}
}

func TestUnmarshalArrayFromAnyLength(t *testing.T) {
	anyLength := []json.Options{gojson.UnmarshalArrayFromAnyLength(true)}
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"longer", `[1,2,3]`, anyLength, newOf([2]int{}), [2]int{1, 2}},
		{"longer by default", `[1,2,3]`, nil, newOf([2]int{}), nil},
		{"shorter", `[1]`, anyLength, newOf([3]int{7, 7, 7}), [3]int{1, 0, 0}},
		{"shorter by default", `[1]`, nil, newOf([3]int{}), nil},
		{"invalid UTF-8 past the length", "[1,2,\"\xff\"]", anyLength, newOf([2]int{}), nil},
		{"a name twice past the length", `[1,2,{"a":1,"a":2}]`, anyLength, newOf([2]int{}), nil},
		{"bytes longer", `"AQID"`, anyLength, newOf([2]byte{}), [2]byte{1, 2}},
		{"bytes shorter", `"AQID"`, anyLength, newOf([4]byte{7, 7, 7, 7}), [4]byte{1, 2, 3, 0}},
		{"bytes shorter by default", `"AQID"`, nil, newOf([4]byte{}), nil},
	})
}

func TestParseBytesWithLooseRFC4648(t *testing.T) {
	loose := []json.Options{gojson.ParseBytesWithLooseRFC4648(true)}
	in := `"AQ` + "\\r\\n" + `ID"`
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"line feed", in, loose, newOf([]byte(nil)), []byte{1, 2, 3}},
		{"line feed by default", in, nil, newOf([]byte(nil)), nil},
		{"line feed into an array", in, loose, newOf([3]byte{}), [3]byte{1, 2, 3}},
		{"wrong character", `"AQ*ID"`, loose, newOf([]byte(nil)), nil},
	})
}

func TestParseTimeWithLooseRFC3339(t *testing.T) {
	loose := []json.Options{gojson.ParseTimeWithLooseRFC3339(true)}
	oneDigit := time.Date(2006, 1, 2, 1, 4, 5, 0, time.UTC)
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"hour of one digit", `"2006-01-02T1:04:05Z"`, loose, newOf(time.Time{}), oneDigit},
		{"hour of one digit by default", `"2006-01-02T1:04:05Z"`, nil, newOf(time.Time{}), nil},
		{"key of an hour of one digit", `{"2006-01-02T1:04:05Z":1}`, loose, func() any { return new(map[time.Time]int) },
			map[time.Time]int{oneDigit: 1}},
		{"wrong text", `"2006-01-02"`, loose, newOf(time.Time{}), nil},
	})
}

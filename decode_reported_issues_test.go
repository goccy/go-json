package json_test

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

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

// numberGrammarCases are numbers which are not valid, and valid ones around them.
var numberGrammarCases = []string{
	`0.`, `1.`, `-2.`, `01`, `00`, `-01`, `-012`, `012`, `1.e5`, `0.e1`, `2.e+3`, `2.e-3`, `-.123`, `.5`, `-`, `-.`,
	`1e`, `1e+`, `1E-`, `1.5e`, `1..2`, `1e5.5`, `1ee5`, `1.2.3`, `--1`, `1-`, `-e1`, `1e+-2`, `0.e`, `+1`,
	`0`, `-0`, `0.5`, `-0.5`, `10`, `1e5`, `1E+5`, `1e-5`, `0e0`, `-0.0e-0`, `123.456e7`,
}

func TestIssue448NumberGrammar(t *testing.T) {
	// A number is checked by the grammar of JSON wherever it is: skipped, decoded into any of the types, or
	// validated.
	for _, n := range numberGrammarCases {
		for _, doc := range []string{n, `[` + n + `]`, `{"x":` + n + `}`, `{"skipped":` + n + `,"x":1}`, `[[` + n + `],1]`} {
			checkUnmarshal(t, doc, func() any { return new(any) })
			checkUnmarshal(t, doc, func() any { return new(struct{}) })
			checkUnmarshal(t, doc, func() any { return new(struct{ X any }) })
			checkUnmarshal(t, doc, func() any { return new(struct{ X float64 }) })
			checkUnmarshal(t, doc, func() any { return new(struct{ X int }) })
			checkUnmarshal(t, doc, func() any { return new(map[string]any) })
			checkUnmarshal(t, doc, func() any { return new([]any) })
			if got, want := json.Valid([]byte(doc)), stdjson.Valid([]byte(doc)); got != want {
				t.Errorf("Valid(%q) = %v, want %v", doc, got, want)
			}
		}
		for _, newValue := range []func() any{
			func() any { return new(float64) }, func() any { return new(float32) }, func() any { return new(int) },
			func() any { return new(uint) }, func() any { return new(json.Number) },
			func() any { return new(struct{ X json.Number }) },
			func() any {
				return new(struct {
					X int `json:",string"`
				})
			},
		} {
			checkUnmarshal(t, n, newValue)
			checkUnmarshal(t, `{"X":`+n+`}`, newValue)
			checkUnmarshal(t, `{"X":"`+n+`"}`, newValue)
		}
	}
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

// tokenDecoder is the decoder of go-json and the one of encoding/json.
type tokenDecoder interface {
	Token() (stdjson.Token, error)
	Decode(any) error
	More() bool
	InputOffset() int64
}

// streamTokens reads the tokens of doc by Token and Decode by turns, as by the calls of pattern: 't' a Token,
// 'd' a Decode into an interface{}, 'm' a More.
func streamTokens(doc string, pattern string, newDecoder func(string) tokenDecoder) string {
	dec := newDecoder(doc)
	var b strings.Builder
	for i := 0; i < 40; i++ {
		switch pattern[i%len(pattern)] {
		case 't':
			tok, err := dec.Token()
			fmt.Fprintf(&b, "T(%v %v %d) ", tok, describeTypeErrorMessage(err), dec.InputOffset())
			if err != nil {
				return b.String()
			}
		case 'd':
			var v any
			err := dec.Decode(&v)
			// the syntax errors of the decoding of a value are compared by their kind only
			fmt.Fprintf(&b, "D(%v %v %d) ", v, describeTypeError(err), dec.InputOffset())
			if err != nil {
				return b.String()
			}
		case 'm':
			fmt.Fprintf(&b, "M(%v) ", dec.More())
		}
	}
	return b.String()
}

// describeTypeErrorMessage describes an error with its message, and the offset of a syntax error.
func describeTypeErrorMessage(err error) string {
	var std *stdjson.SyntaxError
	var goErr *json.SyntaxError
	switch {
	case err == nil:
		return "<nil>"
	case errors.As(err, &std):
		return fmt.Sprintf("syntax %q at %d", std.Error(), std.Offset)
	case errors.As(err, &goErr):
		return fmt.Sprintf("syntax %q at %d", goErr.Error(), goErr.Offset)
	}
	return err.Error()
}

func TestIssue548TokenGrammar(t *testing.T) {
	// The delimiters of the tokens are checked by the grammar, and a value decoded between the tokens follows the
	// comma or the colon before it, as encoding/json has them.
	stdDecoder := func(doc string) tokenDecoder {
		return stdjson.NewDecoder(strings.NewReader(doc))
	}
	goDecoder := func(doc string) tokenDecoder {
		return json.NewDecoder(strings.NewReader(doc))
	}
	for _, doc := range []string{
		`{"hello": "value"` + "\n" + ` "foo": "bar"}`,
		`[1 2]`, `{"a" "b"}`, `{"a":1,}`, `[1,]`, `[,1]`, `{,"a":1}`, `{"a"::1}`, `{"a":1:2}`, `[1:2]`, `{"a",1}`,
		`{1:2}`, `{"a":1]`, `[1}`, `[1,,2]`, `{"a":1,,"b":2}`, `{"a":{"b":1}"c":2}`, `[[1][2]]`, `[true false]`,
		`{"a":1 "b":2}`, `]`, `}`, `,`, `:`, `1 , 2`, `1 2`, `{"a":[1,{"b":2}],"c":"d"} [3]`,
		`[{"a":1},{"a":2}]`, `{"a":{"b":[1,2]},"c":3}`, ``, `[`, `{"a"`, `{"a":`, `[1,`,
	} {
		for _, pattern := range []string{"t", "d", "td", "ttd", "tmd", "tttd", "tdt", "ttdd"} {
			want := streamTokens(doc, pattern, stdDecoder)
			got := streamTokens(doc, pattern, goDecoder)
			if got != want {
				t.Errorf("%q by %q:\n got %s\nwant %s", doc, pattern, got, want)
			}
		}
	}
}

func TestValidByGrammar(t *testing.T) {
	// Valid is the grammar of JSON, as encoding/json has it.
	for _, doc := range []string{
		`[1]]`, `1]`, `{}}`, `["x"]]`, `[-1e+9999]`, `[1.5e+9999]`, `[123123e100000]`, `0.4e0066999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999969999999006`,
		`123` + "\x00", "{\"a\":1}\x00", "\x00", ``, ` `, `null`, ` true `, `"a"`, `"\x"`, "\"\x01\"", `[`, `{"a"`, `{"a":1,}`,
		`[1,2,3]`, `{"a":{"b":[1,{"c":null}]}}`, `01`, `1.`, `-`, "\"\xff\"", `{"a" 1}`, `[1 2]`, `tru`, `nul`, `falsy`,
	} {
		if got, want := json.Valid([]byte(doc)), stdjson.Valid([]byte(doc)); got != want {
			t.Errorf("Valid(%q) = %v, want %v", doc, got, want)
		}
	}
	r := rand.New(rand.NewSource(3))
	for i := 0; i < 20000; i++ {
		doc := randomSkippedValue(r, 1+r.Intn(4))
		if r.Intn(2) == 0 {
			doc = mutate(r, doc)
		}
		if got, want := json.Valid([]byte(doc)), stdjson.Valid([]byte(doc)); got != want {
			t.Fatalf("Valid(%q) = %v, want %v", doc, got, want)
		}
	}
}

func TestFuncFieldTypeError(t *testing.T) {
	// A value of a func is a type error of its field, after which the decoding goes on; null leaves it nil.
	type funcs struct {
		F func()
		A int
	}
	for _, doc := range []string{
		`{"F":1,"A":2}`, `{"F":"x","A":2}`, `{"F":null,"A":2}`, `{"F":[1,{"a":2}],"A":2}`, `{"F":{},"A":2}`,
		`{"F":true,"A":2}`, `{"F":nul,"A":2}`, `{"F":[1 2],"A":2}`,
	} {
		checkUnmarshal(t, doc, func() any { return new(funcs) })
	}
	checkUnmarshal(t, `[true,null]`, func() any { return new([]func()) })
	checkUnmarshal(t, `1`, func() any { return new(func()) })
}

func TestStringOptionJSONNumber(t *testing.T) {
	// The bytes of the string of a json.Number of the string option are checked as encoding/json of the Go version
	// checks them: before Go 1.27, any bytes which start as a number are stored as they are.
	type number struct {
		N json.Number `json:",string"`
		A int
	}
	type stdNumber struct {
		N stdjson.Number `json:",string"`
		A int
	}
	check := func(doc string) {
		t.Helper()
		want, got := new(stdNumber), new(number)
		wantErr := describeTypeError(stdjson.Unmarshal([]byte(doc), want))
		gotErr := describeTypeError(json.Unmarshal([]byte(doc), got))
		if gotErr != wantErr || string(got.N) != string(want.N) || got.A != want.A {
			t.Errorf("%q:\n got %q %d %s\nwant %q %d %s", doc, got.N, got.A, gotErr, want.N, want.A, wantErr)
		}
	}
	for _, s := range []string{
		`1`, `-1.5e3`, `01`, `0x`, `1e`, `-`, `1.`, `1..2`, `--1`, `1 `, ` 1`, `+1`, `00`, `.5`, `12abc`, `1,2`, ``,
		`true`, `false`, `null`, `nul`, `{}`, `\"1\"`, `\"x\"`, `\"1`, `\u0031`, `1\u0030`, ` null`, `\"1\" `,
	} {
		check(`{"N":"` + s + `","A":2}`)
	}
	for _, doc := range []string{`{"N":1,"A":2}`, `{"N":true,"A":2}`, `{"N":{},"A":2}`, `{"N":null,"A":2}`} {
		check(doc)
	}
}

func TestStringOptionWhiteSpace(t *testing.T) {
	// The bytes of a string of the string option are a value with nothing around it: a white space is an error.
	type fields struct {
		B  bool    `json:",string"`
		S  string  `json:",string"`
		I  int     `json:",string"`
		F  float64 `json:",string"`
		PB *bool   `json:",string"`
		U  uint8   `json:",string"`
	}
	for field, values := range map[string][]string{
		"B":  {` true`, `true `, `\ttrue`, `true\n`, ` false `},
		"S":  {` \"a\"`, `\"a\" `, ` \"a\" `},
		"I":  {` 1`, `1 `, ` -1`},
		"F":  {` 1.5`, `1.5 `},
		"PB": {` true`, `true `},
	} {
		for _, value := range values {
			checkUnmarshal(t, `{"`+field+`":"`+value+`","U":"7"}`, func() any { return new(fields) })
		}
	}
}

func TestTimeOfOtherKinds(t *testing.T) {
	// A time.Time is decoded from a string: any other value is given to its UnmarshalJSON before Go 1.27, and is a
	// type error by Go 1.27.
	type times struct {
		T time.Time
		U int
	}
	for _, doc := range []string{
		`{"T":1,"U":2}`, `{"T":true,"U":2}`, `{"T":{"a":[1]},"U":2}`, `{"T":[],"U":2}`, `{"T":null,"U":2}`,
		`{"T":"2020-01-01T00:00:00Z","U":2}`, `{"T":-1.5e3,"U":2}`, `{"T":tru,"U":2}`,
	} {
		checkUnmarshal(t, doc, func() any { return new(times) })
	}
	checkUnmarshal(t, `1`, func() any { return new(time.Time) })
}

// methodErrorValue is a value whose unmarshal methods fail for 0 and x.
type methodErrorValue struct{ V int }

func (v *methodErrorValue) UnmarshalJSON(b []byte) error {
	if string(b) == "0" {
		return errors.New("method error")
	}
	v.V = len(b)
	return nil
}

// methodErrorText is a value whose UnmarshalText fails for x.
type methodErrorText struct{ V string }

func (t *methodErrorText) UnmarshalText(b []byte) error {
	if string(b) == "x" {
		return errors.New("text method error")
	}
	t.V = string(b)
	return nil
}

// methodStdTypeError is a value whose UnmarshalJSON returns a type error of encoding/json.
type methodStdTypeError struct{}

func (*methodStdTypeError) UnmarshalJSON([]byte) error {
	return &stdjson.UnmarshalTypeError{Value: "number", Type: reflect.TypeOf(0)}
}

func TestUnmarshalMethodErrors(t *testing.T) {
	// The error of an unmarshal method stops the decoding before Go 1.27. By Go 1.27, the decoding goes on after
	// it, as after a type error, and the first error is returned: the error of the method as it is.
	type fields struct {
		A int
		B methodErrorValue
		C *methodErrorValue
		T methodErrorText
		E int
	}
	for _, doc := range []string{
		`{"A":1,"B":0,"E":2}`, `{"A":"x","B":0,"E":2}`, `{"B":0,"A":"x","E":2}`,
		`{"T":"x","E":2}`, `{"B":0,"E":"x"}`, `{"B":0,"E":2,"C":1}`, `{"B":1,"T":"y","E":2}`,
	} {
		checkUnmarshal(t, doc, func() any { return new(fields) })
	}
	// a type error of encoding/json has its field set before Go 1.27, as encoding/json sets it
	type stdTypeError struct {
		N methodStdTypeError
		E int
	}
	checkUnmarshal(t, `{"N":1,"E":2}`, func() any { return new(stdTypeError) })
}

func TestTypeErrorPathOfKeys(t *testing.T) {
	// The path of a type error has the keys of the input, which Go 1.27 escapes as the tokens of a JSON pointer.
	for _, doc := range []string{`{"a/b":"x"}`, `{"a~b":"x"}`, `{"a.b":"x"}`, `{"~1/":{"c":"x"}}`} {
		checkUnmarshal(t, doc, func() any { return new(map[string]int) })
		checkUnmarshal(t, doc, func() any { return new(map[string]map[string]int) })
	}
}

func TestNulAfterValue(t *testing.T) {
	// A nul byte in the input is a byte which the grammar doesn't have: after a value, it is a syntax error.
	for _, doc := range []string{"123\x00", "{\"a\":1}\x00", "\"s\" \x00", "[1]\x00 ", "true\x00", "1 \x00 2"} {
		checkUnmarshal(t, doc, func() any { return new(any) })
		checkUnmarshal(t, doc, func() any { return new(struct{ A int }) })
		var got, want any
		gotErr, wantErr := json.UnmarshalOf([]byte(doc), &got), stdjson.Unmarshal([]byte(doc), &want)
		if describeTypeErrorMessage(gotErr) != describeTypeErrorMessage(wantErr) {
			t.Errorf("UnmarshalOf(%q): got %s, want %s", doc, describeTypeErrorMessage(gotErr), describeTypeErrorMessage(wantErr))
		}
	}
}

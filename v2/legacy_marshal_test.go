package json_test

import (
	"encoding"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The options of encoding/json when marshaling, on every version of Go, with the outputs which encoding/json/v2 of
// Go 1.27 writes ( see std_legacy_diff_test.go, which compares with it ).

type legacyMarshalCase struct {
	name string
	in   any
	opts []json.Options
	// want is the output, or "" for an error
	want string
}

func runLegacyMarshalCases(t *testing.T, cases []legacyMarshalCase) {
	t.Helper()
	for _, c := range cases {
		got, err := json.Marshal(c.in, append(c.opts, json.Deterministic(true))...)
		switch {
		case c.want == "" && err == nil:
			t.Errorf("%s: got %s, want an error", c.name, got)
		case c.want != "" && (err != nil || string(got) != c.want):
			t.Errorf("%s: got %s, %v, want %s", c.name, got, err, c.want)
		}
	}
}

// legacyByte is a named byte type, whose slices and arrays are of bytes only with FormatBytesWithLegacySemantics.
type legacyByte byte

// legacyTextByte is a named byte type with methods of marshaling, whose slices and arrays are JSON arrays of its
// values when marshaling, even with FormatBytesWithLegacySemantics: a value is a text of as many b as it is.
type legacyTextByte byte

func (b legacyTextByte) MarshalText() ([]byte, error) {
	return []byte(strings.Repeat("b", int(b))), nil
}

func (b *legacyTextByte) UnmarshalText(text []byte) error {
	*b = legacyTextByte(len(text))
	return nil
}

func TestMarshalFormatBytes(t *testing.T) {
	array := []json.Options{gojson.FormatByteArrayAsArray(true)}
	legacy := []json.Options{gojson.FormatBytesWithLegacySemantics(true)}
	both := append(array[:1:1], legacy...)
	runLegacyMarshalCases(t, []legacyMarshalCase{
		{"byte array as array", [3]byte{1, 2, 3}, array, `[1,2,3]`},
		{"byte array by default", [3]byte{1, 2, 3}, nil, `"AQID"`},
		{"byte array as array stringified", [2]byte{1, 2}, append(array, json.StringifyNumbers(true)), `["1","2"]`},
		{"byte slice with FormatByteArrayAsArray", []byte{1, 2, 3}, array, `"AQID"`},
		{"byte array key as array", map[[2]byte]int{{1, 2}: 1}, array, ``},
		{"byte array key", map[[2]byte]int{{1, 2}: 1}, nil, `{"AQI=":1}`},
		{"named bytes", []legacyByte{1, 2, 3}, legacy, `"AQID"`},
		{"named bytes by default", []legacyByte{1, 2, 3}, nil, `[1,2,3]`},
		{"nil named bytes", []legacyByte(nil), legacy, `""`},
		{"nil named bytes as null", []legacyByte(nil), append(legacy, json.FormatNilSliceAsNull(true)), `null`},
		{"named byte array", [2]legacyByte{1, 2}, legacy, `"AQI="`},
		{"named byte array as array", [2]legacyByte{1, 2}, both, `[1,2]`},
		{"named byte array key", map[[2]legacyByte]int{{1, 2}: 1}, legacy, `{"AQI=":1}`},
		{"text bytes", []legacyTextByte{1, 2}, legacy, `["b","bb"]`},
		{"text byte array", [2]legacyTextByte{1, 2}, legacy, `["b","bb"]`},
		{"named bytes in an interface", []any{[]legacyByte{1}}, legacy, `["AQ=="]`},
	})
}

func TestUnmarshalFormatBytes(t *testing.T) {
	array := []json.Options{gojson.FormatByteArrayAsArray(true)}
	legacy := []json.Options{gojson.FormatBytesWithLegacySemantics(true)}
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"byte array as array", `[1,2]`, array, newOf([2]byte{}), [2]byte{1, 2}},
		{"byte array as array of base64", `"AQI="`, array, newOf([2]byte{}), nil},
		{"byte array key as array", `{"AQI=":1}`, array, func() any { return new(map[[2]byte]int) }, nil},
		{"bytes of a JSON array", `[1,2,3]`, legacy, newOf([]byte(nil)), []byte{1, 2, 3}},
		{"bytes of a JSON array by default", `[1,2,3]`, nil, newOf([]byte(nil)), nil},
		{"byte array of a JSON array", `[1,2]`, legacy, newOf([2]byte{}), [2]byte{1, 2}},
		{"named bytes", `"AQID"`, legacy, newOf([]legacyByte(nil)), []legacyByte{1, 2, 3}},
		{"named bytes by default", `"AQID"`, nil, newOf([]legacyByte(nil)), nil},
		{"named byte array key", `{"AQI=":1}`, legacy, func() any { return new(map[[2]legacyByte]int) },
			map[[2]legacyByte]int{{1, 2}: 1}},
		{"text bytes of a JSON array", `["b","bb"]`, legacy, newOf([]legacyTextByte(nil)), []legacyTextByte{1, 2}},
		{"text bytes of base64", `"AQI="`, legacy, newOf([]legacyTextByte(nil)), []legacyTextByte{1, 2}},
	})
}

func TestMarshalOmitEmptyWithLegacySemantics(t *testing.T) {
	legacy := []json.Options{gojson.OmitEmptyWithLegacySemantics(true)}
	type omit struct {
		I  int            `json:",omitempty"`
		B  bool           `json:",omitempty"`
		P  *[]int         `json:",omitempty"`
		S  struct{}       `json:",omitempty"`
		M  map[string]int `json:",omitempty"`
		SB bool           `json:",omitempty,string"`
	}
	runLegacyMarshalCases(t, []legacyMarshalCase{
		{"zero", omit{}, legacy, `{"S":{}}`},
		{"zero by default", struct {
			I int      `json:",omitempty"`
			S struct{} `json:",omitempty"`
		}{}, nil, `{"I":0}`},
		{"empty", omit{P: &[]int{}, M: map[string]int{}}, legacy, `{"P":[],"S":{}}`},
		{"not empty", omit{I: 1, B: true, M: map[string]int{"a": 1}}, legacy, `{"I":1,"B":true,"S":{},"M":{"a":1}}`},
		{"string option of a bool", omit{SB: true}, legacy, ``},
	})
}

func TestMarshalStringifyWithLegacySemantics(t *testing.T) {
	stringify := []json.Options{gojson.StringifyWithLegacySemantics(true)}
	type values struct {
		B bool   `json:",string"`
		S string `json:",string"`
		I int    `json:",string"`
	}
	n := 1
	pn := &n
	runLegacyMarshalCases(t, []legacyMarshalCase{
		{"values", values{B: true, S: "x", I: 3}, stringify, `{"B":"true","S":"\"x\"","I":"3"}`},
		{"values by default", values{}, nil, ``},
		{"pointer to a pointer", struct {
			P **int `json:",string"`
		}{&pn}, stringify, ``},
	})
}

// legacyLitPtrJSON is a struct whose MarshalJSON has a pointer receiver, which writes "nil" for a nil pointer.
type legacyLitPtrJSON struct{ X int }

func (p *legacyLitPtrJSON) MarshalJSON() ([]byte, error) {
	if p == nil {
		return []byte(`"nil"`), nil
	}
	return []byte(`"called"`), nil
}

// legacyLitNode is a recursive struct of a value of legacyLitPtrJSON.
type legacyLitNode struct {
	V    legacyLitPtrJSON
	Next *legacyLitNode
}

// legacyLitPtrText is a struct whose MarshalText has a pointer receiver, which writes "nil" for a nil pointer.
type legacyLitPtrText struct{ X int }

func (p *legacyLitPtrText) MarshalText() ([]byte, error) {
	if p == nil {
		return []byte("nil"), nil
	}
	return []byte("text"), nil
}

// legacyLitValueText is a struct whose MarshalText has a value receiver, which can't be called for a nil pointer.
type legacyLitValueText struct{ X int }

func (legacyLitValueText) MarshalText() ([]byte, error) { return []byte("value"), nil }

// legacyLitPtrMap is a map whose MarshalJSON has a pointer receiver.
type legacyLitPtrMap map[string]int

func (*legacyLitPtrMap) MarshalJSON() ([]byte, error) { return []byte(`"map"`), nil }

// legacyLitRecSlice is a slice of itself, whose MarshalJSON has a pointer receiver: its elements are addressable.
type legacyLitRecSlice []legacyLitRecSlice

func (*legacyLitRecSlice) MarshalJSON() ([]byte, error) { return []byte(`"r"`), nil }

// legacyLitDeclined is a struct whose MarshalJSONTo, of a pointer receiver, declines to write it, before its
// MarshalJSON, of a pointer receiver too.
type legacyLitDeclined struct{ X int }

func (*legacyLitDeclined) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

func (*legacyLitDeclined) MarshalJSON() ([]byte, error) { return []byte(`"json"`), nil }

// legacyLitText is a string whose MarshalText is not called for a key of a map with CallMethodsWithLegacySemantics.
type legacyLitText string

func (legacyLitText) MarshalText() ([]byte, error) { return []byte("text"), nil }

func TestMarshalCallMethodsWithLegacySemantics(t *testing.T) {
	legacy := []json.Options{gojson.CallMethodsWithLegacySemantics(true), json.Deterministic(true)}
	// a function of another type, which makes the functions of marshaling be looked up
	withFuncs := append(legacy, json.WithMarshalers(json.MarshalFunc(func(time.Month) ([]byte, error) {
		return []byte(`"m"`), nil
	})))
	runLegacyMarshalCases(t, []legacyMarshalCase{
		{"value", legacyLitPtrJSON{}, legacy, `{"X":0}`},
		{"value by default", legacyLitPtrJSON{}, nil, `"called"`},
		{"pointer", &legacyLitPtrJSON{}, legacy, `"called"`},
		{"array", [1]legacyLitPtrJSON{}, legacy, `[{"X":0}]`},
		{"array of a pointer", &[1]legacyLitPtrJSON{}, legacy, `["called"]`},
		{"slice", []legacyLitPtrJSON{{}}, legacy, `["called"]`},
		{"map", map[string]legacyLitPtrJSON{"a": {}}, legacy, `{"a":{"X":0}}`},
		{"recursive", legacyLitNode{Next: &legacyLitNode{}}, legacy, `{"V":{"X":0},"Next":{"V":"called","Next":null}}`},
		{"key of a string kind", map[legacyLitText]int{"k": 1}, legacy, `{"k":1}`},
		{"key of a string kind by default", map[legacyLitText]int{"k": 1}, nil, `{"text":1}`},
		{"nil key", map[*legacyLitText]int{nil: 1}, legacy, `{"":1}`},
		{"interface of a nil pointer", struct{ M gojson.Marshaler }{(*legacyLitPtrJSON)(nil)}, legacy, `{"M":"nil"}`},
		{"interface of a nil pointer by default", struct{ M gojson.Marshaler }{(*legacyLitPtrJSON)(nil)}, nil, `{"M":null}`},
		{"pointer with functions", &legacyLitPtrJSON{}, withFuncs, `"called"`},
		{"pointer of an array with functions", &[1]legacyLitPtrJSON{}, withFuncs, `["called"]`},
		{"value with functions", legacyLitPtrJSON{}, withFuncs, `{"X":0}`},
		{"declined pointer", &legacyLitDeclined{}, legacy, `"json"`},
		{"declined value", legacyLitDeclined{}, legacy, `{"X":0}`},
		{"declined elements", []legacyLitDeclined{{}}, legacy, `["json"]`},
		{"key of an interface of a nil pointer", map[encoding.TextMarshaler]int{(*legacyLitPtrText)(nil): 1}, legacy, `{"nil":1}`},
		{"key of an interface of a nil pointer of a method of the value", map[encoding.TextMarshaler]int{(*legacyLitValueText)(nil): 1},
			legacy, ""},
		{"interface of a nil pointer of a method of the value", []encoding.TextMarshaler{(*legacyLitValueText)(nil)}, legacy, `[null]`},
		{"map", legacyLitPtrMap{"k": 1}, legacy, `{"k":1}`},
		{"map of a pointer", &legacyLitPtrMap{"k": 1}, legacy, `"map"`},
		{"map of a value of a map", map[string]legacyLitPtrMap{"a": {"k": 1}}, legacy, `{"a":{"k":1}}`},
		{"map of an interface", []any{legacyLitPtrMap{"k": 1}}, legacy, `[{"k":1}]`},
		{"map of an array", struct{ A [1]legacyLitPtrMap }{[1]legacyLitPtrMap{{"k": 1}}}, legacy, `{"A":[{"k":1}]}`},
		{"map of an element", []legacyLitPtrMap{{"k": 1}}, legacy, `["map"]`},
		{"map of a pointer field", struct{ F *legacyLitPtrMap }{&legacyLitPtrMap{"k": 1}}, legacy, `{"F":"map"}`},
		{"map of pointer elements", []*legacyLitPtrMap{{"k": 1}}, legacy, `["map"]`},
		{"recursive slice", legacyLitRecSlice{legacyLitRecSlice{}, nil}, legacy, `["r","r"]`},
		{"recursive slice by default", legacyLitRecSlice{legacyLitRecSlice{}, nil}, nil, `"r"`},
	})
}

// legacyLitFailingText is a struct whose MarshalText fails, of a value receiver.
type legacyLitFailingText struct{ X int }

func (legacyLitFailingText) MarshalText() ([]byte, error) { return nil, errors.New("text failed") }

// legacyLitConstKey is an int whose MarshalText names every key of a map alike.
type legacyLitConstKey int

func (legacyLitConstKey) MarshalText() ([]byte, error) { return []byte("k"), nil }

// legacyLitFailingWriter is a writer which fails.
type legacyLitFailingWriter struct{}

var errLegacyLitWrite = errors.New("write failed")

func (legacyLitFailingWriter) Write([]byte) (int, error) { return 0, errLegacyLitWrite }

// With ReportErrorsWithLegacySemantics, the errors of marshaling are the ones of encoding/json, and the errors of the
// definitions of structs and the `string` option of the values it doesn't apply to are not errors.
func TestMarshalReportErrorsWithLegacySemantics(t *testing.T) {
	legacy := gojson.ReportErrorsWithLegacySemantics(true)
	marshal := func(v any) error {
		_, err := json.Marshal(v, legacy)
		return err
	}
	var typeErr *gojson.UnsupportedTypeError
	if err := marshal(struct{ C chan int }{}); !errors.As(err, &typeErr) || typeErr.Type.String() != "chan int" {
		t.Errorf("chan: got %v, want an UnsupportedTypeError of chan int", err)
	}
	var valueErr *gojson.UnsupportedValueError
	if err := marshal(math.NaN()); !errors.As(err, &valueErr) || valueErr.Str != "NaN" || !math.IsNaN(valueErr.Value.Float()) {
		t.Errorf("NaN: got %v, want an UnsupportedValueError of NaN", err)
	}
	for _, v := range []any{map[float64]int{math.Inf(-1): 1}, map[any]int{math.Inf(-1): 1}} {
		if err := marshal(v); !errors.As(err, &valueErr) || valueErr.Str != "-Inf" || !math.IsInf(valueErr.Value.Float(), -1) {
			t.Errorf("key of -Inf of %T: got %v, want an UnsupportedValueError of -Inf", v, err)
		}
	}
	var marshalerErr *gojson.MarshalerError
	if err := marshal([]any{legacyLitFailingText{}}); !errors.As(err, &marshalerErr) ||
		marshalerErr.Type.String() != "*json_test.legacyLitFailingText" ||
		!strings.Contains(err.Error(), "error calling MarshalText for type") {
		t.Errorf("MarshalText: got %v, want a MarshalerError of MarshalText of the pointer", err)
	}
	var syntaxErr *gojson.SyntaxError
	if err := marshal("\xff"); !errors.As(err, &syntaxErr) || syntaxErr.Offset != 1 {
		t.Errorf("invalid UTF-8: got %v, want a SyntaxError after the string", err)
	}
	if err := marshal(map[legacyLitConstKey]int{1: 1, 2: 2}); !errors.As(err, &syntaxErr) {
		t.Errorf("duplicate names of a method: got %v, want a SyntaxError", err)
	}
	if err := json.MarshalEncode(jsontext.NewEncoder(legacyLitFailingWriter{}), make([]int, 10000), legacy); err != errLegacyLitWrite {
		t.Errorf("MarshalEncode to a failing writer: got %v, want its error", err)
	}
	if err := json.MarshalWrite(legacyLitFailingWriter{}, 1, legacy); err != errLegacyLitWrite {
		t.Errorf("MarshalWrite to a failing writer: got %v, want its error", err)
	}
	runLegacyMarshalCases(t, []legacyMarshalCase{
		{"conflicting names", struct {
			A int `json:"x"`
			B int `json:"x"`
			C int
		}{}, []json.Options{legacy}, `{"C":0}`},
		{"conflicting names by default", struct {
			A int `json:"x"`
			B int `json:"x"`
			C int
		}{}, nil, ""},
		{"string option of a bool", struct {
			B bool `json:",string"`
		}{}, []json.Options{legacy}, `{"B":false}`},
		{"string option of an interface", struct {
			I any `json:",string"`
		}{1}, []json.Options{legacy}, `{"I":"1"}`},
		{"string option of an interface with StringifyWithLegacySemantics", struct {
			I any `json:",string"`
		}{1}, []json.Options{legacy, gojson.StringifyWithLegacySemantics(true)}, `{"I":1}`},
		{"string option of a bool by default", struct {
			B bool `json:",string"`
		}{}, nil, ""},
	})
}

func TestMarshalFormatDurationAsNano(t *testing.T) {
	nano := []json.Options{gojson.FormatDurationAsNano(true)}
	d := time.Duration(-3)
	runLegacyMarshalCases(t, []legacyMarshalCase{
		{"value", time.Duration(1500), nano, `1500`},
		{"value by default", time.Duration(1500), nil, ``},
		{"pointer", &d, nano, `-3`},
		{"stringified", time.Duration(1500), append(nano, json.StringifyNumbers(true)), `"1500"`},
		{"key", map[time.Duration]int{5: 1, -2: 2}, nano, `{"-2":2,"5":1}`},
		{"key by default", map[time.Duration]int{5: 1}, nil, ``},
		{"string option", struct {
			S time.Duration `json:",string"`
		}{2}, nano, `{"S":"2"}`},
		{"string option by default", struct {
			S time.Duration `json:",string"`
		}{2}, nil, ``},
		{"in an interface", []any{time.Duration(7)}, nano, `[7]`},
	})
}

func TestUnmarshalFormatDurationAsNano(t *testing.T) {
	nano := []json.Options{gojson.FormatDurationAsNano(true)}
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"value", `1500`, nano, newOf(time.Duration(0)), time.Duration(1500)},
		{"value by default", `1500`, nil, newOf(time.Duration(0)), nil},
		{"fraction", `1.5`, nano, newOf(time.Duration(0)), nil},
		{"string", `"1500"`, nano, newOf(time.Duration(0)), nil},
		{"stringified", `"1500"`, append(nano, json.StringifyNumbers(true)), newOf(time.Duration(0)), time.Duration(1500)},
		{"key", `{"5":1}`, nano, func() any { return new(map[time.Duration]int) }, map[time.Duration]int{5: 1}},
		{"string option", `{"S":"3"}`, nano, func() any {
			return new(struct {
				S time.Duration `json:",string"`
			})
		}, struct {
			S time.Duration `json:",string"`
		}{3}},
	})
}

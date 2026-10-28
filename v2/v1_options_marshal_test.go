package json_test

import (
	"encoding"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The options of encoding/json when marshaling, on every version of Go, with the outputs which encoding/json/v2 of
// Go 1.27 writes ( see std_v1_options_diff_test.go, which compares with it ).

type v1MarshalCase struct {
	name string
	in   any
	opts []json.Options
	// want is the output, or "" for an error
	want string
}

func runV1MarshalCases(t *testing.T, cases []v1MarshalCase) {
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

// v1Byte is a named byte type, whose slices and arrays are of bytes only with FormatBytesWithLegacySemantics.
type v1Byte byte

// v1TextByte is a named byte type with methods of marshaling, whose slices and arrays are JSON arrays of its values
// when marshaling, even with FormatBytesWithLegacySemantics: a value is a text of as many b as it is.
type v1TextByte byte

func (b v1TextByte) MarshalText() ([]byte, error) {
	return []byte(strings.Repeat("b", int(b))), nil
}

func (b *v1TextByte) UnmarshalText(text []byte) error {
	*b = v1TextByte(len(text))
	return nil
}

func TestMarshalFormatBytes(t *testing.T) {
	array := []json.Options{gojson.FormatByteArrayAsArray(true)}
	formatBytes := []json.Options{gojson.FormatBytesWithLegacySemantics(true)}
	both := []json.Options{gojson.FormatByteArrayAsArray(true), gojson.FormatBytesWithLegacySemantics(true)}
	runV1MarshalCases(t, []v1MarshalCase{
		{"byte array as array", [3]byte{1, 2, 3}, array, `[1,2,3]`},
		{"byte array by default", [3]byte{1, 2, 3}, nil, `"AQID"`},
		{"byte array as array stringified", [2]byte{1, 2}, append(array, json.StringifyNumbers(true)), `["1","2"]`},
		{"byte slice with FormatByteArrayAsArray", []byte{1, 2, 3}, array, `"AQID"`},
		{"byte array key as array", map[[2]byte]int{{1, 2}: 1}, array, ``},
		{"byte array key", map[[2]byte]int{{1, 2}: 1}, nil, `{"AQI=":1}`},
		{"named bytes", []v1Byte{1, 2, 3}, formatBytes, `"AQID"`},
		{"named bytes by default", []v1Byte{1, 2, 3}, nil, `[1,2,3]`},
		{"nil named bytes", []v1Byte(nil), formatBytes, `""`},
		{"nil named bytes as null", []v1Byte(nil), append(formatBytes, json.FormatNilSliceAsNull(true)), `null`},
		{"named byte array", [2]v1Byte{1, 2}, formatBytes, `"AQI="`},
		{"named byte array as array", [2]v1Byte{1, 2}, both, `[1,2]`},
		{"named byte array key", map[[2]v1Byte]int{{1, 2}: 1}, formatBytes, `{"AQI=":1}`},
		{"text bytes", []v1TextByte{1, 2}, formatBytes, `["b","bb"]`},
		{"text byte array", [2]v1TextByte{1, 2}, formatBytes, `["b","bb"]`},
		{"named bytes in an interface", []any{[]v1Byte{1}}, formatBytes, `["AQ=="]`},
	})
}

func TestUnmarshalFormatBytes(t *testing.T) {
	array := []json.Options{gojson.FormatByteArrayAsArray(true)}
	formatBytes := []json.Options{gojson.FormatBytesWithLegacySemantics(true)}
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"byte array as array", `[1,2]`, array, outOf([2]byte{}), [2]byte{1, 2}},
		{"byte array as array of base64", `"AQI="`, array, outOf([2]byte{}), nil},
		{"byte array key as array", `{"AQI=":1}`, array, outOf(map[[2]byte]int(nil)), nil},
		{"bytes of a JSON array", `[1,2,3]`, formatBytes, outOf([]byte(nil)), []byte{1, 2, 3}},
		{"bytes of a JSON array by default", `[1,2,3]`, nil, outOf([]byte(nil)), nil},
		{"byte array of a JSON array", `[1,2]`, formatBytes, outOf([2]byte{}), [2]byte{1, 2}},
		{"named bytes", `"AQID"`, formatBytes, outOf([]v1Byte(nil)), []v1Byte{1, 2, 3}},
		{"named bytes by default", `"AQID"`, nil, outOf([]v1Byte(nil)), nil},
		{"named byte array key", `{"AQI=":1}`, formatBytes, outOf(map[[2]v1Byte]int(nil)),
			map[[2]v1Byte]int{{1, 2}: 1}},
		{"text bytes of a JSON array", `["b","bb"]`, formatBytes, outOf([]v1TextByte(nil)), []v1TextByte{1, 2}},
		{"text bytes of base64", `"AQI="`, formatBytes, outOf([]v1TextByte(nil)), []v1TextByte{1, 2}},
	})
}

func TestMarshalOmitEmptyWithLegacySemantics(t *testing.T) {
	omitEmpty := []json.Options{gojson.OmitEmptyWithLegacySemantics(true)}
	type omit struct {
		I  int            `json:",omitempty"`
		B  bool           `json:",omitempty"`
		P  *[]int         `json:",omitempty"`
		S  struct{}       `json:",omitempty"`
		M  map[string]int `json:",omitempty"`
		SB bool           `json:",omitempty,string"`
	}
	runV1MarshalCases(t, []v1MarshalCase{
		{"zero", omit{}, omitEmpty, `{"S":{}}`},
		{"zero by default", struct {
			I int      `json:",omitempty"`
			S struct{} `json:",omitempty"`
		}{}, nil, `{"I":0}`},
		{"empty", omit{P: &[]int{}, M: map[string]int{}}, omitEmpty, `{"P":[],"S":{}}`},
		{"not empty", omit{I: 1, B: true, M: map[string]int{"a": 1}}, omitEmpty, `{"I":1,"B":true,"S":{},"M":{"a":1}}`},
		{"string option of a bool", omit{SB: true}, omitEmpty, ``},
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
	runV1MarshalCases(t, []v1MarshalCase{
		{"values", values{B: true, S: "x", I: 3}, stringify, `{"B":"true","S":"\"x\"","I":"3"}`},
		{"values by default", values{}, nil, ``},
		{"pointer to a pointer", struct {
			P **int `json:",string"`
		}{&pn}, stringify, ``},
	})
}

// v1PtrMarshaler is a struct whose MarshalJSON has a pointer receiver, which CallMethodsWithLegacySemantics calls
// for an addressable value only, and which writes "nil" for a nil pointer.
type v1PtrMarshaler struct{ X int }

func (p *v1PtrMarshaler) MarshalJSON() ([]byte, error) {
	if p == nil {
		return []byte(`"nil"`), nil
	}
	return []byte(`"called"`), nil
}

// v1Node is a recursive struct of a value of v1PtrMarshaler, which is addressable through the pointer.
type v1Node struct {
	V    v1PtrMarshaler
	Next *v1Node
}

// v1PtrText is a struct whose MarshalText has a pointer receiver, which writes "nil" for a nil pointer.
type v1PtrText struct{ X int }

func (p *v1PtrText) MarshalText() ([]byte, error) {
	if p == nil {
		return []byte("nil"), nil
	}
	return []byte("text"), nil
}

// v1ValueText is a struct whose MarshalText has a value receiver, which can't be called for a nil pointer.
type v1ValueText struct{ X int }

func (v1ValueText) MarshalText() ([]byte, error) { return []byte("value"), nil }

// v1PtrMap is a map whose MarshalJSON has a pointer receiver: a map is stored directly in an interface value, and
// its opcodes read it by a pointer to it where it is not addressable.
type v1PtrMap map[string]int

func (*v1PtrMap) MarshalJSON() ([]byte, error) { return []byte(`"map"`), nil }

// v1RecSlice is a slice of itself, whose MarshalJSON has a pointer receiver: its elements are addressable.
type v1RecSlice []v1RecSlice

func (*v1RecSlice) MarshalJSON() ([]byte, error) { return []byte(`"r"`), nil }

// v1Declined is a struct whose MarshalJSONTo, of a pointer receiver, declines to write it, before its
// MarshalJSON, of a pointer receiver too.
type v1Declined struct{ X int }

func (*v1Declined) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

func (*v1Declined) MarshalJSON() ([]byte, error) { return []byte(`"json"`), nil }

// v1StringText is a string whose MarshalText, of a value receiver, is not called for a key of a map with
// CallMethodsWithLegacySemantics, as the key is of a string kind.
type v1StringText string

func (v1StringText) MarshalText() ([]byte, error) { return []byte("text"), nil }

func TestMarshalCallMethodsWithLegacySemantics(t *testing.T) {
	callMethods := []json.Options{gojson.CallMethodsWithLegacySemantics(true)}
	// a function of another type, which makes the functions of marshaling be looked up
	withFuncs := []json.Options{gojson.CallMethodsWithLegacySemantics(true), json.WithMarshalers(json.MarshalFunc(func(time.Month) ([]byte, error) {
		return []byte(`"m"`), nil
	}))}
	runV1MarshalCases(t, []v1MarshalCase{
		{"value", v1PtrMarshaler{}, callMethods, `{"X":0}`},
		{"value by default", v1PtrMarshaler{}, nil, `"called"`},
		{"pointer", &v1PtrMarshaler{}, callMethods, `"called"`},
		{"array", [1]v1PtrMarshaler{}, callMethods, `[{"X":0}]`},
		{"array of a pointer", &[1]v1PtrMarshaler{}, callMethods, `["called"]`},
		{"slice", []v1PtrMarshaler{{}}, callMethods, `["called"]`},
		{"map", map[string]v1PtrMarshaler{"a": {}}, callMethods, `{"a":{"X":0}}`},
		{"recursive", v1Node{Next: &v1Node{}}, callMethods, `{"V":{"X":0},"Next":{"V":"called","Next":null}}`},
		{"key of a string kind", map[v1StringText]int{"k": 1}, callMethods, `{"k":1}`},
		{"key of a string kind by default", map[v1StringText]int{"k": 1}, nil, `{"text":1}`},
		{"nil key", map[*v1StringText]int{nil: 1}, callMethods, `{"":1}`},
		{"interface of a nil pointer", struct{ M gojson.Marshaler }{(*v1PtrMarshaler)(nil)}, callMethods, `{"M":"nil"}`},
		{"interface of a nil pointer by default", struct{ M gojson.Marshaler }{(*v1PtrMarshaler)(nil)}, nil, `{"M":null}`},
		{"pointer with functions", &v1PtrMarshaler{}, withFuncs, `"called"`},
		{"pointer of an array with functions", &[1]v1PtrMarshaler{}, withFuncs, `["called"]`},
		{"value with functions", v1PtrMarshaler{}, withFuncs, `{"X":0}`},
		{"declined pointer", &v1Declined{}, callMethods, `"json"`},
		{"declined value", v1Declined{}, callMethods, `{"X":0}`},
		{"declined elements", []v1Declined{{}}, callMethods, `["json"]`},
		{"key of an interface of a nil pointer", map[encoding.TextMarshaler]int{(*v1PtrText)(nil): 1}, callMethods, `{"nil":1}`},
		{"key of an interface of a nil pointer of a method of the value", map[encoding.TextMarshaler]int{(*v1ValueText)(nil): 1},
			callMethods, ""},
		{"interface of a nil pointer of a method of the value", []encoding.TextMarshaler{(*v1ValueText)(nil)}, callMethods, `[null]`},
		{"map", v1PtrMap{"k": 1}, callMethods, `{"k":1}`},
		{"map of a pointer", &v1PtrMap{"k": 1}, callMethods, `"map"`},
		{"map of a value of a map", map[string]v1PtrMap{"a": {"k": 1}}, callMethods, `{"a":{"k":1}}`},
		{"map of an interface", []any{v1PtrMap{"k": 1}}, callMethods, `[{"k":1}]`},
		{"map of an array", struct{ A [1]v1PtrMap }{[1]v1PtrMap{{"k": 1}}}, callMethods, `{"A":[{"k":1}]}`},
		{"map of an element", []v1PtrMap{{"k": 1}}, callMethods, `["map"]`},
		{"map of a pointer field", struct{ F *v1PtrMap }{&v1PtrMap{"k": 1}}, callMethods, `{"F":"map"}`},
		{"map of pointer elements", []*v1PtrMap{{"k": 1}}, callMethods, `["map"]`},
		{"recursive slice", v1RecSlice{v1RecSlice{}, nil}, callMethods, `["r","r"]`},
		{"recursive slice by default", v1RecSlice{v1RecSlice{}, nil}, nil, `"r"`},
	})
}

// reportErrText is a struct whose MarshalText fails, of a value receiver.
type reportErrText struct{ X int }

func (reportErrText) MarshalText() ([]byte, error) { return nil, errors.New("text failed") }

// reportConstKey is an int whose MarshalText names every key of a map alike.
type reportConstKey int

func (reportConstKey) MarshalText() ([]byte, error) { return []byte("k"), nil }

// v1LitFailingWriter is a writer which fails.
type v1LitFailingWriter struct{}

var errV1LitWrite = errors.New("write failed")

func (v1LitFailingWriter) Write([]byte) (int, error) { return 0, errV1LitWrite }

// With ReportErrorsWithLegacySemantics, the errors of marshaling are the ones of encoding/json, and the errors of the
// definitions of structs and the `string` option of the values it doesn't apply to are not errors.
func TestMarshalReportErrorsWithLegacySemantics(t *testing.T) {
	reportErrors := gojson.ReportErrorsWithLegacySemantics(true)
	marshal := func(v any) error {
		_, err := json.Marshal(v, reportErrors)
		return err
	}
	var typeErr *gojson.UnsupportedTypeError
	if err := marshal(struct{ C chan int }{}); !errors.As(err, &typeErr) || typeErr.Type != reflect.TypeFor[chan int]() {
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
	if err := marshal([]any{reportErrText{}}); !errors.As(err, &marshalerErr) || marshalerErr.Type != reflect.TypeFor[*reportErrText]() ||
		!strings.Contains(err.Error(), "error calling MarshalText for type") {
		t.Errorf("MarshalText: got %v, want a MarshalerError of MarshalText of the pointer", err)
	}
	var syntaxErr *gojson.SyntaxError
	if err := marshal("\xff"); !errors.As(err, &syntaxErr) || syntaxErr.Offset != 1 {
		t.Errorf("invalid UTF-8: got %v, want a SyntaxError after the string", err)
	}
	if err := marshal(map[reportConstKey]int{1: 1, 2: 2}); !errors.As(err, &syntaxErr) {
		t.Errorf("duplicate names of a method: got %v, want a SyntaxError", err)
	}
	// a year of five digits, which RFC 3339 has no text of
	if err := marshal(map[time.Time]int{time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC): 1}); !errors.As(err, &marshalerErr) ||
		marshalerErr.Type != reflect.TypeFor[*time.Time]() || !strings.Contains(err.Error(), "error calling MarshalJSON for type") {
		t.Errorf("time key out of range: got %v, want a MarshalerError of MarshalJSON of *time.Time", err)
	}
	// MarshalEncode writes to the writer after the whole value is marshaled, as MarshalWrite does
	if err := json.MarshalEncode(jsontext.NewEncoder(v1LitFailingWriter{}), 1, reportErrors); err != errV1LitWrite {
		t.Errorf("MarshalEncode to a failing writer: got %v, want its error", err)
	}
	if err := json.MarshalWrite(v1LitFailingWriter{}, 1, reportErrors); err != errV1LitWrite {
		t.Errorf("MarshalWrite to a failing writer: got %v, want its error", err)
	}
	opts := []json.Options{reportErrors}
	runV1MarshalCases(t, []v1MarshalCase{
		{"conflicting names", struct {
			A int `json:"x"`
			B int `json:"x"`
			C int
		}{}, opts, `{"C":0}`},
		{"conflicting names by default", struct {
			A int `json:"x"`
			B int `json:"x"`
			C int
		}{}, nil, ""},
		{"string option of a bool", struct {
			B bool `json:",string"`
		}{}, opts, `{"B":false}`},
		{"string option of an interface", struct {
			I any `json:",string"`
		}{1}, opts, `{"I":"1"}`},
		{"string option of an interface with StringifyWithLegacySemantics", struct {
			I any `json:",string"`
		}{1}, append(opts, gojson.StringifyWithLegacySemantics(true)), `{"I":1}`},
		{"string option of an interface of an array", struct {
			I any `json:",string"`
		}{[]any{1, &[]int{2}}}, opts, `{"I":[1,[2]]}`},
		{"string option of an interface of a pointer", struct {
			I any `json:",string"`
		}{new(int)}, opts, `{"I":"0"}`},
		{"string option of a bool by default", struct {
			B bool `json:",string"`
		}{}, nil, ""},
	})
}

func TestMarshalFormatDurationAsNano(t *testing.T) {
	nano := []json.Options{gojson.FormatDurationAsNano(true)}
	d := time.Duration(-3)
	runV1MarshalCases(t, []v1MarshalCase{
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
	type stringOption = struct {
		S time.Duration `json:",string"`
	}
	runV1UnmarshalCases(t, []v1UnmarshalCase{
		{"value", `1500`, nano, outOf(time.Duration(0)), time.Duration(1500)},
		{"value by default", `1500`, nil, outOf(time.Duration(0)), nil},
		{"fraction", `1.5`, nano, outOf(time.Duration(0)), nil},
		{"string", `"1500"`, nano, outOf(time.Duration(0)), nil},
		{"stringified", `"1500"`, append(nano, json.StringifyNumbers(true)), outOf(time.Duration(0)), time.Duration(1500)},
		{"key", `{"5":1}`, nano, outOf(map[time.Duration]int(nil)), map[time.Duration]int{5: 1}},
		{"string option", `{"S":"3"}`, nano, outOf(stringOption{}), stringOption{3}},
	})
}

package json_test

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// The type errors of the decoding are the ones of encoding/json of the running Go: a value which is not of the
// type of its destination is skipped, the decoding goes on, and the first such error is returned, unless the
// input has a syntax error, which is returned instead. Its Value, Type, Offset, Struct and Field, and its
// message, are the ones of encoding/json, whose rules differ between Go 1.27, which is made of
// encoding/json/v2, and the earlier versions.

type typeErrorInner struct {
	F int
	G string
	H float64
}

type typeErrorEmbedded struct {
	E int
}

type typeErrorOuter struct {
	typeErrorEmbedded
	A      typeErrorInner
	L      []typeErrorInner
	M      map[string]typeErrorInner
	P      *typeErrorInner
	X      int
	U      uint8
	S      []int
	Arr    [2]int
	B      bool
	Str    string
	Bytes  []byte
	C      complex128
	N      json.Number
	Tagged int     `json:"tagged"`
	Q      int     `json:"q,string"`
	Q8     int8    `json:"q8,string"`
	QB     bool    `json:"qb,string"`
	QF     float32 `json:"qf,string"`
	QS     string  `json:"qs,string"`
	QP     *int    `json:"qp,string"`
	Last   int
}

// typeErrorValues are the JSON values which are put where a value of each field is.
var typeErrorValues = []string{
	`1`, `-1`, `1.5`, `1e3`, `300`, `"s"`, `"1"`, `""`, `"true"`, `"1.5"`, `"300"`, `"99999999999999999999"`,
	`true`, `null`, `[1]`, `[]`, `{"F":1}`, `{}`,
}

// typeErrorPlaces are the places of a value in a document of typeErrorOuter, %s being the value.
var typeErrorPlaces = []string{
	`{"X":%s,"Last":1}`,
	`{"U":%s,"Last":1}`,
	`{"E":%s,"Last":1}`,
	`{"A":{"F":%s,"G":"g"},"Last":1}`,
	`{"A":{"G":%s},"Last":1}`,
	`{"A":{"H":%s},"Last":1}`,
	`{"A":%s,"Last":1}`,
	`{"L":[{"F":1},{"F":%s}],"Last":1}`,
	`{"L":%s,"Last":1}`,
	`{"M":{"k":{"G":%s}},"Last":1}`,
	`{"M":{"k":%s},"Last":1}`,
	`{"M":%s,"Last":1}`,
	`{"P":{"F":%s},"Last":1}`,
	`{"P":%s,"Last":1}`,
	`{"S":[1,%s,3],"Last":1}`,
	`{"S":%s,"Last":1}`,
	`{"Arr":[%s,2],"Last":1}`,
	`{"Arr":%s,"Last":1}`,
	`{"B":%s,"Last":1}`,
	`{"Str":%s,"Last":1}`,
	`{"Bytes":%s,"Last":1}`,
	`{"C":%s,"Last":1}`,
	`{"N":%s,"Last":1}`,
	`{"tagged":%s,"Last":1}`,
	`{"q":%s,"Last":1}`,
	`{"q8":%s,"Last":1}`,
	`{"qb":%s,"Last":1}`,
	`{"qf":%s,"Last":1}`,
	`{"qs":%s,"Last":1}`,
	`{"qp":%s,"Last":1}`,
	`{"X":%s,"A":{"F":"first"},"Last":1}`,
	`{"A":{"F":"first"},"X":%s,"Last":1}`,
}

// describeTypeError returns what the test compares of an error: its kind, and the fields and the message of a
// type error.
func describeTypeError(err error) string {
	if err == nil {
		return "<nil>"
	}
	var stdErr *stdjson.UnmarshalTypeError
	if errors.As(err, &stdErr) {
		return fmt.Sprintf("UnmarshalTypeError Value=%q Type=%v Offset=%d Struct=%q Field=%q: %s",
			stdErr.Value, stdErr.Type, stdErr.Offset, stdErr.Struct, stdErr.Field, err)
	}
	var goErr *json.UnmarshalTypeError
	if errors.As(err, &goErr) {
		return fmt.Sprintf("UnmarshalTypeError Value=%q Type=%v Offset=%d Struct=%q Field=%q: %s",
			goErr.Value, goErr.Type, goErr.Offset, goErr.Struct, goErr.Field, err)
	}
	var stdSyntax *stdjson.SyntaxError
	var goSyntax *json.SyntaxError
	if errors.As(err, &stdSyntax) || errors.As(err, &goSyntax) {
		// the messages of the syntax errors are not compared here
		return "SyntaxError"
	}
	return err.Error()
}

func checkTypeError(t *testing.T, doc string, newValue func() any) {
	t.Helper()
	want, got := newValue(), newValue()
	wantErr := describeTypeError(stdjson.Unmarshal([]byte(doc), want))
	gotErr := describeTypeError(json.Unmarshal([]byte(doc), got))
	if gotErr != wantErr {
		t.Errorf("%s:\n got %s\nwant %s", doc, gotErr, wantErr)
		return
	}
	if wantErr == "SyntaxError" {
		// encoding/json validates the whole input before it decodes anything, and go-json decodes the values
		// before a syntax error: the values decoded before it differ.
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: decoded %+v, want %+v", doc, reflect.ValueOf(got).Elem(), reflect.ValueOf(want).Elem())
	}
}

func TestDecodeTypeErrors(t *testing.T) {
	for _, place := range typeErrorPlaces {
		for _, value := range typeErrorValues {
			checkTypeError(t, fmt.Sprintf(place, value), func() any { return new(typeErrorOuter) })
		}
	}
}

func TestDecodeTypeErrorsOfValues(t *testing.T) {
	// A value which is not in a struct, of each kind, as a value of a slice, of a map and of an interface.
	targets := []func() any{
		func() any { return new(int) }, func() any { return new(uint) }, func() any { return new(int8) },
		func() any { return new(float32) }, func() any { return new(float64) }, func() any { return new(bool) },
		func() any { return new(string) }, func() any { return new([]int) }, func() any { return new([2]int) },
		func() any { return new(map[string]int) }, func() any { return new(map[int]int) },
		func() any { return new(typeErrorInner) }, func() any { return new([]byte) },
		func() any { return new(complex128) }, func() any { return new(json.Number) },
		func() any { return new(*int) }, func() any { return new([]typeErrorInner) },
		func() any { return new(map[string]typeErrorInner) }, func() any { return new(any) },
	}
	for _, newValue := range targets {
		for _, value := range typeErrorValues {
			checkTypeError(t, value, newValue)
			checkTypeError(t, " "+value+" ", newValue)
		}
	}
	for _, doc := range []string{
		`{"x":1}`, `{"1":1,"x":2,"3":3}`, `{"300":1}`, `{"-1":1}`, `{"1.5":1}`, `{"true":1,"false":2}`, `{}`,
	} {
		checkTypeError(t, doc, func() any { return new(map[int8]int) })
		checkTypeError(t, doc, func() any { return new(map[uint]int) })
		checkTypeError(t, doc, func() any { return new(map[float64]int) })
		checkTypeError(t, doc, func() any { return new(map[bool]int) })
		checkTypeError(t, doc, func() any { return new(map[typeErrorEmbedded]int) })
	}
}

func TestDecodeTypeErrorsBeforeSyntaxErrors(t *testing.T) {
	// A syntax error anywhere in the input is returned instead of a type error before it.
	for _, doc := range []string{
		`{"X":"s","Last":1,}`,
		`{"X":"s","Last":1} x`,
		`{"X":"s","A":{"F":1,}}`,
		`{"X":"s","L":[1 2]}`,
		`{"X":"s","Str":"\x"}`,
		`{"A":{"F":"s"},"X":tru}`,
	} {
		checkTypeError(t, doc, func() any { return new(typeErrorOuter) })
	}
}

func TestDecodeTypeErrorsOfStream(t *testing.T) {
	// The decoder of a stream returns the type error of a value, and decodes the next ones.
	const input = `{"X":1} {"X":"s","Last":2} {"A":{"F":true}} {"X":3}`
	std := stdjson.NewDecoder(strings.NewReader(input))
	dec := json.NewDecoder(strings.NewReader(input))
	for i := 0; i < 4; i++ {
		var want, got typeErrorOuter
		wantErr := describeTypeError(std.Decode(&want))
		gotErr := describeTypeError(dec.Decode(&got))
		if gotErr != wantErr {
			t.Errorf("value %d:\n got %s\nwant %s", i, gotErr, wantErr)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("value %d: decoded %+v, want %+v", i, got, want)
		}
	}
}

func TestDecodeTypeErrorsNotKept(t *testing.T) {
	// A type error of a decoding which fails with a syntax error is not kept for the next decoding, which reuses
	// the context.
	for i := 0; i < 100; i++ {
		var v typeErrorOuter
		if err := json.Unmarshal([]byte(`{"X":"s","Last":1,}`), &v); err == nil {
			t.Fatal("expected a syntax error")
		}
		if err := json.Unmarshal([]byte(`{"X":1,"Last":1}`), &v); err != nil {
			t.Fatalf("got %v after a syntax error", err)
		}
		if err := json.UnmarshalOf([]byte(`{"X":"s","Last":1} x`), &v); err == nil {
			t.Fatal("expected a syntax error")
		}
		if err := json.UnmarshalOf([]byte(`{"X":1,"Last":1}`), &v); err != nil {
			t.Fatalf("got %v after a syntax error", err)
		}
	}
}

// typeErrorMethods is an interface type with methods, whose value is decoded into what it points to if it is a
// pointer, and is kept else.
type typeErrorMethods interface {
	Method()
}

type typeErrorJSONValue struct{ Raw string }

func (v *typeErrorJSONValue) Method() {}

func (v *typeErrorJSONValue) UnmarshalJSON(b []byte) error {
	v.Raw = string(b)
	return nil
}

type typeErrorTextValue struct{ Text string }

func (v *typeErrorTextValue) Method() {}

func (v *typeErrorTextValue) UnmarshalText(b []byte) error {
	v.Text = string(b)
	return nil
}

type typeErrorPlainValue struct{ A int }

func (v *typeErrorPlainValue) Method() {}

type typeErrorValueType struct{ A int }

func (v typeErrorValueType) Method() {}

func TestDecodeTypeErrorsOfInterfaceValues(t *testing.T) {
	// The value which an interface value holds is decoded into what it points to if it is a pointer, with the
	// unmarshalers of the pointer type, whatever the interface type is. A value of an interface type with
	// methods which is not a pointer is kept: null sets it to nil, and anything else is a type error.
	var nilPlain *typeErrorPlainValue
	values := []func() typeErrorMethods{
		func() typeErrorMethods { return nil },
		func() typeErrorMethods { return &typeErrorJSONValue{} },
		func() typeErrorMethods { return &typeErrorTextValue{} },
		func() typeErrorMethods { return &typeErrorPlainValue{A: 9} },
		func() typeErrorMethods { return typeErrorValueType{A: 9} },
		func() typeErrorMethods { return nilPlain },
	}
	type holder struct {
		V    typeErrorMethods
		Last int
	}
	type anyHolder struct {
		V    any
		Last int
	}
	for round := 0; round < 2; round++ {
		for _, value := range []string{`{"A":1}`, `null`, `"text"`, `[1]`, `1`, `true`, ` false`, ` 12 `, ` {}`} {
			for _, newValue := range values {
				checkTypeError(t, `{"V":`+value+`,"Last":1}`, func() any { return &holder{V: newValue()} })
				checkTypeError(t, `{"V":`+value+`,"Last":1}`, func() any { return &anyHolder{V: newValue()} })
			}
		}
	}
}

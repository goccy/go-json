package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/goccy/go-json"
)

// mapKeyKindUnmarshalJSON is an integer with a method of unmarshaling JSON, which a key of a map doesn't use.
type mapKeyKindUnmarshalJSON int

func (k *mapKeyKindUnmarshalJSON) UnmarshalJSON(b []byte) error {
	n, err := strconv.Atoi(string(b))
	*k = mapKeyKindUnmarshalJSON(n * 10)
	return err
}

type mapKeyKindText string

func (k *mapKeyKindText) UnmarshalText(b []byte) error {
	*k = mapKeyKindText("t" + string(b))
	return nil
}

// mapKeyKindError returns the kind of the error, which both packages report by types of their own.
func mapKeyKindError(err error) string {
	var syntax *json.SyntaxError
	var stdSyntax *stdjson.SyntaxError
	var typ *json.UnmarshalTypeError
	var stdTyp *stdjson.UnmarshalTypeError
	switch {
	case err == nil:
		return "none"
	case errors.As(err, &syntax), errors.As(err, &stdSyntax):
		return "syntax"
	case errors.As(err, &typ), errors.As(err, &stdTyp):
		return "type"
	}
	return "other"
}

// A key of an object is a string: anything else is a syntax error, whatever the type of the keys of the map is, as
// encoding/json reports it.
func TestDecodeMapKeyNotString(t *testing.T) {
	values := []func() any{
		func() any { return new(map[int]int) },
		func() any { return new(map[string]int) },
		func() any { return new(map[any]int) },
		func() any { return new(map[bool]int) },
		func() any { return new(map[float64]int) },
		func() any { return new(map[*int]int) },
		func() any { return new(map[mapKeyKindText]int) },
		func() any { return new(map[fmt.Stringer]int) },
		func() any { return new(map[mapKeyKindUnmarshalJSON]int) },
	}
	inputs := []string{`{null:1}`, `{5:1}`, `{true:1}`, `{[1]:2}`, `{{}:2}`, `{ 5:1}`, `{"5":1,null:2}`, `{"5":1, 6:2}`, `{"5":1,}`}
	for _, value := range values {
		for _, in := range inputs {
			t.Run(fmt.Sprintf("%T%s", value(), in), func(t *testing.T) {
				if got := mapKeyKindError(stdjson.Unmarshal([]byte(in), value())); got != "syntax" {
					t.Fatalf("encoding/json: got %s error", got)
				}
				if got := mapKeyKindError(json.Unmarshal([]byte(in), value())); got != "syntax" {
					t.Errorf("Unmarshal: got %s error, want a syntax error", got)
				}
				if got := mapKeyKindError(json.NewDecoder(bytes.NewReader([]byte(in))).Decode(value())); got != "syntax" {
					t.Errorf("Decoder: got %s error, want a syntax error", got)
				}
			})
		}
	}
}

// A key is decoded by its kind, without a method of unmarshaling JSON, and a map whose keys are interfaces with
// methods has no key decoded into it, as encoding/json of the Go version does.
func TestDecodeMapKeyKind(t *testing.T) {
	values := []func() any{
		func() any { return new(map[mapKeyKindUnmarshalJSON]int) },
		func() any { return new(map[*mapKeyKindUnmarshalJSON]int) },
		func() any { return new(map[fmt.Stringer]int) },
	}
	inputs := []string{`{}`, `{"5":1}`, `{"5":1,"6":2}`, `{"x":1,"6":2}`}
	for _, value := range values {
		for _, in := range inputs {
			t.Run(fmt.Sprintf("%T%s", value(), in), func(t *testing.T) {
				want, got := value(), value()
				wantErr, err := stdjson.Unmarshal([]byte(in), want), json.Unmarshal([]byte(in), got)
				if mapKeyKindError(err) != mapKeyKindError(wantErr) {
					t.Fatalf("got error %v, want %v", err, wantErr)
				}
				m := reflect.ValueOf(got).Elem()
				if _, ok := got.(*map[fmt.Stringer]int); ok && m.Len() != 0 {
					t.Errorf("got %d entries, want none", m.Len())
				}
				if wantErr == nil && mapKeyPointerDumpMap(m) != mapKeyPointerDumpMap(reflect.ValueOf(want).Elem()) {
					t.Errorf("got %s, want %s", mapKeyPointerDumpMap(m), mapKeyPointerDumpMap(reflect.ValueOf(want).Elem()))
				}
			})
		}
	}
}

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"math"
	"net/netip"
	"reflect"
	"testing"
	"time"

	json "github.com/goccy/go-json"
)

type ifaceKeyName string

type ifaceKeyInt int16

type ifaceKeyText struct{ v string }

func (t ifaceKeyText) MarshalText() ([]byte, error) { return []byte("text:" + t.v), nil }

type ifaceKeyPtrText struct{ v string }

func (t *ifaceKeyPtrText) MarshalText() ([]byte, error) {
	if t == nil {
		return []byte("nil"), nil
	}
	return []byte("ptr:" + t.v), nil
}

type ifaceKeyTextError struct{}

func (ifaceKeyTextError) MarshalText() ([]byte, error) { return nil, fmt.Errorf("no text") }

type ifaceKeyJSON struct{}

func (ifaceKeyJSON) MarshalJSON() ([]byte, error) { return []byte(`"json"`), nil }

type ifaceKeyNamedText string

func (ifaceKeyNamedText) MarshalText() ([]byte, error) { return []byte("named-text"), nil }

type ifaceKeyEscapedText string

func (t ifaceKeyEscapedText) MarshalText() ([]byte, error) { return []byte(t), nil }

type ifaceKeyStringer interface{ String() string }

type ifaceKeyStringerValue string

func (s ifaceKeyStringerValue) String() string { return string(s) }

// ifaceKeyErrorKind is the kind of an error, which the errors of go-json and of encoding/json are compared by.
func ifaceKeyErrorKind(err error, std bool) string {
	switch {
	case err == nil:
		return ""
	case std && errors.As(err, new(*stdjson.UnsupportedTypeError)), !std && errors.As(err, new(*json.UnsupportedTypeError)):
		return "unsupported type"
	case std && errors.As(err, new(*stdjson.UnsupportedValueError)), !std && errors.As(err, new(*json.UnsupportedValueError)):
		return "unsupported value"
	case std && errors.As(err, new(*stdjson.MarshalerError)), !std && errors.As(err, new(*json.MarshalerError)):
		return "marshaler"
	}
	return fmt.Sprintf("%T", err)
}

func TestEncodeMapInterfaceKeys(t *testing.T) {
	// A map whose keys are of an interface type is encoded as encoding/json of the Go it is built with encodes
	// it: refused, or, where encoding/json is built on encoding/json/v2, with the names of the dynamic values of
	// its keys, whatever they hold, by Marshal, MarshalIndent and an Encoder.
	var nilKey any
	var nilPtr *string
	var nilText *ifaceKeyPtrText
	s := "pointed"
	ps := &s
	cases := []struct {
		name  string
		value any
	}{
		{"string", map[any]any{"x": 1}},
		{"invalid UTF-8 string", map[any]any{"a\xffb": 1, "\xe3\x81": 2, "<\u2028": 3}},
		{"named string", map[any]any{ifaceKeyName("n"): 1}},
		{"int", map[any]any{1: "a"}},
		{"int kinds", map[any]any{int8(-8): 1, int16(-16): 2, int32(-32): 3, int64(math.MinInt64): 4, ifaceKeyInt(5): 5}},
		{"uint kinds", map[any]any{uint(1): 1, uint8(8): 2, uint16(16): 3, uint32(32): 4, uint64(math.MaxUint64): 5, uintptr(9): 6}},
		{"float32", map[any]any{float32(1.5): 1, float32(2): 2, float32(1e21): 3, float32(1e-7): 4}},
		{"float64", map[any]any{1.5: 1}},
		{"float64 nan", map[any]any{math.NaN(): 1}},
		{"bool", map[any]any{true: 1}},
		{"nil", map[any]any{nilKey: 1}},
		{"struct", map[any]any{struct{}{}: 1}},
		{"array", map[any]any{[2]int{1, 2}: 1}},
		{"complex", map[any]any{complex(1, 2): 1}},
		{"pointer", map[any]any{ps: 1}},
		{"pointer to pointer", map[any]any{&ps: 1}},
		{"nil pointer", map[any]any{nilPtr: 1}},
		{"text marshaler", map[any]any{ifaceKeyText{"a"}: 1}},
		{"pointer text marshaler", map[any]any{&ifaceKeyPtrText{"b"}: 1}},
		{"value of pointer text marshaler", map[any]any{ifaceKeyPtrText{"c"}: 1}},
		{"nil pointer text marshaler", map[any]any{nilText: 1}},
		{"text marshaler error", map[any]any{ifaceKeyTextError{}: 1}},
		{"json marshaler", map[any]any{ifaceKeyJSON{}: 1}},
		{"named text marshaler", map[any]any{ifaceKeyNamedText("x"): 1}},
		{"time", map[any]any{time.Date(2026, 9, 25, 12, 0, 0, 5, time.UTC): 1}},
		{"netip", map[any]any{netip.MustParseAddr("192.0.2.1"): 1}},
		{"sorted", map[any]any{"b": 1, 10: 2, "a": 3, 2: 4, ifaceKeyName("c"): 5, -1: 6, float32(0.5): 7}},
		{"escaped", map[any]any{"<&>": 1, "\u2028": 2, "a\"b": 3, "\n": 4, "\u00e9": 5, "\\": 6}},
		{"values", map[any]string{"a": "x", 2: "y"}},
		{"struct values", map[any]struct{ A int }{"a": {1}, 2: {2}}},
		{"nested", map[string]any{"m": map[any]any{"x": map[any]any{3: []any{true, nil}}}}},
		{"in a struct", struct {
			M map[any]int `json:"m"`
			N int
		}{M: map[any]int{"k": 1}, N: 2}},
		{"empty", map[any]any{}},
		{"nil map", map[any]any(nil)},
		{"text marshaler keys with escapes", map[ifaceKeyEscapedText]int{"<&>": 1, "\u2028": 2, "a\"b": 3, "\t": 4, "b": 5}},
		{"stringer keys", map[ifaceKeyStringer]int{ifaceKeyStringerValue("s"): 1}},
		{"stringer float64 keys", map[fmt.Stringer]int{time.Duration(3): 1}},
		{"text marshaler keys", map[interface{ MarshalText() ([]byte, error) }]int{ifaceKeyText{"t"}: 1, &ifaceKeyPtrText{"p"}: 2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want, wantErr := stdjson.Marshal(tc.value)
			got, gotErr := json.Marshal(tc.value)
			if ifaceKeyErrorKind(gotErr, false) != ifaceKeyErrorKind(wantErr, true) {
				t.Fatalf("Marshal: got the error %v, want %v", gotErr, wantErr)
			}
			if wantErr == nil && !bytes.Equal(got, want) {
				t.Fatalf("Marshal: got %s, want %s", got, want)
			}

			want, wantErr = stdjson.MarshalIndent(tc.value, ">", "  ")
			got, gotErr = json.MarshalIndent(tc.value, ">", "  ")
			if ifaceKeyErrorKind(gotErr, false) != ifaceKeyErrorKind(wantErr, true) {
				t.Fatalf("MarshalIndent: got the error %v, want %v", gotErr, wantErr)
			}
			if wantErr == nil && !bytes.Equal(got, want) {
				t.Fatalf("MarshalIndent: got %s, want %s", got, want)
			}

			// the VMs of the colors and of the options: the same entries, in no order when they are not sorted.
			for _, opts := range [][]json.EncodeOptionFunc{
				{json.Colorize(json.DefaultColorScheme)},
				{json.UnorderedMap()},
			} {
				_, gotErr := json.MarshalWithOption(tc.value, opts...)
				if ifaceKeyErrorKind(gotErr, false) != ifaceKeyErrorKind(wantErr, true) {
					t.Fatalf("MarshalWithOption: got the error %v, want %v", gotErr, wantErr)
				}
				if wantErr != nil || len(opts) == 0 {
					continue
				}
				if _, err := json.MarshalIndentWithOption(tc.value, "", " ", opts...); err != nil {
					t.Fatalf("MarshalIndentWithOption: %v", err)
				}
			}
			if wantErr == nil {
				want, _ := stdjson.Marshal(tc.value)
				got, err := json.MarshalWithOption(tc.value, json.UnorderedMap())
				if err != nil {
					t.Fatal(err)
				}
				var wantValue, gotValue any
				if err := stdjson.Unmarshal(want, &wantValue); err != nil {
					t.Fatal(err)
				}
				if err := stdjson.Unmarshal(got, &gotValue); err != nil {
					t.Fatalf("UnorderedMap: %s: %v", got, err)
				}
				if !reflect.DeepEqual(gotValue, wantValue) {
					t.Fatalf("UnorderedMap: got %s, want %s", got, want)
				}
			}

			var wantBuf, gotBuf bytes.Buffer
			wantEnc, gotEnc := stdjson.NewEncoder(&wantBuf), json.NewEncoder(&gotBuf)
			wantEnc.SetEscapeHTML(false)
			gotEnc.SetEscapeHTML(false)
			wantErr, gotErr = wantEnc.Encode(tc.value), gotEnc.Encode(tc.value)
			if ifaceKeyErrorKind(gotErr, false) != ifaceKeyErrorKind(wantErr, true) {
				t.Fatalf("Encoder: got the error %v, want %v", gotErr, wantErr)
			}
			if wantErr == nil && !bytes.Equal(gotBuf.Bytes(), wantBuf.Bytes()) {
				t.Fatalf("Encoder: got %s, want %s", gotBuf.Bytes(), wantBuf.Bytes())
			}
		})
	}
}

package json_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
)

// The MarshalJSONTo and UnmarshalJSONFrom methods of the Encoder and the Decoder of github.com/goccy/go-json/jsontext,
// which the root package calls before MarshalJSON and UnmarshalJSON, on every version of Go.

type toValue struct{ N int }

func (v toValue) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(jsontext.Value(`{"to" : ` + strings.Repeat("1", v.N) + `}`))
}

type toPointer struct{ N int }

func (v *toPointer) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.Int(int64(v.N)))
}

// toAndJSON has both methods: MarshalJSONTo is called.
type toAndJSON struct{}

func (toAndJSON) MarshalJSON() ([]byte, error)              { return []byte(`"json"`), nil }
func (toAndJSON) MarshalJSONTo(enc *jsontext.Encoder) error { return enc.WriteToken(jsontext.String("to")) }

// toDeclinedJSON declines for MarshalJSON, toDeclined for its fields.
type toDeclinedJSON struct{}

func (toDeclinedJSON) MarshalJSON() ([]byte, error)          { return []byte(`"json"`), nil }
func (toDeclinedJSON) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

type toDeclined struct {
	A int    `json:"a"`
	B string `json:"b,omitempty"`
}

func (toDeclined) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

// toFails fails, toTwice writes two values, toNothing none, toDeclinedLate declines after it wrote.
type (
	toFails        struct{}
	toTwice        struct{}
	toNothing      struct{}
	toDeclinedLate struct{}
)

func (toFails) MarshalJSONTo(*jsontext.Encoder) error { return errors.New("failed") }
func (toTwice) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.Int(1)); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.Int(2))
}
func (toNothing) MarshalJSONTo(*jsontext.Encoder) error { return nil }
func (toDeclinedLate) MarshalJSONTo(enc *jsontext.Encoder) error {
	_ = enc.WriteToken(jsontext.Int(1))
	return errors.ErrUnsupported
}

type toHTML struct{}

func (toHTML) MarshalJSONTo(enc *jsontext.Encoder) error { return enc.WriteToken(jsontext.String("<&>")) }

func TestEncodeMarshalJSONTo(t *testing.T) {
	p := &toPointer{7}
	tests := []struct {
		name     string
		in       any
		want     string
		indented string
	}{
		{"value", toValue{2}, `{"to":11}`, "{\n \"to\": 11\n}"},
		{"pointer", p, `7`, `7`},
		{"pointer receiver of an addressable value", &struct{ P toPointer }{toPointer{3}}, `{"P":3}`, "{\n \"P\": 3\n}"},
		{"pointer receiver of a value which is not addressable", toPointer{4}, `{"N":4}`, "{\n \"N\": 4\n}"},
		{"nil pointer", (*toPointer)(nil), `null`, `null`},
		{"before MarshalJSON", toAndJSON{}, `"to"`, `"to"`},
		{"declined to MarshalJSON", toDeclinedJSON{}, `"json"`, `"json"`},
		{"declined to the fields", toDeclined{A: 1}, `{"a":1}`, "{\n \"a\": 1\n}"},
		{"declined behind a pointer", &toDeclined{A: 2, B: "b"}, `{"a":2,"b":"b"}`, "{\n \"a\": 2,\n \"b\": \"b\"\n}"},
		{"in a slice", []toValue{{1}, {3}}, `[{"to":1},{"to":111}]`, "[\n {\n  \"to\": 1\n },\n {\n  \"to\": 111\n }\n]"},
		{"in a slice of pointers", []*toPointer{p, nil}, `[7,null]`, "[\n 7,\n null\n]"},
		{"in a map", map[string]toValue{"k": {1}}, `{"k":{"to":1}}`, "{\n \"k\": {\n  \"to\": 1\n }\n}"},
		{"in an interface", map[string]any{"k": toDeclined{A: 3}}, `{"k":{"a":3}}`, "{\n \"k\": {\n  \"a\": 3\n }\n}"},
		{"escaped for HTML", toHTML{}, `"\u003c\u0026\u003e"`, `"\u003c\u0026\u003e"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.in)
			if err != nil || string(got) != tt.want {
				t.Errorf("got %s, %v, want %s", got, err, tt.want)
			}
			got, err = json.MarshalIndent(tt.in, "", " ")
			if err != nil || string(got) != tt.indented {
				t.Errorf("indented: got %q, %v, want %q", got, err, tt.indented)
			}
		})
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(toHTML{}); err != nil || b.String() != "\"<&>\"\n" {
		t.Errorf("without HTML escaping: got %q, %v", b.String(), err)
	}
}

// The errors of MarshalJSONTo are MarshalerErrors of the pointer to the value, as encoding/json reports them.
func TestEncodeMarshalJSONToErrors(t *testing.T) {
	for _, tt := range []struct {
		in   any
		want string
	}{
		{toFails{}, "json: error calling MarshalJSONTo for type *json_test.toFails: failed"},
		{toTwice{}, "json: error calling MarshalJSONTo for type *json_test.toTwice: must read or write exactly one value"},
		{toNothing{}, "json: error calling MarshalJSONTo for type *json_test.toNothing: must read or write exactly one value"},
		{toDeclinedLate{}, "json: error calling MarshalJSONTo for type *json_test.toDeclinedLate: unsupported calls must not read or write any tokens"},
	} {
		_, err := json.Marshal(tt.in)
		var marshalerErr *json.MarshalerError
		if !errors.As(err, &marshalerErr) || err.Error() != tt.want {
			t.Errorf("%T: got the error %v, want %s", tt.in, err, tt.want)
		}
	}
}

type fromValue struct{ Got string }

func (v *fromValue) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	raw, err := dec.ReadValue()
	v.Got = "from " + string(raw)
	return err
}

// fromAndJSON has both methods: UnmarshalJSONFrom is called.
type fromAndJSON struct{ Got string }

func (v *fromAndJSON) UnmarshalJSON(b []byte) error { v.Got = "json " + string(b); return nil }
func (v *fromAndJSON) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	raw, err := dec.ReadValue()
	v.Got = "from " + string(raw)
	return err
}

// fromDeclinedJSON declines for UnmarshalJSON, fromDeclined for its fields.
type fromDeclinedJSON struct{ Got string }

func (v *fromDeclinedJSON) UnmarshalJSON(b []byte) error        { v.Got = "json " + string(b); return nil }
func (*fromDeclinedJSON) UnmarshalJSONFrom(*jsontext.Decoder) error { return errors.ErrUnsupported }

type fromDeclined struct {
	A int `json:"a"`
}

func (*fromDeclined) UnmarshalJSONFrom(*jsontext.Decoder) error { return errors.ErrUnsupported }

// fromFails fails, fromNothing reads nothing, fromDeclinedLate declines after it read.
type (
	fromFails        struct{}
	fromNothing      struct{}
	fromDeclinedLate struct{}
)

func (*fromFails) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	_, _ = dec.ReadValue()
	return errors.New("failed")
}
func (*fromNothing) UnmarshalJSONFrom(*jsontext.Decoder) error { return nil }
func (*fromDeclinedLate) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	_, _ = dec.ReadToken()
	return errors.ErrUnsupported
}

// fromKey is a key of a map, which is decoded by UnmarshalText, not by UnmarshalJSONFrom.
type fromKey string

func (k *fromKey) UnmarshalJSONFrom(*jsontext.Decoder) error { *k = "from"; return nil }
func (k *fromKey) UnmarshalText(b []byte) error            { *k = fromKey("text " + string(b)); return nil }

type fromHolder struct {
	V  fromValue
	P  *fromValue
	VJ fromAndJSON
	DJ fromDeclinedJSON
	D  fromDeclined
	L  []fromValue
	M  map[string]fromValue
	K  map[fromKey]int
	N  int
}

func TestDecodeUnmarshalJSONFrom(t *testing.T) {
	in := `{"V":"v","P":[1, 2],"VJ":{"x":1},"DJ":"d","D":{"a":3},"L":["l",{}],"M":{"m":true},"K":{"k":1},"N":4}`
	want := fromHolder{
		V: fromValue{`from "v"`}, P: &fromValue{"from [1, 2]"}, VJ: fromAndJSON{`from {"x":1}`},
		DJ: fromDeclinedJSON{`json "d"`}, D: fromDeclined{A: 3}, L: []fromValue{{`from "l"`}, {"from {}"}},
		M: map[string]fromValue{"m": {"from true"}}, K: map[fromKey]int{"text k": 1}, N: 4,
	}
	for name, decode := range map[string]func(v any) error{
		"Unmarshal": func(v any) error { return json.Unmarshal([]byte(in), v) },
		"Decoder":   func(v any) error { return json.NewDecoder(strings.NewReader(in)).Decode(v) },
	} {
		var got fromHolder
		if err := decode(&got); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got.V != want.V || *got.P != *want.P || got.VJ != want.VJ || got.DJ != want.DJ || got.D != want.D ||
			len(got.L) != 2 || got.L[0] != want.L[0] || got.L[1] != want.L[1] || got.M["m"] != want.M["m"] ||
			got.K["text k"] != 1 || len(got.K) != 1 || got.N != want.N {
			t.Errorf("%s: got %+v, want %+v", name, got, want)
		}
	}
}

// The errors of UnmarshalJSONFrom are its own, as encoding/json reports them, after which the decoding goes on.
func TestDecodeUnmarshalJSONFromErrors(t *testing.T) {
	for _, tt := range []struct {
		in   string
		v    any
		want string
	}{
		{`{"a":1}`, new(fromFails), "failed"},
		{`1`, new(fromNothing), "must read or write exactly one value"},
		{`[1]`, new(fromDeclinedLate), "unsupported calls must not read or write any tokens"},
	} {
		if err := json.Unmarshal([]byte(tt.in), tt.v); err == nil || err.Error() != tt.want {
			t.Errorf("%T: got the error %v, want %s", tt.v, err, tt.want)
		}
	}
	var got struct {
		F fromFails
		N int
	}
	if err := json.Unmarshal([]byte(`{"F":{},"N":1}`), &got); err == nil || got.N != 1 {
		t.Errorf("after the error: got %+v, %v", got, err)
	}
}

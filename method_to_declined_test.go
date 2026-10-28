package json_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
)

// The values whose MarshalJSONTo declines, which are written by the representation after the method, and the
// UnmarshalJSONFrom of a value of the string option.

// toList declines when N is 0: the values of its type in it call the method again.
type toList struct {
	N    int
	Next *toList `json:",omitempty"`
}

func (v toList) MarshalJSONTo(enc *jsontext.Encoder) error {
	if v.N == 0 {
		return errors.ErrUnsupported
	}
	return enc.WriteToken(jsontext.Int(int64(v.N * 100)))
}

// toPtrJSON declines for the MarshalJSON of its pointer, which an addressable value has.
type toPtrJSON struct{ X int }

func (toPtrJSON) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }
func (*toPtrJSON) MarshalJSON() ([]byte, error)          { return []byte(`"pointer json"`), nil }

// toMixed has MarshalJSON and its pointer MarshalJSONTo, which comes first for an addressable value.
type toMixed struct{}

func (toMixed) MarshalJSON() ([]byte, error)                { return []byte(`"json"`), nil }
func (*toMixed) MarshalJSONTo(enc *jsontext.Encoder) error { return enc.WriteToken(jsontext.String("to")) }

// toChan and toComplex decline, and have no representation after the method.
type (
	toChan    chan int
	toComplex complex128
)

func (toChan) MarshalJSONTo(*jsontext.Encoder) error    { return errors.ErrUnsupported }
func (toComplex) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

// plainDeclined is toDeclined without the method, which a declined value is written as.
type plainDeclined struct {
	A int    `json:"a"`
	B string `json:"b,omitempty"`
}

func TestEncodeMarshalJSONToDeclined(t *testing.T) {
	for _, tt := range []struct {
		in   any
		want string
	}{
		{toList{0, &toList{N: 1}}, `{"N":0,"Next":100}`},
		{&toList{0, &toList{0, &toList{N: 2}}}, `{"N":0,"Next":{"N":0,"Next":200}}`},
		{[]toList{{0, &toList{N: 3}}}, `[{"N":0,"Next":300}]`},
		{toPtrJSON{}, `{"X":0}`},
		{&toPtrJSON{}, `"pointer json"`},
		{[]toPtrJSON{{}}, `["pointer json"]`},
		{toMixed{}, `"json"`},
		{&toMixed{}, `"to"`},
		{[]toMixed{{}}, `["to"]`},
		{&struct{ M toMixed }{}, `{"M":"to"}`},
	} {
		if got, err := json.Marshal(tt.in); err != nil || string(got) != tt.want {
			t.Errorf("%#v: got %s, %v, want %s", tt.in, got, err, tt.want)
		}
	}
	for _, v := range []any{toChan(nil), toComplex(1)} {
		var unsupported *json.UnsupportedTypeError
		if _, err := json.Marshal(v); !errors.As(err, &unsupported) {
			t.Errorf("%T: got the error %v, want an UnsupportedTypeError", v, err)
		}
	}
	// a cycle through declined values is found, as any other
	cycle := &toList{}
	cycle.Next = cycle
	for _, v := range []any{cycle, []*toList{cycle}, *cycle} {
		var unsupported *json.UnsupportedValueError
		if _, err := json.Marshal(v); !errors.As(err, &unsupported) {
			t.Errorf("%T: got the error %v, want an UnsupportedValueError", v, err)
		}
	}
	// a field of an interface type with MarshalJSONTo is written by its dynamic value
	if got, err := json.Marshal(struct{ F interface{ MarshalJSONTo(*jsontext.Encoder) error } }{toValue{1}}); err != nil || string(got) != `{"F":{"to":1}}` {
		t.Errorf("a field of an interface type: got %s, %v", got, err)
	}
}

// A declined value is written as the value without the method, indented and colored as it.
func TestEncodeMarshalJSONToDeclinedFormat(t *testing.T) {
	declined := struct {
		X int
		D toDeclined
		L []toDeclined
	}{1, toDeclined{A: 2, B: "<"}, []toDeclined{{A: 3}}}
	plain := struct {
		X int
		D plainDeclined
		L []plainDeclined
	}{1, plainDeclined{A: 2, B: "<"}, []plainDeclined{{A: 3}}}
	for _, opts := range [][]json.EncodeOptionFunc{nil, {json.Colorize(json.DefaultColorScheme)}, {json.DisableHTMLEscape()}} {
		got, err := json.MarshalWithOption(declined, opts...)
		want, _ := json.MarshalWithOption(plain, opts...)
		if err != nil || string(got) != string(want) {
			t.Errorf("got %q, %v, want %q", got, err, want)
		}
		got, err = json.MarshalIndentWithOption(declined, ">", "  ", opts...)
		want, _ = json.MarshalIndentWithOption(plain, ">", "  ", opts...)
		if err != nil || string(got) != string(want) {
			t.Errorf("indented: got %q, %v, want %q", got, err, want)
		}
	}
	encode := func(v any) string {
		var b bytes.Buffer
		enc := json.NewEncoder(&b)
		enc.SetIndent("", " ")
		if err := enc.Encode(v); err != nil {
			t.Fatal(err)
		}
		return b.String()
	}
	if got, want := encode(declined), encode(plain); got != want {
		t.Errorf("Encoder with an indentation: got %q, want %q", got, want)
	}
}

// fromLength is a number which UnmarshalJSONFrom sets to the length of the value, also with the string option.
type fromLength int

func (v *fromLength) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	raw, err := dec.ReadValue()
	*v = fromLength(len(raw))
	return err
}

func TestDecodeUnmarshalJSONFromStringOption(t *testing.T) {
	var got struct {
		A fromLength  `json:",string"`
		P *fromLength `json:",string"`
	}
	if err := json.Unmarshal([]byte(`{"A":"12","P":"123"}`), &got); err != nil || got.A != 4 || got.P == nil || *got.P != 5 {
		t.Errorf("got %v, %v, %v", got.A, got.P, err)
	}
}

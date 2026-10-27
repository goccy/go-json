//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjson "encoding/json"
	"encoding/json/jsontext"
	"errors"
	"strconv"
	"testing"

	"github.com/goccy/go-json"
)

// The MarshalJSONTo and UnmarshalJSONFrom methods of the Encoder and the Decoder of encoding/json/jsontext, which
// encoding/json calls where it is built on encoding/json/v2, and so does the root package.

type stdToValue struct{ N int }

func (v stdToValue) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(jsontext.Value(`{"to" : ` + strconv.Itoa(v.N) + `}`))
}

type stdToPointer struct{ N int }

func (v *stdToPointer) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.Int(int64(v.N)))
}

type stdToAndJSON struct{}

func (stdToAndJSON) MarshalJSON() ([]byte, error)              { return []byte(`"json"`), nil }
func (stdToAndJSON) MarshalJSONTo(enc *jsontext.Encoder) error { return enc.WriteToken(jsontext.String("to")) }

type stdToDeclined struct {
	A int    `json:"a"`
	B string `json:"b,omitempty"`
}

func (stdToDeclined) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

type stdToHTML struct{}

func (stdToHTML) MarshalJSONTo(enc *jsontext.Encoder) error { return enc.WriteToken(jsontext.String("<&>")) }

type (
	stdToFails struct{}
	stdToTwice struct{}
)

func (stdToFails) MarshalJSONTo(*jsontext.Encoder) error { return errors.New("failed") }
func (stdToTwice) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.Int(1)); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.Int(2))
}

func TestEncodeStdMarshalJSONTo(t *testing.T) {
	p := &stdToPointer{7}
	compareStdEncoding(t,
		stdToValue{2}, p, toPointer{4}, &struct{ P stdToPointer }{stdToPointer{3}}, (*stdToPointer)(nil),
		stdToAndJSON{}, stdToDeclined{A: 1}, &stdToDeclined{A: 2, B: "b"}, stdToHTML{},
		[]stdToValue{{1}, {3}}, []*stdToPointer{p, nil}, map[string]stdToValue{"k": {1}},
		map[string]any{"k": stdToDeclined{A: 3}}, [2]stdToValue{{1}, {2}},
	)
	for _, v := range []any{stdToFails{}, stdToTwice{}} {
		_, stdErr := stdjson.Marshal(v)
		_, err := json.Marshal(v)
		if err == nil || stdErr == nil || err.Error() != stdErr.Error() {
			t.Errorf("%T: got the error %v, want %v", v, err, stdErr)
		}
	}
}

type stdFromValue struct{ Got string }

func (v *stdFromValue) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	raw, err := dec.ReadValue()
	v.Got = "from " + string(raw)
	return err
}

type stdFromAndJSON struct{ Got string }

func (v *stdFromAndJSON) UnmarshalJSON(b []byte) error { v.Got = "json " + string(b); return nil }
func (v *stdFromAndJSON) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	raw, err := dec.ReadValue()
	v.Got = "from " + string(raw)
	return err
}

type stdFromDeclined struct {
	A int `json:"a"`
}

func (*stdFromDeclined) UnmarshalJSONFrom(*jsontext.Decoder) error { return errors.ErrUnsupported }

type (
	stdFromFails   struct{}
	stdFromNothing struct{}
)

func (*stdFromFails) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	_, _ = dec.ReadValue()
	return errors.New("failed")
}
func (*stdFromNothing) UnmarshalJSONFrom(*jsontext.Decoder) error { return nil }

type stdFromKey string

func (k *stdFromKey) UnmarshalJSONFrom(*jsontext.Decoder) error { *k = "from"; return nil }
func (k *stdFromKey) UnmarshalText(b []byte) error            { *k = stdFromKey("text " + string(b)); return nil }

type stdFromHolder struct {
	V  stdFromValue
	P  *stdFromValue
	VJ stdFromAndJSON
	D  stdFromDeclined
	L  []stdFromValue
	M  map[string]stdFromValue
	K  map[stdFromKey]int
	N  int
}

func TestDecodeStdUnmarshalJSONFrom(t *testing.T) {
	compareUnmarshal(t, `{"V":"v","P":[1, 2],"VJ":{"x":1},"D":{"a":3},"L":["l",{}],"M":{"m":true},"K":{"k":1},"N":4}`,
		func() any { return new(stdFromHolder) })
	compareUnmarshal(t, `null`, func() any { return new(stdFromValue) })
	compareUnmarshal(t, `{"P":null}`, func() any { return new(stdFromHolder) })
	for _, tt := range []struct {
		in string
		v  func() any
	}{
		{`{"a":1}`, func() any { return new(stdFromFails) }},
		{`1`, func() any { return new(stdFromNothing) }},
		{`[{},{}]`, func() any { return new([]stdFromFails) }},
	} {
		stdErr := stdjson.Unmarshal([]byte(tt.in), tt.v())
		err := json.Unmarshal([]byte(tt.in), tt.v())
		if err == nil || stdErr == nil || err.Error() != stdErr.Error() {
			t.Errorf("%s: got the error %v, want %v", tt.in, err, stdErr)
		}
	}
}

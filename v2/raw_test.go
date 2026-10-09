package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The raw value which a method or a function returns is formatted into the output, not where it is: it may be the
// memory of the caller, as the value of a json.RawMessage is.
func TestMarshalKeepsRawValues(t *testing.T) {
	shared := []byte(`[ "<a>" , { "b" : null } ]`)
	tests := []struct {
		name string
		in   any
		opts []json.Options
		want string
	}{
		{name: "RawMessage", in: struct{ R stdjson.RawMessage }{stdjson.RawMessage(shared)}, want: `{"R":["<a>",{"b":null}]}`},
		{name: "NilRawMessage", in: struct{ R stdjson.RawMessage }{}, want: `{"R":null}`},
		{
			name: "EscapeForHTML", in: struct{ R stdjson.RawMessage }{stdjson.RawMessage(shared)},
			opts: []json.Options{jsontext.EscapeForHTML(true)}, want: "{\"R\":[\"\\u003ca\\u003e\",{\"b\":null}]}",
		},
		{
			name: "MarshalFunc", in: []int{1, 2},
			opts: []json.Options{json.WithMarshalers(json.MarshalFunc(func(int) ([]byte, error) { return shared, nil }))},
			want: `[["<a>",{"b":null}],["<a>",{"b":null}]]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := string(shared)
			got, err := json.Marshal(tt.in, tt.opts...)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal:\ngot  %s\nwant %s", got, tt.want)
			}
			if string(shared) != orig {
				t.Errorf("the raw value was changed to %s", shared)
			}
		})
	}
}

type rawOptionFields struct {
	Z int
	R stdjson.RawMessage
	N stdjson.Number
	P *stdjson.Number `json:",string"`
	A int
}

// The options which rewrite raw values rewrite the raw values only, and a json.Number, which is one: the members
// of the structs and the maps are written in their order, through any of the calls.
func TestMarshalRawValueOptions(t *testing.T) {
	big := stdjson.Number("1e20")
	in := rawOptionFields{R: stdjson.RawMessage(`{"b":1.0,"a":1E2}`), N: "1.50", P: &big}
	tests := []struct {
		name string
		opts []json.Options
		want string
	}{
		{"ReorderRawObjects", []json.Options{jsontext.ReorderRawObjects(true)},
			`{"Z":0,"R":{"a":1E2,"b":1.0},"N":1.50,"P":"1e20","A":0}`},
		{"CanonicalizeRawFloats", []json.Options{jsontext.CanonicalizeRawFloats(true)},
			`{"Z":0,"R":{"b":1,"a":100},"N":1.5,"P":"1e20","A":0}`},
		{"all with Multiline", []json.Options{jsontext.ReorderRawObjects(true), jsontext.CanonicalizeRawFloats(true),
			jsontext.Multiline(true), jsontext.WithIndent(" ")},
			"{\n \"Z\": 0,\n \"R\": {\n  \"a\": 100,\n  \"b\": 1\n },\n \"N\": 1.5,\n \"P\": \"1e20\",\n \"A\": 0\n}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := json.Marshal(in, tt.opts...); err != nil || string(got) != tt.want {
				t.Errorf("Marshal: got %s %v, want %s", got, err, tt.want)
			}
			var b bytes.Buffer
			if err := json.MarshalWrite(&b, in, tt.opts...); err != nil || b.String() != tt.want {
				t.Errorf("MarshalWrite: got %q %v, want %q", b.String(), err, tt.want)
			}
			b.Reset()
			if err := json.MarshalEncode(jsontext.NewEncoder(&b, tt.opts...), in); err != nil || b.String() != tt.want+"\n" {
				t.Errorf("MarshalEncode: got %q %v, want %q", b.String(), err, tt.want+"\n")
			}
		})
	}
}

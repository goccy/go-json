package json_test

import (
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

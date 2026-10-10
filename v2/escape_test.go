package json_test

import (
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

type escapedNames struct {
	A string `json:"a<\u2028>"`
	B int
}

// escapingCall writes its value by MarshalEncode with options of its own.
type escapingCall struct {
	V    any
	opts []json.Options
}

func (v escapingCall) MarshalJSONTo(e *jsontext.Encoder) error {
	return json.MarshalEncode(e, v.V, v.opts...)
}

// The names of the fields are escaped as the strings are, and a string is written as the call which wrote it escaped
// it, also in a call of MarshalEncode with other escapes and in the output formatted after.
func TestMarshalEscapes(t *testing.T) {
	names := escapedNames{A: "<\u2028>"}
	js := jsontext.EscapeForJS(true)
	html := jsontext.EscapeForHTML(true)
	noEscapes := []json.Options{jsontext.EscapeForHTML(false), jsontext.EscapeForJS(false)}
	tests := []struct {
		name string
		v    any
		opts []json.Options
		want string
	}{
		{"names", names, nil, "{\"a<\u2028>\":\"<\u2028>\",\"B\":0}"},
		{"names for JavaScript", names, []json.Options{js}, `{"a<\u2028>":"<\u2028>","B":0}`},
		{"names for HTML", names, []json.Options{html}, "{\"a\\u003c\u2028\\u003e\":\"\\u003c\u2028\\u003e\",\"B\":0}"},
		{"names for both", names, []json.Options{html, js}, `{"a\u003c\u2028\u003e":"\u003c\u2028\u003e","B":0}`},
		{"a nested call escapes", escapingCall{"<\u2028>", []json.Options{js}}, nil, `"<\u2028>"`},
		{"a nested call doesn't escape", escapingCall{"<\u2028>", noEscapes}, []json.Options{js, html}, "\"<\u2028>\""},
		{"a nested call in one which doesn't escape",
			escapingCall{escapingCall{"<\u2028>", []json.Options{js}}, noEscapes}, nil, `"<\u2028>"`},
		{"a nested call formatted", []any{escapingCall{"<", noEscapes}},
			[]json.Options{html, jsontext.Multiline(true), jsontext.WithIndent(" ")}, "[\n \"<\"\n]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := json.Marshal(tt.v, tt.opts...); err != nil || string(got) != tt.want {
				t.Errorf("got %s %v, want %s", got, err, tt.want)
			}
		})
	}
}

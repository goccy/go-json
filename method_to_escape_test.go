package json_test

import (
	"bytes"
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
)

// toString writes its string by MarshalJSONTo, and jsonString the same by MarshalJSON: their outputs are escaped
// the same.
type (
	toString   string
	jsonString string
)

func (v toString) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.String(string(v)))
}

func (v jsonString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(v) + `"`), nil
}

func TestEncodeMarshalJSONToEscapes(t *testing.T) {
	for _, s := range []string{"plain", "<a&b>", "line" + string(rune(0x2028)) + "end" + string(rune(0x2029)), "é" + string(rune(0x20ac))} {
		for _, escape := range []bool{true, false} {
			encode := func(v any) string {
				var b bytes.Buffer
				enc := json.NewEncoder(&b)
				enc.SetEscapeHTML(escape)
				if err := enc.Encode(v); err != nil {
					t.Fatal(err)
				}
				return b.String()
			}
			got, want := encode([]any{toString(s), map[string]toString{"k": toString(s)}}), encode([]any{jsonString(s), map[string]jsonString{"k": jsonString(s)}})
			if got != want {
				t.Errorf("%q, escaped for HTML %v: got %q, want %q", s, escape, got, want)
			}
		}
	}
}

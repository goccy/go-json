package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

// TestDecodeNumberSyntaxErrors decodes numbers which are not numbers by the grammar, or which another value
// follows without a separator, into values of every kind of the decoders of numbers, by Unmarshal and by a
// Decoder, and compares the errors with the ones of encoding/json.
func TestDecodeNumberSyntaxErrors(t *testing.T) {
	numbers := []string{"1.", "1.x", "-", "-x", "1e", "1e+", "1ex", "01", "1x", "-01", "1.5x", "2]", "0x1", "1.e5", "1e5.", "--1", "1-2"}
	wrappers := []string{"%s", "[%s]", `{"a":%s}`, "[%s,1]", `{"a":%s,"b":1}`}
	targets := []func() any{
		func() any { return new(any) },
		func() any { return new(float64) },
		func() any { return new(float32) },
		func() any { return new(json.Number) },
		func() any { return new(int) },
		func() any { return new(uint) },
		func() any { return new([]any) },
		func() any { return new(map[string]any) },
		func() any { return new(struct{ A int }) },
		func() any { return new(struct{ A json.Number }) },
		func() any { return new(struct{ A float64 }) },
		func() any { return new([]float64) },
		func() any { return new(map[string]json.Number) },
	}
	for _, n := range numbers {
		for _, w := range wrappers {
			in := fmt.Sprintf(w, n)
			for _, mk := range targets {
				v := mk()
				got, want := json.Unmarshal([]byte(in), v), stdjson.Unmarshal([]byte(in), mk())
				if fmt.Sprint(got) != fmt.Sprint(want) {
					t.Errorf("Unmarshal of %s into %T:\n got: %v\nwant: %v", in, v, got, want)
				}
				got = json.NewDecoder(strings.NewReader(in)).Decode(mk())
				want = stdjson.NewDecoder(strings.NewReader(in)).Decode(mk())
				if fmt.Sprint(got) != fmt.Sprint(want) {
					t.Errorf("Decode of %s into %T:\n got: %v\nwant: %v", in, v, got, want)
				}
			}
		}
	}
}

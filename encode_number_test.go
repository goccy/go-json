package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

// TestEncodeNumberGrammar marshals json.Number values, valid and not, as a value, in fields with and without the
// options omitempty and string, in a slice, in a map and through an interface, and compares the output and the
// error with encoding/json: a json.Number which is not a JSON number by its grammar is an error.
func TestEncodeNumberGrammar(t *testing.T) {
	type field struct {
		N stdjson.Number `json:"n"`
	}
	type omitEmpty struct {
		N stdjson.Number `json:"n,omitempty"`
	}
	type quoted struct {
		N stdjson.Number `json:"n,string"`
	}
	type quotedOmitEmpty struct {
		N stdjson.Number `json:"n,string,omitempty"`
	}
	type pointer struct {
		N *stdjson.Number `json:"n"`
	}
	for _, s := range []string{
		"", "0", "-0", "12", "-12.5e+3", "1E5", "0.000001", "123456789012345678901234567890",
		"1-2", "--", "-", "e", "1e", "1e+", "01", "1.", ".5", "+1", " 1", "1 ", "1.5.5", "NaN", "0x10", "1_000",
	} {
		n := stdjson.Number(s)
		for _, v := range []any{
			n, &n, field{n}, omitEmpty{n}, quoted{n}, quotedOmitEmpty{n}, pointer{&n}, []stdjson.Number{n, n},
			map[string]stdjson.Number{"k": n}, []any{n}, map[string]any{"k": n},
		} {
			got, err := json.Marshal(v)
			want, stdErr := stdjson.Marshal(v)
			if fmt.Sprint(err) != fmt.Sprint(stdErr) || string(got) != string(want) {
				t.Errorf("Marshal(%T of %q) = %s, %v; want %s, %v", v, s, got, err, want, stdErr)
			}
		}
	}
}

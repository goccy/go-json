//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

// The syntax errors of Compact and Indent, and of the outputs of MarshalJSON, are the ones of encoding/json of the
// latest Go, its messages and offsets, which go-json reports on every version of Go.
func TestFormatErrorsAsLatestGo(t *testing.T) {
	offsetOf := func(err error) int64 {
		var g *json.SyntaxError
		if errors.As(err, &g) {
			return g.Offset
		}
		var s *stdjson.SyntaxError
		if errors.As(err, &s) {
			return s.Offset
		}
		return -1
	}
	for _, in := range formatInputs {
		var got, want bytes.Buffer
		gerr := json.Compact(&got, []byte(in))
		werr := stdjson.Compact(&want, []byte(in))
		if fmt.Sprint(gerr) != fmt.Sprint(werr) || offsetOf(gerr) != offsetOf(werr) {
			t.Errorf("Compact(%q): got %v (%d); want %v (%d)", in, gerr, offsetOf(gerr), werr, offsetOf(werr))
		}
		gerr = json.Indent(&got, []byte(in), ">", "\t")
		werr = stdjson.Indent(&want, []byte(in), ">", "\t")
		if fmt.Sprint(gerr) != fmt.Sprint(werr) || offsetOf(gerr) != offsetOf(werr) {
			t.Errorf("Indent(%q): got %v (%d); want %v (%d)", in, gerr, offsetOf(gerr), werr, offsetOf(werr))
		}
		// encoding/json names the pointer type, which go-json doesn't: the error of the output is compared.
		_, gerr = json.Marshal(rawOutput(in))
		_, werr = stdjson.Marshal(rawOutput(in))
		var gm *json.MarshalerError
		var wm *stdjson.MarshalerError
		if errors.As(gerr, &gm) != errors.As(werr, &wm) || gm != nil && fmt.Sprint(gm.Unwrap()) != fmt.Sprint(wm.Unwrap()) {
			t.Errorf("Marshal of the output %q: got %v; want %v", in, gerr, werr)
		}
	}
}

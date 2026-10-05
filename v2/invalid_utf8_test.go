package json_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// A string of invalid UTF-8 stops every function of marshaling with the same error, and the calls after it work.
func TestMarshalInvalidUTF8(t *testing.T) {
	in := map[string]any{"a": []any{"ok", "\xff"}}
	_, want := json.Marshal(in)
	var serr *jsontext.SyntacticError
	if !errors.As(want, &serr) || serr.JSONPointer != "/a/1" {
		t.Fatalf("Marshal error = %v; want a SyntacticError at /a/1", want)
	}
	var buf bytes.Buffer
	if err := json.MarshalWrite(&buf, in); err == nil || err.Error() != want.Error() {
		t.Errorf("MarshalWrite error = %v; want %v", err, want)
	}
	enc := jsontext.NewEncoder(&buf)
	if err := json.MarshalEncode(enc, in); err == nil || err.Error() != want.Error() {
		t.Errorf("MarshalEncode error = %v; want %v", err, want)
	}
	for range 3 {
		out, err := json.Marshal(map[string]any{"a": []any{"ok"}})
		if err != nil || string(out) != `{"a":["ok"]}` {
			t.Fatalf("Marshal after the error = %s, %v", out, err)
		}
	}
}

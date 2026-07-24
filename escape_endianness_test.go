// escape_endianness_test.go
//
// Regression tests for the big-endian (e.g. s390x) string-escaping bug, where
// the SWAR escape scanner in internal/encoder/string.go mislocated the first
// byte needing escaping and emitted the first `"` of a value unescaped,
// producing invalid JSON

package json_test

import (
	stdjson "encoding/json"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	gojson "github.com/goccy/go-json"
)

// TestMarshalEmbeddedJSONStrings mirrors annotation values
// that are themselves serialized JSON objects/arrays. The first interior quote
// of each value must be escaped; on s390x it was not.
func TestMarshalEmbeddedJSONStrings(t *testing.T) {
	m := map[string]string{
		"Annotations":  `{"io.kubernetes.container.hash":"a539e690","restartCount":"0"}`,
		"Volumes":      `[{"container_path":"/host","host_path":"/","readonly":false}]`,
		"leadingQuote": `"x`, // quote as the very first byte: the s390x-broken case
		"plain":        "no escaping needed",
	}

	got, err := gojson.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !stdjson.Valid(got) {
		t.Fatalf("produced invalid JSON: %s", got)
	}

	var round map[string]string
	if err := stdjson.Unmarshal(got, &round); err != nil {
		t.Fatalf("output not parseable by encoding/json: %v\n%s", err, got)
	}
	if !reflect.DeepEqual(round, m) {
		t.Fatalf("round-trip mismatch:\n got %#v\nwant %#v", round, m)
	}
}

// TestMarshalStringEscapingMatchesStdlib checks a fixed set of strings that
// straddle 8-byte SWAR chunk boundaries, where the byte-offset computation is
// endianness-sensitive.
//
// Comparison is semantic (round-trip decode) rather than byte-equal because
// go-json intentionally encodes \b and \f as \u0008/\u000c while encoding/json
// uses the named escapes \b/\f. Both are valid per RFC 8259 §7 and decode to
// the same value.
func TestMarshalStringEscapingMatchesStdlib(t *testing.T) {
	cases := []string{
		``,
		`"`,
		`\`,
		`"abcdefg`,            // quote at byte 0 of chunk 0
		`abcdefg"`,            // quote at byte 7 of chunk 0
		`abcdefgh"ij`,         // quote at byte 0 of chunk 1
		`abcdefghijklmno"pqr`, // quote at byte 7 of chunk 1
		"\x00\x1f",            // control characters (includes \b 0x08, \f 0x0c)
		`<&>`,                 // HTML-significant characters
		strings.Repeat(`"`, 9),
		"héllo wörld",    // multibyte UTF-8
		"\xe4\xb8\xad",   // CJK byte sequence (中)
		"tab\tnewline\n", // common escapes
	}
	for _, s := range cases {
		got, err := gojson.Marshal(s)
		if err != nil {
			t.Fatalf("Marshal(%q): %v", s, err)
		}
		if !stdjson.Valid(got) {
			t.Errorf("Marshal(%q) produced invalid JSON: %s", s, got)
			continue
		}
		assertSemanticEqual(t, s, got)
	}
}

// FuzzMarshalStringMatchesStdlib asserts that goccy/go-json produces valid JSON
// that round-trips identically to encoding/json for all valid UTF-8 strings.
// This is the most effective check for the endianness bug: run it on s390x
// (see file header).
//
// Note: byte-for-byte equality with encoding/json is intentionally NOT
// required. go-json encodes \b (0x08) as \u0008 and \f (0x0c) as \u000c,
// while encoding/json uses the named escapes \b/\f. Both are valid per
// RFC 8259 §7 and decode to identical values. The contract here is semantic
// equivalence, not identical wire bytes.
func FuzzMarshalStringMatchesStdlib(f *testing.F) {
	seeds := []string{
		``, `"`, `\`, `{"a":"b"}`, "\x00\x1f", `<&>`,
		strings.Repeat(`"`, 9), `abcdefgh"ij`, "héllo", "\xe4\xb8\xad",
		"\b", "\f", "00\b", "00\f", // explicit seeds for the known \b/\f divergence
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			t.Skip() // invalid UTF-8: not part of this contract
		}
		got, err := gojson.Marshal(s)
		if err != nil {
			t.Fatalf("Marshal(%q) returned unexpected error: %v", s, err)
		}
		if !stdjson.Valid(got) {
			t.Fatalf("Marshal(%q) produced invalid JSON: %s", s, got)
		}
		assertSemanticEqual(t, s, got)
	})
}

// assertSemanticEqual decodes got (produced by go-json) and the encoding/json
// output for the same input s, then compares the decoded values with
// reflect.DeepEqual. This correctly handles cases where go-json and
// encoding/json choose different but equally valid escape sequences.
func assertSemanticEqual(t *testing.T, s string, got []byte) {
	t.Helper()
	want, _ := stdjson.Marshal(s)

	var gotVal, wantVal string
	if err := stdjson.Unmarshal(got, &gotVal); err != nil {
		t.Fatalf("could not decode go-json output for %q: %v — raw: %s", s, err, got)
	}
	if err := stdjson.Unmarshal(want, &wantVal); err != nil {
		t.Fatalf("could not decode stdlib output for %q: %v — raw: %s", s, err, want)
	}
	if !reflect.DeepEqual(gotVal, wantVal) {
		t.Fatalf("semantic mismatch for %q:\n got=%s (decoded: %q)\nwant=%s (decoded: %q)",
			s, got, gotVal, want, wantVal)
	}
}

// internal/encoder/string_escape_index_test.go
//
// goccy/go-json has no standalone escapeIndex function; the SWAR escape scan is
// inlined into the append*String functions in string.go. These tests drive
// those functions directly, which exercises the scan together with the
// stringToUint64Slice loader — specifically the s390x branch selected by the
// bigEndian compile-time constant in string.go.
//
// The boundary cases (quote at byte 0 / 7 / 8) are exactly where the s390x
// byte-offset computation used to be wrong, emitting the first `"` unescaped
// Output is compared against encoding/json, which goccy
// aims to match byte-for-byte.

package encoder

import (
	"bytes"
	stdjson "encoding/json"
	"strings"
	"testing"
)

// stdHTMLEscaped mirrors appendHTMLString: encoding/json HTML-escapes by default.
func stdHTMLEscaped(s string) []byte {
	b, err := stdjson.Marshal(s)
	if err != nil {
		panic(err)
	}
	return b
}

// stdNoHTMLEscaped mirrors appendString: HTML escaping disabled.
func stdNoHTMLEscaped(s string) []byte {
	var buf bytes.Buffer
	enc := stdjson.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		panic(err)
	}
	return bytes.TrimRight(buf.Bytes(), "\n") // Encode appends a trailing newline
}

func TestAppendStringEscaping(t *testing.T) {
	cases := []string{
		`"abcdefg`,                     // quote at byte 0 of chunk 0 (the s390x bug)
		`abcdefg"`,                     // quote at byte 7 of chunk 0
		`abcdefgh"ijk`,                 // quote at byte 0 of chunk 1
		`{"hash":"a539e690"}`,          // embedded-JSON value, like cri-o annotations
		`[{"container_path":"/host"}]`, // embedded-JSON array value
		"plain ascii, nothing special",
		"\x00\x01\x1f control chars",
		"tab\tnewline\ncr\r",
		`back\slash and "quotes"`,
		"unicode héllo wörld 中文 🚀",
		strings.Repeat(`"`, 9),
		"",
	}

	for _, s := range cases {
		if got, want := appendHTMLString(nil, s), stdHTMLEscaped(s); !bytes.Equal(got, want) {
			t.Errorf("appendHTMLString(%q):\n got = %s\n want = %s", s, got, want)
		}
		if got, want := appendString(nil, s), stdNoHTMLEscaped(s); !bytes.Equal(got, want) {
			t.Errorf("appendString(%q):\n got = %s\n want = %s", s, got, want)
		}
	}
}

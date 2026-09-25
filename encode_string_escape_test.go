package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"math/rand"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// randomTextString returns a random string of valid UTF-8 with runs of plain bytes of every length between the
// bytes which need an escape by one option or another.
func randomTextString(r *rand.Rand, n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		switch k := r.Intn(24); {
		case k == 0:
			b.WriteString([]string{"\n", "\r", "\t", `"`, `\`, "\x00", "\x1f", "\x7f", "\b", "\f"}[r.Intn(10)])
		case k == 1:
			b.WriteString([]string{"<", ">", "&"}[r.Intn(3)])
		case k == 2:
			b.WriteString([]string{"é", "あ", "😀", "\u2028", "\u2029"}[r.Intn(5)])
		default:
			b.WriteByte(byte(' ' + r.Intn(95)))
		}
	}
	return b.String()
}

func TestEncodeEscapedStrings(t *testing.T) {
	// Strings of every length with bytes to escape anywhere, encoded with every combination of the options of
	// the escapes, as encoding/json does for the ones it has, and decoded back to themselves for the others.
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 5000; i++ {
		s := randomTextString(r, r.Intn(i%300+1))
		v := struct{ S string }{s}

		want, _ := stdjson.Marshal(v)
		got, err := json.Marshal(v)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("%q:\n got %s %v\nwant %s", s, got, err, want)
		}

		var buf bytes.Buffer
		enc := stdjson.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.Encode(v)
		got, err = json.MarshalWithOption(v, json.DisableHTMLEscape())
		if err != nil || !bytes.Equal(got, bytes.TrimSuffix(buf.Bytes(), []byte("\n"))) {
			t.Fatalf("%q: without the escape of HTML:\n got %s %v\nwant %s", s, got, err, buf.Bytes())
		}

		for _, opts := range [][]json.EncodeOptionFunc{
			{json.DisableNormalizeUTF8()},
			{json.DisableNormalizeUTF8(), json.DisableHTMLEscape()},
		} {
			got, err := json.MarshalWithOption(v, opts...)
			if err != nil {
				t.Fatal(err)
			}
			var back struct{ S string }
			if err := stdjson.Unmarshal(got, &back); err != nil || back.S != s {
				t.Fatalf("%q: without the normalization of UTF-8: %s decodes to %q %v", s, got, back.S, err)
			}
		}
	}
}

func TestEncodeInvalidUTF8(t *testing.T) {
	// A byte of a string which is not valid UTF-8 is replaced by U+FFFD as encoding/json of the running Go does
	// ( escaped before Go 1.27, as it is by Go 1.27 ), in strings of every length and at every position, by the
	// scan of words and of SIMD, in values, in object keys and in the strings of the string option.
	type withString struct {
		S string `json:"s,string"`
	}
	invalid := []string{"\xff", "\xe3\x81", "\xed\xa0\x80"}
	others := []string{"", "<", " ", "あ", `"`}
	for n := 0; n < 100; n++ {
		for pos := 0; pos <= n; pos += 1 + n/16 {
			for _, bad := range invalid {
				for _, other := range others {
					s := strings.Repeat("a", n)
					s = s[:pos] + bad + other + s[pos:]
					values := []any{s, map[string]int{s: 1}, withString{S: s}}
					for _, v := range values {
						want, err := stdjson.Marshal(v)
						if err != nil {
							t.Fatal(err)
						}
						got, err := json.Marshal(v)
						if err != nil || string(got) != string(want) {
							t.Fatalf("Marshal(%q): got %q, %v, want %q", v, got, err, want)
						}
						wantIndent, _ := stdjson.MarshalIndent(v, "", " ")
						gotIndent, err := json.MarshalIndent(v, "", " ")
						if err != nil || string(gotIndent) != string(wantIndent) {
							t.Fatalf("MarshalIndent(%q): got %q, %v, want %q", v, gotIndent, err, wantIndent)
						}
						var wantBuf, gotBuf strings.Builder
						stdEnc := stdjson.NewEncoder(&wantBuf)
						stdEnc.SetEscapeHTML(false)
						_ = stdEnc.Encode(v)
						enc := json.NewEncoder(&gotBuf)
						enc.SetEscapeHTML(false)
						if err := enc.Encode(v); err != nil || gotBuf.String() != wantBuf.String() {
							t.Fatalf("Encode(%q) without HTML escapes: got %q, %v, want %q", v, gotBuf.String(), err, wantBuf.String())
						}
					}
				}
			}
		}
	}
}

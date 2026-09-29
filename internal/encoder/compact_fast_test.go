package encoder

import (
	"bytes"
	stdjson "encoding/json"
	"strconv"
	"strings"
	"testing"
)

// appendCompactOutput must accept an output only if compact leaves it as it is: what it accepts is copied.
// Whatever it rejects is compacted, so a rejection is never wrong, but the valid and compact outputs must be
// accepted for the speed.
func TestAppendCompactOutput(t *testing.T) {
	accepted := []string{
		`"hello"`, `""`, `"hello world"`, `0`, `-0`, `123`, `-12.5e+3`, `1E-2`, `true`, `false`, `null`,
		`{}`, `[]`, `{"a":1}`, `{"a":{"b":[1,2,{"c":null}]},"d":"e"}`, `[[[[]]]]`, `[1,"2",true,null,{"x":[]}]`,
		`{"a b":"c d"}`, strings.Repeat("[", 64) + strings.Repeat("]", 64),
		// escapes, and characters which are not ASCII, which compact leaves as they are.
		`"a\"b"`, `"a\\b"`, `"a\nb"`, `"\"\\\/\b\f\n\r\t\u00e9\uD83D\uDE00\uABCD"`, `{"\"k\"":"\\"}`, `["\\","\""]`,
		`"日本語"`, `"a — b"`, `{"é":"😀"}`, "\"\xff\xfe\"", `"\u2028"`,
		// several quotes in the eight bytes read at once, of which the first one in memory ends the string.
		`["x","y","z"]`, `["ab","cd","e"]`, `["a\\bcdefg","ijklmnop","q"]`, `{"k":"ab","l":"cdefghij"}`,
	}
	rejected := []string{
		``, ` `, `{ }`, `{"a": 1}`, `[1, 2]`, "[1,\n2]", `{"a":1} `, ` 1`, "\"a\tb\"",
		`{`, `[`, `}`, `]`, `{"a"}`, `{"a":}`, `{"a":1,}`, `[1,]`,
		`[1 2]`, `{"a":1 "b":2}`, `tru`, `nul`, `truex`, `01`, `1.`, `.5`, `1e`, `1e+`, `-`, `--1`, `+1`, `"abc`,
		`{"a":1}}`, `[1]]`, `{a:1}`, `{1:2}`, `[1]x`, "\x00", "\"\x01\"", strings.Repeat("[", 65) + strings.Repeat("]", 65),
		`{"a":1}{"b":2}`, `1 2`, `"a" "b"`,
		// several quotes in the eight bytes read at once: a string ended at a later one would make these valid.
		`""}a,:,{{`, `"\\]"[b]a}"`, `["ab" "cd"]`, `["ab" "c\\"]`, `{"k":"ab" "cd"}`,
		// escapes which are not ones of JSON, or which end the string.
		`"\"`, `"\`, `"\\"x`, `"\u00"`, `"\u00g0"`, `"\x41"`, `"\U0041"`, `"\'"`, `"\u12`, `["\"]`, `{"\":1}`,
		// commas and colons which are not in their places.
		`[,]`, `[1,,2]`, `{"a":1,,"b":2}`, `{,}`, `{"a"::1}`, `{"a":1:2}`, `[1:2]`, `,`, `:`, `{"a",1}`, `["a":1]`,
		// brackets which don't match.
		`[}`, `{]`, `[{]}`, `{"a":[}`,
		// values which are not values.
		`[undefined]`, `[NaN]`, `[Infinity]`, `[.1]`, `[1.]`, `[1e5.0]`, `[0x1]`, `[1_000]`, `['a']`, `[True]`,
	}
	for _, escape := range []bool{false, true} {
		for _, s := range accepted {
			// what compact does must be the same.
			compacted, err := compact([]byte("x"), append([]byte(s), nul), escape)
			if err != nil {
				t.Fatalf("escape=%v: compact of %q: %v", escape, s, err)
			}
			got, ok := appendCompactOutput([]byte("x"), []byte(s), escape, false)
			if !ok || string(got) != string(compacted) {
				t.Fatalf("escape=%v: %q must be accepted as %q, got %q, %v", escape, s, compacted, got, ok)
			}
		}
		for _, s := range rejected {
			if got, ok := appendCompactOutput([]byte("x"), []byte(s), escape, false); ok {
				t.Fatalf("escape=%v: %q must be rejected, got %q", escape, s, got)
			}
		}
	}
	// the characters which compact escapes for HTML are escaped as it escapes them, anywhere in a long output too.
	for _, s := range []string{`"<html>"`, `"a&b"`, "\"a\u2028b\"", "{\"\u2029\":1}", "\"\xe2\x80\xa8\"", "\"\xe2\x80\xa7\xe2\"",
		`["` + strings.Repeat("a<b>c&d — “e” \u2028 \\n", 20) + `"]`} {
		for _, escape := range []bool{false, true} {
			want, err := compact(nil, append([]byte(s), nul), escape)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := appendCompactOutput(nil, []byte(s), escape, false)
			if !ok || string(got) != string(want) {
				t.Fatalf("escape=%v: %q must be accepted as %q, got %q, %v", escape, s, want, got, ok)
			}
		}
	}
}

// Every byte in every position of a valid document must give the same answer as compact: accepted only if the
// document is compact, and then appended as compact appends it.
func TestAppendCompactOutputAgainstCompact(t *testing.T) {
	docs := []string{
		`{"ab":[1,-2.5e3,"cd",true,null,{}],"e":{"f":[[]]}}`, `"abc def"`, `12345`, `[true,false]`, `-0.5E+10`,
		`{"":""}`, `[[[{"a":[]}]]]`, `{"a":1,"b":2}`, `[1,2,3]`, `null`,
	}
	for _, doc := range docs {
		for pos := 0; pos < len(doc); pos++ {
			for c := 0; c < 256; c++ {
				b := []byte(doc)
				b[pos] = byte(c)
				for _, escape := range []bool{false, true} {
					compacted, err := compact(nil, append(append([]byte(nil), b...), nul), escape)
					unescaped, _ := compact(nil, append(append([]byte(nil), b...), nul), false)
					same := err == nil && bytes.Equal(unescaped, b)
					out, ok := appendCompactOutput(nil, b, escape, false)
					if ok && (!same || !bytes.Equal(out, compacted)) {
						t.Fatalf("escape=%v: %q is accepted as %q but compact gives %q, %v", escape, b, out, compacted, err)
					}
					// what is accepted must be valid by encoding/json: the check must be as strict as it.
					if ok && !stdjson.Valid(b) {
						t.Fatalf("escape=%v: %q is accepted but not valid", escape, b)
					}
					// compact accepts a number which is not one of JSON, such as 02: the check doesn't.
					if !ok && same && stdjson.Valid(b) && c >= 0x20 {
						t.Fatalf("escape=%v: %q is compact but rejected", escape, b)
					}
				}
			}
		}
	}
}

func BenchmarkCompactOutput(b *testing.B) {
	obj := `{"id":123456,"name":"alice","tags":["a","b","c"],"score":12.5,"active":true,"nested":{"x":1,"y":null}}`
	for _, s := range []string{`"hello"`, obj, "[" + strings.Repeat(obj+",", 20) + obj + "]"} {
		src := []byte(s)
		dst := make([]byte, 0, len(s)+16)
		b.Run("check/"+strconv.Itoa(len(s)), func(b *testing.B) {
			b.SetBytes(int64(len(s)))
			for i := 0; i < b.N; i++ {
				if _, ok := appendCompactOutput(dst[:0], src, true, false); !ok {
					b.Fatal("rejected")
				}
			}
		})
		withNul := append([]byte(s), nul)
		b.Run("compact/"+strconv.Itoa(len(s)), func(b *testing.B) {
			b.SetBytes(int64(len(s)))
			for i := 0; i < b.N; i++ {
				if _, err := compact(dst[:0], withNul, true); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// FuzzAppendCompactOutput checks that whatever is accepted is valid by encoding/json and is appended as compact
// appends it, so that the check is as strict as both.
func FuzzAppendCompactOutput(f *testing.F) {
	for _, s := range []string{`{"a":[1,"b",true,null,{"c":-1.5e3}]}`, `"x"`, `[]`, `0`, `{"k":"v","n":[1,2]}`, ` `, `[1,]`} {
		f.Add([]byte(s), true)
		f.Add([]byte(s), false)
	}
	f.Fuzz(func(t *testing.T, src []byte, escape bool) {
		out, ok := appendCompactOutput([]byte("x"), src, escape, false)
		if !ok {
			return
		}
		if !stdjson.Valid(src) {
			t.Fatalf("escape=%v: %q is accepted but not valid", escape, src)
		}
		compacted, err := compact([]byte("x"), append(append([]byte(nil), src...), nul), escape)
		if err != nil || !bytes.Equal(compacted, out) {
			t.Fatalf("escape=%v: %q is accepted as %q but compact gives %q, %v", escape, src, out, compacted, err)
		}
		// what encoding/json compacts must not be accepted either.
		var buf bytes.Buffer
		if stdjson.Compact(&buf, src) == nil && !bytes.Equal(buf.Bytes(), src) {
			t.Fatalf("escape=%v: %q is accepted but encoding/json compacts it to %q", escape, src, buf.Bytes())
		}
	})
}

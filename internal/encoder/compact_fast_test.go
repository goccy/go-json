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
	}
	rejected := []string{
		``, ` `, `{ }`, `{"a": 1}`, `[1, 2]`, "[1,\n2]", `{"a":1} `, ` 1`, `"a\"b"`, `"a\\b"`, `"a\nb"`,
		`"日本語"`, "\" \"", `{`, `[`, `}`, `]`, `{"a"}`, `{"a":}`, `{"a":1,}`, `[1,]`,
		`[1 2]`, `{"a":1 "b":2}`, `tru`, `nul`, `truex`, `01`, `1.`, `.5`, `1e`, `1e+`, `-`, `--1`, `+1`, `"abc`,
		`{"a":1}}`, `[1]]`, `{a:1}`, `{1:2}`, `[1]x`, "\x00", "\"\x01\"", strings.Repeat("[", 65) + strings.Repeat("]", 65),
		`{"a":1}{"b":2}`, `1 2`, `"a" "b"`,
		// an escape, alone or in a combination: the escapes are left to compact.
		`"\"`, `"\`, `"\\"x`, `"\u00"`, `["\"]`, `{"\":1}`,
		// commas and colons which are not in their places.
		`[,]`, `[1,,2]`, `{"a":1,,"b":2}`, `{,}`, `{"a"::1}`, `{"a":1:2}`, `[1:2]`, `,`, `:`, `{"a",1}`, `["a":1]`,
		// brackets which don't match.
		`[}`, `{]`, `[{]}`, `{"a":[}`,
		// values which are not values.
		`[undefined]`, `[NaN]`, `[Infinity]`, `[.1]`, `[1.]`, `[1e5.0]`, `[0x1]`, `[1_000]`, `['a']`, `[True]`,
	}
	for _, escape := range []bool{false, true} {
		for _, s := range accepted {
			if escape && strings.ContainsAny(s, "<>&") {
				continue
			}
			got, ok := appendCompactOutput([]byte("x"), []byte(s), escape)
			if !ok || string(got) != "x"+s {
				t.Fatalf("escape=%v: %q must be accepted, got %q, %v", escape, s, got, ok)
			}
			// what compact does must be the same.
			if compacted, err := compact(nil, append([]byte(s), nul), escape); err != nil || string(compacted) != s {
				t.Fatalf("escape=%v: compact of %q gives %q, %v", escape, s, compacted, err)
			}
		}
		for _, s := range rejected {
			if got, ok := appendCompactOutput([]byte("x"), []byte(s), escape); ok {
				t.Fatalf("escape=%v: %q must be rejected, got %q", escape, s, got)
			}
		}
	}
	// what is not escaped for HTML when the option is off is accepted then.
	for _, s := range []string{`"<html>"`, `"a&b"`} {
		if _, ok := appendCompactOutput(nil, []byte(s), false); !ok {
			t.Fatalf("%q must be accepted without the escape of HTML", s)
		}
		if _, ok := appendCompactOutput(nil, []byte(s), true); ok {
			t.Fatalf("%q must be rejected with the escape of HTML", s)
		}
	}
}

// Every byte in every position of a valid document must give the same answer as compact:
// accepted only if compact returns the document as it is.
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
					same := err == nil && bytes.Equal(compacted, b)
					_, ok := appendCompactOutput(nil, b, escape)
					if ok && !same {
						t.Fatalf("escape=%v: %q is accepted but compact gives %q, %v", escape, b, compacted, err)
					}
					// what is accepted must be valid by encoding/json: the check must be as strict as it.
					if ok && !stdjson.Valid(b) {
						t.Fatalf("escape=%v: %q is accepted but not valid", escape, b)
					}
					// compact accepts a number which is not one of JSON, such as 02: the check doesn't.
					if !ok && same && stdjson.Valid(b) && c < 0x80 && c >= 0x20 && c != '\\' && !(escape && (c == '<' || c == '>' || c == '&')) {
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
				if _, ok := appendCompactOutput(dst[:0], src, true); !ok {
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

// FuzzAppendCompactOutput checks that whatever is accepted is valid by encoding/json and is left as it is by
// compact, so that the check is as strict as both.
func FuzzAppendCompactOutput(f *testing.F) {
	for _, s := range []string{`{"a":[1,"b",true,null,{"c":-1.5e3}]}`, `"x"`, `[]`, `0`, `{"k":"v","n":[1,2]}`, ` `, `[1,]`} {
		f.Add([]byte(s), true)
		f.Add([]byte(s), false)
	}
	f.Fuzz(func(t *testing.T, src []byte, escape bool) {
		out, ok := appendCompactOutput([]byte("x"), src, escape)
		if !ok {
			return
		}
		if string(out) != "x"+string(src) {
			t.Fatalf("escape=%v: %q is not copied as it is: %q", escape, src, out)
		}
		if !stdjson.Valid(src) {
			t.Fatalf("escape=%v: %q is accepted but not valid", escape, src)
		}
		compacted, err := compact(nil, append(append([]byte(nil), src...), nul), escape)
		if err != nil || !bytes.Equal(compacted, src) {
			t.Fatalf("escape=%v: %q is accepted but compact gives %q, %v", escape, src, compacted, err)
		}
		// what encoding/json compacts must not be accepted either.
		var buf bytes.Buffer
		if stdjson.Compact(&buf, src) == nil && !bytes.Equal(buf.Bytes(), src) {
			t.Fatalf("escape=%v: %q is accepted but encoding/json compacts it to %q", escape, src, buf.Bytes())
		}
	})
}

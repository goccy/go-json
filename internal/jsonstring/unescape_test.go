package jsonstring

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

// randomStringLiteral is a JSON string of random pieces: plain characters, multi-byte ones, escapes of one
// character, \u escapes of any code, surrogate pairs and surrogates which are not in pairs, and invalid UTF-8.
func randomStringLiteral(r *rand.Rand) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for range r.IntN(24) {
		switch r.IntN(9) {
		case 0:
			sb.WriteString("plain text ")
		case 1:
			sb.WriteString("é漢🙂")
		case 2:
			sb.WriteString([]string{`\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`}[r.IntN(8)])
		case 3:
			fmt.Fprintf(&sb, `\u%04x`, r.IntN(0x10000))
		case 4:
			fmt.Fprintf(&sb, `\u%04X\u%04x`, 0xd800+r.IntN(0x400), 0xdc00+r.IntN(0x400))
		case 5:
			fmt.Fprintf(&sb, `\u%04x`, 0xd800+r.IntN(0x800))
		case 6:
			sb.WriteByte(byte(0x80 + r.IntN(0x80)))
		case 7:
			sb.WriteString(strings.Repeat("x", r.IntN(40)))
		default:
			fmt.Fprintf(&sb, `\u%04x\u%04x`, 0xd800+r.IntN(0x400), r.IntN(0x10000))
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// AppendUnescaped and ReplaceInvalidUTF8 make the value which encoding/json decodes a string to.
func TestAppendUnescaped(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for range 50000 {
		lit := randomStringLiteral(r)
		var want string
		if err := json.Unmarshal([]byte(lit), &want); err != nil {
			t.Fatalf("encoding/json can't decode %q: %v", lit, err)
		}
		b := []byte(lit)
		for _, prefix := range []string{"", "xyz"} {
			got := AppendUnescaped([]byte(prefix), b[1:len(b)-1])
			got = ReplaceInvalidUTF8(got, len(prefix))
			if string(got) != prefix+want {
				t.Fatalf("value of %q: got %q, want %q", lit, got[len(prefix):], want)
			}
		}
	}
}

// Unquote takes a JSON string as encoding/json decodes it, and rejects the bytes which are not one.
func TestUnquote(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 6))
	for range 20000 {
		lit := randomStringLiteral(r)
		var want string
		if err := json.Unmarshal([]byte(lit), &want); err != nil {
			t.Fatalf("encoding/json can't decode %q: %v", lit, err)
		}
		got, ok := Unquote([]byte(lit))
		if !ok || string(got) != want {
			t.Fatalf("Unquote(%q) = %q, %v; want %q", lit, got, ok, want)
		}
	}
	for _, tc := range []struct {
		lit  string
		want string
		ok   bool
	}{
		{`""`, "", true},
		{`"abc"`, "abc", true},
		{`"\'"`, "'", true},
		{`"aA"`, "aA", true},
		{"\"\xff\"", "�", true},
		{`"`, "", false},
		{`abc`, "", false},
		{`"a"b"`, "", false},
		{"\"a\nb\"", "", false},
		{`"\q"`, "", false},
		{`"\"`, "", false},
		{`"\u12"`, "", false},
		{`"\u12zz"`, "", false},
		{`"\ud800\uzzzz"`, "", false},
	} {
		got, ok := Unquote([]byte(tc.lit))
		if ok != tc.ok || string(got) != tc.want {
			t.Errorf("Unquote(%q) = %q, %v; want %q, %v", tc.lit, got, ok, tc.want, tc.ok)
		}
	}
}

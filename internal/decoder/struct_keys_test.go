package decoder

import (
	"encoding/binary"
	"math/rand"
	"testing"
	"unicode/utf8"
)

// keyLengthInWordByBytes is keyLengthInWord byte by byte.
func keyLengthInWordByBytes(w uint64) int {
	for i := 0; i < 8; i++ {
		c := byte(w >> (8 * i))
		if c == '"' || c == '\\' || c < 0x20 || c >= 0x80 {
			return i
		}
	}
	return 8
}

func TestKeyLengthInWord(t *testing.T) {
	// Every byte at every position, followed and preceded by random bytes, and random words whose bytes are
	// mostly the ones of keys.
	r := rand.New(rand.NewSource(1))
	var b [8]byte
	check := func() {
		t.Helper()
		w := binary.LittleEndian.Uint64(b[:])
		if got, want := keyLengthInWord(w), keyLengthInWordByBytes(w); got != want {
			t.Fatalf("%q: got %d, want %d", b[:], got, want)
		}
	}
	for pos := 0; pos < 8; pos++ {
		for c := 0; c < 256; c++ {
			for i := 0; i < 16; i++ {
				for j := range b {
					switch {
					case j < pos:
						b[j] = byte(0x20 + r.Intn(0x60))
					case j == pos:
						b[j] = byte(c)
					default:
						b[j] = byte(r.Intn(256))
					}
				}
				check()
			}
		}
	}
	chars := []byte("aZ09_-\"\\\x00\x1f\x20\x7f\x80\xff")
	for i := 0; i < 1000000; i++ {
		for j := range b {
			b[j] = chars[r.Intn(len(chars))]
		}
		check()
	}
}

func TestStructKeysShortWithNul(t *testing.T) {
	// A key of a field with a nul byte is not found by a shorter key whose word is the same.
	d := newStructDecoder("", "")
	withNul := &structFieldSet{key: "a\x00"}
	plain := &structFieldSet{key: "b"}
	d.setFields([]*structFieldSet{withNul, plain})
	for _, tc := range []struct {
		in   string
		want *structFieldSet
	}{
		{`"a":1`, nil},
		{`"A":1`, nil},
		{`"b":1`, plain},
		{`"B":1`, plain},
		{`"a\u0000":1`, withNul},
	} {
		buf := make([]byte, len(tc.in)+1, len(tc.in)+64)
		copy(buf, tc.in)
		field, _, err := d.decodeKey(buf, 0, false)
		if err != nil || field != tc.want {
			t.Fatalf("%s: got %v %v, want %v", tc.in, field, err, tc.want)
		}
	}
}

func TestStructKeysMayFold(t *testing.T) {
	// mayFold tells by the bytes of a key what mayFoldDecoded tells by its runes, for a key of valid UTF-8.
	var fields []*structFieldSet
	for _, key := range []string{"k", "s", "名前", "Ärger", "ſ", "Ωmega", "𝔸x"} {
		fields = append(fields, &structFieldSet{key: key})
	}
	k := newStructKeys(fields)
	if k.hasRuneError {
		t.Fatal("no key has utf8.RuneError")
	}
	r := rand.New(rand.NewSource(1))
	runes := []rune{'a', 'K', 'S', 'K', 'ſ', '名', '前', '後', 'Ä', 'ä', 'Ω', 'ω', 'Ω', '𝔸', '𝔹', 'é', 'ÿ', 'Ā'}
	for i := 0; i < 100000; i++ {
		var key []byte
		for j := r.Intn(4) + 1; j > 0; j-- {
			if r.Intn(4) == 0 {
				key = utf8.AppendRune(key, rune(r.Intn(0x10ffff)))
			} else {
				key = utf8.AppendRune(key, runes[r.Intn(len(runes))])
			}
		}
		if got, want := k.mayFold(key), k.mayFoldDecoded(key); got != want {
			t.Fatalf("%q: got %v, want %v", key, got, want)
		}
		// the check of the first rune never rejects a key which may fold
		var w [8]byte
		copy(w[:], key)
		if k.mayFold(key) && !k.firstRuneMayFold(key, binary.LittleEndian.Uint64(w[:])) {
			t.Fatalf("%q: rejected by its first rune", key)
		}
	}
	// a key of runes which fold to a key may be of it
	for _, key := range []string{"K", "ſ", "名前", "äRGER", "Ωmega", "𝔸X"} {
		if !k.mayFold([]byte(key)) {
			t.Fatalf("%q: not found", key)
		}
	}
}

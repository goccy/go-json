package decoder

import (
	"encoding/binary"
	"math/bits"
	"math/rand"
	"testing"
	"unicode/utf8"
)

// firstByteOf returns the position in the word of its first byte which is special, or 8 if it has none.
func firstByteOf(w uint64, special func(c byte) bool) int {
	for i := 0; i < 8; i++ {
		if special(byte(w >> (8 * i))) {
			return i
		}
	}
	return 8
}

func TestKeyEndMasks(t *testing.T) {
	// The first byte of the masks is the first byte which ends a simple key ( specialKeyBytes ) or a string
	// ( keyEndBytes ): the bytes after it may be wrong.
	// Every byte at every position, followed and preceded by random bytes, and random words whose bytes are
	// mostly the ones of keys.
	r := rand.New(rand.NewSource(1))
	var b [8]byte
	check := func() {
		t.Helper()
		w := binary.LittleEndian.Uint64(b[:])
		if got, want := bits.TrailingZeros64(specialKeyBytes(w))/8, firstByteOf(w, func(c byte) bool {
			return c == '"' || c == '\\' || c < 0x20 || c >= 0x80
		}); got != want {
			t.Fatalf("specialKeyBytes %q: got %d, want %d", b[:], got, want)
		}
		if got, want := bits.TrailingZeros64(keyEndBytes(w))/8, firstByteOf(w, func(c byte) bool {
			return c == '"' || c == '\\' || c < 0x20
		}); got != want {
			t.Fatalf("keyEndBytes %q: got %d, want %d", b[:], got, want)
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
		// no check rejects a key which folds to a key of a field
		var w [8]byte
		copy(w[:], key)
		w0 := binary.LittleEndian.Uint64(w[:])
		if k.find(appendFoldedKey(nil, key)) != nil && !(k.leadMayFold(w0) && k.firstRuneMayFold(key, w0) && k.mayFold(key)) {
			t.Fatalf("%q: the key of a field is rejected", key)
		}
	}
	// a key of runes which fold to a key may be of it
	for _, key := range []string{"K", "ſ", "名前", "äRGER", "Ωmega", "𝔸X"} {
		if !k.mayFold([]byte(key)) {
			t.Fatalf("%q: not found", key)
		}
	}
}

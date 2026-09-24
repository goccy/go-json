package decoder

import (
	"encoding/binary"
	"math/rand"
	"testing"
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

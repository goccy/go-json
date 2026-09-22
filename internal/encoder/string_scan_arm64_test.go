package encoder

import (
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

func TestScanStringNEON(t *testing.T) {
	special := []byte{0x00, 0x1f, '"', '\\', '<', '>', '&', 0x7f, 0x80, 0xe3, 0xff, '\n', ' ', '!', '#', '=', '?', '[', ']'}
	for index := range stringEscapes {
		e := &stringEscapes[index]
		expected := func(s string) int {
			for i := 0; i < len(s); i++ {
				if e.table[s[i]] {
					return 1
				}
			}
			return 0
		}
		for length := 16; length <= 100; length++ {
			base := strings.Repeat("a", length)
			check := func(s string) {
				t.Helper()
				got := scanStringNEON(unsafe.Pointer(unsafe.StringData(s)), len(s), &e.tables)
				if expected := expected(s); got != expected {
					t.Fatalf("escape %d, %q: expected %d but got %d", index, s, expected, got)
				}
			}
			check(base)
			for pos := 0; pos < length; pos++ {
				for _, c := range special {
					b := []byte(base)
					b[pos] = c
					check(string(b))
				}
			}
		}
	}
}

// The scan must not read out of the string.
func TestScanStringNEONAtEndOfMemory(t *testing.T) {
	for length := 16; length <= 80; length++ {
		mem := []byte(strings.Repeat("a", length))
		for start := 0; start+16 <= length; start++ {
			s := string(mem[start:])
			if got := scanStringNEON(unsafe.Pointer(unsafe.StringData(s)), len(s), &stringEscapes[stringEscapeHTML|stringEscapeNormalize].tables); got != 0 {
				t.Fatalf("%q: got %d", s, got)
			}
		}
	}
}

// the scans alone: the one by words and the one by SIMD.
func BenchmarkScanString(b *testing.B) {
	e := &stringEscapes[stringEscapeHTML|stringEscapeNormalize]
	for _, n := range []int{16, 36, 100, 1000} {
		s := strings.Repeat("a", n)
		p := unsafe.Pointer(unsafe.StringData(s))
		b.Run("Words/"+strconv.Itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if e.hasLooseEscape(p, n) {
					b.Fatal("found")
				}
			}
		})
		b.Run("NEON/"+strconv.Itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if scanStringNEON(p, n, &e.tables) != 0 {
					b.Fatal("found")
				}
			}
		})
	}
}

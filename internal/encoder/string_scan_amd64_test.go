package encoder

import (
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

func TestScanStringAVX2(t *testing.T) {
	if !runtime.HasAVX2 {
		t.Skip("AVX2 is not supported")
	}
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
		for length := 32; length <= 130; length++ {
			base := strings.Repeat("a", length)
			check := func(s string) {
				t.Helper()
				got := scanStringAVX2(unsafe.Pointer(unsafe.StringData(s)), len(s), &e.tables)
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
					if pos+1 < length {
						b[pos+1] = '"'
						check(string(b))
					}
				}
			}
		}
	}
}

// The scan must not read out of the string.
func TestScanStringAVX2AtEndOfMemory(t *testing.T) {
	if !runtime.HasAVX2 {
		t.Skip("AVX2 is not supported")
	}
	for length := 32; length <= 100; length++ {
		mem := []byte(strings.Repeat("a", length))
		for start := 0; start+32 <= length; start++ {
			s := string(mem[start:])
			if got := scanStringAVX2(unsafe.Pointer(unsafe.StringData(s)), len(s), &stringEscapes[stringEscapeHTML|stringEscapeNormalize].tables); got != 0 {
				t.Fatalf("%q: got %d", s, got)
			}
		}
	}
}

// the scans alone: the one by words and the one by SIMD.
func BenchmarkScanString(b *testing.B) {
	e := &stringEscapes[stringEscapeHTML|stringEscapeNormalize]
	for _, n := range []int{36, 100, 1000} {
		s := strings.Repeat("a", n)
		p := unsafe.Pointer(unsafe.StringData(s))
		b.Run("Words/"+strconv.Itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if e.hasLooseEscape(p, n) {
					b.Fatal("found")
				}
			}
		})
		if runtime.HasAVX2 {
			b.Run("AVX2/"+strconv.Itoa(n), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if scanStringAVX2(p, n, &e.tables) != 0 {
						b.Fatal("found")
					}
				}
			})
		}
	}
}

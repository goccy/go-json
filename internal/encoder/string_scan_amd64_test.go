package encoder

import (
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

func TestScanStringAVX2(t *testing.T) {
	if !hasAVX2 {
		t.Skip("AVX2 is not supported")
	}
	special := []byte{0x00, 0x1f, '"', '\\', '<', '>', '&', 0x7f, 0x80, 0xe3, 0xff, '\n', ' ', '!', '#', '=', '?', '[', ']'}
	for index := range stringEscapes {
		e := &stringEscapes[index]
		chars := uint64(uint8(e.chars[0])) | uint64(uint8(e.chars[1]))<<8 | uint64(uint8(e.chars[2]))<<16
		expectedIndex := func(s string) int {
			for i := 0; i < len(s); i++ {
				if e.table[s[i]] {
					return i
				}
			}
			return len(s)
		}
		for length := 32; length <= 130; length++ {
			base := strings.Repeat("a", length)
			check := func(s string) {
				t.Helper()
				got := scanStringAVX2(unsafe.Pointer(unsafe.StringData(s)), len(s), chars, e.high)
				if expected := expectedIndex(s); got != expected {
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
	if !hasAVX2 {
		t.Skip("AVX2 is not supported")
	}
	for length := 32; length <= 100; length++ {
		mem := []byte(strings.Repeat("a", length))
		for start := 0; start+32 <= length; start++ {
			s := string(mem[start:])
			if got := scanStringAVX2(unsafe.Pointer(unsafe.StringData(s)), len(s), 0x262e3c, msb); got != len(s) {
				t.Fatalf("%q: got %d", s, got)
			}
		}
	}
}

// the scans alone, to find where the time of the SIMD scan goes.
func BenchmarkScanString(b *testing.B) {
	e := &stringEscapes[stringEscapeHTML|stringEscapeNormalize]
	chars := uint64(uint8(e.chars[0])) | uint64(uint8(e.chars[1]))<<8 | uint64(uint8(e.chars[2]))<<16
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
		if hasAVX2 {
			b.Run("AVX2/"+strconv.Itoa(n), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if scanStringAVX2(p, n, chars, e.high) != n {
						b.Fatal("found")
					}
				}
			})
			b.Run("SSE/"+strconv.Itoa(n), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if scanStringSSE(p, n, chars, e.high) != n {
						b.Fatal("found")
					}
				}
			})
		}
	}
}

func TestScanStringSSE(t *testing.T) {
	if !hasAVX2 {
		t.Skip("AVX2 is not supported")
	}
	e := &stringEscapes[stringEscapeHTML]
	chars := uint64(uint8(e.chars[0])) | uint64(uint8(e.chars[1]))<<8 | uint64(uint8(e.chars[2]))<<16
	for length := 16; length <= 70; length++ {
		for pos := -1; pos < length; pos++ {
			b := []byte(strings.Repeat("a", length))
			expected := length
			if pos >= 0 {
				b[pos] = '<'
				expected = pos
			}
			if got := scanStringSSE(unsafe.Pointer(&b[0]), length, chars, e.high); got != expected {
				t.Fatalf("%q: expected %d but got %d", b, expected, got)
			}
		}
	}
}

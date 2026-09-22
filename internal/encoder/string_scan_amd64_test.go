package encoder

import (
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

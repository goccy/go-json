package jsonstring

import (
	"math/bits"
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

// The candidates of the optimizations of the escapes of strings, which are measured against what is used now on
// the machine of the CI: amd64 can't be measured on the machine they are developed on.

// escapeMask returns the mask of the bytes of n which may need an escape: a control character, '"', '\\' and
// a byte which is not ASCII. The most significant bit of such a byte is set.
//
// A byte after one which needs an escape may be set too, because of the borrow of the subtraction,
// so only the first byte of the mask is exact.
func escapeMask(n uint64) uint64 {
	// `n` is in the mask to check whether the most significant bit of a byte of the input is set
	// ( the byte is outside of ASCII ).
	mask := n | (n - (lsb * 0x20)) |
		((n ^ (lsb * '"')) - lsb) |
		((n ^ (lsb * '\\')) - lsb)
	return mask & msb
}

// escapeMaskHTML is escapeMask which also has '<', '>' and '&'.
func escapeMaskHTML(n uint64) uint64 {
	mask := n | (n - (lsb * 0x20)) |
		((n ^ (lsb * '"')) - lsb) |
		((n ^ (lsb * '\\')) - lsb) |
		((n ^ (lsb * '<')) - lsb) |
		((n ^ (lsb * '>')) - lsb) |
		((n ^ (lsb * '&')) - lsb)
	return mask & msb
}

// shortStringWord returns the bytes of a string of 1 to 7 bytes as a word, without reading out of the string,
// and the index of the string which each byte of the word is of. The bytes which are not used are 'a'.
func shortStringWord(p unsafe.Pointer, n int) (uint64, [8]uint8) {
	if n >= 4 {
		// the two halves overlap.
		word := uint64(*(*uint32)(p)) | uint64(*(*uint32)(unsafe.Add(p, n-4)))<<32
		m := uint8(n - 4)
		return word, [8]uint8{0, 1, 2, 3, m, m + 1, m + 2, m + 3}
	}
	// the first, the middle and the last byte are every byte of 1 to 3 bytes.
	word := uint64(*(*uint8)(p)) |
		uint64(*(*uint8)(unsafe.Add(p, n>>1)))<<8 |
		uint64(*(*uint8)(unsafe.Add(p, n-1)))<<16 |
		0x6161616161000000
	return word, [8]uint8{0, uint8(n >> 1), uint8(n - 1)}
}

// firstEscape returns the index of the first byte of s which may need an escape, or len(s) if there is none.
// The string is always read by words: the last word of a string overlaps the previous one, and a string shorter
// than a word is gathered into a word.
func firstEscape(s string) int {
	n := len(s)
	p := unsafe.Pointer(unsafe.StringData(s))
	if n < 8 {
		if n == 0 {
			return 0
		}
		word, index := shortStringWord(p, n)
		if mask := escapeMask(word); mask != 0 {
			return int(index[bits.TrailingZeros64(mask)/8])
		}
		return n
	}
	i := 0
	for ; i+8 <= n; i += 8 {
		if mask := escapeMask(*(*uint64)(unsafe.Add(p, i))); mask != 0 {
			return i + bits.TrailingZeros64(mask)/8
		}
	}
	if i < n {
		if mask := escapeMask(*(*uint64)(unsafe.Add(p, n-8))); mask != 0 {
			return n - 8 + bits.TrailingZeros64(mask)/8
		}
	}
	return n
}

// firstEscapeHTML is firstEscape which also finds '<', '>' and '&'.
func firstEscapeHTML(s string) int {
	n := len(s)
	p := unsafe.Pointer(unsafe.StringData(s))
	if n < 8 {
		if n == 0 {
			return 0
		}
		word, index := shortStringWord(p, n)
		if mask := escapeMaskHTML(word); mask != 0 {
			return int(index[bits.TrailingZeros64(mask)/8])
		}
		return n
	}
	i := 0
	for ; i+8 <= n; i += 8 {
		if mask := escapeMaskHTML(*(*uint64)(unsafe.Add(p, i))); mask != 0 {
			return i + bits.TrailingZeros64(mask)/8
		}
	}
	if i < n {
		if mask := escapeMaskHTML(*(*uint64)(unsafe.Add(p, n-8))); mask != 0 {
			return n - 8 + bits.TrailingZeros64(mask)/8
		}
	}
	return n
}

// appendStringByWords is appendNormalizedHTMLString which scans the string only by words.
func appendStringByWords(buf []byte, s string) []byte {
	if j := firstEscapeHTML(s); j != len(s) {
		return variantNormalizedHTML(buf, s)
	}
	buf = append(buf, '"')
	return append(append(buf, s...), '"')
}

// appendStringReserved is appendStringByWords which checks the capacity once and writes the quotes directly.
func appendStringReserved(buf []byte, s string) []byte {
	if j := firstEscapeHTML(s); j != len(s) {
		return variantNormalizedHTML(buf, s)
	}
	l := len(buf)
	if cap(buf)-l < len(s)+2 {
		buf = variantReserve(buf, len(s)+2)
	}
	buf = buf[:l+len(s)+2]
	buf[l] = '"'
	copy(buf[l+1:], s)
	buf[l+1+len(s)] = '"'
	return buf
}

// appendStringShortCopy scans as appendNormalizedHTMLString does ( by bytes under a word ), but it checks the
// capacity once, writes the quotes directly and copies a string up to 16 bytes without calling memmove.
func appendStringShortCopy(buf []byte, s string) []byte {
	n := len(s)
	l := len(buf)
	if cap(buf)-l < n+2 {
		buf = variantReserve(buf, n+2)
	}
	src := unsafe.Pointer(unsafe.StringData(s))
	if n < 8 {
		for i := 0; i < n; i++ {
			if needEscapeHTMLNormalizeUTF8[*(*byte)(unsafe.Add(src, i))] {
				return variantNormalizedHTML(buf, s)
			}
		}
	} else {
		i := 0
		for ; i+8 <= n; i += 8 {
			if escapeMaskHTML(*(*uint64)(unsafe.Add(src, i))) != 0 {
				return variantNormalizedHTML(buf, s)
			}
		}
		if i < n && escapeMaskHTML(*(*uint64)(unsafe.Add(src, n-8))) != 0 {
			return variantNormalizedHTML(buf, s)
		}
	}
	buf = buf[:l+n+2]
	dst := unsafe.Pointer(unsafe.SliceData(buf[l:]))
	*(*byte)(dst) = '"'
	dst = unsafe.Add(dst, 1)
	switch {
	case n > 16:
		copy(buf[l+1:], s)
	case n >= 8:
		// the two words overlap.
		*(*uint64)(dst) = *(*uint64)(src)
		*(*uint64)(unsafe.Add(dst, n-8)) = *(*uint64)(unsafe.Add(src, n-8))
	case n >= 4:
		*(*uint32)(dst) = *(*uint32)(src)
		*(*uint32)(unsafe.Add(dst, n-4)) = *(*uint32)(unsafe.Add(src, n-4))
	case n > 0:
		// the first, the middle and the last byte are every byte of 1 to 3 bytes.
		*(*byte)(dst) = *(*byte)(src)
		*(*byte)(unsafe.Add(dst, n>>1)) = *(*byte)(unsafe.Add(src, n>>1))
		*(*byte)(unsafe.Add(dst, n-1)) = *(*byte)(unsafe.Add(src, n-1))
	}
	*(*byte)(unsafe.Add(dst, n)) = '"'
	return buf
}

//go:noinline
func variantReserve(b []byte, n int) []byte {
	grown := make([]byte, len(b), 2*cap(b)+n)
	copy(grown, b)
	return grown
}

var variantStrings = []string{
	"", "abc", "test42", "active", "127.0.0.1", "user_agent_long", "de305d54-75b4-431b-adb2-eb6b9e546014",
	strings.Repeat("abcdefghij", 10), strings.Repeat("abcdefghij", 100),
}

// The escape functions of the escapers of encoding/json, as functions of the buffer and the string.

func variantNormalizedHTML(buf []byte, s string) []byte {
	buf, _ = appendNormalizedHTMLString(V1Escaper(1, 1), buf, s)
	return buf
}

func variantHTML(buf []byte, s string) []byte {
	buf, _ = appendHTMLString(V1Escaper(1, 0), buf, s)
	return buf
}

func variantNormalized(buf []byte, s string) []byte {
	buf, _ = appendNormalizedString(V1Escaper(0, 1), buf, s)
	return buf
}

func variantPlain(buf []byte, s string) []byte {
	buf, _ = appendString(V1Escaper(0, 0), buf, s)
	return buf
}

func BenchmarkVariant_String(b *testing.B) {
	for _, s := range variantStrings {
		for _, variant := range []struct {
			name string
			f    func([]byte, string) []byte
		}{
			{"Current", variantNormalizedHTML},
			{"ByWords", appendStringByWords},
			{"Reserved", appendStringReserved},
			{"ShortCopy", appendStringShortCopy},
		} {
			s, f := s, variant.f
			b.Run(variant.name+"/"+strconv.Itoa(len(s)), func(b *testing.B) {
				buf := make([]byte, 0, 4096)
				for i := 0; i < b.N; i++ {
					buf = f(buf[:0], s)
				}
			})
		}
	}
}

func TestVariantsEncodeTheSame(t *testing.T) {
	var inputs []string
	for length := 0; length <= 40; length++ {
		inputs = append(inputs, strings.Repeat("a", length))
		for pos := 0; pos < length; pos++ {
			for _, c := range []byte{'"', '<', 0x01, 0xe3} {
				b := []byte(strings.Repeat("b", length))
				b[pos] = c
				inputs = append(inputs, string(b))
			}
		}
	}
	for _, s := range inputs {
		expected := string(variantNormalizedHTML([]byte("x"), s))
		for name, f := range map[string]func([]byte, string) []byte{
			"ByWords":   appendStringByWords,
			"Reserved":  appendStringReserved,
			"ShortCopy": appendStringShortCopy,
		} {
			// without and with the capacity.
			for _, buf := range [][]byte{[]byte("x"), append(make([]byte, 0, 128), 'x')} {
				if got := string(f(buf, s)); got != expected {
					t.Fatalf("%s(%q): expected %s but got %s", name, s, expected, got)
				}
			}
		}
	}
}

// the four functions which AppendQuoted chooses by the options must cost the same for a string without an escape.
func BenchmarkVariant_StringFlags(b *testing.B) {
	for _, s := range []string{"test42", "user_agent_long", "de305d54-75b4-431b-adb2-eb6b9e546014"} {
		for _, variant := range []struct {
			name string
			f    func([]byte, string) []byte
		}{
			{"NormalizedHTML", variantNormalizedHTML},
			{"HTML", variantHTML},
			{"Normalized", variantNormalized},
			{"Plain", variantPlain},
		} {
			s, f := s, variant.f
			b.Run(variant.name+"/"+strconv.Itoa(len(s)), func(b *testing.B) {
				buf := make([]byte, 0, 4096)
				for i := 0; i < b.N; i++ {
					buf = f(buf[:0], s)
				}
			})
		}
	}
}

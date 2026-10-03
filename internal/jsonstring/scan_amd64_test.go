package jsonstring

import (
	"math/rand"
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

func TestAppendEscapedSIMD(t *testing.T) {
	// The functions of the escapes append the same bytes with the loop of the escapes by SIMD as without it, for
	// strings of every length with bytes to escape by every option anywhere, including strings which are not
	// valid UTF-8 and strings mostly of characters which are not ASCII, across the blocks and the segments.
	if !runtime.HasAVX2 {
		t.Skip("AVX2 is not supported")
	}
	defer func() { hasEscapeLoop = true }()
	r := rand.New(rand.NewSource(1))
	pieces := []string{"a", "Z", "0", " ", "\n", "\r", "\t", "\"", "\\", "\x00", "\x08", "\x0c", "\x1f", "\x7f", "<", ">", "&",
		"é", "あ", "語", "😀", "\U0010ffff", "\u2027", "\u2028", "\u2029", "\u202a", "\ufffd", "\xff", "\x80", "\xc3", "\xe3\x81",
		"\xf0\x9f", "\xc0\x80", "\xe0\x80\x80", "\xed\xa0\x80", "\xf0\x80\x80\x80", "\xf4\x90\x80\x80", "\xf5\x80\x80\x80"}
	check := func(s string) {
		t.Helper()
		for index := range stringEscapes {
			e := &stringEscapes[index]
			hasEscapeLoop = false
			want, wantValid := e.appendEscaped(e, []byte("prefix"), s)
			hasEscapeLoop = true
			got, gotValid := e.appendEscaped(e, []byte("prefix"), s)
			if string(got) != string(want) || gotValid != wantValid {
				t.Fatalf("%d %q:\n got %q\nwant %q", index, s, got, want)
			}
		}
	}
	for n := 0; n < 3000; n++ {
		var b strings.Builder
		for k := r.Intn(n%400 + 1); k > 0; k-- {
			if r.Intn(6) == 0 {
				b.WriteString(pieces[r.Intn(len(pieces))])
			} else {
				b.WriteByte(byte('a' + r.Intn(26)))
			}
		}
		check(b.String())
	}
	valid := []string{"é", "ß", "あ", "語", "😀", "\U0010ffff", "\u2027", "\u202a", "a", " "}
	for n := 0; n < 1000; n++ {
		var b strings.Builder
		for k := r.Intn(n%3000 + 1); k > 0; k-- {
			if r.Intn(200) == 0 {
				b.WriteString(pieces[r.Intn(len(pieces))])
			} else {
				b.WriteString(valid[r.Intn(len(valid))])
			}
		}
		check(b.String())
	}
}

func TestAppendHTMLEscapedSIMD(t *testing.T) {
	// The output of a marshaler is escaped for HTML the same with the loop of the escapes by SIMD as without it,
	// with the characters to escape and the other characters which start as U+2028 and U+2029 do anywhere.
	if !runtime.HasAVX2 {
		t.Skip("AVX2 is not supported")
	}
	defer func() { hasEscapeLoop = true }()
	r := rand.New(rand.NewSource(1))
	pieces := []string{"<", ">", "&", "\u2028", "\u2029", "\u2027", "—", "“", "\xe2", "\xe2\x80", "é", "\\n", "\\\""}
	for n := 0; n < 3000; n++ {
		var b strings.Builder
		b.WriteByte('"')
		for k := r.Intn(n%400 + 1); k > 0; k-- {
			if r.Intn(8) == 0 {
				b.WriteString(pieces[r.Intn(len(pieces))])
			} else {
				b.WriteByte(byte('a' + r.Intn(26)))
			}
		}
		b.WriteByte('"')
		src := []byte(b.String())
		hasEscapeLoop = false
		want := AppendHTMLEscaped([]byte("prefix"), src)
		hasEscapeLoop = true
		got := AppendHTMLEscaped([]byte("prefix"), src)
		if string(got) != string(want) {
			t.Fatalf("%q:\n got %q\nwant %q", src, got, want)
		}
	}
}

// TestAppendNormalizedStrings without the loop of the escapes by SIMD.
func TestAppendNormalizedStringsWithoutEscapeLoop(t *testing.T) {
	defer func(loop bool) { hasEscapeLoop = loop }(hasEscapeLoop)
	hasEscapeLoop = false
	checkAppendNormalizedStrings(t)
}

// A character which goes on after a block of the loop of the escapes of UTF-8, valid or not, followed by a block
// of ASCII, of characters which are not ASCII or with a byte to escape, is appended as it is without the loop, at
// every position around the ends of the blocks: the loop validates it with the next block, or stops before it.
func TestEscapeUTF8AVX2AcrossBlocks(t *testing.T) {
	if !runtime.HasAVX2 {
		t.Skip("AVX2 is not supported")
	}
	defer func(loop bool) { hasEscapeLoop = loop }(hasEscapeLoop)
	text := strings.Repeat("あいう", 40)
	prefix := func(n int) string {
		// n bytes of characters of three bytes, and ASCII after the last whole one.
		s := text[:n/3*3]
		return s + strings.Repeat("x", n-len(s))
	}
	pieces := []string{"", "\xe3", "\xe3\x81", "\xe3\x81\x82", "\xf0\x9f", "\xf0\x9f\x98", "\xf0\x9f\x98\x80", "\xc3", "\xc3\xa9",
		"\xe2\x80\xa8", "\xe2\x80\xa9", "\xe2\x80", "\xed\xa0\x80", "\x80", "\xff"}
	tails := []string{strings.Repeat("a", 70), strings.Repeat("あ", 24), `"` + strings.Repeat("b", 69), "é" + strings.Repeat("c", 68), ""}
	for n := 0; n <= 100; n++ {
		for _, piece := range pieces {
			for _, tail := range tails {
				s := prefix(n) + piece + tail
				for index := range stringEscapes {
					e := &stringEscapes[index]
					hasEscapeLoop = false
					want, wantValid := e.appendEscaped(e, []byte("prefix"), s)
					hasEscapeLoop = true
					got, gotValid := e.appendEscaped(e, []byte("prefix"), s)
					if string(got) != string(want) || gotValid != wantValid {
						t.Fatalf("%d %q:\n got %q\nwant %q", index, s, got, want)
					}
				}
			}
		}
	}
}

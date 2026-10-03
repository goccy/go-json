package jsonstring

import (
	"bytes"
	"math/rand"
	"testing"
	"unicode/utf8"
)

// The tables must tell the same as the table of every byte, for every combination of the options.
func TestNibbleTables(t *testing.T) {
	for index := range stringEscapes {
		e := &stringEscapes[index]
		for b := 0; b < 256; b++ {
			got := e.tables.Lo[b&0xf]&e.tables.Hi[b>>4] != 0
			if got != e.table[b] {
				t.Fatalf("escape %d, byte %#x: expected %v but got %v", index, b, e.table[b], got)
			}
		}
	}
}

// Every byte which the tables of the options without the normalization of UTF-8 tell to escape, and every byte of
// ASCII which the others tell to, has an escape sequence: the functions of the escapes write the sequence of such
// a byte without looking at it otherwise.
func TestEscapeSequences(t *testing.T) {
	for index := range stringEscapes {
		e := &stringEscapes[index]
		for b := 0; b < 256; b++ {
			if !e.table[b] || (b >= 0x80 && e.high != 0) {
				continue
			}
			if escapeSequences[b] == 0 {
				t.Fatalf("escape %d, byte %#x: no escape sequence", index, b)
			}
		}
	}
}

// utf8BlockEmulated does what the loop of the escapes by AVX2 does for a block of 32 bytes which has a byte which
// is not ASCII, with the tables, byte by byte ( see escapeUTF8AVX2 ): the block starts at the start of a
// character, so the bytes before it are taken as ASCII. problem is whether the block has invalid UTF-8 or U+2028
// or U+2029, and tail is the number of the bytes of the character which the last bytes start, if it goes on after
// the block.
func utf8BlockEmulated(b *[32]byte) (problem bool, tail int) {
	t := &utf8ValidationTables
	at := func(i int) byte {
		if i < 0 {
			return 0
		}
		return b[i]
	}
	for i := 0; i < 32; i++ {
		prev1, prev2, prev3 := at(i-1), at(i-2), at(i-3)
		special := t.Byte1High[prev1>>4] & t.Byte1Low[prev1&0xf] & t.Byte2High[b[i]>>4]
		var must23 byte
		if prev2 >= 0xe0 || prev3 >= 0xf0 {
			must23 = 0x80
		}
		if special^must23 != 0 {
			problem = true
		}
		if prev2 == 0xe2 && prev1 == 0x80 && b[i]&0xfe == 0xa8 {
			problem = true
		}
	}
	switch {
	case b[31] >= 0xc0:
		tail = 1
	case b[30] >= 0xe0:
		tail = 2
	case b[29] >= 0xf0:
		tail = 3
	}
	return problem, tail
}

// utf8ValidPrefix is whether p is the start of a valid character of UTF-8 which goes on after it.
func utf8ValidPrefix(p []byte) bool {
	if len(p) == 0 {
		return true
	}
	var n int
	switch c := p[0]; {
	case c >= 0xc2 && c <= 0xdf:
		n = 2
	case c >= 0xe0 && c <= 0xef:
		n = 3
	case c >= 0xf0 && c <= 0xf4:
		n = 4
	default:
		return false
	}
	if len(p) >= n {
		return false
	}
	for second := 0x80; second <= 0xbf; second++ {
		r := append([]byte{}, p...)
		if len(r) == 1 {
			r = append(r, byte(second))
		}
		for len(r) < n {
			r = append(r, 0x80)
		}
		if utf8.Valid(r) {
			return true
		}
	}
	return false
}

// checkUTF8Block checks what the loop of the escapes decides for a block against the decoder of UTF-8 of Go. A
// block without a problem must be valid UTF-8 without U+2028 and U+2029 up to its tail, which is not written but
// looked at again from its start by the next block. A block with a problem must have one: in the bytes up to the
// tail, or a tail which starts no valid character; else a valid string would be escaped by the loop of the bytes.
func checkUTF8Block(t *testing.T, b *[32]byte) {
	t.Helper()
	problem, tail := utf8BlockEmulated(b)
	head := b[:32-tail]
	headValid := utf8.Valid(head) && !bytes.Contains(head, []byte("\u2028")) && !bytes.Contains(head, []byte("\u2029"))
	if !problem && !headValid {
		t.Fatalf("% x: a problem is missed ( tail %d )", b[:], tail)
	}
	if problem && headValid && utf8ValidPrefix(b[32-tail:]) {
		t.Fatalf("% x: a problem is found in valid UTF-8 ( tail %d )", b[:], tail)
	}
}

func TestUTF8ValidationTables(t *testing.T) {
	ascii := func() *[32]byte {
		var b [32]byte
		for i := range b {
			b[i] = 'a'
		}
		return &b
	}
	// every two bytes at the start, across the halves of 16 bytes and at the end of a block.
	for _, at := range []int{0, 15, 30} {
		for x := 0; x < 256; x++ {
			for y := 0; y < 256; y++ {
				b := ascii()
				b[at], b[at+1] = byte(x), byte(y)
				checkUTF8Block(t, b)
			}
		}
	}
	// every three and four bytes of the bytes which decide the errors, anywhere near the halves and the end.
	interesting := []byte{0x00, 0x41, 0x7f, 0x80, 0x8f, 0x90, 0x9f, 0xa0, 0xa8, 0xa9, 0xbf, 0xc0, 0xc1, 0xc2, 0xdf,
		0xe0, 0xe2, 0xed, 0xee, 0xef, 0xf0, 0xf1, 0xf4, 0xf5, 0xf8, 0xff}
	for _, at := range []int{0, 13, 14, 15, 16, 28, 29} {
		for _, x := range interesting {
			for _, y := range interesting {
				for _, z := range interesting {
					b := ascii()
					b[at], b[at+1], b[at+2] = x, y, z
					checkUTF8Block(t, b)
					if at+3 < 32 {
						for _, w := range interesting {
							b[at+3] = w
							checkUTF8Block(t, b)
						}
					}
				}
			}
		}
	}
	// blocks of characters of every length, cut anywhere, with invalid bytes among them.
	r := rand.New(rand.NewSource(1))
	pieces := []string{"a", "\"", "é", "ß", "あ", "語", "‧", " ", " ", "‪", "😀", "\U0010ffff",
		"�", "\xed\x9f\xbf", "\xee\x80\x80", "\xed\xa0\x80", "\xc0\x80", "\xe0\x80\x80", "\xf0\x80\x80\x80",
		"\xf4\x90\x80\x80", "\xf5\x80\x80\x80", "\xff", "\x80", "\xe3\x81"}
	for n := 0; n < 1000000; n++ {
		var s []byte
		for len(s) < 32 {
			if r.Intn(8) == 0 {
				s = append(s, pieces[15+r.Intn(len(pieces)-15)]...)
			} else {
				s = append(s, pieces[r.Intn(15)]...)
			}
		}
		var b [32]byte
		copy(b[:], s)
		checkUTF8Block(t, &b)
	}
}

// The loop of the escapes by SIMD escapes the bytes of ASCII of the options which normalize UTF-8 by the tables of
// the options which don't ( see appendEscapedSIMD ): they must escape the same bytes of ASCII.
func TestEscapeTablesOfASCII(t *testing.T) {
	for _, index := range []int{0, stringEscapeHTML} {
		for b := 0; b < 0x80; b++ {
			if stringEscapes[index].table[b] != stringEscapes[index|stringEscapeNormalize].table[b] {
				t.Fatalf("escape %d, byte %#x: the tables differ", index, b)
			}
		}
	}
}

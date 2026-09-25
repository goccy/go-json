package encoder

import "testing"

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
			if !e.table[b] || (b >= 0x80 && index&stringEscapeNormalize != 0) {
				continue
			}
			if escapeSequences[b] == 0 {
				t.Fatalf("escape %d, byte %#x: no escape sequence", index, b)
			}
		}
	}
}

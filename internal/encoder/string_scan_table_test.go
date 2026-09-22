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

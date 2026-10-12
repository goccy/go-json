package decoder

import (
	"math/rand"
	"strings"
	"testing"
	"unsafe"
)

// referenceMasks computes the masks of a block one byte at a time.
func referenceMasks(block []byte) scanMasks {
	var m scanMasks
	for i, c := range block {
		bit := uint64(1) << uint(i)
		switch c {
		case '"':
			m.quote |= bit
		case '\\':
			m.backslash |= bit
		case '{', '[':
			m.open |= bit
		case '}', ']':
			m.closing |= bit
		}
	}
	return m
}

func randomBlock(r *rand.Rand) []byte {
	const chars = `{}[]"\abc01 ,:{}[]"\\\`
	block := make([]byte, scanBlockSize)
	for i := range block {
		if r.Intn(8) == 0 {
			block[i] = byte(r.Intn(256))
		} else {
			block[i] = chars[r.Intn(len(chars))]
		}
	}
	return block
}

func TestScanBlock(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 10000; i++ {
		block := randomBlock(r)
		want := referenceMasks(block)
		var got scanMasks
		scanBlock(unsafe.Pointer(&block[0]), &got)
		if got != want {
			t.Fatalf("scanBlock(%q) = %+v, want %+v", block, got, want)
		}
		scanBlockWords(unsafe.Pointer(&block[0]), &got)
		if got != want {
			t.Fatalf("scanBlockWords(%q) = %+v, want %+v", block, got, want)
		}
	}
}

// referenceSkipCompound finds the end of the object or the array open at cursor one byte at a time.
// A byte after a backslash is escaped wherever it is: it is neither a quote nor a bracket.
func referenceSkipCompound(buf []byte, cursor, depth int64) (int64, bool) {
	inString, escaped := false, false
	for pos := cursor; pos < int64(len(buf)); pos++ {
		c := buf[pos]
		switch {
		case escaped:
			escaped = false
			continue
		case c == '\\':
			escaped = true
			continue
		case inString:
			if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 {
				return pos + 1, true
			}
		}
	}
	return int64(len(buf)), false
}

// randomValue returns a random object or array, well formed or not, with strings, escapes and
// random bytes, of a length which crosses the blocks in every way.
func randomValue(r *rand.Rand) []byte {
	var b strings.Builder
	n := r.Intn(300)
	pieces := []string{`{`, `}`, `[`, `]`, `"`, `\`, `\"`, `\\`, `\\\`, `"a"`, `"\"b\""`, `:`, `,`, `1`, ` `, "\n", `"\\"`, `"\\\"`}
	b.WriteByte("{["[r.Intn(2)])
	for i := 0; i < n; i++ {
		if r.Intn(10) == 0 {
			b.WriteByte(byte(r.Intn(256)))
		} else {
			b.WriteString(pieces[r.Intn(len(pieces))])
		}
	}
	return []byte(b.String())
}

func TestScanCompoundAgainstReference(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for i := 0; i < 20000; i++ {
		value := randomValue(r)
		buf := NewInput(value)
		wantEnd, wantFound := referenceSkipCompound(buf, 1, 1)
		sc := compoundScanner{depth: 1, maxDepth: maxDecodeNestingDepth}
		gotEnd, gotFound, err := sc.scan(buf, 1, int64(len(buf)))
		if err != nil {
			t.Fatalf("%q: %v", value, err)
		}
		if gotFound != wantFound || (gotFound && gotEnd != wantEnd) {
			t.Fatalf("%q: got (%d, %v), want (%d, %v)", value, gotEnd, gotFound, wantEnd, wantFound)
		}
		// The same value scanned in pieces, as a stream does, ends at the same place.
		sc = compoundScanner{depth: 1, maxDepth: maxDecodeNestingDepth}
		pos := int64(1)
		for pos < int64(len(buf)) {
			lim := min(pos+int64(r.Intn(70)), int64(len(buf)))
			end, found, err := sc.scan(buf, pos, lim)
			if err != nil {
				t.Fatalf("%q: %v", value, err)
			}
			if found {
				if !wantFound || end != wantEnd {
					t.Fatalf("%q in pieces: got %d, want (%d, %v)", value, end, wantEnd, wantFound)
				}
				break
			}
			if wantFound && wantEnd <= lim {
				t.Fatalf("%q in pieces: not found before %d, want %d", value, lim, wantEnd)
			}
			pos = end
		}
	}
}

func TestScanCompoundMaxDepth(t *testing.T) {
	for _, n := range []int{9999, 10000, 10001} {
		buf := NewInput([]byte(strings.Repeat("[", n) + strings.Repeat("]", n)))
		_, err := skipValue(buf, 0, 0)
		if (err != nil) != (n > maxDecodeNestingDepth) {
			t.Fatalf("%d brackets: err = %v", n, err)
		}
		_, err = skipValue(buf, 0, 1)
		if (err != nil) != (n+1 > maxDecodeNestingDepth) {
			t.Fatalf("%d brackets nested once: err = %v", n, err)
		}
	}
}

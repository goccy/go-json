package decoder

import (
	"math/bits"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
	// bit5 is the bit which distinguishes '{' from '[' and '}' from ']'.
	bit5 = 0x2020202020202020
)

// load64 returns the eight bytes at pos as a little-endian word.
func load64(buf []byte, pos int64) uint64 {
	return *(*uint64)(unsafe.Add((*sliceHeader)(unsafe.Pointer(&buf)).data, pos))
}

// byteMask returns the word whose byte is 0x80 where the byte of w is c, and 0 elsewhere.
func byteMask(w, c uint64) uint64 {
	x := w ^ (c * lsb)
	return ^(((x & ^uint64(msb)) + ^uint64(msb)) | x) & msb
}

// compoundScanner finds the end of an object or an array without decoding it: the position after
// the bracket which closes it. It keeps its state between the calls of scan, so that a stream may
// scan what it has read so far and go on after reading more.
//
// The bytes are scanned eight at a time: the quotes toggle whether the bytes are inside a string,
// and the brackets outside strings count the depth. A word with a backslash, and the bytes after
// the last whole word, are scanned one by one.
type compoundScanner struct {
	// depth is the number of the brackets opened but not closed.
	depth int64
	// maxDepth is the depth beyond which the nesting is too deep.
	maxDepth int64
	inString bool
	// escaped is whether the next byte is escaped by the backslash before it.
	escaped bool
}

// scan scans buf[pos:lim] and returns the position after the bracket which closes the value, or,
// when the value doesn't end before lim, the position it stopped at and false.
func (sc *compoundScanner) scan(buf []byte, pos, lim int64) (int64, bool, error) {
	depth, inString, escaped := sc.depth, sc.inString, sc.escaped
	for pos < lim {
		for pos+8 <= lim {
			w := load64(buf, pos)
			if byteMask(w, '\\') != 0 || escaped {
				// scanned one by one below
				break
			}
			// inside has 0x80 in the bytes which are inside a string, after the byte itself is read:
			// the parity of the quotes up to the byte, flipped if the word begins inside a string.
			inside := byteMask(w, '"')
			inside ^= inside << 8
			inside ^= inside << 16
			inside ^= inside << 32
			if inString {
				inside ^= msb
			}
			inString = inside&(1<<63) != 0
			folded := w &^ bit5
			open := byteMask(folded, '[') &^ inside
			closing := byteMask(folded, ']') &^ inside
			closes := int64(bits.OnesCount64(closing))
			opens := int64(bits.OnesCount64(open))
			if closes < depth && depth+opens <= sc.maxDepth {
				// the value can't end in this word, nor go too deep
				depth += opens - closes
				pos += 8
				continue
			}
			for brackets := open | closing; brackets != 0; brackets &= brackets - 1 {
				i := bits.TrailingZeros64(brackets) / 8
				if open&(1<<(i*8+7)) != 0 {
					depth++
					if depth > sc.maxDepth {
						return 0, false, errors.ErrExceededMaxDepth(buf[pos], pos)
					}
				} else {
					depth--
					if depth == 0 {
						return pos + int64(i) + 1, true, nil
					}
				}
			}
			pos += 8
		}
		// The last bytes, or a word with a backslash: one by one.
		chunk := min(pos+8, lim)
		for pos < chunk {
			c := buf[pos]
			pos++
			if inString {
				switch {
				case escaped:
					escaped = false
				case c == '\\':
					escaped = true
				case c == '"':
					inString = false
				}
				continue
			}
			switch c {
			case '"':
				inString = true
			case '{', '[':
				depth++
				if depth > sc.maxDepth {
					return 0, false, errors.ErrExceededMaxDepth(c, pos-1)
				}
			case '}', ']':
				depth--
				if depth == 0 {
					return pos, true, nil
				}
			}
		}
	}
	sc.depth, sc.inString, sc.escaped = depth, inString, escaped
	return pos, false, nil
}

// skipCompound returns the position after the end of the object or the array which is open at cursor:
// depth brackets are open, nesting is how deep the value is nested in the input.
// The buffer ends with a nul byte, which no value may reach.
func skipCompound(buf []byte, cursor, depth, nesting int64) (int64, error) {
	sc := compoundScanner{depth: depth, maxDepth: maxDecodeNestingDepth - nesting}
	end, found, err := sc.scan(buf, cursor, int64(len(buf)))
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, errors.ErrUnexpectedEndOfJSON("object or array", end)
	}
	return end, nil
}

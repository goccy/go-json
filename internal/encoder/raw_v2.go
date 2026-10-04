package encoder

import (
	"bytes"
	"encoding/binary"
	"math/bits"
	"unicode/utf8"
	"unsafe"

	"github.com/goccy/go-json/internal/jsonstring"
)

// The raw value which a method or a function of the v2 semantics returns is formatted as the options want it,
// which checks it as well: valid JSON, valid UTF-8 and no duplicate names. Most of the values are compact
// already, which are checked here by a walk, as the outputs of the marshalers of v1 are ( see
// appendCompactOutput ), and appended as they are, but for the strings with escapes, which are written again in
// the form of the formatting: the escapes which it doesn't write are written as the characters they are.

// lookedAtByRawCheck are the characters which the escaping of the options escapes, in a raw value which is appended
// by the formatting otherwise ( see AppendFormattedRaw ): U+2028 and U+2029 ( EscapeForJS ), whose first byte is
// 0xe2, and '<', '>' and '&' ( EscapeForHTML ). The control characters, which are in each class, are not
// appended here anyway.
var lookedAtByRawCheck = [4]*jsonstring.ByteClass{
	nil,
	jsonstring.NewByteClass(0xe2),
	jsonstring.NewByteClass('<', '>', '&'),
	jsonstring.NewByteClass('<', '>', '&', 0xe2),
}

// maxNamesOfRawCheck is the number of the names of the objects being read, and maxDepthOfRawCheck the depth of the
// values, up to which a raw value is appended here: a larger one is formatted. They are small, as the arrays of the
// walk are cleared for every value.
const (
	maxNamesOfRawCheck = 32
	maxDepthOfRawCheck = 16
)

// AppendFormattedRaw appends src, a raw value, to dst as the formatting of the v2 semantics with the escaping of
// ctx appends it, and reports true, if src is compact, valid JSON whose strings are valid UTF-8 without a control
// character or an escape of a surrogate, whose names have no escape, and whose objects have no duplicate names.
// Otherwise it reports false, and src is to be formatted: dst may have been written after its length then.
//
// It reads src once, and once more for the escaping of the options, if any, and for UTF-8, if src has a byte
// which is not ASCII.
func AppendFormattedRaw(ctx *RuntimeContext, dst, src []byte) ([]byte, bool) {
	flags := ctx.Option.Flag
	escape := 0
	if flags&NormalizeUTF8Option != 0 {
		escape |= 1
	}
	if flags&HTMLEscapeOption != 0 {
		escape |= 2
	}
	if len(src) == 0 || (escape != 0 && lookedAtByRawCheck[escape].Has(src)) {
		return dst, false
	}
	w := rawWalk{ctx: ctx, src: src, dst: dst, ascii: true, escape: escape}
	if !w.walk() || !(w.ascii || utf8.Valid(src)) {
		return dst, false
	}
	return append(w.dst, src[w.copied:]...), true
}

// rawWalk is the state of the walk of AppendFormattedRaw over src: dst is the output, which has src up to copied,
// with the strings of escapes before it written again.
type rawWalk struct {
	ctx    *RuntimeContext
	src    []byte
	dst    []byte
	copied int
	// ascii is whether every byte read is ASCII.
	ascii bool
	// escape is the escaping of the options ( see lookedAtByRawCheck ).
	escape int
	// stack are the closing brackets of the values being read, of depth levels; names are the offsets of the names
	// of the objects being read, numNames of them, and first the index in names of the first name of each level.
	stack    [maxDepthOfRawCheck]byte
	first    [maxDepthOfRawCheck]int32
	names    [maxNamesOfRawCheck]int32
	numNames int32
	depth    int
}

// name reads the name of an object at i, which must be new in the object, and returns the index after it, or -1.
func (w *rawWalk) name(i int) int {
	src := w.src
	n := len(src)
	if i >= n || src[i] != '"' || w.numNames == maxNamesOfRawCheck {
		return -1
	}
	end, escaped := w.stringEnd(i + 1)
	if end < 0 || escaped {
		return -1
	}
	size := end - i
	for _, start := range w.names[w.first[w.depth-1]:w.numNames] {
		// the name with its quotes, which has no escape: the quote ends the one before as well.
		if s := int(start); s+size < end && src[s+size-1] == '"' && bytes.Equal(src[s:s+size], src[i:end]) {
			return -1
		}
	}
	w.names[w.numNames] = int32(i)
	w.numNames++
	if end >= n || src[end] != ':' {
		return -1
	}
	return end + 1
}

// walk reports whether src is a value which AppendFormattedRaw appends.
func (w *rawWalk) walk() bool {
	src := w.src
	n := len(src)
	i := 0
	for {
		// a value.
		if i >= n {
			return false
		}
		switch c := src[i]; {
		case c == '"':
			end, escaped := w.stringEnd(i + 1)
			if end < 0 {
				return false
			}
			if escaped && !w.requote(i, end) {
				return false
			}
			i = end
		case c == '{':
			if w.depth == maxDepthOfRawCheck {
				return false
			}
			w.stack[w.depth] = '}'
			w.first[w.depth] = w.numNames
			w.depth++
			i++
			if i < n && src[i] == '}' {
				i++
				w.depth--
				break
			}
			if i = w.name(i); i < 0 {
				return false
			}
			continue
		case c == '[':
			if w.depth == maxDepthOfRawCheck {
				return false
			}
			w.stack[w.depth] = ']'
			w.first[w.depth] = w.numNames
			w.depth++
			i++
			if i < n && src[i] == ']' {
				i++
				w.depth--
				break
			}
			continue
		case c == 't':
			if !bytes.HasPrefix(src[i:], []byte("true")) {
				return false
			}
			i += 4
		case c == 'f':
			if !bytes.HasPrefix(src[i:], []byte("false")) {
				return false
			}
			i += 5
		case c == 'n':
			if !bytes.HasPrefix(src[i:], []byte("null")) {
				return false
			}
			i += 4
		case c == '-' || ('0' <= c && c <= '9'):
			if i = numberEnd(src, i); i < 0 {
				return false
			}
		default:
			return false
		}
		// after a value.
		for {
			if w.depth == 0 {
				return i == n
			}
			if i >= n {
				return false
			}
			switch src[i] {
			case ',':
				i++
				if w.stack[w.depth-1] == '}' {
					if i = w.name(i); i < 0 {
						return false
					}
				}
			case w.stack[w.depth-1]:
				w.depth--
				w.numNames = w.first[w.depth]
				i++
				continue
			default:
				return false
			}
			break
		}
	}
}

// stringEnd returns the index after the quote which ends the string whose content starts at i, and whether the
// string has an escape; or -1 if the string has a control character or an invalid escape, or an escape of a
// surrogate, which the formatting reports or writes as it is, or doesn't end. It clears ascii if the string has a
// byte which is not ASCII. The bytes are looked at by words: the lowest byte of the mask of each kind is exact.
func (w *rawWalk) stringEnd(i int) (int, bool) {
	src := w.src
	n := len(src)
	p := unsafe.Pointer(unsafe.SliceData(src))
	escaped := false
	for {
		for ; i+8 <= n; i += 8 {
			x := binary.LittleEndian.Uint64((*[8]byte)(unsafe.Add(p, i))[:])
			q := x ^ (lsb * '"')
			b := x ^ (lsb * '\\')
			mask := ((q-lsb)&^q | (b-lsb)&^b | (x-lsb*0x20)&^x) & msb
			if w.ascii {
				mask |= x & msb
			}
			if mask != 0 {
				i += bits.TrailingZeros64(mask) / 8
				break
			}
		}
		for ; i < n; i++ {
			if c := src[i]; c == '"' || c == '\\' || c < 0x20 || (c >= 0x80 && w.ascii) {
				break
			}
		}
		if i >= n {
			return -1, false
		}
		switch c := src[i]; {
		case c == '"':
			return i + 1, escaped
		case c == '\\':
			if i+1 < n && src[i+1] == 'u' && i+3 < n && (src[i+2]|0x20) == 'd' && src[i+3] >= '8' {
				// \uD800 to \uDFFF: a surrogate, which the formatting checks with the one after it.
				return -1, false
			}
			if i = escapeEnd(src, i+1); i < 0 {
				return -1, false
			}
			escaped = true
		case c >= 0x80:
			w.ascii = false
			i++
		default:
			return -1, false
		}
	}
}

// requote writes the string src[start:end], which has escapes, as the formatting writes it: an escaped character
// which needs no escape as it is, and the others by the escapes of the escapers of the strings, in one pass. With
// the escaping of the options, the string is written by the escaper of the options.
func (w *rawWalk) requote(start, end int) bool {
	w.dst = append(w.dst, w.src[w.copied:start]...)
	w.copied = end
	if w.escape != 0 {
		ctx := w.ctx
		ctx.MarshalBuf = jsonstring.AppendUnescaped(ctx.MarshalBuf[:0], w.src[start+1:end-1])
		var valid bool
		w.dst, valid = jsonstring.AppendQuoted(textEscaper(ctx), w.dst, unsafe.String(unsafe.SliceData(ctx.MarshalBuf), len(ctx.MarshalBuf)))
		return valid
	}
	dst := append(w.dst, '"')
	// the escapes are valid, of no surrogate ( see stringEnd ), and the rest is checked for UTF-8 with the value.
	s := w.src[start+1 : end-1]
	for {
		j := bytes.IndexByte(s, '\\')
		if j < 0 {
			dst = append(dst, s...)
			break
		}
		dst = append(dst, s[:j]...)
		r, size := rune(s[j+1]), 2
		switch r {
		case 'b':
			r = '\b'
		case 'f':
			r = '\f'
		case 'n':
			r = '\n'
		case 'r':
			r = '\r'
		case 't':
			r = '\t'
		case 'u':
			r, size = hexRune(s[j+2:j+6]), 6
		}
		switch {
		case r == '"' || r == '\\' || r < 0x20:
			dst = jsonstring.AppendASCIIEscape(dst, byte(r))
		case r < utf8.RuneSelf:
			dst = append(dst, byte(r))
		default:
			dst = utf8.AppendRune(dst, r)
		}
		s = s[j+size:]
	}
	w.dst = append(dst, '"')
	return true
}

// hexRune returns the character of the four hexadecimal digits of an escape, which are valid.
func hexRune(h []byte) rune {
	var r rune
	for _, c := range h[:4] {
		switch {
		case c <= '9':
			c -= '0'
		case c <= 'F':
			c -= 'A' - 10
		default:
			c -= 'a' - 10
		}
		r = r<<4 | rune(c)
	}
	return r
}

// This files's string processing codes are inspired by https://github.com/segmentio/encoding.
// The license notation is as follows.
//
// # MIT License
//
// Copyright (c) 2019 Segment.io, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package encoder

import (
	"encoding/binary"
	"math/bits"
	"unsafe"
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
)

var hex = "0123456789abcdef"

// stringEscape is what decides whether a string has a byte to escape, for a combination of the options.
type stringEscape struct {
	// chars are three characters to escape in addition to the control characters, '"' and '\\',
	// repeated in every byte of a word. They are '<', '>' and '&', or '"' for nothing more.
	chars [3]uint64
	// high is msb if a byte which is not ASCII is to be looked at, to normalize UTF-8, or 0.
	high uint64
	// table is whether a byte may need an escape.
	table *[256]bool
	// tables is table as the tables of the scan by SIMD.
	tables nibbleTables
	// appendEscaped appends a string which may have a byte to escape.
	appendEscaped func(buf []byte, s string) []byte
}

const (
	stringEscapeNormalize = 1 << iota
	stringEscapeHTML
)

// stringEscapes is indexed by the combination of stringEscapeNormalize and stringEscapeHTML.
var stringEscapes = [4]stringEscape{
	0: {
		chars:         [3]uint64{lsb * '"', lsb * '"', lsb * '"'},
		table:         &needEscape,
		appendEscaped: appendString,
	},
	stringEscapeNormalize: {
		chars:         [3]uint64{lsb * '"', lsb * '"', lsb * '"'},
		high:          msb,
		table:         &needEscapeNormalizeUTF8,
		appendEscaped: appendNormalizedString,
	},
	stringEscapeHTML: {
		chars:         [3]uint64{lsb * '<', lsb * '>', lsb * '&'},
		table:         &needEscapeHTML,
		appendEscaped: appendHTMLString,
	},
	stringEscapeHTML | stringEscapeNormalize: {
		chars:         [3]uint64{lsb * '<', lsb * '>', lsb * '&'},
		high:          msb,
		table:         &needEscapeHTMLNormalizeUTF8,
		appendEscaped: appendNormalizedHTMLString,
	},
}

// escapeTables are the tables of stringEscapes, which the functions of the escapes refer to by this array:
// stringEscapes refers to the functions, so they can't refer to it.
var escapeTables [4]*nibbleTables

func init() {
	for i := range stringEscapes {
		stringEscapes[i].tables = newNibbleTables(stringEscapes[i].table)
		escapeTables[i] = &stringEscapes[i].tables
	}
}

// The functions below return the mask of the bytes of a word which may need an escape: the most significant bit
// of such a byte is set. A byte after one to escape may be set too, because of the borrow of the subtractions.
// A mask is made of parts, each of which is small enough to be inlined: a call for every word costs more than
// the mask itself.
//
// `x - lsb` sets the most significant bit of a byte of x which is zero, and of one which already has it.
// If UTF-8 is normalized, a byte which is not ASCII is looked at anyway, so the latter doesn't matter and the
// loose masks are enough. Otherwise `&^ x` leaves only the former, so that a string which is not ASCII is not
// taken for one to escape.

// looseCommonMask is the mask of the control characters, '"', '\\' and the bytes which are not ASCII.
func looseCommonMask(n uint64) uint64 {
	return n | (n - lsb*0x20) | ((n ^ (lsb * '"')) - lsb) | ((n ^ (lsb * '\\')) - lsb)
}

// looseCharsMask is the mask of chars, and of the bytes which are not ASCII.
func (e *stringEscape) looseCharsMask(n uint64) uint64 {
	return ((n ^ e.chars[0]) - lsb) | ((n ^ e.chars[1]) - lsb) | ((n ^ e.chars[2]) - lsb)
}

// exactCommonMask is the mask of the control characters, '"' and '\\'.
func exactCommonMask(n uint64) uint64 {
	quote := n ^ (lsb * '"')
	backslash := n ^ (lsb * '\\')
	return ((n - lsb*0x20) &^ n) | ((quote - lsb) &^ quote) | ((backslash - lsb) &^ backslash)
}

// exactCharsMask is the mask of chars.
func (e *stringEscape) exactCharsMask(n uint64) uint64 {
	c0, c1, c2 := n^e.chars[0], n^e.chars[1], n^e.chars[2]
	return ((c0 - lsb) &^ c0) | ((c1 - lsb) &^ c1) | ((c2 - lsb) &^ c2)
}

// hasExactEscape is whether a word of the string has a byte to escape, not counting the bytes which are not
// ASCII. The string is 8 bytes or longer.
func (e *stringEscape) hasExactEscape(src unsafe.Pointer, n int) bool {
	i := 0
	for ; i+8 <= n; i += 8 {
		if w := *(*uint64)(unsafe.Add(src, i)); (exactCommonMask(w)|e.exactCharsMask(w))&msb != 0 {
			return true
		}
	}
	if i < n {
		w := *(*uint64)(unsafe.Add(src, n-8))
		return (exactCommonMask(w)|e.exactCharsMask(w))&msb != 0
	}
	return false
}

// hasLooseEscape is whether a word of the string may have a byte to escape, or has a byte which is not ASCII.
// The string is 8 bytes or longer. It is cheaper than hasExactEscape, and it is exact for a string of ASCII,
// which most of the strings are.
func (e *stringEscape) hasLooseEscape(src unsafe.Pointer, n int) bool {
	i := 0
	for ; i+8 <= n; i += 8 {
		if w := *(*uint64)(unsafe.Add(src, i)); (looseCommonMask(w)|e.looseCharsMask(w))&msb != 0 {
			return true
		}
	}
	if i < n {
		// the last word overlaps the previous one.
		w := *(*uint64)(unsafe.Add(src, n-8))
		return (looseCommonMask(w)|e.looseCharsMask(w))&msb != 0
	}
	return false
}

// maxInlineCopySize is the size of the longest string which is copied without calling memmove.
const maxInlineCopySize = 16

// growForString returns buf which has the capacity to append n bytes, growing as append does.
//
//go:noinline
func growForString(buf []byte, n int) []byte {
	grown := make([]byte, len(buf), 2*cap(buf)+n)
	copy(grown, buf)
	return grown
}

// AppendString appends the string as a JSON string.
//
// Most of the strings have nothing to escape and are short, so that case is handled here: the capacity is
// checked once, the string is looked at by words ( by bytes if it is shorter than a word ), and a short string
// is copied by a few loads and stores instead of a call of memmove, which costs more than the copy itself.
// A string which may have a byte to escape is left to the function for the options.
func AppendString(ctx *RuntimeContext, buf []byte, s string) []byte {
	index := 0
	if ctx.Option.Flag&HTMLEscapeOption != 0 {
		index = stringEscapeHTML
	}
	if ctx.Option.Flag&NormalizeUTF8Option != 0 {
		index |= stringEscapeNormalize
	}
	escape := &stringEscapes[index]

	n := len(s)
	if n == 0 {
		return append(buf, '"', '"')
	}
	src := unsafe.Pointer(unsafe.StringData(s))
	if n < 4 {
		table := escape.table
		for i := 0; i < n; i++ {
			if table[*(*byte)(unsafe.Add(src, i))] {
				return escape.appendEscaped(buf, s)
			}
		}
	} else if n < 8 {
		// the two halves overlap, so that the word has every byte of the string and nothing else.
		w := uint64(*(*uint32)(src)) | uint64(*(*uint32)(unsafe.Add(src, n-4)))<<32
		if (looseCommonMask(w)|escape.looseCharsMask(w))&msb != 0 &&
			(escape.high != 0 || (exactCommonMask(w)|escape.exactCharsMask(w))&msb != 0) {
			return escape.appendEscaped(buf, s)
		}
	} else if found, ok := escape.hasEscapeSIMD(src, n); ok {
		if found {
			return escape.appendEscaped(buf, s)
		}
	} else if escape.hasLooseEscape(src, n) && (escape.high != 0 || escape.hasExactEscape(src, n)) {
		// a byte which is not ASCII is to be looked at only if UTF-8 is normalized.
		return escape.appendEscaped(buf, s)
	}

	l := len(buf)
	if cap(buf)-l < n+2 {
		buf = growForString(buf, n+2)
	}
	buf = buf[:l+n+2]
	dst := unsafe.Pointer(unsafe.SliceData(buf[l:]))
	*(*byte)(dst) = '"'
	dst = unsafe.Add(dst, 1)
	switch {
	case n > maxInlineCopySize:
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

func appendNormalizedHTMLString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	var i int
	// the bytes up to the first one which may need an escape are skipped by words, or by SIMD: the first word
	// is looked at here, which is where the escape of a short string is.
	j := 0
	if valLen >= 8 {
		if m := maskNormalizedHTML(firstWord(s)); m != 0 {
			j = bits.TrailingZeros64(m) / 8
		} else {
			j = skipNormalizedHTML(s, 8)
		}
	}
	for j < valLen {
		c := s[j]

		if !needEscapeHTMLNormalizeUTF8[c] {
			// the bytes up to the next one which may need an escape are skipped by words, but a few ones,
			// which are not worth a call
			j++
			if valLen-j >= 8 {
				j = skipNormalizedHTML(s, j)
			}
			continue
		}

		switch c {
		case '\\', '"':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', c)
			i = j + 1
			j = j + 1
			continue

		case '\n':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'n')
			i = j + 1
			j = j + 1
			continue

		case '\r':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'r')
			i = j + 1
			j = j + 1
			continue

		case '\t':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 't')
			i = j + 1
			j = j + 1
			continue

		case '<', '>', '&':
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
			continue

		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C, 0x0E, 0x0F, // 0x00-0x0F
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F: // 0x10-0x1F
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
			continue
		}
		state, size := decodeRuneInString(s[j:])
		switch state {
		case runeErrorState:
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\ufffd`...)
			i = j + 1
			j = j + 1
			continue
			// U+2028 is LINE SEPARATOR.
			// U+2029 is PARAGRAPH SEPARATOR.
			// They are both technically valid characters in JSON strings,
			// but don't work in JSONP, which has to be evaluated as JavaScript,
			// and can lead to security holes there. It is valid JSON to
			// escape them, so we do so unconditionally.
			// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		case lineSepState:
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u2028`...)
			i = j + 3
			j = j + 3
			continue
		case paragraphSepState:
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u2029`...)
			i = j + 3
			j = j + 3
			continue
		}
		j += size
	}

	return append(append(buf, s[i:]...), '"')
}

func appendHTMLString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	var i int
	// the bytes up to the first one which may need an escape are skipped by words, or by SIMD: the first word
	// is looked at here, which is where the escape of a short string is.
	j := 0
	if valLen >= 8 {
		if m := maskHTML(firstWord(s)); m != 0 {
			j = bits.TrailingZeros64(m) / 8
		} else {
			j = skipHTML(s, 8)
		}
	}
	for j < valLen {
		c := s[j]

		if !needEscapeHTML[c] {
			// the bytes up to the next one which may need an escape are skipped by words, but a few ones,
			// which are not worth a call
			j++
			if valLen-j >= 8 {
				j = skipHTML(s, j)
			}
			continue
		}

		switch c {
		case '\\', '"':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', c)
			i = j + 1
			j = j + 1
			continue

		case '\n':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'n')
			i = j + 1
			j = j + 1
			continue

		case '\r':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'r')
			i = j + 1
			j = j + 1
			continue

		case '\t':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 't')
			i = j + 1
			j = j + 1
			continue

		case '<', '>', '&':
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
			continue

		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C, 0x0E, 0x0F, // 0x00-0x0F
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F: // 0x10-0x1F
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
			continue
		}
		j++
	}

	return append(append(buf, s[i:]...), '"')
}

func appendNormalizedString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	var i int
	// the bytes up to the first one which may need an escape are skipped by words, or by SIMD: the first word
	// is looked at here, which is where the escape of a short string is.
	j := 0
	if valLen >= 8 {
		if m := maskNormalized(firstWord(s)); m != 0 {
			j = bits.TrailingZeros64(m) / 8
		} else {
			j = skipNormalized(s, 8)
		}
	}
	for j < valLen {
		c := s[j]

		if !needEscapeNormalizeUTF8[c] {
			// the bytes up to the next one which may need an escape are skipped by words, but a few ones,
			// which are not worth a call
			j++
			if valLen-j >= 8 {
				j = skipNormalized(s, j)
			}
			continue
		}

		switch c {
		case '\\', '"':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', c)
			i = j + 1
			j = j + 1
			continue

		case '\n':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'n')
			i = j + 1
			j = j + 1
			continue

		case '\r':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'r')
			i = j + 1
			j = j + 1
			continue

		case '\t':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 't')
			i = j + 1
			j = j + 1
			continue

		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C, 0x0E, 0x0F, // 0x00-0x0F
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F: // 0x10-0x1F
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
			continue
		}

		state, size := decodeRuneInString(s[j:])
		switch state {
		case runeErrorState:
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\ufffd`...)
			i = j + 1
			j = j + 1
			continue
			// U+2028 is LINE SEPARATOR.
			// U+2029 is PARAGRAPH SEPARATOR.
			// They are both technically valid characters in JSON strings,
			// but don't work in JSONP, which has to be evaluated as JavaScript,
			// and can lead to security holes there. It is valid JSON to
			// escape them, so we do so unconditionally.
			// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		case lineSepState:
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u2028`...)
			i = j + 3
			j = j + 3
			continue
		case paragraphSepState:
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u2029`...)
			i = j + 3
			j = j + 3
			continue
		}
		j += size
	}

	return append(append(buf, s[i:]...), '"')
}

func appendString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	var i int
	// the bytes up to the first one which may need an escape are skipped by words, or by SIMD: the first word
	// is looked at here, which is where the escape of a short string is.
	j := 0
	if valLen >= 8 {
		if m := maskPlain(firstWord(s)); m != 0 {
			j = bits.TrailingZeros64(m) / 8
		} else {
			j = skipPlain(s, 8)
		}
	}
	for j < valLen {
		c := s[j]

		if !needEscape[c] {
			// the bytes up to the next one which may need an escape are skipped by words, but a few ones,
			// which are not worth a call
			j++
			if valLen-j >= 8 {
				j = skipPlain(s, j)
			}
			continue
		}

		switch c {
		case '\\', '"':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', c)
			i = j + 1
			j = j + 1
			continue

		case '\n':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'n')
			i = j + 1
			j = j + 1
			continue

		case '\r':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'r')
			i = j + 1
			j = j + 1
			continue

		case '\t':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 't')
			i = j + 1
			j = j + 1
			continue

		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C, 0x0E, 0x0F, // 0x00-0x0F
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F: // 0x10-0x1F
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
			continue
		}
		j++
	}

	return append(append(buf, s[i:]...), '"')
}

// The functions below return the position of the first byte from j which may need an escape, or of a byte of the
// last seven bytes of the string, which are left to the byte by byte loop: the bytes before it are skipped by
// SIMD where the CPU has it and the rest is long enough ( see indexEscapeSIMD ), and else a word at a time. After the first escape of a string, most of its bytes are still plain: a text or a source code has an
// escape every few dozen bytes, which a loop over its bytes one by one spends most of the time of the encoding in.
// The masks are the ones of the options ( see stringEscapes ): the loose ones, which have the bytes which are not
// ASCII, if UTF-8 is normalized, and else the exact ones, so that a text which is not ASCII is skipped too.

func skipNormalizedHTML(s string, j int) int {
	if k, ok := indexEscapeSIMD(unsafe.Pointer(unsafe.StringData(s[j:])), len(s)-j, escapeTables[stringEscapeHTML|stringEscapeNormalize]); ok {
		return j + k
	}
	for ; j+8 <= len(s); j += 8 {
		w := binary.LittleEndian.Uint64(unsafe.Slice(unsafe.StringData(s[j:]), 8))
		if m := maskNormalizedHTML(w); m != 0 {
			return j + bits.TrailingZeros64(m)/8
		}
	}
	return j
}

func skipHTML(s string, j int) int {
	if k, ok := indexEscapeSIMD(unsafe.Pointer(unsafe.StringData(s[j:])), len(s)-j, escapeTables[stringEscapeHTML]); ok {
		return j + k
	}
	for ; j+8 <= len(s); j += 8 {
		w := binary.LittleEndian.Uint64(unsafe.Slice(unsafe.StringData(s[j:]), 8))
		if m := maskHTML(w); m != 0 {
			return j + bits.TrailingZeros64(m)/8
		}
	}
	return j
}

func skipNormalized(s string, j int) int {
	if k, ok := indexEscapeSIMD(unsafe.Pointer(unsafe.StringData(s[j:])), len(s)-j, escapeTables[stringEscapeNormalize]); ok {
		return j + k
	}
	for ; j+8 <= len(s); j += 8 {
		w := binary.LittleEndian.Uint64(unsafe.Slice(unsafe.StringData(s[j:]), 8))
		if m := maskNormalized(w); m != 0 {
			return j + bits.TrailingZeros64(m)/8
		}
	}
	return j
}

func skipPlain(s string, j int) int {
	if k, ok := indexEscapeSIMD(unsafe.Pointer(unsafe.StringData(s[j:])), len(s)-j, escapeTables[0]); ok {
		return j + k
	}
	for ; j+8 <= len(s); j += 8 {
		w := binary.LittleEndian.Uint64(unsafe.Slice(unsafe.StringData(s[j:]), 8))
		if m := maskPlain(w); m != 0 {
			return j + bits.TrailingZeros64(m)/8
		}
	}
	return j
}

// The masks of a word of the options: the top bit of the first byte which may need an escape is set, and maybe
// the ones of the bytes after it ( see looseCommonMask ). They are small enough to be inlined.

func maskNormalizedHTML(w uint64) uint64 {
	return (looseCommonMask(w) | ((w ^ (lsb * '<')) - lsb) | ((w ^ (lsb * '>')) - lsb) | ((w ^ (lsb * '&')) - lsb)) & msb
}

func maskHTML(w uint64) uint64 {
	lt, gt, amp := w^(lsb*'<'), w^(lsb*'>'), w^(lsb*'&')
	return (exactCommonMask(w) | ((lt - lsb) &^ lt) | ((gt - lsb) &^ gt) | ((amp - lsb) &^ amp)) & msb
}

func maskNormalized(w uint64) uint64 {
	return looseCommonMask(w) & msb
}

func maskPlain(w uint64) uint64 {
	return exactCommonMask(w) & msb
}

// firstWord returns the first eight bytes of the string, which has eight bytes or more, as a little-endian word.
func firstWord(s string) uint64 {
	return binary.LittleEndian.Uint64(unsafe.Slice(unsafe.StringData(s), 8))
}

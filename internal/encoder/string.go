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
	"unsafe"
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
)

var hex = "0123456789abcdef"

func AppendString(ctx *RuntimeContext, buf []byte, s string) []byte {
	if ctx.Option.Flag&HTMLEscapeOption != 0 {
		if ctx.Option.Flag&NormalizeUTF8Option != 0 {
			return appendNormalizedHTMLString(buf, s)
		}
		return appendHTMLString(buf, s)
	}
	if ctx.Option.Flag&NormalizeUTF8Option != 0 {
		return appendNormalizedString(buf, s)
	}
	return appendString(buf, s)
}

// appendString encodes a JSON string without HTML escaping and without UTF-8 normalization.
// Uses resumed SIMD scanning: after each escape, jumps back to SIMD instead of byte-by-byte.
// Uses ASCIIOnly scanner that skips non-ASCII bytes (>= 0x80) since they're safe in this mode.
func appendString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	// Pre-grow buffer for the common case (no escaping)
	if cap(buf)-len(buf) < valLen+1 {
		newBuf := make([]byte, len(buf), 2*len(buf)+valLen+1)
		copy(newBuf, buf)
		buf = newBuf
	}

	p := stringptr(s)
	var i, j int

	for j < valLen {
		// Resumed SIMD scanning: use ASCIIOnly scanner for segments >= 16 bytes
		remaining := valLen - j
		if remaining >= 16 {
			offset := scanEscapeBasicASCIIOnly(unsafe.Pointer(uintptr(p)+uintptr(j)), remaining)
			j += offset
			if j >= valLen {
				break
			}
		} else {
			c := s[j]
			if !needEscape[c] {
				j++
				continue
			}
		}

		c := s[j]
		switch c {
		case '\\', '"':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', c)
			i = j + 1
			j = j + 1
		case '\n':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'n')
			i = j + 1
			j = j + 1
		case '\r':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'r')
			i = j + 1
			j = j + 1
		case '\t':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 't')
			i = j + 1
			j = j + 1
		default:
			if c < 0x20 {
				buf = append(buf, s[i:j]...)
				buf = append(buf, `\u00`...)
				buf = append(buf, hex[c>>4], hex[c&0xF])
				i = j + 1
				j = j + 1
			} else {
				// Not actually an escape char (can happen in tail path)
				j++
			}
		}
	}

	return append(append(buf, s[i:]...), '"')
}

// appendHTMLString encodes a JSON string with HTML escaping but without UTF-8 normalization.
// Uses ASCIIOnly scanner that skips non-ASCII bytes.
func appendHTMLString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	if cap(buf)-len(buf) < valLen+1 {
		newBuf := make([]byte, len(buf), 2*len(buf)+valLen+1)
		copy(newBuf, buf)
		buf = newBuf
	}

	p := stringptr(s)
	var i, j int

	for j < valLen {
		remaining := valLen - j
		if remaining >= 16 {
			offset := scanEscapeHTMLASCIIOnly(unsafe.Pointer(uintptr(p)+uintptr(j)), remaining)
			j += offset
			if j >= valLen {
				break
			}
		} else {
			c := s[j]
			if !needEscapeHTML[c] {
				j++
				continue
			}
		}

		c := s[j]
		switch c {
		case '\\', '"':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', c)
			i = j + 1
			j = j + 1
		case '\n':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'n')
			i = j + 1
			j = j + 1
		case '\r':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'r')
			i = j + 1
			j = j + 1
		case '\t':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 't')
			i = j + 1
			j = j + 1
		case '<', '>', '&':
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
		default:
			if c < 0x20 {
				buf = append(buf, s[i:j]...)
				buf = append(buf, `\u00`...)
				buf = append(buf, hex[c>>4], hex[c&0xF])
				i = j + 1
				j = j + 1
			} else {
				j++
			}
		}
	}

	return append(append(buf, s[i:]...), '"')
}

// appendNormalizedString encodes a JSON string with UTF-8 normalization but without HTML escaping.
// Uses the original scanner that flags non-ASCII bytes for UTF-8 validation.
func appendNormalizedString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	if cap(buf)-len(buf) < valLen+1 {
		newBuf := make([]byte, len(buf), 2*len(buf)+valLen+1)
		copy(newBuf, buf)
		buf = newBuf
	}

	p := stringptr(s)
	var i, j int

	for j < valLen {
		// For normalize paths, we must flag non-ASCII bytes, so use the original scanner
		remaining := valLen - j
		if remaining >= 16 {
			offset := scanEscapeBasic(unsafe.Pointer(uintptr(p)+uintptr(j)), remaining)
			j += offset
			if j >= valLen {
				break
			}
		} else {
			c := s[j]
			if !needEscapeNormalizeUTF8[c] {
				j++
				continue
			}
		}

		c := s[j]
		switch c {
		case '\\', '"':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', c)
			i = j + 1
			j = j + 1
		case '\n':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'n')
			i = j + 1
			j = j + 1
		case '\r':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'r')
			i = j + 1
			j = j + 1
		case '\t':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 't')
			i = j + 1
			j = j + 1
		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C, 0x0E, 0x0F, // 0x00-0x0F
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F: // 0x10-0x1F
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
		default:
			// Must be a non-ASCII byte (>= 0x80) - validate UTF-8
			state, size := decodeRuneInString(s[j:])
			switch state {
			case runeErrorState:
				buf = append(buf, s[i:j]...)
				buf = append(buf, `\ufffd`...)
				i = j + 1
				j = j + 1
			case lineSepState:
				buf = append(buf, s[i:j]...)
				buf = append(buf, `\u2028`...)
				i = j + 3
				j = j + 3
			case paragraphSepState:
				buf = append(buf, s[i:j]...)
				buf = append(buf, `\u2029`...)
				i = j + 3
				j = j + 3
			default:
				j += size
			}
		}
	}

	return append(append(buf, s[i:]...), '"')
}

// appendNormalizedHTMLString encodes a JSON string with both HTML escaping and UTF-8 normalization.
// Uses the original HTML scanner that flags non-ASCII bytes.
func appendNormalizedHTMLString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	if cap(buf)-len(buf) < valLen+1 {
		newBuf := make([]byte, len(buf), 2*len(buf)+valLen+1)
		copy(newBuf, buf)
		buf = newBuf
	}

	p := stringptr(s)
	var i, j int

	for j < valLen {
		remaining := valLen - j
		if remaining >= 16 {
			offset := scanEscapeHTML(unsafe.Pointer(uintptr(p)+uintptr(j)), remaining)
			j += offset
			if j >= valLen {
				break
			}
		} else {
			c := s[j]
			if !needEscapeHTMLNormalizeUTF8[c] {
				j++
				continue
			}
		}

		c := s[j]
		switch c {
		case '\\', '"':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', c)
			i = j + 1
			j = j + 1
		case '\n':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'n')
			i = j + 1
			j = j + 1
		case '\r':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 'r')
			i = j + 1
			j = j + 1
		case '\t':
			buf = append(buf, s[i:j]...)
			buf = append(buf, '\\', 't')
			i = j + 1
			j = j + 1
		case '<', '>', '&':
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C, 0x0E, 0x0F, // 0x00-0x0F
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F: // 0x10-0x1F
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
		default:
			// Must be a non-ASCII byte (>= 0x80) - validate UTF-8
			state, size := decodeRuneInString(s[j:])
			switch state {
			case runeErrorState:
				buf = append(buf, s[i:j]...)
				buf = append(buf, `\ufffd`...)
				i = j + 1
				j = j + 1
			case lineSepState:
				buf = append(buf, s[i:j]...)
				buf = append(buf, `\u2028`...)
				i = j + 3
				j = j + 3
			case paragraphSepState:
				buf = append(buf, s[i:j]...)
				buf = append(buf, `\u2029`...)
				i = j + 3
				j = j + 3
			default:
				j += size
			}
		}
	}

	return append(append(buf, s[i:]...), '"')
}

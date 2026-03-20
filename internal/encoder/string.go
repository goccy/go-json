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
func appendString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')

	// Fast path for short strings: use uint64 chunk scan + byte-by-byte tail
	if valLen < 16 {
		if valLen >= 8 {
			p := stringptr(s)
			n := *(*uint64)(p)
			mask := n | (n - (lsb * 0x20)) |
				((n ^ (lsb * '"')) - lsb) |
				((n ^ (lsb * '\\')) - lsb)
			if (mask & msb) == 0 {
				allSafe := true
				for k := 8; k < valLen; k++ {
					if needEscape[s[k]] {
						allSafe = false
						break
					}
				}
				if allSafe {
					return append(append(buf, s...), '"')
				}
			}
		} else {
			allSafe := true
			for k := 0; k < valLen; k++ {
				if needEscape[s[k]] {
					allSafe = false
					break
				}
			}
			if allSafe {
				return append(append(buf, s...), '"')
			}
		}
		// Short string with escapes: fall through to byte-by-byte loop
		var i, j int
		for j < valLen {
			c := s[j]
			if !needEscape[c] {
				j++
				continue
			}
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
					j++
				}
			}
		}
		return append(append(buf, s[i:]...), '"')
	}

	// SIMD path for longer strings (>= 16 bytes)
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
				j++
			}
		}
	}

	return append(append(buf, s[i:]...), '"')
}

// appendHTMLString encodes a JSON string with HTML escaping but without UTF-8 normalization.
func appendHTMLString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')

	// Fast path for short strings
	if valLen < 16 {
		if valLen >= 8 {
			p := stringptr(s)
			n := *(*uint64)(p)
			mask := n | (n - (lsb * 0x20)) |
				((n ^ (lsb * '"')) - lsb) |
				((n ^ (lsb * '\\')) - lsb) |
				((n ^ (lsb * '<')) - lsb) |
				((n ^ (lsb * '>')) - lsb) |
				((n ^ (lsb * '&')) - lsb)
			if (mask & msb) == 0 {
				allSafe := true
				for k := 8; k < valLen; k++ {
					if needEscapeHTML[s[k]] {
						allSafe = false
						break
					}
				}
				if allSafe {
					return append(append(buf, s...), '"')
				}
			}
		} else {
			allSafe := true
			for k := 0; k < valLen; k++ {
				if needEscapeHTML[s[k]] {
					allSafe = false
					break
				}
			}
			if allSafe {
				return append(append(buf, s...), '"')
			}
		}
		// Short string with escapes: fall through to byte-by-byte loop
		var i, j int
		for j < valLen {
			c := s[j]
			if !needEscapeHTML[c] {
				j++
				continue
			}
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

	// SIMD path for longer strings
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
func appendNormalizedString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')

	// Fast path for short strings
	if valLen < 16 {
		if valLen >= 8 {
			p := stringptr(s)
			n := *(*uint64)(p)
			mask := n | (n - (lsb * 0x20)) |
				((n ^ (lsb * '"')) - lsb) |
				((n ^ (lsb * '\\')) - lsb)
			if (mask & msb) == 0 {
				allSafe := true
				for k := 8; k < valLen; k++ {
					if needEscapeNormalizeUTF8[s[k]] {
						allSafe = false
						break
					}
				}
				if allSafe {
					return append(append(buf, s...), '"')
				}
			}
		} else {
			allSafe := true
			for k := 0; k < valLen; k++ {
				if needEscapeNormalizeUTF8[s[k]] {
					allSafe = false
					break
				}
			}
			if allSafe {
				return append(append(buf, s...), '"')
			}
		}
	}

	// Long string / short string with escapes
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
func appendNormalizedHTMLString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')

	// Fast path for short strings
	if valLen < 16 {
		if valLen >= 8 {
			p := stringptr(s)
			n := *(*uint64)(p)
			mask := n | (n - (lsb * 0x20)) |
				((n ^ (lsb * '"')) - lsb) |
				((n ^ (lsb * '\\')) - lsb) |
				((n ^ (lsb * '<')) - lsb) |
				((n ^ (lsb * '>')) - lsb) |
				((n ^ (lsb * '&')) - lsb)
			if (mask & msb) == 0 {
				allSafe := true
				for k := 8; k < valLen; k++ {
					if needEscapeHTMLNormalizeUTF8[s[k]] {
						allSafe = false
						break
					}
				}
				if allSafe {
					return append(append(buf, s...), '"')
				}
			}
		} else {
			allSafe := true
			for k := 0; k < valLen; k++ {
				if needEscapeHTMLNormalizeUTF8[s[k]] {
					allSafe = false
					break
				}
			}
			if allSafe {
				return append(append(buf, s...), '"')
			}
		}
	}

	// Long string / short string with escapes
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

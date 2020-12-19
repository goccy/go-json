package json

import (
	"math/bits"
	"reflect"
	"unicode/utf8"
	"unsafe"
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
)

var needEscapeWithHTML = [256]bool{
	'"':  true,
	'&':  true,
	'<':  true,
	'>':  true,
	'\\': true,
	0x00: true,
	0x01: true,
	0x02: true,
	0x03: true,
	0x04: true,
	0x05: true,
	0x06: true,
	0x07: true,
	0x08: true,
	0x09: true,
	0x0a: true,
	0x0b: true,
	0x0c: true,
	0x0d: true,
	0x0e: true,
	0x0f: true,
	0x10: true,
	0x11: true,
	0x12: true,
	0x13: true,
	0x14: true,
	0x15: true,
	0x16: true,
	0x17: true,
	0x18: true,
	0x19: true,
	0x1a: true,
	0x1b: true,
	0x1c: true,
	0x1d: true,
	0x1e: true,
	0x1f: true,
	/* 0x20 - 0x7f */
	0x80: true,
	0x81: true,
	0x82: true,
	0x83: true,
	0x84: true,
	0x85: true,
	0x86: true,
	0x87: true,
	0x88: true,
	0x89: true,
	0x8a: true,
	0x8b: true,
	0x8c: true,
	0x8d: true,
	0x8e: true,
	0x8f: true,
	0x90: true,
	0x91: true,
	0x92: true,
	0x93: true,
	0x94: true,
	0x95: true,
	0x96: true,
	0x97: true,
	0x98: true,
	0x99: true,
	0x9a: true,
	0x9b: true,
	0x9c: true,
	0x9d: true,
	0x9e: true,
	0x9f: true,
	0xa0: true,
	0xa1: true,
	0xa2: true,
	0xa3: true,
	0xa4: true,
	0xa5: true,
	0xa6: true,
	0xa7: true,
	0xa8: true,
	0xa9: true,
	0xaa: true,
	0xab: true,
	0xac: true,
	0xad: true,
	0xae: true,
	0xaf: true,
	0xb0: true,
	0xb1: true,
	0xb2: true,
	0xb3: true,
	0xb4: true,
	0xb5: true,
	0xb6: true,
	0xb7: true,
	0xb8: true,
	0xb9: true,
	0xba: true,
	0xbb: true,
	0xbc: true,
	0xbd: true,
	0xbe: true,
	0xbf: true,
	0xc0: true,
	0xc1: true,
	0xc2: true,
	0xc3: true,
	0xc4: true,
	0xc5: true,
	0xc6: true,
	0xc7: true,
	0xc8: true,
	0xc9: true,
	0xca: true,
	0xcb: true,
	0xcc: true,
	0xcd: true,
	0xce: true,
	0xcf: true,
	0xd0: true,
	0xd1: true,
	0xd2: true,
	0xd3: true,
	0xd4: true,
	0xd5: true,
	0xd6: true,
	0xd7: true,
	0xd8: true,
	0xd9: true,
	0xda: true,
	0xdb: true,
	0xdc: true,
	0xdd: true,
	0xde: true,
	0xdf: true,
	0xe0: true,
	0xe1: true,
	0xe2: true,
	0xe3: true,
	0xe4: true,
	0xe5: true,
	0xe6: true,
	0xe7: true,
	0xe8: true,
	0xe9: true,
	0xea: true,
	0xeb: true,
	0xec: true,
	0xed: true,
	0xee: true,
	0xef: true,
	0xf0: true,
	0xf1: true,
	0xf2: true,
	0xf3: true,
	0xf4: true,
	0xf5: true,
	0xf6: true,
	0xf7: true,
	0xf8: true,
	0xf9: true,
	0xfa: true,
	0xfb: true,
	0xfc: true,
	0xfd: true,
	0xfe: true,
	0xff: true,
}

var needEscape = [256]bool{
	'"':  true,
	'\\': true,
	0x00: true,
	0x01: true,
	0x02: true,
	0x03: true,
	0x04: true,
	0x05: true,
	0x06: true,
	0x07: true,
	0x08: true,
	0x09: true,
	0x0a: true,
	0x0b: true,
	0x0c: true,
	0x0d: true,
	0x0e: true,
	0x0f: true,
	0x10: true,
	0x11: true,
	0x12: true,
	0x13: true,
	0x14: true,
	0x15: true,
	0x16: true,
	0x17: true,
	0x18: true,
	0x19: true,
	0x1a: true,
	0x1b: true,
	0x1c: true,
	0x1d: true,
	0x1e: true,
	0x1f: true,
	/* 0x20 - 0x7f */
	0x80: true,
	0x81: true,
	0x82: true,
	0x83: true,
	0x84: true,
	0x85: true,
	0x86: true,
	0x87: true,
	0x88: true,
	0x89: true,
	0x8a: true,
	0x8b: true,
	0x8c: true,
	0x8d: true,
	0x8e: true,
	0x8f: true,
	0x90: true,
	0x91: true,
	0x92: true,
	0x93: true,
	0x94: true,
	0x95: true,
	0x96: true,
	0x97: true,
	0x98: true,
	0x99: true,
	0x9a: true,
	0x9b: true,
	0x9c: true,
	0x9d: true,
	0x9e: true,
	0x9f: true,
	0xa0: true,
	0xa1: true,
	0xa2: true,
	0xa3: true,
	0xa4: true,
	0xa5: true,
	0xa6: true,
	0xa7: true,
	0xa8: true,
	0xa9: true,
	0xaa: true,
	0xab: true,
	0xac: true,
	0xad: true,
	0xae: true,
	0xaf: true,
	0xb0: true,
	0xb1: true,
	0xb2: true,
	0xb3: true,
	0xb4: true,
	0xb5: true,
	0xb6: true,
	0xb7: true,
	0xb8: true,
	0xb9: true,
	0xba: true,
	0xbb: true,
	0xbc: true,
	0xbd: true,
	0xbe: true,
	0xbf: true,
	0xc0: true,
	0xc1: true,
	0xc2: true,
	0xc3: true,
	0xc4: true,
	0xc5: true,
	0xc6: true,
	0xc7: true,
	0xc8: true,
	0xc9: true,
	0xca: true,
	0xcb: true,
	0xcc: true,
	0xcd: true,
	0xce: true,
	0xcf: true,
	0xd0: true,
	0xd1: true,
	0xd2: true,
	0xd3: true,
	0xd4: true,
	0xd5: true,
	0xd6: true,
	0xd7: true,
	0xd8: true,
	0xd9: true,
	0xda: true,
	0xdb: true,
	0xdc: true,
	0xdd: true,
	0xde: true,
	0xdf: true,
	0xe0: true,
	0xe1: true,
	0xe2: true,
	0xe3: true,
	0xe4: true,
	0xe5: true,
	0xe6: true,
	0xe7: true,
	0xe8: true,
	0xe9: true,
	0xea: true,
	0xeb: true,
	0xec: true,
	0xed: true,
	0xee: true,
	0xef: true,
	0xf0: true,
	0xf1: true,
	0xf2: true,
	0xf3: true,
	0xf4: true,
	0xf5: true,
	0xf6: true,
	0xf7: true,
	0xf8: true,
	0xf9: true,
	0xfa: true,
	0xfb: true,
	0xfc: true,
	0xfd: true,
	0xfe: true,
	0xff: true,
}

// htmlSafeSet holds the value true if the ASCII character with the given
// array position can be safely represented inside a JSON string, embedded
// inside of HTML <script> tags, without any additional escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), the backslash character ("\"), HTML opening and closing
// tags ("<" and ">"), and the ampersand ("&").
var htmlSafeSet = [256]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      false,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      false,
	'=':      true,
	'>':      false,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
	0x80:     false,
	0x81:     false,
	0x82:     false,
	0x83:     false,
	0x84:     false,
	0x85:     false,
	0x86:     false,
	0x87:     false,
	0x88:     false,
	0x89:     false,
	0x8a:     false,
	0x8b:     false,
	0x8c:     false,
	0x8d:     false,
	0x8e:     false,
	0x8f:     false,
	0x90:     false,
	0x91:     false,
	0x92:     false,
	0x93:     false,
	0x94:     false,
	0x95:     false,
	0x96:     false,
	0x97:     false,
	0x98:     false,
	0x99:     false,
	0x9a:     false,
	0x9b:     false,
	0x9c:     false,
	0x9d:     false,
	0x9e:     false,
	0x9f:     false,
	0xa0:     false,
	0xa1:     false,
	0xa2:     false,
	0xa3:     false,
	0xa4:     false,
	0xa5:     false,
	0xa6:     false,
	0xa7:     false,
	0xa8:     false,
	0xa9:     false,
	0xaa:     false,
	0xab:     false,
	0xac:     false,
	0xad:     false,
	0xae:     false,
	0xaf:     false,
	0xb0:     false,
	0xb1:     false,
	0xb2:     false,
	0xb3:     false,
	0xb4:     false,
	0xb5:     false,
	0xb6:     false,
	0xb7:     false,
	0xb8:     false,
	0xb9:     false,
	0xba:     false,
	0xbb:     false,
	0xbc:     false,
	0xbd:     false,
	0xbe:     false,
	0xbf:     false,
	0xc0:     false,
	0xc1:     false,
	0xc2:     false,
	0xc3:     false,
	0xc4:     false,
	0xc5:     false,
	0xc6:     false,
	0xc7:     false,
	0xc8:     false,
	0xc9:     false,
	0xca:     false,
	0xcb:     false,
	0xcc:     false,
	0xcd:     false,
	0xce:     false,
	0xcf:     false,
	0xd0:     false,
	0xd1:     false,
	0xd2:     false,
	0xd3:     false,
	0xd4:     false,
	0xd5:     false,
	0xd6:     false,
	0xd7:     false,
	0xd8:     false,
	0xd9:     false,
	0xda:     false,
	0xdb:     false,
	0xdc:     false,
	0xdd:     false,
	0xde:     false,
	0xdf:     false,
	0xe0:     false,
	0xe1:     false,
	0xe2:     false,
	0xe3:     false,
	0xe4:     false,
	0xe5:     false,
	0xe6:     false,
	0xe7:     false,
	0xe8:     false,
	0xe9:     false,
	0xea:     false,
	0xeb:     false,
	0xec:     false,
	0xed:     false,
	0xee:     false,
	0xef:     false,
	0xf0:     false,
	0xf1:     false,
	0xf2:     false,
	0xf3:     false,
	0xf4:     false,
	0xf5:     false,
	0xf6:     false,
	0xf7:     false,
	0xf8:     false,
	0xf9:     false,
	0xfa:     false,
	0xfb:     false,
	0xfc:     false,
	0xfd:     false,
	0xfe:     false,
	0xff:     false,
}

// safeSet holds the value true if the ASCII character with the given array
// position can be represented inside a JSON string without any further
// escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), and the backslash character ("\").
var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

var hex = "0123456789abcdef"

// escapeIndex finds the index of the first char in `s` that requires escaping.
// A char requires escaping if it's outside of the range of [0x20, 0x7F] or if
// it includes a double quote or backslash.
// If no chars in `s` require escaping, the return value is -1.
func escapeIndex(s string) int {
	chunks := stringToUint64Slice(s)
	for _, n := range chunks {
		// combine masks before checking for the MSB of each byte. We include
		// `n` in the mask to check whether any of the *input* byte MSBs were
		// set (i.e. the byte was outside the ASCII range).
		mask := n | below(n, 0x20) | contains(n, '"') | contains(n, '\\')
		if (mask & msb) != 0 {
			return bits.TrailingZeros64(mask&msb) / 8
		}
	}

	valLen := len(s)
	for i := len(chunks) * 8; i < valLen; i++ {
		if needEscape[s[i]] {
			return i
		}
	}

	return -1
}

// escapeIndex finds the index of the first char in `s` that requires escaping.
// A char requires escaping if it's outside of the range of [0x20, 0x7F] or if
// it includes a double quote or backslash.
// Also, the chars <, > and & require escaping.
// If no chars in `s` require escaping, the return value is -1.
func escapeIndexWithHTMLEscape(s string) int {
	chunks := stringToUint64Slice(s)
	for _, n := range chunks {
		// combine masks before checking for the MSB of each byte. We include
		// `n` in the mask to check whether any of the *input* byte MSBs were
		// set (i.e. the byte was outside the ASCII range).
		mask := n | (n - (lsb * 0x20)) |
			((n ^ (lsb * '"')) - lsb) |
			((n ^ (lsb * '\\')) - lsb) |
			((n ^ (lsb * '<')) - lsb) |
			((n ^ (lsb * '>')) - lsb) |
			((n ^ (lsb * '&')) - lsb)
		if (mask & msb) != 0 {
			return bits.TrailingZeros64(mask&msb) / 8
		}
	}

	valLen := len(s)
	for i := len(chunks) * 8; i < valLen; i++ {
		if needEscapeWithHTML[s[i]] {
			return i
		}
	}

	return -1
}

// below return a mask that can be used to determine if any of the bytes
// in `n` are below `b`. If a byte's MSB is set in the mask then that byte was
// below `b`. The result is only valid if `b`, and each byte in `n`, is below
// 0x80.
func below(n uint64, b byte) uint64 {
	return n - expand(b)
}

// contains returns a mask that can be used to determine if any of the
// bytes in `n` are equal to `b`. If a byte's MSB is set in the mask then
// that byte is equal to `b`. The result is only valid if `b`, and each
// byte in `n`, is below 0x80.
func contains(n uint64, b byte) uint64 {
	return (n ^ expand(b)) - lsb
}

// expand puts the specified byte into each of the 8 bytes of a uint64.
func expand(b byte) uint64 {
	return lsb * uint64(b)
}

func stringToUint64Slice(s string) []uint64 {
	return *(*[]uint64)(unsafe.Pointer(&reflect.SliceHeader{
		Data: ((*reflect.StringHeader)(unsafe.Pointer(&s))).Data,
		Len:  len(s) / 8,
		Cap:  len(s) / 8,
	}))
}

func encodeEscapedString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	var (
		i, j int
	)
	if valLen >= 8 {
		chunks := stringToUint64Slice(s)
		for _, n := range chunks {
			// combine masks before checking for the MSB of each byte. We include
			// `n` in the mask to check whether any of the *input* byte MSBs were
			// set (i.e. the byte was outside the ASCII range).
			mask := n | (n - (lsb * 0x20)) |
				((n ^ (lsb * '"')) - lsb) |
				((n ^ (lsb * '\\')) - lsb) |
				((n ^ (lsb * '<')) - lsb) |
				((n ^ (lsb * '>')) - lsb) |
				((n ^ (lsb * '&')) - lsb)
			if (mask & msb) != 0 {
				j = bits.TrailingZeros64(mask&msb) / 8
				goto ESCAPE_END
			}
		}
		for i := len(chunks) * 8; i < valLen; i++ {
			if needEscapeWithHTML[s[i]] {
				j = i
				goto ESCAPE_END
			}
		}
		// no found any escape characters.
		return append(append(buf, s...), '"')
	}
ESCAPE_END:
	for j < valLen {
		c := s[j]

		if !needEscapeWithHTML[c] {
			// fast path: most of the time, printable ascii characters are used
			j++
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
		}

		// This encodes bytes < 0x20 except for \t, \n and \r.
		if c < 0x20 {
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
			continue
		}

		r, size := utf8.DecodeRuneInString(s[j:])

		if r == utf8.RuneError && size == 1 {
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\ufffd`...)
			i = j + size
			j = j + size
			continue
		}

		switch r {
		case '\u2028', '\u2029':
			// U+2028 is LINE SEPARATOR.
			// U+2029 is PARAGRAPH SEPARATOR.
			// They are both technically valid characters in JSON strings,
			// but don't work in JSONP, which has to be evaluated as JavaScript,
			// and can lead to security holes there. It is valid JSON to
			// escape them, so we do so unconditionally.
			// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u202`...)
			buf = append(buf, hex[r&0xF])
			i = j + size
			j = j + size
			continue
		}

		j += size
	}

	return append(append(buf, s[i:]...), '"')
}

func encodeNoEscapedString(buf []byte, s string) []byte {
	valLen := len(s)
	if valLen == 0 {
		return append(buf, `""`...)
	}
	buf = append(buf, '"')
	var escapeIdx int
	if valLen >= 8 {
		if escapeIdx = escapeIndex(s); escapeIdx < 0 {
			return append(append(buf, s...), '"')
		}
	}

	i := 0
	j := escapeIdx
	for j < valLen {
		c := s[j]

		if c >= 0x20 && c <= 0x7f && c != '\\' && c != '"' {
			// fast path: most of the time, printable ascii characters are used
			j++
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
		}

		// This encodes bytes < 0x20 except for \t, \n and \r.
		if c < 0x20 {
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u00`...)
			buf = append(buf, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = j + 1
			continue
		}

		r, size := utf8.DecodeRuneInString(s[j:])

		if r == utf8.RuneError && size == 1 {
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\ufffd`...)
			i = j + size
			j = j + size
			continue
		}

		switch r {
		case '\u2028', '\u2029':
			// U+2028 is LINE SEPARATOR.
			// U+2029 is PARAGRAPH SEPARATOR.
			// They are both technically valid characters in JSON strings,
			// but don't work in JSONP, which has to be evaluated as JavaScript,
			// and can lead to security holes there. It is valid JSON to
			// escape them, so we do so unconditionally.
			// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
			buf = append(buf, s[i:j]...)
			buf = append(buf, `\u202`...)
			buf = append(buf, hex[r&0xF])
			i = j + size
			j = j + size
			continue
		}

		j += size
	}

	return append(append(buf, s[i:]...), '"')
}

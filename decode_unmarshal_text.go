package json

import (
	"encoding"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

type unmarshalTextDecoder struct {
	typ *rtype
}

func newUnmarshalTextDecoder(typ *rtype) *unmarshalTextDecoder {
	return &unmarshalTextDecoder{typ: typ}
}

func (d *unmarshalTextDecoder) decodeStream(s *stream, p uintptr) error {
	s.skipWhiteSpace()
	start := s.cursor
	if err := s.skipValue(); err != nil {
		return err
	}
	src := s.buf[start:s.cursor]
	if s, ok := unquoteBytes(src); ok {
		src = s
	}
	v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: d.typ,
		ptr: unsafe.Pointer(p),
	}))
	if err := v.(encoding.TextUnmarshaler).UnmarshalText(src); err != nil {
		return err
	}
	return nil
}

func (d *unmarshalTextDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	if s, ok := unquoteBytes(src); ok {
		src = s
	}
	v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: d.typ,
		ptr: unsafe.Pointer(p),
	}))
	if err := v.(encoding.TextUnmarshaler).UnmarshalText(src); err != nil {
		return 0, err
	}
	return end, nil
}

func unquoteBytes(s []byte) (t []byte, ok bool) {
	length := len(s)
	if length < 2 || s[0] != '"' || s[length-1] != '"' {
		return
	}
	s = s[1 : length-1]
	length -= 2

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < length {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == length {
		return s, true
	}

	b := make([]byte, length+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < length {
		// Out of room? Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= length {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range s[2:6] {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}

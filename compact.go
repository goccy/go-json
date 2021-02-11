package json

import (
	"bytes"
)

type asciiSet [8]uint32

func makeASCIISet(chars string) asciiSet {
	var as asciiSet
	for _, c := range chars {
		as[c>>5] |= 1 << uint(c&31)
	}
	return as
}

func (as *asciiSet) Contains(c byte) bool {
	return (as[c>>5] & (1 << uint(c&31))) != 0
}

var flushSet = makeASCIISet("\" \t\n\r")
var escapeSet = makeASCIISet("<>&")

func compact(dst *bytes.Buffer, src []byte, escape bool) error {
	start := 0
	cursor := 0
	for {
		offset := -1
		flushSet := flushSet
		for i, c := range src[cursor:] {
			if flushSet.Contains(c) {
				offset = i
				break
			}
		}
		if offset == -1 {
			break
		}
		cursor += offset
		c := src[cursor]
		if c == '"' { // string literal
			cursor++
			escapeSet := escapeSet
			for escaped := false; cursor < len(src); cursor++ {
				c := src[cursor]
				if !escaped {
					if c == '"' {
						break
					}
					if c == '\\' {
						escaped = true
					} else if escape && escapeSet.Contains(c) {
						if start < cursor {
							dst.Write(src[start:cursor])
						}
						start = cursor + 1
						dst.Write([]byte{'\\', 'u', '0', '0', hex[c>>4], hex[c&0xf]})
					}
				} else {
					escaped = false
					if escape && escapeSet.Contains(c) {
						return errInvalidCharacter(c, "escaped string", int64(cursor))
					}
				}
			}
			if cursor == len(src) {
				return errUnexpectedEndOfJSON("string", int64(len(src)))
			}
		} else { // whitespaces
			if start < cursor {
				dst.Write(src[start:cursor])
			}
			start = cursor + 1
		}
		cursor++
	}
	if start < len(src) {
		dst.Write(src[start:])
	}
	return nil
}

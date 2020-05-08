package json

import (
	"errors"
)

var (
	isWhiteSpace = [256]bool{}
)

func init() {
	isWhiteSpace[' '] = true
	isWhiteSpace['\n'] = true
	isWhiteSpace['\t'] = true
	isWhiteSpace['\r'] = true
}

func skipWhiteSpace(buf []byte, cursor int) int {
LOOP:
	if isWhiteSpace[buf[cursor]] {
		cursor++
		goto LOOP
	}
	return cursor
}

func skipValue(buf []byte, cursor int) (int, error) {
	cursor = skipWhiteSpace(buf, cursor)
	braceCount := 0
	bracketCount := 0
	buflen := len(buf)
	for {
		switch buf[cursor] {
		case '\000':
			return cursor, errors.New("unexpected error value")
		case '{':
			braceCount++
		case '[':
			bracketCount++
		case '}':
			braceCount--
			if braceCount == -1 && bracketCount == 0 {
				return cursor, nil
			}
		case ']':
			bracketCount--
		case ',':
			if bracketCount == 0 && braceCount == 0 {
				return cursor, nil
			}
		case '"':
			cursor++

			for ; cursor < buflen; cursor++ {
				if buf[cursor] != '"' {
					continue
				}
				if buf[cursor-1] == '\\' {
					continue
				}
				if bracketCount == 0 && braceCount == 0 {
					return cursor + 1, nil
				}
				break
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			cursor++
			for ; cursor < buflen; cursor++ {
				tk := int(buf[cursor])
				if (int('0') <= tk && tk <= int('9')) || tk == '.' || tk == 'e' || tk == 'E' {
					continue
				}
				break
			}
			if bracketCount == 0 && braceCount == 0 {
				return cursor, nil
			}
			continue
		}
		cursor++
	}
	return cursor, errors.New("unexpected error value")
}

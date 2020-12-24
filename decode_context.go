package json

var (
	isWhiteSpace = [256]bool{}
)

func init() {
	isWhiteSpace[' '] = true
	isWhiteSpace['\n'] = true
	isWhiteSpace['\t'] = true
	isWhiteSpace['\r'] = true
}

func skipWhiteSpace(buf []byte, cursor int64) int64 {
LOOP:
	if isWhiteSpace[buf[cursor]] {
		cursor++
		goto LOOP
	}
	return cursor
}

func skipValue(buf []byte, cursor int64) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	braceCount := 0
	bracketCount := 0
	buflen := int64(len(buf))
	start := cursor
	for {
		switch buf[cursor] {
		case nul:
			if start == cursor {
				return cursor, errUnexpectedEndOfJSON("value of object", cursor)
			}
			if braceCount == 0 && bracketCount == 0 {
				return cursor, nil
			}
			return cursor, errUnexpectedEndOfJSON("value of object", cursor)
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
			if braceCount == 0 && bracketCount == -1 {
				return cursor, nil
			}
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
		case 't':
			if cursor+3 >= buflen {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			if buf[cursor+1] != 'r' {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			if buf[cursor+2] != 'u' {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			if buf[cursor+3] != 'e' {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			cursor += 4
			if bracketCount == 0 && braceCount == 0 {
				return cursor, nil
			}
			continue
		case 'f':
			if cursor+4 >= buflen {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			if buf[cursor+1] != 'a' {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			if buf[cursor+2] != 'l' {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			if buf[cursor+3] != 's' {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			if buf[cursor+4] != 'e' {
				return 0, errUnexpectedEndOfJSON("bool of object", cursor)
			}
			cursor += 5
			if bracketCount == 0 && braceCount == 0 {
				return cursor, nil
			}
			continue
		case 'n':
			if cursor+3 >= buflen {
				return 0, errUnexpectedEndOfJSON("null", cursor)
			}
			if buf[cursor+1] != 'u' {
				return 0, errUnexpectedEndOfJSON("null", cursor)
			}
			if buf[cursor+2] != 'l' {
				return 0, errUnexpectedEndOfJSON("null", cursor)
			}
			if buf[cursor+3] != 'l' {
				return 0, errUnexpectedEndOfJSON("null", cursor)
			}
			cursor += 4
			if bracketCount == 0 && braceCount == 0 {
				return cursor, nil
			}
			continue
		}
		cursor++
	}
	return cursor, errUnexpectedEndOfJSON("value of object", cursor)
}

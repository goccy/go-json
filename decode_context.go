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

func skipWhiteSpace(buf []byte, cursor int) int {
	buflen := len(buf)
	for ; cursor < buflen; cursor++ {
		if isWhiteSpace[buf[cursor]] {
			continue
		}
		return cursor
	}
	return buflen
}

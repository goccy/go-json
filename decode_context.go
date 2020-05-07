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
LOOP:
	if isWhiteSpace[buf[cursor]] {
		cursor++
		goto LOOP
	}
	return cursor
}

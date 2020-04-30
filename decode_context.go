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

type context struct {
	cursor int
	buf    []byte
	buflen int
}

func (c *context) setBuf(buf []byte) {
	c.buf = buf
	c.buflen = len(buf)
	c.cursor = 0
}

func (c *context) skipWhiteSpace() int {
	buflen := c.buflen
	buf := c.buf
	for cursor := c.cursor; cursor < buflen; cursor++ {
		if isWhiteSpace[buf[cursor]] {
			continue
		}
		c.cursor = cursor
		return cursor
	}
	return buflen
}

func newContext() *context {
	return &context{}
}

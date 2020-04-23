package json

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
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
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

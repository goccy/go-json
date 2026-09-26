package decoder

import (
	"fmt"

	"github.com/goccy/go-json/internal/errors"
)

// syntaxErrorAt returns the syntax error of the byte at cursor, which is not what the grammar has in the place
// which what tells, as encoding/json has it: the end of the input if it is the nul byte which ends the buffer,
// and else the byte, which the offset is after. It is not inlined, so that its callers keep their size.
//
//go:noinline
func syntaxErrorAt(buf []byte, cursor int64, what string) error {
	if cursor == int64(len(buf))-1 && buf[cursor] == nul {
		return errors.ErrSyntax("unexpected end of JSON input", cursor)
	}
	return errors.ErrSyntax(fmt.Sprintf("invalid character %s %s", quoteChar(buf[cursor]), what), cursor+1)
}

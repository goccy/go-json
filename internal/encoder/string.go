package encoder

import (
	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonstring"
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
)

// StringEscaper is the escaper of the strings of the options. It is small enough to be inlined into the VM, which
// calls jsonstring.AppendQuoted with it.
func StringEscaper(ctx *RuntimeContext) *jsonstring.Escaper {
	return jsonstring.EscaperOf(uint(ctx.Option.Flag))
}

// the options of the escaper are the bits of its index which jsonstring.EscaperOf takes: the length is 0 only then.
var _ [0]struct{} = [uint(NormalizeUTF8Option^jsonstring.EscapeNormalize) | uint(HTMLEscapeOption^jsonstring.EscapeHTML) |
	uint(TextEscapeOption^jsonstring.EscapeText) | uint(RejectInvalidUTF8Option^jsonstring.EscapeStrict)]struct{}{}

// textEscaper is the escaper of a string which the encoder writes by a function, not by the VM: of
// RejectInvalidUTF8Option, it reports invalid UTF-8 instead of panicking ( see InvalidUTF8 ).
func textEscaper(ctx *RuntimeContext) *jsonstring.Escaper {
	return jsonstring.EscaperOf(uint(ctx.Option.Flag &^ RejectInvalidUTF8Option))
}

// InvalidUTF8 returns the output of a string which has invalid UTF-8, written by textEscaper, after the output
// before it: an error of RejectInvalidUTF8Option, or the string with U+FFFD.
func InvalidUTF8(ctx *RuntimeContext, before, after []byte) ([]byte, error) {
	if ctx.Option.Flag&RejectInvalidUTF8Option == 0 {
		return after, nil
	}
	return before, &errors.TextError{Err: errors.ErrInvalidUTF8}
}

// AppendString appends the string as a JSON string, escaped as the options want it ( see jsonstring.AppendQuoted ).
func AppendString(ctx *RuntimeContext, buf []byte, s string) []byte {
	buf, _ = jsonstring.AppendQuoted(StringEscaper(ctx), buf, s)
	return buf
}

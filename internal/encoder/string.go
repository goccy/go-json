package encoder

import (
	"github.com/goccy/go-json/internal/jsonstring"
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
)

// StringEscaper is the escaper of the strings of the options. It is small enough to be inlined into the VM, which
// calls jsonstring.AppendQuoted with it.
func StringEscaper(ctx *RuntimeContext) *jsonstring.Escaper {
	return jsonstring.V1EscaperOf(uint(ctx.Option.Flag))
}

// the bits of the options which jsonstring.V1EscaperOf takes are the ones of HTMLEscapeOption and
// NormalizeUTF8Option: the length is 0 only then.
var _ [0]struct{} = [uint(HTMLEscapeOption^1<<jsonstring.V1HTMLBit) | uint(NormalizeUTF8Option^1<<jsonstring.V1NormalizeBit)]struct{}{}

// AppendString appends the string as a JSON string, escaped as the options want it ( see jsonstring.AppendQuoted ).
func AppendString(ctx *RuntimeContext, buf []byte, s string) []byte {
	buf, _ = jsonstring.AppendQuoted(StringEscaper(ctx), buf, s)
	return buf
}

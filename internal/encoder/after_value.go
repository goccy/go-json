package encoder

import (
	"bytes"
	"unsafe"
)

// The opcode after a value ( OpAfterValue ) changes what the value wrote, for the semantics of encoding/json/v2
// which depend on the output of a value: a member of omitempty is unwritten if its value is empty. The output is
// compact, a value followed by a comma, and what the value wrote is found from the end of the output.

// AppendAfterValue returns the output after the value which the opcode follows, by its function.
func AppendAfterValue(ctx *RuntimeContext, code *Opcode, b []byte) ([]byte, error) {
	return code.Marshaler.appendValue(ctx, b, nil)
}

// newAfterValueCode returns the opcode after a value which calls fn with the output.
func newAfterValueCode(ctx *compileContext, fn AppendFunc) *Opcode {
	code := &Opcode{
		Op:         OpAfterValue,
		Idx:        opcodeOffset(ctx.ptrIndex),
		DisplayIdx: ctx.opcodeIndex,
		Indent:     ctx.indent,
		Marshaler:  &MarshalerCall{appendValue: fn},
	}
	ctx.incOpcodeIndex()
	return code
}

// emptyValues are the values which a member of omitempty is omitted for, as the v2 semantics define it: the
// values which would be encoded as null, "", {} or [].
var emptyValues = [...][]byte{[]byte(`null`), []byte(`""`), []byte(`{}`), []byte(`[]`)}

// unwriteEmptyMember returns the function which unwrites the last member of the output, whose key of keyLen bytes
// is followed by its value and a comma, if the value is empty.
//
// A value which ends with the bytes of an empty value after the colon of its key is that value itself: within a
// longer string the quote would be escaped, a longer object or array would end with its own bracket, and null is
// not the end of another value.
func unwriteEmptyMember(keyLen int) AppendFunc {
	return func(_ *RuntimeContext, b []byte, _ unsafe.Pointer) ([]byte, error) {
		end := len(b) - 1 // the comma
		for _, empty := range emptyValues {
			start := end - len(empty)
			if start >= keyLen && b[start-1] == ':' && bytes.Equal(b[start:end], empty) {
				return b[:start-keyLen], nil
			}
		}
		return b, nil
	}
}

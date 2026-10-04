package encoder

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

// newUnwriteEmptyCode returns the opcode after the value of a member of omitempty, whose key is key, which
// unwrites the member if the value is empty ( see UnwriteEmptyMember ).
func newUnwriteEmptyCode(ctx *compileContext, key string) *Opcode {
	code := &Opcode{
		Op:         OpUnwriteEmpty,
		Idx:        opcodeOffset(ctx.ptrIndex),
		Key:        key,
		DisplayIdx: ctx.opcodeIndex,
		Indent:     ctx.indent,
	}
	ctx.incOpcodeIndex()
	return code
}

// UnwriteEmptyMember unwrites the last member of the output b, a member of the key followed by its value and a
// comma, if the value is empty, as the v2 semantics define it: null, "", {} or []. The opcode of the field may
// have written nothing, for a nil pointer which omitempty omits: the last member is then of another field, or
// there is none, and the output is kept.
//
// A value which ends with the bytes of an empty value after a colon is that value itself, the value of a member
// of the object: within a longer string the quote would be escaped, a longer object or array would end with its
// own bracket, and null is not the end of another value.
func UnwriteEmptyMember(ctx *RuntimeContext, b []byte, key string) []byte {
	end := len(b) - 1 // the comma
	n := 2
	if end < n {
		return b
	}
	switch b[end-1] {
	case '"':
		if b[end-2] != '"' {
			return b
		}
	case '}':
		if b[end-2] != '{' {
			return b
		}
	case ']':
		if b[end-2] != '[' {
			return b
		}
	case 'l':
		n = 4
		if end < n || string(b[end-n:end]) != "null" {
			return b
		}
	default:
		return b
	}
	if start := end - n - len(key); start >= 0 && string(b[start:end-n]) == key {
		ctx.Rewrote(start)
		return b[:start]
	}
	return b
}

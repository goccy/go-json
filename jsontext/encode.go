package jsontext

import (
	"bytes"
	"io"
	"math"
)

// Encoder is a streaming encoder of raw JSON tokens and values. It writes a stream of top-level JSON values,
// each followed by a line feed.
//
// Encoder.WriteToken and Encoder.WriteValue calls may be interleaved. For example, the JSON value
//
//	{"name":"value","array":[null,false,true,3.14159],"object":{"k":"v"}}
//
// can be written by these calls (ignoring errors):
//
//	e.WriteToken(BeginObject)        // {
//	e.WriteToken(String("name"))     // "name"
//	e.WriteToken(String("value"))    // "value"
//	e.WriteValue(Value(`"array"`))   // "array"
//	e.WriteToken(BeginArray)         // [
//	e.WriteToken(Null)               // null
//	e.WriteToken(False)              // false
//	e.WriteValue(Value("true"))      // true
//	e.WriteToken(Float(3.14159))     // 3.14159
//	e.WriteToken(EndArray)           // ]
//	e.WriteValue(Value(`"object"`))  // "object"
//	e.WriteValue(Value(`{"k":"v"}`)) // {"k":"v"}
//	e.WriteToken(EndObject)          // }
//
// which is one of many sequences of calls which write it.
type Encoder struct {
	e encoder
}

// encoder is the state of an Encoder.
type encoder struct {
	cfg config
	st  stack
	vs  valueScanner

	w    io.Writer
	bb   *bytes.Buffer // w, whose array the output is appended to
	buf  []byte        // the output which is not written to w yet
	base int64         // the offset of buf[0] in the output

	ws        whitespace
	esc       escapeFlags
	validUTF8 bool
	check     bool // check duplicate names
}

// NewEncoder constructs a streaming encoder which writes to w, with the options. It writes its buffer to w when
// the buffer is full enough, or when a top-level value is complete.
//
// If w is a bytes.Buffer, the encoder appends to the buffer directly, without copying the output from a buffer
// of its own.
func NewEncoder(w io.Writer, opts ...Options) *Encoder {
	e := new(Encoder)
	e.Reset(w, opts...)
	return e
}

// Reset resets the encoder to write to w anew, with the options.
func (e *Encoder) Reset(w io.Writer, opts ...Options) {
	e.e.reset(w, opts)
}

func (e *encoder) reset(w io.Writer, opts []Options) {
	e.cfg = config{}
	e.cfg.apply(opts)
	e.st.reset()
	e.w = w
	e.bb, _ = w.(*bytes.Buffer)
	e.buf = e.buf[:0]
	if e.bb != nil {
		e.buf = e.bb.AvailableBuffer()
	}
	e.base = 0
	e.ws = e.cfg.whitespace()
	e.esc = e.cfg.escapes()
	e.validUTF8 = !e.cfg.has(allowInvalidUTF8)
	e.check = !e.cfg.has(allowDuplicateNames)
	e.vs = valueScanner{
		st:         &e.st,
		validUTF8:  e.validUTF8,
		checkNames: e.check,
		final:      true,
		write:      true,
		ws:         e.ws,
		esc:        e.esc,
		preserve:   e.cfg.has(preserveRawStrings),
		canonInts:  e.cfg.has(canonicalizeRawInts),
		canonFlts:  e.cfg.has(canonicalizeRawFloats),
		reorder:    e.cfg.has(reorderRawObjects),
		scratch:    e.vs.scratch[:0],
		members:    e.vs.members[:0],
	}
}

// Options returns the options which the encoder was constructed with.
func (e *Encoder) Options() Options {
	c := e.e.cfg
	return &c
}

// OutputOffset returns the offset in the output after the last token or value which was written. The encoder
// may not have written all of it to the writer yet.
func (e *Encoder) OutputOffset() int64 {
	return e.e.base + int64(len(e.e.buf))
}

// AvailableBuffer returns an empty buffer which may have room, to build a Value for the next call of
// Encoder.WriteValue:
//
//	b := e.AvailableBuffer()
//	b = append(b, '"')
//	b = appendString(b, v) // append the string formatting of v
//	b = append(b, '"')
//	... := e.WriteValue(b)
//
// WriteValue takes a valid JSON value: a Value built as raw bytes takes more care than a Token of a
// constructor such as String.
func (e *Encoder) AvailableBuffer() []byte {
	return e.e.buf[len(e.e.buf):]
}

// StackDepth returns the number of the objects and arrays which are open in the output: 0 at the top level,
// before any token, after a top-level value and between top-level values, 1 in a top-level object or array, and
// so on.
func (e *Encoder) StackDepth() int {
	return e.e.st.depth()
}

// StackIndex returns the kind and the length of the level i of the stack, which is from 0 to
// Encoder.StackDepth: KindInvalid for the level 0, KindBeginObject for an object, KindBeginArray for an array.
// The length of an object counts its names and values: a complete object has an even length.
func (e *Encoder) StackIndex(i int) (Kind, int64) {
	return e.e.st.index(i)
}

// StackPointer returns a JSON Pointer (RFC 6901) to the last value which was written.
func (e *Encoder) StackPointer() Pointer {
	return e.e.st.pointer(pointLast)
}

// failAt is a SyntacticError at the position pos of the output buffer.
func (e *encoder) failAt(err error, pos int, where int) error {
	return &SyntacticError{ByteOffset: e.base + int64(pos), JSONPointer: e.st.pointer(where), Err: err}
}

// appendDelim appends the comma or colon and the white space before a token of kind k, for which the grammar
// was checked.
func (e *encoder) appendDelim(b []byte, k Kind) []byte {
	l := e.st.last()
	if e.st.depth() == 0 {
		return b
	}
	if k == KindEndObject || k == KindEndArray {
		if l.count > 0 && e.ws.multiline {
			b = e.ws.appendLine(b, e.st.depth()-1)
		}
		return b
	}
	if l.needValue() {
		b = append(b, ':')
		if e.ws.colon {
			b = append(b, ' ')
		}
		return b
	}
	if l.count > 0 {
		b = e.ws.appendComma(b)
	}
	if e.ws.multiline {
		b = e.ws.appendLine(b, e.st.depth())
	}
	return b
}

// WriteToken writes the next token and advances the write offset.
//
// The token must fit the grammar: a number where the encoder needs an object name, which is always a string,
// or the end of an object where an array ends, is an error. At an invalid token, it reports a SyntacticError,
// and the state of the encoder doesn't change: the offset of the error is Encoder.OutputOffset plus the
// delimiters and white space which would precede the token.
func (e *Encoder) WriteToken(t Token) error {
	return e.e.writeToken(t)
}

func (e *encoder) writeToken(t Token) error {
	k := t.Kind()
	l := e.st.last()
	switch k {
	case KindEndObject, KindEndArray:
		if e.st.depth() == 0 || l.object != (k == KindEndObject) {
			return e.failAt(errMismatchDelim, len(e.buf)+e.delimLen(k), pointNext)
		}
		if l.needValue() {
			return e.failAt(errMissingValue, len(e.buf)+e.delimLen(KindString), pointAt)
		}
	case KindInvalid:
		return e.failAt(errInvalidToken, len(e.buf), pointNext)
	default:
		if l.needName() && k != KindString {
			return e.failAt(ErrNonStringName, len(e.buf)+e.delimLen(k), pointAt)
		}
	}
	b := e.appendDelim(e.buf, k)
	pos := len(b)
	var err error
	switch k {
	case KindNull, KindTrue, KindFalse:
		b = append(b, literalOf(k)...)
	case KindBeginObject, KindBeginArray:
		if err := e.st.push(k == KindBeginObject); err != nil {
			return e.failAt(err, pos, pointNext)
		}
		e.buf = append(b, byte(k))
		e.endValue()
		return nil
	case KindEndObject, KindEndArray:
		e.st.pop()
		e.buf = append(b, byte(k))
		e.endValue()
		return nil
	case KindString:
		b, err = e.appendString(b, t, l.needName())
	default:
		b, err = e.appendNumberToken(b, t)
	}
	if err != nil {
		return err
	}
	e.st.last().count++
	e.buf = b
	e.endValue()
	return nil
}

// delimLen is the length of the delimiter and the white space which precede a token of kind k.
func (e *encoder) delimLen(k Kind) int {
	var buf [64]byte
	return len(e.appendDelim(buf[:0], k))
}

// appendString appends the string token t, a name of an object if name is set.
func (e *encoder) appendString(b []byte, t Token, name bool) ([]byte, error) {
	pos := len(b)
	switch t.form {
	case formString:
		var valid bool
		b, valid = appendQuoted(b, t.str, e.esc)
		if !valid && e.validUTF8 {
			return b, e.failAt(errInvalidUTF8, pos, pointNext)
		}
		if name {
			return b, e.insertName(pos, []byte(t.str), !valid)
		}
		return b, nil
	case formFloat, formFloat32:
		// NaN or an infinity
		b = append(b, '"')
		b = appendFloatText(b, math.Float64frombits(t.num), 64)
		b = append(b, '"')
		if name {
			return b, e.insertName(pos, b[pos+1:len(b)-1], false)
		}
		return b, nil
	}
	raw := t.raw()
	_, f, err := scanString(raw, strModeOf(e.validUTF8, false))
	if err != nil {
		return b, e.failAt(err, pos, pointNext)
	}
	b = appendRawString(b, raw, f, e.esc, e.cfg.has(preserveRawStrings))
	if name {
		var buf [64]byte
		return b, e.insertName(pos, appendUnquoted(buf[:0], raw), false)
	}
	return b, nil
}

// insertName adds a name of the innermost object, written at pos.
func (e *encoder) insertName(pos int, name []byte, mangle bool) error {
	if mangle {
		name = []byte(string(bytes.ToValidUTF8(name, []byte("�"))))
	}
	if !e.st.insertName(name, e.check) {
		ptr := e.st.namePointer(name)
		return &SyntacticError{ByteOffset: e.base + int64(pos), JSONPointer: ptr, Err: ErrDuplicateName}
	}
	return nil
}

// appendNumberToken appends the number token t.
func (e *encoder) appendNumberToken(b []byte, t Token) ([]byte, error) {
	if t.form != formRaw {
		return t.appendNumber(b), nil
	}
	raw := t.raw()
	if n, err := scanNumber(raw); err != nil || n < len(raw) {
		return b, e.failAt(invalidChar(raw[n:], "in number (expecting digit)"), len(b)+n, pointNext)
	}
	return e.vs.appendNumber(b, raw), nil
}

// endValue ends a top-level value which is complete: a line feed follows it, and the output is written.
func (e *encoder) endValue() {
	if e.st.depth() > 0 {
		if len(e.buf) > 1<<16 {
			e.flush()
		}
		return
	}
	if !e.cfg.has(omitTopLevelNewline) {
		e.buf = append(e.buf, '\n')
	}
	e.flush()
}

// flush writes the buffer to the writer.
func (e *encoder) flush() error {
	if e.w == nil || len(e.buf) == 0 {
		return nil
	}
	n, err := e.w.Write(e.buf)
	e.base += int64(n)
	if e.bb != nil {
		e.buf = e.bb.AvailableBuffer()
	} else {
		e.buf = e.buf[:copy(e.buf, e.buf[n:])]
	}
	return err
}

// WriteValue writes the next raw value and advances the write offset. The encoder checks that the value is
// valid, and formats its white space and strings as the options say. With AllowInvalidUTF8, invalid UTF-8 is
// mangled as the Unicode replacement character, U+FFFD.
//
// The value must fit the grammar (see the examples of Encoder.WriteToken). At an invalid value, it reports a
// SyntacticError, and the state of the encoder doesn't change: the offset of the error is Encoder.OutputOffset
// plus the offset of the error in v.
func (e *Encoder) WriteValue(v Value) error {
	return e.e.writeValue(v)
}

func (e *encoder) writeValue(v Value) error {
	l := e.st.last()
	k := v.Kind()
	name := l.needName()
	b := e.appendDelim(e.buf, k)
	pos := len(b)
	e.vs.out = b
	end, serr := e.vs.scan(v, 0)
	b = e.vs.out
	e.vs.out = nil
	if serr != nil {
		return &SyntacticError{ByteOffset: e.base + int64(pos+serr.pos), JSONPointer: serr.ptr, Err: serr.err}
	}
	if end = skipSpace(v, end); end < len(v) {
		e.st.last().count--
		return &SyntacticError{ByteOffset: e.base + int64(pos+end), Err: invalidChar(v[end:], "after top-level value")}
	}
	if name {
		e.st.last().count--
		if k != KindString {
			return e.failAt(ErrNonStringName, pos, pointAt)
		}
		// the name, which the scanner wrote, is unquoted from the output
		var buf [64]byte
		if err := e.insertName(pos, appendUnquoted(buf[:0], b[pos:]), false); err != nil {
			return err
		}
		e.st.last().count++
	}
	e.buf = b
	e.endValue()
	return nil
}

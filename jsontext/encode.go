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

	// the options as init takes them, which follow from cfg, and the options of vs: derived is the cfg which
	// they were set from, once derivedOK is set, so that an encoder which is used again with the same options
	// doesn't set them again.
	derived   config
	derivedOK bool
	ws        whitespace
	esc       escapeFlags
	validUTF8 bool
	check     bool // check duplicate names

	avail []byte // the buffer of AvailableBuffer, which the output doesn't share
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
	if w == nil {
		panic("jsontext: invalid nil io.Writer")
	}
	e.e.reset(w, opts)
}

func (e *encoder) reset(w io.Writer, opts []Options) {
	var c config // opts may hold the options of e, which Options returned
	c.apply(opts)
	e.cfg = c
	e.init(w)
}

// init resets the state of the encoder to write to w with the options of e.cfg.
func (e *encoder) init(w io.Writer) {
	e.st.reset()
	if e.bb != nil {
		e.buf = nil // the array of the bytes.Buffer written before, which stays its own
	}
	e.w = w
	e.bb, _ = w.(*bytes.Buffer)
	e.buf = e.buf[:0]
	if e.bb != nil {
		e.buf = e.bb.AvailableBuffer()
	}
	e.base = 0
	if !e.derivedOK || !e.cfg.same(&e.derived) {
		e.derive()
	}
	// the fields are set one by one: a literal of the scanner would be built in a temporary and copied.
	vs := &e.vs
	vs.write, vs.keep = true, false
	vs.spaced = e.ws.multiline || e.ws.colon || e.ws.comma
	vs.run = 0
	vs.wsFrom, vs.wsEnd = 0, 0
	if vs.in != nil || vs.out != nil {
		vs.in, vs.out = nil, nil
	}
}

// derive sets the fields which follow from the options, e.cfg.
func (e *encoder) derive() {
	e.ws = e.cfg.whitespace()
	e.esc = e.cfg.escapes()
	e.validUTF8 = !e.cfg.has(allowInvalidUTF8)
	e.check = !e.cfg.has(allowDuplicateNames)
	vs := &e.vs
	vs.st = &e.st
	vs.validUTF8 = e.validUTF8
	vs.checkNames = e.check
	vs.lazyNames = !e.check
	vs.decoding = false
	vs.final = true
	vs.ws = e.ws
	vs.esc = e.esc
	vs.preserve = e.cfg.has(preserveRawStrings)
	vs.canonInts = e.cfg.has(canonicalizeRawInts)
	vs.canonFlts = e.cfg.has(canonicalizeRawFloats)
	vs.reorder = e.cfg.has(reorderRawObjects)
	vs.lines = vs.lines[:0] // the lines of the white space of other options
	// whitespace may have set the defaults of Multiline in cfg: the options which were given differ from it, and
	// are derived again.
	e.derived, e.derivedOK = e.cfg, true
}

// Options returns the options of the encoder: the ones it was constructed or last reset with, and, for Multiline,
// the space after a colon and the indentation it takes if they are not set. They follow a later Reset.
func (e *Encoder) Options() Options {
	return &e.e.cfg
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
	// a buffer of its own: the room of the output would be written over by the value built in it.
	if e.e.avail == nil {
		e.e.avail = make([]byte, 0, 64)
	}
	return e.e.avail[:0]
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
	if l.needValue() {
		// a colon, which is also the delimiter of the offset of an error of an end delimiter there
		b = append(b, ':')
		if e.ws.colon {
			b = append(b, ' ')
		}
		return b
	}
	if k == KindEndObject || k == KindEndArray {
		if l.count > 0 && e.ws.multiline {
			b = e.ws.appendLine(b, e.st.depth()-1)
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
	if misplaced(l, k, e.st.depth()) {
		return e.misplacedError(l, k)
	}
	b := e.buf
	switch {
	case e.vs.spaced || len(e.st.levels) == 1 || l.count == 0 || k == KindEndObject || k == KindEndArray:
		b = e.appendDelim(b, k)
	case l.object && l.count&1 == 1:
		b = append(b, ':')
	default:
		b = append(b, ',')
	}
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
		return e.endValue()
	case KindEndObject, KindEndArray:
		e.st.pop()
		e.buf = append(b, byte(k))
		return e.endValue()
	case KindString:
		b, err = e.appendString(b, t, l.needName())
	default:
		if t.src == &floatSource {
			b = appendFloat(b, math.Float64frombits(t.num), 64) // a finite number, whose kind is KindNumber
		} else {
			b, err = e.appendNumberToken(b, t)
		}
	}
	if err != nil {
		return err
	}
	l.count++
	e.buf = b
	return e.endValue()
}

// misplaced reports whether a token of kind k doesn't fit the grammar in the innermost level l at depth.
func misplaced(l *level, k Kind, depth int) bool {
	switch k {
	case KindEndObject, KindEndArray:
		return depth == 0 || l.object != (k == KindEndObject) || l.needValue()
	case KindInvalid:
		return true
	}
	return l.needName() && k != KindString
}

// misplacedError is the error of a token of kind k which is misplaced in the innermost level l.
func (e *encoder) misplacedError(l *level, k Kind) error {
	switch {
	case k == KindInvalid:
		return e.failAt(errInvalidToken, len(e.buf), pointNext)
	case k != KindEndObject && k != KindEndArray:
		return e.failAt(ErrNonStringName, len(e.buf)+e.delimLen(k), pointAt)
	case e.st.depth() == 0 || l.object != (k == KindEndObject):
		return e.failAt(errMismatchDelim, len(e.buf)+e.delimLen(k), pointNext)
	}
	return e.failAt(errMissingValue, len(e.buf)+e.delimLen(k), pointAt)
}

// delimLen is the length of the delimiter and the white space which precede a token of kind k.
func (e *encoder) delimLen(k Kind) int {
	var buf [64]byte
	return len(e.appendDelim(buf[:0], k))
}

// appendString appends the string token t, a name of an object if name is set.
func (e *encoder) appendString(b []byte, t Token, name bool) ([]byte, error) {
	pos := len(b)
	switch t.source().form {
	case formString:
		var valid bool
		b, valid = appendQuoted(b, t.str, e.esc)
		if !valid && e.validUTF8 {
			return b, e.failAt(errInvalidUTF8, pos, pointNext)
		}
		if name {
			return b, e.insertName(pos, readOnlyBytes(t.str), !valid) // the name is copied
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
		name = appendValidUTF8(nil, name)
	}
	if !e.st.insertName(name, e.check) {
		ptr := e.st.namePointer(name)
		return &SyntacticError{ByteOffset: e.base + int64(pos), JSONPointer: ptr, Err: ErrDuplicateName}
	}
	return nil
}

// appendNumberToken appends the number token t.
func (e *encoder) appendNumberToken(b []byte, t Token) ([]byte, error) {
	if t.source().form != formRaw {
		return t.appendNumber(b), nil
	}
	raw := t.raw()
	if n, err := scanNumber(raw); err != nil || n < len(raw) {
		return b, e.failAt(invalidChar(raw[n:], "in number (expecting digit)"), len(b)+n, pointNext)
	}
	if e.vs.numberKept(raw) {
		return append(b, raw...), nil
	}
	return appendCanonicalNumber(b, raw), nil
}

// flushSize is the size of the output from which it is written before a top-level value is complete: the
// buffer of a writer stays about as small as it, unless a token or value is larger.
const flushSize = 4 << 10

// endValue ends a token or value: after a complete top-level value, a line feed, and the output is written,
// as it is when the buffer is large.
func (e *encoder) endValue() error {
	// it is inlined where the output is kept.
	if len(e.st.levels) > 1 && len(e.buf) <= flushSize {
		return nil
	}
	return e.endOutput()
}

func (e *encoder) endOutput() error {
	if e.st.depth() > 0 {
		return e.flush()
	}
	if !e.cfg.has(omitTopLevelNewline) {
		e.buf = append(e.buf, '\n')
	}
	return e.flush()
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
	if err != nil {
		return &ioError{write: true, err: err}
	}
	return nil
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
	if e.bb != nil && overlaps(v, e.buf[len(e.buf):]) {
		// a value in the free space of the bytes.Buffer, whose output is written over it before it is read
		v = bytes.Clone(v)
	}
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
	if end < len(v) {
		// white space after the value, or an error
		if end = skipSpace(v, end); end < len(v) {
			e.st.last().count--
			return e.failAt(invalidChar(v[end:], "after top-level value"), pos+end, pointAt)
		}
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
	return e.endValue()
}

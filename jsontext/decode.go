package jsontext

import (
	"bytes"
	"io"
)

// Decoder is a streaming decoder of raw JSON tokens and values. It reads a stream of top-level JSON values,
// separated by optional white space.
//
// Decoder.ReadToken and Decoder.ReadValue calls may be interleaved. For example, the JSON value
//
//	{"name":"value","array":[null,false,true,3.14159],"object":{"k":"v"}}
//
// can be read by these calls (ignoring errors):
//
//	d.ReadToken() // {
//	d.ReadToken() // "name"
//	d.ReadToken() // "value"
//	d.ReadValue() // "array"
//	d.ReadToken() // [
//	d.ReadToken() // null
//	d.ReadToken() // false
//	d.ReadValue() // true
//	d.ReadToken() // 3.14159
//	d.ReadToken() // ]
//	d.ReadValue() // "object"
//	d.ReadValue() // {"k":"v"}
//	d.ReadToken() // }
//
// which is one of many sequences of calls which read it.
type Decoder struct {
	d decoder
}

// decoder is the state of a Decoder.
type decoder struct {
	cfg config
	st  stack
	vs  valueScanner

	r      io.Reader
	bb     *bytes.Buffer // r, which is read in place
	buf    []byte        // the input which was read and is kept: from the last token or value on
	shared bool          // buf is the array of bb, which the decoder must not write to
	base   int64         // the offset of buf[0] in the input

	pos                int // the read offset: the end of the last token or value
	prevStart, prevEnd int // the last token or value
	rerr               error
	peekErr            error // an error which PeekKind found, which the next read reports
}

// NewDecoder constructs a streaming decoder which reads from r.
//
// If r is a bytes.Buffer, the decoder reads the buffer in place, without copying its contents. The buffer must
// not be written to while the decoder is used.
func NewDecoder(r io.Reader, opts ...Options) *Decoder {
	d := new(Decoder)
	d.Reset(r, opts...)
	return d
}

// Reset resets the decoder to read from r anew, with the options.
func (d *Decoder) Reset(r io.Reader, opts ...Options) {
	s := &d.d
	s.cfg = config{}
	s.cfg.apply(opts)
	s.st.reset()
	s.r = r
	s.bb, _ = r.(*bytes.Buffer)
	if s.shared {
		s.buf, s.shared = nil, false
	}
	s.buf = s.buf[:0]
	s.base, s.pos, s.prevStart, s.prevEnd = 0, 0, 0, 0
	s.rerr, s.peekErr = nil, nil
	s.vs = valueScanner{
		st:         &s.st,
		validUTF8:  !s.cfg.has(allowInvalidUTF8),
		checkNames: !s.cfg.has(allowDuplicateNames),
		decoding:   true,
		scratch:    s.vs.scratch[:0],
	}
}

// Options returns the options which the decoder was constructed with.
func (d *Decoder) Options() Options {
	c := d.d.cfg
	return &c
}

func (d *decoder) prevOffset() int64  { return d.base + int64(d.prevStart) }
func (d *decoder) prevBuffer() []byte { return d.buf[d.prevStart:d.prevEnd] }

// minRead is the room which the buffer has for a read.
const minRead = 512

// fetch reads more input to the end of the buffer. It may move the buffer, and drop the input before the last
// token: it returns the number of bytes by which the positions in the buffer moved back. At the end of the
// input, it returns the error of the reader.
func (d *decoder) fetch() (int, error) {
	if d.rerr != nil {
		return 0, d.rerr
	}
	if d.bb != nil {
		b := d.bb.Next(d.bb.Len())
		if len(b) == 0 {
			d.rerr = io.EOF
			return 0, io.EOF
		}
		if len(d.buf) == 0 && !d.shared {
			d.buf, d.shared = b, true
			return 0, nil
		}
		shift := d.makeRoom(len(b))
		d.buf = append(d.buf, b...)
		return shift, nil
	}
	shift := d.makeRoom(minRead)
	for tries := 0; ; tries++ {
		n, err := d.r.Read(d.buf[len(d.buf):cap(d.buf)])
		d.buf = d.buf[:len(d.buf)+n]
		if err != nil {
			d.rerr = err
		}
		if n > 0 {
			return shift, nil
		}
		if err != nil {
			return shift, err
		}
		if tries == 100 {
			d.rerr = io.ErrNoProgress
			return shift, d.rerr
		}
	}
}

// makeRoom makes room for n more bytes at the end of the buffer, dropping the input before the last token only
// when the buffer is full, so that each byte is moved a few times at most. It returns by how much the
// positions moved back.
func (d *decoder) makeRoom(n int) int {
	if cap(d.buf)-len(d.buf) >= n && !d.shared {
		return 0
	}
	shift := d.prevStart
	keep := len(d.buf) - shift
	if keep+n <= cap(d.buf) && !d.shared {
		copy(d.buf, d.buf[shift:])
		d.buf = d.buf[:keep]
	} else {
		b := make([]byte, keep, max(2*cap(d.buf), keep+n, 4096))
		copy(b, d.buf[shift:])
		d.buf, d.shared = b, false
	}
	d.base += int64(shift)
	d.pos -= shift
	d.prevStart -= shift
	d.prevEnd -= shift
	return shift
}

// skip returns the position of the first byte from pos which is not white space, reading more input as it
// needs. At the end of the input, it returns the error of the reader.
func (d *decoder) skip(pos int) (int, error) {
	for {
		if pos = skipSpace(d.buf, pos); pos < len(d.buf) {
			return pos, nil
		}
		shift, err := d.fetch()
		pos -= shift
		if err != nil {
			return pos, err
		}
	}
}

// failAt is a SyntacticError at the position pos of the buffer, in the value where points.
func (d *decoder) failAt(err error, pos int, where int) error {
	return &SyntacticError{ByteOffset: d.base + int64(pos), JSONPointer: d.st.pointer(where), Err: err}
}

// endError is the error at the end of the input at pos: io.EOF where a top-level value may start, or else an
// unexpected one. A reader error which is not io.EOF is returned as it is.
func (d *decoder) endError(pos int, err error, where int) error {
	if err != io.EOF {
		return err
	}
	if d.st.depth() == 0 && pos == skipSpace(d.buf, d.pos) && where != pointNext {
		return io.EOF
	}
	return d.failAt(io.ErrUnexpectedEOF, pos, where)
}

// PeekKind returns the kind of the next token, without reading it.
//
// It returns KindInvalid at an error, which it keeps for the next read call to report: a PeekKind call must
// be followed by a read call.
func (d *Decoder) PeekKind() Kind {
	s := &d.d
	if s.peekErr != nil {
		return KindInvalid
	}
	pos, err := s.beforeToken(true)
	if err != nil {
		if err != io.EOF {
			s.peekErr = err
		}
		return KindInvalid
	}
	return kindOf(s.buf[pos])
}

// beforeToken skips the white space and the delimiter before the next token, and returns its position. peek is
// set for PeekKind, which takes an end delimiter of any kind: the read which follows checks it.
//
// A comma or a colon is taken first, and then compared with the delimiter which the next token needs: one which
// is not needed, or not the one needed, is an invalid character itself.
func (d *decoder) beforeToken(peek bool) (int, error) {
	pos, err := d.skip(d.pos)
	if err != nil {
		return pos, d.endError(pos, err, pointAt)
	}
	found, at := byte(0), pos
	need := d.delimBefore(d.buf[pos], peek)
	if c := d.buf[pos]; c == ',' || c == ':' {
		found = c
		if pos, err = d.skip(pos + 1); err != nil {
			if err != io.EOF || found == need {
				return pos, d.endError(pos, err, pointAt)
			}
		} else {
			need = d.delimBefore(d.buf[pos], peek)
		}
	}
	switch {
	case found == need:
		return pos, nil
	case found == 0:
		return pos, d.missingDelim(pos)
	case need == 0:
		return at, d.failAt(invalidChar(d.buf[at:], "at start of value"), at, pointAt)
	default:
		// the delimiter of the other kind is reported as the character where the needed one is
		return at, d.missingDelim(at)
	}
}

// delimBefore is the delimiter which the next token, which starts with c, needs before it: a comma, a colon, or
// none.
func (d *decoder) delimBefore(c byte, peek bool) byte {
	l := d.st.last()
	switch {
	case d.st.depth() == 0:
		return 0
	case l.needValue():
		return ':'
	case l.count == 0:
		return 0
	case c == '}' && (l.object || peek), c == ']' && (!l.object || peek):
		return 0 // the end
	}
	return ','
}

// missingDelim is the error of the character at pos, where a delimiter is needed.
func (d *decoder) missingDelim(pos int) error {
	l := d.st.last()
	switch {
	case l.needValue():
		return d.failAt(invalidChar(d.buf[pos:], "after object name (expecting ':')"), pos, pointAt)
	case l.object:
		// the end of an array in an object is reported in the level which contains the object.
		where := pointAt
		if d.buf[pos] == ']' {
			where = pointOut
		}
		return d.failAt(invalidChar(d.buf[pos:], "after object value (expecting ',' or '}')"), pos, where)
	}
	return d.failAt(invalidChar(d.buf[pos:], "after array element (expecting ',' or ']')"), pos, pointAt)
}

// ReadToken reads the next Token and advances the read offset. The token is only valid until the next Peek,
// Read or Skip call. It returns io.EOF at the end of the input.
func (d *Decoder) ReadToken() (Token, error) {
	return d.d.readToken()
}

func (d *decoder) readToken() (Token, error) {
	if err := d.peekErr; err != nil {
		d.peekErr = nil
		return Token{}, err
	}
	pos, err := d.beforeToken(false)
	if err != nil {
		return Token{}, err
	}
	l := d.st.last()
	c := d.buf[pos]
	switch c {
	case '{', '[':
		if l.needName() {
			return Token{}, d.failAt(ErrNonStringName, pos, pointAt)
		}
		if err := d.st.push(c == '{'); err != nil {
			return Token{}, d.failAt(err, pos, pointNext)
		}
		d.done(pos, pos+1)
		if c == '{' {
			return BeginObject, nil
		}
		return BeginArray, nil
	case '}', ']':
		switch {
		case d.st.depth() == 0:
			return Token{}, d.failAt(invalidChar(d.buf[pos:], "at start of value"), pos, pointAt)
		case l.object != (c == '}'):
			if l.needValue() {
				return Token{}, d.failAt(invalidChar(d.buf[pos:], "after object value (expecting ',' or '}')"), pos, pointHere)
			}
			if l.object {
				return Token{}, d.failAt(invalidChar(d.buf[pos:], "at start of value"), pos, pointAt)
			}
			return Token{}, d.failAt(invalidChar(d.buf[pos:], "at start of value"), pos, pointNext)
		case l.needValue():
			return Token{}, d.failAt(errMissingValue, pos, pointAt)
		}
		d.st.pop()
		d.done(pos, pos+1)
		if c == '}' {
			return EndObject, nil
		}
		return EndArray, nil
	}
	name := l.needName()
	where := pointNext
	if name {
		where = pointAt
	}
	start, end, f, err := d.scanToken(pos, where)
	if err != nil {
		return Token{}, err
	}
	if name {
		if c != '"' {
			return Token{}, d.failAt(ErrNonStringName, start, pointAt)
		}
		if err := d.insertName(start, end, f); err != nil {
			return Token{}, err
		}
	}
	d.st.last().count++
	d.done(start, end)
	switch k := kindOf(c); k {
	case KindNull:
		return Null, nil
	case KindTrue:
		return True, nil
	case KindFalse:
		return False, nil
	default:
		return Token{dec: d, num: uint64(d.base + int64(start)), kind: k, form: formRaw}, nil
	}
}

// done records the token or value at buf[start:end] as the last one read.
func (d *decoder) done(start, end int) {
	d.prevStart, d.prevEnd, d.pos = start, end, end
}

// insertName adds the name of the string at buf[start:end] to the innermost object.
func (d *decoder) insertName(start, end int, f strFlags) error {
	s := d.buf[start:end]
	if f&(strEscaped|strInvalidUTF8) != 0 {
		d.vs.scratch = appendUnquoted(d.vs.scratch[:0], s)
	} else {
		d.vs.scratch = append(d.vs.scratch[:0], s[1:len(s)-1]...)
	}
	if !d.st.insertName(d.vs.scratch, !d.cfg.has(allowDuplicateNames)) {
		ptr := d.st.namePointer(d.vs.scratch)
		return &SyntacticError{ByteOffset: d.base + int64(start), JSONPointer: ptr, Err: ErrDuplicateName}
	}
	return nil
}

// scanToken scans the string, number or literal at pos, reading more input while the buffer cuts it. It
// returns the position of the token, which a read may have moved, and its end.
func (d *decoder) scanToken(pos int, where int) (int, int, strFlags, error) {
	var (
		n   int
		f   strFlags
		st  numState
		err error
	)
	c := d.buf[pos]
	validUTF8 := !d.cfg.has(allowInvalidUTF8)
	for {
		b := d.buf[pos:]
		switch c {
		case '"':
			if n > 0 {
				n, f, err = scanStringFrom(b, n, f, strModeOf(validUTF8, d.rerr == nil))
			} else {
				n, f, err = scanString(b, strModeOf(validUTF8, d.rerr == nil))
			}
		case 'n', 't', 'f':
			n, err = scanLiteral(b, literalOf(Kind(c)))
		default:
			if kindOf(c) != KindNumber {
				if (c == ',' || c == ':') && d.st.last().count == 0 {
					where = pointAt // a delimiter before the first value
				}
				return pos, pos, 0, d.failAt(invalidChar(b, "at start of value"), pos, where)
			}
			n, st, err = scanNumberFrom(b, n, st)
			if err == nil && n == len(b) && d.rerr == nil {
				err = io.ErrUnexpectedEOF // the number may continue
			}
		}
		if err != io.ErrUnexpectedEOF || d.rerr != nil {
			break
		}
		shift, ferr := d.fetch()
		pos -= shift
		if ferr != nil && ferr != io.EOF {
			return pos, pos, 0, ferr
		}
		if c == 'n' || c == 't' || c == 'f' {
			n = 0
		}
	}
	if err == nil {
		return pos, pos + n, f, nil
	}
	if err == io.ErrUnexpectedEOF && kindOf(c) == KindNumber {
		if d.rerr != io.EOF {
			return pos, pos, 0, d.rerr
		}
		n = 0 // a number which the input cuts is reported at its start
	} else if err == io.ErrUnexpectedEOF && d.rerr != io.EOF {
		return pos, pos, 0, d.rerr
	}
	return pos, pos, 0, d.failAt(err, pos+n, where)
}

// ReadValue returns the next raw JSON value and advances the read offset. The value has no leading and
// trailing white space, and holds the exact bytes of the input, which may contain invalid UTF-8 if
// AllowInvalidUTF8 is set.
//
// The value is only valid until the next Peek, Read or Skip call, and must not be changed while the decoder is
// used. If the next token is the end of an object or array, it reports a SyntacticError and doesn't change the
// state of the decoder. It returns io.EOF at the end of the input.
func (d *Decoder) ReadValue() (Value, error) {
	s := &d.d
	start, end, err := s.readValue()
	if err != nil {
		return nil, err
	}
	return s.buf[start:end], nil
}

// SkipValue is ReadValue, whose value is dropped.
func (d *Decoder) SkipValue() error {
	_, _, err := d.d.readValue()
	return err
}

func (d *decoder) readValue() (int, int, error) {
	if err := d.peekErr; err != nil {
		d.peekErr = nil
		return 0, 0, err
	}
	pos, err := d.beforeToken(false)
	if err != nil {
		return 0, 0, err
	}
	l := d.st.last()
	switch c := d.buf[pos]; {
	case c == '}' || c == ']':
		if d.st.depth() == 0 || l.object != (c == '}') {
			return 0, 0, d.failAt(invalidChar(d.buf[pos:], "at start of value"), pos, pointNext)
		}
		return 0, 0, d.failAt(invalidChar(d.buf[pos:], "at start of value"), pos, pointAt)
	case l.needName():
		if c != '"' {
			if kindOf(c) != KindInvalid {
				return 0, 0, d.failAt(ErrNonStringName, pos, pointAt)
			}
			return 0, 0, d.failAt(invalidChar(d.buf[pos:], "at start of value"), pos, pointAt)
		}
		start, end, f, err := d.scanToken(pos, pointAt)
		if err != nil {
			return 0, 0, err
		}
		if err := d.insertName(start, end, f); err != nil {
			return 0, 0, err
		}
		l.count++
		d.done(start, end)
		return start, end, nil
	}
	count := l.count
	rp := resumePoint{pos: pos, base: d.st.depth()}
	start := pos
	for {
		d.vs.final = d.rerr != nil
		end, serr := d.vs.resume(d.buf, &rp)
		if serr == nil {
			d.done(start, end)
			return start, end, nil
		}
		if serr != errCut {
			d.vs.restore(rp.base, count)
			return 0, 0, &SyntacticError{ByteOffset: d.base + int64(serr.pos), JSONPointer: serr.ptr, Err: serr.err}
		}
		shift, ferr := d.fetch()
		rp.pos -= shift
		start -= shift
		if ferr != nil && ferr != io.EOF {
			d.vs.restore(rp.base, count)
			return 0, 0, ferr
		}
	}
}

// InputOffset returns the offset in the input after the last token or value which was read. The decoder may
// have read more of the input into its buffer.
func (d *Decoder) InputOffset() int64 {
	return d.d.base + int64(d.d.pos)
}

// UnreadBuffer returns the input in the buffer of the decoder which was read from the reader but not by
// Decoder.ReadToken or Decoder.ReadValue: it may be any number of bytes, which may not be valid JSON. How much
// input the decoder buffers may change over time.
//
// The input after the last token or value is this buffer followed by the rest of the reader.
//
// The buffer must not be changed while the decoder is used, and is valid until the next Peek, Read or Skip
// call.
func (d *Decoder) UnreadBuffer() []byte {
	return d.d.buf[d.d.pos:]
}

// StackDepth returns the number of the objects and arrays which are open in the input which was read: 0 at the
// top level, before any token, after a top-level value and between top-level values, 1 in a top-level object
// or array, and so on.
func (d *Decoder) StackDepth() int {
	return d.d.st.depth()
}

// StackIndex returns the kind and the length of the level i of the stack, which is from 0 to
// Decoder.StackDepth: KindInvalid for the level 0, KindBeginObject for an object, KindBeginArray for an array.
// The length of an object counts its names and values: a complete object has an even length.
func (d *Decoder) StackIndex(i int) (Kind, int64) {
	return d.d.st.index(i)
}

// StackPointer returns a JSON Pointer (RFC 6901) to the last value which was read.
func (d *Decoder) StackPointer() Pointer {
	return d.d.st.pointer(pointLast)
}

// index is StackIndex of the stack.
func (s *stack) index(i int) (Kind, int64) {
	l := s.levels[i]
	switch {
	case i == 0:
		return KindInvalid, l.count
	case l.object:
		return KindBeginObject, l.count
	}
	return KindBeginArray, l.count
}

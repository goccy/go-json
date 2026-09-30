package jsontext

import (
	"bytes"
	"io"
	"slices"
	"unicode/utf16"
	"unicode/utf8"
)

// valueScanner reads a JSON value, and writes it formatted if write is set. The levels of the value are pushed
// on the stack of the encoder or decoder while the value is read, so that the pointers of the errors are the
// ones of the stream.
//
// A decoder which reads its input by parts scans a value which the end of its buffer cuts up to there, and
// continues the scan from a resumePoint after it read more.
type valueScanner struct {
	st *stack
	// the options
	validUTF8  bool
	checkNames bool
	decoding   bool // the messages and the offsets of the errors of a decoder
	final      bool // the end of the input is the end of the buffer

	// the output, when the value is written
	write     bool
	out       []byte
	ws        whitespace
	esc       escapeFlags
	preserve  bool
	canonInts bool
	canonFlts bool
	reorder   bool

	// the members of the objects which are reordered, by level
	members []memberList
	scratch []byte
}

// The places in a value where a scan continues.
const (
	atValue  uint8 = iota
	atNext         // after a value: a comma or an end delimiter
	atName         // an object name
	atColon        // the colon after an object name
	atOpened       // after a begin delimiter: a value, or the end delimiter
)

// resumePoint is where a scan which the end of the buffer cut continues: the place, and the progress in a
// string or number which was cut.
type resumePoint struct {
	at   uint8
	pos  int // the start of the token which was cut, or the position of the place
	tok  int // the length of the part of the token which was taken, 0 if none
	str  strFlags
	num  numState
	base int // the depth of the stack before the value
}

// scanError is an error of the value at an offset, with the pointer where it was found.
type scanError struct {
	err error
	pos int
	ptr Pointer
}

func (s *valueScanner) fail(err error, pos int, where int) *scanError {
	return &scanError{err: err, pos: pos, ptr: s.st.pointer(where)}
}

// errCut reports a scan which the end of the buffer cut, which continues from the resume point.
var errCut = &scanError{err: io.ErrUnexpectedEOF}

// scan reads the value at b[i:], and returns the position after it, and counts it in the innermost level of the
// stack. At an error, the stack is left as it was.
func (s *valueScanner) scan(b []byte, i int) (int, *scanError) {
	count := s.st.last().count
	rp := resumePoint{pos: i, base: s.st.depth()}
	end, err := s.resume(b, &rp)
	if err != nil {
		s.restore(rp.base, count)
	}
	return end, err
}

// restore pops the levels of a value which was not read to its end.
func (s *valueScanner) restore(base int, count int64) {
	for s.st.depth() > base {
		s.st.pop()
	}
	s.st.last().count = count
	s.members = s.members[:0]
}

// resume scans the value from the resume point. If the end of the buffer cuts it and the input continues, it
// returns errCut, with the resume point where the scan continues.
//
//nolint:maintidx // one loop of gotos over the grammar, whose state stays in its variables
func (s *valueScanner) resume(b []byte, rp *resumePoint) (int, *scanError) {
	var err error
	var n int
	i := rp.pos
	base := rp.base
	switch rp.at {
	case atNext:
		goto next
	case atName:
		goto name
	case atColon:
		goto colon
	case atOpened:
		goto opened
	}

value:
	if i = skipSpace(b, i); i == len(b) {
		return s.cut(atValue, i, pointAt, rp)
	}
	switch c := b[i]; c {
	case '{', '[':
		if err := s.st.push(c == '{'); err != nil {
			return i, s.fail(err, i, pointNext)
		}
		if s.write {
			s.out = append(s.out, c)
			if c == '{' && s.reorder {
				s.members = append(s.members, memberList{start: len(s.out)})
			}
		}
		i++
		goto opened
	case '"':
		var f strFlags
		if rp.tok > 0 {
			n, f, err = scanStringFrom(b[i:], rp.tok, rp.str, strModeOf(s.validUTF8, !s.final))
			rp.tok = 0
		} else {
			n, f, err = scanString(b[i:], strModeOf(s.validUTF8, !s.final))
		}
		if err != nil {
			if err == io.ErrUnexpectedEOF && !s.final {
				rp.tok, rp.str = n, f
				return s.cut(atValue, i, pointNext, rp)
			}
			return i + n, s.fail(err, i+n, pointNext)
		}
		if s.write {
			s.out = appendRawString(s.out, b[i:i+n], f, s.esc, s.preserve)
		}
		i += n
	case 'n', 't', 'f':
		lit := literalOf(Kind(c))
		n, err = scanLiteral(b[i:], lit)
		if err != nil {
			if err == io.ErrUnexpectedEOF && !s.final {
				return s.cut(atValue, i, pointNext, rp)
			}
			return i + n, s.fail(err, i+n, pointNext)
		}
		if s.write {
			s.out = append(s.out, lit...)
		}
		i += n
	default:
		if kindOf(c) != KindNumber {
			if l := &s.st.levels[base]; s.decoding && base > 0 && l.count > 1 && (c == ']' && l.object || c == '}' && !l.object) {
				// encoding/json/jsontext reports an end delimiter where a value is, in a value which a decoder
				// reads in an object or array, as the error of the object or array where the value is, when it
				// is the end of the other kind.
				if l.object {
					return i, s.fail(invalidChar(b[i:], "after object value (expecting ',' or '}')"), i, pointHere)
				}
				return i, s.fail(invalidChar(b[i:], "after array element (expecting ',' or ']')"), i, pointHere)
			}
			return i, s.fail(invalidChar(b[i:], "at start of value"), i, pointNext)
		}
		var st numState
		if rp.tok > 0 {
			n, st, err = scanNumberFrom(b[i:], rp.tok, rp.num)
			rp.tok = 0
		} else {
			n, st, err = scanNumberFrom(b[i:], 0, numState{})
		}
		if i+n == len(b) && !s.final && (err == nil || err == io.ErrUnexpectedEOF) {
			rp.tok, rp.num = n, st
			return s.cut(atValue, i, pointNext, rp)
		}
		if err != nil {
			// a decoder reports a number which the input cuts at its start.
			if err != io.ErrUnexpectedEOF {
				// the position of the invalid character
			} else if s.decoding {
				n = 0
			} else {
				n = st.done
			}
			return i + n, s.fail(err, i+n, pointNext)
		}
		if s.write {
			s.out = s.appendNumber(s.out, b[i:i+n])
		}
		i += n
	}
	s.st.last().count++

next:
	if s.st.depth() == base {
		return i, nil
	}
	if i = skipSpace(b, i); i == len(b) {
		return s.cut(atNext, i, pointAt, rp)
	}
	if l := s.st.last(); l.object {
		switch b[i] {
		case ',':
			i++
			if s.write {
				if s.reorder {
					s.members[len(s.members)-1].endMember(len(s.out))
				}
				s.out = s.ws.appendComma(s.out)
			}
			goto name
		case '}':
			i++
			s.closeLevel('}', false)
			goto next
		}
		return i, s.fail(invalidChar(b[i:], "after object value (expecting ',' or '}')"), i, pointAt)
	}
	switch b[i] {
	case ',':
		i++
		if s.write {
			s.out = s.ws.appendComma(s.out)
			s.newLine(0)
		}
		goto value
	case ']':
		i++
		s.closeLevel(']', false)
		goto next
	}
	if s.decoding {
		return i, s.fail(invalidChar(b[i:], "after array element (expecting ',' or ']')"), i, pointAt)
	}
	return i, s.fail(invalidChar(b[i:], "after array value (expecting ',' or ']')"), i, pointAt)

opened:
	if i = skipSpace(b, i); i == len(b) {
		return s.cut(atOpened, i, pointAt, rp)
	}
	if l := s.st.last(); b[i] == '}' && l.object || b[i] == ']' && !l.object {
		i++
		s.closeLevel(b[i-1], true)
		goto next
	} else if l.object {
		goto name
	}
	if s.write {
		s.newLine(0)
	}
	goto value

name:
	if i = skipSpace(b, i); i == len(b) {
		return s.cut(atName, i, pointAt, rp)
	}
	if b[i] != '"' {
		return i, s.fail(invalidChar(b[i:], "at start of string (expecting '\"')"), i, pointAt)
	}
	{
		var f strFlags
		if rp.tok > 0 {
			n, f, err = scanStringFrom(b[i:], rp.tok, rp.str, strModeOf(s.validUTF8, !s.final))
			rp.tok = 0
		} else {
			n, f, err = scanString(b[i:], strModeOf(s.validUTF8, !s.final))
		}
		if err != nil {
			if err == io.ErrUnexpectedEOF && !s.final {
				rp.tok, rp.str = n, f
				return s.cut(atName, i, pointAt, rp)
			}
			return i + n, s.fail(err, i+n, pointAt)
		}
		name := b[i : i+n]
		if f&(strEscaped|strInvalidUTF8) != 0 {
			s.scratch = appendUnquoted(s.scratch[:0], name)
		} else {
			s.scratch = append(s.scratch[:0], name[1:n-1]...)
		}
		if !s.st.insertName(s.scratch, s.checkNames) {
			return i, &scanError{err: ErrDuplicateName, pos: i, ptr: s.st.namePointer(s.scratch)}
		}
		s.st.last().count++
		if s.write {
			s.newLine(0)
			if s.reorder {
				s.members[len(s.members)-1].startMember(len(s.out), s.scratch)
			}
			s.out = appendRawString(s.out, name, f, s.esc, s.preserve)
		}
		i += n
	}

colon:
	if i = skipSpace(b, i); i == len(b) {
		return s.cut(atColon, i, pointAt, rp)
	}
	if b[i] != ':' {
		return i, s.fail(invalidChar(b[i:], "after object name (expecting ':')"), i, pointAt)
	}
	i++
	if s.write {
		s.out = append(s.out, ':')
		if s.ws.colon {
			s.out = append(s.out, ' ')
		}
	}
	goto value
}

// cut is the end of the scan at the end of the buffer, at the place at, where the token or the place starts at
// i: the input continues, or the value is not complete.
func (s *valueScanner) cut(at uint8, i int, where int, rp *resumePoint) (int, *scanError) {
	if s.final {
		pos := i + rp.tok
		if rp.tok == 0 {
			pos = i
		}
		return pos, s.fail(io.ErrUnexpectedEOF, pos, where)
	}
	rp.at, rp.pos = at, i
	return i, errCut
}

// closeLevel pops the innermost object or array, and writes its end delimiter.
func (s *valueScanner) closeLevel(end byte, empty bool) {
	if s.write {
		if end == '}' && s.reorder {
			m := &s.members[len(s.members)-1]
			m.endMember(len(s.out))
			s.out = m.reorder(s.out)
			s.members = s.members[:len(s.members)-1]
		}
		if !empty {
			s.newLine(-1)
		}
		s.out = append(s.out, end)
	}
	s.st.pop()
}

// newLine starts a new line of a multiline output at the depth of the stack plus delta.
func (s *valueScanner) newLine(delta int) {
	if s.ws.multiline {
		s.out = s.ws.appendLine(s.out, s.st.depth()+delta)
	}
}

// appendComma appends a comma, and the space after it if the options want it.
func (w *whitespace) appendComma(b []byte) []byte {
	if w.comma {
		return append(b, ',', ' ')
	}
	return append(b, ',')
}

// appendLine appends a line feed, the prefix and the indentation of the depth.
func (w *whitespace) appendLine(b []byte, depth int) []byte {
	b = append(b, '\n')
	b = append(b, w.prefix...)
	for range depth {
		b = append(b, w.indent...)
	}
	return b
}

// appendNumber appends the number b, canonicalized if the options say so.
func (s *valueScanner) appendNumber(dst, b []byte) []byte {
	if string(b) == "-0" && (s.canonInts || s.canonFlts) {
		return append(dst, '0') // -0 is 0 under either option
	}
	if isInteger(b) {
		if s.canonInts {
			return appendCanonicalNumber(dst, b)
		}
	} else if s.canonFlts {
		return appendCanonicalNumber(dst, b)
	}
	return append(dst, b...)
}

// memberList are the members of an object which ReorderRawObjects reorders: the output of each member, from the
// start of its name to the end of its value.
type memberList struct {
	start   int // the start of the output of the members
	members []member
}

type member struct {
	name       string // unquoted
	start, end int
}

func (m *memberList) startMember(pos int, name []byte) {
	m.members = append(m.members, member{name: string(name), start: pos, end: -1})
}

func (m *memberList) endMember(pos int) {
	if k := len(m.members); k > 0 && m.members[k-1].end < 0 {
		m.members[k-1].end = pos
	}
}

// reorder sorts the members in out, which are separated by commas and white space, by their names.
func (m *memberList) reorder(out []byte) []byte {
	// members of the same name, or of names which are the same once invalid UTF-8 is mangled, are ordered by
	// their output.
	compare := func(a, b member) int {
		if c := compareMembers(a, b); c != 0 {
			return c
		}
		return bytes.Compare(out[a.start:a.end], out[b.start:b.end])
	}
	if len(m.members) < 2 || slices.IsSortedFunc(m.members, compare) {
		return out
	}
	// the separator of the members, from the end of the first one to the start of the second one.
	sep := string(out[m.members[0].end:m.members[1].start])
	lead := string(out[m.start:m.members[0].start])
	tail := string(out[m.members[len(m.members)-1].end:])
	ordered := slices.Clone(m.members)
	slices.SortFunc(ordered, compare)
	buf := []byte(lead)
	for i, x := range ordered {
		if i > 0 {
			buf = append(buf, sep...)
		}
		buf = append(buf, out[x.start:x.end]...)
	}
	buf = append(buf, tail...)
	return append(out[:m.start], buf...)
}

// compareMembers orders names by their UTF-16 code units, as RFC 8785, section 3.2.3 does.
func compareMembers(a, b member) int {
	x, y := a.name, b.name
	for len(x) > 0 && len(y) > 0 {
		rx, nx := utf8.DecodeRuneInString(x)
		ry, ny := utf8.DecodeRuneInString(y)
		if rx != ry {
			ux, uy := utf16Units(rx), utf16Units(ry)
			if ux[0] != uy[0] {
				return int(ux[0]) - int(uy[0])
			}
			return int(ux[1]) - int(uy[1])
		}
		x, y = x[nx:], y[ny:]
	}
	return len(x) - len(y)
}

// utf16Units are the UTF-16 code units of r, the second one 0 if it has one.
func utf16Units(r rune) [2]uint16 {
	if r1, r2 := utf16.EncodeRune(r); r1 != utf8.RuneError {
		return [2]uint16{uint16(r1), uint16(r2)}
	}
	return [2]uint16{uint16(r), 0}
}

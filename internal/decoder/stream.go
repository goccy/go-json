package decoder

import (
	"bytes"
	"io"
	"math"
	"math/bits"
	"unicode/utf8"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonnum"
	"github.com/goccy/go-json/internal/runtime"
)

const (
	// initBufSize is the size of the buffer of a new stream.
	initBufSize = 512
	// bufPadding is the number of the bytes kept free after the data in the buffer of a stream:
	// one for the nul byte which ends the value being decoded, the rest so that a scan reads
	// a whole block from any byte of the value.
	bufPadding = scanBlockSize
)

// Stream reads the values of a JSON text from an io.Reader one by one.
//
// A value is decoded by the same decoders as a byte slice is: the stream reads until its buffer
// holds the whole value ( scanValue ), puts a nul byte after the value and lets the decoder work
// on the buffer. The decoded strings refer to the buffer, so a byte which has been read is never
// overwritten: when the buffer is full, the bytes which are not consumed yet are copied to a new one.
type Stream struct {
	// buf[:length] is what has been read, buf[length] is nul, and bufPadding bytes are free at the end.
	buf    []byte
	length int64
	// cursor is the first byte which is not consumed yet.
	cursor int64
	// offset is the number of the bytes discarded before buf[0]: the total offset is offset + cursor.
	offset int64
	r      io.Reader
	// readErr is the error the reader returned. Once set, the reader is never read again:
	// the data read before the error is consumed first, then the error is reported.
	readErr error
	ctx     *RuntimeContext
	Option  *Option
	// prevEnd is the total offset of the end of the previous value, which the offset of a type error may be
	// relative to ( see StreamOffsetBase ).
	prevEnd int64
	// tokenState is where the stream is between the tokens, and tokenStack the states of the objects and the
	// arrays which Token opened, the innermost last ( see Token ).
	tokenState tokenState
	tokenStack []tokenState
	// err is the error which a read failed with, which every read returns after it, as encoding/json does, and
	// errOffset the total offset which InputOffset reports after it: the end of the previous value.
	err       error
	errOffset int64
	// readStart is the total offset where the read of the value of Decode, or of the value of Token, started, and
	// scanned the number of the bytes which these reads read before it, by which encoding/json before Go 1.27
	// reports the offset of a syntax error of the value ( see scannedOffset ).
	readStart int64
	scanned   int64
	// more is what More peeked at, which InputOffset reports ( see streamPeek ).
	more streamPeek
	// keeps is set while the bytes from the total offset keepFrom are kept in the buffer, before the cursor too:
	// a read of a token which fails leaves the stream where it was ( see Token ).
	keeps    bool
	keepFrom int64
	// index is where the objects and the arrays of the value which is decoded element by element end, which the
	// lookups extend ( see indexAt ). It is nil while no value is decoded so, and else elements, whose entries are
	// kept for the next value, which allocates them no more.
	index, elements *bracketIndex
	// filled is set when the last read filled the room of the buffer: the reader may have more, which a larger
	// buffer takes at once ( see grow ). Only what a read returned tells it.
	filled bool
}

func NewStream(r io.Reader) *Stream {
	opt := &Option{}
	return &Stream{
		r:      r,
		buf:    make([]byte, initBufSize+bufPadding),
		ctx:    &RuntimeContext{Option: opt},
		Option: opt,
	}
}

// markPrevEnd keeps the end of the previous value of the stream: the offsets of the type errors of the next value
// may be relative to it ( see StreamOffsetBase ), and InputOffset reports it after a read which fails.
func (s *Stream) markPrevEnd() {
	s.prevEnd = s.offset + s.cursor
}

// fail keeps err, the error which a read failed with, which every read returns after it, as encoding/json does.
//
//go:noinline
func (s *Stream) fail(err error) error {
	s.err, s.errOffset, s.tokenState = err, s.prevEnd, tokenFailed
	return err
}

// TotalOffset returns the number of the bytes consumed from the reader.
func (s *Stream) TotalOffset() int64 {
	return s.offset + s.cursor
}

// Buffered returns a reader of the data read from the reader but not consumed yet.
func (s *Stream) Buffered() io.Reader {
	return bytes.NewReader(s.buf[s.cursor:s.length])
}

// read reads more bytes from the reader into the buffer.
// It returns false when nothing more can be read: the reader ended or failed ( readErr ).
func (s *Stream) read() bool {
	if s.readErr != nil {
		return false
	}
	if int64(len(s.buf))-s.length <= bufPadding {
		s.grow()
	}
	room := int64(len(s.buf)) - bufPadding - s.length
	n, err := s.r.Read(s.buf[s.length : s.length+room])
	s.filled = int64(n) == room
	s.length += int64(n)
	s.buf[s.length] = nul
	if err != nil {
		s.readErr = err
	}
	return n > 0 || err == nil
}

// grow moves the bytes which are not consumed yet to a new buffer, so that the reader
// has room to read into. The old buffer is left to the strings which refer to it.
func (s *Stream) grow() {
	// the bytes from the cursor are moved, or from the one which is kept before it
	from := s.cursor
	if keep := s.keepFrom - s.offset; s.keeps && keep >= 0 && keep < from {
		from = keep
	}
	remain := s.length - from
	size := int64(len(s.buf)) - bufPadding
	if remain*2 > size {
		// The data which is not consumed yet fills more than half of the buffer: it is a value
		// larger than the buffer, which is grown faster so that it is copied fewer times.
		size *= 4
	}
	if r, ok := s.r.(interface{ Len() int }); ok {
		// The reader knows how much is left ( bytes.Reader, strings.Reader, bytes.Buffer ):
		// the buffer holds all of it at once, so that nothing is copied again.
		size = remain + int64(r.Len())
	}
	buf := s.buf
	x := s.index
	if x != nil && s.filled {
		// the buffer of a value decoded element by element grows four times as large up to elementReadSize while the
		// reader fills it, so that a value is read by large reads, which bring what the reader has, without waiting
		// for more: a reader which brings less is read into the same buffer
		size = max(size, min(4*(int64(len(buf))-bufPadding), elementReadSize))
	}
	if x == nil || !x.owns(buf) || size > int64(len(buf))-bufPadding || (s.Option.Flags&NoCopyStringOption) != 0 {
		buf = make([]byte, size+bufPadding)
		if x != nil {
			// whose strings are copied, which nothing refers to but them: it is reused in place
			x.buf = unsafe.SliceData(buf)
		}
	}
	copy(buf, s.buf[from:s.length])
	s.offset += from
	s.buf = buf
	s.length = remain
	s.cursor -= from
}

// fill makes sure that the buffer holds the byte at the cursor.
func (s *Stream) fill() bool {
	for s.cursor >= s.length {
		if !s.read() {
			return false
		}
	}
	return true
}

// endError is the error reported when the reader has nothing more: io.EOF or what the reader returned.
func (s *Stream) endError() error {
	if s.readErr != nil {
		return s.readErr
	}
	return io.EOF
}

// truncatedValueError is the error reported when the reader ends in the middle of the value at the cursor: the first
// syntax error of what has been read, as encoding/json reads a value byte by byte, or the unexpected end.
//
//go:noinline
func (s *Stream) truncatedValueError() error {
	return s.truncatedValueErrorAt(0)
}

// truncatedValueErrorAt is truncatedValueError for a value nested in depth objects and arrays.
func (s *Stream) truncatedValueErrorAt(depth int64) error {
	if err := s.valueSyntaxErrorAt(depth); err != nil {
		return err
	}
	return s.unexpectedEndError()
}

// unexpectedEndError is the error reported when the reader ends in the middle of a value.
func (s *Stream) unexpectedEndError() error {
	if s.readErr != nil && s.readErr != io.EOF {
		return s.readErr
	}
	return io.ErrUnexpectedEOF
}

// skipWhiteSpace moves the cursor to the next byte which is not white space.
// It returns false when the reader has nothing more.
func (s *Stream) skipWhiteSpace() bool {
	// a byte which is not white space, as the tokens of most values are not separated by any, without a call
	if s.cursor < s.length && !isWhiteSpace[s.buf[s.cursor]] {
		return true
	}
	return s.skipWhiteSpaceRest()
}

// skipWhiteSpaceRest is skipWhiteSpace without its check of the byte at the cursor first.
func (s *Stream) skipWhiteSpaceRest() bool {
	for {
		// the nul byte at the end of what has been read stops the skip
		if s.cursor = skipWhiteSpace(s.buf, s.cursor); s.cursor < s.length {
			return true
		}
		if !s.read() {
			return false
		}
	}
}

// peek moves the cursor to the next byte which is not a white space, as skipWhiteSpace does, but leaves it where it
// is when the input ends with white spaces only: the offset of the input doesn't count them, as encoding/json
// doesn't count what it peeks at. It returns false when the reader has nothing more.
func (s *Stream) peek() bool {
	var rel int64
	for {
		// the position is kept relative to the cursor, because a read may move the data ( grow )
		for pos := s.cursor + rel; pos < s.length; pos++ {
			if !isWhiteSpace[s.buf[pos]] {
				s.cursor = pos
				return true
			}
		}
		rel = s.length - s.cursor
		if !s.read() {
			return false
		}
	}
}

// prepare moves the cursor to the beginning of the next value: the white space is skipped, and, between the
// tokens of an object or an array, the comma or the colon which the grammar has before the value.
func (s *Stream) prepare() error {
	if s.tokenState != tokenTopValue {
		return s.prepareInTokens()
	}
	s.readStart = s.offset + s.cursor
	// the values of a stream are separated by white space mostly
	if !s.skipWhiteSpaceRest() {
		return s.fail(s.endError())
	}
	return nil
}

// isLiteralChar is true for the bytes of a number, true, false and null.
var isLiteralChar = [256]bool{
	'0': true, '1': true, '2': true, '3': true, '4': true, '5': true, '6': true, '7': true, '8': true, '9': true,
	'-': true, '+': true, '.': true, 'e': true, 'E': true,
	't': true, 'r': true, 'u': true, 'f': true, 'a': true, 'l': true, 's': true, 'n': true,
}

// scanValue finds the end of the value which begins at the cursor, reading from the reader as needed,
// and returns the position after the value. The value is not validated: the decoder does it.
// The cursor is not moved: the value is in buf[cursor:end], but the buffer may have been replaced.
func (s *Stream) scanValue() (int64, error) {
	return s.scanValueUpTo(math.MaxInt64)
}

// scanValueUpTo is scanValue which reads for an object or an array only until limit bytes of it have been read: if it
// doesn't end within them, it returns errLongValue, and the value is decoded element by element ( see scanFailed ).
// A string, a number, true, false and null are read whole.
func (s *Stream) scanValueUpTo(limit int64) (int64, error) {
	if !s.fill() {
		return 0, s.endError()
	}
	switch c := s.buf[s.cursor]; {
	case c == '{' || c == '[':
		return s.scanCompound(limit)
	case c == '"':
		return s.scanString()
	case isLiteralChar[c]:
		return s.scanLiteral()
	}
	// The decoder reports the invalid character.
	return s.cursor + 1, nil
}

// scanCompound finds the end of the object or the array at the cursor: the position after the bracket which closes
// it. If it doesn't end within limit bytes, it returns errLongValue.
func (s *Stream) scanCompound(limit int64) (int64, error) {
	// The position is kept relative to the cursor, because a read may move the data ( grow ).
	var rel int64
	sc := compoundScanner{maxDepth: maxDecodeNestingDepth}
	for {
		inString, escaped := sc.inString, sc.escaped
		pos, found, err := sc.scan(s.buf, s.cursor+rel, s.length)
		if err != nil {
			return 0, s.totalOffsetError(err)
		}
		if found {
			return pos, nil
		}
		// a reader which tells its length has the input in memory, which is read whole at once ( see grow ), without
		// a wait
		if _, inMemory := s.r.(interface{ Len() int }); !inMemory {
			if pos-s.cursor >= limit {
				return 0, errLongValue
			}
			// a value which doesn't end in what has been read ends at a control character in a string of it, which a
			// string can't have: the decoder reports it, and the input after it is not read
			if c := controlInString(s.buf, s.cursor+rel, pos, inString, escaped); c >= 0 {
				return c + 1, nil
			}
		}
		rel = pos - s.cursor
		if !s.read() {
			return 0, s.truncatedValueError()
		}
	}
}

// controlInString returns the position of the first control character in a string in buf[pos:end], which a string
// can't have, or -1. The bytes are in a string from pos if inString is set, and escaped if escaped is set. The control
// characters are found word by word, and only the bytes before one are looked at for the strings: most values have
// none, or line breaks between their tokens.
func controlInString(buf []byte, pos, end int64, inString, escaped bool) int64 {
	for {
		q := pos
		for q+8 <= end {
			if w := load64(buf, q); (w-0x20*lsb)&^w&msb != 0 {
				break
			}
			q += 8
		}
		for q < end && buf[q] >= ' ' {
			q++
		}
		if q >= end {
			return -1
		}
		if inString, escaped = stringState(buf, pos, q, inString, escaped); inString {
			return q
		}
		pos = q + 1
	}
}

// stringState returns whether the byte at end is in a string, and escaped, from the state at pos.
func stringState(buf []byte, pos, end int64, inString, escaped bool) (bool, bool) {
	for ; pos < end; pos++ {
		switch c := buf[pos]; {
		case escaped:
			escaped = false
		case inString:
			if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
		case c == '"':
			inString = true
		}
	}
	return inString, escaped
}

// scanString finds the end of the string at the cursor: the position after the quote which closes it, or after a
// control character or the byte of an escape which is not one, which a string can't have: the decoder reports it, and
// the input after it is not read.
func (s *Stream) scanString() (int64, error) {
	end, _, _, err := s.scanStringInfo()
	return end, err
}

// scanStringInfo is scanString which also returns what the string has, as stringDecoder.scanString does, and whether
// it ends at its quote, and is valid then.
func (s *Stream) scanStringInfo() (int64, stringInfo, bool, error) {
	var (
		rel     int64 = 1 // after the opening quote
		escaped bool
		// hex is the number of the hexadecimal digits of a \u escape which are not read yet
		hex int
		// high has the high bits of the bytes, and firstEscape is the offset of the first backslash in the string
		high        uint64
		firstEscape int64 = -1
	)
	for {
		buf := s.buf
		pos := s.cursor + rel
		end := s.length
		// the words of the bytes up to a quote, a backslash or a control character, which the nul byte at the end of
		// what has been read is
		if !escaped && hex == 0 {
			for n := int64(len(buf)); pos+8 <= n; pos += 8 {
				w := load64(buf, pos)
				if stop := keyEndBytes(w); stop != 0 {
					// a quote, a backslash or a control character
					i := bits.TrailingZeros64(stop) / 8
					high |= w & (1<<(uint(i)*8&63) - 1)
					pos += int64(i)
					break
				}
				high |= w
			}
			if pos < end && buf[pos] == '"' {
				return pos + 1, stringInfo{firstEscape: int(firstEscape), nonASCII: high&msb != 0}, true, nil
			}
		}
		lim := min(pos+8, end)
		for pos < lim {
			c := buf[pos]
			high |= uint64(c)
			pos++
			switch {
			case c < ' ':
				return pos, stringInfo{}, false, nil
			case hex > 0:
				// an escape which is not one ends the string, so that the decoder reports it
				if !isHexDigit(c) {
					return pos, stringInfo{}, false, nil
				}
				hex--
			case escaped:
				escaped = false
				switch c {
				case 'u':
					hex = 4
				case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				default:
					return pos, stringInfo{}, false, nil
				}
			case c == '\\':
				escaped = true
				if firstEscape < 0 {
					firstEscape = pos - 1 - (s.cursor + 1)
				}
			case c == '"':
				return pos, stringInfo{firstEscape: int(firstEscape), nonASCII: high&msb != 0}, true, nil
			}
		}
		rel = pos - s.cursor
		if pos < end {
			continue
		}
		if !s.read() {
			return 0, stringInfo{}, false, s.truncatedValueError()
		}
	}
}

// scanLiteral finds the end of the number, true, false or null at the cursor: the position after true, false or null,
// or of the first byte which differs from it, and the position of the first byte which can't belong to the number by
// its grammar, or the end of the input. A number is read until a byte ends it, as encoding/json reads it, but a byte
// which makes it invalid ends it without a read more. The bytes of a number may be more than one value, as 01 is 0 and
// 1: Decode then decodes the first alone ( see decodeError ).
func (s *Stream) scanLiteral() (int64, error) {
	if lit := literalOf(s.buf[s.cursor]); lit != "" {
		for i := 1; i < len(lit); i++ {
			// a read may bring nothing, and the next one more
			for s.cursor+int64(i) >= s.length {
				if !s.read() {
					if s.readErr != io.EOF {
						return 0, s.readErr
					}
					return s.length, nil
				}
			}
			if s.buf[s.cursor+int64(i)] != lit[i] {
				return s.cursor + int64(i) + 1, nil
			}
		}
		return s.cursor + int64(len(lit)), nil
	}
	for {
		n, place := jsonnum.Scan(s.buf[s.cursor:s.length])
		if end := s.cursor + int64(n); end < s.length {
			if place == jsonnum.Valid {
				return end, nil
			}
			// the byte where the number is invalid
			return end + 1, nil
		}
		if !s.read() {
			if s.readErr != io.EOF {
				return 0, s.readErr
			}
			return s.length, nil
		}
	}
}

// literalOf returns true, false or null, which c starts, or "".
func literalOf(c byte) string {
	switch c {
	case 't':
		return "true"
	case 'f':
		return "false"
	case 'n':
		return "null"
	}
	return ""
}

// DecoderOf returns the decoder of the type, from the recent decoders of the context of the stream.
func (s *Stream) DecoderOf(typ unsafe.Pointer) (Decoder, error) {
	return s.ctx.DecoderOf(typ)
}

// Decode decodes the next value of the stream into p, of the pointer type typ, by dec. If the value has a type
// error, the stream goes on after the value, as the one of encoding/json does.
func (s *Stream) Decode(dec Decoder, typ unsafe.Pointer, p unsafe.Pointer) error {
	s.markPrevEnd()
	if err := s.prepare(); err != nil {
		return err
	}
	// a value which the buffer doesn't hold whole is decoded element by element, without a read for it
	end, err := s.scanValueUpTo(0)
	if err != nil {
		return s.scanFailed(err, dec, typ, p)
	}
	ctx := s.ctx
	ctx.Option = s.Option
	var cursor int64
	if c := s.buf[s.cursor]; c != 't' && c != 'f' && c != 'n' {
		// The end of the value is known exactly: a nul byte is put there while the value is decoded,
		// so that the decoder never reads the next value, whatever it makes of a malformed one.
		// A number needs it too, because the decoder validates the byte which ends it.
		saved := s.buf[end]
		s.buf[end] = nul
		ctx.Buf = s.buf[:end+1]
		cursor, err = dec.Decode(ctx, s.cursor, 0, p)
		s.buf[end] = saved
	} else {
		// true, false and null end at the first byte which can't belong to them, and the decoder
		// reports that byte when it is wrong: it sees the whole buffer, ended by the nul byte.
		ctx.Buf = s.buf[:s.length+1]
		cursor, err = dec.Decode(ctx, s.cursor, 0, p)
	}
	ctx.Buf = nil
	if err != nil {
		return s.decodeError(err, dec, typ, p, end)
	}
	if ctx.HasTypeError() {
		return s.typeError(typ, cursor)
	}
	s.cursor = cursor
	if countsScanned {
		s.scanned += s.offset + cursor - s.readStart
	}
	s.more.reset()
	if s.tokenState != tokenTopValue {
		s.tokenValueEnd()
	}
	return nil
}

// callDecoder decodes the value at the cursor, which ends at end, by dec into p, nested in depth objects and arrays,
// and returns the position after it. The decoded strings are copied out of the buffer, which they would keep alive.
func (s *Stream) callDecoder(dec Decoder, p unsafe.Pointer, depth, end int64) (int64, error) {
	ctx := s.ctx
	var cursor int64
	var err error
	if c := s.buf[s.cursor]; c != 't' && c != 'f' && c != 'n' {
		// The end of the value is known exactly: a nul byte is put there while the value is decoded,
		// so that the decoder never reads the next value, whatever it makes of a malformed one.
		// A number needs it too, because the decoder validates the byte which ends it.
		saved := s.buf[end]
		s.buf[end] = nul
		ctx.Buf = s.buf[:end+1]
		ctx.origin = ctx.Buf
		cursor, err = dec.Decode(ctx, s.cursor, depth, p)
		s.buf[end] = saved
	} else {
		// true, false and null end at the first byte which can't belong to them, and the decoder
		// reports that byte when it is wrong: it sees the whole buffer, ended by the nul byte.
		ctx.Buf = s.buf[:s.length+1]
		ctx.origin = ctx.Buf
		cursor, err = dec.Decode(ctx, s.cursor, depth, p)
	}
	ctx.Buf, ctx.origin = nil, nil
	return cursor, err
}

// decodeError returns the error of the decoder, whose offset is made the one in the whole input, of the value at
// the cursor, which scanValue ended at end. A type error before it is not kept for the next value.
//
// A syntax error is the first one of the value by the grammar, as encoding/json reports it: it checks the value
// before it decodes it. The decoder never writes the buffer, which keeps the input.
//
//go:noinline
func (s *Stream) decodeError(err error, dec Decoder, typ, p unsafe.Pointer, end int64) error {
	if c := s.buf[s.cursor]; c == '-' || c-'0' <= 9 {
		// the end of the number by the grammar, which an invalid number has where it is invalid
		if n, _ := jsonnum.Scan(s.buf[s.cursor:end]); s.cursor+int64(n) < end {
			first := s.cursor + int64(n)
			// the bytes of the numbers which scanLiteral read are more than one value, as 01 is 0 and 1: the first
			// is decoded alone, and the next one after it
			s.ctx.DiscardTypeError()
			return s.decodeFirstNumber(dec, typ, p, first)
		}
	}
	s.ctx.DiscardTypeError()
	// a syntax error of the value is the error, whatever the decoder met before it, as encoding/json checks the value
	// before it decodes it
	if serr := s.valueSyntaxError(); serr != nil {
		return s.fail(serr)
	}
	return s.fail(s.totalOffsetError(err))
}

// decodeFirstNumber decodes the number at the cursor, which ends at end by the grammar, as Decode does.
func (s *Stream) decodeFirstNumber(dec Decoder, typ, p unsafe.Pointer, end int64) error {
	ctx := s.ctx
	cursor, err := s.callDecoder(dec, p, 0, end)
	if err != nil {
		s.ctx.DiscardTypeError()
		if serr := s.valueSyntaxError(); serr != nil {
			return s.fail(serr)
		}
		return s.fail(s.totalOffsetError(err))
	}
	if ctx.HasTypeError() {
		return s.typeError(typ, cursor)
	}
	s.cursor = cursor
	if countsScanned {
		s.scanned += s.offset + cursor - s.readStart
	}
	s.more.reset()
	if s.tokenState != tokenTopValue {
		s.tokenValueEnd()
	}
	return nil
}

// valueSyntaxError returns the first syntax error of the value at the cursor by the grammar, in what has been read:
// the error of a byte, with its offset in the whole input, or the unexpected end of the input, where encoding/json
// reads more. A number, true, false or null is followed by the byte after it, or by the end of the input: the scan
// of the value has read either. It is nil if what has been read has none.
//
//go:noinline
func (s *Stream) valueSyntaxError() error {
	return s.valueSyntaxErrorAt(0)
}

// valueSyntaxErrorAt is valueSyntaxError for a value nested in depth objects and arrays.
func (s *Stream) valueSyntaxErrorAt(depth int64) error {
	err := s.wholeCharError(func() error {
		_, err := skipValue(s.buf[:s.length+1], s.cursor, depth)
		return err
	})
	if err == nil {
		return nil
	}
	return s.grammarError(err)
}

// wholeCharError returns the syntax error which scan finds in what has been read, or nil. An error at the last
// character of what has been read, which a read cut, is found again once the character is whole ( see fillLastChar ),
// and only then: a read after an error before it would wait for input which the error doesn't need.
func (s *Stream) wholeCharError(scan func() error) error {
	err := scan()
	e, ok := err.(*errors.SyntaxError)
	if !ok || errors.IsAtEnd(err) || s.length == 0 {
		return err
	}
	start := s.length - 1
	for start > 0 && start > s.length-utf8.UTFMax && !utf8.RuneStart(s.buf[start]) {
		start--
	}
	if utf8.FullRune(s.buf[start:s.length]) || e.Offset <= start {
		return err
	}
	s.fillLastChar()
	return scan()
}

// fillLastChar reads more if a read ended in the middle of a UTF-8 sequence, so that a syntax error at it shows the
// character, as encoding/json does, whatever the reads of the input. Only that character is completed: a read may end
// in the middle of another one.
func (s *Stream) fillLastChar() {
	if s.length == 0 {
		return
	}
	// the start of the last character, in the last bytes, kept relative to the cursor, as a read may move the data
	start := s.length - 1
	for start > 0 && start > s.length-utf8.UTFMax && !utf8.RuneStart(s.buf[start]) {
		start--
	}
	rel := start - s.cursor
	for s.cursor+rel >= 0 && !utf8.FullRune(s.buf[s.cursor+rel:s.length]) && s.read() {
	}
}

// grammarError returns the syntax error err of a scan of the grammar in what has been read, as valueSyntaxError
// reports it.
func (s *Stream) grammarError(err error) error {
	if errors.IsAtEnd(err) {
		return s.unexpectedEndError()
	}
	if at, ok := errors.ValueStartAt(err); ok {
		if merr := s.valueStartMismatch(at); merr != nil {
			return merr
		}
	}
	return s.scannedOffset(s.totalOffsetError(err))
}

// scannedOffset makes the offset of the syntax error of the value read from readStart the one which encoding/json
// of the Go version reports ( see countsScanned ).
func (s *Stream) scannedOffset(err error) error {
	if e, ok := err.(*errors.SyntaxError); ok && countsScanned {
		e.Offset += s.scanned - s.readStart
	}
	return err
}

// typeError returns the type error of the value of the pointer type typ which Decode decoded from the cursor up
// to end, and moves the cursor after the value: the stream goes on after it, as the one of encoding/json does.
// Its offset is relative to the value, as encoding/json of the Go version reports it ( see StreamOffsetBase ).
// It takes the decoder of the type again, so that Decode keeps nothing more across its call of the decoder.
//
//go:noinline
func (s *Stream) typeError(typ unsafe.Pointer, end int64) error {
	start := s.cursor
	s.cursor = end
	if countsScanned {
		s.scanned += s.offset + end - s.readStart
	}
	s.more.reset()
	if s.tokenState != tokenTopValue {
		s.tokenValueEnd()
	}
	ctx := s.ctx
	dec, err := s.DecoderOf(typ)
	if err != nil {
		ctx.DiscardTypeError()
		return err
	}
	// the path of the error is walked in the value ended by the nul byte, as it was decoded
	saved := s.buf[end]
	s.buf[end] = nul
	ctx.Buf = s.buf[:end+1]
	base := StreamOffsetBase(s.prevEnd, s.offset+start) - s.offset
	err = ctx.TypeError(dec, runtime.TypeOfPtr(typ), start, base)
	ctx.Buf = nil
	s.buf[end] = saved
	return err
}

// totalOffsetError makes the offset of an error of the decoder, which is relative to the buffer,
// the offset in the whole input.
func (s *Stream) totalOffsetError(err error) error {
	switch e := err.(type) {
	case *errors.SyntaxError:
		e.Offset += s.offset
	case *errors.UnmarshalTypeError:
		e.Offset += s.offset
	}
	return err
}

// tokenString decodes the string at the cursor by the string decoder, as Decode does.
func (s *Stream) tokenString() (any, error) {
	s.readStart = s.offset + s.cursor
	end, err := s.scanValue()
	if err != nil {
		return nil, err
	}
	saved := s.buf[end]
	s.buf[end] = nul
	literal, _, err := tokenStringDecoder.decodeByte(s.buf[:end+1], s.cursor)
	s.buf[end] = saved
	if err != nil {
		if serr := s.valueSyntaxError(); serr != nil {
			return nil, serr
		}
		return nil, s.totalOffsetError(err)
	}
	s.cursor = end
	if countsScanned {
		s.scanned += s.offset + end - s.readStart
	}
	if len(literal) != 0 && !isPartOf(literal, s.buf) {
		// the bytes of the escaped string, which decodeLiteral made: they are the string's own
		return unsafe.String(unsafe.SliceData(literal), len(literal)), nil
	}
	return string(literal), nil
}

var tokenStringDecoder = newStringDecoder("", "")

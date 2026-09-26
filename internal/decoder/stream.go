package decoder

import (
	"bytes"
	"io"
	"math/bits"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
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
	n, err := s.r.Read(s.buf[s.length : int64(len(s.buf))-bufPadding])
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
	remain := s.length - s.cursor
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
	buf := make([]byte, size+bufPadding)
	copy(buf, s.buf[s.cursor:s.length])
	s.offset += s.cursor
	s.buf = buf
	s.length = remain
	s.cursor = 0
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
	for {
		if !s.fill() {
			return false
		}
		switch s.buf[s.cursor] {
		case ' ', '\n', '\t', '\r':
			s.cursor++
		default:
			return true
		}
	}
}

// prepare moves the cursor to the beginning of the next value: the white space is skipped, and, between the
// tokens of an object or an array, the comma or the colon which the grammar has before the value.
func (s *Stream) prepare() error {
	if s.tokenState != tokenTopValue {
		return s.prepareInTokens()
	}
	if !s.skipWhiteSpace() {
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
	if !s.fill() {
		return 0, s.endError()
	}
	switch c := s.buf[s.cursor]; {
	case c == '{' || c == '[':
		return s.scanCompound()
	case c == '"':
		return s.scanString()
	case isLiteralChar[c]:
		return s.scanLiteral()
	}
	// The decoder reports the invalid character.
	return s.cursor + 1, nil
}

// scanCompound finds the end of the object or the array at the cursor: the position after the bracket
// which closes it.
func (s *Stream) scanCompound() (int64, error) {
	// The position is kept relative to the cursor, because a read may move the data ( grow ).
	var rel int64
	sc := compoundScanner{maxDepth: maxDecodeNestingDepth}
	for {
		pos, found, err := sc.scan(s.buf, s.cursor+rel, s.length)
		if err != nil {
			return 0, s.totalOffsetError(err)
		}
		if found {
			return pos, nil
		}
		rel = pos - s.cursor
		if !s.read() {
			return 0, s.unexpectedEndError()
		}
	}
}

// scanString finds the end of the string at the cursor: the position after the quote which closes it.
func (s *Stream) scanString() (int64, error) {
	var (
		rel     int64 = 1 // after the opening quote
		escaped bool
	)
	for {
		buf := s.buf
		pos := s.cursor + rel
		end := s.length
		for pos+8 <= end {
			w := load64(buf, pos)
			if byteMask(w, '\\') != 0 || escaped {
				break
			}
			if quote := byteMask(w, '"'); quote != 0 {
				return pos + int64(bits.TrailingZeros64(quote)/8) + 1, nil
			}
			pos += 8
		}
		lim := min(pos+8, end)
		for pos < lim {
			c := buf[pos]
			pos++
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				return pos, nil
			}
		}
		rel = pos - s.cursor
		if pos < end {
			continue
		}
		if !s.read() {
			return 0, s.unexpectedEndError()
		}
	}
}

// scanLiteral finds the end of the number, true, false or null at the cursor: the position of
// the first byte which can't belong to it, or the end of the input.
func (s *Stream) scanLiteral() (int64, error) {
	var rel int64
	for {
		buf := s.buf
		pos := s.cursor + rel
		end := s.length
		for pos < end {
			if !isLiteralChar[buf[pos]] {
				return pos, nil
			}
			pos++
		}
		rel = pos - s.cursor
		if !s.read() {
			if s.readErr != io.EOF {
				return 0, s.readErr
			}
			return s.cursor + rel, nil
		}
	}
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
	end, err := s.scanValue()
	if err != nil {
		return s.fail(err)
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
		return s.decodeError(err)
	}
	if ctx.HasTypeError() {
		return s.typeError(typ, cursor)
	}
	s.cursor = cursor
	if s.tokenState != tokenTopValue {
		s.tokenValueEnd()
	}
	return nil
}

// decodeError returns the error of the decoder, whose offset is made the one in the whole input. A type error
// before it is not kept for the next value.
//
//go:noinline
func (s *Stream) decodeError(err error) error {
	s.ctx.DiscardTypeError()
	return s.fail(s.totalOffsetError(err))
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
	end, err := s.scanValue()
	if err != nil {
		return nil, err
	}
	saved := s.buf[end]
	s.buf[end] = nul
	literal, _, err := tokenStringDecoder.decodeByte(s.buf[:end+1], s.cursor)
	s.buf[end] = saved
	if err != nil {
		return nil, s.totalOffsetError(err)
	}
	s.cursor = end
	return string(literal), nil
}

var tokenStringDecoder = newStringDecoder("", "")

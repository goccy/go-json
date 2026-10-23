package json

import (
	"cmp"
	"errors"
	"io"
	"math"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/decoder"
	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
)

// An UnmarshalJSONFrom method and a function of UnmarshalFromFunc read a jsontext.Decoder, which the decoder gives
// them at the place of their value in its input: the jsontext.Decoder is attached to the input there, so that its
// stack and its errors point where the value is, and the decoding goes on where it stopped reading.

func init() {
	decoder.V2Hooks.UnmarshalerFromType = reflect.TypeFor[UnmarshalerFrom]()
	decoder.V2Hooks.UnmarshalFrom = func(ctx *decoder.RuntimeContext, cursor, depth int64, typ reflect.Type, p unsafe.Pointer) (int64, error) {
		m, _ := reflect.NewAt(typ, p).Interface().(UnmarshalerFrom)
		return unmarshalFrom(ctx, cursor, depth, typ, fromMethod{m})
	}
	decoder.V2Hooks.SkipStreamValue = func(s *decoder.Stream, depth int64) error {
		_, err := unmarshalFrom(s.Context(), 0, depth, nil, fromSkip{})
		return err
	}
}

// fromState is the state of the calls of the methods and the functions which read a jsontext.Decoder, which a
// context keeps from a call of Unmarshal to the next ( see decoder.V2Call.From ): the decoder, which is attached to
// the input of the call at the place of each value which they decode, and the levels of the input before the
// place, which are found only when the decoder needs them, for its stack or a JSON pointer ( see Levels ).
type fromState struct {
	// att is where dec is attached, of the input of the call, att.In, which is set if dec was attached in the call,
	// by the options of the call.
	dec jsontext.Decoder
	att textcoder.DecoderAttachment
	// at is the place where dec was attached last, in front of the delimiter before the value, which the detach
	// moves att.Pos from: tracked follows the levels of the input up to it.
	at      int
	tracked tracker
	// stream is the stream of the call, if it decodes one, the levels of whose path are the ones around the place
	// ( see decoder.V2Call.Path ), followed by the ones of the value in the buffer which starts at valueStart of
	// att.In, within which the place is, which tracked follows from there, and levels what Levels returned last.
	// reader reads on from the stream for a value which the buffer doesn't hold whole, into room, which the decoder
	// is given the bytes of the buffer of the stream from the place in, and moves them and reads into.
	stream     *decoder.Stream
	valueStart int
	levels     []textcoder.Level
	reader     streamReader
	room       []byte
}

// maxKeptRoom is the size of the largest room which a state keeps: one which grew for a long token is dropped.
const maxKeptRoom = 64 << 10

// Levels returns the levels of the input in front of the place where the decoder is attached ( see
// textcoder.Outer ). The places of a call only go forward, so the input is read once, as far as the last place.
func (st *fromState) Levels() []textcoder.Level {
	if st.stream == nil {
		return st.tracked.sync(st.att.In[:st.at], math.MaxInt)
	}
	st.levels = pathLevels(st.levels, st.stream.Context().V2Call().Path)
	if st.at > st.valueStart {
		// a place within the value, whose levels follow the ones of the path
		inner := st.tracked.sync(st.att.In[st.valueStart:st.at], math.MaxInt)
		st.levels[len(st.levels)-1].Count += inner[0].Count
		st.levels = append(st.levels, inner[1:]...)
	}
	return st.levels
}

// pathLevels returns the levels of the path of a stream in dst: an object which has no member yet has no name.
func pathLevels(dst, path []textcoder.Level) []textcoder.Level {
	dst = append(dst[:0], path...)
	for i := range dst {
		if dst[i].Count == 0 {
			dst[i].Name = nil
		}
	}
	return dst
}

// streamReader reads on from a stream, for a decoder attached at its value ( see decoder.Stream.ReadOn ).
type streamReader struct{ s *decoder.Stream }

func (r *streamReader) Read(p []byte) (int, error) { return r.s.ReadOn(p) }

// release ends the call of the state, which a method or a function used: the input of the call and its options,
// which may hold the functions of the caller, are dropped, and the levels are found again from the start of the
// input of the next call.
func (st *fromState) release() {
	textcoder.ReleaseDecoder(unsafe.Pointer(&st.dec))
	st.att.In, st.att.Opts, st.att.Outer, st.att.R, st.att.Rest = nil, nil, nil, nil, nil
	st.tracked.release()
	st.tracked.stale = true
	st.stream, st.reader.s = nil, nil
	clear(st.levels[:cap(st.levels)])
	if cap(st.room) > maxKeptRoom {
		st.room = nil
	}
}

// fromCall is the call of an UnmarshalJSONFrom method or of a function of UnmarshalFromFunc: a type of the call,
// rather than a func value, which the closure of the method or of the function would allocate.
type fromCall interface {
	call(dec *jsontext.Decoder) error
}

// fromMethod is the call of UnmarshalJSONFrom of m.
type fromMethod struct{ m UnmarshalerFrom }

func (c fromMethod) call(dec *jsontext.Decoder) error { return c.m.UnmarshalJSONFrom(dec) }

// fromFunc is the call of a function of UnmarshalFromFunc with v.
type fromFunc[T any] struct {
	fn func(*jsontext.Decoder, T) error
	v  T
}

func (c fromFunc[T]) call(dec *jsontext.Decoder) error { return c.fn(dec, c.v) }

// fromSkip is the skip of a value which a stream doesn't hold whole, which is checked as encoding/json/v2 checks a
// value it skips ( see decoder.V2Hooks.SkipStreamValue ).
type fromSkip struct{}

func (fromSkip) call(dec *jsontext.Decoder) error { return dec.SkipValue() }

// unmarshalFrom calls c, the UnmarshalJSONFrom method or a function of UnmarshalFromFunc of a value of typ, with the
// decoder of the state attached to the input at the value at cursor, in depth levels, and returns the cursor after
// the value. c must read exactly that value, or else return errors.ErrUnsupported without reading anything, which
// leaves the value to the next function, or to the next method or the kind of the type ( decoder.ErrUseDefault ).
// Its error is reported at the place where it stopped reading, as encoding/json/v2 reports it.
func unmarshalFrom[C fromCall](ctx *decoder.RuntimeContext, cursor, depth int64, typ reflect.Type, c C) (int64, error) {
	call := ctx.V2Call()
	st := (*fromState)(call.From)
	if st == nil {
		st = new(fromState)
		st.tracked.stale = true
		call.From = unsafe.Pointer(st)
	}
	att := &st.att
	in, start, readsOn := st.place(ctx, cursor, depth)
	if start < 0 {
		// no value: encoding/json/v2 calls no method at the end of the input, which is a syntax error.
		return 0, io.ErrUnexpectedEOF
	}
	att.In, att.Opts, att.Same = in, call.Config, att.In != nil
	att.Outer, att.OuterDepth = st, int(depth)
	st.at = att.Pos
	textcoder.AttachDecoder(unsafe.Pointer(&st.dec), att)
	depthBefore, countBefore := att.Depth, att.Count
	err := c.call(&st.dec)
	textcoder.DetachDecoder(unsafe.Pointer(&st.dec), att)
	if readsOn {
		// the stream goes on where the decoder stopped, which may have read on, and the room is kept
		st.stream.Resume(att.End, att.Rest)
		st.room = att.In[:0]
	}
	if err == nil && (att.Depth != depthBefore || att.Count != countBefore+1) {
		err = errNonSingularValue
	}
	if err == nil {
		return int64(att.Pos), nil
	}
	if errors.Is(err, errors.ErrUnsupported) {
		if att.Depth == depthBefore && att.Count == countBefore {
			return 0, decoder.ErrUseDefault
		}
		err = errUnsupportedMutation
	}
	if _, ok := err.(*jsontext.SyntacticError); ok || textcoder.IsReadError(err) {
		// a syntax error and an error of the reader, which are not wrapped, as encoding/json/v2 doesn't wrap them
		return 0, err
	}
	// the place of the error, where it stopped reading: the value if it read nothing, the value it read if it read
	// one, and else the level it is in, after the last token it read.
	serr, _ := err.(*SemanticError)
	if serr == nil {
		serr = &SemanticError{Err: err}
	}
	serr.Err = toUnexpectedEOF(serr.Err)
	serr.action = cmp.Or(serr.action, "unmarshal")
	where, offset := 0, att.PrevStart
	switch {
	case att.Depth == depthBefore && att.Count == countBefore:
		where, offset = +1, att.Base+start
	case att.Depth == depthBefore && att.Count == countBefore+1:
		where = -1
	}
	serr.ByteOffset = cmp.Or(serr.ByteOffset, offset)
	if serr.JSONPointer == "" {
		serr.JSONPointer = jsontext.Pointer(textcoder.DecoderPointer(unsafe.Pointer(&st.dec), where))
	}
	if serr.GoType == nil {
		serr.GoType = typ
	}
	return 0, serr
}

// place sets the place of the attachment of the state for the value at cursor, which a method or a function which
// reads a jsontext.Decoder is called for, and returns the input which the decoder reads, the start of the value in it,
// and whether the decoder reads on from the stream of the call, for a value which its buffer doesn't hold whole. The
// start is -1 at the end of the input, which has no value, at the top level, depth 0. The input is the one of the call, the buffer of the stream,
// or the part of it which holds the value of the stream which is decoded whole, which the value at cursor is in ( see
// decoder.V2Call.Value ).
func (st *fromState) place(ctx *decoder.RuntimeContext, cursor, depth int64) ([]byte, int64, bool) {
	call := ctx.V2Call()
	att := &st.att
	s := call.Stream
	st.stream, att.R, att.Base, att.Own = s, nil, 0, false
	var in []byte
	start := cursor
	if s == nil || !call.Value.ReadsOn {
		buf := ctx.Buf
		for isSpace(buf[start]) {
			start++ // the nul byte after the input ends the white space
		}
		in = buf[:len(buf)-1]
		if s == nil {
			if depth == 0 && start == int64(len(in)) {
				return nil, -1, false
			}
			att.Pos, att.Object, att.Top, att.Count = placeBefore(in, start)
			return in, start, false
		}
	}
	v := call.Value
	read, base := s.Read()
	att.Base, st.valueStart = base, int(v.Start)
	st.tracked.stale = true
	if v.ReadsOn {
		// the bytes of the buffer from the place, which the decoder reads on after in a room of its own
		st.room = append(st.room[:0], read[v.From:]...)
		in, start = st.room, v.Start-v.From
		att.Base, st.valueStart = base+v.From, int(start)
		v.Start, v.From = start, 0
		st.reader.s = s
		att.R, att.Own = &st.reader, true
	}
	if start == v.Start {
		// the value itself, whose place is the end of the token before it, in the innermost level of the path
		path := call.Path
		l := path[len(path)-1]
		att.Pos, att.Object, att.Top, att.Count = int(v.From), l.Object, len(path) == 1, l.Count
	} else {
		att.Pos, att.Object, att.Top, att.Count = placeBefore(in, start)
	}
	return in, start, v.ReadsOn
}

// placeBefore returns the place in front of the value at start of in, which a decoder attached to it reads from: the
// end of the token before the value, in front of the delimiter and the white space, where the input offset of
// encoding/json/v2 is, and the innermost level which the delimiter tells, whether it is an object, whether it is
// the top level, and its count, of which only the parity, which the delimiter needs, is right: the real one is in
// the levels of the state. The value is a name of an object if it is the key of a map.
func placeBefore(in []byte, start int64) (int, bool, bool, int64) {
	name := isName(in, start)
	i := start - 1
	for i >= 0 && isSpace(in[i]) {
		i--
	}
	if i < 0 {
		return 0, false, true, 0
	}
	if in[i] != ':' && in[i] != ',' {
		// the first name of an object, or the first element of an array, after its begin
		return int(i) + 1, name, false, 0
	}
	end := i
	for end > 0 && isSpace(in[end-1]) {
		end--
	}
	switch {
	case in[i] == ':':
		return int(end), true, false, 1
	case name:
		// the next name of an object
		return int(end), true, false, 2
	}
	// the next element of an array
	return int(end), false, false, 1
}

// isName reports whether the value at start of in is the name of a member of an object: a string which a colon
// follows.
func isName(in []byte, start int64) bool {
	if start == int64(len(in)) || in[start] != '"' {
		return false
	}
	i := start + 1
	for ; i < int64(len(in)) && in[i] != '"'; i++ {
		if in[i] == '\\' {
			i++
		}
	}
	for i++; i < int64(len(in)) && isSpace(in[i]); i++ {
	}
	return i < int64(len(in)) && in[i] == ':'
}

// isSpace reports whether c is white space of JSON.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

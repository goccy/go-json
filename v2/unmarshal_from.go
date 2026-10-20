package json

import (
	"cmp"
	"errors"
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
}

// fromState is the state of the calls of the methods and the functions which read a jsontext.Decoder, which a
// context keeps from a call of Unmarshal to the next ( see decoder.V2Call.From ): the decoder, which is attached to
// the input of the call at the place of each value which they decode, and the levels of the input before the
// place, which are found only when the decoder needs them, for its stack or a JSON pointer ( see Levels ).
type fromState struct {
	dec jsontext.Decoder
	att textcoder.DecoderAttachment
	// attached is whether dec was attached in the call, by the options of the call.
	attached bool
	// in is the input of the call, and at the place where dec was attached last, in front of the delimiter before
	// the value: tracked follows the levels of in up to it.
	in      []byte
	at      int
	tracked tracker
}

// Levels returns the levels of the input in front of the place where the decoder is attached ( see
// textcoder.Outer ). The places of a call only go forward, so the input is read once, as far as the last place.
func (st *fromState) Levels() []textcoder.Level {
	return st.tracked.sync(st.in[:st.at], math.MaxInt)
}

// release ends the call of the state, which a method or a function used: the input of the call and its options,
// which may hold the functions of the caller, are dropped, and the levels are found again from the start of the
// input of the next call.
func (st *fromState) release() {
	textcoder.ReleaseDecoder(unsafe.Pointer(&st.dec))
	st.att.In, st.att.Opts, st.att.Outer = nil, nil, nil
	st.in, st.attached = nil, false
	st.tracked.release()
	st.tracked.stale = true
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
	buf := ctx.Buf
	start := cursor
	for isSpace(buf[start]) {
		start++ // the nul byte after the input ends the white space
	}
	in := buf[:len(buf)-1]
	att := &st.att
	att.In, att.Base, att.Opts, att.Same = in, 0, call.Config, st.attached
	att.Pos, att.Object, att.Top, att.Count = placeBefore(in, start)
	att.Outer, att.OuterDepth = st, int(depth)
	st.in, st.at = in, att.Pos
	textcoder.AttachDecoder(unsafe.Pointer(&st.dec), att)
	st.attached = true
	depthBefore, countBefore := att.Depth, att.Count
	err := c.call(&st.dec)
	textcoder.DetachDecoder(unsafe.Pointer(&st.dec), att)
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
	if _, ok := err.(*jsontext.SyntacticError); ok {
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
	where, offset := 0, int64(att.PrevStart)
	switch {
	case att.Depth == depthBefore && att.Count == countBefore:
		where, offset = +1, start
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

// placeBefore returns the place in front of the value at start of in, which a decoder attached to it reads from: the
// end of the token before the value, in front of the delimiter, and the innermost level which the delimiter tells,
// whether it is an object, whether it is the top level, and its count, of which only the parity, which the
// delimiter needs, is right: the real one is in the levels of the state. The value is a name of an object if it is
// the key of a map.
func placeBefore(in []byte, start int64) (int, bool, bool, int64) {
	name := isName(in, start)
	i := start - 1
	for i >= 0 && isSpace(in[i]) {
		i--
	}
	switch {
	case i < 0:
		return 0, false, true, 0
	case in[i] == ':':
		return int(i), true, false, 1
	case in[i] == ',':
		// the next name of an object, or the next element of an array
		if name {
			return int(i), true, false, 2
		}
		return int(i), false, false, 1
	}
	// the first name of an object, or the first element of an array
	return int(start), name, false, 0
}

// isName reports whether the value at start of in is the name of a member of an object: a string which a colon
// follows.
func isName(in []byte, start int64) bool {
	if in[start] != '"' {
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

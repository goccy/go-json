package decoder

import (
	"bytes"
	"io"
	"reflect"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonfields"
	"github.com/goccy/go-json/internal/textcoder"
)

// The values of a stream of the semantics of encoding/json/v2, which the v2 json package decodes by UnmarshalRead. A
// value which the buffer holds whole is decoded by the decoders of v2, as Unmarshal decodes it, and an object or an
// array which it doesn't is decoded element by element, as Stream.Decode does for v1 ( see stream_elements.go ), by
// the rules of v2: the decoding stops at the first error, the names of an object are matched and checked as the
// decoders of v2 do, a value of a type which a method or a function reads whole is read whole, as encoding/json/v2
// reads it, and a method or a function which reads a jsontext.Decoder is given one which reads on from the reader of
// the stream. The strings are copied out of the buffer, as Unmarshal copies them out of its input, so that the buffer
// is reused.
//
// The objects and the arrays which a value is decoded element by element in are the path of the call ( see
// V2Call.Path ), which the v2 json package finds the place of an error with: the innermost function records the part
// of the buffer to read again ( see StreamPlace ), from the end of the token before the member, the element or the
// value which failed. That token, a delimiter or the bracket which begins the level, is kept in the buffer until the
// value after it is decoded, or begins to be decoded element by element: a decoder is attached at it, and the place
// of an error found again from it.

// streamsV2 are the streams of the calls of the v2 json package, which keep their buffers and their contexts.
var streamsV2 = sync.Pool{New: func() any { return NewStream(nil) }}

// maxReusedStreamBuffer is the size of the largest buffer which a stream of the pool keeps: one which grew for a long
// token or a value read whole is dropped.
const maxReusedStreamBuffer = 4 * (initBufSize + bufPadding)

// TakeStreamV2 returns a stream which reads r, for a decoding of v2 by the state of the call of its context ( see
// RuntimeContext.V2Call ), whose path is the top level.
func TakeStreamV2(r io.Reader) *Stream {
	s := streamsV2.Get().(*Stream)
	s.r = r
	call := s.ctx.V2Call()
	call.Stream = s
	call.Path = append(call.Path[:0], textcoder.Level{})
	return s
}

// ReleaseStreamV2 puts the stream back, without its reader and the state of its call.
func ReleaseStreamV2(s *Stream) {
	call := s.ctx.v2
	call.Stream, call.Value, call.Place = nil, StreamValue{}, StreamPlace{}
	if len(call.names.entries) != 0 {
		call.names.truncate(0, 0)
	}
	if call.skippers != [3]unsafe.Pointer{} {
		releaseSkippers(call)
	}
	if len(s.buf) > maxReusedStreamBuffer {
		s.buf = make([]byte, initBufSize+bufPadding)
	}
	s.length, s.cursor, s.offset = 0, 0, 0
	s.buf[0] = nul
	s.r, s.readErr, s.err = nil, nil, nil
	s.keeps, s.keepFrom = false, 0
	s.ctx.Option.Flags, s.ctx.Option.V2 = 0, 0
	streamsV2.Put(s)
}

// StartAt makes the stream, which has read nothing, start at the offset of the input, which its reader reads from.
func (s *Stream) StartAt(offset int64) {
	s.offset = offset
}

// Offset returns the offset in the input of the cursor of the stream.
func (s *Stream) Offset() int64 {
	return s.offset + s.cursor
}

// Drop empties the buffer of the stream, which a decoder attached at its value read on from and failed after: the
// stream has nothing to give back ( see Read ), as the decoding of the value fails.
func (s *Stream) Drop() {
	s.offset += s.length
	s.cursor, s.length = 0, 0
	s.buf[0] = nul
}

// Unread returns the bytes of the buffer which the stream read after its cursor.
func (s *Stream) Unread() []byte {
	return s.buf[s.cursor:s.length]
}

// Context returns the context of the stream, which its values are decoded by.
func (s *Stream) Context() *RuntimeContext {
	return s.ctx
}

// Read returns what has been read and not dropped, and the offset of its first byte in the input.
func (s *Stream) Read() ([]byte, int64) {
	return s.buf[:s.length], s.offset
}

// ReadError returns the error of the reader, which the stream reports at the end of what it read, or nil if the
// reader reached the end of the input, or has not failed.
func (s *Stream) ReadError() error {
	if s.readErr == io.EOF {
		return nil
	}
	return s.readErr
}

// ReadOn reads more of the input into p, for a decoder attached at the value of the stream ( see V2Call.Value ):
// initBufSize bytes at most, as the stream reads, and the error of the reader is kept, as the stream keeps it.
func (s *Stream) ReadOn(p []byte) (int, error) {
	if s.readErr != nil {
		return 0, s.readErr
	}
	n, err := s.r.Read(p[:min(len(p), initBufSize)])
	if err != nil {
		s.readErr = err
	}
	return n, err
}

// Resume makes the stream go on at the offset end of the input, where a decoder attached at its value stopped: in the
// buffer, or, if the decoder read on, before rest, what it read and didn't use, which the buffer takes.
func (s *Stream) Resume(end int64, rest []byte) {
	if rest == nil {
		s.cursor = end - s.offset
		return
	}
	if int64(len(rest)) > int64(len(s.buf))-bufPadding {
		s.buf = make([]byte, int64(len(rest))+bufPadding)
	}
	s.length = int64(copy(s.buf, rest))
	s.buf[s.length] = nul
	s.offset, s.cursor = end, 0
}

// DecodeV2 decodes the next value of the stream by dec, a decoder of v2, into p: whole if the buffer holds it, and
// else element by element. The value is nested in depth objects and arrays, the ones of the path but the top level.
// It returns the error of the reader, io.EOF at the end of the input, if no value is left, and the error of the
// decoding, whose place is recorded ( see V2Call.Place ) unless it is an error of encoding/json/v2 at its place.
func (s *Stream) DecodeV2(dec Decoder, p unsafe.Pointer, depth int64) error {
	s.ctx.v2.Place.Set = false
	// the end of the token before the value, before the white space, which is kept while the value is looked for
	from := s.offset + s.cursor
	keeps, keepFrom := s.keepFromHere(from)
	if !s.skipSpaceBefore() {
		s.keeps, s.keepFrom = keeps, keepFrom
		return s.endError()
	}
	return s.decodeTopV2(dec, p, depth, from, keeps, keepFrom)
}

// DecodeV2After is DecodeV2 for a value at the offset start of the input, after from, the offset of the end of the
// token before it, which the stream reads from: the white space and the delimiter between them, which were checked,
// are skipped. It never returns io.EOF.
func (s *Stream) DecodeV2After(dec Decoder, p unsafe.Pointer, depth, from, start int64) error {
	// the names of the objects of the value which an error ended are dropped after it: the value of a method, whose
	// error the method may ignore, leaves none to the objects around it ( see callNames )
	entries, copies := s.ctx.v2.names.mark()
	err := s.decodeV2After(dec, p, depth, from, start)
	s.ctx.v2.names.truncate(entries, copies)
	return err
}

func (s *Stream) decodeV2After(dec Decoder, p unsafe.Pointer, depth, from, start int64) error {
	s.ctx.v2.Place.Set = false
	keeps, keepFrom := s.keepFromHere(from)
	for s.offset+s.length <= start {
		if !s.read() {
			s.keeps, s.keepFrom = keeps, keepFrom
			return s.unexpectedEndError()
		}
	}
	s.cursor = start - s.offset
	return s.decodeTopV2(dec, p, depth, from, keeps, keepFrom)
}

// decodeTopV2 decodes the value at the cursor of DecodeV2, after from, the end of the token before it, which is kept
// while the value is looked for: keeps and keepFrom are what the stream kept before.
func (s *Stream) decodeTopV2(dec Decoder, p unsafe.Pointer, depth, from int64, keeps bool, keepFrom int64) error {
	switch dec.(type) {
	case *v2UnmarshalFromDecoder, *v2FuncDecoder, *ptrDecoder, *basicPtrDecoder:
		// a method or a function which reads a jsontext.Decoder, which may read on past the value, as the one of
		// encoding/json/v2 reads the values after it, also through a pointer
		s.keeps, s.keepFrom = keeps, keepFrom
		s.beginElements()
		return s.decodeNestedV2(dec, p, depth, from)
	}
	end, err := s.scanValue()
	s.keeps, s.keepFrom = keeps, keepFrom
	switch {
	case err == errLongValue:
		s.beginElements()
		return s.decodeNestedV2(dec, p, depth, from)
	case err != nil:
		return s.placeAt(err, from)
	}
	return s.decodeWholeV2(dec, p, depth, end, from)
}

// beginElements makes the index of the objects and the arrays which the value at the cursor is decoded element by
// element in start at the value ( see bracketIndex ).
func (s *Stream) beginElements() {
	if s.elements == nil {
		// the entries of the first steps of the lookups ( see indexFirstStep ), which a slice grown from nothing
		// would take a few allocations for
		s.elements = &bracketIndex{closes: make([]bracketClose, 0, indexFirstStep/4)}
	}
	*s.elements = bracketIndex{closes: s.elements.closes[:0], end: s.offset + s.cursor, buf: s.elements.buf}
}

// AfterTopLevel moves the stream to the end of the input after its top-level value, which only white space may
// follow, and reports whether it is there: if not, the place of the byte after the value is recorded ( see
// V2Call.Place ). It returns the error of the reader.
func (s *Stream) AfterTopLevel() (bool, error) {
	if s.skipWhiteSpace() {
		s.setPlace(s.offset+s.cursor, s.length)
		return false, nil
	}
	if s.readErr != io.EOF {
		return false, s.readErr
	}
	return true, nil
}

// setPlace records the place of an error of the decoding, which the buffer has from the total offset from up to end,
// unless an inner value recorded its own ( see V2Call.Place ).
func (s *Stream) setPlace(from, end int64) {
	if pl := &s.ctx.v2.Place; !pl.Set {
		*pl = StreamPlace{From: from - s.offset, End: end, Set: true}
	}
}

// placeAt records the place of err, which the buffer has from the total offset from up to what has been read, and
// returns err.
func (s *Stream) placeAt(err error, from int64) error {
	s.setPlace(from, s.length)
	return err
}

// placeValue is placeAt for the value at the cursor, which ends at end.
func (s *Stream) placeValue(err error, from, end int64) error {
	s.setPlace(from, end)
	return err
}

// keepFromHere makes the buffer keep the bytes from the total offset from, which reads would drop, and returns what
// it kept before, which the caller sets back.
func (s *Stream) keepFromHere(from int64) (bool, int64) {
	keeps, keepFrom := s.keeps, s.keepFrom
	if !s.keeps || s.keepFrom > from {
		s.keeps, s.keepFrom = true, from
	}
	return keeps, keepFrom
}

// SetPath sets the path of the call of the stream to levels, the place of the next value which it decodes, copied
// into path, whose arrays it reuses, and returns the path it had, which the caller sets back by RestorePath.
func (s *Stream) SetPath(path, levels []textcoder.Level) []textcoder.Level {
	call := s.ctx.v2
	saved := call.Path
	path = path[:0]
	for _, l := range levels {
		n := len(path)
		if n < cap(path) {
			// the entry, whose name keeps its array: the name of the level is copied, as the names of the path are
			// copied into ( see setName )
			path = path[:n+1]
			path[n].Object, path[n].Count = l.Object, l.Count
		} else {
			path = append(path, textcoder.Level{Object: l.Object, Count: l.Count})
		}
		name := append(path[n].Name[:0], l.Name...)
		if name == nil && l.Name != nil {
			name = []byte{} // the name "", which is a name
		}
		path[n].Name = name
	}
	call.Path = path
	return saved
}

// RestorePath sets the path of the call of the stream back to saved, which SetPath returned, and returns the path
// which it set, whose arrays the caller keeps for the next SetPath.
func (s *Stream) RestorePath(saved []textcoder.Level) []textcoder.Level {
	call := s.ctx.v2
	path := call.Path
	call.Path = saved
	return path
}

// pushLevel adds the level of an object or an array which is decoded element by element to the path, and returns
// its index.
func (s *Stream) pushLevel(object bool) int {
	call := s.ctx.v2
	if n := len(call.Path); n < cap(call.Path) {
		// the entry, whose name keeps its array for the names of the level
		call.Path = call.Path[:n+1]
		l := &call.Path[n]
		l.Object, l.Count, l.Name = object, 0, l.Name[:0]
		return n
	}
	call.Path = append(call.Path, textcoder.Level{Object: object})
	return len(call.Path) - 1
}

// level returns the level i of the path.
func (s *Stream) level(i int) *textcoder.Level {
	return &s.ctx.v2.Path[i]
}

// popLevel removes the level i, which ended, from the path.
func (s *Stream) popLevel(i int) {
	s.ctx.v2.Path = s.ctx.v2.Path[:i]
}

// decodeNestedV2 decodes the value at the cursor, nested in depth objects and arrays, by dec into p: whole if the
// buffer holds it, and else element by element. from is the total offset of the end of the token before the value,
// which the buffer has, and keeps while the value is looked for and read whole, but not while it is decoded element by
// element: a decoder attached at the value is attached there before anything more is read.
func (s *Stream) decodeNestedV2(dec Decoder, p unsafe.Pointer, depth, from int64) error {
	switch d := dec.(type) {
	case *ptrDecoder:
		if s.buf[s.cursor] != 'n' {
			// the value which the pointer points to, which is set to a new one if it is nil: its decoder may be one
			// of a method or a function which reads a jsontext.Decoder, and null sets the pointer to nil
			ptr := *(*unsafe.Pointer)(p)
			if ptr == nil {
				ptr = newValue(d.typ)
				*(*unsafe.Pointer)(p) = ptr
			}
			return s.decodeNestedV2(d.dec, ptr, depth, from)
		}
	case *basicPtrDecoder:
		if s.buf[s.cursor] != 'n' {
			// as a ptrDecoder, by the allocator of the type
			ptr := *(*unsafe.Pointer)(p)
			if ptr == nil {
				ptr = d.alloc()
				*(*unsafe.Pointer)(p) = ptr
			}
			return s.decodeNestedV2(d.dec, ptr, depth, from)
		}
	case *v2EmbeddedPtrDecoder:
		// a field through an embedded pointer, which is set to a new value if it is nil
		if *(*unsafe.Pointer)(p) == nil {
			if !d.settable {
				return s.placeAt(&errors.UnmarshalError{GoType: d.root, Start: s.cursor, NoKind: true, Err: errNilField}, from)
			}
			*(*unsafe.Pointer)(p) = newValue(d.elem)
		}
		return s.decodeNestedV2(d.dec, unsafe.Add(*(*unsafe.Pointer)(p), d.offset), depth, from)
	case *v2UnmarshalFromDecoder:
		if err := s.decodeAttached(d.typ, p, depth, from, nil); err != ErrUseDefault {
			return err
		}
		return s.decodeNestedV2(d.next, p, depth, from)
	case *v2FuncDecoder:
		return s.decodeFuncsV2(d, p, depth, from)
	case *intDecoder, *uintDecoder, *floatDecoder, *boolDecoder, v2StringDecoder:
		// a string, a number or a literal which the buffer has whole is decoded in it, as v1 decodes it, after from,
		// which a read for the rest of a string keeps
		if c := s.buf[s.cursor]; c != '{' && c != '[' {
			keeps, keepFrom := s.keepFromHere(from)
			decoded := s.decodeScalar(dec, p, depth)
			s.keeps, s.keepFrom = keeps, keepFrom
			if decoded {
				return nil
			}
		}
	}
	keeps, keepFrom := s.keepFromHere(from)
	end, whole, err := s.elementEnd(depth)
	if err == nil && !whole {
		// a value decoded element by element doesn't keep the bytes before it
		s.keeps, s.keepFrom = keeps, keepFrom
		var decoded bool
		if decoded, err = s.decodeElementsV2(dec, p, depth, from); decoded {
			return err
		}
		start := s.offset + s.cursor
		s.keepFromHere(from)
		if end, err = s.wholeEnd(depth); err != nil {
			err = s.wholeValueError(err, start)
		}
	}
	if err == nil {
		err = s.decodeWholeV2(dec, p, depth, end, from)
	} else {
		err = s.placeAt(err, from)
	}
	s.keeps, s.keepFrom = keeps, keepFrom
	return err
}

// wholeValueError returns err, an error of wholeEnd, which read the value at the total offset start whole for its
// decoder: a syntax error is the one of the read of the whole value, which encoding/json/v2 reads whole for the
// decoder.
func (s *Stream) wholeValueError(err error, start int64) error {
	if s.grammarFailure(err) && err != s.readErr {
		return &errors.UnmarshalError{Start: start - s.offset, Err: err, Method: errors.UnmarshalWholeValue}
	}
	return err
}

// decodeWholeV2 decodes the value at the cursor, which ends at end, nested in depth objects and arrays, by dec into
// p, as Unmarshal decodes it: a type error which a decoder of v1 recorded is the error of the value. from is the
// total offset of the end of the token before the value.
func (s *Stream) decodeWholeV2(dec Decoder, p unsafe.Pointer, depth, end, from int64) error {
	ctx := s.ctx
	call := ctx.v2
	call.Value = StreamValue{From: from - s.offset, Start: s.cursor}
	stop := s.length
	if c := s.buf[s.cursor]; c != 't' && c != 'f' && c != 'n' {
		// the end of the value is known: a nul byte is put there while the value is decoded, as Decode puts it
		stop = end
	}
	saved := s.buf[stop]
	s.buf[stop] = nul
	ctx.Buf = s.buf[:stop+1]
	ctx.origin = ctx.Buf
	cursor, err := dec.Decode(ctx, s.cursor, depth, p)
	if err == nil && ctx.typeError != nil {
		err = ctx.v2Error()
	}
	s.buf[stop] = saved
	ctx.Buf, ctx.origin = nil, nil
	call.Value = StreamValue{}
	if err != nil {
		ctx.DiscardTypeError()
		return s.placeValue(err, from, end)
	}
	s.cursor = cursor
	return nil
}

// decodeAttached decodes the value at the cursor, nested in depth objects and arrays, which the buffer doesn't hold
// whole, into p, of typ, by its UnmarshalJSONFrom method, or by fn, a function of UnmarshalFromFunc, if it is set:
// they are given a jsontext.Decoder attached at its place, after from, the end of the token before it, which reads
// on from the reader of the stream. It returns ErrUseDefault if they declined the value.
func (s *Stream) decodeAttached(typ reflect.Type, p unsafe.Pointer, depth, from int64, fn *UnmarshalFunc) error {
	ctx := s.ctx
	call := ctx.v2
	call.Value = StreamValue{From: from - s.offset, Start: s.cursor, ReadsOn: true}
	start := s.offset + s.cursor
	var err error
	if fn != nil {
		_, err = fn.Decode(ctx, s.cursor, depth, p, typ)
	} else {
		_, err = V2Hooks.UnmarshalFrom(ctx, s.cursor, depth, typ, p)
	}
	call.Value = StreamValue{}
	if err == ErrUseDefault {
		// the value, which the decoder read nothing of, is decoded from its start
		s.cursor = start - s.offset
	}
	return err
}

// decodeFuncsV2 decodes the value at the cursor, nested in depth objects and arrays, into p by the functions of d in
// their order, as v2FuncDecoder.Decode does: a function which reads a jsontext.Decoder is given one attached at its
// place, and any other the value read whole. from is the total offset of the end of the token before the value.
func (s *Stream) decodeFuncsV2(d *v2FuncDecoder, p unsafe.Pointer, depth, from int64) error {
	for i, fn := range d.funcs {
		if fn.FromDecoder {
			if err := s.decodeAttached(d.typ, p, depth, from, fn); err != ErrUseDefault {
				return err
			}
			continue
		}
		// a function which takes the value whole, which is read whole for it and the ones after it
		start := s.offset + s.cursor
		keeps, keepFrom := s.keepFromHere(from)
		end, whole, err := s.elementEnd(depth)
		if err == nil && !whole {
			if end, err = s.wholeEnd(depth); err != nil {
				err = s.wholeValueError(err, start)
			}
		}
		if err == nil {
			err = s.decodeWholeV2(&v2FuncDecoder{typ: d.typ, funcs: d.funcs[i:], next: d.next}, p, depth, end, from)
		} else {
			err = s.placeAt(err, from)
		}
		s.keeps, s.keepFrom = keeps, keepFrom
		return err
	}
	return s.decodeNestedV2(d.next, p, depth, from)
}

// decodeElementsV2 decodes the object or the array at the cursor, nested in depth objects and arrays, by dec into p
// element by element, after from, the total offset of the end of the token before it. It reports false, having read
// nothing, if dec doesn't decode it so: a value of a type which reads it whole, as encoding/json/v2 reads a value of
// a string, a number, bytes or a method whole before its error. A value of another kind than its type, which
// encoding/json/v2 reports at its first token, is the error.
func (s *Stream) decodeElementsV2(dec Decoder, p unsafe.Pointer, depth, from int64) (bool, error) {
	c := s.buf[s.cursor]
	var typ reflect.Type
	switch d := dec.(type) {
	case *v2StructDecoder:
		if c == '{' {
			if d.initErr != nil {
				return true, s.placeAt(&errors.UnmarshalError{GoType: d.initErr.GoType, Start: s.cursor, Err: d.initErr.Err}, from)
			}
			return true, s.decodeStructElementsV2(d, p, depth, from)
		}
		typ = d.typ
	case *v2MapDecoder:
		if c == '{' {
			return true, s.decodeMapElementsV2(d, p, depth, from)
		}
		typ = d.typ
	case *v2AnyMapDecoder:
		if c == '{' {
			m := *(*map[string]any)(p)
			if m == nil {
				m = map[string]any{}
			}
			err := s.decodeAnyObjectV2(d.iface, m, depth, from)
			*(*map[string]any)(p) = m
			return true, err
		}
		typ = interfaceMapType
	case *v2SliceDecoder:
		if c == '[' {
			return true, s.decodeSliceElementsV2(d, p, depth, from)
		}
		typ = d.typ
	case *v2ArrayDecoder:
		if c == '[' {
			return true, s.decodeArrayElementsV2(d, p, depth, from)
		}
		typ = d.typ
	case *boolDecoder:
		typ = d.typ
	case *v2InvalidDecoder:
		// a type of no representation, which is the error before the value is read
		return true, s.placeAt(&errors.UnmarshalError{GoType: d.typ, Start: s.cursor, NoKind: true, Err: d.err}, from)
	case *v2InterfaceDecoder:
		return s.decodeInterfaceElementsV2(d, p, depth, from)
	default:
		return false, nil
	}
	// an object or an array of a type of another kind, whose error is at its first token
	return true, s.placeAt(&errors.UnmarshalError{GoType: typ, Start: s.cursor}, from)
}

// openElementsV2 moves the cursor after the bracket which begins the object or the array whose elements are read,
// nested in depth objects and arrays, after from, the total offset of the end of the token before it, and returns the
// depth of its elements and the index of its level in the path.
func (s *Stream) openElementsV2(object bool, depth, from int64) (int64, int, error) {
	depth, err := s.openElements(depth)
	if err != nil {
		return 0, 0, s.placeAt(err, from)
	}
	return depth, s.pushLevel(object), nil
}

// nextMemberV2 moves the cursor after the name of the next member of the object whose members are read, and returns
// the name, quoted as the buffer has it: first is set before the first one. It reports false after the end of the
// object.
func (s *Stream) nextMemberV2(first bool) ([]byte, bool, error) {
	_, _, start, ok, err := s.nextMember(first)
	if err != nil || !ok {
		return nil, ok, err
	}
	return s.buf[start:s.cursor], true, nil
}

// memberValueV2 moves the cursor from after the name of a member to its value, and returns the total offset of the
// end of the name, the token before the value.
func (s *Stream) memberValueV2() (int64, error) {
	nameEnd := s.offset + s.cursor
	if err := s.memberValue(); err != nil {
		return 0, err
	}
	return nameEnd, nil
}

// copyName copies name, the name of the next member of an object, before its colon is read, which a read may drop
// the name in the buffer for, into the spare name of the path ( see setName ).
func (s *Stream) copyName(name []byte) {
	call := s.ctx.v2
	if call.spareName = append(call.spareName[:0], name...); call.spareName == nil {
		// the name "", which is a name, as a level without a name has none ( see textcoder.Level )
		call.spareName = []byte{}
	}
}

// setName makes the name which copyName copied the last name of the level lv of the path, of its member k, whose
// colon is read: the level takes the spare name, and gives its former name, which the levels inside the value of
// the member and the next member copy theirs into.
func (s *Stream) setName(lv, k int) {
	call := s.ctx.v2
	l := &call.Path[lv]
	l.Count = int64(2*k + 1)
	l.Name, call.spareName = call.spareName, l.Name[:0]
}

// decodeStructElementsV2 decodes the object at the cursor into the struct at p, as v2StructDecoder.Decode does, after
// from, the total offset of the end of the token before it.
func (s *Stream) decodeStructElementsV2(d *v2StructDecoder, p unsafe.Pointer, depth, from int64) error {
	depth, lv, err := s.openElementsV2(true, depth, from)
	if err != nil {
		return err
	}
	ctx := s.ctx
	// the loop of the options, which are looked at once for the object, as v2StructDecoder.Decode chooses it
	if !d.foldAlways && d.fallback == nil &&
		ctx.Option.V2&(V2MatchCaseInsensitiveNames|V2RejectUnknownMembers|V2AllowDuplicateNames) == V2AllowDuplicateNames {
		return s.decodeStructMembersV2(d, p, depth, lv)
	}
	rejectDuplicateNames := ctx.Option.V2&V2AllowDuplicateNames == 0
	fold := d.foldAlways || ctx.Option.V2&V2MatchCaseInsensitiveNames != 0
	var seen uint64
	var more []uint64
	var unknown nameSet
	for k := 0; ; k++ {
		// the end of the token before the member, which the place of its errors is read again from, kept until the
		// value of the member
		step := s.offset + s.cursor
		keeps, keepFrom := s.keepFromHere(step)
		quoted, ok, err := s.nextMemberV2(k == 0)
		if err != nil {
			return s.placeAt(err, step)
		}
		if !ok {
			unknown.end(ctx.v2)
			s.keeps, s.keepFrom = keeps, keepFrom
			s.popLevel(lv)
			return nil
		}
		start := s.cursor - int64(len(quoted))
		var field *structFieldSet
		var plain bool
		if fold {
			field, _, plain = d.foldedFieldV2(ctx, s.buf[:s.length+1], start)
		} else {
			field, _, plain = plainFieldV2(d.keyDecoder.keys, s.buf[:s.length+1], start)
		}
		name := quoted[1 : len(quoted)-1]
		if !plain {
			// a name which v2StructDecoder.Decode finds by memberOf
			candidate, _, err := d.keyDecoder.decodeKey(s.buf[:s.length+1], start, false)
			if err != nil {
				return s.placeAt(err, step)
			}
			if field, name, err = d.memberOf(ctx, quoted, candidate); err != nil {
				if e, ok := err.(*errors.UnmarshalError); ok {
					e.Start = start
				}
				return s.placeAt(err, step)
			}
		}
		switch {
		case field != nil:
			if !rejectDuplicateNames {
				break
			}
			if idx := field.fieldIdx; idx < 64 {
				bit := uint64(1) << uint(idx)
				if seen&bit != 0 {
					return s.placeAt(&errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}, step)
				}
				seen |= bit
			} else if fieldSeen(idx, &more, len(d.fields)) {
				return s.placeAt(&errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}, step)
			}
		case ctx.Option.V2&V2RejectUnknownMembers != 0 && d.fallback == nil:
			return s.placeAt(&errors.UnmarshalError{GoType: d.typ, Start: start, Err: errors.ErrUnknownName}, step)
		case rejectDuplicateNames:
			if !unknown.addCopy(ctx.V2Call(), nil, name) {
				return s.placeAt(&errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}, step)
			}
		}
		var raw []byte
		var rawValue *[]byte
		if field == nil && d.fallback != nil && d.fallback.mapType == nil {
			// a member which a jsontext.Value takes, checked before its value is read, with its name as the input
			// has it
			if rawValue, err = d.fallback.openRaw(p); err != nil {
				return s.placeAt(&errors.UnmarshalError{GoType: jsonfields.RawValueType, Start: start, Err: err}, step)
			}
			raw = bytes.Clone(quoted)
		}
		s.copyName(name)
		nameEnd, err := s.memberValueV2()
		if err != nil {
			return s.placeAt(err, step)
		}
		s.setName(lv, k)
		s.keeps, s.keepFrom = keeps, keepFrom
		switch {
		case field != nil:
			err = s.decodeNestedV2(field.dec, unsafe.Add(p, field.offset), depth, nameEnd)
		case rawValue != nil:
			err = s.appendRawV2(rawValue, raw, depth, nameEnd)
		case d.fallback != nil:
			err = s.decodeFallbackV2(d.fallback, p, depth, nameEnd, s.level(lv).Name)
		default:
			err = s.skipValueV2(depth, nameEnd)
		}
		if err != nil {
			return err
		}
		s.level(lv).Count = int64(2*k + 2)
	}
}

// decodeStructMembersV2 is the loop of decodeStructElementsV2 with V2AllowDuplicateNames, for a struct whose names
// are matched as they are, which has no fallback, without V2RejectUnknownMembers: the members of the object whose
// level of the path is lv, after its bracket, nested in depth objects and arrays, are decoded as decodeStructElements
// decodes them for v1, and the ones of no field skipped. The bytes from the end of the token before a member, the
// place of its errors, are kept while its name and a string, a number or a literal of its value are read, which the
// buffer has whole: as decodeStructElementsV2 keeps them for the name and decodeNestedV2 for the value.
func (s *Stream) decodeStructMembersV2(d *v2StructDecoder, p unsafe.Pointer, depth int64, lv int) error {
	ctx := s.ctx
	// what the caller keeps: the bytes of a member are kept from the end of the token before it, unless the caller
	// keeps them, from before the object
	keeps, keepFrom := s.keeps, s.keepFrom
	for k := 0; ; k++ {
		step := s.offset + s.cursor
		if !keeps {
			s.keeps, s.keepFrom = true, step
		}
		_, _, start, ok, err := s.nextMember(k == 0)
		if err != nil {
			return s.placeAt(err, step)
		}
		if !ok {
			s.keeps, s.keepFrom = keeps, keepFrom
			s.popLevel(lv)
			return nil
		}
		field, _, plain := plainFieldV2(d.keyDecoder.keys, s.buf[:s.length+1], start)
		name := s.buf[start+1 : s.cursor-1]
		if !plain {
			// a name which v2StructDecoder.Decode finds by memberOf
			candidate, _, err := d.keyDecoder.decodeKey(s.buf[:s.length+1], start, false)
			if err != nil {
				return s.placeAt(err, step)
			}
			if field, name, err = d.memberOf(ctx, s.buf[start:s.cursor], candidate); err != nil {
				if e, ok := err.(*errors.UnmarshalError); ok {
					e.Start = start
				}
				return s.placeAt(err, step)
			}
		}
		s.copyName(name)
		nameEnd := s.offset + s.cursor
		if err := s.memberValue(); err != nil {
			return s.placeAt(err, step)
		}
		s.setName(lv, k)
		switch {
		case field == nil:
			s.keeps, s.keepFrom = keeps, keepFrom
			err = s.skipValueV2(depth, nameEnd)
		case s.buf[s.cursor] != '{' && s.buf[s.cursor] != '[' && s.decodeScalar(field.dec, unsafe.Add(p, field.offset), depth):
			// a string, a number or a literal which the buffer has whole, as decodeNestedV2 decodes it: the decoders
			// of the values of v2 which decodeScalar decodes are the ones which decodeNestedV2 gives it
			s.keeps, s.keepFrom = keeps, keepFrom
		default:
			s.keeps, s.keepFrom = keeps, keepFrom
			err = s.decodeNestedV2(field.dec, unsafe.Add(p, field.offset), depth, nameEnd)
		}
		if err != nil {
			return err
		}
		s.level(lv).Count = int64(2*k + 2)
	}
}

// appendRawV2 appends the member of the name raw, as the input has it, whose value is at the cursor, to b, a
// jsontext.Value fallback which openRaw made ready, as v2Fallback.decodeMember does. The value, which encoding/json/v2
// reads whole, is nested in depth objects and arrays, after from, the total offset of the end of the name before it.
func (s *Stream) appendRawV2(b *[]byte, raw []byte, depth, from int64) error {
	keeps, keepFrom := s.keepFromHere(from)
	start := s.offset + s.cursor
	end, whole, err := s.elementEnd(depth)
	if err == nil && !whole {
		if end, err = s.wholeEnd(depth); err != nil {
			err = s.wholeValueError(err, start)
		}
	}
	// the value is read: nothing more is read for it
	s.keeps, s.keepFrom = keeps, keepFrom
	if err != nil {
		return s.placeAt(err, from)
	}
	ctx := s.ctx
	ctx.Buf = s.buf[:s.length+1]
	_, err = ctx.skipValueV2(s.cursor, depth)
	ctx.Buf = nil
	if err != nil {
		return s.placeValue(&errors.UnmarshalError{GoType: jsonfields.RawValueType, Start: s.cursor, Err: err}, from, end)
	}
	*b = append(*b, raw...)
	*b = append(*b, ':')
	*b = append(*b, s.buf[s.cursor:end]...)
	*b = append(*b, '}')
	s.cursor = end
	return nil
}

// decodeFallbackV2 decodes the value at the cursor, of the member of the name, which is of no field of the struct at
// p, into its map fallback, as v2Fallback.decodeMember does. from is the total offset of the end of the name before
// the value, nested in depth objects and arrays.
func (s *Stream) decodeFallbackV2(fb *v2Fallback, p unsafe.Pointer, depth, from int64, name []byte) error {
	m := valueAt(ptrTypeOf(fb.mapType), fb.target(p))
	if m.IsNil() {
		m.Set(reflect.MakeMap(fb.mapType))
	}
	key := reflect.ValueOf(string(name)).Convert(fb.mapType.Key())
	v := reflect.New(fb.mapType.Elem())
	if old := m.MapIndex(key); old.IsValid() {
		v.Elem().Set(old)
	}
	err := s.decodeNestedV2(fb.value, v.UnsafePointer(), depth, from)
	m.SetMapIndex(key, v.Elem())
	return err
}

// skipValueV2 skips the value at the cursor, nested in depth objects and arrays, after from, the total offset of the
// end of the token before it, as encoding/json/v2 skips it: checked for the text which v2 rejects, by the decoder if
// the buffer holds it whole, and else by a jsontext.Decoder attached at its place, which reads on from the reader of
// the stream.
func (s *Stream) skipValueV2(depth, from int64) error {
	keeps, keepFrom := s.keepFromHere(from)
	end, whole, err := s.elementEnd(depth)
	// the value is found, or the decoder is attached before anything more is read
	s.keeps, s.keepFrom = keeps, keepFrom
	if err != nil {
		return s.placeAt(err, from)
	}
	if !whole {
		call := s.ctx.v2
		call.Value = StreamValue{From: from - s.offset, Start: s.cursor, ReadsOn: true}
		err := V2Hooks.SkipStreamValue(s, depth)
		call.Value = StreamValue{}
		return err
	}
	ctx := s.ctx
	ctx.Buf = s.buf[:s.length+1]
	_, err = ctx.skipValueV2(s.cursor, depth)
	ctx.Buf = nil
	if err != nil {
		return s.placeValue(err, from, end)
	}
	s.cursor = end
	return nil
}

// decodeMapElementsV2 decodes the object at the cursor into the map at p, as v2MapDecoder.Decode does, after from, the
// total offset of the end of the token before it.
func (s *Stream) decodeMapElementsV2(d *v2MapDecoder, p unsafe.Pointer, depth, from int64) error {
	depth, lv, err := s.openElementsV2(true, depth, from)
	if err != nil {
		return err
	}
	ctx := s.ctx
	m := valueAt(d.ptrType, p)
	if m.IsNil() {
		m.Set(reflect.MakeMap(d.typ))
	}
	rejectDuplicateNames := ctx.Option.V2&V2AllowDuplicateNames == 0
	var seen reflect.Value
	if rejectDuplicateNames && m.Len() > 0 {
		seen = reflect.MakeMap(reflect.MapOf(d.typ.Key(), emptyStructType))
	}
	// names are the names of the keys, if two keys may have the same name
	var names nameSet
	keyNames := rejectDuplicateNames && !uniqueKeyNames(d.keyDecoder, d.typ.Key().Kind(), ctx.Option.V2&V2AllowInvalidUTF8 != 0)
	k, v := reflect.New(d.typ.Key()), reflect.New(d.typ.Elem())
	for i := 0; ; i++ {
		// the end of the token before the entry, which the place of its errors is read again from, kept until the
		// value of the entry
		step := s.offset + s.cursor
		keeps, keepFrom := s.keepFromHere(step)
		quoted, ok, err := s.nextMemberV2(i == 0)
		if err != nil {
			return s.placeAt(err, step)
		}
		if !ok {
			names.end(ctx.v2)
			s.keeps, s.keepFrom = keeps, keepFrom
			s.popLevel(lv)
			return nil
		}
		start := s.cursor - int64(len(quoted))
		// the name is decoded as a value of the key type, which the buffer holds whole
		k.Elem().SetZero()
		ctx.Buf = s.buf[:s.length+1]
		ctx.origin = ctx.Buf
		call := ctx.v2
		call.Value = StreamValue{From: step - s.offset, Start: start}
		_, err = d.keyDecoder.Decode(ctx, start, depth, k.UnsafePointer())
		call.Value = StreamValue{}
		ctx.Buf, ctx.origin = nil, nil
		if err == nil && ctx.typeError != nil {
			ctx.Buf = s.buf[:s.length+1]
			err = ctx.v2Error()
			ctx.Buf = nil
		}
		if err != nil {
			ctx.DiscardTypeError()
			return s.placeAt(err, step)
		}
		name := unquotedName(quoted)
		if keyNames && !names.addCopy(ctx.V2Call(), nil, name) {
			return s.placeAt(&errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}, step)
		}
		key := k.Elem()
		if old := m.MapIndex(key); old.IsValid() {
			if rejectDuplicateNames && (!seen.IsValid() || seen.MapIndex(key).IsValid()) {
				return s.placeAt(&errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}, step)
			}
			v.Elem().Set(old)
		} else {
			v.Elem().SetZero()
		}
		s.copyName(name)
		nameEnd, err := s.memberValueV2()
		if err != nil {
			return s.placeAt(err, step)
		}
		s.setName(lv, i)
		s.keeps, s.keepFrom = keeps, keepFrom
		err = s.decodeNestedV2(d.valueDecoder, v.UnsafePointer(), depth, nameEnd)
		m.SetMapIndex(key, v.Elem())
		if seen.IsValid() {
			seen.SetMapIndex(key, reflect.Zero(emptyStructType))
		}
		if err != nil {
			return err
		}
		s.level(lv).Count = int64(2*i + 2)
	}
}

// nextElementV2 moves the cursor to the next element of the array whose elements are read, which has idx elements
// before it, and returns the total offset of the end of the token before it, which the place of its errors is read
// again from, as the place of the element. It reports false after the end of the array.
func (s *Stream) nextElementV2(idx int) (int64, bool, error) {
	step := s.offset + s.cursor
	keeps, keepFrom := s.keepFromHere(step)
	ok, err := s.nextElement(idx == 0)
	s.keeps, s.keepFrom = keeps, keepFrom
	return step, ok, err
}

// decodeSliceElementsV2 decodes the array at the cursor into the slice at p, as v2SliceDecoder.Decode does, after
// from, the total offset of the end of the token before it.
func (s *Stream) decodeSliceElementsV2(d *v2SliceDecoder, p unsafe.Pointer, depth, from int64) error {
	if d.zero && (*sliceHeader)(p).cap > 0 {
		sv := valueAt(d.slicePtrType, p)
		sv.Slice(0, sv.Cap()).Clear()
	}
	depth, lv, err := s.openElementsV2(false, depth, from)
	if err != nil {
		return err
	}
	dst := (*sliceHeader)(p)
	// the elements are decoded into the slice up to limit, and into a buffer after it, as sliceDecoder.Decode does
	t := sliceTarget{d: d.sliceDecoder, dst: dst, limit: max(dst.cap, inPlaceSliceLength)}
	runs := readsWhole(d.valueDecoder)
	for idx := 0; ; {
		step, ok, err := s.nextElementV2(idx)
		if err != nil || !ok {
			t.finish(idx, idx, err)
			if err != nil {
				return s.placeAt(err, step)
			}
			s.popLevel(lv)
			return nil
		}
		if runs {
			n, err := s.decodeRunV2(&t, lv, idx, depth, step)
			if err != nil {
				t.finish(n, n+1, err)
				return err
			}
			if n > idx {
				idx = n
				continue
			}
		}
		if err := s.decodeNestedV2(d.valueDecoder, t.elem(idx), depth, step); err != nil {
			t.finish(idx, idx+1, err)
			return err
		}
		t.decoded(idx)
		idx++
		s.level(lv).Count = int64(idx)
	}
}

// decodeRunV2 decodes the elements of the array whose elements are read into t, from the element idx at the cursor,
// after step, the total offset of the end of the token before it: the ones which the buffer holds whole, up to the
// last one which is an object or an array that ends in what has been read, as decodeRun does, each as decodeWholeV2
// decodes it, without the lookup of its end. It returns the number of the elements of the array decoded, which the
// level lv of the path counts: idx if the buffer holds no such element. The decoder of the elements reads a value
// whole ( see readsWhole ).
func (s *Stream) decodeRunV2(t *sliceTarget, lv, idx int, depth, step int64) (int, error) {
	s.indexAt(depth)
	end, ok := s.elements.lastElementClose(s.offset+s.cursor, depth)
	if !ok {
		return idx, nil
	}
	ctx := s.ctx
	call := ctx.v2
	dec := t.d.valueDecoder
	// a nul byte is put after the last element while the run is decoded, as decodeRun puts it
	stop := end - s.offset + 1
	saved := s.buf[stop]
	s.buf[stop] = nul
	buf := s.buf[:stop+1]
	ctx.Buf, ctx.origin = buf, buf
	var err error
	for {
		call.Value = StreamValue{From: step - s.offset, Start: s.cursor}
		next, derr := dec.Decode(ctx, s.cursor, depth, t.elem(idx))
		if derr == nil && ctx.typeError != nil {
			derr = ctx.v2Error()
		}
		if derr != nil {
			ctx.DiscardTypeError()
			err = s.placeValue(derr, step, stop)
			break
		}
		t.decoded(idx)
		idx++
		s.level(lv).Count = int64(idx)
		s.cursor, step = next, s.offset+next
		// the next element is in the run if it starts before its end
		c := skipWhiteSpace(buf, next)
		if buf[c] != ',' {
			break
		}
		if c = skipWhiteSpace(buf, c+1); c >= stop {
			break
		}
		s.cursor = c
	}
	s.buf[stop] = saved
	ctx.Buf, ctx.origin = nil, nil
	call.Value = StreamValue{}
	return idx, err
}

// readsWhole reports whether decodeNestedV2 decodes a value which the buffer holds whole by dec as decodeWholeV2
// decodes it: unless dec, or the decoder of the value which dec decodes through pointers, is one of a method or a
// function which reads a jsontext.Decoder, which is given one attached at the value. The pointers of a type which
// points to itself, through other types or not, lead to no other decoder: slow, which follows them at half the
// pace, meets dec in them.
func readsWhole(dec Decoder) bool {
	slow := dec
	for odd := false; ; odd = !odd {
		switch d := dec.(type) {
		case *v2UnmarshalFromDecoder, *v2FuncDecoder:
			return false
		case *ptrDecoder:
			dec = d.dec
		case *basicPtrDecoder:
			dec = d.dec
		default:
			return true
		}
		if odd {
			slow = pointeeDecoder(slow)
		}
		if dec == slow {
			return true
		}
	}
}

// pointeeDecoder returns the decoder of the value which dec, a decoder of a pointer, decodes through it.
func pointeeDecoder(dec Decoder) Decoder {
	switch d := dec.(type) {
	case *ptrDecoder:
		return d.dec
	case *basicPtrDecoder:
		return d.dec
	}
	return dec
}

// decodeArrayElementsV2 decodes the array at the cursor into the array at p, as v2ArrayDecoder.Decode does, after
// from, the total offset of the end of the token before it.
func (s *Stream) decodeArrayElementsV2(d *v2ArrayDecoder, p unsafe.Pointer, depth, from int64) error {
	depth, lv, err := s.openElementsV2(false, depth, from)
	if err != nil {
		return err
	}
	var lengthErr error
	for idx := 0; ; idx++ {
		step, ok, err := s.nextElementV2(idx)
		if err != nil {
			return s.placeAt(err, step)
		}
		if !ok {
			for ; idx < d.alen; idx++ {
				d.elem.zero(unsafe.Add(p, uintptr(idx)*d.elem.size))
				lengthErr = errArrayUnderflow
			}
			if lengthErr != nil {
				// the error of the length is at the end of the array, which tells it
				return s.placeAt(&errors.UnmarshalError{GoType: d.typ, Start: s.cursor - 1, Err: lengthErr}, step)
			}
			s.popLevel(lv)
			return nil
		}
		if idx < d.alen {
			ep := unsafe.Add(p, uintptr(idx)*d.elem.size)
			d.elem.zero(ep)
			err = s.decodeNestedV2(d.valueDecoder, ep, depth, step)
		} else {
			err = s.skipValueV2(depth, step)
			lengthErr = errArrayOverflow
		}
		if err != nil {
			return err
		}
		s.level(lv).Count = int64(idx + 1)
	}
}

// decodeInterfaceElementsV2 decodes the object or the array at the cursor into the interface value at p, as
// v2InterfaceDecoder.Decode does, after from, the total offset of the end of the token before it. It reports false,
// having read nothing, for a value which is read whole.
func (s *Stream) decodeInterfaceElementsV2(d *v2InterfaceDecoder, p unsafe.Pointer, depth, from int64) (bool, error) {
	ctx := s.ctx
	if (*[2]unsafe.Pointer)(p)[0] != nil {
		// a value which the interface value holds is decoded into
		iface := valueAt(d.ptrType, p)
		held := iface.Elem()
		if !d.fromAny && (held.Type() == interfaceMapType || held.Type() == anySliceType) {
			return true, s.decodeMergingV2(d, p, depth, from)
		}
		v := reflect.New(held.Type())
		v.Elem().Set(held)
		dec, err := ctx.DecoderOfV2(ptrTypeOf(held.Type()))
		if err != nil {
			return true, err
		}
		err = s.decodeNestedV2(dec, v.UnsafePointer(), depth, from)
		iface.Set(v.Elem())
		return true, err
	}
	switch {
	case !d.empty:
		return true, s.placeAt(&errors.UnmarshalError{GoType: d.typ, Start: s.cursor, NoKind: true, Err: errNilInterface}, from)
	case d.fromAny:
		// the value of the type of its kind, which the functions of unmarshaling decode
		v := reflect.New(anyValueTypes[s.buf[s.cursor]])
		dec, err := ctx.DecoderOfV2(ptrTypeOf(v.Elem().Type()))
		if err != nil {
			return true, err
		}
		err = s.decodeNestedV2(dec, v.UnsafePointer(), depth, from)
		valueAt(d.ptrType, p).Set(v.Elem())
		return true, err
	}
	return true, s.decodeMergingV2(d, p, depth, from)
}

// decodeMergingV2 decodes the object or the array at the cursor into the interface{} at p, as decodeMerging does:
// into the map[string]interface{} or the []interface{} which it holds, or a new one.
func (s *Stream) decodeMergingV2(d *v2InterfaceDecoder, p unsafe.Pointer, depth, from int64) error {
	held := *(*any)(p)
	if s.buf[s.cursor] == '{' {
		m, ok := held.(map[string]any)
		if held != nil && !ok {
			return s.placeAt(&errors.UnmarshalError{GoType: reflect.TypeOf(held), Start: s.cursor}, from)
		}
		if m == nil {
			m = map[string]any{}
			*(*any)(p) = m
		}
		return s.decodeAnyObjectV2(d, m, depth, from)
	}
	a, ok := held.([]any)
	if held != nil && !ok {
		return s.placeAt(&errors.UnmarshalError{GoType: reflect.TypeOf(held), Start: s.cursor}, from)
	}
	err := s.decodeAnyArrayV2(d, &a, depth, from)
	*(*any)(p) = a
	return err
}

// decodeAnyObjectV2 decodes the object at the cursor into m, as decodeObjectInto does, after from, the total offset
// of the end of the token before it.
func (s *Stream) decodeAnyObjectV2(d *v2InterfaceDecoder, m map[string]any, depth, from int64) error {
	depth, lv, err := s.openElementsV2(true, depth, from)
	if err != nil {
		return err
	}
	ctx := s.ctx
	var names nameSet
	for k := 0; ; k++ {
		// the end of the token before the member, which the place of its errors is read again from, kept until the
		// value of the member
		step := s.offset + s.cursor
		keeps, keepFrom := s.keepFromHere(step)
		quoted, ok, err := s.nextMemberV2(k == 0)
		if err != nil {
			return s.placeAt(err, step)
		}
		if !ok {
			names.end(ctx.v2)
			s.keeps, s.keepFrom = keeps, keepFrom
			s.popLevel(lv)
			return nil
		}
		if ctx.Option.V2&V2AllowInvalidUTF8 == 0 && hasNotASCIIOrEscape(quoted) && !validV2Text(quoted) {
			return s.placeAt(errV2Text, step)
		}
		name := ctx.copyString(unquotedName(quoted))
		if ctx.Option.V2&V2AllowDuplicateNames == 0 && !names.addCopy(ctx.V2Call(), nil, unsafe.Slice(unsafe.StringData(name), len(name))) {
			start := s.cursor - int64(len(quoted))
			return s.placeAt(&errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}, step)
		}
		s.copyName(unsafe.Slice(unsafe.StringData(name), len(name)))
		nameEnd, err := s.memberValueV2()
		if err != nil {
			return s.placeAt(err, step)
		}
		s.setName(lv, k)
		s.keeps, s.keepFrom = keeps, keepFrom
		// the value of the name in the map is decoded into, as decodeElem does: the decoder of the interface{}
		// decodes a name of an object twice into its first value ( see v2InterfaceDecoder.Decode )
		value := m[name]
		err = s.decodeNestedV2(d, unsafe.Pointer(&value), depth, nameEnd)
		m[name] = value
		if err != nil {
			return err
		}
		s.level(lv).Count = int64(2*k + 2)
	}
}

// decodeAnyArrayV2 decodes the array at the cursor into *a from its first element, as decodeArrayInto does, after
// from, the total offset of the end of the token before it.
func (s *Stream) decodeAnyArrayV2(d *v2InterfaceDecoder, a *[]any, depth, from int64) error {
	depth, lv, err := s.openElementsV2(false, depth, from)
	if err != nil {
		return err
	}
	*a = (*a)[:0]
	if *a == nil {
		*a = []any{}
	}
	for idx := 0; ; idx++ {
		step, ok, err := s.nextElementV2(idx)
		if err != nil {
			return s.placeAt(err, step)
		}
		if !ok {
			s.popLevel(lv)
			return nil
		}
		*a = append(*a, nil)
		if err := s.decodeNestedV2(d, unsafe.Pointer(&(*a)[len(*a)-1]), depth, step); err != nil {
			return err
		}
		s.level(lv).Count = int64(idx + 1)
	}
}

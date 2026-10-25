package json

import (
	"io"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/decoder"
	"github.com/goccy/go-json/internal/options"
	iruntime "github.com/goccy/go-json/internal/runtime"
	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
)

// UnmarshalDecode deserializes a Go value from a jsontext.Decoder according to the provided unmarshal or decode
// options (while ignoring marshal or encode options). The options provided take precedence over options already
// applied on the jsontext.Decoder and only apply for the duration of the unmarshal call.
//
// The input may be a stream of zero or more JSON values. UnmarshalDecode unmarshals only the next JSON value in
// the stream. If there are no more top-level JSON values, it reports io.EOF. The output must be a non-nil pointer.
// See Unmarshal for details about the conversion of JSON into a Go value.
//
// The value is read from the input of the decoder as UnmarshalRead reads it, with bounded memory, or, for the decoder
// which an UnmarshalJSONFrom method or a function of UnmarshalFromFunc is given, as Unmarshal or UnmarshalRead reads
// it. After a SemanticError, the decoder goes on after the value, whose rest is read, and checked by its grammar, or,
// if the rest has an error, every read of the decoder fails with that error; but not after the error of a method or
// a function which read a jsontext.Decoder of the value from a reader. After any other error, the decoder is where it
// was before the value, as after an error of Decoder.ReadValue, but if the decoder read the value from a reader past
// what it had read itself and dropped the start of the value, which it reads with bounded memory: every read of the
// decoder fails with the error then.
func UnmarshalDecode(in *jsontext.Decoder, out any, opts ...Options) error {
	e := (*emptyInterface)(unsafe.Pointer(&out))
	err := unmarshalDecode(in, e.typ, e.ptr, opts)
	// the decoder refers to the value by its address only.
	runtime.KeepAlive(out)
	return err
}

// UnmarshalDecodeOf decodes the next value of a jsontext.Decoder into the value out points to, as UnmarshalDecode
// does, through a value in the heap which is reused, as UnmarshalOf does.
func UnmarshalDecodeOf[T any](in *jsontext.Decoder, out *T, opts ...Options) error {
	ptrType := reflect.TypeFor[*T]()
	if out == nil {
		return &SemanticError{action: "unmarshal", GoType: ptrType, Err: errNonNilReference}
	}
	ctx := decoder.TakeRuntimeContext()
	p := reusedValue(ctx, out)
	err := unmarshalDecode(in, iruntime.TypePtr(ptrType), p, opts)
	returnValue(p, out)
	decoder.ReleaseRuntimeContext(ctx)
	return err
}

// decodeCall is the state of a call of UnmarshalDecode, which the calls reuse: the levels of the place of the value,
// the input of the decoder which a stream reads, the stream, which a call takes once, and the config of the options
// of the call.
type decodeCall struct {
	levels []textcoder.Level
	input  lentInput
	stream *decoder.Stream
	cfg    options.Config
}

var decodeCalls = sync.Pool{New: func() any { return new(decodeCall) }}

// unmarshalDecode decodes the next value of in into the value at p, of the pointer type typ, by the options of the
// decoder and opts.
func unmarshalDecode(in *jsontext.Decoder, typ, p unsafe.Pointer, opts []Options) error {
	dc := decodeCalls.Get().(*decodeCall)
	err := dc.decode(in, typ, p, opts)
	// the levels and the input refer to the decoder, and the config may hold the functions of the caller, which the
	// pool must not keep alive
	clear(dc.levels)
	dc.levels = dc.levels[:0]
	dc.input = lentInput{}
	if dc.cfg.Set&configHolds != 0 {
		dc.cfg = options.Config{}
	} else {
		// the flags, which refer to nothing
		dc.cfg.Set, dc.cfg.Value = 0, 0
	}
	decodeCalls.Put(dc)
	return err
}

// decode is unmarshalDecode by the state of the call.
func (dc *decodeCall) decode(in *jsontext.Decoder, typ, p unsafe.Pointer, opts []Options) error {
	d := unsafe.Pointer(in)
	levels, name, outer, depth := textcoder.DecoderPlace(d, dc.levels[:0])
	dc.levels = levels
	cfg := &dc.cfg
	in.Options().ApplyTo(cfg)
	if len(opts) > 0 {
		if name {
			orig := *cfg
			cfg.Apply(opts)
			if err := optionsChange(&orig, cfg, levels[len(levels)-1]); err != nil && err != errChangingWhitespace {
				// the error before the name, which is not read
				start, berr := textcoder.BeginValue(d)
				if berr != nil {
					return berr
				}
				return &SemanticError{
					action: "unmarshal", ByteOffset: start, JSONPointer: jsontext.Pointer(textcoder.DecoderPointer(d, +1)),
					GoType: iruntime.TypeOfPtr(typ), Err: err,
				}
			}
		} else {
			cfg.Apply(opts)
		}
	}
	var err error
	switch st, attached := outer.(*fromState); {
	case name:
		err = decodeName(in, typ, p, cfg)
	case attached && st.att.In != nil && st.att.R == nil:
		err = decodeInPlace(d, in, st, levels, depth, typ, p, cfg)
	case attached && st.att.In != nil:
		err = decodeOnStream(d, in, st, levels, depth, typ, p, cfg)
	default:
		err = decodeLent(d, in, dc, levels, depth, typ, p, cfg)
	}
	return err
}

// decodeName decodes the name of an object which is the next value of the decoder, which reads it, into the value at
// p, of the pointer type typ, by the options of cfg, as Unmarshal decodes it.
func decodeName(in *jsontext.Decoder, typ, p unsafe.Pointer, cfg *options.Config) error {
	v, err := in.ReadValue()
	if err != nil {
		return err
	}
	ctx := decoder.TakeRuntimeContext()
	setCall(ctx, cfg)
	err = decodeInput(ctx, ctx.SetInput(v), typ, p, cfg)
	clearCall(ctx)
	ctx.Option.V2 = 0
	decoder.ReleaseRuntimeContext(ctx)
	if se, ok := err.(*SemanticError); ok {
		// the error of the name, which is at its place in the input of the decoder
		se.ByteOffset += in.InputOffset() - int64(len(v))
		se.JSONPointer = jsontext.Pointer(textcoder.DecoderPointer(unsafe.Pointer(in), -1))
	}
	return err
}

// decodeInPlace decodes the next value of the decoder at d, which st attached at a value which the input of a call
// holds whole, into the value at p, of the pointer type typ, by the options of cfg: the value is
// decoded in the input, where the decoder reads it, which goes on after it. levels and depth are the place of the
// value ( see textcoder.DecoderPlace ).
func decodeInPlace(d unsafe.Pointer, in *jsontext.Decoder, st *fromState, levels []textcoder.Level, depth int, typ, p unsafe.Pointer, cfg *options.Config) error {
	from := in.InputOffset()
	start, err := textcoder.BeginValue(d)
	if err != nil {
		return err
	}
	// the input with the nul byte after it, which the attach leaves out
	buf, base := st.att.In[:len(st.att.In)+1], st.att.Base
	ctx := decoder.TakeRuntimeContext()
	setCall(ctx, cfg)
	var next int64
	dec, err := outputDecoder(ctx, typ, p)
	if err == nil {
		decoder.SetBuf(ctx, buf)
		next, err = dec.Decode(ctx, start-base, int64(depth), p)
		if terr := ctx.V2TypeError(); terr != nil {
			err = terr
		}
		decoder.SetBuf(ctx, nil)
		ctx.DiscardTypeError()
	}
	if err == nil {
		textcoder.EndValue(d, start, base+next, nil)
	} else {
		// the error at its place, found again from the end of the token before the value, under its levels
		in := buf[from-base : len(buf)-1]
		err = unmarshalError(&errorInput{buf: buf[:len(buf)-1], base: base, att: decodePlace(in, from, levels, cfg, nil)}, err, cfg)
		err = afterError(d, err, start, buf[start-base], func() (int64, []byte, error) {
			return textcoder.SkipToEnd(decodePlace(in, from, levels, skipOptions, nil), depth, levels[depth].Count)
		}, nil)
	}
	clearCall(ctx)
	ctx.Option.V2 = 0
	decoder.ReleaseRuntimeContext(ctx)
	return err
}

// decodePlace returns the place of the next value of a decoder, whose levels are levels, for an error of the value
// or the rest of it: the input in after the end of the token before the value, at the offset base, and r, the reader
// of the input after in, or nil.
func decodePlace(in []byte, base int64, levels []textcoder.Level, cfg *options.Config, r io.Reader) *textcoder.DecoderAttachment {
	l := levels[len(levels)-1]
	return &textcoder.DecoderAttachment{
		In: in, Base: base, R: r, Object: l.Object, Top: len(levels) == 1, Count: l.Count,
		Outer: &streamPath{path: levels}, OuterDepth: len(levels) - 1, Opts: cfg,
	}
}

// skipOptions are the options of the skip of the rest of a value after a SemanticError: the grammar is checked, but
// not the text which v2 rejects, as names met twice, which the decoding may have reported first.
var skipOptions = func() *options.Config {
	c := new(options.Config)
	c.Apply([]Options{jsontext.AllowDuplicateNames(true), jsontext.AllowInvalidUTF8(true)})
	return c
}()

// afterError leaves the decoder at d where UnmarshalDecode leaves it after err, the error of its next value, which
// starts at the offset start, and begins with the byte first: after a SemanticError, after the value, which skip
// reads to its end, or, if the rest of the value has an error, failing every read with it; after a SyntacticError of
// an object or an array, failing every read with it, as encoding/json/v2 reads them token by token and stops at the
// error, so that a loop over the values of the input doesn't decode the value again; else, or if skip is nil, before
// the value, which putBack puts back unless it is nil, the decoder having its input. If putBack can't put the input
// back, every read fails with err. It returns err.
func afterError(d unsafe.Pointer, err error, start int64, first byte, skip func() (int64, []byte, error), putBack func() bool) error {
	switch err.(type) {
	case *SemanticError:
		if skip == nil {
			break
		}
		end, rest, serr := skip()
		if serr != nil {
			// the value has no end, and the skip may have read on from the reader of the decoder
			textcoder.FailDecoder(d, serr)
		} else {
			textcoder.EndValue(d, start, end, rest)
		}
		return err
	case *jsontext.SyntacticError:
		if first == '{' || first == '[' {
			textcoder.FailDecoder(d, err)
			return err
		}
	}
	if putBack != nil && !putBack() {
		textcoder.FailDecoder(d, err)
	}
	return err
}

// lentInput is the input of a decoder which a stream reads: the bytes of its buffer, and then its reader, which the
// stream reads once it read them all.
type lentInput struct {
	buf []byte
	r   io.Reader
	// readR is set once r was read.
	readR bool
}

func (l *lentInput) Read(p []byte) (int, error) {
	if len(l.buf) > 0 {
		n := copy(p, l.buf)
		l.buf = l.buf[n:]
		return n, nil
	}
	l.readR = true
	return l.r.Read(p)
}

// decodeLent decodes the next value of the decoder at d, of in, into the value at p, of the pointer type typ, by the
// options of cfg: the stream of dc reads the input of the decoder by its input, from the end of the token before
// the value, as UnmarshalRead reads it, and the decoder goes on after the value. levels and depth are the place of
// the value ( see textcoder.DecoderPlace ).
func decodeLent(d unsafe.Pointer, in *jsontext.Decoder, dc *decodeCall, levels []textcoder.Level, depth int, typ, p unsafe.Pointer, cfg *options.Config) error {
	from := in.InputOffset()
	start, err := textcoder.BeginValue(d)
	if err != nil {
		return err
	}
	buf, r := textcoder.DecoderInput(d, from)
	// the first byte of the value, which BeginValue read
	first := buf[start-from]
	input := &dc.input
	s := dc.stream
	if s == nil {
		// the stream of the call, as the pool of TakeStreamV2 makes one
		s = decoder.NewStream(nil)
		dc.stream = s
	}
	s.BeginV2(input)
	// the buffer of the decoder, which the stream reads in place if it can, and its offset, by which the decoder is
	// found to have dropped bytes of its input for the stream
	borrowed := s.BorrowV2(d, from)
	var base int64
	if borrowed {
		*input = lentInput{r: r}
		_, base = s.Read()
	} else {
		*input = lentInput{buf: buf, r: r}
		s.StartAt(from)
	}
	ctx := s.Context()
	s.SetPath(ctx.V2Call().Path, levels)
	setCall(ctx, cfg)
	err = decodeStreamValue(s, typ, p, int64(depth), from, start)
	// moved is set if the decoder has the input of the value no more where it was: the stream read it from the
	// reader of the decoder, or through the decoder into its buffer, which moved its bytes to read more, dropping the
	// ones before the ones which the stream kept. The decoder goes on from the rest which the stream read then. The
	// bytes which the stream read through the decoder into the room after its input are where the decoder reads them.
	moved := input.readR
	if borrowed && !moved {
		// the offset of the buffer of the decoder: the one of what the stream read, while it reads the buffer in
		// place, or the decoder's, once the stream left the buffer for its own, which a decoder attached at a value
		// reads on from ( see decoder.Stream.ReadOn )
		_, b0 := s.Read()
		if !s.ReadsInPlace() {
			_, b0, _ = textcoder.DecoderBuffer(d)
		}
		moved = b0 != base
	}
	if err == nil {
		var rest []byte
		if moved {
			// the input after the value, which the stream read
			rest = s.Unread()
		}
		textcoder.EndValue(d, start, s.Offset(), rest)
	} else {
		err = streamError(s, err, cfg)
		var putBack func() bool
		if moved {
			// the stream read past the input of the decoder, which has it no more where it was
			putBack = func() bool { return putBackFrom(d, s, from) }
		}
		err = afterError(d, err, start, first, streamSkip(s, moved, depth, levels[depth].Count), putBack)
	}
	clearCall(ctx)
	s.EndV2()
	return err
}

// putBackFrom gives the decoder at d the input of the stream from the offset from, where it was before its next
// value, which the stream read and failed at, and reports false if the stream doesn't have it.
func putBackFrom(d unsafe.Pointer, s *decoder.Stream, from int64) bool {
	buf, base := s.Read()
	if from < base {
		return false
	}
	textcoder.PutBack(d, from, buf[from-base:])
	return true
}

// decodeStreamValue decodes the value of the stream at the offset start, after from, the end of the token before it,
// nested in depth objects and arrays, into the value at p, of the pointer type typ.
func decodeStreamValue(s *decoder.Stream, typ, p unsafe.Pointer, depth, from, start int64) error {
	dec, err := outputDecoder(s.Context(), typ, p)
	if err != nil {
		return err
	}
	return s.DecodeV2After(dec, p, depth, from, start)
}

// streamSkip returns the skip of the rest of the value of a stream whose decoding failed, from the place of its error
// ( see decoder.StreamPlace ), nested in depth objects and arrays which have count tokens before the value, or nil
// if the error has no place, the error of a method or a function which read the value from a reader. The skip
// returns the input after the value which was read past the input of the decoder, which read is set for if the stream
// read past it, as the decoder doesn't have it.
func streamSkip(s *decoder.Stream, read bool, depth int, count int64) func() (int64, []byte, error) {
	call := s.Context().V2Call()
	pl := call.Place
	if !pl.Set {
		return nil
	}
	buf, base := s.Read()
	return func() (int64, []byte, error) {
		place := decodePlace(buf[pl.From:], base+pl.From, call.Path, skipOptions, &streamReader{s: s})
		end, rest, err := textcoder.SkipToEnd(place, depth, count)
		if err == nil && rest == nil && read {
			// the end is in what the stream read
			rest = buf[end-base:]
		}
		return end, rest, err
	}
}

// decodeOnStream decodes the next value of the decoder at d, which st attached at a value of a stream which reads on
// from the reader of the stream, into the value at p, of the pointer type typ, by the options of cfg:
// the stream takes the input of the decoder back, decodes the value, and gives the decoder the input after it. levels
// and depth are the place of the value ( see textcoder.DecoderPlace ).
func decodeOnStream(d unsafe.Pointer, in *jsontext.Decoder, st *fromState, levels []textcoder.Level, depth int, typ, p unsafe.Pointer, cfg *options.Config) error {
	from := in.InputOffset()
	start, err := textcoder.BeginValue(d)
	if err != nil {
		return err
	}
	s := st.stream
	buf, _ := textcoder.DecoderInput(d, from)
	// the first byte of the value, which BeginValue read
	first := buf[start-from]
	s.Resume(from, buf)
	// the state of the call of the method, which the call of the value sets aside: its options, its path, its value
	// and the place of its error, and its state of the methods, as a method of the value is given a decoder of its
	// own
	ctx := s.Context()
	call := ctx.V2Call()
	savedFuncs, savedConfig, savedValue, savedPlace, savedFrom := call.Funcs, call.Config, call.Value, call.Place, call.From
	savedOption := *ctx.Option
	if st.nested == nil {
		st.nested = new(fromState)
		st.nested.tracked.stale = true
	}
	call.From = unsafe.Pointer(st.nested)
	savedPath := s.SetPath(st.nestedPath, levels)
	setCall(ctx, cfg)
	err = decodeStreamValue(s, typ, p, int64(depth), from, start)
	if err == nil {
		textcoder.EndValue(d, start, s.Offset(), s.Unread())
	} else {
		err = streamError(s, err, cfg)
		err = afterError(d, err, start, first, streamSkip(s, true, depth, levels[depth].Count), func() bool {
			// the decoder gave the stream its input
			return putBackFrom(d, s, from)
		})
		if textcoder.DecoderFailed(d) {
			// the stream, which the decoder reads, lost its input, or the start of the value: it gives nothing back
			// to the decoders before it
			s.Drop()
		}
	}
	clearCall(ctx)
	st.nestedPath = s.RestorePath(savedPath)
	call.Funcs, call.Config, call.Value, call.Place, call.From = savedFuncs, savedConfig, savedValue, savedPlace, savedFrom
	*ctx.Option = savedOption
	return err
}

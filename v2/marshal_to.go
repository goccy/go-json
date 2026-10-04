package json

import (
	"errors"
	"math"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
)

// A MarshalJSONTo method and a function of MarshalToFunc write to a jsontext.Encoder, which the encoder gives
// them after its output so far: the encoder is set at the place of the output, which is found from the output
// itself, so that its stack and its errors point where the value is.

func init() {
	encoder.V2Hooks.MarshalerToType = reflect.TypeFor[MarshalerTo]()
	encoder.V2Hooks.AppendRaw = appendRaw
	encoder.V2Hooks.MarshalTo = func(ctx *encoder.RuntimeContext, b []byte, typ reflect.Type, recv any) ([]byte, error) {
		m, _ := recv.(MarshalerTo)
		return marshalTo(ctx, b, typ, m, nil)
	}
	encoder.V2Hooks.MarshalToOf = marshalToOf
}

// ifaceWords are the words of an interface value of a non-empty interface.
type ifaceWords struct {
	tab  unsafe.Pointer
	data unsafe.Pointer
}

// marshalToOf returns the function which calls MarshalJSONTo of the value of typ at an address, or appendDefault
// if it declines. The interface value of the method is made from the address and the table of the methods of the
// pointer to typ, which is found once: the data word of the pointer is the address itself.
func marshalToOf(typ reflect.Type, appendDefault encoder.AppendFunc) encoder.AppendFunc {
	proto, _ := reflect.New(typ).Interface().(MarshalerTo)
	tab := (*ifaceWords)(unsafe.Pointer(&proto)).tab
	return func(ctx *encoder.RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
		var m MarshalerTo
		w := (*ifaceWords)(unsafe.Pointer(&m))
		w.tab, w.data = tab, p
		out, err := marshalTo(ctx, b, typ, m, nil)
		if err == encoder.ErrUseDefault {
			return appendDefault(ctx, b, p)
		}
		return out, err
	}
}

// callState is the state of a call of Marshal, MarshalWrite or MarshalEncode, which the encoder holds in its
// options.
type callState struct {
	// cfg are the options of the call, and orig the ones of the jsontext.Encoder of MarshalEncode.
	cfg, orig options.Config
	// outerEnc is the jsontext.Encoder which MarshalEncode writes to, or nil for an output of its own, and place
	// and placeTop are its innermost level ( see textcoder.Place ). outer are the levels of the output where the
	// value is written, the top level first, and ptr is its JSON pointer: they are found when they are needed, for
	// an error or the stack of an encoder, if outerKnown is not set. base is the offset of the value in that
	// output.
	outerEnc   *jsontext.Encoder
	outerPlace textcoder.Level
	placeTop   bool
	outer      []textcoder.Level
	ptr        jsontext.Pointer
	outerKnown bool
	base       int64
	// attachCtx and attachOut are the context and the output of the call of a method which writes to enc, whose
	// levels Levels returns.
	attachCtx *encoder.RuntimeContext
	attachOut []byte
	// attachSortedName is whether the method writes the name of a key of a sorted map, whose entries before it
	// are not in their place yet: the names of the map are not given.
	attachSortedName bool
	// attached is whether enc was attached in the call, with its options, and place is where it is attached.
	attached bool
	place    textcoder.Attachment
	// opened is whether an encoding which failed stopped in an object or an array it opened.
	opened bool
	// enc is the encoder which the methods and the functions write to, at the place of the levels.
	enc     jsontext.Encoder
	levels  []textcoder.Level
	tracked tracker
}

// setPlace sets the place of the output where the value is written: of enc, whose innermost level is inner, or
// the top level of an output of its own if enc is nil. A pointer is written only if it changes, as a write of it
// costs a write barrier while the GC marks.
func (st *callState) setPlace(enc *jsontext.Encoder, inner textcoder.Level, top bool) {
	if st.outerEnc != enc {
		st.outerEnc = enc
	}
	st.outerPlace, st.placeTop, st.outerKnown = inner, top, false
}

// outerLevels returns the levels of the output where the value is written ( see callState.outer ).
func (st *callState) outerLevels() []textcoder.Level {
	if !st.outerKnown {
		if st.outerEnc == nil {
			st.outer = append(st.outer[:0], textcoder.Level{})
			st.ptr = ""
		} else {
			st.outer, _ = textcoder.Position(st.outerEnc, st.outer[:0])
			st.ptr = pointerOf(st.outer, +1)
		}
		st.outerKnown = true
	}
	return st.outer
}

// at returns the offset and the JSON pointer of the place after out, the output of the encoder so far, in the
// whole output.
func (st *callState) at(out []byte) (int64, jsontext.Pointer) {
	st.outerLevels()
	return st.base + int64(len(out)), st.ptr + nextPointer(out)
}

// Levels returns the levels of the place where enc is attached for a method ( see textcoder.Outer ).
func (st *callState) Levels() []textcoder.Level {
	levels := st.trackedLevels(st.attachCtx, st.attachOut)
	if st.attachSortedName {
		levels[len(levels)-1].Name = nil
	}
	return levels
}

// levelsOf returns the levels of the whole output after out, the output of the encoder so far ( see levelsOf ):
// the ones of the output where the value is written, and the ones of out within it.
func (st *callState) levelsOf(out []byte) []textcoder.Level {
	return st.withOuter(levelsOf(out, nil))
}

// trackedLevels is levelsOf of the output of ctx, which is read from where it was read the last time: the output
// grows, and is written again only where ctx records it.
func (st *callState) trackedLevels(ctx *encoder.RuntimeContext, out []byte) []textcoder.Level {
	inner := st.tracked.sync(out, ctx.RewriteFrom)
	ctx.RewriteFrom = math.MaxInt
	return st.withOuter(inner)
}

// withOuter returns the levels of the output where the value is written, followed by inner, the levels of the
// output of the encoder within it.
func (st *callState) withOuter(inner []textcoder.Level) []textcoder.Level {
	levels := append(st.levels[:0], st.outerLevels()...)
	if len(levels) == 0 {
		levels = append(levels, textcoder.Level{})
	}
	levels[len(levels)-1].Count += inner[0].Count
	st.levels = append(levels, inner[1:]...)
	return st.levels
}

// tracker follows the levels of the output of the encoder as it grows, so that a method which writes to an
// encoder is given its place without reading the whole output again: the levels after the first n bytes, and the
// levels at some offsets before, to read the output again from after a part of it was written again.
type tracker struct {
	n      int
	levels []textcoder.Level
	// marks are the places where the levels were kept, in markLevels: the levels of the marks one after another.
	marks      []trackMark
	markLevels []textcoder.Level
	// stale is whether the tracker was reset, which it is made ready for when it is used: most calls don't.
	stale bool
	// used is whether the tracker was used since it was released: its levels refer to the output.
	used bool
}

// trackMark is the place n, whose levels are markLevels[start:end].
type trackMark struct {
	n, start, end int
}

// checkTracked is called with the levels which the tracker followed, by a test which compares them with the
// ones of the whole output.
var checkTracked func(out []byte, levels []textcoder.Level)

// trackMarkInterval is the length of the output between the marks: what is read again at most after a rewrite.
const trackMarkInterval = 4 << 10

// sync returns the levels of out, which was written again from rewriteFrom if that is within what was read. The
// names of the levels are in out, which is not written again before them ( see scanLevels ): the levels are
// copied without a copy of their names.
func (t *tracker) sync(out []byte, rewriteFrom int) []textcoder.Level {
	t.used = true
	if t.stale {
		t.n, t.stale = 0, false
		t.levels = append(t.levels[:0], textcoder.Level{})
		t.marks, t.markLevels = t.marks[:0], t.markLevels[:0]
	}
	if at := min(rewriteFrom, len(out)); at < t.n {
		for len(t.marks) > 0 && t.marks[len(t.marks)-1].n > at {
			t.markLevels = t.markLevels[:t.marks[len(t.marks)-1].start]
			t.marks = t.marks[:len(t.marks)-1]
		}
		if len(t.marks) == 0 {
			t.n = 0
			t.levels = append(t.levels[:0], textcoder.Level{})
		} else {
			m := t.marks[len(t.marks)-1]
			t.n = m.n
			t.levels = append(t.levels[:0], t.markLevels[m.start:m.end]...)
		}
	}
	t.levels = scanLevels(out[t.n:], t.levels)
	t.n = len(out)
	last := 0
	if len(t.marks) > 0 {
		last = t.marks[len(t.marks)-1].n
	}
	if t.n-last >= trackMarkInterval {
		start := len(t.markLevels)
		t.markLevels = append(t.markLevels, t.levels...)
		t.marks = append(t.marks, trackMark{n: t.n, start: start, end: len(t.markLevels)})
	}
	if checkTracked != nil {
		checkTracked(out, t.levels)
	}
	return t.levels
}

// release drops the references of the levels to the output, which the state of a call, which is kept, would keep
// otherwise.
func (t *tracker) release() {
	if t.used {
		clear(t.levels[:cap(t.levels)])
		clear(t.markLevels[:cap(t.markLevels)])
		t.used = false
	}
}

// takeCallState returns the state for a call which encodes by ctx, whose options are opts: the one kept with ctx,
// which a call uses at a time. The state holds the options: they are applied by the calls of their interface,
// which a value on the stack would escape by.
func takeCallState(ctx *encoder.RuntimeContext, opts []Options) *callState {
	if st := (*callState)(ctx.V2State); st != nil && len(opts) == 0 {
		return st
	}
	// a new state, or options: a call, which most calls don't make.
	return newCallState(ctx, opts)
}

func newCallState(ctx *encoder.RuntimeContext, opts []Options) *callState {
	st := (*callState)(ctx.V2State)
	if st == nil {
		st = new(callState)
		ctx.V2State = unsafe.Pointer(st)
	}
	st.cfg.Apply(opts)
	return st
}

// releaseCallState ends the call of the state: the options, which may hold the functions of the caller, are
// cleared if they were set.
func releaseCallState(st *callState) {
	st.tracked.release()
	if st.attachOut != nil {
		st.attachCtx, st.attachOut, st.place.Out = nil, nil, nil
	}
	if st.outerEnc != nil {
		st.outerEnc = nil
	}
	if st.cfg.Set != 0 {
		st.cfg = options.Config{}
	}
	if st.orig.Set != 0 {
		st.orig = options.Config{}
	}
}

// stateOf returns the state of the call which ctx encodes for.
func stateOf(ctx *encoder.RuntimeContext) *callState {
	return ctx.Option.V2.(*callState)
}

// marshalTo calls MarshalJSONTo of m, or fn, a function of MarshalToFunc, for a value of typ, which writes the
// value after b, the output so far. b ends with the delimiter before the value, if any, which the encoder writes
// itself.
func marshalTo(ctx *encoder.RuntimeContext, b []byte, typ reflect.Type, m MarshalerTo, fn func(*jsontext.Encoder) error) ([]byte, error) {
	st := stateOf(ctx)
	out := b
	// the place of the value, found from the delimiter before it, which the encoder writes itself: the levels
	// around it are found only if the encoder needs them ( see callState.Levels ).
	pl := &st.place
	pl.Object, pl.Top, pl.Count = st.outerPlace.Object, st.placeTop, st.outerPlace.Count
	if n := len(out); n > 0 {
		switch out[n-1] {
		case ':':
			// the value of a member.
			out, pl.Object, pl.Top, pl.Count = out[:n-1], true, false, 1
		case ',':
			if ctx.KeyName {
				// the name of an entry of a map after another one.
				out, pl.Object, pl.Top, pl.Count = out[:n-1], true, false, 2
			} else {
				// an element of an array after another one.
				out, pl.Object, pl.Top, pl.Count = out[:n-1], false, false, 1
			}
		case '{':
			pl.Object, pl.Top, pl.Count = true, false, 0
		case '[':
			pl.Object, pl.Top, pl.Count = false, false, 0
		default:
			// a value after a value, at the top level, which the encoder doesn't write.
			return b, errNonSingularValue
		}
	}
	// the pointers are written only if they change, as a write of a pointer costs a write barrier while the GC
	// marks: the output is mostly the same array, whose length is set by a slice of itself.
	if st.attachCtx != ctx {
		st.attachCtx = ctx
	}
	if unsafe.SliceData(st.attachOut) == unsafe.SliceData(out) && cap(st.attachOut) == cap(out) {
		st.attachOut = st.attachOut[:len(out)]
	} else {
		st.attachOut = out
	}
	st.attachSortedName = ctx.KeyName && ctx.Option.Flag&encoder.UnorderedMapOption == 0
	// the levels around the place, which the call gives for any place: at the top level, the encoder has them.
	if unsafe.SliceData(pl.Out) == unsafe.SliceData(out) && cap(pl.Out) == cap(out) {
		pl.Out = pl.Out[:len(out)]
	} else {
		pl.Out = out
	}
	pl.Base, pl.Same = st.base, st.attached
	if !st.attached {
		pl.Outer, pl.Opts = st, &st.cfg
	}
	textcoder.Attach(unsafe.Pointer(&st.enc), pl)
	st.attached = true
	depth, count, skip := pl.Depth, pl.Count, pl.Skip
	var err error
	if m != nil {
		// called by the interface, without the method value, which would be one more call.
		err = m.MarshalJSONTo(&st.enc)
	} else {
		err = fn(&st.enc)
	}
	textcoder.Detach(unsafe.Pointer(&st.enc), pl)
	written, newDepth, newCount := pl.Out, pl.Depth, pl.Count
	if skip > 0 && len(written) >= skip {
		// the delimiter before the value, which the encoder of MarshalEncode writes.
		written = written[:copy(written, written[skip:])]
	}
	if err == nil && (newDepth != depth || newCount != count+1) {
		err = errNonSingularValue
	}
	if err != nil {
		if errors.Is(err, errors.ErrUnsupported) {
			if newDepth == depth && newCount == count {
				return b, encoder.ErrUseDefault
			}
			err = errUnsupportedMutation
		}
		return b, &ierrors.MethodError{GoType: typ, Err: err, Kind: ierrors.MethodJSONTo, Out: written,
			Where: whereOf(depth, count, newDepth, newCount)}
	}
	return written, nil
}

// whereOf is where an error of a method which wrote to the encoder points: to the next value if it wrote
// nothing, to the value it wrote if it wrote one at the place of its value, or to the level it stopped in.
func whereOf(depth int, count int64, newDepth int, newCount int64) int8 {
	switch {
	case depth == newDepth && count == newCount:
		return +1
	case depth == newDepth && count+1 == newCount:
		return -1
	}
	return 0
}

// appendRaw appends the raw value which a method or a function returned, after b: checked and formatted as the
// options want it. Its error points to the place in the whole output.
func appendRaw(ctx *encoder.RuntimeContext, b, raw []byte) ([]byte, error) {
	st := stateOf(ctx)
	if st.cfg.Value&rawRewriting == 0 {
		if out, ok := encoder.AppendFormattedRaw(ctx, b, raw); ok {
			return out, nil
		}
	}
	// raw is not changed: it is the memory of the caller, as the value of a json.RawMessage.
	out, err := jsontext.AppendFormat(b, raw, &st.cfg, compact)
	if err != nil {
		if serr, ok := err.(*jsontext.SyntacticError); ok {
			// the place in the value is after the place of the value in the output.
			pos, ptr := st.at(b)
			at := *serr
			at.ByteOffset += pos
			at.JSONPointer = ptr + serr.JSONPointer
			return b, &at
		}
		return b, err
	}
	return out, nil
}

// rawRewriting are the options of the formatting of a raw value which encoder.AppendFormattedRaw doesn't follow.
const rawRewriting = options.CanonicalizeRawInts | options.CanonicalizeRawFloats | options.ReorderRawObjects |
	options.PreserveRawStrings

// compact is the format of the output of the encoder, which formats it as the options want it after.
var compact = &options.Config{
	Set: options.Multiline | options.SpaceAfterColon | options.SpaceAfterComma,
}

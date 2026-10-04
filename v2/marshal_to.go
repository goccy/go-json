package json

import (
	"errors"
	"reflect"
	"sync"

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
		return marshalTo(ctx, b, typ, m.MarshalJSONTo)
	}
}

// callState is the state of a call of Marshal, MarshalWrite or MarshalEncode, which the encoder holds in its
// options.
type callState struct {
	cfg *options.Config
	// outer are the levels of the output where the value is written, the top level first: of the jsontext.Encoder
	// which MarshalEncode writes to, or the top level only. base is the offset of the value in that output, and
	// ptr is its JSON pointer.
	outer []textcoder.Level
	base  int64
	ptr   jsontext.Pointer
	// opened is whether an encoding which failed stopped in an object or an array it opened.
	opened bool
	// enc is the encoder which the methods and the functions write to.
	enc    jsontext.Encoder
	levels []textcoder.Level
}

// reset sets the state for a call which writes at the place of the output which outer and base are of.
func (st *callState) reset(c *options.Config, outer []textcoder.Level, base int64) {
	st.cfg, st.base, st.opened = c, base, false
	if outer == nil {
		st.outer = append(st.outer[:0], textcoder.Level{})
		st.ptr = ""
		return
	}
	st.outer = append(st.outer[:0], outer...)
	st.ptr = pointerOf(outer, +1)
}

// at returns the offset and the JSON pointer of the place after out, the output of the encoder so far, in the
// whole output.
func (st *callState) at(out []byte) (int64, jsontext.Pointer) {
	return st.base + int64(len(out)), st.ptr + nextPointer(out)
}

// levelsOf returns the levels of the whole output after out, the output of the encoder so far ( see levelsOf ):
// the ones of the output where the value is written, and the ones of out within it.
func (st *callState) levelsOf(out []byte) []textcoder.Level {
	levels := append(st.levels[:0], st.outer...)
	inner := levelsOf(out, nil)
	levels[len(levels)-1].Count += inner[0].Count
	st.levels = append(levels, inner[1:]...)
	return st.levels
}

var callStates = sync.Pool{New: func() any { return new(callState) }}

// stateOf returns the state of the call which ctx encodes for.
func stateOf(ctx *encoder.RuntimeContext) *callState {
	return ctx.Option.V2.(*callState)
}

// marshalTo calls fn, a MarshalJSONTo method or a function of MarshalToFunc for a value of typ, which writes the
// value after b, the output so far. b ends with the delimiter before the value, if any, which the encoder writes
// itself.
func marshalTo(ctx *encoder.RuntimeContext, b []byte, typ reflect.Type, fn func(*jsontext.Encoder) error) ([]byte, error) {
	st := stateOf(ctx)
	out := b
	if n := len(out); n > 0 && (out[n-1] == ',' || out[n-1] == ':') {
		out = out[:n-1]
	}
	st.levels = st.levelsOf(out)
	depth, count := len(st.levels)-1, st.levels[len(st.levels)-1].Count
	if top := &st.levels[depth]; top.Object && count%2 == 0 && ctx.Option.Flag&encoder.UnorderedMapOption == 0 {
		// the name of a key of a sorted map, whose names are checked after the entries are sorted.
		top.Names = nil
	}
	textcoder.Attach(&st.enc, out, st.base, st.levels, st.cfg)
	err := fn(&st.enc)
	written, newDepth, newCount := textcoder.Detach(&st.enc)
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
	v := jsontext.Value(raw)
	if err := v.Format(st.cfg, compact); err != nil {
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
	return append(b, v...), nil
}

// compact is the format of the output of the encoder, which formats it as the options want it after.
var compact = &options.Config{
	Set: options.Multiline | options.SpaceAfterColon | options.SpaceAfterComma,
}

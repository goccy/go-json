//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The stack of the encoder which a MarshalJSONTo method is given, which the encoder finds only when the method
// asks for it, is the one which encoding/json/v2 gives: the methods log it, for the same values.

// stackLog is what the methods saw, one line for each call.
type stackLog struct{ lines []string }

func (l *stackLog) add(ptr string, depth int, kinds []string) {
	l.lines = append(l.lines, fmt.Sprintf("%s %d %s", ptr, depth, strings.Join(kinds, ",")))
}

// stdRecorder and recorder write their value after they log the stack, the first one to encoding/json/v2, the
// second one to the v2 json package. nested is whether the value is written by MarshalEncode, or by WriteValue
// of its JSON.
type stdRecorder struct {
	V      any
	log    *stackLog
	nested bool
	// invalid makes the method write a value which is not valid JSON.
	invalid bool
}

func (r stdRecorder) MarshalJSONTo(enc *stdjsontext.Encoder) error {
	if r.invalid {
		return enc.WriteValue([]byte(`{"a":1,"a":2}`))
	}
	var kinds []string
	for i := range enc.StackDepth() + 1 {
		k, n := enc.StackIndex(i)
		kinds = append(kinds, fmt.Sprintf("%v:%d", k, n))
	}
	r.log.add(string(enc.StackPointer()), enc.StackDepth(), kinds)
	if r.nested {
		return stdjson.MarshalEncode(enc, r.V, stdjson.Deterministic(true))
	}
	b, err := stdjson.Marshal(r.V, stdjson.Deterministic(true))
	if err != nil {
		return err
	}
	return enc.WriteValue(b)
}

type recorder struct {
	V       any
	log     *stackLog
	nested  bool
	invalid bool
}

func (r recorder) MarshalJSONTo(enc *jsontext.Encoder) error {
	if r.invalid {
		return enc.WriteValue([]byte(`{"a":1,"a":2}`))
	}
	var kinds []string
	for i := range enc.StackDepth() + 1 {
		k, n := enc.StackIndex(i)
		kinds = append(kinds, fmt.Sprintf("%v:%d", k, n))
	}
	r.log.add(string(enc.StackPointer()), enc.StackDepth(), kinds)
	if r.nested {
		return json.MarshalEncode(enc, r.V, json.Deterministic(true))
	}
	b, err := json.Marshal(r.V, json.Deterministic(true))
	if err != nil {
		return err
	}
	return enc.WriteValue(b)
}

// recorderStruct has members around a recorder, of names which need escapes.
type recorderStruct[R any] struct {
	A   int `json:"a~b/c"`
	R   R   `json:"r"`
	Arr []R `json:"arr"`
	Z   string
}

func TestMarshalToStackSameAsStd(t *testing.T) {
	for seed := range uint64(300) {
		var stdLog, ourLog stackLog
		// the same random values, of the recorders of each package.
		build := func(mk func(v any, nested, invalid bool) any) any {
			r := rand.New(rand.NewPCG(seed, 1))
			var value func(depth int) any
			value = func(depth int) any {
				n := r.IntN(6)
				if depth > 3 {
					n = 0
				}
				switch n {
				case 0:
					return mk(r.IntN(10), r.IntN(2) == 0, r.IntN(20) == 0)
				case 1:
					m := map[string]any{}
					for i := range r.IntN(4) {
						m[fmt.Sprintf("k%d", i)] = value(depth + 1)
					}
					return m
				case 2:
					list := make([]any, r.IntN(4))
					for i := range list {
						list[i] = value(depth + 1)
					}
					return list
				case 3:
					return mk(value(depth+1), r.IntN(2) == 0, false)
				default:
					return map[string]any{"s": strings.Repeat("x", r.IntN(5000)), "v": value(depth + 1), "l": []any{1, value(depth + 1)}}
				}
			}
			return value(0)
		}
		stdIn := build(func(v any, nested, invalid bool) any {
			return stdRecorder{V: v, log: &stdLog, nested: nested, invalid: invalid}
		})
		ourIn := build(func(v any, nested, invalid bool) any {
			return recorder{V: v, log: &ourLog, nested: nested, invalid: invalid}
		})
		stdIn = recorderStruct[any]{A: 1, R: stdIn, Arr: []any{stdIn, 2}, Z: "z"}
		ourIn = recorderStruct[any]{A: 1, R: ourIn, Arr: []any{ourIn, 2}, Z: "z"}
		want, wantErr := stdjson.Marshal(stdIn, stdjson.Deterministic(true))
		got, err := json.Marshal(ourIn, json.Deterministic(true))
		if (err == nil) != (wantErr == nil) {
			t.Fatalf("%d: Marshal error = %v; std %v", seed, err, wantErr)
		}
		if err != nil {
			// the place of the error of the method, which the encoder reports.
			var serr *jsontext.SyntacticError
			var stdSerr *stdjsontext.SyntacticError
			if !errors.As(err, &serr) || !errors.As(wantErr, &stdSerr) ||
				serr.ByteOffset != stdSerr.ByteOffset || string(serr.JSONPointer) != string(stdSerr.JSONPointer) {
				t.Fatalf("%d: Marshal error = %v; std %v", seed, err, wantErr)
			}
			continue
		}
		if string(got) != string(want) {
			t.Fatalf("%d: Marshal:\ngot  %s\nwant %s", seed, got, want)
		}
		if g, w := strings.Join(ourLog.lines, "\n"), strings.Join(stdLog.lines, "\n"); g != w {
			t.Fatalf("%d: the stacks which the methods saw:\ngot\n%s\nwant\n%s", seed, g, w)
		}
	}
}

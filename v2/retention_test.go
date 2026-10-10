package json

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/jsontext"
)

// callerName is the name of the object of the caller of MarshalEncode.
const callerName = "caller-name."

// stackReader reads the stack of its encoder, which has the levels of the encoder of MarshalEncode.
type stackReader struct{}

func (stackReader) MarshalJSONTo(e *jsontext.Encoder) error {
	_ = e.StackPointer()
	return e.WriteToken(jsontext.Int(1))
}

// The state of a call of MarshalEncode keeps nothing of the output of its encoder for the pool: the names of the
// objects it is in, which a method reads by the stack of its encoder, may be large. The names of the output of the
// call itself are of the output of the context, which the pool keeps anyway.
func TestMarshalEncodeKeepsNoOutput(t *testing.T) {
	for name, v := range map[string]any{
		"a method reads the stack": []any{stackReader{}},
		"an error at a place":      []any{func() {}},
	} {
		t.Run(name, func(t *testing.T) {
			var b bytes.Buffer
			enc := jsontext.NewEncoder(&b)
			if err := enc.WriteToken(jsontext.BeginObject); err != nil {
				t.Fatal(err)
			}
			if err := enc.WriteToken(jsontext.String(strings.Repeat(callerName, 64))); err != nil {
				t.Fatal(err)
			}
			_ = MarshalEncode(enc, v)
			ctx := encoder.TakeRuntimeContext()
			defer encoder.ReleaseRuntimeContext(ctx)
			st := (*callState)(ctx.V2State)
			if st == nil || cap(st.levels) == 0 && cap(st.outer) == 0 {
				t.Skip("the pool gave another context")
			}
			for _, l := range append(st.levels[:cap(st.levels)], st.outer[:cap(st.outer)]...) {
				if bytes.Contains(l.Name, []byte(callerName)) {
					t.Fatalf("a level keeps the name of the caller %.20q…", l.Name)
				}
			}
			if strings.Contains(string(st.ptr), callerName) {
				t.Fatalf("the pointer %.20q… of the caller is kept", st.ptr)
			}
		})
	}
}

// MarshalOf with functions keeps neither the functions nor what they refer to after the call: the opcodes of the
// functions of a call are not kept by the pooled context.
func TestMarshalOfKeepsNoFunctions(t *testing.T) {
	collected := make(chan struct{}, 1)
	func() {
		captured := new([64]byte)
		runtime.SetFinalizer(captured, func(*[64]byte) { collected <- struct{}{} })
		funcs := WithMarshalers(MarshalFunc(func(v int) ([]byte, error) {
			_ = captured
			return []byte("1"), nil
		}))
		if _, err := MarshalOf(struct{ A, B int }{1, 2}, funcs); err != nil {
			t.Fatal(err)
		}
	}()
	// Only one GC: the pool releases what it has after a few, whatever it holds.
	runtime.GC()
	select {
	case <-collected:
	case <-time.After(10 * time.Second):
		t.Fatal("the functions are kept alive after the call")
	}
}

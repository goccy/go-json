package decoder

import (
	"reflect"
	"testing"

	"github.com/goccy/go-json/internal/runtime"
)

// The recent decoders of a context return the decoder of the type asked, whatever types share its set
// and in whatever order they are asked: more types than the table holds are asked for by turns.
func TestRecentDecodersEvict(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeOf((*int)(nil)), reflect.TypeOf((*string)(nil)), reflect.TypeOf((*bool)(nil)),
		reflect.TypeOf((*float64)(nil)), reflect.TypeOf((*[]int)(nil)), reflect.TypeOf((*[]string)(nil)),
		reflect.TypeOf((*map[string]int)(nil)), reflect.TypeOf((*map[string]string)(nil)), reflect.TypeOf((*any)(nil)),
		reflect.TypeOf((*struct{ A int })(nil)), reflect.TypeOf((*struct{ B string })(nil)), reflect.TypeOf((*[2]int)(nil)),
		reflect.TypeOf((*int8)(nil)), reflect.TypeOf((*int16)(nil)), reflect.TypeOf((*int32)(nil)),
		reflect.TypeOf((*uint)(nil)), reflect.TypeOf((*uint8)(nil)), reflect.TypeOf((*uint16)(nil)),
		reflect.TypeOf((*uint32)(nil)), reflect.TypeOf((*uint64)(nil)), reflect.TypeOf((*float32)(nil)),
		reflect.TypeOf((*[]bool)(nil)), reflect.TypeOf((*[]float64)(nil)), reflect.TypeOf((*map[string]bool)(nil)),
		reflect.TypeOf((*[]any)(nil)), reflect.TypeOf((*map[string]any)(nil)), reflect.TypeOf((*struct{ C []int })(nil)),
		reflect.TypeOf((*struct{ D map[string]int })(nil)), reflect.TypeOf((**int)(nil)), reflect.TypeOf((**string)(nil)),
		reflect.TypeOf((*[3]string)(nil)), reflect.TypeOf((*[]*int)(nil)), reflect.TypeOf((*struct{ E *int })(nil)),
		reflect.TypeOf((*int64)(nil)),
	}
	ctx := TakeRuntimeContext()
	defer ReleaseRuntimeContext(ctx)
	for round := 0; round < 4; round++ {
		for i := range types {
			// by turns, forward and backward, so that every pair of a set evicts each other
			typ := types[i]
			if round%2 == 1 {
				typ = types[len(types)-1-i]
			}
			got, err := ctx.DecoderOf(runtime.TypePtr(typ))
			if err != nil {
				t.Fatal(err)
			}
			want, err := CompileToGetDecoder(runtime.TypePtr(typ))
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("round %d: the decoder of %v is %T, want %T", round, typ, got, want)
			}
		}
	}
	// A type which is not a pointer is an error, and is not remembered.
	if _, err := ctx.DecoderOf(runtime.TypePtr(reflect.TypeOf(0))); err == nil {
		t.Fatal("expected an error for a type which is not a pointer")
	}
}

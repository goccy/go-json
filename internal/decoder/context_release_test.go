package decoder

import (
	"context"
	"testing"
)

type releaseTestKey struct{}

// A context which goes back to the pool keeps nothing of what the caller gave to the call: the pool would keep it
// alive for the calls after it.
func TestReleaseRuntimeContextDropsWhatTheCallerGave(t *testing.T) {
	path, err := new(PathBuilder).Build([]rune("$[*]"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := TakeRuntimeContext()
	ctx.Option.Context = context.WithValue(context.Background(), releaseTestKey{}, new(int))
	ctx.SetPath(path)
	opt := ctx.Option
	ReleaseRuntimeContext(ctx)
	if opt.Context != nil || opt.Path != nil || ctx.path != (Path{}) {
		t.Fatalf("the context of the pool keeps the context %v, the path %v, the copy %v", opt.Context, opt.Path, ctx.path)
	}
}

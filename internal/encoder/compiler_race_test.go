package encoder

import (
	"reflect"
	goruntime "runtime"
	"sync"
	"testing"

	"github.com/goccy/go-json/internal/runtime"
)

// The opcodes of a type are compiled by every goroutine which encodes the type for the first time at the same
// time, and all of them must get the ones which are cached.
func TestCompileToGetCodeSetConcurrently(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeOf(struct{ A string }{}),
		reflect.TypeOf(struct{ B int }{}),
		reflect.TypeOf(struct{ C []string }{}),
		reflect.TypeOf(struct{ D map[string]int }{}),
		reflect.TypeOf([]struct{ E *int }{}),
	}
	workers := goruntime.GOMAXPROCS(0) * 4
	results := make([][]*OpcodeSet, workers)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx := TakeRuntimeContext()
			defer ReleaseRuntimeContext(ctx)
			<-start
			for _, typ := range types {
				codeSet, err := CompileToGetCodeSet(ctx, uintptr(runtime.TypePtr(typ)))
				if err != nil {
					t.Error(err)
					return
				}
				results[i] = append(results[i], codeSet)
			}
		}(i)
	}
	close(start)
	wg.Wait()

	ctx := TakeRuntimeContext()
	defer ReleaseRuntimeContext(ctx)
	for j, typ := range types {
		cached, err := CompileToGetCodeSet(ctx, uintptr(runtime.TypePtr(typ)))
		if err != nil {
			t.Fatal(err)
		}
		if cached.Type != typ {
			t.Fatalf("expected the opcodes of %s but got the ones of %s", typ, cached.Type)
		}
		for i := range results {
			if len(results[i]) > j && results[i][j] != cached {
				t.Fatalf("%s: a goroutine got the opcodes which are not cached", typ)
			}
		}
	}
}

func BenchmarkCompileToGetCodeSet(b *testing.B) {
	typeptr := uintptr(runtime.TypePtr(reflect.TypeOf(struct{ Name string }{})))
	b.RunParallel(func(pb *testing.PB) {
		ctx := TakeRuntimeContext()
		defer ReleaseRuntimeContext(ctx)
		for pb.Next() {
			_, _ = CompileToGetCodeSet(ctx, typeptr)
		}
	})
}

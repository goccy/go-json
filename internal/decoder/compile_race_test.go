package decoder

import (
	"reflect"
	goruntime "runtime"
	"sync"
	"testing"

	"github.com/goccy/go-json/internal/runtime"
)

// The decoder of a type is compiled by every goroutine which decodes the type for the first time at the same
// time, and all of them must get the one which is cached.
func TestCompileToGetDecoderConcurrently(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeOf(&struct{ A string }{}),
		reflect.TypeOf(&struct{ B int }{}),
		reflect.TypeOf(&struct{ C []string }{}),
		reflect.TypeOf(&map[string]int{}),
		reflect.TypeOf(&[]struct{ E *int }{}),
	}
	workers := goruntime.GOMAXPROCS(0) * 4
	results := make([][]Decoder, workers)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			for _, typ := range types {
				dec, err := CompileToGetDecoder(runtime.TypePtr(typ))
				if err != nil {
					t.Error(err)
					return
				}
				results[i] = append(results[i], dec)
			}
		}(i)
	}
	close(start)
	wg.Wait()

	for j, typ := range types {
		cached, err := CompileToGetDecoder(runtime.TypePtr(typ))
		if err != nil {
			t.Fatal(err)
		}
		for i := range results {
			if len(results[i]) > j && results[i][j] != cached {
				t.Fatalf("%s: a goroutine got a decoder which is not cached", typ)
			}
		}
	}
}

func BenchmarkCompileToGetDecoder(b *testing.B) {
	typ := runtime.TypePtr(reflect.TypeOf(&struct{ Name string }{}))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = CompileToGetDecoder(typ)
		}
	})
}

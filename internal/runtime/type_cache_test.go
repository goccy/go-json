package runtime

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"
)

func TestTypeCache(t *testing.T) {
	var cache TypeCache[int]
	if cache.Load(0x1000) != nil {
		t.Fatal("an empty cache has a value")
	}

	// the addresses of the types are aligned and close to each other, and the table grows many times.
	const num = 5000
	values := make([]int, num)
	addr := func(i int) uintptr { return 0x100000 + uintptr(i)*32 }
	for i := range values {
		values[i] = i
		if got := cache.Store(addr(i), &values[i]); got != &values[i] {
			t.Fatalf("Store(%d): expected the value which is stored", i)
		}
		// every value stored so far stays, also after the table grew.
		for _, j := range []int{0, i / 2, i} {
			if got := cache.Load(addr(j)); got != &values[j] {
				t.Fatalf("Load(%d) after Store(%d): unexpected value %v", j, i, got)
			}
		}
	}
	for i := range values {
		if got := cache.Load(addr(i)); got != &values[i] {
			t.Fatalf("Load(%d): unexpected value %v", i, got)
		}
		// an address between the types, and one after them, has no value.
		if cache.Load(addr(i)+8) != nil || cache.Load(addr(num+i)) != nil {
			t.Fatalf("Load near %d: a value for an address which is not stored", i)
		}
	}

	// the value which was stored first stays.
	other := -1
	if got := cache.Store(addr(10), &other); got != &values[10] {
		t.Fatal("Store replaced the value")
	}
	if got := cache.Load(addr(10)); got != &values[10] {
		t.Fatal("Store replaced the value")
	}
}

func TestTypeCacheConcurrently(t *testing.T) {
	var cache TypeCache[int]
	const (
		workers = 16
		num     = 2000
	)
	addr := func(i int) uintptr { return 0x7f000000 + uintptr(i)*64 }
	results := make([][]*int, workers)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			results[w] = make([]*int, num)
			for i := 0; i < num; i++ {
				// every worker stores its own value for every type, while the others load and store.
				v := cache.Load(addr(i))
				if v == nil {
					own := i
					v = cache.Store(addr(i), &own)
				}
				if *v != i {
					t.Errorf("worker %d: the value of %d is %d", w, i, *v)
					return
				}
				results[w][i] = v
			}
		}(w)
	}
	wg.Wait()

	for i := 0; i < num; i++ {
		cached := cache.Load(addr(i))
		if cached == nil {
			t.Fatalf("no value for %d", i)
		}
		for w := range results {
			if results[w][i] != cached {
				t.Fatalf("worker %d got a value for %d which is not cached", w, i)
			}
		}
	}
}

// typeCacheTypes returns the addresses of n objects of the size of a type descriptor, laid out as the types of a
// program are: aligned and close to each other.
func typeCacheTypes(n int) []uintptr {
	objs := make([][80]byte, n)
	keepTypeCacheTypes = append(keepTypeCacheTypes, objs)
	types := make([]uintptr, n)
	for i := range objs {
		types[i] = uintptr(unsafe.Pointer(&objs[i]))
	}
	return types
}

var keepTypeCacheTypes [][][80]byte

var typeCacheSink *int

// BenchmarkTypeCacheLoad looks up types by turns in tables of several sizes.
func BenchmarkTypeCacheLoad(b *testing.B) {
	for _, n := range []int{16, 1024, 8192} {
		types := typeCacheTypes(n)
		v := 1
		var cache TypeCache[int]
		for _, typ := range types {
			cache.Store(typ, &v)
		}
		for _, k := range []int{1, 8, 128} {
			if k > n {
				continue
			}
			look := make([]uintptr, k)
			for i := range look {
				look[i] = types[(i*7919)%n]
			}
			b.Run(fmt.Sprintf("%d types/%d by turns", n, k), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					typeCacheSink = cache.Load(look[i%k])
				}
			})
		}
	}
}

// BenchmarkTypeCacheStore stores n types one by one into an empty table, as a program compiles its types: the
// table is made larger only when it gets half full, so the stores cost in proportion to n.
func BenchmarkTypeCacheStore(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		types := typeCacheTypes(n)
		v := 1
		b.Run(fmt.Sprintf("%d types", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var cache TypeCache[int]
				for _, typ := range types {
					cache.Store(typ, &v)
				}
			}
		})
	}
}

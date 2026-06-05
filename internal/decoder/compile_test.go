package decoder

import (
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

type decoderCacheTestDecoder struct {
	id int
}

func (d *decoderCacheTestDecoder) Decode(*RuntimeContext, int64, int64, unsafe.Pointer) (int64, error) {
	return 0, nil
}

func (d *decoderCacheTestDecoder) DecodePath(*RuntimeContext, int64, int64) ([][]byte, int64, error) {
	return nil, 0, nil
}

func (d *decoderCacheTestDecoder) DecodeStream(*Stream, int64, unsafe.Pointer) error {
	return nil
}

type decoderCacheTestStruct struct {
	A int `json:"a"`
}

func TestDecoderFastCacheReturnsStoredDecoder(t *testing.T) {
	index, ok := decoderCacheTestIndex()
	if !ok {
		t.Skip("decoder fast cache is disabled for this runtime")
	}
	clearDecoderCacheTestSlot(index)
	defer clearDecoderCacheTestSlot(index)

	if got := loadCachedDecoder(index); got != nil {
		t.Fatalf("loadCachedDecoder() = %T, want nil", got)
	}

	first := &decoderCacheTestDecoder{id: 1}
	if got := storeCachedDecoder(index, first); got != first {
		t.Fatalf("storeCachedDecoder(first) = %p, want %p", got, first)
	}
	if got := loadCachedDecoder(index); got != first {
		t.Fatalf("loadCachedDecoder() = %p, want %p", got, first)
	}

	second := &decoderCacheTestDecoder{id: 2}
	if got := storeCachedDecoder(index, second); got != first {
		t.Fatalf("storeCachedDecoder(second) = %p, want cached %p", got, first)
	}
	if got := loadCachedDecoder(index); got != first {
		t.Fatalf("loadCachedDecoder() after second store = %p, want %p", got, first)
	}
}

func TestDecoderFastCacheConcurrentStoresReturnWinner(t *testing.T) {
	index, ok := decoderCacheTestIndex()
	if !ok {
		t.Skip("decoder fast cache is disabled for this runtime")
	}
	clearDecoderCacheTestSlot(index)
	defer clearDecoderCacheTestSlot(index)

	const goroutineNum = 64
	start := make(chan struct{})
	results := make(chan Decoder, goroutineNum)
	var wg sync.WaitGroup
	wg.Add(goroutineNum)

	for i := 0; i < goroutineNum; i++ {
		dec := &decoderCacheTestDecoder{id: i}
		go func() {
			defer wg.Done()
			<-start
			results <- storeCachedDecoder(index, dec)
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	winner := loadCachedDecoder(index)
	if winner == nil {
		t.Fatal("loadCachedDecoder() = nil, want cached decoder")
	}
	for got := range results {
		if got != winner {
			t.Fatalf("storeCachedDecoder() = %p, want cached winner %p", got, winner)
		}
	}
}

func decoderCacheTestIndex() (uintptr, bool) {
	initDecoder()
	typ := runtime.Type2RType(reflect.TypeOf((*decoderCacheTestStruct)(nil)))
	typeptr := uintptr(unsafe.Pointer(typ))
	if typeptr > typeAddr.MaxTypeAddr || typeptr < typeAddr.BaseTypeAddr {
		return 0, false
	}
	return (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift, true
}

func clearDecoderCacheTestSlot(index uintptr) {
	cachedDecoder[index].Store(nil)
}

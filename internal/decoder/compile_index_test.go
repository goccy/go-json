package decoder

import (
	"sync/atomic"
	"testing"

	"github.com/goccy/go-json/internal/runtime"
)

func TestDecoderCacheIndexFallsBackWhenPastSlice(t *testing.T) {
	initDecoder()
	origAddr := typeAddr
	origCache := cachedDecoder
	t.Cleanup(func() {
		typeAddr = origAddr
		cachedDecoder = origCache
	})

	typeAddr = &runtime.TypeAddr{
		BaseTypeAddr: 1000,
		MaxTypeAddr:  5000,
		AddrRange:    4000,
		AddrShift:    0,
	}
	cachedDecoder = make([]atomic.Pointer[Decoder], 2)

	// In-range for min/max but past the allocated cache.
	if _, slow := decoderCacheIndex(4500); !slow {
		t.Fatal("expected slow path when index >= len(cachedDecoder)")
	}
	if _, slow := decoderCacheIndex(10); !slow {
		t.Fatal("expected slow path when typeptr < BaseTypeAddr")
	}
	if _, slow := decoderCacheIndex(8000); !slow {
		t.Fatal("expected slow path when typeptr > MaxTypeAddr")
	}
	if idx, slow := decoderCacheIndex(1001); slow || idx != 1 {
		t.Fatalf("expected fast path index=1, got idx=%d slow=%v", idx, slow)
	}
}

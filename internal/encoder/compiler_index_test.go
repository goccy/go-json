package encoder

import (
	"sync/atomic"
	"testing"

	"github.com/goccy/go-json/internal/runtime"
)

func TestEncoderCacheIndexFallsBackWhenPastSlice(t *testing.T) {
	initEncoder()
	origAddr := typeAddr
	origCache := cachedOpcodeSets
	t.Cleanup(func() {
		typeAddr = origAddr
		cachedOpcodeSets = origCache
	})

	typeAddr = &runtime.TypeAddr{
		BaseTypeAddr: 1000,
		MaxTypeAddr:  5000,
		AddrRange:    4000,
		AddrShift:    0,
	}
	cachedOpcodeSets = make([]atomic.Pointer[OpcodeSet], 2)

	if _, slow := encoderCacheIndex(4500); !slow {
		t.Fatal("expected slow path when index >= len(cachedOpcodeSets)")
	}
	if _, slow := encoderCacheIndex(10); !slow {
		t.Fatal("expected slow path when typeptr < BaseTypeAddr")
	}
	if _, slow := encoderCacheIndex(8000); !slow {
		t.Fatal("expected slow path when typeptr > MaxTypeAddr")
	}
	if idx, slow := encoderCacheIndex(1001); slow || idx != 1 {
		t.Fatalf("expected fast path index=1, got idx=%d slow=%v", idx, slow)
	}
}

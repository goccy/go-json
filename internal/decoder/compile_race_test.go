package decoder

import (
	"os"
	goruntime "runtime"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

func TestCompileToGetDecoder(t *testing.T) {
	type TestType struct{ Name string }
	var v any = &TestType{}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ
	workers := goruntime.GOMAXPROCS(0)

	tests := []struct {
		Name         string
		Fn           func(typ *runtime.Type) (Decoder, error)
		RaceExpected bool
	}{
		{
			Name:         "no_race",
			Fn:           CompileToGetDecoderNoRace,
			RaceExpected: true,
		},
		{
			Name: "race",
			Fn:   CompileToGetDecoderRace,
		},
		{
			Name: "atomic",
			Fn:   CompileToGetDecoderAtomic,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			if tt.RaceExpected && os.Getenv("SHOW_COMPILE_RACE") == "" {
				t.Skip()
			}

			start := make(chan struct{})
			wg := sync.WaitGroup{}
			wg.Add(workers)

			for i := 0; i < workers; i++ {
				go func() {
					ctx := TakeRuntimeContext()
					<-start
					_, _ = tt.Fn(typ)
					ReleaseRuntimeContext(ctx)
					wg.Done()
				}()
			}

			close(start)
			wg.Wait()

		})
	}
}

func BenchmarkCompileToGetDecoder(b *testing.B) {
	type TestType struct{ Name string }
	var v any = &TestType{}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ

	tests := []struct {
		Name         string
		Fn           func(typ *runtime.Type) (Decoder, error)
		RaceExpected bool
	}{
		{
			Name:         "no_race",
			Fn:           CompileToGetDecoderNoRace,
			RaceExpected: true,
		},
		{
			Name: "race",
			Fn:   CompileToGetDecoderRace,
		},
		{
			Name: "atomic",
			Fn:   CompileToGetDecoderAtomic,
		},
	}
	for _, tt := range tests {
		b.Run(tt.Name, func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {

					ctx := TakeRuntimeContext()
					_, _ = tt.Fn(typ)
					ReleaseRuntimeContext(ctx)

				}
			})
		})
	}
}

var (
	initPureOnce      sync.Once
	cachedDecoderPure []Decoder
)

func initEncoderPure() {
	initPureOnce.Do(func() {
		typeAddr = runtime.AnalyzeTypeAddr()
		if typeAddr == nil {
			typeAddr = &runtime.TypeAddr{}
		}
		cachedDecoderPure = make([]Decoder, typeAddr.AddrRange>>typeAddr.AddrShift+1)
	})
}

func CompileToGetDecoderNoRace(typ *runtime.Type) (Decoder, error) {
	initEncoderPure()
	typeptr := uintptr(unsafe.Pointer(typ))
	if typeptr > typeAddr.MaxTypeAddr {
		return compileToGetDecoderSlowPath(typeptr, typ)
	}

	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if dec := cachedDecoderPure[index]; dec != nil {
		return dec, nil
	}

	dec, err := compileHead(typ, map[uintptr]Decoder{})
	if err != nil {
		return nil, err
	}
	cachedDecoderPure[index] = dec
	return dec, nil
}

var decMu_test sync.RWMutex

func CompileToGetDecoderRace(typ *runtime.Type) (Decoder, error) {
	initEncoderPure()
	typeptr := uintptr(unsafe.Pointer(typ))
	if typeptr > typeAddr.MaxTypeAddr {
		return compileToGetDecoderSlowPath(typeptr, typ)
	}

	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	decMu_test.RLock()
	if dec := cachedDecoderPure[index]; dec != nil {
		decMu_test.RUnlock()
		return dec, nil
	}
	decMu_test.RUnlock()

	dec, err := compileHead(typ, map[uintptr]Decoder{})
	if err != nil {
		return nil, err
	}
	decMu_test.Lock()
	cachedDecoderPure[index] = dec
	decMu_test.Unlock()
	return dec, nil
}

var (
	initAtomicOnce      sync.Once
	cachedDecoderAtomic []atomic.Pointer[Decoder]
)

func initDecoderAtomic() {
	initAtomicOnce.Do(func() {
		typeAddr = runtime.AnalyzeTypeAddr()
		if typeAddr == nil {
			typeAddr = &runtime.TypeAddr{}
		}
		cachedDecoderAtomic = make([]atomic.Pointer[Decoder], typeAddr.AddrRange>>typeAddr.AddrShift+1)
	})
}

func CompileToGetDecoderAtomic(typ *runtime.Type) (Decoder, error) {
	initDecoderAtomic()
	typeptr := uintptr(unsafe.Pointer(typ))
	if typeptr > typeAddr.MaxTypeAddr || typeptr < typeAddr.BaseTypeAddr {
		return compileToGetDecoderSlowPath(typeptr, typ)
	}

	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if dec := cachedDecoderAtomic[index].Load(); dec != nil {
		return *dec, nil
	}

	dec, err := compileHead(typ, map[uintptr]Decoder{})
	if err != nil {
		return nil, err
	}
	cachedDecoderAtomic[index].Store(&dec)
	return dec, nil
}

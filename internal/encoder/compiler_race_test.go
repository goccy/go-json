package encoder

import (
	"os"
	goruntime "runtime"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

func TestCompileToGetCodeSetDataRace(t *testing.T) {
	type TestType struct{ Name string }
	var v any = TestType{}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ
	typeptr := uintptr(unsafe.Pointer(typ))
	workers := goruntime.GOMAXPROCS(0)

	type testCase struct {
		Name         string
		Fn           func(ctx *RuntimeContext, typeptr uintptr) (*OpcodeSet, error)
		RaceExpected bool
	}
	testCases := []testCase{
		{
			Name:         "no_race",
			Fn:           CompileToGetCodeSetNoRace,
			RaceExpected: true,
		},
		{
			Name: "race",
			Fn:   CompileToGetCodeSetRace,
		},
		{
			Name: "sync_map",
			Fn:   CompileToGetCodeSetSyncMap,
		},
		{
			Name: "sync_atomic",
			Fn:   CompileToGetCodeSetAtomic,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			if tc.RaceExpected && os.Getenv("SHOW_COMPILE_RACE") == "" {
				t.Skip()
			}

			start := make(chan struct{})
			wg := sync.WaitGroup{}
			wg.Add(workers)

			for i := 0; i < workers; i++ {
				go func() {
					ctx := TakeRuntimeContext()
					<-start
					_, _ = tc.Fn(ctx, typeptr)
					ReleaseRuntimeContext(ctx)
					wg.Done()
				}()
			}

			close(start)
			wg.Wait()

		})
	}
}

func BenchmarkCompileToGetCodeSet(b *testing.B) {
	initEncoderPure()
	initEncoderAtomic()

	type TestType struct {
		Name string
	}
	var v any = TestType{}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ
	typeptr := uintptr(unsafe.Pointer(typ))

	type testCase struct {
		Name         string
		Fn           func(ctx *RuntimeContext, typeptr uintptr) (*OpcodeSet, error)
		RaceExpected bool
	}
	testCases := []testCase{
		{
			Name:         "no_race",
			Fn:           CompileToGetCodeSetNoRace,
			RaceExpected: true,
		},
		{
			Name: "race",
			Fn:   CompileToGetCodeSetRace,
		},
		{
			Name: "sync_map",
			Fn:   CompileToGetCodeSetSyncMap,
		},
		{
			Name: "atomic",
			Fn:   CompileToGetCodeSetAtomic,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.Name, func(b *testing.B) {

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					ctx := TakeRuntimeContext()
					_, _ = tc.Fn(ctx, typeptr)
					ReleaseRuntimeContext(ctx)
				}
			})

		})
	}
}

var (
	initEncoderPureOnce  sync.Once
	cachedOpcodeSetsPure []*OpcodeSet
)

func initEncoderPure() {
	initEncoderPureOnce.Do(func() {
		typeAddr = runtime.AnalyzeTypeAddr()
		if typeAddr == nil {
			typeAddr = &runtime.TypeAddr{}
		}
		cachedOpcodeSetsPure = make([]*OpcodeSet, typeAddr.AddrRange>>typeAddr.AddrShift+1)
	})
}

func CompileToGetCodeSetNoRace(ctx *RuntimeContext, typeptr uintptr) (*OpcodeSet, error) {
	initEncoderPure()
	if typeptr > typeAddr.MaxTypeAddr || typeptr < typeAddr.BaseTypeAddr {
		codeSet, err := compileToGetCodeSetSlowPath(typeptr)
		if err != nil {
			return nil, err
		}
		return getFilteredCodeSetIfNeeded(ctx, codeSet)
	}
	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if codeSet := cachedOpcodeSetsPure[index]; codeSet != nil {
		filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
		if err != nil {
			return nil, err
		}
		return filtered, nil
	}
	codeSet, err := newCompiler().compile(typeptr)
	if err != nil {
		return nil, err
	}
	filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
	if err != nil {
		return nil, err
	}
	cachedOpcodeSetsPure[index] = codeSet
	return filtered, nil
}

var setsMu_test sync.RWMutex

func CompileToGetCodeSetRace(ctx *RuntimeContext, typeptr uintptr) (*OpcodeSet, error) {
	initEncoderPure()
	if typeptr > typeAddr.MaxTypeAddr || typeptr < typeAddr.BaseTypeAddr {
		codeSet, err := compileToGetCodeSetSlowPath(typeptr)
		if err != nil {
			return nil, err
		}
		return getFilteredCodeSetIfNeeded(ctx, codeSet)
	}
	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	setsMu_test.RLock()
	if codeSet := cachedOpcodeSetsPure[index]; codeSet != nil {
		filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
		if err != nil {
			setsMu_test.RUnlock()
			return nil, err
		}
		setsMu_test.RUnlock()
		return filtered, nil
	}
	setsMu_test.RUnlock()

	codeSet, err := newCompiler().compile(typeptr)
	if err != nil {
		return nil, err
	}
	filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
	if err != nil {
		return nil, err
	}
	setsMu_test.Lock()
	cachedOpcodeSetsPure[index] = codeSet
	setsMu_test.Unlock()
	return filtered, nil
}

var cachedOpcodeSetsMp sync.Map

func CompileToGetCodeSetSyncMap(ctx *RuntimeContext, typeptr uintptr) (*OpcodeSet, error) {
	if typeptr > typeAddr.MaxTypeAddr || typeptr < typeAddr.BaseTypeAddr {
		codeSet, err := compileToGetCodeSetSlowPath(typeptr)
		if err != nil {
			return nil, err
		}
		return getFilteredCodeSetIfNeeded(ctx, codeSet)
	}
	//index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if v, has := cachedOpcodeSetsMp.Load(typeptr); has {
		codeSet := v.(*OpcodeSet)
		filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
		if err != nil {
			return nil, err
		}
		return filtered, nil
	}
	codeSet, err := newCompiler().compile(typeptr)
	if err != nil {
		return nil, err
	}
	filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
	if err != nil {
		return nil, err
	}
	cachedOpcodeSetsMp.Store(typeptr, codeSet)
	return filtered, nil
}

var (
	initEncoderAtomicOnce  sync.Once
	cachedOpcodeSetsAtomic []atomic.Pointer[OpcodeSet]
)

func initEncoderAtomic() {
	initEncoderAtomicOnce.Do(func() {
		typeAddr = runtime.AnalyzeTypeAddr()
		if typeAddr == nil {
			typeAddr = &runtime.TypeAddr{}
		}
		cachedOpcodeSetsAtomic = make([]atomic.Pointer[OpcodeSet], typeAddr.AddrRange>>typeAddr.AddrShift+1)
	})
}

func CompileToGetCodeSetAtomic(ctx *RuntimeContext, typeptr uintptr) (*OpcodeSet, error) {
	initEncoderAtomic()
	if typeptr > typeAddr.MaxTypeAddr || typeptr < typeAddr.BaseTypeAddr {
		codeSet, err := compileToGetCodeSetSlowPath(typeptr)
		if err != nil {
			return nil, err
		}
		return getFilteredCodeSetIfNeeded(ctx, codeSet)
	}
	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if codeSet := cachedOpcodeSetsAtomic[index].Load(); codeSet != nil {
		filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
		if err != nil {
			return nil, err
		}
		return filtered, nil
	}
	codeSet, err := newCompiler().compile(typeptr)
	if err != nil {
		return nil, err
	}
	filtered, err := getFilteredCodeSetIfNeeded(ctx, codeSet)
	if err != nil {
		return nil, err
	}
	cachedOpcodeSetsAtomic[index].Store(codeSet)
	return filtered, nil
}

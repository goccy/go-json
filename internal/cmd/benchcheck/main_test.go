package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseBenchOutput(t *testing.T) {
	output := `goos: darwin
goarch: arm64
pkg: benchmark
Benchmark_Decode_SmallStruct_Unmarshal_GoJson-8   	 5000000	       250.5 ns/op	     256 B/op	       2 allocs/op
Benchmark_MarshalBytes_GoJson/32-8                	10000000	       100 ns/op
Benchmark_MarshalBytes_GoJson/32-8                	10000000	        90 ns/op
BenchmarkThroughput-8                             	    1000	    12.5 MB/s	   2000 ns/op
BenchmarkNoTime-8                                 	    1000	    12.5 MB/s
BenchmarkBelowTimerResolution-8                   	       1
BenchmarkLogged-8
--- BENCH: BenchmarkLogged-8
    bench_test.go:10: message
PASS
ok  	benchmark	10.000s
`
	got := parseBenchOutput(output, 8)
	want := measurement{
		"Benchmark_Decode_SmallStruct_Unmarshal_GoJson": 250.5,
		"Benchmark_MarshalBytes_GoJson/32":              90,
		"BenchmarkThroughput":                           2000,
		// the testing package omits ns/op when it is zero.
		"BenchmarkNoTime":               0,
		"BenchmarkBelowTimerResolution": 0,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result: got %v, want %v", got, want)
	}
}

func TestParseBenchOutputSingleProc(t *testing.T) {
	got := parseBenchOutput("BenchmarkFoo-8 \t 100 \t 10 ns/op\n", 1)
	want := measurement{"BenchmarkFoo-8": 10}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result: got %v, want %v", got, want)
	}
}

// statusesOf returns the statuses of the results by the name.
func statusesOf(cmp *comparison) map[string]status {
	statuses := map[string]status{}
	for _, result := range cmp.sortedResults() {
		statuses[result.Name] = result.Status
	}
	return statuses
}

func TestComparison(t *testing.T) {
	t.Run("the fastest result of each side is compared", func(t *testing.T) {
		cmp := newComparison(3, 15)
		cmp.add("BenchmarkA", measurement{"BenchmarkA": 100}, measurement{"BenchmarkA": 130})
		// the base happened to be slow in the second attempt: a degraded head must not pass with it.
		cmp.add("BenchmarkA", measurement{"BenchmarkA": 500}, measurement{"BenchmarkA": 125})
		want := []benchResult{{Name: "BenchmarkA", Func: "BenchmarkA", Status: statusRegressed, BaseNs: 100, HeadNs: 125, Attempts: 2, baseSeen: true}}
		if got := cmp.sortedResults(); !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected results:\n got %+v\nwant %+v", got, want)
		}
		if !cmp.degraded() {
			t.Fatal("degradation must be detected")
		}
	})

	t.Run("noise of a single benchmark is not a degradation", func(t *testing.T) {
		cmp := newComparison(3, 15)
		// one benchmark is 8% slower and the others are as fast as the base: the mean is about +0.8%.
		cmp.add("BenchmarkNoise", measurement{"BenchmarkNoise": 100}, measurement{"BenchmarkNoise": 108})
		for _, name := range []string{"BenchmarkB", "BenchmarkC", "BenchmarkD", "BenchmarkE", "BenchmarkF", "BenchmarkG", "BenchmarkH", "BenchmarkI", "BenchmarkJ"} {
			cmp.add(name, measurement{name: 100}, measurement{name: 100})
		}
		if cmp.degraded() {
			t.Fatalf("must not be degraded: mean %+.2f%%", cmp.meanDeltaPercent())
		}
		if got := cmp.pendingFuncs(); len(got) != 0 {
			t.Fatalf("nothing must be measured again: %v", got)
		}
	})

	t.Run("mean beyond the tolerance is a degradation", func(t *testing.T) {
		cmp := newComparison(3, 15)
		// every benchmark is 5% slower: none of them is beyond the tolerance of a single benchmark.
		cmp.add("BenchmarkA", measurement{"BenchmarkA": 100}, measurement{"BenchmarkA": 105})
		cmp.add("BenchmarkB", measurement{"BenchmarkB": 200}, measurement{"BenchmarkB": 210})
		cmp.add("BenchmarkFast", measurement{"BenchmarkFast": 100}, measurement{"BenchmarkFast": 102})
		if got, want := statusesOf(cmp), (map[string]status{"BenchmarkA": statusOK, "BenchmarkB": statusOK, "BenchmarkFast": statusOK}); !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected statuses: got %v, want %v", got, want)
		}
		if !cmp.degraded() {
			t.Fatalf("degradation must be detected: mean %+.2f%%", cmp.meanDeltaPercent())
		}
		// only what makes the mean slow is measured again.
		if got, want := cmp.pendingFuncs(), []string{"BenchmarkA", "BenchmarkB"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected pending: got %v, want %v", got, want)
		}
		// the noise is gone in the next attempt.
		cmp.add("BenchmarkA", measurement{"BenchmarkA": 100}, measurement{"BenchmarkA": 100})
		cmp.add("BenchmarkB", measurement{"BenchmarkB": 200}, measurement{"BenchmarkB": 201})
		if cmp.degraded() {
			t.Fatalf("must not be degraded: mean %+.2f%%", cmp.meanDeltaPercent())
		}
	})

	t.Run("degradation of a single benchmark is detected", func(t *testing.T) {
		cmp := newComparison(3, 15)
		cmp.add("BenchmarkDegraded", measurement{"BenchmarkDegraded": 100}, measurement{"BenchmarkDegraded": 130})
		for _, name := range []string{"BenchmarkB", "BenchmarkC", "BenchmarkD", "BenchmarkE", "BenchmarkF", "BenchmarkG", "BenchmarkH", "BenchmarkI", "BenchmarkJ", "BenchmarkK", "BenchmarkL"} {
			cmp.add(name, measurement{name: 100}, measurement{name: 100})
		}
		if mean := cmp.meanDeltaPercent(); mean > 3 {
			t.Fatalf("the mean must be within the tolerance for this test: %+.2f%%", mean)
		}
		if got := statusesOf(cmp)["BenchmarkDegraded"]; got != statusRegressed {
			t.Fatalf("unexpected status: %v", got)
		}
		if !cmp.degraded() {
			t.Fatal("degradation must be detected")
		}
		if got, want := cmp.pendingFuncs(), []string{"BenchmarkDegraded"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected pending: got %v, want %v", got, want)
		}
	})

	t.Run("benchmark which the base doesn't have", func(t *testing.T) {
		cmp := newComparison(3, 15)
		cmp.add("BenchmarkNew", measurement{}, measurement{"BenchmarkNew": 10})
		cmp.add("BenchmarkSub", measurement{"BenchmarkSub/a": 100}, measurement{"BenchmarkSub/a": 100, "BenchmarkSub/new": 1000})
		want := map[string]status{"BenchmarkNew": statusNew, "BenchmarkSub/a": statusOK, "BenchmarkSub/new": statusNew}
		if got := statusesOf(cmp); !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected statuses: got %v, want %v", got, want)
		}
		if cmp.degraded() {
			t.Fatal("a new benchmark is not a degradation")
		}
	})
}

func TestCache(t *testing.T) {
	c := &cache{dir: filepath.Join(t.TempDir(), "cache"), readable: true}
	key := baselineKey{
		Config:    benchConfig{Dir: "benchmarks", Bench: "."},
		Machine:   machine{Hostname: "host", GOOS: "linux", GOARCH: "amd64", NumCPU: 4, GoVersion: "go1.21.0"},
		BenchHash: "hash",
	}
	if got, err := c.loadBaseline("commit", key); err != nil || got != nil {
		t.Fatalf("unexpected result for empty cache: %v, %v", got, err)
	}
	want := &baseline{Commit: "commit", Key: key, Results: measurement{"BenchmarkFoo": 1.5}}
	if err := c.storeBaseline(want); err != nil {
		t.Fatal(err)
	}
	got, err := c.loadBaseline("commit", key)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected baseline: got %+v, want %+v", got, want)
	}

	otherMachine := key
	otherMachine.Machine.Hostname = "other"
	if got, err := c.loadBaseline("commit", otherMachine); err != nil || got != nil {
		t.Fatalf("baseline of the other machine must not be used: %v, %v", got, err)
	}
	if got, err := c.loadBaseline("other-commit", key); err != nil || got != nil {
		t.Fatalf("baseline of the other commit must not be used: %v, %v", got, err)
	}

	c.readable = false
	if got, err := c.loadBaseline("commit", key); err != nil || got != nil {
		t.Fatalf("cache must be ignored: %v, %v", got, err)
	}
}

func TestHashDir(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	hash := func() string {
		t.Helper()
		h, err := hashDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	write("a_test.go", "package a")
	write(filepath.Join("testdata", "a.json"), "{}")
	first := hash()
	if got := hash(); got != first {
		t.Fatalf("hash must be stable: %s != %s", got, first)
	}
	write(filepath.Join("testdata", "a.json"), "[]")
	second := hash()
	if second == first {
		t.Fatal("hash must change when a content is changed")
	}
	write("b_test.go", "")
	if got := hash(); got == second {
		t.Fatal("hash must change when a file is added")
	}
}

func TestComparisonWithZeroMeasurement(t *testing.T) {
	cmp := newComparison(3, 15)
	cmp.add("BenchmarkBothZero", measurement{"BenchmarkBothZero": 0}, measurement{"BenchmarkBothZero": 0})
	cmp.add("BenchmarkZeroBase", measurement{"BenchmarkZeroBase": 0}, measurement{"BenchmarkZeroBase": 20})
	cmp.add("BenchmarkZeroBase", measurement{"BenchmarkZeroBase": 0}, measurement{"BenchmarkZeroBase": 10})
	cmp.add("BenchmarkZeroHead", measurement{"BenchmarkZeroHead": 10}, measurement{"BenchmarkZeroHead": 0})

	want := []benchResult{
		{Name: "BenchmarkBothZero", Func: "BenchmarkBothZero", Status: statusOK, Attempts: 1, baseSeen: true},
		{Name: "BenchmarkZeroBase", Func: "BenchmarkZeroBase", Status: statusRegressed, BaseNs: 0, HeadNs: 10, Attempts: 2, baseSeen: true},
		{Name: "BenchmarkZeroHead", Func: "BenchmarkZeroHead", Status: statusOK, BaseNs: 10, HeadNs: 0, Attempts: 1, baseSeen: true},
	}
	if got := cmp.sortedResults(); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected results:\n got %+v\nwant %+v", got, want)
	}
	// the ratio to zero can't be represented, so such benchmarks are not a part of the mean.
	if got := cmp.meanDeltaPercent(); got != 0 {
		t.Fatalf("unexpected mean: %v", got)
	}
	if got := formatDelta(0, 10); got != "-" {
		t.Fatalf("unexpected delta for zero base: %s", got)
	}
	if got := formatDelta(100, 110); got != "+10.00%" {
		t.Fatalf("unexpected delta: %s", got)
	}
}

func TestShardFuncs(t *testing.T) {
	funcs := []string{"Benchmark_D", "Benchmark_A", "Benchmark_C", "Benchmark_B", "Benchmark_E"}
	seen := map[string]int{}
	for shard := 0; shard < 3; shard++ {
		for _, fn := range shardFuncs(funcs, shard, 3) {
			seen[fn]++
		}
	}
	if len(seen) != len(funcs) {
		t.Fatalf("the shards have %d functions, want %d", len(seen), len(funcs))
	}
	for fn, n := range seen {
		if n != 1 {
			t.Fatalf("%s is in %d shards", fn, n)
		}
	}
	if got := shardFuncs(funcs, 0, 1); len(got) != len(funcs) {
		t.Fatalf("one shard has %d functions", len(got))
	}
}

func TestBenchmarkGroup(t *testing.T) {
	for name, want := range map[string]string{
		"BenchmarkCodeDecoder":                          groupDecode,
		"BenchmarkUnmarshalFloat64":                     groupDecode,
		"BenchmarkUnmarshalUnmapped":                    groupDecode,
		"Benchmark_Decode_SmallStruct_Unmarshal_GoJson": groupDecode,
		"Benchmark_Encode_SmallStruct_GoJson":           groupEncode,
		"Benchmark_MarshalBytes_GoJson":                 groupEncode,
		"Benchmark_Compact_GoJson":                      groupEncode,
		"BenchmarkCodeEncoder":                          groupEncode,
	} {
		if got := benchmarkGroup(name); got != want {
			t.Errorf("%s: got %s, want %s", name, got, want)
		}
	}
	funcs := []string{"Benchmark_Encode_A_GoJson", "Benchmark_Decode_A_GoJson", "BenchmarkUnmarshalX"}
	if got := groupFuncs(funcs, groupDecode); len(got) != 2 {
		t.Fatalf("decode: %v", got)
	}
	if got := groupFuncs(funcs, ""); len(got) != 3 {
		t.Fatalf("no group: %v", got)
	}
}

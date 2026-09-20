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

func TestComparison(t *testing.T) {
	cmp := newComparison(5)

	// the first attempt.
	cmp.add("BenchmarkFast", measurement{"BenchmarkFast": 100}, measurement{"BenchmarkFast": 90})
	cmp.add("BenchmarkNoise", measurement{"BenchmarkNoise": 100}, measurement{"BenchmarkNoise": 105})
	cmp.add("BenchmarkNew", measurement{}, measurement{"BenchmarkNew": 10})
	cmp.add("BenchmarkDegraded", measurement{"BenchmarkDegraded": 100}, measurement{"BenchmarkDegraded": 130})
	cmp.add(
		"BenchmarkSub",
		measurement{"BenchmarkSub/ok": 100, "BenchmarkSub/recovered": 100},
		measurement{"BenchmarkSub/ok": 100, "BenchmarkSub/recovered": 120},
	)
	if got, want := cmp.pendingFuncs(), []string{"BenchmarkDegraded", "BenchmarkSub"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected pending: got %v, want %v", got, want)
	}

	// the second attempt: the base is measured again together with the working tree.
	cmp.add("BenchmarkDegraded", measurement{"BenchmarkDegraded": 110}, measurement{"BenchmarkDegraded": 132})
	cmp.add(
		"BenchmarkSub",
		measurement{"BenchmarkSub/ok": 100, "BenchmarkSub/recovered": 120},
		// the benchmark which has already reached the base must not be updated.
		measurement{"BenchmarkSub/ok": 500, "BenchmarkSub/recovered": 125},
	)
	if got, want := cmp.pendingFuncs(), []string{"BenchmarkDegraded"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected pending: got %v, want %v", got, want)
	}

	// the third attempt.
	cmp.add("BenchmarkDegraded", measurement{"BenchmarkDegraded": 100}, measurement{"BenchmarkDegraded": 125})

	want := []benchResult{
		{Name: "BenchmarkDegraded", Func: "BenchmarkDegraded", Status: statusRegressed, BaseNs: 110, HeadNs: 132, Attempts: 3},
		{Name: "BenchmarkFast", Func: "BenchmarkFast", Status: statusOK, BaseNs: 100, HeadNs: 90, Attempts: 1},
		{Name: "BenchmarkNew", Func: "BenchmarkNew", Status: statusNew, HeadNs: 10, Attempts: 1},
		{Name: "BenchmarkNoise", Func: "BenchmarkNoise", Status: statusOK, BaseNs: 100, HeadNs: 105, Attempts: 1},
		{Name: "BenchmarkSub/ok", Func: "BenchmarkSub", Status: statusOK, BaseNs: 100, HeadNs: 100, Attempts: 1},
		{Name: "BenchmarkSub/recovered", Func: "BenchmarkSub", Status: statusOK, BaseNs: 120, HeadNs: 125, Attempts: 2},
	}
	if got := cmp.sortedResults(); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected results:\n got %+v\nwant %+v", got, want)
	}
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
	cmp := newComparison(5)
	cmp.add("BenchmarkBothZero", measurement{"BenchmarkBothZero": 0}, measurement{"BenchmarkBothZero": 0})
	cmp.add("BenchmarkZeroBase", measurement{"BenchmarkZeroBase": 0}, measurement{"BenchmarkZeroBase": 20})
	cmp.add("BenchmarkZeroBase", measurement{"BenchmarkZeroBase": 0}, measurement{"BenchmarkZeroBase": 10})
	cmp.add("BenchmarkZeroHead", measurement{"BenchmarkZeroHead": 10}, measurement{"BenchmarkZeroHead": 0})

	want := []benchResult{
		{Name: "BenchmarkBothZero", Func: "BenchmarkBothZero", Status: statusOK, Attempts: 1},
		{Name: "BenchmarkZeroBase", Func: "BenchmarkZeroBase", Status: statusRegressed, BaseNs: 0, HeadNs: 10, Attempts: 2},
		{Name: "BenchmarkZeroHead", Func: "BenchmarkZeroHead", Status: statusOK, BaseNs: 10, HeadNs: 0, Attempts: 1},
	}
	if got := cmp.sortedResults(); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected results:\n got %+v\nwant %+v", got, want)
	}
	if got := formatDelta(0, 10); got != "-" {
		t.Fatalf("unexpected delta for zero base: %s", got)
	}
	if got := formatDelta(100, 110); got != "+10.00%" {
		t.Fatalf("unexpected delta: %s", got)
	}
}

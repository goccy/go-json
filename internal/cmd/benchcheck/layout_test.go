//go:build go1.23

// The linker randomizes the function layout ( -randlayout ) since Go 1.23.

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// testBenchSlowLayout is slower in the binary of the layout 1: 100ms against 20ms, far apart beyond the resolution
// of the timer of any platform.
const testBenchSlowLayout = `package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/lib"
)

func BenchmarkWork(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if strings.Contains(filepath.Base(os.Args[0]), ".layout1.") {
			time.Sleep(80 * time.Millisecond)
		}
		lib.Work()
	}
}
`

func TestDetectKeepsEveryLayout(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	r := newTestRepo(t)
	r.layouts = 2
	// more rounds than layouts: the round N measures the layout N % layouts.
	r.rounds = 4
	r.write("lib.go", testLibFastRefactored)
	r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchSlowLayout)
	out := r.detect(ignoreNoise)
	assertStatuses(t, out.verdict, map[string]status{"BenchmarkWork": statusOK})
	result := out.verdict.Benchmarks[0]
	// the fastest is the layout 0, and the median is between both layouts: 60ms.
	for _, side := range []struct {
		name            string
		fastest, median float64
	}{{"base", result.BaseNs, result.BaseMedianNs}, {"head", result.HeadNs, result.HeadMedianNs}} {
		if side.fastest >= float64(40*time.Millisecond) || side.median < float64(50*time.Millisecond) {
			t.Fatalf("%s: unexpected fastest %v and median %v", side.name, time.Duration(side.fastest), time.Duration(side.median))
		}
	}
}

func TestDetectMeasuresEveryLayout(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	r := newTestRepo(t)
	r.layouts = 2
	r.write("lib.go", testLibFastRefactored)
	r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchRecordingBinary)

	recordDir := t.TempDir()
	t.Setenv(testRecordDirEnv, recordDir)
	out := r.detect(ignoreNoise)
	assertStatuses(t, out.verdict, map[string]status{"BenchmarkWork": statusOK})

	entries, err := os.ReadDir(recordDir)
	if err != nil {
		t.Fatal(err)
	}
	ran := map[string]bool{}
	for _, entry := range entries {
		ran[entry.Name()] = true
	}
	for _, side := range []string{"base", "head"} {
		for layout := 0; layout < r.layouts; layout++ {
			name := filepath.Base(binaryPath("", side+".layout"+strconv.Itoa(layout)+".test"))
			if !ran[name] {
				t.Fatalf("%s was not measured: %v", name, ran)
			}
		}
	}
}

//go:build go1.23

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// The linker randomizes the function layout ( -randlayout ) since Go 1.23.
func TestCheckMeasuresEveryLayout(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	r := newTestRepo(t)
	r.layouts = 2
	r.write("lib.go", testLibFastRefactored)
	r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchRecordingBinary)

	recordDir := t.TempDir()
	t.Setenv(testRecordDirEnv, recordDir)
	out := r.check(ignoreNoise)
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

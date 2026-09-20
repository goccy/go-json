package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// measurement maps a benchmark name ( without the GOMAXPROCS suffix ) to its ns/op.
type measurement map[string]float64

// mergeFastest merges other into m, keeping the faster result for each benchmark.
func (m measurement) mergeFastest(other measurement) {
	for name, nsPerOp := range other {
		if prev, exists := m[name]; !exists || nsPerOp < prev {
			m[name] = nsPerOp
		}
	}
}

const (
	benchmarkPrefix = "Benchmark"
	nsPerOpUnit     = "ns/op"
)

// parseBenchOutput extracts ns/op of every benchmark result line from the output of the benchmarks.
//
// The output is a text format owned by the testing package, specified at
// https://go.dev/design/14313-benchmark-format :
//
//	Benchmark<name>[-<procs>] <iterations> [<value> <unit>...]
//
// The specification requires readers to ignore every line that is not a benchmark result line,
// so such lines are skipped here as well.
// If the same benchmark is reported more than once, the fastest result is kept.
func parseBenchOutput(output string, procs int) measurement {
	result := measurement{}
	for _, line := range strings.Split(output, "\n") {
		name, nsPerOp, ok := parseBenchLine(line, procs)
		if !ok {
			continue
		}
		result.mergeFastest(measurement{name: nsPerOp})
	}
	return result
}

func parseBenchLine(line string, procs int) (string, float64, bool) {
	fields := strings.Fields(line)
	// name and iterations, followed by value/unit pairs.
	if len(fields) < 2 || len(fields)%2 != 0 {
		return "", 0, false
	}
	if !strings.HasPrefix(fields[0], benchmarkPrefix) {
		return "", 0, false
	}
	if _, err := strconv.ParseUint(fields[1], 10, 64); err != nil {
		return "", 0, false
	}
	for i := 2; i < len(fields); i += 2 {
		if fields[i+1] != nsPerOpUnit {
			continue
		}
		v, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return "", 0, false
		}
		return trimProcsSuffix(fields[0], procs), v, true
	}
	// the testing package omits ns/op when it is zero:
	// the elapsed time was below the resolution of the timer, or the benchmark suppressed the metric.
	return trimProcsSuffix(fields[0], procs), 0, true
}

// trimProcsSuffix removes the "-<GOMAXPROCS>" suffix which the testing package appends
// to the benchmark name when GOMAXPROCS is not 1.
func trimProcsSuffix(name string, procs int) string {
	if procs == 1 {
		return name
	}
	return strings.TrimSuffix(name, "-"+strconv.Itoa(procs))
}

// suite is a compiled test binary of the benchmarks.
type suite struct {
	binary string
	// dir is the directory of the benchmark package.
	// The binary runs there, as `go test` does, so that the benchmarks can read their testdata.
	dir string
}

// buildSuite compiles the benchmarks in dir into binary.
// If modFile is not empty, it is used instead of go.mod in dir.
func buildSuite(ctx context.Context, dir, modFile, binary string) (*suite, error) {
	args := []string{"test", "-c", "-o", binary}
	if modFile != "" {
		args = append(args, "-modfile", modFile)
	}
	args = append(args, ".")
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to build benchmarks in %s: %w", dir, err)
	}
	return &suite{binary: binary, dir: dir}, nil
}

func (s *suite) exec(ctx context.Context, args ...string) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, s.binary, args...)
	cmd.Dir = s.dir
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// a crashed benchmark reports its reason to stdout.
		os.Stderr.Write(stdout.Bytes())
		return "", fmt.Errorf("failed to run %s %s: %w", s.binary, strings.Join(args, " "), err)
	}
	return stdout.String(), nil
}

// funcs returns the names of the top-level benchmark functions matched by pattern.
func (s *suite) funcs(ctx context.Context, pattern string) ([]string, error) {
	out, err := s.exec(ctx, "-test.list", pattern)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, name := range strings.Fields(out) {
		// -test.list also prints tests, fuzz targets and examples.
		// The testing package distinguishes them only by the prefix of the function name.
		if strings.HasPrefix(name, benchmarkPrefix) {
			names = append(names, name)
		}
	}
	return names, nil
}

// run measures a top-level benchmark function, including its sub-benchmarks.
// The result is empty if the suite doesn't have the function.
func (s *suite) run(ctx context.Context, fn, benchTime string) (measurement, error) {
	args := []string{"-test.run", "^$", "-test.bench", "^" + regexp.QuoteMeta(fn) + "$", "-test.count", "1"}
	if benchTime != "" {
		args = append(args, "-test.benchtime", benchTime)
	}
	out, err := s.exec(ctx, args...)
	if err != nil {
		return nil, err
	}
	return parseBenchOutput(out, runtime.GOMAXPROCS(0)), nil
}

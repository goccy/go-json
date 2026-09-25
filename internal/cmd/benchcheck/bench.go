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

// suite is the compiled test binaries of the benchmarks.
//
// A change of the library shifts where the linker places every function of the binary,
// and the alignment of a hot loop alone makes a benchmark faster or slower by more than 10%,
// even if the code it runs is not changed at all.
// Measuring the same binary again never removes this difference,
// so a suite has the binaries of the same code with different function layouts,
// and a benchmark is measured with all of them.
type suite struct {
	// binaries is indexed by the layout.
	binaries []string
	// dir is the directory of the benchmark package.
	// The binaries run there, as `go test` does, so that the benchmarks can read their testdata.
	dir string
}

// buildSuite compiles the benchmarks in dir into the binaries named after name in outDir, one per layout.
// If modFile is not empty, it is used instead of go.mod in dir.
//
// The layout 0 is the default one of the linker, and the others are randomized by the linker
// with the layout number as the seed ( -randlayout, supported since Go 1.23 ).
// Only the first build compiles the packages: the rest are linked from the build cache.
//
// The layout N is linked with the seed N + seedOffset of -randlayout ( the seed 0 is the layout of the linker ).
func buildSuite(ctx context.Context, dir, modFile, outDir, name string, layouts, seedOffset int) (*suite, error) {
	s := &suite{dir: dir}
	for layout := 0; layout < layouts; layout++ {
		binary := binaryPath(outDir, fmt.Sprintf("%s.layout%d.test", name, layout))
		args := []string{"test", "-c", "-o", binary}
		if seed := layout + seedOffset; seed != 0 {
			args = append(args, "-ldflags", "-randlayout="+strconv.Itoa(seed))
		}
		if modFile != "" {
			args = append(args, "-modfile", modFile)
		}
		args = append(args, ".")
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to build benchmarks in %s ( layout %d ): %w", dir, layout, err)
		}
		s.binaries = append(s.binaries, binary)
	}
	return s, nil
}

func (s *suite) exec(ctx context.Context, binary string, args ...string) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = s.dir
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// a crashed benchmark reports its reason to stdout.
		os.Stderr.Write(stdout.Bytes())
		return "", fmt.Errorf("failed to run %s %s: %w", binary, strings.Join(args, " "), err)
	}
	return stdout.String(), nil
}

// comparedLibraries are the libraries which the benchmarks compare go-json with.
// Their benchmarks exist to be compared with go-json by hand. They are never measured here,
// because the purpose of this command is to notice the degradation of go-json.
var comparedLibraries = map[string]struct{}{
	"EasyJson":      {},
	"EncodingJson":  {},
	"FFJson":        {},
	"FastJson":      {},
	"GoJay":         {},
	"GoJayUnsafe":   {},
	"Jettison":      {},
	"JsonIter":      {},
	"SegmentioJson": {},
	"Sonic":         {},
	"SonicFast":     {},
	"SonicFastest":  {},
	"SonicStd":      {},
	"StdLib":        {},
}

// isComparedLibraryBenchmark reports whether the benchmark function measures a library other than go-json.
//
// A test binary identifies a benchmark only by its function name, so the library can't be carried
// by anything else: the benchmarks are named Benchmark_<operation>_<library>, and the library is
// the part after the last underscore. Every other name, including the one without an underscore,
// is a benchmark of go-json.
func isComparedLibraryBenchmark(name string) bool {
	i := strings.LastIndexByte(name, '_')
	if i < 0 {
		return false
	}
	_, exists := comparedLibraries[name[i+1:]]
	return exists
}

// The groups of the benchmarks, whose means are judged apart: a change of the decoder moves only the decode
// benchmarks, which are fewer than the encode ones, so in a mean of both it would count for less than one of
// the encoder.
const (
	groupEncode = "encode"
	groupDecode = "decode"
)

// benchmarkGroup returns the group of the benchmark function.
//
// A test binary identifies a benchmark only by its function name, so the group can't be carried by anything
// else: a benchmark of decoding has Decode or Unmarshal in its name ( BenchmarkCodeDecoder, Benchmark_Decode_*,
// BenchmarkUnmarshal* ), and every other one is a benchmark of encoding.
func benchmarkGroup(name string) string {
	if strings.Contains(name, "Decode") || strings.Contains(name, "Unmarshal") {
		return groupDecode
	}
	return groupEncode
}

// funcs returns the names of the top-level benchmark functions of go-json matched by pattern.
func (s *suite) funcs(ctx context.Context, pattern string) ([]string, error) {
	// every layout has the same functions.
	out, err := s.exec(ctx, s.binaries[0], "-test.list", pattern)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, name := range strings.Fields(out) {
		// -test.list also prints tests, fuzz targets and examples.
		// The testing package distinguishes them only by the prefix of the function name.
		if !strings.HasPrefix(name, benchmarkPrefix) {
			continue
		}
		if isComparedLibraryBenchmark(name) {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

// run measures a top-level benchmark function, including its sub-benchmarks, with the binary of the layout.
// The result is empty if the suite doesn't have the function.
func (s *suite) run(ctx context.Context, layout int, fn, benchTime string) (measurement, error) {
	args := []string{"-test.run", "^$", "-test.bench", "^" + regexp.QuoteMeta(fn) + "$", "-test.count", "1"}
	if benchTime != "" {
		args = append(args, "-test.benchtime", benchTime)
	}
	out, err := s.exec(ctx, s.binaries[layout], args...)
	if err != nil {
		return nil, err
	}
	return parseBenchOutput(out, runtime.GOMAXPROCS(0)), nil
}

package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The durations are far longer than the resolution of the timer of any platform
// ( about 16ms at worst ), so that a single iteration is always measured as a positive time
// and the slow one is always distinguishable from the fast one.
const (
	testLibFast = `package lib

import "time"

func Work() int {
	time.Sleep(20 * time.Millisecond)
	return 0
}
`
	testLibFastRefactored = `package lib

import "time"

const workDuration = 20 * time.Millisecond

func Work() int {
	time.Sleep(workDuration)
	return 0
}
`
	testLibSlow = `package lib

import "time"

func Work() int {
	time.Sleep(100 * time.Millisecond)
	return 0
}
`
	// testLibExtra is a new API in a file of its own.
	testLibExtra = `package lib

func Extra() int {
	return Work() + 1
}
`
	testBench = `package bench

import (
	"testing"

	"example.com/lib"
)

func BenchmarkWork(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lib.Work()
	}
}

func BenchmarkSub(b *testing.B) {
	for _, name := range []string{"a", "b"} {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lib.Work()
			}
		})
	}
}
`
	// testRecordDirEnv is the directory where a benchmark records the binary which ran it: a file named after the
	// binary has the symbols of the binary, with their addresses and sizes ( go tool nm -size ).
	testRecordDirEnv         = "BENCHCHECK_TEST_RECORD_DIR"
	testBenchRecordingBinary = `package bench

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"example.com/lib"
)

func BenchmarkWork(b *testing.B) {
	symbols, err := exec.Command("go", "tool", "nm", "-size", os.Args[0]).Output()
	if err != nil {
		b.Fatal(err)
	}
	name := filepath.Join(os.Getenv("` + testRecordDirEnv + `"), filepath.Base(os.Args[0]))
	if err := os.WriteFile(name, symbols, 0o600); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		lib.Work()
	}
}
`
	testBenchWithComparedLibrary = `package bench

import (
	"testing"

	"example.com/lib"
)

func BenchmarkWork(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lib.Work()
	}
}

func Benchmark_Work_GoJson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lib.Work()
	}
}

func Benchmark_Work_EncodingJson(b *testing.B) {
	b.Fatal("benchmark of the compared library must not run")
}
`
	testBenchWithNewAPI = testBench + `
func BenchmarkExtra(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lib.Extra()
	}
}
`
	// testLibDecodeUnsupported has an API which doesn't support its work yet, which testLibDecodeSupported does.
	testLibDecodeUnsupported = testLibFast + `
func Decode() error {
	return errNotSupported
}

type notSupported struct{}

func (notSupported) Error() string { return "not supported yet" }

var errNotSupported error = notSupported{}
`
	testLibDecodeSupported = testLibFast + `
func Decode() error {
	Work()
	return nil
}
`
	// testBenchDecode has two functions of Decode next to each other, so that one is measured with the base first,
	// and the other with the working tree first.
	testBenchDecode = testBench + `
func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := lib.Decode(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeTwice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := lib.Decode(); err != nil {
			b.Fatal(err)
		}
	}
}
`
	// testLibDecodeInterrupted has an API which interrupts the process, as a Ctrl-C does a test binary which a
	// benchcheck runs.
	testLibDecodeInterrupted = `package lib

import (
	"os"
	"time"
)

func Work() int {
	time.Sleep(20 * time.Millisecond)
	return 0
}

func Decode() error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	if err := p.Signal(os.Interrupt); err != nil {
		return err
	}
	// the signal ends the process
	time.Sleep(time.Minute)
	return nil
}
`
	testBenchDecodeOnly = `package bench

import (
	"testing"

	"example.com/lib"
)

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := lib.Decode(); err != nil {
			b.Fatal(err)
		}
	}
}
`
)

type testRepo struct {
	t   *testing.T
	dir string
	// layouts is the number of the function layouts of a measurement, and rounds the number of its rounds, as many
	// as the layouts if it is zero.
	layouts int
	rounds  int
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()
	for _, command := range []string{"git", "go"} {
		if _, err := exec.LookPath(command); err != nil {
			t.Skipf("%s command is required: %v", command, err)
		}
	}
	r := &testRepo{t: t, dir: t.TempDir(), layouts: 1}
	r.git("init", "-q")
	r.git("symbolic-ref", "HEAD", "refs/heads/master")
	r.write("go.mod", "module example.com/lib\n\ngo 1.19\n")
	r.write("lib.go", testLibFast)
	r.write(filepath.Join("benchmarks", "go.mod"), "module bench\n\ngo 1.19\n\nrequire example.com/lib v0.0.0\n\nreplace example.com/lib => ../\n")
	r.write(filepath.Join("benchmarks", "bench_test.go"), testBench)
	r.commit("initial commit")
	return r
}

func (r *testRepo) git(args ...string) string {
	r.t.Helper()
	args = append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("failed to run git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func (r *testRepo) write(name, content string) {
	r.t.Helper()
	path := filepath.Join(r.dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		r.t.Fatal(err)
	}
}

func (r *testRepo) commit(message string) {
	r.t.Helper()
	r.git("add", "go.mod", "lib.go", "benchmarks")
	r.git("commit", "-q", "-m", message)
}

func (r *testRepo) detect(tolerance float64) *outcome {
	r.t.Helper()
	out, err := r.detectResult(tolerance)
	if err != nil {
		r.t.Fatal(err)
	}
	return out
}

// detectResult is detect, which returns the error of the run too.
func (r *testRepo) detectResult(tolerance float64) (*outcome, error) {
	r.t.Helper()
	rounds := r.rounds
	if rounds == 0 {
		rounds = r.layouts
	}
	c := &detector{
		opt: &options{
			config:   benchConfig{Dir: "benchmarks", Bench: ".", BenchTime: "1x", Rounds: rounds, Layouts: r.layouts},
			attempts: defaultAttempts,
			// the tests slow the library down several times, which is beyond both of the tolerances.
			tolerance:       tolerance,
			singleTolerance: tolerance,
		},
		repo:  &repository{root: r.dir},
		cache: &cache{dir: filepath.Join(r.dir, ".git", "benchcheck"), readable: true},
	}
	out, err := c.detect(context.Background())
	// the temporary worktree of the base commit must always be removed.
	if got := strings.Count(r.git("worktree", "list", "--porcelain"), "worktree "); got != 1 {
		r.t.Fatalf("temporary worktree is left: %d worktrees exist", got)
	}
	return out, err
}

func statusOf(t *testing.T, v *verdict) map[string]status {
	t.Helper()
	if v == nil {
		t.Fatal("verdict is required")
	}
	statuses := map[string]status{}
	for _, result := range v.Benchmarks {
		statuses[result.Name] = result.Status
	}
	return statuses
}

func assertStatuses(t *testing.T, v *verdict, want map[string]status) {
	t.Helper()
	got := statusOf(t, v)
	if len(got) != len(want) {
		t.Fatalf("unexpected statuses: got %v, want %v", got, want)
	}
	for name, s := range want {
		if got[name] != s {
			t.Fatalf("unexpected statuses: got %v, want %v", got, want)
		}
	}
}

func TestDetect(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	allOK := map[string]status{
		"BenchmarkWork":  statusOK,
		"BenchmarkSub/a": statusOK,
		"BenchmarkSub/b": statusOK,
	}
	r := newTestRepo(t)

	t.Run("HEAD is the base commit", func(t *testing.T) {
		if out := r.detect(ignoreNoise); out.verdict != nil {
			t.Fatalf("nothing must be compared: %+v", out.verdict)
		}
	})

	t.Run("uncommitted changes on the base commit are measured", func(t *testing.T) {
		r.write("lib.go", testLibFastRefactored)
		out := r.detect(ignoreNoise)
		if out.cachedVerdict || out.cachedBaseline {
			t.Fatalf("nothing must be cached yet: %+v", out)
		}
		assertStatuses(t, out.verdict, allOK)
	})

	r.git("checkout", "-q", "-b", "feature")
	r.commit("refactor")

	t.Run("committed changes are measured with the cached result of the base", func(t *testing.T) {
		out := r.detect(ignoreNoise)
		if out.cachedVerdict {
			t.Fatal("verdict of uncommitted changes must not be cached")
		}
		if !out.cachedBaseline {
			t.Fatal("result of the base must be cached")
		}
		assertStatuses(t, out.verdict, allOK)
	})

	t.Run("verdict of the same commit is cached", func(t *testing.T) {
		out := r.detect(ignoreNoise)
		if !out.cachedVerdict {
			t.Fatal("verdict must be cached")
		}
		assertStatuses(t, out.verdict, allOK)
	})

	t.Run("verdict is not shared between different options", func(t *testing.T) {
		if out := r.detect(ignoreNoise / 10); out.cachedVerdict {
			t.Fatal("verdict of the other options must not be used")
		}
	})

	t.Run("uncommitted changes are always measured", func(t *testing.T) {
		r.write("lib.go", testLibSlow)
		for i := 0; i < 2; i++ {
			out := r.detect(defaultTolerance)
			if out.cachedVerdict {
				t.Fatal("verdict must not be cached")
			}
			if !out.cachedBaseline {
				t.Fatal("result of the base must be cached")
			}
			assertStatuses(t, out.verdict, map[string]status{
				"BenchmarkWork":  statusRegressed,
				"BenchmarkSub/a": statusRegressed,
				"BenchmarkSub/b": statusRegressed,
			})
			for _, result := range out.verdict.Benchmarks {
				if result.Attempts != defaultAttempts {
					t.Fatalf("all attempts must be used: %+v", result)
				}
			}
			if err := report(out.verdict); !errors.Is(err, errDegraded) {
				t.Fatalf("degradation must be reported: %v", err)
			}
		}
	})

	t.Run("degraded commit is cached as degraded", func(t *testing.T) {
		r.commit("degrade")
		if out := r.detect(defaultTolerance); out.cachedVerdict {
			t.Fatal("verdict must not be cached yet")
		}
		out := r.detect(defaultTolerance)
		if !out.cachedVerdict {
			t.Fatal("verdict must be cached")
		}
		if err := report(out.verdict); !errors.Is(err, errDegraded) {
			t.Fatalf("degradation must be reported: %v", err)
		}
	})

	t.Run("benchmark of the API which the base doesn't have", func(t *testing.T) {
		r.write("lib.go", testLibFast)
		// the head has the files which are not tracked yet.
		r.write("extra.go", testLibExtra)
		r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchWithNewAPI)
		out := r.detect(ignoreNoise)
		if out.cachedBaseline {
			t.Fatal("result of the base measured by the other benchmarks must not be used")
		}
		assertStatuses(t, out.verdict, map[string]status{
			"BenchmarkWork":  statusOK,
			"BenchmarkSub/a": statusOK,
			"BenchmarkSub/b": statusOK,
			"BenchmarkExtra": statusNew,
		})
	})
}

func TestDetectBenchmarkWhichFailsInTheBase(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	t.Run("is new", func(t *testing.T) {
		r := newTestRepo(t)
		r.layouts, r.rounds = 2, 2
		r.write("lib.go", testLibDecodeUnsupported)
		r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchDecode)
		r.commit("an API which doesn't support its work yet")
		r.write("lib.go", testLibDecodeSupported)
		want := map[string]status{
			"BenchmarkWork":        statusOK,
			"BenchmarkSub/a":       statusOK,
			"BenchmarkSub/b":       statusOK,
			"BenchmarkDecode":      statusNew,
			"BenchmarkDecodeTwice": statusNew,
		}
		out := r.detect(ignoreNoise)
		assertStatuses(t, out.verdict, want)
		// the results of the base are not kept: the base is measured again, with the same statuses.
		out = r.detect(ignoreNoise)
		if out.cachedBaseline {
			t.Fatal("the results of a base which failed a function must not be kept")
		}
		assertStatuses(t, out.verdict, want)
	})
	t.Run("of every function is an error", func(t *testing.T) {
		r := newTestRepo(t)
		r.write("lib.go", testLibDecodeUnsupported)
		r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchDecodeOnly)
		r.commit("only an API which doesn't support its work yet")
		r.write("lib.go", testLibDecodeSupported)
		if _, err := r.detectResult(ignoreNoise); !errors.Is(err, errBaseFailsEveryFunc) {
			t.Fatalf("got %v, want the error of a base which fails every benchmark function", err)
		}
	})
	t.Run("ended by a signal is an error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("a process can't send itself an interrupt on Windows")
		}
		r := newTestRepo(t)
		r.write("lib.go", testLibDecodeInterrupted)
		r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchDecode)
		r.commit("an API which interrupts the process")
		r.write("lib.go", testLibDecodeSupported)
		if _, err := r.detectResult(ignoreNoise); err == nil || errors.Is(err, errBaseFailsEveryFunc) {
			t.Fatalf("got %v, want the error of the base which was ended by a signal", err)
		}
	})
	t.Run("in the working tree is an error", func(t *testing.T) {
		r := newTestRepo(t)
		r.write("lib.go", testLibDecodeSupported)
		r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchDecode)
		r.commit("an API which supports its work")
		r.write("lib.go", testLibDecodeUnsupported)
		var exitErr *exec.ExitError
		if _, err := r.detectResult(ignoreNoise); !errors.As(err, &exitErr) || errors.Is(err, errBaseFailsEveryFunc) {
			t.Fatalf("got %v, want the error of the benchmark which fails in the working tree", err)
		}
	})
}

func TestDetectBuildsBothSidesAlike(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	r := newTestRepo(t)
	r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchRecordingBinary)
	r.commit("record the binaries")
	// a change which is not code: both sides build identical code.
	r.write("notes.txt", "notes\n")
	recordDir := t.TempDir()
	t.Setenv(testRecordDirEnv, recordDir)
	out := r.detect(ignoreNoise)
	assertStatuses(t, out.verdict, map[string]status{"BenchmarkWork": statusOK})

	symbols := map[string][]string{}
	for _, side := range []string{"base", "head"} {
		b, err := os.ReadFile(filepath.Join(recordDir, filepath.Base(binaryPath("", side+".layout0.test"))))
		if err != nil {
			t.Fatal(err)
		}
		symbols[side] = strings.Split(string(b), "\n")
	}
	// the paths of the files of the library are in the binaries, and a path of another length moves the symbols
	// after them.
	base, head := symbols["base"], symbols["head"]
	if len(base) != len(head) {
		t.Fatalf("the binaries of identical code have %d and %d symbols", len(base), len(head))
	}
	for i := range base {
		if base[i] != head[i] {
			t.Fatalf("the binaries of identical code have other symbols:\nbase: %s\nhead: %s", base[i], head[i])
		}
	}
}

func TestCopyWorkingTree(t *testing.T) {
	r := newTestRepo(t)
	r.write("notes.txt", "notes\n")
	r.write(filepath.Join(".config", "settings"), "tracked\n")
	r.write(".gitignore", "ignored.txt\n")
	r.git("add", "notes.txt", ".config", ".gitignore")
	r.commit("add files which are not code")

	r.write("lib.go", testLibSlow)
	if err := os.Remove(filepath.Join(r.dir, "notes.txt")); err != nil {
		t.Fatal(err)
	}
	r.write("extra.go", testLibExtra)
	r.write(filepath.Join(".scratch", "out.txt"), "scratch\n")
	r.write(filepath.Join("_x", "broken.go"), "broken\n")
	r.write("ignored.txt", "ignored\n")

	dst := t.TempDir()
	if err := (&repository{root: r.dir}).copyWorkingTree(context.Background(), dst); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"go.mod": "module example.com/lib\n\ngo 1.19\n",
		// the changes which are not committed yet.
		"lib.go":   testLibSlow,
		"extra.go": testLibExtra,
		filepath.Join("benchmarks", "bench_test.go"): testBench,
		// tracked files are copied wherever they are.
		filepath.Join(".config", "settings"): "tracked\n",
		".gitignore":                         "ignored.txt\n",
	} {
		got, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Fatalf("%s must be copied: %v", name, err)
		}
		if string(got) != want {
			t.Fatalf("%s: got %q, want %q", name, got, want)
		}
	}
	for _, name := range []string{
		// removed and not committed yet.
		"notes.txt",
		// untracked files which the go command doesn't read.
		filepath.Join(".scratch", "out.txt"),
		filepath.Join("_x", "broken.go"),
		"ignored.txt",
		// the repository itself.
		".git",
	} {
		if _, err := os.Lstat(filepath.Join(dst, name)); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("%s must not be copied: %v", name, err)
		}
	}
}

func TestIgnoredByGo(t *testing.T) {
	for name, want := range map[string]bool{
		"encode.go":                        false,
		"internal/encoder/vm/vm.go":        false,
		"benchmarks/testdata/code.json.gz": false,
		".prof/out.txt":                    true,
		"benchmarks/_scratch/a.go":         true,
		"internal/.cache/b":                true,
		".golangci.yml":                    true,
	} {
		if got := ignoredByGo(name); got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
}

func TestDetectIgnoresComparedLibraries(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	r := newTestRepo(t)
	r.write("lib.go", testLibFastRefactored)
	r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchWithComparedLibrary)
	out := r.detect(ignoreNoise)
	assertStatuses(t, out.verdict, map[string]status{
		"Benchmark_Work_GoJson": statusOK,
		"BenchmarkWork":         statusOK,
	})
}

func TestIsComparedLibraryBenchmark(t *testing.T) {
	for name, want := range map[string]bool{
		"Benchmark_Encode_SmallStruct_EncodingJson":                    true,
		"Benchmark_Decode_SmallStruct_GoJayUnsafe":                     true,
		"Benchmark_Decode_GitHubREST_Unmarshal_SonicFastestValidating": true,
		"Benchmark_Encode_SmallStruct_GoJson":                          false,
		"Benchmark_Encode_SmallStruct_GoJsonNoEscape":                  false,
		"Benchmark_Encode_FilterByMap":                                 false,
		"BenchmarkUnmarshalString":                                     false,
		// the library is only the part after the last underscore.
		"Benchmark_EncodingJson_Compatible_GoJson": false,
	} {
		if got := isComparedLibraryBenchmark(name); got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
}

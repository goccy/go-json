package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
	testLibWithNewAPI = testLibFast + `
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
	// testRecordDirEnv is the directory where a benchmark records the name of the binary which ran it.
	testRecordDirEnv         = "BENCHCHECK_TEST_RECORD_DIR"
	testBenchRecordingBinary = `package bench

import (
	"os"
	"path/filepath"
	"testing"

	"example.com/lib"
)

func BenchmarkWork(b *testing.B) {
	name := filepath.Join(os.Getenv("` + testRecordDirEnv + `"), filepath.Base(os.Args[0]))
	if err := os.WriteFile(name, nil, 0o600); err != nil {
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
)

type testRepo struct {
	t   *testing.T
	dir string
	// layouts is the number of the function layouts ( and of the rounds ) of a measurement.
	layouts int
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

func (r *testRepo) check(tolerance float64) *outcome {
	r.t.Helper()
	c := &checker{
		opt: &options{
			config:   benchConfig{Dir: "benchmarks", Bench: ".", BenchTime: "1x", Rounds: r.layouts, Layouts: r.layouts},
			attempts: defaultAttempts,
			// the tests slow the library down several times, which is beyond both of the tolerances.
			tolerance:       tolerance,
			singleTolerance: tolerance,
		},
		repo:  &repository{root: r.dir},
		cache: &cache{dir: filepath.Join(r.dir, ".git", "benchcheck"), readable: true},
	}
	out, err := c.check(context.Background())
	if err != nil {
		r.t.Fatal(err)
	}
	// the temporary worktree of the base commit must always be removed.
	if got := strings.Count(r.git("worktree", "list", "--porcelain"), "worktree "); got != 1 {
		r.t.Fatalf("temporary worktree is left: %d worktrees exist", got)
	}
	return out
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

func TestCheck(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	allOK := map[string]status{
		"BenchmarkWork":  statusOK,
		"BenchmarkSub/a": statusOK,
		"BenchmarkSub/b": statusOK,
	}
	r := newTestRepo(t)

	t.Run("HEAD is the base commit", func(t *testing.T) {
		if out := r.check(ignoreNoise); out.verdict != nil {
			t.Fatalf("nothing must be compared: %+v", out.verdict)
		}
	})

	t.Run("uncommitted changes on the base commit are measured", func(t *testing.T) {
		r.write("lib.go", testLibFastRefactored)
		out := r.check(ignoreNoise)
		if out.cachedVerdict || out.cachedBaseline {
			t.Fatalf("nothing must be cached yet: %+v", out)
		}
		assertStatuses(t, out.verdict, allOK)
	})

	r.git("checkout", "-q", "-b", "feature")
	r.commit("refactor")

	t.Run("committed changes are measured with the cached result of the base", func(t *testing.T) {
		out := r.check(ignoreNoise)
		if out.cachedVerdict {
			t.Fatal("verdict of uncommitted changes must not be cached")
		}
		if !out.cachedBaseline {
			t.Fatal("result of the base must be cached")
		}
		assertStatuses(t, out.verdict, allOK)
	})

	t.Run("verdict of the same commit is cached", func(t *testing.T) {
		out := r.check(ignoreNoise)
		if !out.cachedVerdict {
			t.Fatal("verdict must be cached")
		}
		assertStatuses(t, out.verdict, allOK)
	})

	t.Run("verdict is not shared between different options", func(t *testing.T) {
		if out := r.check(ignoreNoise / 10); out.cachedVerdict {
			t.Fatal("verdict of the other options must not be used")
		}
	})

	t.Run("uncommitted changes are always measured", func(t *testing.T) {
		r.write("lib.go", testLibSlow)
		for i := 0; i < 2; i++ {
			out := r.check(defaultTolerance)
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
		if out := r.check(defaultTolerance); out.cachedVerdict {
			t.Fatal("verdict must not be cached yet")
		}
		out := r.check(defaultTolerance)
		if !out.cachedVerdict {
			t.Fatal("verdict must be cached")
		}
		if err := report(out.verdict); !errors.Is(err, errDegraded) {
			t.Fatalf("degradation must be reported: %v", err)
		}
	})

	t.Run("benchmark of the API which the base doesn't have", func(t *testing.T) {
		r.write("lib.go", testLibWithNewAPI)
		r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchWithNewAPI)
		out := r.check(ignoreNoise)
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

func TestCheckIgnoresComparedLibraries(t *testing.T) {
	// the tolerance to make the verdict independent of the measurement noise.
	const ignoreNoise = 1e12

	r := newTestRepo(t)
	r.write("lib.go", testLibFastRefactored)
	r.write(filepath.Join("benchmarks", "bench_test.go"), testBenchWithComparedLibrary)
	out := r.check(ignoreNoise)
	assertStatuses(t, out.verdict, map[string]status{
		"Benchmark_Work_GoJson": statusOK,
		"BenchmarkWork":         statusOK,
	})
}

func TestIsComparedLibraryBenchmark(t *testing.T) {
	for name, want := range map[string]bool{
		"Benchmark_Encode_SmallStruct_EncodingJson":   true,
		"Benchmark_Decode_SmallStruct_GoJayUnsafe":    true,
		"Benchmark_Encode_SmallStruct_GoJson":         false,
		"Benchmark_Encode_SmallStruct_GoJsonNoEscape": false,
		"Benchmark_Encode_FilterByMap":                false,
		"BenchmarkUnmarshalString":                    false,
		// the library is only the part after the last underscore.
		"Benchmark_EncodingJson_Compatible_GoJson": false,
	} {
		if got := isComparedLibraryBenchmark(name); got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
}

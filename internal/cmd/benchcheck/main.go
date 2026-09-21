// benchcheck detects performance degradation by comparing the benchmark results
// of the current working tree with those of the base branch.
//
// A benchmark is treated as degraded if it never reaches the result of the base branch
// within the allowed number of attempts. Every benchmark must reach it.
//
// To make the comparison robust against the load of the machine changing over time,
// the base and the working tree are measured alternately for each benchmark function.
// An additional attempt measures both of them again for the benchmarks which have not reached the base.
//
// The measured results are cached by commit hash:
//   - the result of the base branch is reused as long as it is measured on the same machine.
//   - if the working tree is clean, the verdict for the pair of HEAD and the base commit is reused,
//     so that the command ends immediately.
//   - if the working tree has uncommitted changes, the working tree is always measured.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"
)

const (
	defaultAttempts = 3
	// A measurement consists of multiple short rounds rather than a single long run,
	// because the fastest round is hardly affected by the temporary load of the machine.
	// Every layout is measured once by default.
	defaultRounds = 4
	// defaultLayouts is the number of the function layouts with which a benchmark is measured.
	// The fastest layout of each side is compared, so the more layouts are measured,
	// the less the result depends on how well a single layout happens to fit the code.
	defaultLayouts   = 4
	defaultBenchTime = "300ms"
	// defaultTolerance absorbs the run-to-run noise of identical code measured on the same machine.
	defaultTolerance = 5.0
)

var errDegraded = errors.New("performance degradation detected")

type options struct {
	baseRef   string
	config    benchConfig
	attempts  int
	tolerance float64
	cacheDir  string
	noCache   bool
}

func parseOptions() (*options, error) {
	var opt options
	flag.StringVar(&opt.baseRef, "base", "", "git ref to compare with ( default: master, or origin/master if master doesn't exist )")
	flag.StringVar(&opt.config.Dir, "dir", "benchmarks", "directory of the benchmarks, relative to the repository root")
	flag.StringVar(&opt.config.Bench, "bench", ".", "pattern of the top-level benchmark functions to run")
	flag.StringVar(&opt.config.BenchTime, "benchtime", defaultBenchTime, "duration of each round of a benchmark ( go test -benchtime )")
	flag.IntVar(&opt.config.Rounds, "rounds", defaultRounds, "number of rounds of a measurement: the fastest round is the result of the measurement")
	flag.IntVar(&opt.config.Layouts, "layouts", defaultLayouts, "number of function layouts of the benchmark binary: the round N is measured with the layout N % layouts ( more than 1 requires Go 1.23 or later )")
	flag.IntVar(&opt.attempts, "attempts", defaultAttempts, "max number of measurements to reach the result of the base")
	flag.Float64Var(&opt.tolerance, "tolerance", defaultTolerance, "measurement noise ( in percent ) which is not treated as a degradation")
	flag.StringVar(&opt.cacheDir, "cache-dir", "", "directory to store the results ( default: <git common dir>/benchcheck )")
	flag.BoolVar(&opt.noCache, "no-cache", false, "ignore the cached results and measure again")
	flag.Parse()
	if flag.NArg() != 0 {
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(flag.Args(), " "))
	}
	if opt.config.Rounds < 1 {
		return nil, fmt.Errorf("rounds must be greater than zero: %d", opt.config.Rounds)
	}
	if opt.config.Layouts < 1 {
		return nil, fmt.Errorf("layouts must be greater than zero: %d", opt.config.Layouts)
	}
	if opt.config.Layouts > opt.config.Rounds {
		return nil, fmt.Errorf("layouts must not be greater than rounds, otherwise some of them are never measured: layouts %d, rounds %d", opt.config.Layouts, opt.config.Rounds)
	}
	if opt.attempts < 1 {
		return nil, fmt.Errorf("attempts must be greater than zero: %d", opt.attempts)
	}
	if opt.tolerance < 0 {
		return nil, fmt.Errorf("tolerance must not be negative: %v", opt.tolerance)
	}
	return &opt, nil
}

func currentMachine(ctx context.Context) (machine, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return machine{}, err
	}
	// the version of the go command which builds the benchmarks.
	out, err := exec.CommandContext(ctx, "go", "env", "GOVERSION").Output()
	if err != nil {
		return machine{}, fmt.Errorf("failed to get go version: %w", err)
	}
	return machine{
		Hostname:  hostname,
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
		GoVersion: strings.TrimSpace(string(out)),
	}, nil
}

func binaryPath(dir, name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

type checker struct {
	opt      *options
	repo     *repository
	cache    *cache
	tmpDir   string
	benchDir string

	baseCommit string
	base       *suite
	cleanups   []func()
}

func (c *checker) cleanup() {
	for i := len(c.cleanups) - 1; i >= 0; i-- {
		c.cleanups[i]()
	}
	c.cleanups = nil
	c.base = nil
}

// baseSuite builds the benchmarks for the base commit on the first call.
//
// The benchmarks of the working tree are built with the library of the base commit,
// so that exactly the same benchmarks are compared.
// If they can't be built with the base commit ( e.g. they use an API added after it ),
// the benchmarks of the base commit itself are used instead.
func (c *checker) baseSuite(ctx context.Context) (*suite, error) {
	if c.base != nil {
		return c.base, nil
	}
	fmt.Printf("build benchmarks for base %s\n", c.baseCommit)
	baseDir := filepath.Join(c.tmpDir, "base")
	cleanup, err := c.repo.checkout(ctx, baseDir, c.baseCommit)
	if err != nil {
		return nil, err
	}
	c.cleanups = append(c.cleanups, cleanup)

	modFile, err := createModFile(ctx, c.tmpDir, c.benchDir, baseDir)
	if err != nil {
		return nil, err
	}
	base, err := buildSuite(ctx, c.benchDir, modFile, c.tmpDir, "base", c.opt.config.Layouts)
	if err != nil {
		fmt.Printf("benchmarks of the working tree can't be built with base %s: use the benchmarks of the base\n", c.baseCommit)
		base, err = buildSuite(ctx, filepath.Join(baseDir, c.opt.config.Dir), "", c.tmpDir, "base", c.opt.config.Layouts)
		if err != nil {
			return nil, err
		}
	}
	c.base = base
	return base, nil
}

// formatDelta returns how much slower headNs is than baseNs, in percent.
func formatDelta(baseNs, headNs float64) string {
	if baseNs == 0 {
		// the ratio to zero can't be represented.
		return "-"
	}
	return fmt.Sprintf("%+.2f%%", (headNs-baseNs)/baseNs*100)
}

func printProgress(base, head measurement) {
	names := make([]string, 0, len(head))
	for name := range head {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		headNs := head[name]
		baseNs, exists := base[name]
		if !exists {
			fmt.Printf("%s\tbase: -\thead: %.2f ns/op\n", name, headNs)
			continue
		}
		fmt.Printf("%s\tbase: %.2f ns/op\thead: %.2f ns/op\t%s\n", name, baseNs, headNs, formatDelta(baseNs, headNs))
	}
}

// measureFunc measures the benchmark function fn of the base and of the working tree.
// The rounds of both sides are interleaved so that they run under the same load,
// and each round uses the next layout so that the fastest layout of each side is compared.
// If base is nil, only the working tree is measured.
func (c *checker) measureFunc(ctx context.Context, fn string, base, head *suite, baseFirst bool) (measurement, measurement, error) {
	type side struct {
		suite  *suite
		result measurement
	}
	baseSide := &side{suite: base, result: measurement{}}
	headSide := &side{suite: head, result: measurement{}}
	sides := []*side{baseSide, headSide}
	if !baseFirst {
		sides = []*side{headSide, baseSide}
	}
	for round := 0; round < c.opt.config.Rounds; round++ {
		for _, s := range sides {
			if s.suite == nil {
				continue
			}
			result, err := s.suite.run(ctx, round%c.opt.config.Layouts, fn, c.opt.config.BenchTime)
			if err != nil {
				return nil, nil, err
			}
			s.result.mergeFastest(result)
		}
	}
	return baseSide.result, headSide.result, nil
}

// measure runs the benchmarks until all of them reach the result of the base or the attempts are exhausted.
func (c *checker) measure(ctx context.Context, out *outcome) ([]benchResult, error) {
	m, err := currentMachine(ctx)
	if err != nil {
		return nil, err
	}
	benchHash, err := hashDir(c.benchDir)
	if err != nil {
		return nil, err
	}
	key := baselineKey{Config: c.opt.config, Machine: m, BenchHash: benchHash}
	cached, err := c.cache.loadBaseline(c.baseCommit, key)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		fmt.Printf("use cached benchmark results of base %s\n", c.baseCommit)
		out.cachedBaseline = true
	}

	fmt.Println("build benchmarks for working tree")
	head, err := buildSuite(ctx, c.benchDir, "", c.tmpDir, "head", c.opt.config.Layouts)
	if err != nil {
		return nil, err
	}
	funcs, err := head.funcs(ctx, c.opt.config.Bench)
	if err != nil {
		return nil, err
	}
	if len(funcs) == 0 {
		return nil, fmt.Errorf("no benchmark matched %q in %s", c.opt.config.Bench, c.benchDir)
	}

	cmp := newComparison(c.opt.tolerance)
	for attempt := 1; attempt <= c.opt.attempts && len(funcs) != 0; attempt++ {
		fmt.Printf("measure %d benchmark functions ( attempt %d/%d )\n", len(funcs), attempt, c.opt.attempts)
		// the cached results are valid only for the first attempt:
		// a retry exists to get rid of the noise, so both sides are measured again at the same time.
		useCache := attempt == 1 && cached != nil
		measured := measurement{}
		for i, fn := range funcs {
			var baseSuite *suite
			if !useCache {
				baseSuite, err = c.baseSuite(ctx)
				if err != nil {
					return nil, err
				}
			}
			// alternate the order so that neither side is systematically favored.
			baseResult, headResult, err := c.measureFunc(ctx, fn, baseSuite, head, i%2 == 0)
			if err != nil {
				return nil, err
			}
			if useCache {
				baseResult = cached.Results
			} else {
				measured.mergeFastest(baseResult)
			}
			cmp.add(fn, baseResult, headResult)
			printProgress(baseResult, headResult)
		}
		if attempt == 1 && cached == nil {
			if err := c.cache.storeBaseline(&baseline{Commit: c.baseCommit, Key: key, Results: measured}); err != nil {
				return nil, err
			}
		}
		funcs = cmp.pendingFuncs()
	}
	return cmp.sortedResults(), nil
}

// outcome is the result of a check.
type outcome struct {
	// verdict is nil if there is nothing to compare, because HEAD is the base commit without any change.
	verdict        *verdict
	cachedVerdict  bool
	cachedBaseline bool
}

func (c *checker) check(ctx context.Context) (*outcome, error) {
	head, err := c.repo.commit(ctx, "HEAD")
	if err != nil {
		return nil, err
	}
	base, err := c.repo.baseCommit(ctx, c.opt.baseRef)
	if err != nil {
		return nil, err
	}
	dirty, err := c.repo.dirty(ctx)
	if err != nil {
		return nil, err
	}
	if head == base && !dirty {
		fmt.Printf("HEAD is the base commit %s: nothing to compare\n", base)
		return &outcome{}, nil
	}

	key := verdictKey{Config: c.opt.config, Attempts: c.opt.attempts, Tolerance: c.opt.tolerance}
	if !dirty {
		cached, err := c.cache.loadVerdict(head, base, key)
		if err != nil {
			return nil, err
		}
		if cached != nil {
			fmt.Printf("use cached verdict of %s against base %s\n", head, base)
			return &outcome{verdict: cached, cachedVerdict: true}, nil
		}
	}

	tmpDir, err := os.MkdirTemp("", "benchcheck-")
	if err != nil {
		return nil, err
	}
	c.cleanups = append(c.cleanups, func() { os.RemoveAll(tmpDir) })
	defer c.cleanup()
	c.tmpDir = tmpDir
	c.benchDir = filepath.Join(c.repo.root, c.opt.config.Dir)
	c.baseCommit = base

	out := &outcome{}
	results, err := c.measure(ctx, out)
	if err != nil {
		return nil, err
	}
	out.verdict = &verdict{Head: head, Base: base, Key: key, Benchmarks: results}
	if !dirty {
		if err := c.cache.storeVerdict(out.verdict); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (c *checker) run(ctx context.Context) error {
	out, err := c.check(ctx)
	if err != nil {
		return err
	}
	if out.verdict == nil {
		return nil
	}
	return report(out.verdict)
}

func report(v *verdict) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "benchmark\tbase ns/op\thead ns/op\tdelta\tattempts\tstatus")
	regressed := []string{}
	for _, result := range v.Benchmarks {
		switch result.Status {
		case statusOK, statusRegressed:
			delta := formatDelta(result.BaseNs, result.HeadNs)
			fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%s\t%d\t%s\n", result.Name, result.BaseNs, result.HeadNs, delta, result.Attempts, result.Status)
		case statusNew:
			fmt.Fprintf(w, "%s\t-\t%.2f\t-\t%d\t%s\n", result.Name, result.HeadNs, result.Attempts, result.Status)
		}
		if result.Status == statusRegressed {
			regressed = append(regressed, result.Name)
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if len(regressed) == 0 {
		fmt.Printf("OK: all benchmarks reached base %s ( tolerance %v%% )\n", v.Base, v.Key.Tolerance)
		return nil
	}
	fmt.Printf("FAIL: %d benchmarks never reached base %s within %d attempts ( tolerance %v%% )\n", len(regressed), v.Base, v.Key.Attempts, v.Key.Tolerance)
	for _, name := range regressed {
		fmt.Printf("  %s\n", name)
	}
	return errDegraded
}

func _main() error {
	opt, err := parseOptions()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, err := openRepository(ctx)
	if err != nil {
		return err
	}
	cacheDir := opt.cacheDir
	if cacheDir == "" {
		cacheDir, err = repo.defaultCacheDir(ctx)
		if err != nil {
			return err
		}
	}
	c := &checker{
		opt:   opt,
		repo:  repo,
		cache: &cache{dir: cacheDir, readable: !opt.noCache},
	}
	return c.run(ctx)
}

func main() {
	if err := _main(); err != nil {
		if !errors.Is(err, errDegraded) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

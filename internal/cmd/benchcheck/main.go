// benchcheck detects performance degradation by comparing the benchmark results
// of the current working tree with those of the base branch.
//
// The working tree is treated as degraded if the mean of all the benchmarks is slower than the base beyond
// the tolerance, or if a single benchmark is slower beyond the tolerance of a single benchmark. A single
// benchmark of identical code differs by several percent on a shared machine, so its tolerance is large
// and the mean, whose noise is far smaller, is what notices a small degradation.
//
// Both sides are built alike, from copies of their libraries in one temporary directory, and linked with several
// function layouts. A benchmark is compared by the fastest of its layouts and by their median, and is slower
// beyond the tolerance only if it is by both: see comparison.judge.
//
// To make the comparison robust against the load of the machine changing over time,
// the base and the working tree are measured alternately for each benchmark function.
// While the working tree is degraded, an additional attempt measures both of them again for the benchmarks
// which make the mean slow. The noise of a measurement only makes it slower, so the fastest result of all
// the attempts is what is kept, for each layout of each side.
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
	"io/fs"
	"math"
	"math/rand/v2"
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
	defaultRounds = 8
	// defaultLayouts is the number of the function layouts with which a benchmark is measured.
	// The fastest layout of each side and the median of the layouts are compared, so the more layouts are
	// measured, the less the result depends on how well a single layout happens to fit the code.
	//
	// The layouts and the rounds are by -calibrate on the machines of the CI ( 37 decode benchmarks, two
	// machines for each setting ): identical code linked with other layouts differs by at most 2.4%, 1.8% at
	// the 90th percentile, with 8 layouts of 300ms; 4 layouts left up to 5.8%, and more rounds of the same
	// layouts gained nothing, since how the functions are laid out moves a benchmark more than the load does.
	defaultLayouts   = 8
	defaultBenchTime = "300ms"
	// defaultTolerance is for the mean of all the benchmarks. The mean of identical code measured
	// on a shared machine stays within about 1.5%.
	defaultTolerance = 3.0
	// defaultSingleTolerance is for a single benchmark, which differs by up to about 8% for identical code
	// on a shared machine even if the function layout is varied, and by up to 50% in the fastest layout or in the
	// median on the machines where the layouts of a benchmark have two clusters of results ( see suite ). It
	// exists to notice the degradation of only a few benchmarks, which hardly moves the mean.
	defaultSingleTolerance = 15.0
)

var errDegraded = errors.New("performance degradation detected")

type options struct {
	baseRef  string
	config   benchConfig
	attempts int
	// tolerance is for the mean of all the benchmarks, and singleTolerance is for a single benchmark.
	tolerance       float64
	singleTolerance float64
	cacheDir        string
	noCache         bool
	// calibrate measures HEAD against itself, to tell how much the measurement of identical code differs.
	calibrate bool
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
	flag.StringVar(&opt.config.Group, "group", "", "encode or decode: measure only the benchmarks of the group, whose mean is judged apart from the other group ( default: both, in one mean )")
	shard := flag.String("shard", "", "i/n: measure only the i-th of n parts of the benchmark functions ( 0 <= i < n ), so that n machines measure all of them in parallel. The mean is the one of the part")
	flag.Float64Var(&opt.tolerance, "tolerance", defaultTolerance, "how much slower ( in percent ) the mean of all the benchmarks may be")
	flag.Float64Var(&opt.singleTolerance, "single-tolerance", defaultSingleTolerance, "how much slower ( in percent ) a single benchmark may be")
	flag.StringVar(&opt.cacheDir, "cache-dir", "", "directory to store the results ( default: <git common dir>/benchcheck )")
	flag.BoolVar(&opt.noCache, "no-cache", false, "ignore the cached results and measure again")
	flag.BoolVar(&opt.calibrate, "calibrate", false, "measure HEAD against itself, the head linked with other layouts than the base, and print how much the results differ: the noise of the measurement which a comparison has to exceed ( never cached )")
	flag.Parse()
	if flag.NArg() != 0 {
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(flag.Args(), " "))
	}
	switch opt.config.Group {
	case "", groupEncode, groupDecode:
	default:
		return nil, fmt.Errorf("group must be %s or %s: %q", groupEncode, groupDecode, opt.config.Group)
	}
	opt.config.Shards = 1
	if *shard != "" {
		if _, err := fmt.Sscanf(*shard, "%d/%d", &opt.config.Shard, &opt.config.Shards); err != nil {
			return nil, fmt.Errorf("shard must be i/n: %q", *shard)
		}
		if opt.config.Shards < 1 || opt.config.Shard < 0 || opt.config.Shard >= opt.config.Shards {
			return nil, fmt.Errorf("shard must be i/n with 0 <= i < n: %q", *shard)
		}
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
	if opt.singleTolerance < 0 {
		return nil, fmt.Errorf("single-tolerance must not be negative: %v", opt.singleTolerance)
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
		TempDir:   os.TempDir(),
	}, nil
}

func binaryPath(dir, name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

type checker struct {
	opt   *options
	repo  *repository
	cache *cache
	// tmpDir has the libraries of both sides, the base at base and the working tree at head ( see headSuite ).
	tmpDir string
	// benchDir is the directory of the benchmarks of the working tree, in its copy.
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

// headSuite copies the working tree to head in tmpDir and builds its benchmarks there, with the library of the copy.
//
// The binaries of both sides are built alike: the library of each side is in tmpDir, at base or at head, and is
// what the replace directive of its own copy of go.mod points to ( see createModFile ). The paths of the files of
// the library are in the binary, and a path of another length moves the data after them: the same code built in
// two directories differed by 12 KB in the addresses of its data, and the fastest layout of a small encode
// benchmark by 12%. With paths of the same length, the binaries of identical code have the same layout.
func (c *checker) headSuite(ctx context.Context) (*suite, error) {
	headDir := filepath.Join(c.tmpDir, "head")
	if err := c.repo.copyWorkingTree(ctx, headDir); err != nil {
		return nil, err
	}
	modFile := filepath.Join(c.tmpDir, "head.mod")
	if err := createModFile(ctx, modFile, c.benchDir, headDir); err != nil {
		return nil, err
	}
	// A calibration links the head with other layouts than the base, as a change of the code moves the functions.
	seedOffset := 0
	if c.opt.calibrate {
		seedOffset = c.opt.config.Layouts
	}
	return buildSuite(ctx, c.benchDir, modFile, c.tmpDir, "head", c.opt.config.Layouts, seedOffset)
}

// baseSuite builds the benchmarks for the base commit on the first call, as headSuite builds the ones of the
// working tree.
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

	modFile := filepath.Join(c.tmpDir, "base.mod")
	if err := createModFile(ctx, modFile, c.benchDir, baseDir); err != nil {
		return nil, err
	}
	base, err := buildSuite(ctx, c.benchDir, modFile, c.tmpDir, "base", c.opt.config.Layouts, 0)
	if err != nil {
		fmt.Printf("benchmarks of the working tree can't be built with base %s: use the benchmarks of the base\n", c.baseCommit)
		benchDir := filepath.Join(baseDir, c.opt.config.Dir)
		if err := createModFile(ctx, modFile, benchDir, baseDir); err != nil {
			return nil, err
		}
		base, err = buildSuite(ctx, benchDir, modFile, c.tmpDir, "base", c.opt.config.Layouts, 0)
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

// printProgress prints the results of an attempt of a benchmark function: the fastest layout and the median of
// the layouts of each side.
func printProgress(base, head layoutMeasurement) {
	for _, name := range head.names() {
		headNs, headMedianNs, _ := head.stats(name)
		baseNs, baseMedianNs, exists := base.stats(name)
		if !exists {
			fmt.Printf("%s\tbase: -\thead: %.2f ns/op ( median %.2f )\n", name, headNs, headMedianNs)
			continue
		}
		fmt.Printf("%s\tbase: %.2f ns/op ( median %.2f )\thead: %.2f ns/op ( median %.2f )\t%s ( median %s )\n",
			name, baseNs, baseMedianNs, headNs, headMedianNs, formatDelta(baseNs, headNs), formatDelta(baseMedianNs, headMedianNs))
	}
}

// measureFunc measures the benchmark function fn of the base and of the working tree.
// The rounds of both sides are interleaved so that they run under the same load,
// and each round uses the next layout so that every layout of each side is measured.
// If base is nil, only the working tree is measured.
func (c *checker) measureFunc(ctx context.Context, fn string, base, head *suite, baseFirst bool) (layoutMeasurement, layoutMeasurement, error) {
	type side struct {
		suite  *suite
		result layoutMeasurement
	}
	baseSide := &side{suite: base, result: newLayoutMeasurement(c.opt.config.Layouts)}
	headSide := &side{suite: head, result: newLayoutMeasurement(c.opt.config.Layouts)}
	sides := []*side{baseSide, headSide}
	if !baseFirst {
		sides = []*side{headSide, baseSide}
	}
	for round := 0; round < c.opt.config.Rounds; round++ {
		for _, s := range sides {
			if s.suite == nil {
				continue
			}
			layout := round % c.opt.config.Layouts
			result, err := s.suite.run(ctx, layout, fn, c.opt.config.BenchTime)
			if err != nil {
				return nil, nil, err
			}
			s.result[layout].mergeFastest(result)
		}
	}
	return baseSide.result, headSide.result, nil
}

// measure runs the benchmarks until all of them reach the result of the base or the attempts are exhausted.
func (c *checker) measure(ctx context.Context, out *outcome) (*measured, error) {
	fmt.Println("build benchmarks for working tree")
	head, err := c.headSuite(ctx)
	if err != nil {
		return nil, err
	}
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

	funcs, err := head.funcs(ctx, c.opt.config.Bench)
	if err != nil {
		return nil, err
	}
	if len(funcs) == 0 {
		return nil, fmt.Errorf("no benchmark matched %q in %s", c.opt.config.Bench, c.benchDir)
	}
	funcs = shardFuncs(groupFuncs(funcs, c.opt.config.Group), c.opt.config.Shard, c.opt.config.Shards)
	if len(funcs) == 0 {
		return nil, fmt.Errorf("no benchmark of the group %q in the shard %d/%d", c.opt.config.Group, c.opt.config.Shard, c.opt.config.Shards)
	}

	cmp := newComparison(c.opt.tolerance, c.opt.singleTolerance)
	for attempt := 1; attempt <= c.opt.attempts && len(funcs) != 0; attempt++ {
		fmt.Printf("measure %d benchmark functions ( attempt %d/%d )\n", len(funcs), attempt, c.opt.attempts)
		// the cached results are valid only for the first attempt:
		// a retry exists to get rid of the noise, so both sides are measured again at the same time.
		useCache := attempt == 1 && cached != nil
		var measured layoutMeasurement
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
				measured = measured.mergeFastest(baseResult)
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
	return &measured{results: cmp.sortedResults(), meanDeltaPercent: cmp.meanDeltaPercent()}, nil
}

// measured is the result of the measurement of all the attempts.
type measured struct {
	results          []benchResult
	meanDeltaPercent float64
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
	// A calibration compares HEAD with itself: it has no base branch to find.
	base := head
	if !c.opt.calibrate {
		base, err = c.repo.baseCommit(ctx, c.opt.baseRef)
		if err != nil {
			return nil, err
		}
	}
	dirty, err := c.repo.dirty(ctx)
	if err != nil {
		return nil, err
	}
	if head == base && !dirty && !c.opt.calibrate {
		fmt.Printf("HEAD is the base commit %s: nothing to compare\n", base)
		return &outcome{}, nil
	}

	key := verdictKey{Config: c.opt.config, Attempts: c.opt.attempts, Tolerance: c.opt.tolerance, SingleTolerance: c.opt.singleTolerance}
	if !dirty && !c.opt.calibrate {
		cached, err := c.cache.loadVerdict(head, base, key)
		if err != nil {
			return nil, err
		}
		if cached != nil {
			fmt.Printf("use cached verdict of %s against base %s\n", head, base)
			return &outcome{verdict: cached, cachedVerdict: true}, nil
		}
	}

	tmpDir, err := makeTempDir()
	if err != nil {
		return nil, err
	}
	c.cleanups = append(c.cleanups, func() { os.RemoveAll(tmpDir) })
	defer c.cleanup()
	c.tmpDir = tmpDir
	c.benchDir = filepath.Join(tmpDir, "head", c.opt.config.Dir)
	c.baseCommit = base

	out := &outcome{}
	result, err := c.measure(ctx, out)
	if err != nil {
		return nil, err
	}
	out.verdict = &verdict{Head: head, Base: base, Key: key, MeanDeltaPercent: result.meanDeltaPercent, Benchmarks: result.results}
	if !dirty && !c.opt.calibrate {
		if err := c.cache.storeVerdict(out.verdict); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// makeTempDir creates a temporary directory in os.TempDir whose name has the same length on every run: a cached
// result of the base was measured with the binaries an earlier run built in its own temporary directory, which must
// be built as the binaries of the head are ( see headSuite ). The result is cached for the same os.TempDir only
// ( see machine ).
func makeTempDir() (string, error) {
	for {
		dir := filepath.Join(os.TempDir(), fmt.Sprintf("benchcheck-%010d", rand.Uint32()))
		err := os.Mkdir(dir, 0o700)
		if err == nil {
			return dir, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", err
		}
	}
}

func (c *checker) run(ctx context.Context) error {
	out, err := c.check(ctx)
	if err != nil {
		return err
	}
	if out.verdict == nil {
		return nil
	}
	if c.opt.calibrate {
		reportCalibration(out.verdict)
		return nil
	}
	return report(out.verdict)
}

// reportCalibration prints the results of identical code and how much they differ, by the fastest layouts, by the
// medians and by both, which a single benchmark has to be beyond to fail ( see comparison.judge ): the mean, the
// 90th percentile and the largest of the differences, whichever side is slower.
func reportCalibration(v *verdict) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "benchmark\tbase ns/op\thead ns/op\tdelta\tbase median\thead median\tdelta of medians")
	var fastest, median, both []float64
	for _, result := range v.Benchmarks {
		if result.Status != statusOK && result.Status != statusRegressed {
			continue
		}
		fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%s\t%.2f\t%.2f\t%s\n", result.Name,
			result.BaseNs, result.HeadNs, formatDelta(result.BaseNs, result.HeadNs),
			result.BaseMedianNs, result.HeadMedianNs, formatDelta(result.BaseMedianNs, result.HeadMedianNs))
		if !result.comparable() {
			continue
		}
		d, m := result.deltaPercent(), result.medianDeltaPercent()
		fastest = append(fastest, math.Abs(d))
		median = append(median, math.Abs(m))
		// both are beyond a difference only if they are slower on the same side.
		var b float64
		if d*m > 0 {
			b = math.Min(math.Abs(d), math.Abs(m))
		}
		both = append(both, b)
	}
	_ = w.Flush()
	if len(fastest) == 0 {
		return
	}
	fmt.Printf("calibration: %d benchmarks, mean of the deltas %+.2f%%\n", len(fastest), v.MeanDeltaPercent)
	for _, s := range []struct {
		name   string
		deltas []float64
	}{{"fastest layout", fastest}, {"median", median}, {"both", both}} {
		sort.Float64s(s.deltas)
		var sum float64
		for _, d := range s.deltas {
			sum += d
		}
		n := len(s.deltas)
		fmt.Printf("  %s: mean |delta| %.2f%%, p90 %.2f%%, max %.2f%%\n", s.name, sum/float64(n), s.deltas[n*9/10], s.deltas[n-1])
	}
}

func report(v *verdict) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// the fastest layout of each side, and the median of the layouts.
	fmt.Fprintln(w, "benchmark\tbase ns/op\thead ns/op\tdelta\tbase median\thead median\tdelta of medians\tattempts\tstatus")
	regressed := []string{}
	for _, result := range v.Benchmarks {
		switch result.Status {
		case statusOK, statusRegressed:
			fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%s\t%.2f\t%.2f\t%s\t%d\t%s\n", result.Name,
				result.BaseNs, result.HeadNs, formatDelta(result.BaseNs, result.HeadNs),
				result.BaseMedianNs, result.HeadMedianNs, formatDelta(result.BaseMedianNs, result.HeadMedianNs),
				result.Attempts, result.Status)
		case statusNew:
			fmt.Fprintf(w, "%s\t-\t%.2f\t-\t-\t%.2f\t-\t%d\t%s\n", result.Name, result.HeadNs, result.HeadMedianNs, result.Attempts, result.Status)
		}
		if result.Status == statusRegressed {
			regressed = append(regressed, result.Name)
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Printf("mean: %+.2f%% ( tolerance %v%% )\n", v.MeanDeltaPercent, v.Key.Tolerance)
	meanDegraded := v.MeanDeltaPercent > v.Key.Tolerance
	if !meanDegraded && len(regressed) == 0 {
		fmt.Printf("OK: not degraded compared with base %s\n", v.Base)
		return nil
	}
	if meanDegraded {
		fmt.Printf("FAIL: the mean of the benchmarks is slower than base %s beyond the tolerance\n", v.Base)
	}
	if len(regressed) != 0 {
		fmt.Printf("FAIL: %d benchmarks are slower than base %s beyond the tolerance of a single benchmark ( %v%% ), by both the fastest layout and the median\n", len(regressed), v.Base, v.Key.SingleTolerance)
		for _, name := range regressed {
			fmt.Printf("  %s\n", name)
		}
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

// shardFuncs returns the functions of the shard-th of shards parts. The functions are dealt to the parts in
// the order of their names, one by one: the benchmarks of a kind, whose names are next to each other, are in
// every part, so that the parts take about as long as each other.
func shardFuncs(funcs []string, shard, shards int) []string {
	if shards <= 1 {
		return funcs
	}
	sorted := append([]string(nil), funcs...)
	sort.Strings(sorted)
	var part []string
	for i, fn := range sorted {
		if i%shards == shard {
			part = append(part, fn)
		}
	}
	return part
}

// groupFuncs returns the functions of the group, or all of them for no group.
func groupFuncs(funcs []string, group string) []string {
	if group == "" {
		return funcs
	}
	var part []string
	for _, fn := range funcs {
		if benchmarkGroup(fn) == group {
			part = append(part, fn)
		}
	}
	return part
}

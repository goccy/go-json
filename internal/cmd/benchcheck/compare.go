package main

import (
	"math"
	"sort"
)

type status string

const (
	// statusOK means the benchmark is not slower than the base beyond the tolerance of a single benchmark.
	statusOK status = "ok"
	// statusRegressed means the benchmark is slower than the base beyond the tolerance of a single benchmark.
	statusRegressed status = "regressed"
	// statusNew means the benchmark doesn't exist in the base, so there is nothing to compare.
	statusNew status = "new"
)

type benchResult struct {
	Name string `json:"name"`
	// Func is the top-level benchmark function which the benchmark belongs to.
	Func   string `json:"func"`
	Status status `json:"status"`
	// BaseNs and HeadNs are the fastest results of all the attempts.
	// The noise of a measurement only makes it slower, so the fastest one is the closest to the real cost.
	// Each side takes its own fastest result: if a pair measured in the same attempt were compared,
	// a base which happened to be slow in an attempt would let a degraded head pass.
	BaseNs   float64 `json:"baseNs"`
	HeadNs   float64 `json:"headNs"`
	Attempts int     `json:"attempts"`
	// baseSeen is whether the base has reported the benchmark.
	// It is not the same as BaseNs != 0: the testing package reports zero ns/op as well.
	baseSeen bool
}

// comparable reports whether the result has both sides to compare.
// ns/op is zero if the testing package omitted it, and the ratio to zero can't be represented.
func (r *benchResult) comparable() bool {
	return r.baseSeen && r.BaseNs > 0 && r.HeadNs > 0
}

// deltaPercent returns how much slower the head is than the base, in percent.
func (r *benchResult) deltaPercent() float64 {
	return (r.HeadNs/r.BaseNs - 1) * 100
}

// comparison collects the results of the attempts and decides whether the head is degraded.
//
// A single benchmark is too noisy to be judged with a small tolerance: on a shared machine the same code
// differs by several percent between two builds. So the head is judged by the mean of all the benchmarks,
// whose noise is far smaller, with meanTolerance. A single benchmark is judged only with singleTolerance,
// which is beyond the noise, because the mean can't notice the degradation of only a few benchmarks.
type comparison struct {
	meanTolerance   float64
	singleTolerance float64
	results         map[string]*benchResult
}

func newComparison(meanTolerance, singleTolerance float64) *comparison {
	return &comparison{
		meanTolerance:   meanTolerance,
		singleTolerance: singleTolerance,
		results:         map[string]*benchResult{},
	}
}

// add records one attempt of the benchmark function fn:
// base and head are the results of the base and of the working tree measured for the attempt.
func (c *comparison) add(fn string, base, head measurement) {
	for name, headNs := range head {
		result, exists := c.results[name]
		if !exists {
			result = &benchResult{Name: name, Func: fn, HeadNs: headNs}
			c.results[name] = result
		}
		result.Attempts++
		if headNs < result.HeadNs {
			result.HeadNs = headNs
		}
		if baseNs, hasBase := base[name]; hasBase {
			if !result.baseSeen || baseNs < result.BaseNs {
				result.BaseNs = baseNs
			}
			result.baseSeen = true
		}
		c.judge(result)
	}
}

func (c *comparison) judge(result *benchResult) {
	switch {
	case !result.baseSeen:
		result.Status = statusNew
	case result.BaseNs == 0 && result.HeadNs > 0:
		// the base was too fast to be measured and the head is not.
		// The ratio can't be represented, so it is not a part of the mean either.
		result.Status = statusRegressed
	case result.comparable() && result.deltaPercent() > c.singleTolerance:
		result.Status = statusRegressed
	default:
		result.Status = statusOK
	}
}

// meanDeltaPercent returns how much slower the head is than the base on average, in percent.
// It is the geometric mean of the ratios, which is the mean of how many times slower each benchmark is.
func (c *comparison) meanDeltaPercent() float64 {
	var (
		sum   float64
		count int
	)
	for _, result := range c.results {
		if !result.comparable() {
			continue
		}
		sum += math.Log(result.HeadNs / result.BaseNs)
		count++
	}
	if count == 0 {
		return 0
	}
	return (math.Exp(sum/float64(count)) - 1) * 100
}

// degraded reports whether the head is degraded: the mean is beyond its tolerance,
// or a benchmark is beyond the tolerance of a single benchmark.
func (c *comparison) degraded() bool {
	if c.meanDeltaPercent() > c.meanTolerance {
		return true
	}
	for _, result := range c.results {
		if result.Status == statusRegressed {
			return true
		}
	}
	return false
}

// pendingFuncs returns the benchmark functions to measure again: nothing if the head is not degraded,
// otherwise the ones which have a benchmark slower than the tolerance of the mean, which are what makes
// the mean slow. Measuring them again gets rid of the noise, because the fastest result is kept.
func (c *comparison) pendingFuncs() []string {
	if !c.degraded() {
		return nil
	}
	seen := map[string]struct{}{}
	funcs := []string{}
	for _, result := range c.results {
		if !result.comparable() || result.deltaPercent() <= c.meanTolerance {
			continue
		}
		if _, exists := seen[result.Func]; exists {
			continue
		}
		seen[result.Func] = struct{}{}
		funcs = append(funcs, result.Func)
	}
	sort.Strings(funcs)
	return funcs
}

func (c *comparison) sortedResults() []benchResult {
	results := make([]benchResult, 0, len(c.results))
	for _, result := range c.results {
		results = append(results, *result)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	return results
}

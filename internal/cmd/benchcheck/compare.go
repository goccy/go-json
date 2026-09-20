package main

import (
	"sort"
)

type status string

const (
	// statusOK means the benchmark reached the result of the base within the allowed attempts.
	statusOK status = "ok"
	// statusRegressed means the benchmark never reached the result of the base.
	statusRegressed status = "regressed"
	// statusNew means the benchmark doesn't exist in the base, so there is nothing to compare.
	statusNew status = "new"
)

type benchResult struct {
	Name string `json:"name"`
	// Func is the top-level benchmark function which the benchmark belongs to.
	Func   string `json:"func"`
	Status status `json:"status"`
	// BaseNs and HeadNs are the pair of the measurements in which HeadNs was the best relative to BaseNs.
	BaseNs   float64 `json:"baseNs"`
	HeadNs   float64 `json:"headNs"`
	Attempts int     `json:"attempts"`
}

// comparison tracks, over multiple attempts, whether each benchmark reached the result of the base.
type comparison struct {
	tolerance float64
	results   map[string]*benchResult
	// compared is the set of the benchmarks whose result holds a measured pair.
	compared map[string]struct{}
}

func newComparison(tolerance float64) *comparison {
	return &comparison{
		tolerance: tolerance,
		results:   map[string]*benchResult{},
		compared:  map[string]struct{}{},
	}
}

// reached reports whether headNs is not slower than baseNs.
// tolerance is the measurement noise ( in percent ) which is not treated as a degradation.
func (c *comparison) reached(baseNs, headNs float64) bool {
	return headNs <= baseNs*(1+c.tolerance/100)
}

// closerToBase reports whether the pair ( baseNs, headNs ) is better than ( otherBaseNs, otherHeadNs ),
// that is, whether headNs/baseNs is less than otherHeadNs/otherBaseNs.
// The ratios are compared without division, because ns/op can be zero.
func closerToBase(baseNs, headNs, otherBaseNs, otherHeadNs float64) bool {
	lhs, rhs := headNs*otherBaseNs, otherHeadNs*baseNs
	if lhs != rhs {
		return lhs < rhs
	}
	return headNs < otherHeadNs
}

// add records one attempt of the benchmark function fn:
// base and head are the results of the base and of the working tree measured for the attempt.
// The benchmarks which have already reached the base are not updated.
func (c *comparison) add(fn string, base, head measurement) {
	for name, headNs := range head {
		result, exists := c.results[name]
		if !exists {
			result = &benchResult{Name: name, Func: fn, Status: statusRegressed}
			c.results[name] = result
		}
		if result.Status != statusRegressed {
			continue
		}
		result.Attempts++
		baseNs, hasBase := base[name]
		if !hasBase {
			if !exists {
				result.Status = statusNew
				result.HeadNs = headNs
			}
			continue
		}
		_, compared := c.compared[name]
		if !compared || closerToBase(baseNs, headNs, result.BaseNs, result.HeadNs) {
			result.BaseNs = baseNs
			result.HeadNs = headNs
			c.compared[name] = struct{}{}
		}
		if c.reached(baseNs, headNs) {
			result.Status = statusOK
		}
	}
}

// pendingFuncs returns the benchmark functions
// which have a benchmark that has not reached the result of the base yet.
func (c *comparison) pendingFuncs() []string {
	seen := map[string]struct{}{}
	funcs := []string{}
	for _, result := range c.results {
		if result.Status != statusRegressed {
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

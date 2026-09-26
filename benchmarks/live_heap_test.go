package benchmark

import (
	"os"
	"strconv"
)

// liveHeap is the live heap of the program which the benchmarks run in, when BENCH_LIVE_HEAP_MB sets its size in
// megabytes. A program which decodes JSON has a live heap of its own, which sets the goal of the GC for every
// library alike, as GOGC makes the goal twice the live heap. Without it, the goal is set by what a library itself
// keeps alive, as its caches and pools: a library which keeps more is collected less often in a benchmark, which
// a real program doesn't see. The benchmarks are compared with and without it.
var liveHeap []byte

func init() {
	if mb, err := strconv.Atoi(os.Getenv("BENCH_LIVE_HEAP_MB")); err == nil && mb > 0 {
		liveHeap = make([]byte, mb<<20)
	}
}

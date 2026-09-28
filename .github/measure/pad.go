package encoder

// measurePad moves the code of the package after it, to measure how much the layout alone moves the benchmarks.
//
//go:noinline
func measurePad(n int) int { return n*3 + 1 }

var measurePadSink = measurePad(1)

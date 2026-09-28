package decoder

// measurePadOf adds two dictionaries to the read-only data, which moves the itabs after them by 32 bytes, to
// measure how much the place of the itabs alone moves the benchmarks.
//
//go:noinline
func measurePadOf[T any](n int) any { return make([]T, n) }

type measurePad1 struct{ n [1]int }
type measurePad2 struct{ n [2]int }

var measurePadSink1 = measurePadOf[measurePad1](1)
var measurePadSink2 = measurePadOf[measurePad2](1)

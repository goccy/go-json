package encoder

//go:noinline
func alignPad(n int) int { return n + 1 }

var alignSink = alignPad(1)

//go:build !race

package json_test

// raceEnabled is whether the race detector is enabled, with which sync.Pool drops the values at random.
const raceEnabled = false

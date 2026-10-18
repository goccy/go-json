package json_test

import (
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// TestUnmarshalDuplicateNamesDeepIsLinear checks that a name which an object deep in an input has twice, which
// AllowDuplicateNames allows, makes the input be decoded again once, not once for each level around the object:
// the time grows with the depth, not with its square.
func TestUnmarshalDuplicateNamesDeepIsLinear(t *testing.T) {
	for _, c := range []struct {
		name        string
		open, close string
	}{
		{"objects", `{"a":`, `}`},
		{"arrays", `[`, `]`},
		{"objects in arrays", `[{"a":`, `}]`},
	} {
		t.Run(c.name, func(t *testing.T) {
			// the fastest of a few decodings, which the noise of the machine only makes slower
			fastest := func(depth int) time.Duration {
				in := []byte(strings.Repeat(c.open, depth) + `{"x":1,"x":2}` + strings.Repeat(c.close, depth))
				best := time.Duration(1<<63 - 1)
				for range 10 {
					start := time.Now()
					var v any
					if err := json.Unmarshal(in, &v, jsontext.AllowDuplicateNames(true)); err != nil {
						t.Fatal(err)
					}
					best = min(best, time.Since(start))
				}
				return best
			}
			// sixteen times as deep: sixteen to thirty times the time if it is linear, 256 times if it is quadratic
			small, large := fastest(250), fastest(4000)
			if large > small*100 {
				t.Fatalf("%v at the depth 4000, %v at the depth 250", large, small)
			}
		})
	}
}

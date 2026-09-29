package benchmark

import (
	"fmt"
	"testing"

	gojson "github.com/goccy/go-json"
)

// TestZZCodeSetCounts prints the lookups of the code sets for an encode of each report payload by go-json, and
// how many of them miss the recent code sets of the context and go to the shared table.
func TestZZCodeSetCounts(t *testing.T) {
	for _, c := range reportConfigs {
		if c.Library != "goccy/go-json" || c.marshal == nil || (c.Category != "std" && c.Category != "fast" && c.Category != "fastest") {
			continue
		}
		for _, p := range reportPayloads() {
			if _, err := c.marshal(p.value); err != nil {
				t.Fatal(err)
			}
			gojson.ResetCodeSetCounts()
			const n = 100
			for i := 0; i < n; i++ {
				if _, err := c.marshal(p.value); err != nil {
					t.Fatal(err)
				}
			}
			lookups, misses, shared := gojson.CodeSetCounts()
			fmt.Printf("count %s %s lookups %.1f misses %.1f shared %.1f\n", c.ID, p.ID, float64(lookups)/n, float64(misses)/n, float64(shared)/n)
		}
	}
}

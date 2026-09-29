package benchmark

import (
	"os"
	"strings"
	"testing"
)

// BenchmarkZZEncode encodes the report payload ZZ_PAYLOAD by the report configuration ZZ_CONFIG
// ( e.g. "fastest/go-json" ), so that a profile is of the encoder alone.
func BenchmarkZZEncode(b *testing.B) {
	var c *reportConfig
	for _, x := range reportConfigs {
		if x.ID == os.Getenv("ZZ_CONFIG") {
			c = x
		}
	}
	var p *reportPayload
	for _, x := range reportPayloads() {
		if x.ID == os.Getenv("ZZ_PAYLOAD") {
			p = x
		}
	}
	if c == nil || p == nil {
		b.Fatal("unknown configuration or payload")
	}
	pretouchSonic()
	if _, err := c.marshal(p.value); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.marshal(p.value); err != nil {
			b.Fatal(err)
		}
	}
}

// TestZZKeys prints the number of the keys and of the bytes of the encode of ZZ_PAYLOAD by go-json.
func TestZZKeys(t *testing.T) {
	for _, x := range reportPayloads() {
		if x.ID != os.Getenv("ZZ_PAYLOAD") {
			continue
		}
		for _, c := range reportConfigs {
			if c.ID == "fastest/go-json" {
				out, err := c.marshal(x.value)
				if err != nil {
					t.Fatal(err)
				}
				var long int
				for _, part := range strings.Split(string(out), "\":")[:strings.Count(string(out), "\":")] {
					if k := strings.LastIndexByte(part, '"'); k >= 0 && len(part)-k+2 > 16 {
						long++
					}
				}
				t.Logf("keys: %d, over 16 bytes: %d, bytes: %d", strings.Count(string(out), "\":"), long, len(out))
			}
		}
	}
}

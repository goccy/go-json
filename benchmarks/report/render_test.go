package report

import (
	"strings"
	"testing"
	"time"
)

func testRun(arch string) *Run {
	return &Run{
		GeneratedAt: time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC),
		GoVersion:   "go1.27.1", GOOS: "linux", GOARCH: arch, CPU: "test CPU",
		Commit: "0123456789abcdef", Repository: "goccy/go-json", BenchTime: "100ms", Rounds: 3,
		Libraries:  []Library{{Name: "encoding/json", Module: "std", Version: "go1.27.1"}, {Name: "goccy/go-json", Module: "github.com/goccy/go-json", Version: "0123456789abcdef"}},
		Categories: []Category{{ID: "std", Title: "Same behavior as encoding/json", Baseline: "encoding/json"}},
		Conditions: []Condition{{ID: "live-heap", Title: "With a live heap"}},
		Payloads:   []Payload{{ID: "small", Title: "Small struct", Bytes: 100}, {ID: "large", Title: "Large struct", Bytes: 1000}},
		Configs: []Config{
			{ID: "encoding/json", Library: "encoding/json", Category: "std", Title: "encoding/json"},
			{ID: "go-json", Library: "goccy/go-json", Category: "std", Title: "goccy/go-json"},
			{ID: "other", Library: "other", Category: "std", Title: "other <lib>"},
		},
		Probes: []ProbeResult{
			{Config: "go-json", Probe: "p", Description: "probe", Required: true, Passed: true},
			{Config: "other", Probe: "p", Description: "probe", Required: true, Passed: false, Detail: "got <x>"},
		},
		Results: []Result{
			{Condition: "live-heap", Op: OpDecode, Payload: "small", Config: "encoding/json", NsPerOp: 400},
			{Condition: "live-heap", Op: OpDecode, Payload: "small", Config: "go-json", NsPerOp: 100},
			{Condition: "live-heap", Op: OpDecode, Payload: "large", Config: "encoding/json", NsPerOp: 4000},
			{Condition: "live-heap", Op: OpDecode, Payload: "large", Config: "go-json", NsPerOp: 2000},
			{Condition: "live-heap", Op: OpDecode, Payload: "small", Config: "other", NsPerOp: 50},
		},
	}
}

func TestRender(t *testing.T) {
	site := &Site{Runs: []*Run{testRun("arm64"), testRun("amd64")}, Files: []string{"results-arm64.json", "results-amd64.json"}}
	page, err := site.HTML()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"arch":"amd64"`, // the results are in the page as JSON
		`"file":"results-amd64.json"`,
		`other \u003clib\u003e`, // escaped in the script
		`"differs":{"other":[`,  // the configuration which fails a probe
		"<h1>Go JSON Benchmarks</h1>",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the page has no %q", want)
		}
	}
	// amd64 is the first run
	if strings.Index(page, `"arch":"amd64"`) > strings.Index(page, `"arch":"arm64"`) {
		t.Error("the runs are not ordered by architecture")
	}
	svg := site.Summary()
	if !strings.HasPrefix(svg, "<svg") || !strings.Contains(svg, "2.83x") || !strings.Contains(svg, "other &lt;lib&gt; †") {
		t.Errorf("unexpected summary: %s", svg)
	}
}

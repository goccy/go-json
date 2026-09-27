// Package report holds the results of the benchmark report of go-json, which compares the JSON libraries of Go
// doing the same work, and renders them as an HTML page and an SVG summary.
//
// A run is measured on one machine ( see report_test.go of the benchmarks ): every library of a comparison is
// measured there, by turns, so that the libraries are compared on the same CPU. The runs of the machines are
// rendered together.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Run is what a run of the report measured on one machine.
type Run struct {
	GeneratedAt time.Time `json:"generatedAt"`
	GoVersion   string    `json:"goVersion"`
	GOOS        string    `json:"goos"`
	GOARCH      string    `json:"goarch"`
	CPU         string    `json:"cpu"`
	// CPUFeatures are the extensions of the instruction set of the CPU which libraries choose their code by, as
	// AVX2 and AVX-512: a library may be several times faster on a CPU with one than on one without.
	CPUFeatures []string `json:"cpuFeatures"`
	// Commit, Repository and RunURL tell where the run comes from: the commit of go-json measured, the repository
	// it is in and the workflow run which measured it.
	Commit     string `json:"commit"`
	Repository string `json:"repository"`
	RunURL     string `json:"runURL"`
	BenchTime  string `json:"benchTime"`
	Rounds     int    `json:"rounds"`

	Libraries  []Library     `json:"libraries"`
	Categories []Category    `json:"categories"`
	Conditions []Condition   `json:"conditions"`
	Payloads   []Payload     `json:"payloads"`
	Configs    []Config      `json:"configs"`
	Probes     []ProbeResult `json:"probes"`
	Results    []Result      `json:"results"`
	// Differences are the payloads on which the result of a configuration is not the one of its category: it is
	// measured, and shown with the difference.
	Differences []Exclusion `json:"differences"`
	// Exclusions are the payloads on which an operation of a configuration fails, which is not measured.
	Exclusions []Exclusion `json:"exclusions"`
}

// Library is a library measured, with the version the run was built with.
type Library struct {
	Name    string `json:"name"`
	Module  string `json:"module"`
	Version string `json:"version"`
}

// Category is a set of configurations which behave the same: only the configurations of a category are compared
// with each other.
type Category struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Baseline is the configuration which the others of the category are compared with.
	Baseline string `json:"baseline"`
}

// Condition is a condition of the process which the benchmarks are measured in.
type Condition struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Payload is an input of the benchmarks: it is decoded into its Go type, and the value is encoded.
type Payload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Bytes       int    `json:"bytes"`
}

// Config is a library called in one way: the functions and the options it is called with.
type Config struct {
	ID       string `json:"id"`
	Library  string `json:"library"`
	Category string `json:"category"`
	// Title is the name of the configuration on the page, and Setting how the library is called.
	Title   string `json:"title"`
	Setting string `json:"setting"`
}

// ProbeResult is whether a configuration behaves as its category requires on a probe: a small input which tells a
// behavior apart. A configuration which fails a required probe is still measured, and shown apart from the others
// of its category, with what differs.
type ProbeResult struct {
	Config      string `json:"config"`
	Probe       string `json:"probe"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Passed      bool   `json:"passed"`
	Detail      string `json:"detail,omitempty"`
}

// Result is the measurement of a configuration on a payload.
type Result struct {
	Condition   string  `json:"condition"`
	Op          string  `json:"op"`
	Payload     string  `json:"payload"`
	Config      string  `json:"config"`
	NsPerOp     float64 `json:"nsPerOp"`
	BytesPerOp  int64   `json:"bytesPerOp"`
	AllocsPerOp int64   `json:"allocsPerOp"`
}

// Exclusion is a configuration on a payload, and why it is noted: the operation fails ( Run.Exclusions ), or its
// result is not the one of its category ( Run.Differences ).
type Exclusion struct {
	Op      string `json:"op"`
	Payload string `json:"payload"`
	Config  string `json:"config"`
	Reason  string `json:"reason"`
}

// Operations are the operations measured.
const (
	OpEncode = "encode"
	OpDecode = "decode"
)

// ReadRun reads a run written by WriteRun.
func ReadRun(path string) (*Run, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r Run
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &r, nil
}

// WriteRun writes a run as JSON.
func WriteRun(path string, r *Run) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// result returns the result of a configuration on a payload, or nil.
func (r *Run) result(cond, op, payload, config string) *Result {
	for i := range r.Results {
		x := &r.Results[i]
		if x.Condition == cond && x.Op == op && x.Payload == payload && x.Config == config {
			return x
		}
	}
	return nil
}

// exclusion returns why a configuration is not measured on a payload, or "".
func (r *Run) exclusion(op, payload, config string) string {
	return find(r.Exclusions, op, payload, config)
}

// difference returns how the result of a configuration on a payload differs from the one of its category, or "".
func (r *Run) difference(op, payload, config string) string {
	return find(r.Differences, op, payload, config)
}

func find(es []Exclusion, op, payload, config string) string {
	for _, e := range es {
		if e.Op == op && e.Payload == payload && e.Config == config {
			return e.Reason
		}
	}
	return ""
}

// configsOf returns the configurations of a category, in their order.
func (r *Run) configsOf(category string) []Config {
	var cs []Config
	for _, c := range r.Configs {
		if c.Category == category {
			cs = append(cs, c)
		}
	}
	return cs
}

// library returns the library of the name.
func (r *Run) library(name string) Library {
	for _, l := range r.Libraries {
		if l.Name == name {
			return l
		}
	}
	return Library{Name: name}
}

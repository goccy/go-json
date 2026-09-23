package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// benchConfig is the set of options which decides what is measured.
type benchConfig struct {
	Dir       string `json:"dir"`
	Bench     string `json:"bench"`
	BenchTime string `json:"benchTime"`
	Rounds    int    `json:"rounds"`
	Layouts   int    `json:"layouts"`
}

// machine identifies the environment in which a measurement is comparable with another one.
// ns/op measured on a different machine must not be used as the baseline.
type machine struct {
	Hostname  string `json:"hostname"`
	GOOS      string `json:"goos"`
	GOARCH    string `json:"goarch"`
	NumCPU    int    `json:"numCpu"`
	GoVersion string `json:"goVersion"`
}

type baselineKey struct {
	Config  benchConfig `json:"config"`
	Machine machine     `json:"machine"`
	// BenchHash is the hash of the benchmark directory of the working tree,
	// because the benchmarks of the working tree are run against the base commit.
	BenchHash string `json:"benchHash"`
}

// verdictKey doesn't contain the machine because a verdict is relative:
// both sides were measured on the same machine, whichever it was.
type verdictKey struct {
	Config    benchConfig `json:"config"`
	Attempts  int         `json:"attempts"`
	Tolerance float64     `json:"tolerance"`
	// SingleTolerance is the tolerance of a single benchmark, and Tolerance is the one of the mean.
	SingleTolerance float64 `json:"singleTolerance"`
}

type baseline struct {
	Commit  string      `json:"commit"`
	Key     baselineKey `json:"key"`
	Results measurement `json:"results"`
}

type verdict struct {
	Head string     `json:"head"`
	Base string     `json:"base"`
	Key  verdictKey `json:"key"`
	// MeanDeltaPercent is how much slower the head is than the base on average, in percent.
	MeanDeltaPercent float64       `json:"meanDeltaPercent"`
	Benchmarks       []benchResult `json:"benchmarks"`
}

type cache struct {
	dir string
	// readable is false when cached data must be ignored and measured again.
	readable bool
}

func keyHash(key any) (string, error) {
	b, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8]), nil
}

func (c *cache) baselinePath(commit string, key baselineKey) (string, error) {
	hash, err := keyHash(key)
	if err != nil {
		return "", err
	}
	return filepath.Join(c.dir, "baseline", fmt.Sprintf("%s-%s.json", commit, hash)), nil
}

func (c *cache) verdictPath(head, base string, key verdictKey) (string, error) {
	hash, err := keyHash(key)
	if err != nil {
		return "", err
	}
	return filepath.Join(c.dir, "verdict", fmt.Sprintf("%s-%s-%s.json", head, base, hash)), nil
}

// load reads the cached value into v. It returns false if the value is not cached.
func (c *cache) load(path string, v any) (bool, error) {
	if !c.readable {
		return false, nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return false, fmt.Errorf("failed to decode %s: %w", path, err)
	}
	return true, nil
}

func (c *cache) store(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// write to a temporary file first so that an interrupted run never leaves a broken cache.
	tmp, err := os.CreateTemp(filepath.Dir(path), "tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func (c *cache) loadBaseline(commit string, key baselineKey) (*baseline, error) {
	path, err := c.baselinePath(commit, key)
	if err != nil {
		return nil, err
	}
	var v baseline
	found, err := c.load(path, &v)
	if err != nil || !found {
		return nil, err
	}
	return &v, nil
}

func (c *cache) storeBaseline(v *baseline) error {
	path, err := c.baselinePath(v.Commit, v.Key)
	if err != nil {
		return err
	}
	return c.store(path, v)
}

func (c *cache) loadVerdict(head, base string, key verdictKey) (*verdict, error) {
	path, err := c.verdictPath(head, base, key)
	if err != nil {
		return nil, err
	}
	var v verdict
	found, err := c.load(path, &v)
	if err != nil || !found {
		return nil, err
	}
	return &v, nil
}

func (c *cache) storeVerdict(v *verdict) error {
	path, err := c.verdictPath(v.Head, v.Base, v.Key)
	if err != nil {
		return err
	}
	return c.store(path, v)
}

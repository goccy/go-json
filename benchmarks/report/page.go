package report

import (
	_ "embed"
	"html/template"
	"strings"
)

// page.html is the page of the report. The results are in the page as JSON, and the charts are drawn from them in
// the browser, so that the reader chooses the machine, the condition, the operation and the metric.
//
//go:embed page.html
var pageHTML string

var pageTemplate = template.Must(template.New("page").Parse(pageHTML))

// pageData is what the page is rendered from.
type pageData struct {
	Generated    string         `json:"generated"`
	GoVersion    string         `json:"goVersion"`
	Commit       string         `json:"commit"`
	ShortCommit  string         `json:"shortCommit"`
	Repository   string         `json:"repository"`
	RunURL       string         `json:"runURL"`
	Source       string         `json:"source"`
	Attestations string         `json:"attestations"`
	Libraries    []Library      `json:"libraries"`
	Categories   []pageCategory `json:"categories"`
	Conditions   []Condition    `json:"conditions"`
	Payloads     []Payload      `json:"payloads"`
	Configs      []Config       `json:"configs"`
	Runs         []pageRun      `json:"runs"`
	Probes       []pageProbe    `json:"probes"`
}

type pageCategory struct {
	Category
	// Requires are the descriptions of the probes the category requires.
	Requires []string `json:"requires"`
}

type pageRun struct {
	Arch      string   `json:"arch"`
	CPU       string   `json:"cpu"`
	Features  []string `json:"features"`
	GOOS      string   `json:"goos"`
	File      string   `json:"file"`
	BenchTime string   `json:"benchTime"`
	Rounds    int      `json:"rounds"`
	// Differs are the required probes each configuration fails, by configuration.
	Differs map[string][]ProbeResult `json:"differs"`
	// Skips are the probes which a configuration is only informed of, of a category which requires none, and
	// which it fails: what it does differently from encoding/json.
	Skips       map[string][]ProbeResult `json:"skips"`
	Differences []Exclusion              `json:"differences"`
	Exclusions  []Exclusion              `json:"exclusions"`
	Results     []Result                 `json:"results"`
}

// pageProbe is a probe and what every configuration did on it, on the first machine.
type pageProbe struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Results     map[string]ProbeResult `json:"results"`
}

// HTML renders the page of the runs.
func (s *Site) HTML() (string, error) {
	s.sortRuns()
	d := pageData{Attestations: s.Attestations}
	if len(s.Runs) > 0 {
		r := s.Runs[0]
		d.Generated = r.GeneratedAt.Format("2006-01-02 15:04 UTC")
		d.GoVersion, d.Commit, d.Repository, d.RunURL = r.GoVersion, r.Commit, r.Repository, r.RunURL
		d.ShortCommit = r.Commit
		if len(d.ShortCommit) > 7 {
			d.ShortCommit = d.ShortCommit[:7]
		}
		if r.Repository != "" && r.Commit != "" {
			d.Source = "https://github.com/" + r.Repository + "/blob/" + r.Commit + "/benchmarks/report_test.go"
		}
		d.Libraries, d.Conditions, d.Payloads, d.Configs = r.Libraries, r.Conditions, r.Payloads, r.Configs
		for _, c := range r.Categories {
			d.Categories = append(d.Categories, pageCategory{Category: c, Requires: probesOf(r, c.ID)})
		}
		index := map[string]int{}
		for _, p := range r.Probes {
			i, ok := index[p.Probe]
			if !ok {
				i = len(d.Probes)
				index[p.Probe] = i
				d.Probes = append(d.Probes, pageProbe{ID: p.Probe, Description: p.Description, Results: map[string]ProbeResult{}})
			}
			d.Probes[i].Results[p.Config] = p
		}
	}
	for i, r := range s.Runs {
		pr := pageRun{
			Arch: r.GOARCH, CPU: r.CPU, Features: r.CPUFeatures, GOOS: r.GOOS, BenchTime: r.BenchTime, Rounds: r.Rounds,
			Differs: map[string][]ProbeResult{}, Skips: map[string][]ProbeResult{},
			Differences: r.Differences, Exclusions: r.Exclusions, Results: r.Results,
		}
		if i < len(s.Files) {
			pr.File = s.Files[i]
		}
		for _, c := range r.Configs {
			if f := failedProbes(r, c.ID); len(f) > 0 {
				pr.Differs[c.ID] = f
			}
			for _, p := range r.Probes {
				if p.Config == c.ID && !p.Required && !p.Passed {
					pr.Skips[c.ID] = append(pr.Skips[c.ID], p)
				}
			}
		}
		d.Runs = append(d.Runs, pr)
	}
	var b strings.Builder
	err := pageTemplate.Execute(&b, d)
	return b.String(), err
}

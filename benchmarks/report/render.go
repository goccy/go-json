package report

import (
	"fmt"
	"html/template"
	"math"
	"sort"
	"strings"
)

// Site is what is rendered from the runs of the machines: the page, the SVG summary, and the files of the runs,
// which the page links to.
type Site struct {
	Runs []*Run
	// Files are the names of the files of the runs, in the order of Runs, which the page links to.
	Files []string
	// Attestations is the URL of the attestations of the files, if they are attested.
	Attestations string
}

// sortRuns orders the runs by architecture, amd64 first.
func (s *Site) sortRuns() {
	idx := make([]int, len(s.Runs))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return s.Runs[idx[a]].GOARCH < s.Runs[idx[b]].GOARCH })
	runs, files := make([]*Run, len(idx)), make([]string, len(idx))
	for i, j := range idx {
		runs[i] = s.Runs[j]
		if j < len(s.Files) {
			files[i] = s.Files[j]
		}
	}
	s.Runs, s.Files = runs, files
}

// cell is the result of a configuration on a payload, as the page shows it.
type cell struct {
	Text     string
	Title    string
	Ratio    string
	Bar      float64
	Excluded bool
	Faster   bool
}

type row struct {
	Config   Config
	Baseline bool
	Cells    []cell
	Mean     string
}

type table struct {
	Op       string
	Payloads []Payload
	Rows     []row
}

type categoryView struct {
	Category Category
	Tables   []table
	// Excluded are the configurations of the category which fail a required probe, and why.
	Excluded []excludedView
	Probes   []string
}

type excludedView struct {
	Config Config
	Failed []ProbeResult
}

type conditionView struct {
	Condition  Condition
	Categories []categoryView
}

type runView struct {
	Run        *Run
	File       string
	Conditions []conditionView
	Probes     probeMatrix
}

type probeMatrix struct {
	Probes  []ProbeResult
	Configs []Config
	Passed  map[string]map[string]*ProbeResult
}

func (m probeMatrix) Get(config, probe string) *ProbeResult {
	return m.Passed[config][probe]
}

// view builds what the page shows of a run.
func view(r *Run, file string) runView {
	v := runView{Run: r, File: file}
	for _, cond := range r.Conditions {
		cv := conditionView{Condition: cond}
		for _, cat := range r.Categories {
			configs := r.configsOf(cat.ID)
			if len(configs) == 0 {
				continue
			}
			catv := categoryView{Category: cat}
			for _, p := range probesOf(r, cat.ID) {
				catv.Probes = append(catv.Probes, p)
			}
			for _, c := range configs {
				if failed := failedProbes(r, c.ID); len(failed) > 0 {
					catv.Excluded = append(catv.Excluded, excludedView{Config: c, Failed: failed})
				}
			}
			for _, op := range []string{OpEncode, OpDecode} {
				t := table{Op: op, Payloads: r.Payloads}
				for _, c := range configs {
					if len(failedProbes(r, c.ID)) > 0 || !hasOp(r, cond.ID, op, c.ID) {
						continue
					}
					t.Rows = append(t.Rows, buildRow(r, cond.ID, op, cat, c))
				}
				if len(t.Rows) > 0 {
					catv.Tables = append(catv.Tables, t)
				}
			}
			cv.Categories = append(cv.Categories, catv)
		}
		v.Conditions = append(v.Conditions, cv)
	}
	v.Probes = buildProbeMatrix(r)
	return v
}

func hasOp(r *Run, cond, op, config string) bool {
	for _, x := range r.Results {
		if x.Condition == cond && x.Op == op && x.Config == config {
			return true
		}
	}
	return false
}

func buildRow(r *Run, cond, op string, cat Category, c Config) row {
	rw := row{Config: c, Baseline: c.ID == cat.Baseline}
	var logSum float64
	var n int
	var maxRatio float64
	ratios := make([]float64, len(r.Payloads))
	for i, p := range r.Payloads {
		base := r.result(cond, op, p.ID, cat.Baseline)
		x := r.result(cond, op, p.ID, c.ID)
		if x != nil && base != nil {
			ratios[i] = base.NsPerOp / x.NsPerOp
			maxRatio = math.Max(maxRatio, ratios[i])
		}
	}
	for i, p := range r.Payloads {
		x := r.result(cond, op, p.ID, c.ID)
		if x == nil {
			reason := r.exclusion(op, p.ID, c.ID)
			if reason == "" {
				reason = "not measured"
			}
			rw.Cells = append(rw.Cells, cell{Text: "–", Title: reason, Excluded: true})
			continue
		}
		cl := cell{
			Text:  formatNs(x.NsPerOp),
			Title: fmt.Sprintf("%s/op, %d B/op, %d allocs/op", formatNs(x.NsPerOp), x.BytesPerOp, x.AllocsPerOp),
		}
		if ratios[i] > 0 {
			cl.Ratio = fmt.Sprintf("%.2fx", ratios[i])
			cl.Faster = ratios[i] > 1
			cl.Bar = 100 * ratios[i] / math.Max(maxRatio, 1)
			logSum += math.Log(ratios[i])
			n++
		}
		rw.Cells = append(rw.Cells, cl)
	}
	if n > 0 {
		rw.Mean = fmt.Sprintf("%.2fx", math.Exp(logSum/float64(n)))
	}
	return rw
}

// probesOf returns the descriptions of the probes which a category requires.
func probesOf(r *Run, category string) []string {
	seen := map[string]bool{}
	var ds []string
	for _, p := range r.Probes {
		if !p.Required || seen[p.Probe] {
			continue
		}
		for _, c := range r.Configs {
			if c.ID == p.Config && c.Category == category {
				seen[p.Probe] = true
				ds = append(ds, p.Description)
				break
			}
		}
	}
	return ds
}

func failedProbes(r *Run, config string) []ProbeResult {
	var fs []ProbeResult
	for _, p := range r.Probes {
		if p.Config == config && p.Required && !p.Passed {
			fs = append(fs, p)
		}
	}
	return fs
}

func buildProbeMatrix(r *Run) probeMatrix {
	m := probeMatrix{Passed: map[string]map[string]*ProbeResult{}}
	seen := map[string]bool{}
	for i := range r.Probes {
		p := &r.Probes[i]
		if m.Passed[p.Config] == nil {
			m.Passed[p.Config] = map[string]*ProbeResult{}
		}
		m.Passed[p.Config][p.Probe] = p
		if !seen[p.Probe] {
			seen[p.Probe] = true
			m.Probes = append(m.Probes, *p)
		}
	}
	for _, c := range r.Configs {
		if m.Passed[c.ID] != nil {
			m.Configs = append(m.Configs, c)
		}
	}
	return m
}

func formatNs(ns float64) string {
	switch {
	case ns >= 1e6:
		return fmt.Sprintf("%.2f ms", ns/1e6)
	case ns >= 1e3:
		return fmt.Sprintf("%.2f µs", ns/1e3)
	default:
		return fmt.Sprintf("%.0f ns", ns)
	}
}

// HTML renders the page of the runs.
func (s *Site) HTML() (string, error) {
	s.sortRuns()
	var views []runView
	for i, r := range s.Runs {
		views = append(views, view(r, s.Files[i]))
	}
	var first *Run
	if len(s.Runs) > 0 {
		first = s.Runs[0]
	}
	var b strings.Builder
	err := pageTemplate.Execute(&b, struct {
		First        *Run
		Views        []runView
		Attestations string
	}{first, views, s.Attestations})
	return b.String(), err
}

var pageTemplate = template.Must(template.New("page").Funcs(template.FuncMap{
	"short": func(s string) string {
		if len(s) > 12 {
			return s[:12]
		}
		return s
	},
	"sourceURL": func(r *Run, path string) string {
		if r.Repository == "" || r.Commit == "" {
			return ""
		}
		return fmt.Sprintf("https://github.com/%s/blob/%s/%s", r.Repository, r.Commit, path)
	},
	"libURL": func(l Library) string {
		switch {
		case l.Module == "std":
			return "https://pkg.go.dev/" + l.Name
		case !strings.HasPrefix(l.Version, "v"):
			// go-json of the repository, whose version is the commit measured
			return "https://" + l.Module + "/commit/" + l.Version
		}
		return "https://pkg.go.dev/" + l.Module + "@" + l.Version
	},
	"date": func(r *Run) string { return r.GeneratedAt.Format("2006-01-02 15:04 UTC") },
}).Parse(pageHTML))

const pageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Go JSON Benchmarks</title>
<style>
:root { --bg: #ffffff; --fg: #1f2328; --muted: #59636e; --line: #d1d9e0; --soft: #f6f8fa; --bar: #9ec5fe; --good: #1a7f37; --bad: #cf222e; --accent: #0969da; }
@media (prefers-color-scheme: dark) { :root:not([data-theme="light"]) { --bg: #0d1117; --fg: #e6edf3; --muted: #9198a1; --line: #3d444d; --soft: #151b23; --bar: #1f4a7a; --good: #3fb950; --bad: #f85149; --accent: #4493f8; } }
:root[data-theme="dark"] { --bg: #0d1117; --fg: #e6edf3; --muted: #9198a1; --line: #3d444d; --soft: #151b23; --bar: #1f4a7a; --good: #3fb950; --bad: #f85149; --accent: #4493f8; }
body { background: var(--bg); color: var(--fg); font: 15px/1.55 -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; margin: 0; }
main { max-width: 1180px; margin: 0 auto; padding: 24px 16px 64px; }
h1 { font-size: 28px; margin: 0 0 4px; }
h2 { font-size: 21px; margin: 40px 0 8px; padding-bottom: 6px; border-bottom: 1px solid var(--line); }
h3 { font-size: 17px; margin: 24px 0 6px; }
a { color: var(--accent); }
p, li { max-width: 860px; }
.muted { color: var(--muted); }
.meta { display: grid; grid-template-columns: max-content 1fr; gap: 2px 16px; margin: 12px 0; font-size: 14px; }
.controls { position: sticky; top: 0; background: var(--bg); padding: 10px 0; border-bottom: 1px solid var(--line); display: flex; flex-wrap: wrap; gap: 16px; z-index: 1; }
.controls fieldset { border: 0; margin: 0; padding: 0; display: flex; gap: 4px; align-items: center; }
.controls legend { float: left; margin-right: 8px; font-size: 13px; color: var(--muted); }
.controls label { border: 1px solid var(--line); border-radius: 6px; padding: 3px 10px; cursor: pointer; font-size: 14px; }
.controls input { position: absolute; opacity: 0; }
.controls input:checked + span { font-weight: 600; }
.controls label:has(input:checked) { background: var(--soft); border-color: var(--accent); }
.scroll { overflow-x: auto; }
table { border-collapse: collapse; font-size: 13.5px; margin: 6px 0 12px; }
th, td { border: 1px solid var(--line); padding: 5px 8px; text-align: right; vertical-align: top; white-space: nowrap; }
th { background: var(--soft); font-weight: 600; }
th:first-child, td:first-child { text-align: left; }
td.cell { position: relative; min-width: 92px; }
td.cell .bar { position: absolute; left: 0; bottom: 0; height: 3px; background: var(--bar); }
td.cell .ratio { display: block; font-size: 12px; color: var(--muted); }
td.cell .ratio.faster { color: var(--good); }
td.excluded { color: var(--muted); text-align: center; }
tr.baseline td:first-child { font-weight: 600; }
.setting { display: block; font-size: 12px; color: var(--muted); white-space: normal; max-width: 280px; }
.run, .cond { display: none; }
.pass { color: var(--good); }
.fail { color: var(--bad); }
code { background: var(--soft); padding: 1px 4px; border-radius: 4px; font-size: 13px; }
pre { background: var(--soft); padding: 10px 12px; border-radius: 6px; overflow-x: auto; }
</style>
</head>
<body>
<main>
<h1>Go JSON Benchmarks</h1>
<p class="muted">The JSON libraries of Go compared doing the same work, measured on GitHub Actions and updated automatically.</p>
{{with .First}}
<div class="meta">
<span class="muted">Measured</span><span>{{date .}}</span>
<span class="muted">Go</span><span>{{.GoVersion}}</span>
{{if .Commit}}<span class="muted">go-json commit</span><span><a href="https://github.com/{{.Repository}}/commit/{{.Commit}}">{{short .Commit}}</a></span>{{end}}
{{if .RunURL}}<span class="muted">Workflow run</span><span><a href="{{.RunURL}}">{{.RunURL}}</a></span>{{end}}
<span class="muted">Source</span><span>{{with sourceURL . "benchmarks/report_test.go"}}<a href="{{.}}">benchmarks/report_test.go</a>{{else}}benchmarks/report_test.go{{end}}</span>
</div>
<h2>Libraries</h2>
<table>
<tr><th>Library</th><th>Version</th></tr>
{{range .Libraries}}<tr><td><a href="{{libURL .}}">{{.Name}}</a></td><td>{{.Version}}</td></tr>
{{end}}</table>
{{end}}

<h2>How to read</h2>
<ul>
<li>Only the configurations of a category are compared with each other: every configuration of a category behaves the same, which is checked on every run ( see <a href="#fairness">Fairness</a> ). A library with an option which makes it do less work is never compared with one doing more.</li>
<li>Each cell is the time of one operation, and the speed relative to the baseline of the category ( higher is faster ). Hover a cell for the allocated bytes and the number of allocations.</li>
<li>The last column is the geometric mean of the relative speeds over the payloads.</li>
</ul>

<div class="controls">
<fieldset><legend>Machine</legend>
{{range $i, $v := .Views}}<label><input type="radio" name="run" value="{{$i}}"{{if eq $i 0}} checked{{end}}><span>{{$v.Run.GOARCH}}</span></label>
{{end}}</fieldset>
<fieldset><legend>Condition</legend>
{{with .First}}{{range $i, $c := .Conditions}}<label><input type="radio" name="cond" value="{{$c.ID}}"{{if eq $i 0}} checked{{end}}><span>{{$c.Title}}</span></label>
{{end}}{{end}}</fieldset>
</div>

{{range $i, $v := .Views}}
<section class="run" data-run="{{$i}}">
<p class="muted">{{$v.Run.GOARCH}}: {{$v.Run.CPU}} · {{$v.Run.GOOS}} · {{$v.Run.GoVersion}} · {{$v.Run.Rounds}} rounds of {{$v.Run.BenchTime}}, median · <a href="{{$v.File}}">raw results</a></p>
{{range $v.Conditions}}
<div class="cond" data-cond="{{.Condition.ID}}">
<p class="muted">{{.Condition.Description}}</p>
{{range .Categories}}
<h2>{{.Category.Title}}</h2>
<p>{{.Category.Description}}</p>
{{range .Tables}}
<h3>{{if eq .Op "encode"}}Encode{{else}}Decode{{end}}</h3>
<div class="scroll"><table>
<tr><th>Library</th>{{range .Payloads}}<th title="{{.Description}}">{{.Title}}<br><span class="muted">{{.Bytes}} B</span></th>{{end}}<th>Mean</th></tr>
{{range .Rows}}<tr{{if .Baseline}} class="baseline"{{end}}><td>{{.Config.Title}}{{if .Baseline}} <span class="muted">( baseline )</span>{{end}}<span class="setting">{{.Config.Setting}}</span></td>
{{range .Cells}}{{if .Excluded}}<td class="excluded" title="{{.Title}}">–</td>{{else}}<td class="cell" title="{{.Title}}">{{.Text}}{{if .Ratio}}<span class="ratio{{if .Faster}} faster{{end}}">{{.Ratio}}</span><span class="bar" style="width: {{printf "%.0f" .Bar}}%"></span>{{end}}</td>{{end}}{{end}}
<td>{{.Mean}}</td></tr>
{{end}}</table></div>
{{end}}
{{if .Excluded}}<p class="muted">Not compared in this category, because they behave differently:</p><ul>
{{range .Excluded}}<li>{{.Config.Title}} ( <code>{{.Config.Setting}}</code> ): {{range $j, $f := .Failed}}{{if $j}}; {{end}}{{$f.Description}} — {{$f.Detail}}{{end}}</li>
{{end}}</ul>{{end}}
{{end}}
</div>
{{end}}
</section>
{{end}}

<h2 id="fairness">Fairness</h2>
<ul>
<li>Every library is measured on the same machine, by turns, in several rounds, and the median of the rounds is shown, so that a change of the machine during the run affects every library alike. The machines of GitHub Actions differ from run to run, so the results of different runs are not compared.</li>
<li>Before it is measured, every configuration is run on probes: small inputs which tell a behavior apart. A configuration which fails a probe its category requires is not compared in that category.</li>
<li>On every payload, a decoded value must be the one encoding/json decodes, and an encoded value must be encoding/json's output byte for byte ( in "Same behavior, without HTML escaping, key sorting and string copying", the same JSON value ). A configuration which differs on a payload is not measured on it.</li>
<li>A value is decoded into a new value for each operation, and the encoders encode a pointer to the value. The input of sonic's fastest configuration is made a string once, outside of the measurement, as sonic decodes from a string.</li>
<li>"With a live heap of 64 MB" is the condition of a real program; "Without a live heap" favors a library which keeps more memory alive, as the GC then runs less often.</li>
</ul>
{{range $i, $v := .Views}}{{if eq $i 0}}
<h3>Probes</h3>
<div class="scroll"><table>
<tr><th>Probe</th>{{range $v.Probes.Configs}}<th>{{.Title}}<br><span class="muted">{{.Category}}</span></th>{{end}}</tr>
{{range $p := $v.Probes.Probes}}<tr><td>{{$p.Description}}</td>{{range $c := $v.Probes.Configs}}{{with $v.Probes.Get $c.ID $p.Probe}}<td class="{{if .Passed}}pass{{else}}fail{{end}}" title="{{.Detail}}">{{if .Passed}}yes{{else}}no{{end}}{{if .Required}}{{else}} <span class="muted">( not required )</span>{{end}}</td>{{else}}<td class="muted">–</td>{{end}}{{end}}</tr>
{{end}}</table></div>
{{end}}{{end}}

<h2>Verify the results</h2>
<p>The result files are produced by the workflow of the repository and attested with GitHub Artifact Attestations, which record the repository, the commit and the workflow run which produced them. Download a result file of this page and verify it with the GitHub CLI:</p>
<pre>gh attestation verify results-amd64.json --repo goccy/go-json</pre>
{{if .Attestations}}<p>The attestations of the repository: <a href="{{.Attestations}}">{{.Attestations}}</a></p>{{end}}
</main>
<script>
(function () {
  function show() {
    var run = document.querySelector('input[name="run"]:checked');
    var cond = document.querySelector('input[name="cond"]:checked');
    document.querySelectorAll('.run').forEach(function (el) { el.style.display = run && el.dataset.run === run.value ? 'block' : 'none'; });
    document.querySelectorAll('.cond').forEach(function (el) { el.style.display = cond && el.dataset.cond === cond.value ? 'block' : 'none'; });
  }
  document.querySelectorAll('.controls input').forEach(function (el) { el.addEventListener('change', show); });
  show();
})();
</script>
</body>
</html>
`

// Summary renders the SVG summary of the runs: for each machine, the speed of every library relative to
// encoding/json in the category of the behavior of encoding/json, with a live heap, by the geometric mean over
// the payloads.
func (s *Site) Summary() string {
	s.sortRuns()
	const (
		width    = 920
		pad      = 16
		rowH     = 22
		labelW   = 190
		barMax   = 190
		panelGap = 24
	)
	type bar struct {
		label string
		ratio float64
	}
	type panel struct {
		title string
		bars  []bar
	}
	var panels []panel
	var date, goVersion string
	for _, r := range s.Runs {
		date, goVersion = r.GeneratedAt.Format("2006-01-02"), r.GoVersion
		for _, op := range []string{OpEncode, OpDecode} {
			p := panel{title: fmt.Sprintf("%s · %s", opTitle(op), r.GOARCH)}
			for _, c := range r.configsOf("std") {
				if strings.Contains(c.ID, "/of") || len(failedProbes(r, c.ID)) > 0 {
					continue
				}
				if ratio := meanRatio(r, "live-heap", op, "std", c.ID); ratio > 0 {
					p.bars = append(p.bars, bar{label: c.Title, ratio: ratio})
				}
			}
			panels = append(panels, p)
		}
	}
	maxRatio := 1.0
	maxRows := 0
	for _, p := range panels {
		for _, b := range p.bars {
			maxRatio = math.Max(maxRatio, b.ratio)
		}
		maxRows = max(maxRows, len(p.bars))
	}
	panelH := 28 + maxRows*rowH
	cols := 2
	rows := (len(panels) + cols - 1) / cols
	height := 56 + rows*(panelH+panelGap) + pad
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" font-family="-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif">`, width, height, width, height)
	// the colors are fixed, on a background of its own, so that it reads the same on a light and a dark page
	b.WriteString(`<style>.t{font-size:13px;fill:#1f2328}.m{font-size:11px;fill:#59636e}.h{font-size:14px;font-weight:600;fill:#1f2328}.b{fill:#218bff}.g{fill:#8c959f}</style>`)
	fmt.Fprintf(&b, `<rect x="0.5" y="0.5" width="%d" height="%d" rx="6" fill="#ffffff" stroke="#d1d9e0"/>`, width-1, height-1)
	fmt.Fprintf(&b, `<text class="h" x="%d" y="%d">Speed relative to encoding/json, with the same behavior as encoding/json ( higher is faster )</text>`, pad, pad+14)
	fmt.Fprintf(&b, `<text class="m" x="%d" y="%d">%s · %s · with a live heap of 64 MB · geometric mean over the payloads</text>`, pad, pad+32, date, template.HTMLEscapeString(goVersion))
	for i, p := range panels {
		x := pad + (i%cols)*((width-2*pad)/cols)
		y := pad + 44 + (i/cols)*(panelH+panelGap)
		fmt.Fprintf(&b, `<text class="h" x="%d" y="%d">%s</text>`, x, y+16, template.HTMLEscapeString(p.title))
		for j, bb := range p.bars {
			yy := y + 28 + j*rowH
			w := float64(barMax) * bb.ratio / maxRatio
			class := "b"
			if bb.label == "encoding/json" {
				class = "g"
			}
			fmt.Fprintf(&b, `<text class="t" x="%d" y="%d">%s</text>`, x, yy+14, template.HTMLEscapeString(trimLabel(bb.label)))
			fmt.Fprintf(&b, `<rect class="%s" x="%d" y="%d" width="%.1f" height="14" rx="2"/>`, class, x+labelW, yy+3, w)
			fmt.Fprintf(&b, `<text class="t" x="%.1f" y="%d">%.2fx</text>`, float64(x+labelW)+w+6, yy+14, bb.ratio)
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func opTitle(op string) string {
	if op == OpEncode {
		return "Encode"
	}
	return "Decode"
}

func trimLabel(s string) string {
	s = strings.TrimSuffix(s, " ( DefaultOptionsV1 )")
	if s == "encoding/json/v2" {
		return "encoding/json/v2 ( v1 )"
	}
	return s
}

// meanRatio returns the geometric mean over the payloads of the speed of a configuration relative to the baseline
// of the category, or 0 if it is measured on none of them.
func meanRatio(r *Run, cond, op, category, config string) float64 {
	var base string
	for _, c := range r.Categories {
		if c.ID == category {
			base = c.Baseline
		}
	}
	var logSum float64
	var n int
	for _, p := range r.Payloads {
		b := r.result(cond, op, p.ID, base)
		x := r.result(cond, op, p.ID, config)
		if b == nil || x == nil {
			continue
		}
		logSum += math.Log(b.NsPerOp / x.NsPerOp)
		n++
	}
	if n == 0 {
		return 0
	}
	return math.Exp(logSum / float64(n))
}

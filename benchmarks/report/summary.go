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
		label   string
		library string
		ratio   float64
		differs bool
	}
	type panel struct {
		title string
		bars  []bar
	}
	var panels []panel
	var date, goVersion string
	differing := map[string]bool{}
	for _, r := range s.Runs {
		date, goVersion = r.GeneratedAt.Format("2006-01-02"), r.GoVersion
		for _, op := range []string{OpEncode, OpDecode} {
			var shown []string
			for _, c := range r.configsOf("std") {
				if !strings.Contains(c.ID, "/of") && r.measured("live-heap", op, c.ID) {
					shown = append(shown, c.ID)
				}
			}
			// the libraries are compared on the same payloads: the ones which all of them do alike
			payloads := r.sharedPayloads("live-heap", op, "std", shown)
			p := panel{title: fmt.Sprintf("%s · %s · %d of %d payloads", opTitle(op), r.GOARCH, len(payloads), len(r.Payloads))}
			// the configurations which behave as encoding/json first, then the others, marked
			for _, differs := range []bool{false, true} {
				for _, c := range r.configsOf("std") {
					if strings.Contains(c.ID, "/of") || (len(failedProbes(r, c.ID)) > 0) != differs {
						continue
					}
					if ratio := meanRatio(r, "live-heap", op, "std", c.ID, payloads); ratio > 0 {
						label := trimLabel(c.Title)
						if differs {
							label += " †"
							differing[c.Title] = true
						}
						p.bars = append(p.bars, bar{label: label, library: c.Library, ratio: ratio, differs: differs})
					}
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
	if len(differing) > 0 {
		height += 16
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" font-family="-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif">`, width, height, width, height)
	// the colors are fixed, on a background of its own, so that it reads the same on a light and a dark page
	b.WriteString(`<style>.t{font-size:13px;fill:#1f2328}.m{font-size:11px;fill:#59636e}.h{font-size:14px;font-weight:600;fill:#1f2328}.one{stroke:#8c959f;stroke-dasharray:3 3}</style>`)
	// the bars of a library are of its color, which the page gives it too; the bars of a configuration which behaves
	// differently are hatched, as on the page
	b.WriteString(`<defs>`)
	for i, c := range libraryColors {
		fmt.Fprintf(&b, `<pattern id="h%d" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><rect width="6" height="6" fill="%s" fill-opacity="0.45"/><rect width="3" height="6" fill="%s"/></pattern>`, i, c.color, c.color)
	}
	b.WriteString(`</defs>`)
	fmt.Fprintf(&b, `<rect x="0.5" y="0.5" width="%d" height="%d" rx="6" fill="#ffffff" stroke="#d1d9e0"/>`, width-1, height-1)
	fmt.Fprintf(&b, `<text class="h" x="%d" y="%d">Speed relative to encoding/json, with the same behavior as encoding/json ( higher is faster )</text>`, pad, pad+14)
	fmt.Fprintf(&b, `<text class="m" x="%d" y="%d">%s · %s · with a live heap of 64 MB · geometric mean over the payloads which every library does alike</text>`, pad, pad+32, date, template.HTMLEscapeString(goVersion))
	for i, p := range panels {
		x := pad + (i%cols)*((width-2*pad)/cols)
		y := pad + 44 + (i/cols)*(panelH+panelGap)
		fmt.Fprintf(&b, `<text class="h" x="%d" y="%d">%s</text>`, x, y+16, template.HTMLEscapeString(p.title))
		for j, bb := range p.bars {
			yy := y + 28 + j*rowH
			w := float64(barMax) * bb.ratio / maxRatio
			i := libraryColorIndex(bb.library)
			fill := libraryColors[i].color
			if bb.differs {
				fill = fmt.Sprintf("url(#h%d)", i)
			}
			fmt.Fprintf(&b, `<text class="t" x="%d" y="%d">%s</text>`, x, yy+14, template.HTMLEscapeString(bb.label))
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%.1f" height="14" rx="2" fill="%s"/>`, x+labelW, yy+3, w, fill)
			fmt.Fprintf(&b, `<text class="t" x="%.1f" y="%d">%.2fx</text>`, float64(x+labelW)+w+6, yy+14, bb.ratio)
		}
		if len(p.bars) > 0 {
			// the line of the speed of encoding/json
			one := float64(x+labelW) + float64(barMax)/maxRatio
			fmt.Fprintf(&b, `<line class="one" x1="%.1f" y1="%d" x2="%.1f" y2="%d"/>`, one, y+26, one, y+28+len(p.bars)*rowH)
		}
	}
	if len(differing) > 0 {
		var names []string
		for n := range differing {
			names = append(names, n)
		}
		sort.Strings(names)
		fmt.Fprintf(&b, `<text class="m" x="%d" y="%d">† %s: behaves differently from encoding/json, so not the same work ( see the page for what differs )</text>`,
			pad, height-pad, template.HTMLEscapeString(strings.Join(names, ", ")))
	}
	b.WriteString(`</svg>`)
	return b.String()
}

// libraryColors are the colors of the libraries, the ones of the page ( page.html ) on a light background.
var libraryColors = []struct{ library, color string }{
	{"encoding/json", "#8c959f"},
	{"encoding/json/v2", "#57606a"},
	{"goccy/go-json", "#0969da"},
	{"bytedance/sonic", "#e16f24"},
	{"json-iterator/go", "#1a7f37"},
	{"segmentio/encoding", "#8250df"},
	{"velox-io/json", "#bf3989"},
	{"", "#59636e"},
}

// libraryColorIndex returns the index of the color of the library in libraryColors: the last one for a library
// which has none.
func libraryColorIndex(library string) int {
	for i, c := range libraryColors {
		if c.library == library {
			return i
		}
	}
	return len(libraryColors) - 1
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
func meanRatio(r *Run, cond, op, category, config string, payloads []string) float64 {
	base := r.baselineOf(category)
	var logSum float64
	var n int
	for _, p := range payloads {
		b := r.result(cond, op, p, base)
		x := r.result(cond, op, p, config)
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

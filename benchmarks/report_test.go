package benchmark

import (
	"bytes"
	stdjson "encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"benchmark/report"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
	segmentio "github.com/segmentio/encoding/json"
)

// The benchmark report compares the JSON libraries of Go doing the same work. It is run by TestReport when
// BENCH_REPORT_OUT names the file to write the results to ( see the Makefile and the workflow of the report ).
//
// A comparison is fair only between configurations which behave the same, so the configurations are grouped in
// categories, and only the ones of a category are compared with each other:
//
//   - "std" behaves as encoding/json: HTML is escaped, the keys of a map are sorted, invalid UTF-8 is replaced,
//     what a Marshaler returns is validated and compacted, the decoded strings are copies of the input, the
//     skipped values are validated, and the keys match the fields regardless of case.
//   - "fast" drops only what a program may do without: HTML is not escaped, the keys of a map are in any order,
//     and the decoded strings may refer to the input. Everything else is as in "std".
//   - "v2" is encoding/json/v2 with its default options, which behaves differently from all the others.
//
// What a category requires is checked, not assumed: every configuration is run on the probes of its category
// before it is measured, and one which fails a required probe is reported and not measured. The result of every
// configuration on every payload is also compared with the one of the category ( see checkPayload ).

// reportCategories are the categories of the configurations.
var reportCategories = []report.Category{
	{
		ID:    "std",
		Title: "Same behavior as encoding/json",
		Description: "HTML is escaped, the keys of a map are sorted, invalid UTF-8 is replaced by U+FFFD, the output of a " +
			"Marshaler is validated and compacted, the decoded strings are copies, the values of the unknown keys are " +
			"validated, and the keys match the fields regardless of case. The output is the same as encoding/json's, byte " +
			"for byte.",
		Baseline: "encoding/json",
	},
	{
		ID:    "fast",
		Title: "Same behavior, without HTML escaping, key sorting and string copying",
		Description: "As \"Same behavior as encoding/json\", except that HTML is not escaped, the keys of a map are in any " +
			"order, and the decoded strings may refer to the input, which must then not be modified while they are used. " +
			"Invalid UTF-8 is still replaced, the output of a Marshaler is still validated, and the values of the unknown " +
			"keys are still validated.",
		Baseline: "go-json/fast",
	},
	{
		ID:    "v2",
		Title: "encoding/json/v2 with its default options",
		Description: "The default behavior of encoding/json/v2 differs from encoding/json's in many ways ( for example, the " +
			"keys match the fields by case, duplicate keys and invalid UTF-8 are errors, and a nil slice is encoded as [] ), " +
			"and no other library behaves the same, so it is shown alone.",
		Baseline: "encoding/json/v2",
	},
}

// reportConfig is a library called in one way.
type reportConfig struct {
	report.Config
	// module is the module of the library, whose version the report shows.
	module string
	// marshal encodes a value.
	marshal func(v any) ([]byte, error)
	// unmarshal returns the function which decodes the data into a value: what it prepares, as the string
	// sonic decodes from, is made once, outside of the measured loop.
	unmarshal func(data []byte) func(v any) error
	// decodeOf is set for go-json's UnmarshalOf, which takes the value by its type: the options it is called with.
	decodeOf []gojson.DecodeOptionFunc
}

func stdUnmarshal(f func([]byte, any) error) func([]byte) func(any) error {
	return func(data []byte) func(any) error {
		return func(v any) error { return f(data, v) }
	}
}

// sonicStringUnmarshal decodes by sonic from a string, as it is at its fastest: the string is made once.
func sonicStringUnmarshal(api sonic.API) func([]byte) func(any) error {
	return func(data []byte) func(any) error {
		s := string(data)
		return func(v any) error { return api.UnmarshalFromString(s, v) }
	}
}

var (
	jsoniterFast = jsoniter.Config{EscapeHTML: false, SortMapKeys: false, ValidateJsonRawMessage: true}.Froze()
	sonicFast    = sonic.Config{CompactMarshaler: true, ValidateString: true}.Froze()
)

// reportConfigs are the configurations measured, by category. The first of a category is its baseline. The
// configurations of encoding/json/v2 are added where the Go version has it ( report_v2_test.go ).
var reportConfigs = []*reportConfig{
	{
		Config: report.Config{ID: "encoding/json", Library: "encoding/json", Category: "std", Title: "encoding/json",
			Setting: "json.Marshal / json.Unmarshal"},
		module: "std", marshal: stdjson.Marshal, unmarshal: stdUnmarshal(stdjson.Unmarshal),
	},
	{
		Config: report.Config{ID: "go-json", Library: "goccy/go-json", Category: "std", Title: "goccy/go-json",
			Setting: "json.Marshal / json.Unmarshal"},
		module: "github.com/goccy/go-json", marshal: gojson.Marshal, unmarshal: stdUnmarshal(gojson.Unmarshal),
	},
	{
		Config: report.Config{ID: "go-json/of", Library: "goccy/go-json", Category: "std", Title: "goccy/go-json ( UnmarshalOf )",
			Setting: "json.UnmarshalOf, which takes the value by its type ( decode only )"},
		module: "github.com/goccy/go-json", decodeOf: []gojson.DecodeOptionFunc{},
	},
	{
		Config: report.Config{ID: "sonic/std", Library: "bytedance/sonic", Category: "std", Title: "bytedance/sonic",
			Setting: "sonic.ConfigStd"},
		module: "github.com/bytedance/sonic", marshal: sonic.ConfigStd.Marshal, unmarshal: stdUnmarshal(sonic.ConfigStd.Unmarshal),
	},
	{
		Config: report.Config{ID: "jsoniter/std", Library: "json-iterator/go", Category: "std", Title: "json-iterator/go",
			Setting: "jsoniter.ConfigCompatibleWithStandardLibrary"},
		module: "github.com/json-iterator/go", marshal: jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
		unmarshal: stdUnmarshal(jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal),
	},
	{
		Config: report.Config{ID: "segmentio/std", Library: "segmentio/encoding", Category: "std", Title: "segmentio/encoding",
			Setting: "json.Marshal / json.Unmarshal"},
		module: "github.com/segmentio/encoding", marshal: segmentio.Marshal, unmarshal: stdUnmarshal(segmentio.Unmarshal),
	},
	{
		Config: report.Config{ID: "go-json/fast", Library: "goccy/go-json", Category: "fast", Title: "goccy/go-json",
			Setting: "json.MarshalWithOption( DisableHTMLEscape, UnorderedMap ) / json.UnmarshalWithOption( DecodeNoCopyString )"},
		module: "github.com/goccy/go-json",
		marshal: func(v any) ([]byte, error) {
			return gojson.MarshalWithOption(v, gojson.DisableHTMLEscape(), gojson.UnorderedMap())
		},
		unmarshal: stdUnmarshal(func(data []byte, v any) error {
			return gojson.UnmarshalWithOption(data, v, gojson.DecodeNoCopyString())
		}),
	},
	{
		Config: report.Config{ID: "go-json/fast-of", Library: "goccy/go-json", Category: "fast", Title: "goccy/go-json ( UnmarshalOf )",
			Setting: "json.UnmarshalOf( DecodeNoCopyString ), which takes the value by its type ( decode only )"},
		module: "github.com/goccy/go-json", decodeOf: []gojson.DecodeOptionFunc{gojson.DecodeNoCopyString()},
	},
	{
		Config: report.Config{ID: "sonic/fast", Library: "bytedance/sonic", Category: "fast", Title: "bytedance/sonic",
			Setting: "sonic.Config{ CompactMarshaler: true, ValidateString: true }; decoded from a string ( UnmarshalFromString )"},
		module: "github.com/bytedance/sonic", marshal: sonicFast.Marshal, unmarshal: sonicStringUnmarshal(sonicFast),
	},
	{
		Config: report.Config{ID: "jsoniter/fast", Library: "json-iterator/go", Category: "fast", Title: "json-iterator/go",
			Setting: "jsoniter.Config{ EscapeHTML: false, SortMapKeys: false, ValidateJsonRawMessage: true }"},
		module: "github.com/json-iterator/go", marshal: jsoniterFast.Marshal, unmarshal: stdUnmarshal(jsoniterFast.Unmarshal),
	},
	{
		Config: report.Config{ID: "segmentio/fast", Library: "segmentio/encoding", Category: "fast", Title: "segmentio/encoding",
			Setting: "json.Append( 0 ) / json.Parse( ZeroCopy )"},
		module:  "github.com/segmentio/encoding",
		marshal: func(v any) ([]byte, error) { return segmentio.Append(nil, v, 0) },
		unmarshal: stdUnmarshal(func(data []byte, v any) error {
			_, err := segmentio.Parse(data, v, segmentio.ZeroCopy)
			return err
		}),
	},
}

// reportPayload is an input of the benchmarks, decoded into its Go type T.
type reportPayload struct {
	report.Payload
	data []byte
	// newValue returns a pointer to a new zero value of T.
	newValue func() any
	// decodeOf decodes the data into a new value of T by go-json's UnmarshalOf.
	decodeOf func(opts []gojson.DecodeOptionFunc) error
	// value is a pointer to the value of the data, decoded by encoding/json, which the encoders encode.
	value any
}

func newReportPayload[T any](id, title, description string, data []byte) *reportPayload {
	v := new(T)
	if err := stdjson.Unmarshal(data, v); err != nil {
		panic(fmt.Sprintf("%s: %v", id, err))
	}
	return &reportPayload{
		Payload:  report.Payload{ID: id, Title: title, Description: description, Bytes: len(data)},
		data:     data,
		newValue: func() any { return new(T) },
		decodeOf: func(opts []gojson.DecodeOptionFunc) error {
			var v T
			return gojson.UnmarshalOf(data, &v, opts...)
		},
		value: v,
	}
}

func reportPayloads() []*reportPayload {
	codeInit()
	return []*reportPayload{
		newReportPayload[SmallPayload]("small", "Small struct", "A flat struct of 9 fields.", SmallFixture),
		newReportPayload[MediumPayload]("medium", "Medium struct", "Nested structs and a slice of structs.", MediumFixture),
		newReportPayload[LargePayload]("large", "Large struct", "Slices of many small structs.", LargeFixture),
		newReportPayload[TwitterStruct]("twitter", "Twitter ( struct )", "A Twitter search result, into its struct ( the payload of sonic's benchmarks ).", []byte(TwitterJson)),
		newReportPayload[any]("twitter-any", "Twitter ( interface{} )", "The same Twitter search result, into interface{}.", []byte(TwitterJson)),
		newReportPayload[[]*GitHubIssue]("github", "GitHub REST issues", "A page of issues of the GitHub REST API.", githubRESTIssues),
		newReportPayload[Response]("openai", "OpenAI Responses", "A response of the OpenAI Responses API, with long escaped strings.", openAIResponsesResponseJSON),
		newReportPayload[codeResponse]("code", "Go source tree", "The large tree of encoding/json's own benchmark ( 1.9 MB ).", codeJSON),
	}
}

// reportConditions are the conditions the benchmarks are measured in.
var reportConditions = []report.Condition{
	{
		ID:    "live-heap",
		Title: "With a live heap of 64 MB",
		Description: "The process keeps 64 MB alive, as a real program keeps its own data: the GC runs as often for every " +
			"library, since its goal is twice the live heap.",
	},
	{
		ID:    "no-live-heap",
		Title: "Without a live heap",
		Description: "The process keeps nothing alive but what the libraries keep, as their caches and pools: a library " +
			"which keeps more is collected less often, which a real program doesn't see.",
	},
}

func TestReport(t *testing.T) {
	out := os.Getenv("BENCH_REPORT_OUT")
	if out == "" {
		t.Skip("BENCH_REPORT_OUT is not set")
	}
	rounds := 3
	if n, err := strconv.Atoi(os.Getenv("BENCH_REPORT_ROUNDS")); err == nil && n > 0 {
		rounds = n
	}
	pretouchSonic()
	payloads := reportPayloads()
	run := &report.Run{
		GeneratedAt: time.Now().UTC(),
		GoVersion:   runtime.Version(),
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		CPU:         reportCPU(),
		Commit:      os.Getenv("GITHUB_SHA"),
		Repository:  os.Getenv("GITHUB_REPOSITORY"),
		Rounds:      rounds,
		Categories:  reportCategories,
		Conditions:  reportConditions,
		Libraries:   reportLibraries(),
	}
	if f := flag.Lookup("test.benchtime"); f != nil {
		run.BenchTime = f.Value.String()
	}
	if os.Getenv("GITHUB_RUN_ID") != "" {
		run.RunURL = os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/actions/runs/" + os.Getenv("GITHUB_RUN_ID")
	}
	for _, p := range payloads {
		run.Payloads = append(run.Payloads, p.Payload)
	}

	// the configurations which pass the probes of their category, and the payloads each of them is measured on
	type job struct {
		op     string
		p      *reportPayload
		c      *reportConfig
		sample []testing.BenchmarkResult
	}
	var jobs []*job
	for _, c := range reportConfigs {
		run.Configs = append(run.Configs, c.Config)
		probes := runProbes(c)
		run.Probes = append(run.Probes, probes...)
		failed := failedProbes(probes)
		for _, op := range []string{report.OpEncode, report.OpDecode} {
			if !c.does(op) {
				continue
			}
			for _, p := range payloads {
				reason := failed
				if reason == "" {
					reason = checkPayload(c, op, p)
				}
				if reason != "" {
					run.Exclusions = append(run.Exclusions, report.Exclusion{Op: op, Payload: p.ID, Config: c.ID, Reason: reason})
					continue
				}
				jobs = append(jobs, &job{op: op, p: p, c: c})
			}
		}
	}

	for _, cond := range reportConditions {
		if cond.ID == "live-heap" {
			liveHeap = make([]byte, 64<<20)
		} else {
			liveHeap = nil
		}
		runtime.GC()
		for _, j := range jobs {
			j.sample = j.sample[:0]
		}
		// the configurations are measured by turns, in rounds, so that a change of the machine during the run
		// affects all of them alike; the median of the rounds is reported.
		for round := 0; round < rounds; round++ {
			for i := range jobs {
				j := jobs[(i+round*len(jobs)/rounds)%len(jobs)]
				j.sample = append(j.sample, measure(j.c, j.op, j.p))
			}
		}
		for _, j := range jobs {
			m := median(j.sample)
			run.Results = append(run.Results, report.Result{
				Condition: cond.ID, Op: j.op, Payload: j.p.ID, Config: j.c.ID,
				NsPerOp:     float64(m.T.Nanoseconds()) / float64(m.N),
				BytesPerOp:  m.AllocedBytesPerOp(),
				AllocsPerOp: m.AllocsPerOp(),
			})
		}
	}
	liveHeap = nil
	if err := report.WriteRun(out, run); err != nil {
		t.Fatal(err)
	}
}

// does reports whether the configuration does the operation.
func (c *reportConfig) does(op string) bool {
	if op == report.OpEncode {
		return c.marshal != nil
	}
	return c.unmarshal != nil || c.decodeOf != nil
}

// measure measures the configuration on the payload.
func measure(c *reportConfig, op string, p *reportPayload) testing.BenchmarkResult {
	return testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(p.data)))
		switch {
		case op == report.OpEncode:
			for i := 0; i < b.N; i++ {
				if _, err := c.marshal(p.value); err != nil {
					b.Fatal(err)
				}
			}
		case c.decodeOf != nil:
			for i := 0; i < b.N; i++ {
				if err := p.decodeOf(c.decodeOf); err != nil {
					b.Fatal(err)
				}
			}
		default:
			unmarshal := c.unmarshal(p.data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := unmarshal(p.newValue()); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func median(rs []testing.BenchmarkResult) testing.BenchmarkResult {
	sort.Slice(rs, func(i, j int) bool { return nsPerOp(rs[i]) < nsPerOp(rs[j]) })
	return rs[len(rs)/2]
}

func nsPerOp(r testing.BenchmarkResult) float64 {
	return float64(r.T.Nanoseconds()) / float64(r.N)
}

// checkPayload returns why the result of the configuration on the payload is not the one of its category, or "":
// a decoded value must be the one encoding/json decodes; an encoded value must be encoding/json's output byte for
// byte in "std", and the same JSON value in "fast". The output of encoding/json/v2 is only checked to be valid.
func checkPayload(c *reportConfig, op string, p *reportPayload) string {
	if op == report.OpDecode {
		if c.decodeOf != nil {
			// UnmarshalOf decodes the value as Unmarshal does: it is checked by a decode into the value.
			v := p.newValue()
			if err := gojson.UnmarshalWithOption(p.data, v, c.decodeOf...); err != nil {
				return err.Error()
			}
			if !reflect.DeepEqual(v, p.value) {
				return "the decoded value differs from encoding/json's"
			}
			return ""
		}
		v := p.newValue()
		if err := c.unmarshal(p.data)(v); err != nil {
			return err.Error()
		}
		if c.Category != "v2" && !reflect.DeepEqual(v, p.value) {
			return "the decoded value differs from encoding/json's"
		}
		return ""
	}
	got, err := c.marshal(p.value)
	if err != nil {
		return err.Error()
	}
	switch c.Category {
	case "std":
		want, _ := stdjson.Marshal(p.value)
		if !bytes.Equal(got, want) {
			return "the output differs from encoding/json's"
		}
	case "fast":
		want, _ := stdjson.Marshal(p.value)
		if !sameJSON(got, want) {
			return "the output is not the same JSON value as encoding/json's"
		}
	default:
		if !stdjson.Valid(got) {
			return "the output is not valid JSON"
		}
	}
	return ""
}

// sameJSON reports whether a and b are the same JSON value, whatever the order of the keys of their objects.
func sameJSON(a, b []byte) bool {
	var x, y any
	da, db := stdjson.NewDecoder(bytes.NewReader(a)), stdjson.NewDecoder(bytes.NewReader(b))
	da.UseNumber()
	db.UseNumber()
	if da.Decode(&x) != nil || db.Decode(&y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}

// reportProbe is a small input which tells a behavior apart. check returns "" if the configuration behaves as the
// category requires, else what it did.
type reportProbe struct {
	id, description string
	op              string
	// categories are the categories which require the behavior.
	categories []string
	check      func(c *reportConfig) string
}

type probeMarshaler struct{}

func (probeMarshaler) MarshalJSON() ([]byte, error) { return []byte(`{ "a" : [ 1 , 2 ] }`), nil }

type probeFields struct {
	Name  string
	Small int8
}

// encodeProbe checks that the configuration encodes v as want, or, where exact is false, as the same JSON value.
func encodeProbe(v any, want string, exact bool) func(c *reportConfig) string {
	return func(c *reportConfig) string {
		got, err := c.marshal(v)
		if err != nil {
			return "error: " + err.Error()
		}
		if exact && string(got) != want || !exact && !sameJSON(got, []byte(want)) {
			return fmt.Sprintf("got %s, want %s", got, want)
		}
		return ""
	}
}

// decodeProbe decodes data into a new value of newValue's type, and checks the value, or that it is an error when
// check is nil.
func decodeProbe(data string, newValue func() any, check func(v any) string) func(c *reportConfig) string {
	return func(c *reportConfig) string {
		v := newValue()
		var err error
		if c.decodeOf != nil {
			err = gojson.UnmarshalWithOption([]byte(data), v, c.decodeOf...)
		} else {
			err = c.unmarshal([]byte(data))(v)
		}
		if check == nil {
			if err == nil {
				return "accepted invalid input"
			}
			return ""
		}
		if err != nil {
			return "error: " + err.Error()
		}
		return check(v)
	}
}

var both = []string{"std", "fast"}

var reportProbes = []reportProbe{
	{id: "escape-html", op: report.OpEncode, categories: []string{"std"},
		description: `HTML characters are escaped: "<&>" is encoded as "\u003c\u0026\u003e".`,
		check:       encodeProbe("<&>", `"\u003c\u0026\u003e"`, true)},
	{id: "sort-map-keys", op: report.OpEncode, categories: []string{"std"},
		description: "The keys of a map are sorted.",
		check:       encodeProbe(map[string]int{"b": 1, "c": 2, "a": 3}, `{"a":3,"b":1,"c":2}`, true)},
	{id: "invalid-utf8-encode", op: report.OpEncode, categories: both,
		description: `Invalid UTF-8 is encoded as U+FFFD.`,
		check:       encodeProbe("a\xffb", `"a\ufffdb"`, false)},
	{id: "floats", op: report.OpEncode, categories: both,
		description: "Floats are formatted as encoding/json formats them.",
		check:       encodeProbe([]float64{1e21, 1e20, 1e-7, 0.000001, 0.1, 123456789, -0.5}, `[1e+21,100000000000000000000,1e-7,0.000001,0.1,123456789,-0.5]`, true)},
	{id: "marshaler-compact", op: report.OpEncode, categories: both,
		description: "The output of a Marshaler is compacted.",
		check:       encodeProbe(probeMarshaler{}, `{"a":[1,2]}`, true)},
	{id: "invalid-utf8-decode", op: report.OpDecode, categories: both,
		description: "Invalid UTF-8 in a string is decoded as U+FFFD.",
		check: decodeProbe("\"a\xffb\"", func() any { return new(string) }, func(v any) string {
			if s := *v.(*string); s != "a�b" {
				return fmt.Sprintf("got %q", s)
			}
			return ""
		})},
	{id: "control-character", op: report.OpDecode, categories: both,
		description: "A control character in a string is an error.",
		check:       decodeProbe("\"a\x01\"", func() any { return new(string) }, nil)},
	{id: "case-insensitive", op: report.OpDecode, categories: both,
		description: "A key matches a field regardless of case.",
		check: decodeProbe(`{"NAME":"x"}`, func() any { return new(probeFields) }, func(v any) string {
			if n := v.(*probeFields).Name; n != "x" {
				return fmt.Sprintf("got %q", n)
			}
			return ""
		})},
	{id: "skipped-value-validated", op: report.OpDecode, categories: both,
		description: "The value of an unknown key is validated.",
		check:       decodeProbe(`{"unknown":[1,,2],"Name":"x"}`, func() any { return new(probeFields) }, nil)},
	{id: "duplicate-key", op: report.OpDecode, categories: both,
		description: "The last of duplicate keys wins.",
		check: decodeProbe(`{"Name":"a","Name":"b"}`, func() any { return new(probeFields) }, func(v any) string {
			if n := v.(*probeFields).Name; n != "b" {
				return fmt.Sprintf("got %q", n)
			}
			return ""
		})},
	{id: "overflow", op: report.OpDecode, categories: both,
		description: "A number out of the range of its field is an error.",
		check:       decodeProbe(`{"Small":300}`, func() any { return new(probeFields) }, nil)},
	{id: "string-copied", op: report.OpDecode, categories: []string{"std"},
		description: "A decoded string is a copy: it doesn't change when the input is modified.",
		check: func(c *reportConfig) string {
			data := []byte(`{"Name":"abcdefghijklmnopqrstuvwxyz"}`)
			v := new(probeFields)
			var err error
			if c.decodeOf != nil {
				err = gojson.UnmarshalWithOption(data, v, c.decodeOf...)
			} else {
				err = c.unmarshal(data)(v)
			}
			if err != nil {
				return "error: " + err.Error()
			}
			copy(data[9:], "ZZZZ")
			if v.Name != "abcdefghijklmnopqrstuvwxyz" {
				return "the string refers to the input"
			}
			return ""
		}},
}

// runProbes runs the probes of the category of the configuration.
func runProbes(c *reportConfig) []report.ProbeResult {
	var rs []report.ProbeResult
	for _, p := range reportProbes {
		if !c.does(p.op) {
			continue
		}
		required := false
		for _, cat := range p.categories {
			required = required || cat == c.Category
		}
		if c.Category == "v2" {
			continue
		}
		detail := safeProbe(p, c)
		rs = append(rs, report.ProbeResult{
			Config: c.ID, Probe: p.id, Description: p.description, Required: required, Passed: detail == "", Detail: detail,
		})
	}
	return rs
}

// safeProbe runs a probe, reporting a panic of the library as its result.
func safeProbe(p reportProbe, c *reportConfig) (detail string) {
	defer func() {
		if r := recover(); r != nil {
			detail = fmt.Sprintf("panic: %v", r)
		}
	}()
	return p.check(c)
}

// failedProbes returns the required probes which failed, or "".
func failedProbes(rs []report.ProbeResult) string {
	var failed []string
	for _, r := range rs {
		if r.Required && !r.Passed {
			failed = append(failed, r.Probe)
		}
	}
	if len(failed) == 0 {
		return ""
	}
	return "fails the required probes: " + strings.Join(failed, ", ")
}

// reportLibraries returns the libraries measured, with the versions the benchmarks are built with.
func reportLibraries() []report.Library {
	versions := map[string]string{"std": runtime.Version()}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, d := range info.Deps {
			v := d.Version
			if d.Replace != nil {
				v = d.Replace.Version
			}
			versions[d.Path] = v
		}
	}
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		// go-json is the one of the repository, replaced by its directory
		versions["github.com/goccy/go-json"] = sha
	}
	seen := map[string]bool{}
	var libs []report.Library
	for _, c := range reportConfigs {
		if seen[c.Library] {
			continue
		}
		seen[c.Library] = true
		v := versions[c.module]
		if v == "" {
			v = "(devel)"
		}
		libs = append(libs, report.Library{Name: c.Library, Module: c.module, Version: v})
	}
	return libs
}

// reportCPU returns the model of the CPU: BENCH_CPU, which the workflow sets, or the model in /proc/cpuinfo.
func reportCPU() string {
	if cpu := os.Getenv("BENCH_CPU"); cpu != "" {
		return cpu
	}
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return runtime.GOARCH
	}
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(k) == "model name" {
			return strings.TrimSpace(v)
		}
	}
	return runtime.GOARCH
}

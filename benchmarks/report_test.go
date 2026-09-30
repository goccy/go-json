package benchmark

import (
	"bytes"
	stdjson "encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
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
// A comparison is fair only between configurations which do the same work, so the results are shown by category of
// behavior, and every library is in every category:
//
//   - "std" behaves as encoding/json: HTML is escaped, the keys of a map are sorted, invalid UTF-8 is replaced,
//     what a Marshaler returns is validated and compacted, the decoded strings are copies of the input, the
//     skipped values are validated, and the keys match the fields regardless of case.
//   - "fast" drops only what a program may do without: HTML is not escaped, the keys of a map are in any order,
//     and the decoded strings may refer to the input. Everything else is as in "std".
//   - "v2" behaves as encoding/json/v2 with its default options ( report_v2_test.go, where Go has it ).
//   - "fastest" is every library with the options which make it fastest, whatever they drop: its probes show what
//     each of them does differently from encoding/json.
//
// In a category, a library is configured by its options to behave as the category requires, as far as its
// options allow, and then as fast as they allow. A library whose options can't make it behave so is still shown,
// configured as close as they allow, and the report says what differs.
//
// What a category requires is checked, not assumed: every configuration is run on the probes of its category, small
// inputs which tell a behavior apart, and the result of every configuration on every payload is compared with the
// one of the baseline of the category ( see checkPayload ). A configuration which behaves differently is still
// measured: the report shows it apart from the others of its category, with what differs. Only an operation which
// fails on a payload is not measured.

// reportCategories are the categories of the configurations. The category of encoding/json/v2 is added where Go
// has it ( report_v2_test.go ).
var reportCategories = []report.Category{
	{
		ID:    "std",
		Title: "Same behavior as encoding/json",
		Description: "HTML is escaped, the keys of a map are sorted, invalid UTF-8 is replaced by U+FFFD, the output of a " +
			"Marshaler is validated and compacted, the decoded strings are copies, the values of the unknown keys are " +
			"validated, and the keys match the fields regardless of case. The output is the same as encoding/json's, byte " +
			"for byte.",
		Baseline: "std/encoding/json",
	},
	{
		ID:    "fast",
		Title: "Same behavior, without HTML escaping, key sorting and string copying",
		Description: "As \"Same behavior as encoding/json\", except that HTML is not escaped, the keys of a map may be in " +
			"any order, and the decoded strings may refer to the input, which must then not be modified while they are " +
			"used. Invalid UTF-8 is still replaced, the output of a Marshaler is still validated, and the values of the " +
			"unknown keys are still validated. Every library is configured as fast as its options allow in this behavior.",
		Baseline: "fast/encoding/json",
	},
	{
		ID:    "fastest",
		Title: "Every library at its fastest",
		Description: "Every library with the options which make it fastest, whatever they make it skip. The configurations " +
			"don't do the same work: what each of them does differently from encoding/json is listed below the charts, " +
			"from the probes.",
		Baseline: "fastest/encoding/json",
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

// stdMarshalNoHTMLEscape encodes as json.Marshal does without HTML escaping, which encoding/json does only by an
// Encoder: its newline is dropped.
func stdMarshalNoHTMLEscape(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := stdjson.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b.Bytes(), []byte{'\n'}), nil
}

func goJSONMarshal(opts ...gojson.EncodeOptionFunc) func(any) ([]byte, error) {
	return func(v any) ([]byte, error) { return gojson.MarshalWithOption(v, opts...) }
}

func goJSONUnmarshal(opts ...gojson.DecodeOptionFunc) func([]byte) func(any) error {
	return stdUnmarshal(func(data []byte, v any) error { return gojson.UnmarshalWithOption(data, v, opts...) })
}

func segmentioMarshal(flags segmentio.AppendFlags) func(any) ([]byte, error) {
	return func(v any) ([]byte, error) { return segmentio.Append(nil, v, flags) }
}

func segmentioUnmarshal(flags segmentio.ParseFlags) func([]byte) func(any) error {
	return stdUnmarshal(func(data []byte, v any) error {
		rest, err := segmentio.Parse(data, v, flags)
		if err == nil && len(bytes.TrimSpace(rest)) != 0 {
			return fmt.Errorf("invalid character after top-level value")
		}
		return err
	})
}

const (
	goJSONModule    = "github.com/goccy/go-json"
	sonicModule     = "github.com/bytedance/sonic"
	jsoniterModule  = "github.com/json-iterator/go"
	segmentioModule = "github.com/segmentio/encoding"
)

// reportConfigs are the configurations measured, by category; the first of a category is its baseline. The
// configurations of encoding/json/v2 are added where Go has it ( report_v2_test.go ).
var reportConfigs = []*reportConfig{
	// std
	{
		Config: report.Config{ID: "std/encoding/json", Library: "encoding/json", Category: "std", Title: "encoding/json",
			Setting: "json.Marshal / json.Unmarshal"},
		module: "std", marshal: stdjson.Marshal, unmarshal: stdUnmarshal(stdjson.Unmarshal),
	},
	{
		Config: report.Config{ID: "std/go-json", Library: "goccy/go-json", Category: "std", Title: "goccy/go-json",
			Setting: "json.Marshal / json.Unmarshal"},
		module: goJSONModule, marshal: gojson.Marshal, unmarshal: stdUnmarshal(gojson.Unmarshal),
	},
	{
		Config: report.Config{ID: "std/go-json/of", Library: "goccy/go-json", Category: "std", Title: "goccy/go-json ( UnmarshalOf )",
			Setting: "json.UnmarshalOf, which takes the value by its type ( decode only )"},
		module: goJSONModule, decodeOf: []gojson.DecodeOptionFunc{},
	},
	{
		Config: report.Config{ID: "std/sonic", Library: "bytedance/sonic", Category: "std", Title: "bytedance/sonic",
			Setting: "sonic.ConfigStd"},
		module: sonicModule, marshal: sonic.ConfigStd.Marshal, unmarshal: stdUnmarshal(sonic.ConfigStd.Unmarshal),
	},
	{
		Config: report.Config{ID: "std/jsoniter", Library: "json-iterator/go", Category: "std", Title: "json-iterator/go",
			Setting: "jsoniter.ConfigCompatibleWithStandardLibrary"},
		module: jsoniterModule, marshal: jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
		unmarshal: stdUnmarshal(jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal),
	},
	{
		Config: report.Config{ID: "std/segmentio", Library: "segmentio/encoding", Category: "std", Title: "segmentio/encoding",
			Setting: "json.Marshal / json.Unmarshal"},
		module: segmentioModule, marshal: segmentio.Marshal, unmarshal: stdUnmarshal(segmentio.Unmarshal),
	},

	// fast
	{
		Config: report.Config{ID: "fast/encoding/json", Library: "encoding/json", Category: "fast", Title: "encoding/json",
			Setting: "json.Encoder with SetEscapeHTML( false ) / json.Unmarshal ( the keys of a map are always sorted and the strings copied )"},
		module: "std", marshal: stdMarshalNoHTMLEscape, unmarshal: stdUnmarshal(stdjson.Unmarshal),
	},
	{
		Config: report.Config{ID: "fast/go-json", Library: "goccy/go-json", Category: "fast", Title: "goccy/go-json",
			Setting: "json.MarshalWithOption( DisableHTMLEscape, UnorderedMap ) / json.UnmarshalWithOption( DecodeNoCopyString )"},
		module:    goJSONModule,
		marshal:   goJSONMarshal(gojson.DisableHTMLEscape(), gojson.UnorderedMap()),
		unmarshal: goJSONUnmarshal(gojson.DecodeNoCopyString()),
	},
	{
		Config: report.Config{ID: "fast/go-json/of", Library: "goccy/go-json", Category: "fast", Title: "goccy/go-json ( UnmarshalOf )",
			Setting: "json.UnmarshalOf( DecodeNoCopyString ), which takes the value by its type ( decode only )"},
		module: goJSONModule, decodeOf: []gojson.DecodeOptionFunc{gojson.DecodeNoCopyString()},
	},
	{
		Config: report.Config{ID: "fast/sonic", Library: "bytedance/sonic", Category: "fast", Title: "bytedance/sonic",
			Setting: "sonic.Config{ CompactMarshaler: true, ValidateString: true }; decoded from a string ( UnmarshalFromString )"},
		module: sonicModule, marshal: sonicFast.Marshal, unmarshal: sonicStringUnmarshal(sonicFast),
	},
	{
		Config: report.Config{ID: "fast/jsoniter", Library: "json-iterator/go", Category: "fast", Title: "json-iterator/go",
			Setting: "jsoniter.Config{ EscapeHTML: false, SortMapKeys: false, ValidateJsonRawMessage: true }"},
		module: jsoniterModule, marshal: jsoniterFast.Marshal, unmarshal: stdUnmarshal(jsoniterFast.Unmarshal),
	},
	{
		Config: report.Config{ID: "fast/segmentio", Library: "segmentio/encoding", Category: "fast", Title: "segmentio/encoding",
			Setting: "json.Append( 0 ) / json.Parse( ZeroCopy )"},
		module: segmentioModule, marshal: segmentioMarshal(0), unmarshal: segmentioUnmarshal(segmentio.ZeroCopy),
	},
}

// the configurations of "fastest"
func init() {
	reportConfigs = append(reportConfigs,
		&reportConfig{
			Config: report.Config{ID: "fastest/encoding/json", Library: "encoding/json", Category: "fastest", Title: "encoding/json",
				Setting: "json.Encoder with SetEscapeHTML( false ) / json.Unmarshal"},
			module: "std", marshal: stdMarshalNoHTMLEscape, unmarshal: stdUnmarshal(stdjson.Unmarshal),
		},
		&reportConfig{
			Config: report.Config{ID: "fastest/go-json", Library: "goccy/go-json", Category: "fastest", Title: "goccy/go-json",
				Setting: "json.MarshalWithOption( DisableHTMLEscape, DisableNormalizeUTF8, UnorderedMap ) / json.UnmarshalOf( DecodeNoCopyString )"},
			module:  goJSONModule,
			marshal: goJSONMarshal(gojson.DisableHTMLEscape(), gojson.DisableNormalizeUTF8(), gojson.UnorderedMap()),
			// DecodeFieldPriorityFirstWin is not set: it is faster only for an object whose fields are all decoded before its
			// end, and a little slower on most of these payloads.
			decodeOf: []gojson.DecodeOptionFunc{gojson.DecodeNoCopyString()},
		},
		&reportConfig{
			Config: report.Config{ID: "fastest/sonic", Library: "bytedance/sonic", Category: "fastest", Title: "bytedance/sonic",
				Setting: "sonic.ConfigFastest; decoded from a string ( UnmarshalFromString )"},
			module: sonicModule, marshal: sonic.ConfigFastest.Marshal, unmarshal: sonicStringUnmarshal(sonic.ConfigFastest),
		},
		&reportConfig{
			Config: report.Config{ID: "fastest/jsoniter", Library: "json-iterator/go", Category: "fastest", Title: "json-iterator/go",
				Setting: "jsoniter.ConfigFastest"},
			module: jsoniterModule, marshal: jsoniter.ConfigFastest.Marshal, unmarshal: stdUnmarshal(jsoniter.ConfigFastest.Unmarshal),
		},
		&reportConfig{
			Config: report.Config{ID: "fastest/segmentio", Library: "segmentio/encoding", Category: "fastest", Title: "segmentio/encoding",
				Setting: "json.Append( TrustRawMessage ) / json.Parse( ZeroCopy )"},
			module: segmentioModule, marshal: segmentioMarshal(segmentio.TrustRawMessage), unmarshal: segmentioUnmarshal(segmentio.ZeroCopy),
		},
	)
}

var (
	jsoniterFast = jsoniter.Config{EscapeHTML: false, SortMapKeys: false, ValidateJsonRawMessage: true}.Froze()
	sonicFast    = sonic.Config{CompactMarshaler: true, ValidateString: true}.Froze()
)

// insertAfter inserts the configuration after the one of the ID.
func insertAfter(id string, c *reportConfig) {
	for i, x := range reportConfigs {
		if x.ID == id {
			reportConfigs = append(reportConfigs[:i+1], append([]*reportConfig{c}, reportConfigs[i+1:]...)...)
			return
		}
	}
	panic("no configuration " + id)
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
		newReportPayload[MessagesResponse]("anthropic", "Anthropic Messages", "A response of the Anthropic Messages API, with text and tool use blocks.", anthropicResponseJSON),
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
	// BENCH_REPORT_ONLY measures only the configurations, operations and payloads whose "config/op/payload" it
	// matches, to look into a result.
	var only *regexp.Regexp
	if expr := os.Getenv("BENCH_REPORT_ONLY"); expr != "" {
		only = regexp.MustCompile(expr)
	}
	pretouchSonic()
	payloads := reportPayloads()
	run := &report.Run{
		GeneratedAt: time.Now().UTC(),
		GoVersion:   runtime.Version(),
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		CPU:         reportCPU(),
		CPUFeatures: reportCPUFeatures(),
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
		run.Probes = append(run.Probes, runProbes(c)...)
		for _, op := range []string{report.OpEncode, report.OpDecode} {
			if !c.does(op) {
				continue
			}
			for _, p := range payloads {
				diff, err := checkPayload(c, op, p)
				if err != nil {
					run.Exclusions = append(run.Exclusions, report.Exclusion{Op: op, Payload: p.ID, Config: c.ID, Reason: err.Error()})
					continue
				}
				if diff != "" {
					run.Differences = append(run.Differences, report.Exclusion{Op: op, Payload: p.ID, Config: c.ID, Reason: diff})
				}
				if only != nil && !only.MatchString(c.ID+"/"+op+"/"+p.ID) {
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

// baselineOf returns the baseline of the category.
func baselineOf(category string) *reportConfig {
	for _, cat := range reportCategories {
		if cat.ID != category {
			continue
		}
		for _, c := range reportConfigs {
			if c.ID == cat.Baseline {
				return c
			}
		}
	}
	panic("no baseline of " + category)
}

// checkPayload returns how the result of the configuration on the payload differs from the one of the baseline of
// its category, or "", and the error of the operation, which then can't be measured: a decoded value must be the
// one the baseline decodes; an encoded value must be the baseline's output byte for byte in "std", where the
// output is defined byte for byte, and the same JSON value in the other categories.
func checkPayload(c *reportConfig, op string, p *reportPayload) (string, error) {
	base := baselineOf(c.Category)
	if op == report.OpDecode {
		v := p.newValue()
		var err error
		if c.decodeOf != nil {
			// UnmarshalOf decodes the value as Unmarshal does: it is checked by a decode into the value.
			err = gojson.UnmarshalWithOption(p.data, v, c.decodeOf...)
		} else {
			err = c.unmarshal(p.data)(v)
		}
		if err != nil || c == base {
			return "", err
		}
		want := p.newValue()
		if err := base.unmarshal(p.data)(want); err != nil {
			return "", fmt.Errorf("the baseline fails: %w", err)
		}
		if !reflect.DeepEqual(v, want) {
			return "the decoded value differs from " + base.Title + "'s", nil
		}
		return "", nil
	}
	got, err := c.marshal(p.value)
	if err != nil || c == base {
		return "", err
	}
	want, err := base.marshal(p.value)
	if err != nil {
		return "", fmt.Errorf("the baseline fails: %w", err)
	}
	if c.Category == "std" {
		if !bytes.Equal(got, want) {
			return "the output differs from " + base.Title + "'s", nil
		}
	} else if !sameJSON(got, want) {
		return "the output is not the same JSON value as " + base.Title + "'s", nil
	}
	return "", nil
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

var (
	v1Like = []string{"std", "fast"}
	every  = []string{"std", "fast", "v2"}
)

type probeSlice struct {
	S []int
	M map[string]int
}

func stringProbe(want string) func(v any) string {
	return func(v any) string {
		if s := *v.(*string); s != want {
			return fmt.Sprintf("got %q", s)
		}
		return ""
	}
}

func nameProbe(want string) func(v any) string {
	return func(v any) string {
		if n := v.(*probeFields).Name; n != want {
			return fmt.Sprintf("got %q", n)
		}
		return ""
	}
}

var reportProbes = []reportProbe{
	{id: "escape-html", op: report.OpEncode, categories: []string{"std"},
		description: `HTML characters are escaped: "<&>" is encoded as "\u003c\u0026\u003e".`,
		check:       encodeProbe("<&>", `"\u003c\u0026\u003e"`, true)},
	{id: "no-escape-html", op: report.OpEncode, categories: []string{"v2"},
		description: `HTML characters are not escaped: "<&>" is encoded as "<&>".`,
		check:       encodeProbe("<&>", `"<&>"`, true)},
	{id: "sort-map-keys", op: report.OpEncode, categories: []string{"std"},
		description: "The keys of a map are sorted.",
		check:       encodeProbe(map[string]int{"b": 1, "c": 2, "a": 3}, `{"a":3,"b":1,"c":2}`, true)},
	{id: "invalid-utf8-encode", op: report.OpEncode, categories: v1Like,
		description: `Invalid UTF-8 is encoded as U+FFFD.`,
		check:       encodeProbe("a\xffb", `"a\ufffdb"`, false)},
	{id: "invalid-utf8-encode-error", op: report.OpEncode, categories: []string{"v2"},
		description: "Invalid UTF-8 is an error when it is encoded.",
		check: func(c *reportConfig) string {
			if out, err := c.marshal("a\xffb"); err == nil {
				return "encoded as " + string(out)
			}
			return ""
		}},
	{id: "nil-null", op: report.OpEncode, categories: v1Like,
		description: "A nil slice and a nil map are encoded as null.",
		check:       encodeProbe(probeSlice{}, `{"S":null,"M":null}`, false)},
	{id: "nil-empty", op: report.OpEncode, categories: []string{"v2"},
		description: "A nil slice is encoded as [] and a nil map as {}.",
		check:       encodeProbe(probeSlice{}, `{"S":[],"M":{}}`, false)},
	{id: "floats", op: report.OpEncode, categories: every,
		description: "Floats are formatted as encoding/json formats them.",
		check:       encodeProbe([]float64{1e21, 1e20, 1e-7, 0.000001, 0.1, 123456789, -0.5}, `[1e+21,100000000000000000000,1e-7,0.000001,0.1,123456789,-0.5]`, true)},
	{id: "marshaler-compact", op: report.OpEncode, categories: v1Like,
		description: "The output of a Marshaler is compacted.",
		check:       encodeProbe(probeMarshaler{}, `{"a":[1,2]}`, true)},
	{id: "invalid-utf8-decode", op: report.OpDecode, categories: v1Like,
		description: "Invalid UTF-8 in a string is decoded as U+FFFD.",
		check:       decodeProbe("\"a\xffb\"", func() any { return new(string) }, stringProbe("a\ufffdb"))},
	{id: "invalid-utf8-decode-error", op: report.OpDecode, categories: []string{"v2"},
		description: "Invalid UTF-8 in a string is an error.",
		check:       decodeProbe("\"a\xffb\"", func() any { return new(string) }, nil)},
	{id: "control-character", op: report.OpDecode, categories: every,
		description: "A control character in a string is an error.",
		check:       decodeProbe("\"a\x01\"", func() any { return new(string) }, nil)},
	{id: "case-insensitive", op: report.OpDecode, categories: v1Like,
		description: "A key matches a field regardless of case.",
		check:       decodeProbe(`{"NAME":"x"}`, func() any { return new(probeFields) }, nameProbe("x"))},
	{id: "case-sensitive", op: report.OpDecode, categories: []string{"v2"},
		description: "A key matches a field only by its case.",
		check:       decodeProbe(`{"NAME":"x"}`, func() any { return new(probeFields) }, nameProbe(""))},
	{id: "skipped-value-validated", op: report.OpDecode, categories: every,
		description: "The value of an unknown key is validated.",
		check:       decodeProbe(`{"unknown":[1,,2],"Name":"x"}`, func() any { return new(probeFields) }, nil)},
	{id: "duplicate-key", op: report.OpDecode, categories: v1Like,
		description: "The last of duplicate keys wins.",
		check:       decodeProbe(`{"Name":"a","Name":"b"}`, func() any { return new(probeFields) }, nameProbe("b"))},
	{id: "duplicate-key-error", op: report.OpDecode, categories: []string{"v2"},
		description: "Duplicate keys are an error.",
		check:       decodeProbe(`{"Name":"a","Name":"b"}`, func() any { return new(probeFields) }, nil)},
	{id: "overflow", op: report.OpDecode, categories: every,
		description: "A number out of the range of its field is an error.",
		check:       decodeProbe(`{"Small":300}`, func() any { return new(probeFields) }, nil)},
	{id: "string-copied", op: report.OpDecode, categories: []string{"std", "v2"},
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
		required, informational := false, false
		for _, cat := range p.categories {
			required = required || cat == c.Category
			// "fastest" requires nothing: the probes of encoding/json's behavior tell what each configuration skips
			informational = informational || c.Category == "fastest" && cat == "std"
		}
		if !required && !informational {
			// a probe of another category tells nothing of this one
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

// cpuFeatures are the extensions of the instruction set which the report shows, by their names in /proc/cpuinfo.
var cpuFeatures = []struct{ flag, name string }{
	{"avx2", "AVX2"}, {"bmi2", "BMI2"}, {"avx512f", "AVX-512F"}, {"avx512bw", "AVX-512BW"}, {"avx512vl", "AVX-512VL"},
	{"asimd", "NEON"}, {"sve", "SVE"}, {"sve2", "SVE2"},
}

// reportCPUFeatures returns the extensions of the instruction set of the CPU, from /proc/cpuinfo: the flags of an
// amd64 CPU and the features of an arm64 one.
func reportCPUFeatures() []string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil
	}
	has := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		k, v, ok := strings.Cut(line, ":")
		if k = strings.TrimSpace(k); !ok || k != "flags" && k != "Features" {
			continue
		}
		for _, f := range strings.Fields(v) {
			has[f] = true
		}
		break
	}
	var fs []string
	for _, f := range cpuFeatures {
		if has[f.flag] {
			fs = append(fs, f.name)
		}
	}
	return fs
}

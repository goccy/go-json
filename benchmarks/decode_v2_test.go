package benchmark

import (
	"io"
	"reflect"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	gojsontext "github.com/goccy/go-json/jsontext"
	gojsonv2 "github.com/goccy/go-json/v2"
)

// The decoding of github.com/goccy/go-json/v2, beside go-json's own of the same input.

func unmarshalV2(data []byte, v any) error { return gojsonv2.Unmarshal(data, v) }

func Benchmark_Decode_TwitterGeneric_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[any](b, []byte(TwitterJson), unmarshalV2)
}

func Benchmark_Decode_CodeAny_Unmarshal_GoJson(b *testing.B) {
	benchDecode[any](b, streamCodeJSON(), gojson.Unmarshal)
}

func Benchmark_Decode_CodeAny_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[any](b, streamCodeJSON(), unmarshalV2)
}

func Benchmark_Decode_LargeSlice_EscapedString_GoJsonV2(b *testing.B) {
	benchDecode[[]string](b, LargeSliceEscapedString, unmarshalV2)
}

// unmarshalV2Insensitive decodes as unmarshalV2 does, with the names matched regardless of their case, and of their
// dashes and underscores, which v1 doesn't ignore: by more than v1 matches them, into the structs of the payloads of
// fields without tags, whose objects have the names in other cases than the fields ( see smallPayloadTagged ).
func unmarshalV2Insensitive(data []byte, v any) error {
	return gojsonv2.Unmarshal(data, v, gojsonv2.MatchCaseInsensitiveNames(true))
}

func Benchmark_Decode_Code_Unmarshal_GoJson(b *testing.B) {
	benchDecode[codeResponse](b, streamCodeJSON(), gojson.Unmarshal)
}

func Benchmark_Decode_Code_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[codeResponse](b, streamCodeJSON(), unmarshalV2)
}

func Benchmark_Decode_TwitterBinding_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[TwitterStruct](b, []byte(TwitterJson), unmarshalV2)
}

// The payloads of fields without tags are decoded into the structs with the tags, without options,
// beside go-json's own into the ones without them ( see smallPayloadTagged ), and with MatchCaseInsensitiveNames
// into the ones without them, which matches the names by more than v1 does.

func Benchmark_Decode_SmallStruct_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[smallPayloadTagged](b, SmallFixture, unmarshalV2)
}

func Benchmark_Decode_MediumStruct_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[mediumPayloadTagged](b, MediumFixture, unmarshalV2)
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[largePayloadTagged](b, LargeFixture, unmarshalV2)
}

func Benchmark_Decode_SmallStruct_Unmarshal_GoJsonV2CaseInsensitive(b *testing.B) {
	benchDecode[SmallPayload](b, SmallFixture, unmarshalV2Insensitive)
}

func Benchmark_Decode_MediumStruct_Unmarshal_GoJsonV2CaseInsensitive(b *testing.B) {
	benchDecode[MediumPayload](b, MediumFixture, unmarshalV2Insensitive)
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJsonV2CaseInsensitive(b *testing.B) {
	benchDecode[LargePayload](b, LargeFixture, unmarshalV2Insensitive)
}

func Benchmark_Decode_InterfaceUnmarshaler_Unmarshal_GoJsonV2(b *testing.B) {
	data := ifaceDocument(`[1,2,3]`)
	v := newIfaceUnmarshalerPayload()
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := gojsonv2.Unmarshal(data, v); err != nil {
			b.Fatal(err)
		}
	}
}

// timesDocument is an array of 100 times of RFC 3339, with fractions of a second, in UTC and in another zone.
var timesDocument = func() []byte {
	b := []byte{'['}
	t := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	zone := time.FixedZone("", 9*60*60)
	for i := 0; i < 100; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		v := t.Add(time.Duration(i) * 1234567891)
		if i%2 == 1 {
			v = v.In(zone)
		}
		b = append(b, '"')
		b = v.AppendFormat(b, time.RFC3339Nano)
		b = append(b, '"')
	}
	return append(b, ']')
}()

func Benchmark_Decode_Times_Unmarshal_GoJson(b *testing.B) {
	benchDecode[[]time.Time](b, timesDocument, gojson.Unmarshal)
}

func Benchmark_Decode_Times_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[[]time.Time](b, timesDocument, unmarshalV2)
}

// TestGitHubPayloadsV2 checks that the v2 decoding of the payloads of the GitHub API gives the values of go-json's
// own: the times of the REST API by UnmarshalJSONFrom, and the fields of the GraphQL API, which have no tags, by
// their names of another case.
func TestGitHubPayloadsV2(t *testing.T) {
	var restWant, restGot []*GitHubIssue
	if err := gojson.Unmarshal(githubRESTIssues, &restWant); err != nil {
		t.Fatal(err)
	}
	if err := unmarshalV2(githubRESTIssues, &restGot); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restGot, restWant) {
		t.Fatal("the REST payload decodes differently")
	}
	var graphqlWant, graphqlGot GitHubGraphQLIssues
	if err := gojson.Unmarshal(githubGraphQLIssues, &graphqlWant); err != nil {
		t.Fatal(err)
	}
	if err := unmarshalV2Insensitive(githubGraphQLIssues, &graphqlGot); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(graphqlGot, graphqlWant) {
		t.Fatal("the GraphQL payload decodes differently")
	}
}

func Benchmark_Decode_GitHubREST_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[[]*GitHubIssue](b, githubRESTIssues, unmarshalV2)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[gitHubGraphQLIssuesTagged](b, githubGraphQLIssues, unmarshalV2)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_GoJsonV2CaseInsensitive(b *testing.B) {
	benchDecode[GitHubGraphQLIssues](b, githubGraphQLIssues, unmarshalV2Insensitive)
}

// v2ReadDecoder decodes the single value of its reader by UnmarshalRead of github.com/goccy/go-json/v2, with the
// options.
type v2ReadDecoder struct {
	r    io.Reader
	opts []gojsonv2.Options
}

func (d *v2ReadDecoder) Decode(v any) error { return gojsonv2.UnmarshalRead(d.r, v, d.opts...) }

func newGoJsonV2ReadDecoder(r io.Reader) streamDecoder { return &v2ReadDecoder{r: r} }

// newGoJsonV2ReadInsensitiveDecoder is newGoJsonV2ReadDecoder for the payloads whose fields have no tags ( see
// unmarshalV2Insensitive ).
func newGoJsonV2ReadInsensitiveDecoder(r io.Reader) streamDecoder {
	return &v2ReadDecoder{r: r, opts: []gojsonv2.Options{gojsonv2.MatchCaseInsensitiveNames(true)}}
}

func Benchmark_Decode_SmallStruct_ChunkedStream_GoJsonV2(b *testing.B) {
	benchStream(b, streamSmallTaggedPayload, newGoJsonV2ReadDecoder)
}

func Benchmark_Decode_MediumStruct_ChunkedStream_GoJsonV2(b *testing.B) {
	benchStream(b, streamMediumTaggedPayload, newGoJsonV2ReadDecoder)
}

func Benchmark_Decode_SmallStruct_ChunkedStream_GoJsonV2CaseInsensitive(b *testing.B) {
	benchStream(b, streamSmallPayload, newGoJsonV2ReadInsensitiveDecoder)
}

func Benchmark_Decode_MediumStruct_ChunkedStream_GoJsonV2CaseInsensitive(b *testing.B) {
	benchStream(b, streamMediumPayload, newGoJsonV2ReadInsensitiveDecoder)
}

func Benchmark_Decode_Twitter_ChunkedStream_GoJsonV2(b *testing.B) {
	benchStream(b, streamTwitterPayload, newGoJsonV2ReadDecoder)
}

func Benchmark_Decode_Code_ChunkedStream_GoJsonV2(b *testing.B) {
	benchStream(b, streamCodePayload, newGoJsonV2ReadDecoder)
}

func Benchmark_Decode_CodeAny_ChunkedStream_GoJsonV2(b *testing.B) {
	benchStream(b, streamCodeAnyPayload, newGoJsonV2ReadDecoder)
}

func Benchmark_Decode_Records_ChunkedStream_GoJsonV2(b *testing.B) {
	benchStream(b, streamRecordsPayload, newGoJsonV2ReadDecoder)
}

func Benchmark_Decode_SkipRecords_ChunkedStream_GoJsonV2(b *testing.B) {
	benchStream(b, streamSkipPayload, newGoJsonV2ReadDecoder)
}

// v2DecodeDecoder decodes the values of its decoder one after another by UnmarshalDecode of
// github.com/goccy/go-json/v2.
type v2DecodeDecoder struct{ dec *gojsontext.Decoder }

func (d *v2DecodeDecoder) Decode(v any) error { return gojsonv2.UnmarshalDecode(d.dec, v) }

func newGoJsonV2DecodeDecoder(r io.Reader) streamDecoder {
	return &v2DecodeDecoder{dec: gojsontext.NewDecoder(r)}
}

func Benchmark_Decode_NDJSON_ChunkedStream_GoJsonV2(b *testing.B) {
	benchStream(b, streamNDJSONPayload, newGoJsonV2DecodeDecoder)
}

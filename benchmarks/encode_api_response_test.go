package benchmark

import (
	"bytes"
	stdjson "encoding/json"
	"testing"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
)

// The benchmarks of encoding the responses of the APIs: the OpenAI Responses API and the Anthropic Messages API
// ( see llm_api_test.go ), whose strings are long and have escapes and a few characters which are not ASCII, and
// a page of issues of the GitHub REST API ( see decode_github_test.go ), whose times have MarshalJSON. They are the
// payloads of the benchmark report, encoded with the settings of its categories:
//
//   - GoJson and SonicStd: the same output as encoding/json.
//   - GoJsonValidateString and SonicValidateString: without HTML escaping and key sorting, but invalid UTF-8 is
//     still replaced and the output of a marshaler still validated ( the category "fast" of the report ).
//   - GoJsonLikeSonic, Sonic and SonicFastest: each library at its fastest.

var (
	openAIResponsesResponse = mustDecode[Response](openAIResponsesResponseJSON)
	anthropicResponse       = mustDecode[MessagesResponse](anthropicResponseJSON)
	githubRESTIssuesValue   = mustDecode[[]*GitHubIssue](githubRESTIssues)

	validateStringOptions = []gojson.EncodeOptionFunc{gojson.DisableHTMLEscape(), gojson.UnorderedMap()}
)

func marshalValidateString(v any) ([]byte, error) {
	return gojson.MarshalWithOption(v, validateStringOptions...)
}

func TestAPIResponsesEncode(t *testing.T) {
	// go-json encodes the responses to the same bytes as encoding/json.
	for _, v := range []any{openAIResponsesResponse, anthropicResponse, githubRESTIssuesValue} {
		want, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gojson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%T encodes differently", v)
		}
	}
}

func Benchmark_Encode_OpenAIResponse_EncodingJson(b *testing.B) {
	benchEncode(b, openAIResponsesResponse, stdjson.Marshal)
}

func Benchmark_Encode_OpenAIResponse_GoJson(b *testing.B) {
	benchEncode(b, openAIResponsesResponse, gojson.Marshal)
}

func Benchmark_Encode_OpenAIResponse_GoJsonValidateString(b *testing.B) {
	benchEncode(b, openAIResponsesResponse, marshalValidateString)
}

func Benchmark_Encode_OpenAIResponse_GoJsonLikeSonic(b *testing.B) {
	benchEncode(b, openAIResponsesResponse, marshalLikeSonic)
}

func Benchmark_Encode_OpenAIResponse_SonicStd(b *testing.B) {
	pretouchSonic()
	benchEncode(b, openAIResponsesResponse, sonic.ConfigStd.Marshal)
}

func Benchmark_Encode_OpenAIResponse_SonicValidateString(b *testing.B) {
	pretouchSonic()
	benchEncode(b, openAIResponsesResponse, sonicFast.Marshal)
}

func Benchmark_Encode_OpenAIResponse_Sonic(b *testing.B) {
	pretouchSonic()
	benchEncode(b, openAIResponsesResponse, sonic.ConfigDefault.Marshal)
}

func Benchmark_Encode_OpenAIResponse_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchEncode(b, openAIResponsesResponse, sonic.ConfigFastest.Marshal)
}

func Benchmark_Encode_AnthropicMessage_EncodingJson(b *testing.B) {
	benchEncode(b, anthropicResponse, stdjson.Marshal)
}

func Benchmark_Encode_AnthropicMessage_GoJson(b *testing.B) {
	benchEncode(b, anthropicResponse, gojson.Marshal)
}

func Benchmark_Encode_AnthropicMessage_GoJsonValidateString(b *testing.B) {
	benchEncode(b, anthropicResponse, marshalValidateString)
}

func Benchmark_Encode_AnthropicMessage_GoJsonLikeSonic(b *testing.B) {
	benchEncode(b, anthropicResponse, marshalLikeSonic)
}

func Benchmark_Encode_AnthropicMessage_SonicStd(b *testing.B) {
	pretouchSonic()
	benchEncode(b, anthropicResponse, sonic.ConfigStd.Marshal)
}

func Benchmark_Encode_AnthropicMessage_SonicValidateString(b *testing.B) {
	pretouchSonic()
	benchEncode(b, anthropicResponse, sonicFast.Marshal)
}

func Benchmark_Encode_AnthropicMessage_Sonic(b *testing.B) {
	pretouchSonic()
	benchEncode(b, anthropicResponse, sonic.ConfigDefault.Marshal)
}

func Benchmark_Encode_AnthropicMessage_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchEncode(b, anthropicResponse, sonic.ConfigFastest.Marshal)
}

func Benchmark_Encode_GitHubREST_EncodingJson(b *testing.B) {
	benchEncode(b, githubRESTIssuesValue, stdjson.Marshal)
}

func Benchmark_Encode_GitHubREST_GoJson(b *testing.B) {
	benchEncode(b, githubRESTIssuesValue, gojson.Marshal)
}

func Benchmark_Encode_GitHubREST_GoJsonValidateString(b *testing.B) {
	benchEncode(b, githubRESTIssuesValue, marshalValidateString)
}

func Benchmark_Encode_GitHubREST_GoJsonLikeSonic(b *testing.B) {
	benchEncode(b, githubRESTIssuesValue, marshalLikeSonic)
}

func Benchmark_Encode_GitHubREST_SonicStd(b *testing.B) {
	pretouchSonic()
	benchEncode(b, githubRESTIssuesValue, sonic.ConfigStd.Marshal)
}

func Benchmark_Encode_GitHubREST_SonicValidateString(b *testing.B) {
	pretouchSonic()
	benchEncode(b, githubRESTIssuesValue, sonicFast.Marshal)
}

func Benchmark_Encode_GitHubREST_Sonic(b *testing.B) {
	pretouchSonic()
	benchEncode(b, githubRESTIssuesValue, sonic.ConfigDefault.Marshal)
}

func Benchmark_Encode_GitHubREST_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchEncode(b, githubRESTIssuesValue, sonic.ConfigFastest.Marshal)
}

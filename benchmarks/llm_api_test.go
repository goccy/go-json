package benchmark

import (
	"bytes"
	stdjson "encoding/json"
	"reflect"
	"testing"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
)

// The benchmarks of the requests and the responses of the APIs of large language models, which coding agents
// exchange at every step: the OpenAI Chat Completions API ( /v1/chat/completions ), the OpenAI Responses API
// ( /v1/responses ) and the Anthropic Messages API ( /v1/messages ).
//
// The payloads are one session of a coding agent in the format of each API, generated after the references of
// the APIs ( see llm_payload_test.go ): a system prompt, tool definitions by JSON Schema, and turns of tool calls
// whose results are source files, then an answer with text and a tool call. Each API has:
//
//   - a request, whose encoding is measured ( Encode ): the conversation so far, sent at every step.
//   - a response, whose decoding is measured ( Decode ).
//   - a stream of the events of the same response, as server-sent events deliver them one JSON document at a
//     time: the decoding of every event is measured ( Decode ...Stream ). Most events carry a few bytes of
//     text, so a stream measures the cost of a call more than the one of a byte.
//
// The types are the ones Go programs decode these APIs into: the ones of github.com/sashabaranov/go-openai for
// the Chat Completions API, and types of the same kind for the others, with json.RawMessage for the input of
// a tool as in the Anthropic SDK.

// JSONSchema is a JSON Schema of the parameters of a tool, as jsonschema.Definition of go-openai.
type JSONSchema struct {
	Type                 string                `json:"type,omitempty"`
	Description          string                `json:"description,omitempty"`
	Enum                 []string              `json:"enum,omitempty"`
	Properties           map[string]JSONSchema `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	Items                *JSONSchema           `json:"items,omitempty"`
	AdditionalProperties any                   `json:"additionalProperties,omitempty"`
	Default              any                   `json:"default,omitempty"`
	Minimum              *float64              `json:"minimum,omitempty"`
	Maximum              *float64              `json:"maximum,omitempty"`
}

// ---------------------------------------------------------------- OpenAI Chat Completions

type ChatCompletionRequest struct {
	Model               string                  `json:"model"`
	Messages            []ChatCompletionMessage `json:"messages"`
	MaxCompletionTokens int                     `json:"max_completion_tokens,omitempty"`
	Temperature         float32                 `json:"temperature,omitempty"`
	Stream              bool                    `json:"stream,omitempty"`
	StreamOptions       *ChatStreamOptions      `json:"stream_options,omitempty"`
	Tools               []ChatTool              `json:"tools,omitempty"`
	ToolChoice          any                     `json:"tool_choice,omitempty"`
	ParallelToolCalls   any                     `json:"parallel_tool_calls,omitempty"`
	ReasoningEffort     string                  `json:"reasoning_effort,omitempty"`
	PromptCacheKey      string                  `json:"prompt_cache_key,omitempty"`
}

type ChatStreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type ChatTool struct {
	Type     string                  `json:"type"`
	Function *ChatFunctionDefinition `json:"function,omitempty"`
}

type ChatFunctionDefinition struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Strict      bool       `json:"strict,omitempty"`
	Parameters  JSONSchema `json:"parameters"`
}

type ChatCompletionMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	Refusal    string         `json:"refusal,omitempty"`
	Name       string         `json:"name,omitempty"`
	ToolCalls  []ChatToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type ChatToolCall struct {
	Index    *int             `json:"index,omitempty"`
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type"`
	Function ChatFunctionCall `json:"function"`
}

type ChatFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             ChatUsage              `json:"usage"`
	SystemFingerprint string                 `json:"system_fingerprint"`
	ServiceTier       string                 `json:"service_tier,omitempty"`
}

type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
	LogProbs     *ChatLogProbs         `json:"logprobs,omitempty"`
}

type ChatLogProbs struct {
	Content []ChatLogProb `json:"content"`
}

type ChatLogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"`
}

type ChatUsage struct {
	PromptTokens            int                          `json:"prompt_tokens"`
	CompletionTokens        int                          `json:"completion_tokens"`
	TotalTokens             int                          `json:"total_tokens"`
	PromptTokensDetails     *ChatPromptTokensDetails     `json:"prompt_tokens_details"`
	CompletionTokensDetails *ChatCompletionTokensDetails `json:"completion_tokens_details"`
}

type ChatPromptTokensDetails struct {
	AudioTokens  int `json:"audio_tokens"`
	CachedTokens int `json:"cached_tokens"`
}

type ChatCompletionTokensDetails struct {
	AudioTokens              int `json:"audio_tokens"`
	ReasoningTokens          int `json:"reasoning_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

type ChatCompletionStreamResponse struct {
	ID                string                       `json:"id"`
	Object            string                       `json:"object"`
	Created           int64                        `json:"created"`
	Model             string                       `json:"model"`
	Choices           []ChatCompletionStreamChoice `json:"choices"`
	SystemFingerprint string                       `json:"system_fingerprint"`
	ServiceTier       string                       `json:"service_tier,omitempty"`
	Usage             *ChatUsage                   `json:"usage,omitempty"`
}

type ChatCompletionStreamChoice struct {
	Index        int                             `json:"index"`
	Delta        ChatCompletionStreamChoiceDelta `json:"delta"`
	LogProbs     *ChatLogProbs                   `json:"logprobs,omitempty"`
	FinishReason string                          `json:"finish_reason"`
}

type ChatCompletionStreamChoiceDelta struct {
	Content   string         `json:"content,omitempty"`
	Role      string         `json:"role,omitempty"`
	ToolCalls []ChatToolCall `json:"tool_calls,omitempty"`
	Refusal   string         `json:"refusal,omitempty"`
}

// ---------------------------------------------------------------- OpenAI Responses

type ResponsesRequest struct {
	Model             string               `json:"model"`
	Instructions      string               `json:"instructions,omitempty"`
	Input             []ResponseInputItem  `json:"input"`
	Tools             []ResponseTool       `json:"tools,omitempty"`
	ToolChoice        string               `json:"tool_choice,omitempty"`
	ParallelToolCalls bool                 `json:"parallel_tool_calls,omitempty"`
	Reasoning         *ResponseReasoning   `json:"reasoning,omitempty"`
	Store             bool                 `json:"store"`
	Stream            bool                 `json:"stream,omitempty"`
	Include           []string             `json:"include,omitempty"`
	PromptCacheKey    string               `json:"prompt_cache_key,omitempty"`
	MaxOutputTokens   int                  `json:"max_output_tokens,omitempty"`
	Text              *ResponseTextOptions `json:"text,omitempty"`
}

type ResponseInputItem struct {
	Type      string            `json:"type"`
	ID        string            `json:"id,omitempty"`
	Status    string            `json:"status,omitempty"`
	Role      string            `json:"role,omitempty"`
	Content   []ResponseContent `json:"content,omitempty"`
	CallID    string            `json:"call_id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Arguments string            `json:"arguments,omitempty"`
	Output    string            `json:"output,omitempty"`
}

type ResponseContent struct {
	Type        string               `json:"type"`
	Text        string               `json:"text"`
	Annotations []ResponseAnnotation `json:"annotations,omitempty"`
	Logprobs    []ResponseLogprob    `json:"logprobs,omitempty"`
}

type ResponseAnnotation struct {
	Type       string `json:"type"`
	URL        string `json:"url,omitempty"`
	Title      string `json:"title,omitempty"`
	StartIndex int    `json:"start_index,omitempty"`
	EndIndex   int    `json:"end_index,omitempty"`
}

type ResponseLogprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
}

type ResponseTool struct {
	Type        string     `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Parameters  JSONSchema `json:"parameters"`
	Strict      bool       `json:"strict"`
}

type ResponseReasoning struct {
	Effort  string `json:"effort,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type ResponseTextOptions struct {
	Format    *ResponseTextFormat `json:"format,omitempty"`
	Verbosity string              `json:"verbosity,omitempty"`
}

type ResponseTextFormat struct {
	Type string `json:"type"`
}

type Response struct {
	ID                 string               `json:"id"`
	Object             string               `json:"object"`
	CreatedAt          int64                `json:"created_at"`
	Status             string               `json:"status"`
	Background         bool                 `json:"background"`
	Error              *ResponseError       `json:"error"`
	IncompleteDetails  *ResponseIncomplete  `json:"incomplete_details"`
	Instructions       string               `json:"instructions"`
	MaxOutputTokens    *int                 `json:"max_output_tokens"`
	MaxToolCalls       *int                 `json:"max_tool_calls"`
	Model              string               `json:"model"`
	Output             []ResponseOutputItem `json:"output"`
	ParallelToolCalls  bool                 `json:"parallel_tool_calls"`
	PreviousResponseID *string              `json:"previous_response_id"`
	PromptCacheKey     string               `json:"prompt_cache_key"`
	Reasoning          ResponseReasoning    `json:"reasoning"`
	SafetyIdentifier   *string              `json:"safety_identifier"`
	ServiceTier        string               `json:"service_tier"`
	Store              bool                 `json:"store"`
	Temperature        float64              `json:"temperature"`
	Text               ResponseTextOptions  `json:"text"`
	ToolChoice         any                  `json:"tool_choice"`
	Tools              []ResponseTool       `json:"tools"`
	TopLogprobs        int                  `json:"top_logprobs"`
	TopP               float64              `json:"top_p"`
	Truncation         string               `json:"truncation"`
	Usage              *ResponseUsage       `json:"usage"`
	User               *string              `json:"user"`
	Metadata           map[string]string    `json:"metadata"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResponseIncomplete struct {
	Reason string `json:"reason"`
}

type ResponseOutputItem struct {
	ID               string            `json:"id"`
	Type             string            `json:"type"`
	Status           string            `json:"status,omitempty"`
	Role             string            `json:"role,omitempty"`
	Content          []ResponseContent `json:"content,omitempty"`
	Summary          []ResponseSummary `json:"summary,omitempty"`
	EncryptedContent *string           `json:"encrypted_content,omitempty"`
	CallID           string            `json:"call_id,omitempty"`
	Name             string            `json:"name,omitempty"`
	Arguments        string            `json:"arguments,omitempty"`
}

type ResponseSummary struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ResponseUsage struct {
	InputTokens        int `json:"input_tokens"`
	InputTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"input_tokens_details"`
	OutputTokens        int `json:"output_tokens"`
	OutputTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
	TotalTokens int `json:"total_tokens"`
}

type ResponseStreamEvent struct {
	Type           string              `json:"type"`
	SequenceNumber int                 `json:"sequence_number"`
	Response       *Response           `json:"response,omitempty"`
	OutputIndex    int                 `json:"output_index"`
	ItemID         string              `json:"item_id,omitempty"`
	ContentIndex   int                 `json:"content_index"`
	SummaryIndex   int                 `json:"summary_index"`
	Item           *ResponseOutputItem `json:"item,omitempty"`
	Part           *ResponseContent    `json:"part,omitempty"`
	Delta          string              `json:"delta,omitempty"`
	Text           string              `json:"text,omitempty"`
	Arguments      string              `json:"arguments,omitempty"`
	Logprobs       []ResponseLogprob   `json:"logprobs,omitempty"`
	Obfuscation    string              `json:"obfuscation,omitempty"`
}

// ---------------------------------------------------------------- Anthropic Messages

type MessagesRequest struct {
	Model      string             `json:"model"`
	MaxTokens  int                `json:"max_tokens"`
	System     []AnthropicBlock   `json:"system,omitempty"`
	Messages   []AnthropicMessage `json:"messages"`
	Tools      []AnthropicTool    `json:"tools,omitempty"`
	ToolChoice *AnthropicChoice   `json:"tool_choice,omitempty"`
	Thinking   *AnthropicThinking `json:"thinking,omitempty"`
	Stream     bool               `json:"stream,omitempty"`
	Metadata   *AnthropicMetadata `json:"metadata,omitempty"`
}

type AnthropicMessage struct {
	Role    string           `json:"role"`
	Content []AnthropicBlock `json:"content"`
}

// AnthropicBlock is a block of content of any type: text, tool_use, tool_result, thinking.
type AnthropicBlock struct {
	Type         string              `json:"type"`
	Text         string              `json:"text,omitempty"`
	ID           string              `json:"id,omitempty"`
	Name         string              `json:"name,omitempty"`
	Input        stdjson.RawMessage  `json:"input,omitempty"`
	ToolUseID    string              `json:"tool_use_id,omitempty"`
	Content      string              `json:"content,omitempty"`
	IsError      bool                `json:"is_error,omitempty"`
	Thinking     string              `json:"thinking,omitempty"`
	Signature    string              `json:"signature,omitempty"`
	CacheControl *AnthropicCacheCtrl `json:"cache_control,omitempty"`
}

type AnthropicCacheCtrl struct {
	Type string `json:"type"`
}

type AnthropicTool struct {
	Name         string              `json:"name"`
	Description  string              `json:"description,omitempty"`
	InputSchema  JSONSchema          `json:"input_schema"`
	CacheControl *AnthropicCacheCtrl `json:"cache_control,omitempty"`
}

type AnthropicChoice struct {
	Type string `json:"type"`
}

type AnthropicThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

type AnthropicMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

type MessagesResponse struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	Role         string           `json:"role"`
	Model        string           `json:"model"`
	Content      []AnthropicBlock `json:"content"`
	StopReason   string           `json:"stop_reason"`
	StopSequence *string          `json:"stop_sequence"`
	Usage        AnthropicUsage   `json:"usage"`
}

type AnthropicUsage struct {
	InputTokens              int                     `json:"input_tokens"`
	CacheCreationInputTokens int                     `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int                     `json:"cache_read_input_tokens"`
	CacheCreation            *AnthropicCacheCreation `json:"cache_creation,omitempty"`
	OutputTokens             int                     `json:"output_tokens"`
	ServiceTier              string                  `json:"service_tier,omitempty"`
}

type AnthropicCacheCreation struct {
	Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens"`
	Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens"`
}

type MessageStreamEvent struct {
	Type         string            `json:"type"`
	Message      *MessagesResponse `json:"message,omitempty"`
	Index        int               `json:"index"`
	ContentBlock *AnthropicBlock   `json:"content_block,omitempty"`
	Delta        *MessageDelta     `json:"delta,omitempty"`
	Usage        *AnthropicUsage   `json:"usage,omitempty"`
}

type MessageDelta struct {
	Type         string  `json:"type,omitempty"`
	Text         string  `json:"text,omitempty"`
	PartialJSON  string  `json:"partial_json,omitempty"`
	Thinking     string  `json:"thinking,omitempty"`
	Signature    string  `json:"signature,omitempty"`
	StopReason   string  `json:"stop_reason,omitempty"`
	StopSequence *string `json:"stop_sequence,omitempty"`
}

// ---------------------------------------------------------------- payloads

var (
	openAIChatRequestJSON       = openAIChatRequestPayload()
	openAIChatResponseJSON      = openAIChatResponsePayload()
	openAIChatStream            = openAIChatStreamPayload()
	openAIResponsesRequestJSON  = openAIResponsesRequestPayload()
	openAIResponsesResponseJSON = openAIResponsesResponsePayload()
	openAIResponsesStream       = openAIResponsesStreamPayload()
	anthropicRequestJSON        = anthropicRequestPayload()
	anthropicResponseJSON       = anthropicResponsePayload()
	anthropicStream             = anthropicStreamPayload()

	openAIChatRequest      = mustDecode[ChatCompletionRequest](openAIChatRequestJSON)
	openAIResponsesRequest = mustDecode[ResponsesRequest](openAIResponsesRequestJSON)
	anthropicRequest       = mustDecode[MessagesRequest](anthropicRequestJSON)
)

func mustDecode[T any](data []byte) *T {
	var v T
	if err := stdjson.Unmarshal(data, &v); err != nil {
		panic(err)
	}
	return &v
}

func init() {
	for _, v := range []any{
		ChatCompletionRequest{}, ChatCompletionResponse{}, ChatCompletionStreamResponse{},
		ResponsesRequest{}, Response{}, ResponseStreamEvent{},
		MessagesRequest{}, MessagesResponse{}, MessageStreamEvent{},
	} {
		if err := sonic.Pretouch(reflect.TypeOf(v)); err != nil {
			panic(err)
		}
	}
}

func TestLLMPayloads(t *testing.T) {
	// go-json decodes the responses and the events as encoding/json does, and encodes the requests to the same
	// bytes.
	checkDecode[ChatCompletionResponse](t, openAIChatResponseJSON)
	checkDecode[Response](t, openAIResponsesResponseJSON)
	checkDecode[MessagesResponse](t, anthropicResponseJSON)
	for _, line := range openAIChatStream {
		checkDecode[ChatCompletionStreamResponse](t, line)
	}
	for _, line := range openAIResponsesStream {
		checkDecode[ResponseStreamEvent](t, line)
	}
	for _, line := range anthropicStream {
		checkDecode[MessageStreamEvent](t, line)
	}
	for _, v := range []any{openAIChatRequest, openAIResponsesRequest, anthropicRequest} {
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

func checkDecode[T any](t *testing.T, data []byte) {
	t.Helper()
	var want, got T
	if err := stdjson.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	if err := gojson.Unmarshal(data, &got); err != nil {
		t.Fatalf("%s: %v", data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s decodes differently:\n got %+v\nwant %+v", data, got, want)
	}
}

// benchDecodeStream decodes every event of a stream into a new value, as a client does with the data of the
// server-sent events.
func benchDecodeStream[T any](b *testing.B, events [][]byte, unmarshal func([]byte, any) error) {
	n := 0
	for _, e := range events {
		n += len(e)
	}
	b.SetBytes(int64(n))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, e := range events {
			var v T
			if err := unmarshal(e, &v); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func benchDecodeStreamOf[T any](b *testing.B, events [][]byte) {
	n := 0
	for _, e := range events {
		n += len(e)
	}
	b.SetBytes(int64(n))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, e := range events {
			var v T
			if err := gojson.UnmarshalOf(e, &v); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func benchEncode(b *testing.B, v any, marshal func(any) ([]byte, error)) {
	data, err := marshal(v)
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func marshalLikeSonic(v any) ([]byte, error) { return gojson.MarshalWithOption(v, likeSonicOptions...) }

// ---------------------------------------------------------------- OpenAI Chat Completions benchmarks

func Benchmark_Decode_OpenAIChatCompletion_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[ChatCompletionResponse](b, openAIChatResponseJSON, stdjson.Unmarshal)
}

func Benchmark_Decode_OpenAIChatCompletion_Unmarshal_GoJson(b *testing.B) {
	benchDecode[ChatCompletionResponse](b, openAIChatResponseJSON, gojson.Unmarshal)
}

func Benchmark_Decode_OpenAIChatCompletion_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[ChatCompletionResponse](b, openAIChatResponseJSON)
}

func Benchmark_Decode_OpenAIChatCompletion_Unmarshal_Sonic(b *testing.B) {
	benchDecode[ChatCompletionResponse](b, openAIChatResponseJSON, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_OpenAIChatCompletion_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[ChatCompletionResponse](b, openAIChatResponseJSON, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_OpenAIChatCompletionStream_Unmarshal_EncodingJson(b *testing.B) {
	benchDecodeStream[ChatCompletionStreamResponse](b, openAIChatStream, stdjson.Unmarshal)
}

func Benchmark_Decode_OpenAIChatCompletionStream_Unmarshal_GoJson(b *testing.B) {
	benchDecodeStream[ChatCompletionStreamResponse](b, openAIChatStream, gojson.Unmarshal)
}

func Benchmark_Decode_OpenAIChatCompletionStream_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeStreamOf[ChatCompletionStreamResponse](b, openAIChatStream)
}

func Benchmark_Decode_OpenAIChatCompletionStream_Unmarshal_Sonic(b *testing.B) {
	benchDecodeStream[ChatCompletionStreamResponse](b, openAIChatStream, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_OpenAIChatCompletionStream_Unmarshal_SonicStd(b *testing.B) {
	benchDecodeStream[ChatCompletionStreamResponse](b, openAIChatStream, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Encode_OpenAIChatCompletionRequest_EncodingJson(b *testing.B) {
	benchEncode(b, openAIChatRequest, stdjson.Marshal)
}

func Benchmark_Encode_OpenAIChatCompletionRequest_GoJson(b *testing.B) {
	benchEncode(b, openAIChatRequest, gojson.Marshal)
}

func Benchmark_Encode_OpenAIChatCompletionRequest_GoJsonLikeSonic(b *testing.B) {
	benchEncode(b, openAIChatRequest, marshalLikeSonic)
}

func Benchmark_Encode_OpenAIChatCompletionRequest_Sonic(b *testing.B) {
	benchEncode(b, openAIChatRequest, sonic.ConfigDefault.Marshal)
}

func Benchmark_Encode_OpenAIChatCompletionRequest_SonicStd(b *testing.B) {
	benchEncode(b, openAIChatRequest, sonic.ConfigStd.Marshal)
}

// ---------------------------------------------------------------- OpenAI Responses benchmarks

func Benchmark_Decode_OpenAIResponse_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[Response](b, openAIResponsesResponseJSON, stdjson.Unmarshal)
}

func Benchmark_Decode_OpenAIResponse_Unmarshal_GoJson(b *testing.B) {
	benchDecode[Response](b, openAIResponsesResponseJSON, gojson.Unmarshal)
}

func Benchmark_Decode_OpenAIResponse_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[Response](b, openAIResponsesResponseJSON)
}

func Benchmark_Decode_OpenAIResponse_Unmarshal_Sonic(b *testing.B) {
	benchDecode[Response](b, openAIResponsesResponseJSON, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_OpenAIResponse_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[Response](b, openAIResponsesResponseJSON, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_OpenAIResponseStream_Unmarshal_EncodingJson(b *testing.B) {
	benchDecodeStream[ResponseStreamEvent](b, openAIResponsesStream, stdjson.Unmarshal)
}

func Benchmark_Decode_OpenAIResponseStream_Unmarshal_GoJson(b *testing.B) {
	benchDecodeStream[ResponseStreamEvent](b, openAIResponsesStream, gojson.Unmarshal)
}

func Benchmark_Decode_OpenAIResponseStream_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeStreamOf[ResponseStreamEvent](b, openAIResponsesStream)
}

func Benchmark_Decode_OpenAIResponseStream_Unmarshal_Sonic(b *testing.B) {
	benchDecodeStream[ResponseStreamEvent](b, openAIResponsesStream, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_OpenAIResponseStream_Unmarshal_SonicStd(b *testing.B) {
	benchDecodeStream[ResponseStreamEvent](b, openAIResponsesStream, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Encode_OpenAIResponseRequest_EncodingJson(b *testing.B) {
	benchEncode(b, openAIResponsesRequest, stdjson.Marshal)
}

func Benchmark_Encode_OpenAIResponseRequest_GoJson(b *testing.B) {
	benchEncode(b, openAIResponsesRequest, gojson.Marshal)
}

func Benchmark_Encode_OpenAIResponseRequest_GoJsonLikeSonic(b *testing.B) {
	benchEncode(b, openAIResponsesRequest, marshalLikeSonic)
}

func Benchmark_Encode_OpenAIResponseRequest_Sonic(b *testing.B) {
	benchEncode(b, openAIResponsesRequest, sonic.ConfigDefault.Marshal)
}

func Benchmark_Encode_OpenAIResponseRequest_SonicStd(b *testing.B) {
	benchEncode(b, openAIResponsesRequest, sonic.ConfigStd.Marshal)
}

// ---------------------------------------------------------------- Anthropic Messages benchmarks

func Benchmark_Decode_AnthropicMessage_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[MessagesResponse](b, anthropicResponseJSON, stdjson.Unmarshal)
}

func Benchmark_Decode_AnthropicMessage_Unmarshal_GoJson(b *testing.B) {
	benchDecode[MessagesResponse](b, anthropicResponseJSON, gojson.Unmarshal)
}

func Benchmark_Decode_AnthropicMessage_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[MessagesResponse](b, anthropicResponseJSON)
}

func Benchmark_Decode_AnthropicMessage_Unmarshal_Sonic(b *testing.B) {
	benchDecode[MessagesResponse](b, anthropicResponseJSON, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_AnthropicMessage_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[MessagesResponse](b, anthropicResponseJSON, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_AnthropicMessageStream_Unmarshal_EncodingJson(b *testing.B) {
	benchDecodeStream[MessageStreamEvent](b, anthropicStream, stdjson.Unmarshal)
}

func Benchmark_Decode_AnthropicMessageStream_Unmarshal_GoJson(b *testing.B) {
	benchDecodeStream[MessageStreamEvent](b, anthropicStream, gojson.Unmarshal)
}

func Benchmark_Decode_AnthropicMessageStream_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeStreamOf[MessageStreamEvent](b, anthropicStream)
}

func Benchmark_Decode_AnthropicMessageStream_Unmarshal_Sonic(b *testing.B) {
	benchDecodeStream[MessageStreamEvent](b, anthropicStream, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_AnthropicMessageStream_Unmarshal_SonicStd(b *testing.B) {
	benchDecodeStream[MessageStreamEvent](b, anthropicStream, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Encode_AnthropicMessageRequest_EncodingJson(b *testing.B) {
	benchEncode(b, anthropicRequest, stdjson.Marshal)
}

func Benchmark_Encode_AnthropicMessageRequest_GoJson(b *testing.B) {
	benchEncode(b, anthropicRequest, gojson.Marshal)
}

func Benchmark_Encode_AnthropicMessageRequest_GoJsonLikeSonic(b *testing.B) {
	benchEncode(b, anthropicRequest, marshalLikeSonic)
}

func Benchmark_Encode_AnthropicMessageRequest_Sonic(b *testing.B) {
	benchEncode(b, anthropicRequest, sonic.ConfigDefault.Marshal)
}

func Benchmark_Encode_AnthropicMessageRequest_SonicStd(b *testing.B) {
	benchEncode(b, anthropicRequest, sonic.ConfigStd.Marshal)
}

// ---------------------------------------------------------------- the fastest configurations ( see sonic_decode_test.go )

// benchDecodeStreamSonicFastest decodes every event of a stream by sonic at its fastest, as strings.
func benchDecodeStreamSonicFastest[T any](b *testing.B, events [][]byte) {
	strs := make([]string, len(events))
	n := 0
	for i, e := range events {
		strs[i] = string(e)
		n += len(e)
	}
	b.SetBytes(int64(n))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, e := range strs {
			var v T
			if err := sonic.ConfigFastest.UnmarshalFromString(e, &v); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// benchDecodeStreamNoCopy decodes every event of a stream by go-json at its fastest.
func benchDecodeStreamNoCopy[T any](b *testing.B, events [][]byte) {
	n := 0
	for _, e := range events {
		n += len(e)
	}
	b.SetBytes(int64(n))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, e := range events {
			var v T
			if err := gojson.UnmarshalOf(e, &v, gojson.DecodeNoCopyString()); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func Benchmark_Decode_OpenAIChatCompletion_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[ChatCompletionResponse](b, openAIChatResponseJSON)
}

func Benchmark_Decode_OpenAIChatCompletion_Unmarshal_SonicFastest(b *testing.B) {
	benchDecodeSonicFastest[ChatCompletionResponse](b, openAIChatResponseJSON)
}

func Benchmark_Decode_OpenAIResponse_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[Response](b, openAIResponsesResponseJSON)
}

func Benchmark_Decode_OpenAIResponse_Unmarshal_SonicFastest(b *testing.B) {
	benchDecodeSonicFastest[Response](b, openAIResponsesResponseJSON)
}

func Benchmark_Decode_AnthropicMessage_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[MessagesResponse](b, anthropicResponseJSON)
}

func Benchmark_Decode_AnthropicMessage_Unmarshal_SonicFastest(b *testing.B) {
	benchDecodeSonicFastest[MessagesResponse](b, anthropicResponseJSON)
}

func Benchmark_Decode_OpenAIChatCompletionStream_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeStreamNoCopy[ChatCompletionStreamResponse](b, openAIChatStream)
}

func Benchmark_Decode_OpenAIChatCompletionStream_Unmarshal_SonicFastest(b *testing.B) {
	benchDecodeStreamSonicFastest[ChatCompletionStreamResponse](b, openAIChatStream)
}

func Benchmark_Decode_OpenAIResponseStream_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeStreamNoCopy[ResponseStreamEvent](b, openAIResponsesStream)
}

func Benchmark_Decode_OpenAIResponseStream_Unmarshal_SonicFastest(b *testing.B) {
	benchDecodeStreamSonicFastest[ResponseStreamEvent](b, openAIResponsesStream)
}

func Benchmark_Decode_AnthropicMessageStream_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeStreamNoCopy[MessageStreamEvent](b, anthropicStream)
}

func Benchmark_Decode_AnthropicMessageStream_Unmarshal_SonicFastest(b *testing.B) {
	benchDecodeStreamSonicFastest[MessageStreamEvent](b, anthropicStream)
}

func Benchmark_Encode_OpenAIChatCompletionRequest_SonicFastest(b *testing.B) {
	benchEncode(b, openAIChatRequest, sonic.ConfigFastest.Marshal)
}

func Benchmark_Encode_OpenAIResponseRequest_SonicFastest(b *testing.B) {
	benchEncode(b, openAIResponsesRequest, sonic.ConfigFastest.Marshal)
}

func Benchmark_Encode_AnthropicMessageRequest_SonicFastest(b *testing.B) {
	benchEncode(b, anthropicRequest, sonic.ConfigFastest.Marshal)
}

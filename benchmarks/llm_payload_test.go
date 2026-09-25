package benchmark

import (
	stdjson "encoding/json"
	"fmt"
	"strings"
)

// The payloads of the APIs of large language models ( see llm_api_test.go ): one session of a coding agent,
// rendered in the format of each API. The session is a system prompt, the definitions of five tools, a request
// of a user and five tool calls with their results, a search, a source file of 15 KB, an edit, the output of the
// tests and a directory; then the answer of the model: its reasoning, a text and a tool call.

// llmToolCall is a call of a tool, and its result.
type llmToolCall struct {
	id, name  string
	arguments string
	result    string
}

type llmSession struct {
	system    string
	tools     []llmTool
	request   string
	calls     []llmToolCall
	remark    string // a text of the model before a call
	followUp  string // a text of the user after the calls
	reasoning string
	summary   string
	signature string
	answer    string
	final     llmToolCall
}

type llmTool struct {
	name, description string
	parameters        JSONSchema
}

func float(v float64) *float64 { return &v }

func llmTools() []llmTool {
	str := func(description string) JSONSchema { return JSONSchema{Type: "string", Description: description} }
	integer := func(description string) JSONSchema { return JSONSchema{Type: "integer", Description: description} }
	boolean := func(description string) JSONSchema { return JSONSchema{Type: "boolean", Description: description} }
	params := func(required []string, properties map[string]JSONSchema) JSONSchema {
		return JSONSchema{Type: "object", Properties: properties, Required: required, AdditionalProperties: false}
	}
	timeout := integer("The timeout in milliseconds, up to 10 minutes")
	timeout.Minimum, timeout.Maximum = float(1), float(600000)
	replaceAll := boolean("Replace every occurrence instead of the first one")
	replaceAll.Default = false
	return []llmTool{
		{"read_file", "Reads a file of the repository. Lines are numbered from 1; a long file can be read by parts with offset and limit.",
			params([]string{"path"}, map[string]JSONSchema{
				"path":   str("The path of the file, relative to the root of the repository"),
				"offset": integer("The line to start reading from, from 1"),
				"limit":  integer("The number of lines to read"),
			})},
		{"edit_file", "Replaces a text of a file by another one. The text to replace must be unique in the file unless replace_all is set, and must match the file exactly, indentation included.",
			params([]string{"path", "old_string", "new_string"}, map[string]JSONSchema{
				"path":        str("The path of the file to edit"),
				"old_string":  str("The text to replace, which must be unique in the file"),
				"new_string":  str("The text to replace it with"),
				"replace_all": replaceAll,
			})},
		{"search", "Searches the files of the repository for a regular expression, and returns the matching lines with their paths and line numbers.",
			params([]string{"pattern"}, map[string]JSONSchema{
				"pattern":          str("A regular expression to search for"),
				"path":             str("The directory to search in"),
				"glob":             str("A glob of the files to search in, e.g. \"*.go\" or \"**/*.ts\""),
				"case_insensitive": boolean("Match regardless of case"),
			})},
		{"run_command", "Runs a shell command at the root of the repository and returns its output.",
			params([]string{"command"}, map[string]JSONSchema{
				"command":     str("The command to run"),
				"description": str("What the command does, in a few words"),
				"timeout":     timeout,
			})},
		{"list_files", "Lists the files of a directory.",
			params([]string{"path"}, map[string]JSONSchema{
				"path": str("The directory to list, relative to the root of the repository"),
			})},
	}
}

// callID returns an identifier of a tool call, as the OpenAI APIs make them.
func callID(r *payloadRand) string {
	return "call_" + r.token(alphanumeric, 24)
}

func newLLMSession() *llmSession {
	r := newPayloadRand(3)
	args := func(o object) string { return string(encodePayload(o)) }
	source := goSource(r, 15000)
	var numbered strings.Builder
	for i, line := range strings.SplitAfter(source, "\n") {
		fmt.Fprintf(&numbered, "%6d\t%s", 160+i, line)
	}
	var matches strings.Builder
	for i, line := range strings.Split(goSource(r, 500), "\n")[:8] {
		fmt.Fprintf(&matches, "internal/decoder/%s.go:%d:%s\n", r.pick(nouns), 40+i*37, line)
	}
	var tests strings.Builder
	for tests.Len() < 2000 {
		name := "TestDecode" + strings.ToUpper(identifier(r)[:1]) + identifier(r)[1:]
		fmt.Fprintf(&tests, "=== RUN   %s\n--- PASS: %s (0.%02ds)\n", name, name, r.intn(100))
	}
	tests.WriteString("PASS\nok  \tgithub.com/example/json-codec/internal/decoder\t0.412s\n")
	var files strings.Builder
	for i := 0; i < 24; i++ {
		fmt.Fprintf(&files, "%s_%s_test.go\n", r.pick(nouns), r.pick(nouns))
	}
	oldCode := goSource(r, 150)
	newCode := goSource(r, 430)
	s := &llmSession{
		system:  markdown(r, 1400),
		tools:   llmTools(),
		request: prose(r, 260),
		calls: []llmToolCall{
			{callID(r), "search", args(object{{"pattern", `func \(d \*structDecoder\) Decode`}, {"path", "internal/decoder"}, {"glob", "*.go"}}), matches.String()},
			{callID(r), "read_file", args(object{{"path", "internal/decoder/struct.go"}, {"offset", 160}, {"limit", 120}}), numbered.String()},
			{callID(r), "edit_file", args(object{{"path", "internal/decoder/struct.go"}, {"old_string", oldCode}, {"new_string", newCode}}), "The file internal/decoder/struct.go has been updated."},
			{callID(r), "run_command", args(object{{"command", "go test ./internal/decoder/... -run TestDecodeStructKeys -count=1"}, {"description", "Run the tests of the decoder"}}), tests.String()},
			{callID(r), "list_files", args(object{{"path", "benchmarks"}}), files.String()},
		},
		remark:    prose(r, 220),
		followUp:  prose(r, 76),
		reasoning: prose(r, 700),
		summary:   "**" + sentence(r) + "**\n\n" + prose(r, 120),
		signature: r.token(base64URL, 702),
		answer:    markdown(r, 950),
		final: llmToolCall{id: callID(r), name: "run_command", arguments: args(object{
			{"command", "cd benchmarks && go test -run '^$' -bench 'Benchmark_Decode_(Short|Long)Keys' -benchtime 300ms -count 3 ."},
			{"description", "Benchmark the decoding of long keys"},
		})},
	}
	return s
}

var llmSessionPayload = newLLMSession()

const (
	openAIModel        = "gpt-5.1-codex"
	openAIModelVersion = "gpt-5.1-codex-2025-11-13"
	anthropicModel     = "claude-sonnet-4-5"
	anthropicVersion   = "claude-sonnet-4-5-20250929"
	llmCreated         = 1790000000
	promptCacheKey     = "session-7f3c2a9e41d84b06"
)

// ---------------------------------------------------------------- OpenAI Chat Completions

func openAIChatRequestPayload() []byte {
	s := llmSessionPayload
	messages := []ChatCompletionMessage{{Role: "system", Content: s.system}, {Role: "user", Content: s.request}}
	for i, c := range s.calls {
		m := ChatCompletionMessage{Role: "assistant", ToolCalls: []ChatToolCall{{ID: c.id, Type: "function", Function: ChatFunctionCall{Name: c.name, Arguments: c.arguments}}}}
		if i == 2 {
			m.Content = s.remark
		}
		messages = append(messages, m, ChatCompletionMessage{Role: "tool", Content: c.result, ToolCallID: c.id})
	}
	messages = append(messages, ChatCompletionMessage{Role: "user", Content: s.followUp})
	var tools []ChatTool
	for _, t := range s.tools {
		tools = append(tools, ChatTool{Type: "function", Function: &ChatFunctionDefinition{Name: t.name, Description: t.description, Strict: true, Parameters: t.parameters}})
	}
	return encodePayload(ChatCompletionRequest{
		Model:               openAIModel,
		Messages:            messages,
		MaxCompletionTokens: 16384,
		Stream:              true,
		StreamOptions:       &ChatStreamOptions{IncludeUsage: true},
		Tools:               tools,
		ToolChoice:          "auto",
		ParallelToolCalls:   false,
		ReasoningEffort:     "medium",
		PromptCacheKey:      promptCacheKey,
	})
}

func openAIChatUsage() object {
	return object{
		{"prompt_tokens", 21874},
		{"completion_tokens", 1243},
		{"total_tokens", 23117},
		{"prompt_tokens_details", object{{"cached_tokens", 19200}, {"audio_tokens", 0}}},
		{"completion_tokens_details", object{{"reasoning_tokens", 384}, {"audio_tokens", 0}, {"accepted_prediction_tokens", 0}, {"rejected_prediction_tokens", 0}}},
	}
}

const openAIChatID = "chatcmpl-SXRVxfCQGgXkH1zxFUbEctT2NLLzP"

func openAIChatResponsePayload() []byte {
	s := llmSessionPayload
	return encodePayload(object{
		{"id", openAIChatID},
		{"object", "chat.completion"},
		{"created", llmCreated},
		{"model", openAIModelVersion},
		{"choices", []any{object{
			{"index", 0},
			{"message", object{
				{"role", "assistant"},
				{"content", s.answer},
				{"refusal", nil},
				{"annotations", []any{}},
				{"tool_calls", []any{object{
					{"id", s.final.id},
					{"type", "function"},
					{"function", object{{"name", s.final.name}, {"arguments", s.final.arguments}}},
				}}},
			}},
			{"logprobs", nil},
			{"finish_reason", "tool_calls"},
		}}},
		{"usage", openAIChatUsage()},
		{"service_tier", "default"},
		{"system_fingerprint", nil},
	})
}

func openAIChatStreamPayload() [][]byte {
	s := llmSessionPayload
	r := newPayloadRand(4)
	chunk := func(choices []any, usage any, obfuscate bool) object {
		o := object{
			{"id", openAIChatID},
			{"object", "chat.completion.chunk"},
			{"created", llmCreated},
			{"model", openAIModelVersion},
			{"service_tier", "default"},
			{"system_fingerprint", nil},
			{"choices", choices},
			{"usage", usage},
		}
		if obfuscate {
			o = append(o, field{"obfuscation", r.token(alphanumeric, r.between(1, 9))})
		}
		return o
	}
	delta := func(delta object, finish any) []any {
		return []any{object{{"index", 0}, {"delta", delta}, {"logprobs", nil}, {"finish_reason", finish}}}
	}
	events := []object{chunk(delta(object{{"role", "assistant"}, {"content", ""}, {"refusal", nil}}, nil), nil, true)}
	for _, piece := range chunks(r, s.answer) {
		events = append(events, chunk(delta(object{{"content", piece}}, nil), nil, true))
	}
	events = append(events, chunk(delta(object{{"tool_calls", []any{object{
		{"index", 0}, {"id", s.final.id}, {"type", "function"}, {"function", object{{"name", s.final.name}, {"arguments", ""}}},
	}}}}, nil), nil, false))
	for _, piece := range chunks(r, s.final.arguments) {
		events = append(events, chunk(delta(object{{"tool_calls", []any{object{{"index", 0}, {"function", object{{"arguments", piece}}}}}}}, nil), nil, true))
	}
	events = append(events,
		chunk(delta(object{}, "tool_calls"), nil, false),
		chunk([]any{}, openAIChatUsage(), true),
	)
	return jsonLines(events)
}

// ---------------------------------------------------------------- OpenAI Responses

func responseTools(s *llmSession) []ResponseTool {
	var tools []ResponseTool
	for _, t := range s.tools {
		tools = append(tools, ResponseTool{Type: "function", Name: t.name, Description: t.description, Parameters: t.parameters, Strict: true})
	}
	return tools
}

func openAIResponsesRequestPayload() []byte {
	s := llmSessionPayload
	r := newPayloadRand(5)
	input := []ResponseInputItem{{Type: "message", Role: "user", Content: []ResponseContent{{Type: "input_text", Text: s.request}}}}
	for i, c := range s.calls {
		if i == 2 {
			input = append(input, ResponseInputItem{Type: "message", ID: "msg_" + r.token(hexDigits, 48), Status: "completed", Role: "assistant",
				Content: []ResponseContent{{Type: "output_text", Text: s.remark, Annotations: []ResponseAnnotation{}, Logprobs: []ResponseLogprob{}}}})
		}
		input = append(input,
			ResponseInputItem{Type: "function_call", ID: "fc_" + r.token(hexDigits, 48), Status: "completed", CallID: c.id, Name: c.name, Arguments: c.arguments},
			ResponseInputItem{Type: "function_call_output", CallID: c.id, Output: c.result},
		)
	}
	input = append(input, ResponseInputItem{Type: "message", Role: "user", Content: []ResponseContent{{Type: "input_text", Text: s.followUp}}})
	return encodePayload(ResponsesRequest{
		Model:           openAIModel,
		Instructions:    s.system,
		Input:           input,
		Tools:           responseTools(s),
		ToolChoice:      "auto",
		Reasoning:       &ResponseReasoning{Effort: "medium", Summary: "auto"},
		Store:           false,
		Stream:          true,
		Include:         []string{"reasoning.encrypted_content"},
		PromptCacheKey:  promptCacheKey,
		MaxOutputTokens: 16384,
		Text:            &ResponseTextOptions{Verbosity: "medium"},
	})
}

// openAIResponse is a response of the Responses API, whose output and usage are the ones given.
func openAIResponse(s *llmSession, status string, output []any, usage any) object {
	return object{
		{"id", "resp_b9dba03eeb9caf3cc6086ed95e6b0cdca2f790d4c8520b8d"},
		{"object", "response"},
		{"created_at", llmCreated},
		{"status", status},
		{"background", false},
		{"billing", object{{"payer", "developer"}}},
		{"error", nil},
		{"incomplete_details", nil},
		{"instructions", s.system},
		{"max_output_tokens", 16384},
		{"max_tool_calls", nil},
		{"model", openAIModelVersion},
		{"output", output},
		{"parallel_tool_calls", false},
		{"previous_response_id", nil},
		{"prompt_cache_key", promptCacheKey},
		{"prompt_cache_retention", nil},
		{"reasoning", object{{"effort", "medium"}, {"summary", "detailed"}}},
		{"safety_identifier", nil},
		{"service_tier", "default"},
		{"store", false},
		{"temperature", 1.0},
		{"text", object{{"format", object{{"type", "text"}}}, {"verbosity", "medium"}}},
		{"tool_choice", "auto"},
		{"tools", responseTools(s)},
		{"top_logprobs", 0},
		{"top_p", 0.98},
		{"truncation", "disabled"},
		{"usage", usage},
		{"user", nil},
		{"metadata", object{}},
	}
}

const (
	responseReasoningID = "rs_94e8f5e183d2b2e0552c89667a822be1598b7cc5f8a7870c"
	responseMessageID   = "msg_3c1f0e8d9b2a47c6a5d4e3f2b1a09876c5d4e3f2a1b0c9d8"
	responseCallID      = "fc_7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f"
)

// openAIResponseOutput returns the items of the output: the reasoning, the message and the function call, with
// their texts as far as they are written.
func openAIResponseOutput(s *llmSession, encrypted, summary, answer, arguments string, done bool) (reasoning, message, call object) {
	status := "in_progress"
	if done {
		status = "completed"
	}
	summaries := []any{}
	if summary != "" {
		summaries = append(summaries, object{{"type", "summary_text"}, {"text", summary}})
	}
	content := []any{}
	if answer != "" {
		content = append(content, object{{"type", "output_text"}, {"annotations", []any{}}, {"logprobs", []any{}}, {"text", answer}})
	}
	reasoning = object{{"id", responseReasoningID}, {"type", "reasoning"}, {"encrypted_content", encrypted}, {"summary", summaries}}
	message = object{{"id", responseMessageID}, {"type", "message"}, {"status", status}, {"content", content}, {"role", "assistant"}}
	call = object{{"id", responseCallID}, {"type", "function_call"}, {"status", status}, {"arguments", arguments}, {"call_id", s.final.id}, {"name", s.final.name}}
	return reasoning, message, call
}

func openAIResponsesUsage() object {
	return object{
		{"input_tokens", 21874},
		{"input_tokens_details", object{{"cached_tokens", 19200}}},
		{"output_tokens", 1243},
		{"output_tokens_details", object{{"reasoning_tokens", 384}}},
		{"total_tokens", 23117},
	}
}

func encryptedReasoning(s *llmSession) string {
	return "gAAAAA" + s.signature + s.signature[:700]
}

func openAIResponsesResponsePayload() []byte {
	s := llmSessionPayload
	reasoning, message, call := openAIResponseOutput(s, encryptedReasoning(s), s.summary, s.answer, s.final.arguments, true)
	return encodePayload(openAIResponse(s, "completed", []any{reasoning, message, call}, openAIResponsesUsage()))
}

func openAIResponsesStreamPayload() [][]byte {
	s := llmSessionPayload
	r := newPayloadRand(6)
	var events []object
	event := func(typ string, fields ...field) {
		events = append(events, append(object{{"type", typ}, {"sequence_number", len(events)}}, fields...))
	}
	obfuscation := func() field { return field{"obfuscation", r.token(alphanumeric, r.between(1, 9))} }
	encrypted := encryptedReasoning(s)
	event("response.created", field{"response", openAIResponse(s, "in_progress", []any{}, nil)})
	event("response.in_progress", field{"response", openAIResponse(s, "in_progress", []any{}, nil)})

	reasoning, _, _ := openAIResponseOutput(s, encrypted, "", "", "", false)
	event("response.output_item.added", field{"output_index", 0}, field{"item", reasoning})
	summaryPart := func(text string) object { return object{{"type", "summary_text"}, {"text", text}} }
	event("response.reasoning_summary_part.added", field{"item_id", responseReasoningID}, field{"output_index", 0}, field{"summary_index", 0}, field{"part", summaryPart("")})
	for _, piece := range chunks(r, s.summary) {
		event("response.reasoning_summary_text.delta", field{"item_id", responseReasoningID}, field{"output_index", 0}, field{"summary_index", 0}, field{"delta", piece}, obfuscation())
	}
	event("response.reasoning_summary_text.done", field{"item_id", responseReasoningID}, field{"output_index", 0}, field{"summary_index", 0}, field{"text", s.summary})
	event("response.reasoning_summary_part.done", field{"item_id", responseReasoningID}, field{"output_index", 0}, field{"summary_index", 0}, field{"part", summaryPart(s.summary)})
	reasoning, _, _ = openAIResponseOutput(s, encrypted, s.summary, "", "", true)
	event("response.output_item.done", field{"output_index", 0}, field{"item", reasoning})

	_, message, _ := openAIResponseOutput(s, encrypted, "", "", "", false)
	event("response.output_item.added", field{"output_index", 1}, field{"item", message})
	textPart := func(text string) object {
		return object{{"type", "output_text"}, {"annotations", []any{}}, {"logprobs", []any{}}, {"text", text}}
	}
	event("response.content_part.added", field{"item_id", responseMessageID}, field{"output_index", 1}, field{"content_index", 0}, field{"part", textPart("")})
	for _, piece := range chunks(r, s.answer) {
		event("response.output_text.delta", field{"item_id", responseMessageID}, field{"output_index", 1}, field{"content_index", 0}, field{"delta", piece}, field{"logprobs", []any{}}, obfuscation())
	}
	event("response.output_text.done", field{"item_id", responseMessageID}, field{"output_index", 1}, field{"content_index", 0}, field{"text", s.answer}, field{"logprobs", []any{}})
	event("response.content_part.done", field{"item_id", responseMessageID}, field{"output_index", 1}, field{"content_index", 0}, field{"part", textPart(s.answer)})
	_, message, _ = openAIResponseOutput(s, encrypted, "", s.answer, "", true)
	event("response.output_item.done", field{"output_index", 1}, field{"item", message})

	_, _, call := openAIResponseOutput(s, encrypted, "", "", "", false)
	event("response.output_item.added", field{"output_index", 2}, field{"item", call})
	for _, piece := range chunks(r, s.final.arguments) {
		event("response.function_call_arguments.delta", field{"item_id", responseCallID}, field{"output_index", 2}, field{"delta", piece}, obfuscation())
	}
	event("response.function_call_arguments.done", field{"item_id", responseCallID}, field{"output_index", 2}, field{"arguments", s.final.arguments})
	_, _, call = openAIResponseOutput(s, encrypted, "", "", s.final.arguments, true)
	event("response.output_item.done", field{"output_index", 2}, field{"item", call})

	reasoning, message, call = openAIResponseOutput(s, encrypted, s.summary, s.answer, s.final.arguments, true)
	event("response.completed", field{"response", openAIResponse(s, "completed", []any{reasoning, message, call}, openAIResponsesUsage())})
	return jsonLines(events)
}

// ---------------------------------------------------------------- Anthropic Messages

func anthropicToolID(r *payloadRand) string {
	return "toolu_01" + r.token(alphanumeric, 22)
}

func anthropicRequestPayload() []byte {
	s := llmSessionPayload
	r := newPayloadRand(7)
	ephemeral := &AnthropicCacheCtrl{Type: "ephemeral"}
	messages := []AnthropicMessage{{Role: "user", Content: []AnthropicBlock{{Type: "text", Text: s.request}}}}
	for i, c := range s.calls {
		id := anthropicToolID(r)
		var assistant []AnthropicBlock
		if i == 2 {
			assistant = append(assistant, AnthropicBlock{Type: "text", Text: s.remark})
		}
		assistant = append(assistant, AnthropicBlock{Type: "tool_use", ID: id, Name: c.name, Input: stdjson.RawMessage(c.arguments)})
		user := []AnthropicBlock{{Type: "tool_result", ToolUseID: id, Content: c.result}}
		if i == len(s.calls)-1 {
			user = append(user, AnthropicBlock{Type: "text", Text: s.followUp, CacheControl: ephemeral})
		}
		messages = append(messages, AnthropicMessage{Role: "assistant", Content: assistant}, AnthropicMessage{Role: "user", Content: user})
	}
	var tools []AnthropicTool
	for _, t := range s.tools {
		tools = append(tools, AnthropicTool{Name: t.name, Description: t.description, InputSchema: t.parameters})
	}
	return encodePayload(MessagesRequest{
		Model:     anthropicModel,
		MaxTokens: 32000,
		System:    []AnthropicBlock{{Type: "text", Text: s.system, CacheControl: ephemeral}},
		Messages:  messages,
		Tools:     tools,
		Thinking:  &AnthropicThinking{Type: "enabled", BudgetTokens: 10000},
		Stream:    true,
		Metadata:  &AnthropicMetadata{UserID: "user_" + r.token(hexDigits, 32)},
	})
}

const anthropicMessageID = "msg_010ZX2h3zvgCT7KHTPEQrjBx"

var anthropicFinalToolID = anthropicToolID(newPayloadRand(8))

func anthropicUsage(outputTokens int) object {
	return object{
		{"input_tokens", 312},
		{"cache_creation_input_tokens", 2362},
		{"cache_read_input_tokens", 19200},
		{"cache_creation", object{{"ephemeral_5m_input_tokens", 2362}, {"ephemeral_1h_input_tokens", 0}}},
		{"output_tokens", outputTokens},
		{"service_tier", "standard"},
	}
}

func anthropicResponsePayload() []byte {
	s := llmSessionPayload
	return encodePayload(object{
		{"id", anthropicMessageID},
		{"type", "message"},
		{"role", "assistant"},
		{"model", anthropicVersion},
		{"content", []any{
			object{{"type", "thinking"}, {"thinking", s.reasoning}, {"signature", s.signature}},
			object{{"type", "text"}, {"text", s.answer}},
			object{{"type", "tool_use"}, {"id", anthropicFinalToolID}, {"name", s.final.name}, {"input", stdjson.RawMessage(s.final.arguments)}},
		}},
		{"stop_reason", "tool_use"},
		{"stop_sequence", nil},
		{"usage", anthropicUsage(1243)},
	})
}

func anthropicStreamPayload() [][]byte {
	s := llmSessionPayload
	r := newPayloadRand(9)
	events := []object{{
		{"type", "message_start"},
		{"message", object{
			{"id", anthropicMessageID},
			{"type", "message"},
			{"role", "assistant"},
			{"model", anthropicVersion},
			{"content", []any{}},
			{"stop_reason", nil},
			{"stop_sequence", nil},
			{"usage", anthropicUsage(1)},
		}},
	}}
	block := func(index int, start object, deltaType, key, text string, extra ...object) {
		events = append(events, object{{"type", "content_block_start"}, {"index", index}, {"content_block", start}})
		if index == 0 {
			events = append(events, object{{"type", "ping"}})
		}
		for _, piece := range chunks(r, text) {
			events = append(events, object{{"type", "content_block_delta"}, {"index", index}, {"delta", object{{"type", deltaType}, {key, piece}}}})
		}
		for _, delta := range extra {
			events = append(events, object{{"type", "content_block_delta"}, {"index", index}, {"delta", delta}})
		}
		events = append(events, object{{"type", "content_block_stop"}, {"index", index}})
	}
	block(0, object{{"type", "thinking"}, {"thinking", ""}, {"signature", ""}}, "thinking_delta", "thinking", s.reasoning,
		object{{"type", "signature_delta"}, {"signature", s.signature}})
	block(1, object{{"type", "text"}, {"text", ""}}, "text_delta", "text", s.answer)
	block(2, object{{"type", "tool_use"}, {"id", anthropicFinalToolID}, {"name", s.final.name}, {"input", object{}}}, "input_json_delta", "partial_json", s.final.arguments)
	events = append(events,
		object{
			{"type", "message_delta"},
			{"delta", object{{"stop_reason", "tool_use"}, {"stop_sequence", nil}}},
			{"usage", object{{"input_tokens", 312}, {"cache_creation_input_tokens", 2362}, {"cache_read_input_tokens", 19200}, {"output_tokens", 1243}}},
		},
		object{{"type", "message_stop"}},
	)
	return jsonLines(events)
}

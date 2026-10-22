package benchmark

import (
	"testing"

	gojson "github.com/goccy/go-json"
	gojsonv2 "github.com/goccy/go-json/v2"
)

// The decoding of github.com/goccy/go-json/v2, beside go-json's own of the same input: the values which its
// Unmarshal decodes so far, without methods of unmarshaling.

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

// unmarshalV2Insensitive decodes as unmarshalV2 does, with the names matched without regard to their case, as v1
// matches them: the fields of the payloads of these benchmarks have no tags, and their objects lowercase names.
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

func Benchmark_Decode_SmallStruct_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[SmallPayload](b, SmallFixture, unmarshalV2Insensitive)
}

func Benchmark_Decode_MediumStruct_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[MediumPayload](b, MediumFixture, unmarshalV2Insensitive)
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJsonV2(b *testing.B) {
	benchDecode[LargePayload](b, LargeFixture, unmarshalV2Insensitive)
}

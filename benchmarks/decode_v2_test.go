package benchmark

import (
	"testing"

	gojson "github.com/goccy/go-json"
	gojsonv2 "github.com/goccy/go-json/v2"
)

// The decoding of github.com/goccy/go-json/v2, beside go-json's own of the same input: the values which its
// Unmarshal decodes so far, interface{} and slices.

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

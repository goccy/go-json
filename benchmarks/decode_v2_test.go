package benchmark

import (
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	gojsonv2 "github.com/goccy/go-json/v2"
)

// The decoding of github.com/goccy/go-json/v2, beside go-json's own of the same input: the values which its
// Unmarshal decodes so far, without UnmarshalJSONFrom.

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

package benchmark

import (
	stdjson "encoding/json"
	"testing"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
)

// The decode benchmarks of go-json beside sonic, on the payloads of the other decode benchmarks and on
// the Twitter payload of sonic ( sonic_bench_test.go ), decoded into its struct ( Binding ) and into
// interface{} ( Generic ). Sonic is sonic.ConfigDefault and SonicStd is sonic.ConfigStd, which validates
// the strings as encoding/json does; the types are compiled before the benchmarks ( sonic_test.go ).

func benchDecode[T any](b *testing.B, data []byte, unmarshal func([]byte, any) error) {
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v T
		if err := unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_Sonic(b *testing.B) {
	benchDecode[SmallPayload](b, SmallFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_SmallStruct_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[SmallPayload](b, SmallFixture, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_MediumStruct_Unmarshal_Sonic(b *testing.B) {
	benchDecode[MediumPayload](b, MediumFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_MediumStruct_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[MediumPayload](b, MediumFixture, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_LargeStruct_Unmarshal_Sonic(b *testing.B) {
	benchDecode[LargePayload](b, LargeFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_LargeStruct_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[LargePayload](b, LargeFixture, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_TwitterBinding_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[TwitterStruct](b, []byte(TwitterJson), stdjson.Unmarshal)
}

func Benchmark_Decode_TwitterBinding_Unmarshal_GoJson(b *testing.B) {
	benchDecode[TwitterStruct](b, []byte(TwitterJson), gojson.Unmarshal)
}

func Benchmark_Decode_TwitterBinding_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	data := []byte(TwitterJson)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v TwitterStruct
		if err := gojson.UnmarshalOf(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

// GoJsonLikeSonic is go-json with the option which does what sonic does by default:
// the strings refer to the input instead of copies.
func Benchmark_Decode_TwitterBinding_Unmarshal_GoJsonLikeSonic(b *testing.B) {
	benchDecode[TwitterStruct](b, []byte(TwitterJson), func(data []byte, v any) error {
		return gojson.UnmarshalWithOption(data, v, gojson.DecodeNoCopyString())
	})
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_GoJsonLikeSonic(b *testing.B) {
	benchDecode[any](b, []byte(TwitterJson), func(data []byte, v any) error {
		return gojson.UnmarshalWithOption(data, v, gojson.DecodeNoCopyString())
	})
}

func Benchmark_Decode_TwitterBinding_Unmarshal_Sonic(b *testing.B) {
	benchDecode[TwitterStruct](b, []byte(TwitterJson), sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_TwitterBinding_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[TwitterStruct](b, []byte(TwitterJson), sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[any](b, []byte(TwitterJson), stdjson.Unmarshal)
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_GoJson(b *testing.B) {
	benchDecode[any](b, []byte(TwitterJson), gojson.Unmarshal)
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_Sonic(b *testing.B) {
	benchDecode[any](b, []byte(TwitterJson), sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[any](b, []byte(TwitterJson), sonic.ConfigStd.Unmarshal)
}

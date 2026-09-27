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
//
// SonicFastest is sonic at its fastest: sonic.ConfigFastest, which skips the values of no field without
// validating them, decoding the input as a string ( UnmarshalFromString ), which sonic doesn't copy, while
// Unmarshal copies the bytes it is given into a string first. Its strings refer to the input, as the ones of
// sonic do by default. GoJsonUnmarshalOfNoCopyString is go-json at its fastest: UnmarshalOf, with the strings
// referring to the input ( DecodeNoCopyString ). SonicFastestValidating is SonicFastest with the checks of
// go-json ( see sonicFastestValidating ), which is the one to compare it with.
//
// Sonic has two decoders on amd64, the JIT one and the one of SONIC_USE_OPTDEC=1, which the environment chooses
// for the process: the comparison is run with both ( see the CI ).

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
	pretouchSonic()
	benchDecode[SmallPayload](b, SmallFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_SmallStruct_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
	benchDecode[SmallPayload](b, SmallFixture, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_MediumStruct_Unmarshal_Sonic(b *testing.B) {
	pretouchSonic()
	benchDecode[MediumPayload](b, MediumFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_MediumStruct_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
	benchDecode[MediumPayload](b, MediumFixture, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_LargeStruct_Unmarshal_Sonic(b *testing.B) {
	pretouchSonic()
	benchDecode[LargePayload](b, LargeFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_LargeStruct_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
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
	pretouchSonic()
	benchDecode[TwitterStruct](b, []byte(TwitterJson), sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_TwitterBinding_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
	benchDecode[TwitterStruct](b, []byte(TwitterJson), sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[any](b, []byte(TwitterJson), stdjson.Unmarshal)
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_GoJson(b *testing.B) {
	benchDecode[any](b, []byte(TwitterJson), gojson.Unmarshal)
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_Sonic(b *testing.B) {
	pretouchSonic()
	benchDecode[any](b, []byte(TwitterJson), sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
	benchDecode[any](b, []byte(TwitterJson), sonic.ConfigStd.Unmarshal)
}

// benchDecodeSonicFastest decodes the data by sonic at its fastest: the data is made a string once, outside of
// the loop, as a program which has the input as a string does.
func benchDecodeSonicFastest[T any](b *testing.B, data []byte) {
	benchDecodeSonicString[T](b, sonic.ConfigFastest, data)
}

// sonicFastestValidating is sonic at its fastest with the checks which encoding/json and go-json do: the values
// which are skipped are validated ( NoValidateJSONSkip is not set ), and so are the strings ( ValidateString: a
// control character is an error, and invalid UTF-8 is replaced ). Its strings refer to the input, as the ones of
// GoJsonUnmarshalOfNoCopyString, which does the same checks.
var sonicFastestValidating = sonic.Config{ValidateString: true}.Froze()

// benchDecodeSonicFastestValidating decodes the data as benchDecodeSonicFastest does, with the checks of go-json.
func benchDecodeSonicFastestValidating[T any](b *testing.B, data []byte) {
	benchDecodeSonicString[T](b, sonicFastestValidating, data)
}

// benchDecodeSonicString decodes the data by api from a string, which is made once, outside of the loop.
func benchDecodeSonicString[T any](b *testing.B, api sonic.API, data []byte) {
	s := string(data)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v T
		if err := api.UnmarshalFromString(s, &v); err != nil {
			b.Fatal(err)
		}
	}
}

// benchDecodeNoCopy decodes the data by go-json at its fastest.
func benchDecodeNoCopy[T any](b *testing.B, data []byte) {
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v T
		if err := gojson.UnmarshalOf(data, &v, gojson.DecodeNoCopyString()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[SmallPayload](b, SmallFixture)
}

func Benchmark_Decode_SmallStruct_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[SmallPayload](b, SmallFixture)
}

func Benchmark_Decode_MediumStruct_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[MediumPayload](b, MediumFixture)
}

func Benchmark_Decode_MediumStruct_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[MediumPayload](b, MediumFixture)
}

func Benchmark_Decode_LargeStruct_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[LargePayload](b, LargeFixture)
}

func Benchmark_Decode_LargeStruct_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[LargePayload](b, LargeFixture)
}

func Benchmark_Decode_TwitterBinding_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[TwitterStruct](b, []byte(TwitterJson))
}

func Benchmark_Decode_TwitterBinding_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[TwitterStruct](b, []byte(TwitterJson))
}

func Benchmark_Decode_TwitterBinding_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[TwitterStruct](b, []byte(TwitterJson))
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[any](b, []byte(TwitterJson))
}

func Benchmark_Decode_TwitterGeneric_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[any](b, []byte(TwitterJson))
}

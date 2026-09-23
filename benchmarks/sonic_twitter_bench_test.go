package benchmark

import (
	stdjson "encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/option"
	gojson "github.com/goccy/go-json"
)

// The benchmarks of the encoder of bytedance/sonic ( encoder/encoder_test.go ), with go-json beside it:
// the twitter payload ( sonic_twitter_test.go ) as a value of interface{} ( Generic ) and as its struct
// ( Binding ), one goroutine and in parallel.
//
//   - Sonic is what sonic's BenchmarkEncoder_*_Sonic does: SortMapKeys, EscapeHTML and CompactMarshaler,
//     which is what go-json does by default, but for the normalization of UTF-8, which go-json also does by
//     default: GoJsonNoNormalize is go-json without it.
//   - SonicFast is what BenchmarkEncoder_*_Sonic_Fast does: no option. GoJsonLikeSonicFast is go-json with
//     the options which do the same: no sort of the keys, no escape of HTML, no normalization of UTF-8.
//   - StdLib is encoding/json, as in sonic.

var (
	twitterGeneric     interface{}
	twitterBinding     TwitterStruct
	sonicTwitter       = sonic.Config{SortMapKeys: true, EscapeHTML: true, CompactMarshaler: true}.Froze()
	sonicTwitterFast   = sonic.Config{}.Froze()
	noNormalizeOptions = []gojson.EncodeOptionFunc{gojson.DisableNormalizeUTF8()}
)

func init() {
	if err := stdjson.Unmarshal([]byte(TwitterJson), &twitterGeneric); err != nil {
		panic(err)
	}
	if err := stdjson.Unmarshal([]byte(TwitterJson), &twitterBinding); err != nil {
		panic(err)
	}
	opts := []option.CompileOption{option.WithCompileRecursiveDepth(10)}
	if depth, err := strconv.Atoi(os.Getenv("SONIC_MAX_INLINE_DEPTH")); err == nil {
		opts = append(opts, option.WithCompileMaxInlineDepth(depth))
	}
	if err := sonic.Pretouch(reflect.TypeOf(&twitterBinding), opts...); err != nil {
		panic(err)
	}
}

func benchTwitter(b *testing.B, marshal func() ([]byte, error)) {
	if _, err := marshal(); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(TwitterJson)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := marshal(); err != nil {
			b.Fatal(err)
		}
	}
}

func benchTwitterParallel(b *testing.B, marshal func() ([]byte, error)) {
	if _, err := marshal(); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(TwitterJson)))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := marshal(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_TwitterGeneric_Sonic(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return sonicTwitter.Marshal(twitterGeneric) })
}

func Benchmark_TwitterGeneric_SonicFast(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return sonicTwitterFast.Marshal(twitterGeneric) })
}

func Benchmark_TwitterGeneric_StdLib(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return stdjson.Marshal(twitterGeneric) })
}

func Benchmark_TwitterGeneric_GoJson(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.Marshal(twitterGeneric) })
}

func Benchmark_TwitterGeneric_GoJsonNoNormalize(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.MarshalWithOption(twitterGeneric, noNormalizeOptions...) })
}

func Benchmark_TwitterGeneric_GoJsonLikeSonicFast(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.MarshalWithOption(twitterGeneric, likeSonicOptions...) })
}

func Benchmark_TwitterBinding_Sonic(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return sonicTwitter.Marshal(&twitterBinding) })
}

func Benchmark_TwitterBinding_SonicFast(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return sonicTwitterFast.Marshal(&twitterBinding) })
}

func Benchmark_TwitterBinding_StdLib(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return stdjson.Marshal(&twitterBinding) })
}

func Benchmark_TwitterBinding_GoJson(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.Marshal(&twitterBinding) })
}

func Benchmark_TwitterBinding_GoJsonNoNormalize(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.MarshalWithOption(&twitterBinding, noNormalizeOptions...) })
}

func Benchmark_TwitterBinding_GoJsonLikeSonicFast(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.MarshalWithOption(&twitterBinding, likeSonicOptions...) })
}

func Benchmark_TwitterParallelGeneric_Sonic(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return sonicTwitter.Marshal(twitterGeneric) })
}

func Benchmark_TwitterParallelGeneric_SonicFast(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return sonicTwitterFast.Marshal(twitterGeneric) })
}

func Benchmark_TwitterParallelGeneric_StdLib(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return stdjson.Marshal(twitterGeneric) })
}

func Benchmark_TwitterParallelGeneric_GoJson(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return gojson.Marshal(twitterGeneric) })
}

func Benchmark_TwitterParallelGeneric_GoJsonLikeSonicFast(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return gojson.MarshalWithOption(twitterGeneric, likeSonicOptions...) })
}

func Benchmark_TwitterParallelBinding_Sonic(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return sonicTwitter.Marshal(&twitterBinding) })
}

func Benchmark_TwitterParallelBinding_SonicFast(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return sonicTwitterFast.Marshal(&twitterBinding) })
}

func Benchmark_TwitterParallelBinding_StdLib(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return stdjson.Marshal(&twitterBinding) })
}

func Benchmark_TwitterParallelBinding_GoJson(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return gojson.Marshal(&twitterBinding) })
}

func Benchmark_TwitterParallelBinding_GoJsonLikeSonicFast(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return gojson.MarshalWithOption(&twitterBinding, likeSonicOptions...) })
}

// what go-json writes for the payload is what encoding/json writes, and what sonic writes with the options
// of the benchmark.
func TestTwitterPayload(t *testing.T) {
	for _, v := range []interface{}{twitterGeneric, &twitterBinding} {
		expected, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gojson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(expected) {
			t.Fatalf("%T: go-json differs from encoding/json", v)
		}
		got, err = sonicTwitter.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(expected) {
			t.Logf("%T: sonic differs from encoding/json at byte %d", v, firstDifference(got, expected))
		}
	}
}

func firstDifference(a, b []byte) int {
	for i := range a {
		if i >= len(b) || a[i] != b[i] {
			return i
		}
	}
	return len(a)
}

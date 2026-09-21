package benchmark

import (
	"bytes"
	"io"
	"testing"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
)

// sonic.ConfigDefault doesn't escape HTML and doesn't sort the keys of a map, which go-json and encoding/json do.
// sonic.ConfigStd does what encoding/json does, so it is the one to compare go-json with.

func Benchmark_Encode_SmallStruct_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(NewSmallPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStruct_SonicStd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(NewSmallPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_Sonic(b *testing.B) {
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_SonicStd(b *testing.B) {
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStruct_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(NewMediumPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStruct_SonicStd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(NewMediumPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_Sonic(b *testing.B) {
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_SonicStd(b *testing.B) {
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStruct_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(NewLargePayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStruct_SonicStd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(NewLargePayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_Sonic(b *testing.B) {
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_SonicStd(b *testing.B) {
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MapInterface_Sonic(b *testing.B) {
	v := benchMapValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MapInterface_SonicStd(b *testing.B) {
	v := benchMapValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Interface_Sonic(b *testing.B) {
	v := []interface{}{1}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Interface_SonicStd(b *testing.B) {
	v := []interface{}{1}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Bool_Sonic(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	enc := sonic.ConfigDefault.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Bool_SonicStd(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	enc := sonic.ConfigStd.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Marshal_Bool_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Marshal_Bool_SonicStd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Int_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(1); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Int_SonicStd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(1); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MarshalJSON_Sonic(b *testing.B) {
	v := &marshaler{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigDefault.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MarshalJSON_SonicStd(b *testing.B) {
	v := &marshaler{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.ConfigStd.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_EncodeBigData_Sonic(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		enc := sonic.ConfigDefault.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&codeStruct); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_EncodeBigData_SonicStd(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		enc := sonic.ConfigStd.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&codeStruct); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_MarshalBigData_Sonic(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := sonic.ConfigDefault.Marshal(&codeStruct); err != nil {
				b.Fatal("Marshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_MarshalBigData_SonicStd(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := sonic.ConfigStd.Marshal(&codeStruct); err != nil {
				b.Fatal("Marshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

// likeSonicOptions are the options of go-json which make it do what sonic.ConfigDefault does:
// the benchmarks of GoJsonLikeSonic are the ones to compare with the benchmarks of Sonic,
// as the ones of GoJson are the ones to compare with the benchmarks of SonicStd.
var likeSonicOptions = []gojson.EncodeOptionFunc{
	gojson.DisableHTMLEscape(),
	gojson.DisableNormalizeUTF8(),
	gojson.UnorderedMap(),
}

func benchLikeSonic(b *testing.B, v interface{}) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalWithOption(v, likeSonicOptions...); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_GoJsonLikeSonic(b *testing.B) {
	benchLikeSonic(b, NewSmallPayload())
}

func Benchmark_Encode_MediumStructCached_GoJsonLikeSonic(b *testing.B) {
	benchLikeSonic(b, NewMediumPayload())
}

func Benchmark_Encode_LargeStructCached_GoJsonLikeSonic(b *testing.B) {
	benchLikeSonic(b, NewLargePayload())
}

func Benchmark_Encode_MapInterface_GoJsonLikeSonic(b *testing.B) {
	benchLikeSonic(b, benchMapValue())
}

package benchmark

import (
	"encoding/json"
	"testing"

	gojay "github.com/francoispqt/gojay"
	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

func Benchmark_Encode_SmallStruct_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(NewSmallPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStruct_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(NewSmallPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStruct_GoJay(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojay.MarshalJSONObject(NewSmallPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStruct_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(NewSmallPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStruct_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(NewMediumPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStruct_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(NewMediumPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStruct_GoJay(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojay.MarshalJSONObject(NewMediumPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStruct_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(NewMediumPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStruct_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(NewLargePayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStruct_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(NewLargePayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStruct_GoJay(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojay.MarshalJSONObject(NewLargePayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStruct_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(NewLargePayload()); err != nil {
			b.Fatal(err)
		}
	}
}

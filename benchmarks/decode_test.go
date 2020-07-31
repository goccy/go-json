package benchmark

import (
	"bytes"
	"encoding/json"
	"testing"

	gojay "github.com/francoispqt/gojay"
	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

func Benchmark_Decode_SmallStruct_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := SmallPayload{}
		if err := json.Unmarshal(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := SmallPayload{}
		if err := jsoniter.Unmarshal(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_GoJayDecode(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(SmallFixture)
	for n := 0; n < b.N; n++ {
		reader.Reset(SmallFixture)
		result := SmallPayload{}
		if err := gojay.NewDecoder(reader).DecodeObject(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_GoJay(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := SmallPayload{}
		if err := gojay.UnmarshalJSONObject(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_GoJayUnsafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		if err := gojay.Unsafe.UnmarshalJSONObject(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_GoJsonDecode(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(SmallFixture)
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		reader.Reset(SmallFixture)
		if err := gojson.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		if err := gojson.Unmarshal(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		if err := gojson.UnmarshalNoEscape(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := MediumPayload{}
		if err := json.Unmarshal(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := MediumPayload{}
		if err := jsoniter.Unmarshal(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_GoJay(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := MediumPayload{}
		if err := gojay.UnmarshalJSONObject(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_GoJayUnsafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		if err := gojay.Unsafe.UnmarshalJSONObject(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		if err := gojson.Unmarshal(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		if err := gojson.UnmarshalNoEscape(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := LargePayload{}
		if err := json.Unmarshal(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := LargePayload{}
		if err := jsoniter.Unmarshal(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_GoJay(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := LargePayload{}
		if err := gojay.UnmarshalJSONObject(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_GoJayUnsafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := gojay.Unsafe.UnmarshalJSONObject(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := gojson.Unmarshal(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := gojson.UnmarshalNoEscape(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

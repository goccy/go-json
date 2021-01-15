package benchmark

import (
	"encoding/json"
	"testing"

	gojay "github.com/francoispqt/gojay"
	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
	segmentiojson "github.com/segmentio/encoding/json"
	"github.com/wI2L/jettison"
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

func Benchmark_Encode_SmallStruct_EasyJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewSmallPayloadEasyJson().MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStruct_Jettison(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(NewSmallPayload()); err != nil {
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

func Benchmark_Encode_SmallStruct_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(NewSmallPayload()); err != nil {
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

func Benchmark_Encode_SmallStruct_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalNoEscape(NewSmallPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_EncodingJson(b *testing.B) {
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_EasyJson(b *testing.B) {
	cached := NewSmallPayloadEasyJson()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := cached.MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_Jettison(b *testing.B) {
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_GoJay(b *testing.B) {
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojay.MarshalJSONObject(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_SegmentioJson(b *testing.B) {
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_GoJson(b *testing.B) {
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_SmallStructCached_GoJsonNoEscape(b *testing.B) {
	cached := NewSmallPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalNoEscape(cached); err != nil {
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

func Benchmark_Encode_MediumStruct_EasyJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewMediumPayloadEasyJson().MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStruct_Jettison(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(NewMediumPayload()); err != nil {
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

func Benchmark_Encode_MediumStruct_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(NewMediumPayload()); err != nil {
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

func Benchmark_Encode_MediumStruct_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalNoEscape(NewMediumPayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_EncodingJson(b *testing.B) {
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_EasyJson(b *testing.B) {
	cached := NewMediumPayloadEasyJson()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := cached.MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_Jettison(b *testing.B) {
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_GoJay(b *testing.B) {
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojay.MarshalJSONObject(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_SegmentioJson(b *testing.B) {
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_GoJson(b *testing.B) {
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MediumStructCached_GoJsonNoEscape(b *testing.B) {
	cached := NewMediumPayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalNoEscape(cached); err != nil {
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

func Benchmark_Encode_LargeStruct_EasyJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewLargePayloadEasyJson().MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStruct_Jettison(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(NewLargePayload()); err != nil {
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

func Benchmark_Encode_LargeStruct_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(NewLargePayload()); err != nil {
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

func Benchmark_Encode_LargeStruct_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalNoEscape(NewLargePayload()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_EncodingJson(b *testing.B) {
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_EasyJson(b *testing.B) {
	cached := NewLargePayloadEasyJson()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := cached.MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_Jettison(b *testing.B) {
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_GoJay(b *testing.B) {
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojay.MarshalJSONObject(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_SegmentioJson(b *testing.B) {
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_GoJson(b *testing.B) {
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(cached); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_LargeStructCached_GoJsonNoEscape(b *testing.B) {
	cached := NewLargePayload()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalNoEscape(cached); err != nil {
			b.Fatal(err)
		}
	}
}

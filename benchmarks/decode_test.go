package benchmark

import (
	"bytes"
	"encoding/json"
	"testing"

	gojay "github.com/francoispqt/gojay"
	gojson "github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
	segmentiojson "github.com/segmentio/encoding/json"
	fastjson "github.com/valyala/fastjson"
)

func Benchmark_Decode_SmallStruct_Unmarshal_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := SmallPayload{}
		if err := json.Unmarshal(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_FastJson(b *testing.B) {
	smallFixture := string(SmallFixture)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var p fastjson.Parser
		if _, err := p.Parse(smallFixture); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := SmallPayload{}
		if err := segmentiojson.Unmarshal(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := SmallPayload{}
		if err := jsoniter.Unmarshal(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_GoJay(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := SmallPayload{}
		if err := gojay.UnmarshalJSONObject(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_GoJayUnsafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		if err := gojay.Unsafe.UnmarshalJSONObject(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		if err := gojson.Unmarshal(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Unmarshal_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		if err := gojson.UnmarshalNoEscape(SmallFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Stream_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(SmallFixture)
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		reader.Reset(SmallFixture)
		if err := json.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Stream_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(SmallFixture)
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		reader.Reset(SmallFixture)
		if err := segmentiojson.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Stream_JsonIter(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(SmallFixture)
	for i := 0; i < b.N; i++ {
		result := SmallPayload{}
		reader.Reset(SmallFixture)
		if err := jsoniter.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Stream_GoJay(b *testing.B) {
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

func Benchmark_Decode_SmallStruct_Stream_GoJson(b *testing.B) {
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

func Benchmark_Decode_MediumStruct_Unmarshal_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := MediumPayload{}
		if err := json.Unmarshal(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Unmarshal_FastJson(b *testing.B) {
	mediumFixture := string(MediumFixture)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var p fastjson.Parser
		if _, err := p.Parse(mediumFixture); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Unmarshal_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		if err := segmentiojson.Unmarshal(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Unmarshal_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := MediumPayload{}
		if err := jsoniter.Unmarshal(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Unmarshal_GoJay(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := MediumPayload{}
		if err := gojay.UnmarshalJSONObject(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Unmarshal_GoJayUnsafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		if err := gojay.Unsafe.UnmarshalJSONObject(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Unmarshal_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		if err := gojson.Unmarshal(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Unmarshal_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		if err := gojson.UnmarshalNoEscape(MediumFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Stream_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(MediumFixture)
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		reader.Reset(MediumFixture)
		if err := json.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Stream_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(MediumFixture)
	for n := 0; n < b.N; n++ {
		reader.Reset(MediumFixture)
		result := MediumPayload{}
		if err := segmentiojson.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Stream_JsonIter(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(MediumFixture)
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		reader.Reset(MediumFixture)
		if err := jsoniter.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Stream_GoJay(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(MediumFixture)
	for n := 0; n < b.N; n++ {
		reader.Reset(MediumFixture)
		result := MediumPayload{}
		if err := gojay.NewDecoder(reader).DecodeObject(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_MediumStruct_Stream_GoJson(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(MediumFixture)
	for i := 0; i < b.N; i++ {
		result := MediumPayload{}
		reader.Reset(MediumFixture)
		if err := gojson.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := LargePayload{}
		if err := json.Unmarshal(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_FastJson(b *testing.B) {
	largeFixture := string(LargeFixture)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var p fastjson.Parser
		if _, err := p.Parse(largeFixture); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := segmentiojson.Unmarshal(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := LargePayload{}
		if err := jsoniter.Unmarshal(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJay(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		result := LargePayload{}
		if err := gojay.UnmarshalJSONObject(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJayUnsafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := gojay.Unsafe.UnmarshalJSONObject(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := gojson.Unmarshal(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJsonNoEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := gojson.UnmarshalNoEscape(LargeFixture, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJsonFirstWinMode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := gojson.UnmarshalWithOption(
			LargeFixture,
			&result,
			gojson.DecodeFieldPriorityFirstWin(),
		); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Unmarshal_GoJsonNoEscapeFirstWinMode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		if err := gojson.UnmarshalNoEscape(
			LargeFixture,
			&result,
			gojson.DecodeFieldPriorityFirstWin(),
		); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Stream_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(LargeFixture)
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		reader.Reset(LargeFixture)
		if err := json.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Stream_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(LargeFixture)
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		reader.Reset(LargeFixture)
		if err := segmentiojson.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Stream_JsonIter(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(LargeFixture)
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		reader.Reset(LargeFixture)
		if err := jsoniter.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Stream_GoJay(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(LargeFixture)
	for n := 0; n < b.N; n++ {
		reader.Reset(LargeFixture)
		result := LargePayload{}
		if err := gojay.NewDecoder(reader).DecodeObject(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Stream_GoJson(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(LargeFixture)
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		reader.Reset(LargeFixture)
		if err := gojson.NewDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeStruct_Stream_GoJsonFirstWinMode(b *testing.B) {
	b.ReportAllocs()
	reader := bytes.NewReader(LargeFixture)
	for i := 0; i < b.N; i++ {
		result := LargePayload{}
		reader.Reset(LargeFixture)
		if err := gojson.NewDecoder(reader).DecodeWithOption(
			&result,
			gojson.DecodeFieldPriorityFirstWin(),
		); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_LargeSlice_EscapedString_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v []string
		if err := gojson.Unmarshal(LargeSliceEscapedString, &v); err != nil {
			b.Fatal(err)
		}
	}
}

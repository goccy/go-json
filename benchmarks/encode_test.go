package benchmark

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	gojay "github.com/francoispqt/gojay"
	gojson "github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
	"github.com/pquerna/ffjson/ffjson"
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

func Benchmark_Encode_SmallStruct_FFJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ffjson.Marshal(NewSmallPayloadFFJson()); err != nil {
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

func Benchmark_Encode_SmallStruct_GoJsonColored(b *testing.B) {
	colorOpt := gojson.Colorize(gojson.DefaultColorScheme)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalWithOption(NewSmallPayload(), colorOpt); err != nil {
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

func Benchmark_Encode_SmallStructCached_FFJson(b *testing.B) {
	cached := NewSmallPayloadFFJson()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ffjson.Marshal(cached); err != nil {
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

func Benchmark_Encode_SmallStructCached_GoJsonColored(b *testing.B) {
	cached := NewSmallPayload()
	colorOpt := gojson.Colorize(gojson.DefaultColorScheme)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalWithOption(cached, colorOpt); err != nil {
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

func Benchmark_Encode_MediumStruct_GoJsonColored(b *testing.B) {
	colorOpt := gojson.Colorize(gojson.DefaultColorScheme)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalWithOption(NewMediumPayload(), colorOpt); err != nil {
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

func Benchmark_Encode_MediumStructCached_GoJsonColored(b *testing.B) {
	cached := NewMediumPayload()
	colorOpt := gojson.Colorize(gojson.DefaultColorScheme)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalWithOption(cached, colorOpt); err != nil {
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

func Benchmark_Encode_LargeStruct_GoJsonColored(b *testing.B) {
	colorOpt := gojson.Colorize(gojson.DefaultColorScheme)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalWithOption(NewLargePayload(), colorOpt); err != nil {
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

func Benchmark_Encode_LargeStructCached_GoJsonColored(b *testing.B) {
	cached := NewLargePayload()
	colorOpt := gojson.Colorize(gojson.DefaultColorScheme)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalWithOption(cached, colorOpt); err != nil {
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

func benchMapValue() map[string]interface{} {
	return map[string]interface{}{
		"a": 1,
		"b": 2.1,
		"c": "hello",
		"d": struct {
			V int
		}{
			V: 1,
		},
		"e": true,
	}
}

func Benchmark_Encode_MapInterface_EncodingJson(b *testing.B) {
	v := benchMapValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MapInterface_JsonIter(b *testing.B) {
	v := benchMapValue()
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MapInterface_Jettison(b *testing.B) {
	v := benchMapValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MapInterface_SegmentioJson(b *testing.B) {
	v := benchMapValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MapInterface_GoJson(b *testing.B) {
	v := benchMapValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Interface_SegmentioJson(b *testing.B) {
	v := []interface{}{1}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Interface_GoJson(b *testing.B) {
	v := []interface{}{1}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Bool_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Bool_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Bool_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	enc := segmentiojson.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Bool_GoJson(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	enc := gojson.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Marshal_Bool_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Marshal_Bool_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Marshal_Bool_Jettison(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Marshal_Bool_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Marshal_Bool_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(true); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Int_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(1); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Int_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(1); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Int_Jettison(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(1); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Int_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(1); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_Int_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(1); err != nil {
			b.Fatal(err)
		}
	}
}

type marshaler struct{}

func (*marshaler) MarshalJSON() ([]byte, error) {
	return []byte(`"hello"`), nil
}

func Benchmark_Encode_MarshalJSON_EncodingJson(b *testing.B) {
	v := &marshaler{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MarshalJSON_JsonIter(b *testing.B) {
	v := &marshaler{}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MarshalJSON_Jettison(b *testing.B) {
	v := &marshaler{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := jettison.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MarshalJSON_SegmentioJson(b *testing.B) {
	v := &marshaler{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := segmentiojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MarshalJSON_GoJson(b *testing.B) {
	v := &marshaler{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

type queryTestX struct {
	XA int
	XB string
	XC *queryTestY
	XD bool
	XE float32
}

type queryTestY struct {
	YA int
	YB string
	YC bool
	YD float32
}

func Benchmark_Encode_FilterByMap(b *testing.B) {
	v := &queryTestX{
		XA: 1,
		XB: "xb",
		XC: &queryTestY{
			YA: 2,
			YB: "yb",
			YC: true,
			YD: 4,
		},
		XD: true,
		XE: 5,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		filteredMap := map[string]interface{}{
			"XA": v.XA,
			"XB": v.XB,
			"XC": map[string]interface{}{
				"YA": v.XC.YA,
				"YB": v.XC.YB,
			},
		}
		if _, err := gojson.Marshal(filteredMap); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_FilterByFieldQuery(b *testing.B) {
	query, err := gojson.BuildFieldQuery(
		"XA",
		"XB",
		gojson.BuildSubFieldQuery("XC").Fields(
			"YA",
			"YB",
		),
	)
	if err != nil {
		b.Fatal(err)
	}
	v := &queryTestX{
		XA: 1,
		XB: "xb",
		XC: &queryTestY{
			YA: 2,
			YB: "yb",
			YC: true,
			YD: 4,
		},
		XD: true,
		XE: 5,
	}
	ctx := gojson.SetFieldQueryToContext(context.Background(), query)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalContext(ctx, v); err != nil {
			b.Fatal(err)
		}
	}
}

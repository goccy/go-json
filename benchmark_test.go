package json_test

import (
	"testing"

	gojay "github.com/francoispqt/gojay"
	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

type T struct {
	A int     `json:"a"`
	B float64 `json:"b"`
	C string  `json:"c"`
}

func (t *T) MarshalJSONObject(enc *gojay.Encoder) {
	enc.IntKey("a", t.A)
	enc.FloatKey("b", t.B)
	enc.StringKey("c", t.C)
}

func (t *T) IsNil() bool {
	return t == nil
}

func newT() *T {
	return &T{A: 1, B: 3.14, C: `hello"world`}
}

func Benchmark_jsoniter(b *testing.B) {
	v := newT()
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_gojay(b *testing.B) {
	v := newT()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojay.MarshalJSONObject(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_gojson(b *testing.B) {
	v := newT()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

package json_test

import (
	"testing"

	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

type T struct {
	A int
	B float64
	C string
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

func Benchmark_gojson(b *testing.B) {
	v := newT()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

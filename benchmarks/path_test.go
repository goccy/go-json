package benchmark

import (
	"testing"

	gojson "github.com/bytedance/sonic"
)

func Benchmark_Decode_SmallStruct_UnmarshalPath_GoJson(b *testing.B) {
	path, err := gojson.CreatePath("$.st")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v int
		if err := path.Unmarshal(SmallFixture, &v); err != nil {
			b.Fatal(err)
		}
		if v != 1 {
			b.Fatal("failed to unmarshal path")
		}
	}
}

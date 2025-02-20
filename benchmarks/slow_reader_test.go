package benchmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	gojson "github.com/bytedance/sonic"
)

// Benchmark decoding from a slow io.Reader that never fills the buffer completely
func Benchmark_Decode_SlowReader_EncodingJson(b *testing.B) {
	var expected LargePayload
	if err := json.Unmarshal(LargeFixture, &expected); err != nil {
		b.Fatal(err)
	}
	for _, chunkSize := range [5]int{16384, 4096, 1024, 256, 64} {
		b.Run(fmt.Sprintf("chunksize %v", chunkSize), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				index = 0
				var got LargePayload
				if err := json.NewDecoder(slowReader{chunkSize: chunkSize}).Decode(&got); err != nil {
					b.Fatal(err)
				}
				if !reflect.DeepEqual(expected, got) {
					b.Fatalf("failed to decode. expected:[%+v] but got:[%+v]", expected, got)
				}
			}
		})
	}
}

func Benchmark_Decode_SlowReader_GoJson(b *testing.B) {
	var expected LargePayload
	if err := json.Unmarshal(LargeFixture, &expected); err != nil {
		b.Fatal(err)
	}
	for _, chunkSize := range []int{16384, 4096, 1024, 256, 64} {
		b.Run(fmt.Sprintf("chunksize %v", chunkSize), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				index = 0
				var got LargePayload
				if err := gojson.NewDecoder(slowReader{chunkSize: chunkSize}).Decode(&got); err != nil {
					b.Fatal(err)
				}
				if !reflect.DeepEqual(expected, got) {
					b.Fatalf("failed to decode. expected:[%+v] but got:[%+v]", expected, got)
				}
			}
		})
	}
}

type slowReader struct {
	chunkSize int
}

var index int

func (s slowReader) Read(p []byte) (n int, err error) {
	smallBuf := make([]byte, Min(s.chunkSize, len(p)))
	x := bytes.NewReader(LargeFixture)
	n, err = x.ReadAt(smallBuf, int64(index))
	index += n
	copy(p, smallBuf)
	return
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

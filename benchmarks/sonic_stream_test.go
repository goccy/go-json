package benchmark

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"
)

// The stream benchmarks of sonic beside the ones of go-json which had none: the same input and reader as the
// GoJson ones of decode_test.go ( *_Stream_GoJson ), bench_test.go ( BenchmarkCodeDecoder ) and slow_reader_test.go
// ( Benchmark_Decode_SlowReader_GoJson ), decoded by the Decoder of sonic.ConfigDefault ( Sonic ) and of
// sonic.ConfigStd ( SonicStd ). The library is the part of the name after the last underscore, by which benchcheck
// leaves them out.

func benchStructStream[T any](b *testing.B, data []byte, newDecoder func(io.Reader) streamDecoder) {
	pretouchSonic()
	b.ReportAllocs()
	reader := bytes.NewReader(data)
	for i := 0; i < b.N; i++ {
		var result T
		reader.Reset(data)
		if err := newDecoder(reader).Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_SmallStruct_Stream_Sonic(b *testing.B) {
	benchStructStream[SmallPayload](b, SmallFixture, newSonicStreamDecoder)
}

func Benchmark_Decode_SmallStruct_Stream_SonicStd(b *testing.B) {
	benchStructStream[SmallPayload](b, SmallFixture, newSonicStdStreamDecoder)
}

func Benchmark_Decode_MediumStruct_Stream_Sonic(b *testing.B) {
	benchStructStream[MediumPayload](b, MediumFixture, newSonicStreamDecoder)
}

func Benchmark_Decode_MediumStruct_Stream_SonicStd(b *testing.B) {
	benchStructStream[MediumPayload](b, MediumFixture, newSonicStdStreamDecoder)
}

func Benchmark_Decode_LargeStruct_Stream_Sonic(b *testing.B) {
	benchStructStream[LargePayload](b, LargeFixture, newSonicStreamDecoder)
}

func Benchmark_Decode_LargeStruct_Stream_SonicStd(b *testing.B) {
	benchStructStream[LargePayload](b, LargeFixture, newSonicStdStreamDecoder)
}

func benchCodeDecoder(b *testing.B, newDecoder func(io.Reader) streamDecoder) {
	pretouchSonic()
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		dec := newDecoder(&buf)
		var r codeResponse
		for pb.Next() {
			buf.Write(codeJSON)
			// hide EOF
			buf.WriteByte('\n')
			buf.WriteByte('\n')
			buf.WriteByte('\n')
			if err := dec.Decode(&r); err != nil {
				if err != io.EOF {
					b.Fatal("Decode:", err)
				}
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func BenchmarkCodeDecoder_Sonic(b *testing.B) { benchCodeDecoder(b, newSonicStreamDecoder) }

func BenchmarkCodeDecoder_SonicStd(b *testing.B) { benchCodeDecoder(b, newSonicStdStreamDecoder) }

func benchSlowReader(b *testing.B, newDecoder func(io.Reader) streamDecoder) {
	pretouchSonic()
	var expected LargePayload
	if err := stdjson.Unmarshal(LargeFixture, &expected); err != nil {
		b.Fatal(err)
	}
	for _, chunkSize := range []int{16384, 4096, 1024, 256, 64} {
		b.Run(fmt.Sprintf("chunksize %v", chunkSize), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				index = 0
				var got LargePayload
				if err := newDecoder(slowReader{chunkSize: chunkSize}).Decode(&got); err != nil {
					b.Fatal(err)
				}
				if !reflect.DeepEqual(expected, got) {
					b.Fatalf("failed to decode. expected:[%+v] but got:[%+v]", expected, got)
				}
			}
		})
	}
}

func Benchmark_Decode_SlowReader_Sonic(b *testing.B) { benchSlowReader(b, newSonicStreamDecoder) }

func Benchmark_Decode_SlowReader_SonicStd(b *testing.B) { benchSlowReader(b, newSonicStdStreamDecoder) }

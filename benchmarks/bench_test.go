// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Large data benchmark.
// The JSON data is a summary of agl's changes in the
// go, webkit, and chromium open source projects.
// We benchmark converting between the JSON form
// and in-memory data structures.

package benchmark

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	stdjson "encoding/json"

	"github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
	segmentiojson "github.com/segmentio/encoding/json"
	"github.com/wI2L/jettison"
)

type codeResponse struct {
	Tree     *codeNode `json:"tree"`
	Username string    `json:"username"`
}

type codeNode struct {
	Name     string      `json:"name"`
	Kids     []*codeNode `json:"kids"`
	CLWeight float64     `json:"cl_weight"`
	Touches  int         `json:"touches"`
	MinT     int64       `json:"min_t"`
	MaxT     int64       `json:"max_t"`
	MeanT    int64       `json:"mean_t"`
}

var codeJSON []byte
var codeStruct codeResponse

func codeInit() {
	f, err := os.Open("testdata/code.json.gz")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(gz)
	if err != nil {
		panic(err)
	}

	codeJSON = data

	if err := stdjson.Unmarshal(codeJSON, &codeStruct); err != nil {
		panic("unmarshal code.json: " + err.Error())
	}
	{
		stdjsonbytes, err := stdjson.Marshal(&codeStruct)
		if err != nil {
			panic("marshal code.json: " + err.Error())
		}
		jsonbytes, err := json.Marshal(&codeStruct)
		if err != nil {
			panic("marshal code.json: " + err.Error())
		}
		if len(stdjsonbytes) != len(jsonbytes) {
			panic(fmt.Sprintf("stdjson = %d but go-json = %d", len(stdjsonbytes), len(jsonbytes)))
		}
	}
	if _, err := json.Marshal(&codeStruct); err != nil {
		panic("marshal code.json: " + err.Error())
	}
	if !bytes.Equal(data, codeJSON) {
		println("different lengths", len(data), len(codeJSON))
		for i := 0; i < len(data) && i < len(codeJSON); i++ {
			if data[i] != codeJSON[i] {
				println("re-marshal: changed at byte", i)
				println("orig: ", string(codeJSON[i-10:i+10]))
				println("new: ", string(data[i-10:i+10]))
				break
			}
		}
		panic("re-marshal code.json: different result")
	}
}

func Benchmark_EncodeBigData_GoJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		enc := json.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&codeStruct); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_EncodeBigData_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		enc := stdjson.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&codeStruct); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_EncodeBigData_JsonIter(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.RunParallel(func(pb *testing.PB) {
		enc := json.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&codeStruct); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_EncodeBigData_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		enc := segmentiojson.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&codeStruct); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_MarshalBigData_GoJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := json.Marshal(&codeStruct); err != nil {
				b.Fatal("Marshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_MarshalBigData_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := stdjson.Marshal(&codeStruct); err != nil {
				b.Fatal("Marshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_MarshalBigData_JsonIter(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := json.Marshal(&codeStruct); err != nil {
				b.Fatal("Marshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_MarshalBigData_Jettison(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := jettison.Marshal(&codeStruct); err != nil {
				b.Fatal("Marshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func Benchmark_MarshalBigData_SegmentioJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := segmentiojson.Marshal(&codeStruct); err != nil {
				b.Fatal("Marshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func benchMarshalBytes(n int, marshaler func(interface{}) ([]byte, error)) func(*testing.B) {
	sample := []byte("hello world")
	// Use a struct pointer, to avoid an allocation when passing it as an
	// interface parameter to Marshal.
	v := &struct {
		Bytes []byte
	}{
		bytes.Repeat(sample, (n/len(sample))+1)[:n],
	}
	return func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := marshaler(v); err != nil {
				b.Fatal("Marshal:", err)
			}
		}
	}
}

func Benchmark_MarshalBytes_EncodingJson(b *testing.B) {
	// 32 fits within encodeState.scratch.
	b.Run("32", benchMarshalBytes(32, stdjson.Marshal))
	// 256 doesn't fit in encodeState.scratch, but is small enough to
	// allocate and avoid the slower base64.NewEncoder.
	b.Run("256", benchMarshalBytes(256, stdjson.Marshal))
	// 4096 is large enough that we want to avoid allocating for it.
	b.Run("4096", benchMarshalBytes(4096, stdjson.Marshal))
}

func Benchmark_MarshalBytes_JsonIter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	// 32 fits within encodeState.scratch.
	b.Run("32", benchMarshalBytes(32, json.Marshal))
	// 256 doesn't fit in encodeState.scratch, but is small enough to
	// allocate and avoid the slower base64.NewEncoder.
	b.Run("256", benchMarshalBytes(256, json.Marshal))
	// 4096 is large enough that we want to avoid allocating for it.
	b.Run("4096", benchMarshalBytes(4096, json.Marshal))
}

func Benchmark_MarshalBytes_GoJson(b *testing.B) {
	b.ReportAllocs()
	// 32 fits within encodeState.scratch.
	b.Run("32", benchMarshalBytes(32, json.Marshal))
	// 256 doesn't fit in encodeState.scratch, but is small enough to
	// allocate and avoid the slower base64.NewEncoder.
	b.Run("256", benchMarshalBytes(256, json.Marshal))
	// 4096 is large enough that we want to avoid allocating for it.
	b.Run("4096", benchMarshalBytes(4096, json.Marshal))
}

func Benchmark_EncodeRawMessage_EncodingJson(b *testing.B) {
	b.ReportAllocs()

	m := struct {
		A int
		B json.RawMessage
	}{}

	b.RunParallel(func(pb *testing.PB) {
		enc := stdjson.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&m); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
}

func Benchmark_EncodeRawMessage_JsonIter(b *testing.B) {
	b.ReportAllocs()

	m := struct {
		A int
		B json.RawMessage
	}{}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	b.RunParallel(func(pb *testing.PB) {
		enc := json.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&m); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
}

func Benchmark_EncodeRawMessage_GoJson(b *testing.B) {
	b.ReportAllocs()

	m := struct {
		A int
		B json.RawMessage
	}{}
	b.RunParallel(func(pb *testing.PB) {
		enc := json.NewEncoder(io.Discard)
		for pb.Next() {
			if err := enc.Encode(&m); err != nil {
				b.Fatal("Encode:", err)
			}
		}
	})
}

func Benchmark_MarshalString_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	j := struct {
		Bar string `json:"bar,string"`
	}{
		Bar: `foobar`,
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := stdjson.Marshal(&j); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_MarshalString_JsonIter(b *testing.B) {
	b.ReportAllocs()
	j := struct {
		Bar string `json:"bar,string"`
	}{
		Bar: `foobar`,
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := json.Marshal(&j); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_MarshalString_GoJson(b *testing.B) {
	b.ReportAllocs()
	j := struct {
		Bar string `json:"bar,string"`
	}{
		Bar: `foobar`,
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := json.Marshal(&j); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCodeDecoder(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		dec := json.NewDecoder(&buf)
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

func BenchmarkUnicodeDecoder(b *testing.B) {
	b.ReportAllocs()
	j := []byte(`"\uD83D\uDE01"`)
	b.SetBytes(int64(len(j)))
	r := bytes.NewReader(j)
	dec := json.NewDecoder(r)
	var out string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := dec.Decode(&out); err != nil {
			if err != io.EOF {
				b.Fatal("Decode:", err)
			}
		}
		r.Seek(0, 0)
	}
}

func BenchmarkDecoderStream(b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()
	var buf bytes.Buffer
	dec := json.NewDecoder(&buf)
	buf.WriteString(`"` + strings.Repeat("x", 1000000) + `"` + "\n\n\n")
	var x interface{}
	if err := dec.Decode(&x); err != nil {
		b.Fatal("Decode:", err)
	}
	ones := strings.Repeat(" 1\n", 300000) + "\n\n\n"
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if i%300000 == 0 {
			buf.WriteString(ones)
		}
		x = nil
		if err := dec.Decode(&x); err != nil || x != 1.0 {
			if err != io.EOF {
				b.Fatalf("Decode: %v after %d", err, i)
			}
		}
	}
}

func BenchmarkCodeUnmarshal(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var r codeResponse
			if err := json.Unmarshal(codeJSON, &r); err != nil {
				b.Fatal("Unmarshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func BenchmarkCodeUnmarshalReuse(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		var r codeResponse
		for pb.Next() {
			if err := json.Unmarshal(codeJSON, &r); err != nil {
				b.Fatal("Unmarshal:", err)
			}
		}
	})
	b.SetBytes(int64(len(codeJSON)))
}

func BenchmarkUnmarshalString(b *testing.B) {
	b.ReportAllocs()
	data := []byte(`"hello, world"`)
	b.RunParallel(func(pb *testing.PB) {
		var s string
		for pb.Next() {
			if err := json.Unmarshal(data, &s); err != nil {
				b.Fatal("Unmarshal:", err)
			}
		}
	})
}

func BenchmarkUnmarshalFloat64(b *testing.B) {
	b.ReportAllocs()
	data := []byte(`3.14`)
	b.RunParallel(func(pb *testing.PB) {
		var f float64
		for pb.Next() {
			if err := json.Unmarshal(data, &f); err != nil {
				b.Fatal("Unmarshal:", err)
			}
		}
	})
}

func BenchmarkUnmarshalInt64(b *testing.B) {
	b.ReportAllocs()
	data := []byte(`3`)
	b.RunParallel(func(pb *testing.PB) {
		var x int64
		for pb.Next() {
			if err := json.Unmarshal(data, &x); err != nil {
				b.Fatal("Unmarshal:", err)
			}
		}
	})
}

func BenchmarkIssue10335(b *testing.B) {
	b.ReportAllocs()
	j := []byte(`{"a":{ }}`)
	b.RunParallel(func(pb *testing.PB) {
		var s struct{}
		for pb.Next() {
			if err := json.Unmarshal(j, &s); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkUnmapped(b *testing.B) {
	b.ReportAllocs()
	j := []byte(`{"s": "hello", "y": 2, "o": {"x": 0}, "a": [1, 99, {"x": 1}]}`)
	b.RunParallel(func(pb *testing.PB) {
		var s struct{}
		for pb.Next() {
			if err := json.Unmarshal(j, &s); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_Compact_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := stdjson.Compact(&buf, codeJSON); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Compact_GoJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := json.Compact(&buf, codeJSON); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Indent_EncodingJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := stdjson.Indent(&buf, codeJSON, "-", " "); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Indent_GoJson(b *testing.B) {
	b.ReportAllocs()
	if codeJSON == nil {
		b.StopTimer()
		codeInit()
		b.StartTimer()
	}
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := json.Indent(&buf, codeJSON, "-", " "); err != nil {
			b.Fatal(err)
		}
	}
}

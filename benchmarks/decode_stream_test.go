package benchmark

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
)

// The benchmarks of decoding from a stream ( Decoder.Decode ) read from a reader which gives the input in chunks of
// a fixed size and has no Len method, as a network connection or a file does. A reader which has one, as the
// bytes.Reader of the *_Stream_* benchmarks ( decode_test.go ), lets go-json read the whole input at once, so these
// never see a value which is larger than the buffer of the decoder.
//
// The payloads are:
//   - a value which fits in the first buffer of a decoder ( SmallStruct, 143 B; the one of go-json has 512 bytes );
//   - values larger than the first buffer, which fit in one chunk or in a few ( MediumStruct, 2 KB, and Twitter,
//     10 KB );
//   - values of megabytes, many times a chunk ( Code, Records: an array of records, CodeAny: Code into interface{} );
//   - many values one after another ( NDJSON: one record per line, decoded until the end of the input );
//   - a large value which is mostly skipped ( SkipRecords: an object of an array of records and the count, decoded
//     into a struct of the count only ).
//
// Each benchmark is run with chunks of 4 KiB and 32 KiB, but a payload which fits in the smaller chunk only with
// that one: a larger chunk can't change what the reader gives.
//
// Besides ns/op, B/op and allocs/op, they report peak-live-B: the most memory which the decoding keeps alive at
// once, beyond what was live before it ( see measureStreamPeak ). A decoder which reads a whole value before
// decoding it keeps the input of the value; one which decodes it while reading keeps only what it hasn't decoded.

// streamChunkSizes are the sizes of the chunks which the reader of the benchmarks gives at most at a time, smallest
// first.
var streamChunkSizes = []int{4 << 10, 32 << 10}

// chunkReader reads data in chunks of at most chunk bytes. It has no Len method, so that a decoder can't know how
// much is left. If sample is set, it is called before a read when the reader has given another sampleStep bytes
// since the last call, and once before the first read.
type chunkReader struct {
	data       []byte
	pos        int
	chunk      int
	sample     func()
	sampleStep int
	nextSample int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if r.sample != nil && r.pos >= r.nextSample {
		r.sample()
		r.nextSample = r.pos + r.sampleStep
	}
	if r.pos == len(r.data) {
		return 0, io.EOF
	}
	n := copy(p[:min(len(p), r.chunk)], r.data[r.pos:])
	r.pos += n
	return n, nil
}

// streamDecoder is the Decoder of each library: encoding/json, go-json and sonic.
type streamDecoder interface {
	Decode(v any) error
}

// streamPayload is the input of a benchmark and the type of its values.
type streamPayload struct {
	data func() []byte
	// newValue returns a pointer to a new value to decode one value of the input into.
	newValue func() any
	// many is set when the input is many values one after another, which are decoded until the end of the input.
	many bool
	// collect is set for the payloads of megabytes, whose decoding is preceded by a collection outside of the
	// measured time. A decoding of them allocates megabytes, so where the collections fall would otherwise move
	// the time and B/op ( a pool emptied by a collection or not ) from one run to another.
	collect bool
}

// decodeStream decodes every value of the input of the payload by dec, each into a new value which is passed to
// each, if set, and dropped before the next is decoded. It returns how many values it decoded.
func (p *streamPayload) decodeStream(dec streamDecoder, each func(v any)) (int, error) {
	n := 0
	for {
		v := p.newValue()
		if err := dec.Decode(v); err != nil {
			if p.many && n > 0 && errors.Is(err, io.EOF) {
				return n, nil
			}
			return n, err
		}
		n++
		if each != nil {
			each(v)
		}
		if !p.many {
			return n, nil
		}
	}
}

// decodeAll returns every value of the input of the payload decoded by dec.
func (p *streamPayload) decodeAll(b *testing.B, dec streamDecoder) []any {
	values := []any{}
	if _, err := p.decodeStream(dec, func(v any) { values = append(values, v) }); err != nil {
		b.Fatal(err)
	}
	return values
}

// benchStream measures decoding every value of the payload from a chunkReader by the decoders of newDecoder.
func benchStream(b *testing.B, p *streamPayload, newDecoder func(io.Reader) streamDecoder) {
	data := p.data()
	for i, chunk := range streamChunkSizes {
		if i > 0 && len(data) <= streamChunkSizes[i-1] {
			break
		}
		peak := measureStreamPeak(b, p, newDecoder, chunk)
		b.Run(fmt.Sprintf("%dKiB", chunk>>10), func(b *testing.B) {
			r := &chunkReader{data: data, chunk: chunk}
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if p.collect {
					b.StopTimer()
					runtime.GC()
					b.StartTimer()
				}
				r.pos = 0
				if _, err := p.decodeStream(newDecoder(r), nil); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(peak, "peak-live-B")
		})
	}
}

// streamPeakSamples is how many times measureStreamPeak samples the live heap while the input is read.
const streamPeakSamples = 64

// measureStreamPeak checks that the decoders of newDecoder decode the values which encoding/json decodes, and
// returns the most memory which a second decoding kept alive at once, beyond what was live before it: the first
// one fills the caches of the library, as the compiled types, which stay alive after it.
//
// The live heap is the heap after a collection ( runtime.GC, then HeapAlloc of runtime.ReadMemStats ). It is
// sampled before the reader gives each 1/streamPeakSamples of the input, while the decoder holds what it has
// read and not decoded yet, and once after the last value is decoded, while the decoder and the value are still
// alive ( of many values, each is dropped before the next is decoded, as a program reading a stream of records
// does ). What a decoder allocates and drops between two samples isn't seen, and the peak which a decoder reaches
// between two samples is seen only as far as it is kept until the next one: the measure is of what the decoder
// keeps while it reads, which is what decoding while reading bounds, not of the garbage it makes ( B/op ).
func measureStreamPeak(b *testing.B, p *streamPayload, newDecoder func(io.Reader) streamDecoder, chunk int) float64 {
	b.Helper()
	data := p.data()
	// the values are dead after the check, so the collections before the measurement free them.
	got := p.decodeAll(b, newDecoder(&chunkReader{data: data, chunk: chunk}))
	if !reflect.DeepEqual(got, p.decodeAll(b, stdjson.NewDecoder(bytes.NewReader(data)))) {
		b.Fatal("the decoded values differ from the ones of encoding/json")
	}
	var base, peak uint64
	var stats runtime.MemStats
	sample := func() {
		runtime.GC()
		runtime.ReadMemStats(&stats)
		if stats.HeapAlloc > base {
			peak = max(peak, stats.HeapAlloc-base)
		}
	}
	r := &chunkReader{data: data, chunk: chunk, sample: sample, sampleStep: max(len(data)/streamPeakSamples, 1)}
	// two collections empty the sync.Pools, whose objects survive one in the victim cache: a buffer which a
	// library takes from its pool is then allocated during the decoding and counted, as in a program whose pools
	// the collections have emptied since its last decoding.
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&stats)
	base = stats.HeapAlloc
	dec := newDecoder(r)
	var kept any
	var keep func(v any)
	if !p.many {
		keep = func(v any) { kept = v }
	}
	if _, err := p.decodeStream(dec, keep); err != nil {
		b.Fatal(err)
	}
	sample()
	runtime.KeepAlive(dec)
	runtime.KeepAlive(kept)
	return float64(peak)
}

func newEncodingJsonStreamDecoder(r io.Reader) streamDecoder { return stdjson.NewDecoder(r) }

func newGoJsonStreamDecoder(r io.Reader) streamDecoder { return gojson.NewDecoder(r) }

func newSonicStreamDecoder(r io.Reader) streamDecoder { return sonic.ConfigDefault.NewDecoder(r) }

func newSonicStdStreamDecoder(r io.Reader) streamDecoder { return sonic.ConfigStd.NewDecoder(r) }

// streamRecord is a record of the generated payloads.
type streamRecord struct {
	ID      int64         `json:"id"`
	Name    string        `json:"name"`
	Email   string        `json:"email"`
	Score   float64       `json:"score"`
	Active  bool          `json:"active"`
	Tags    []string      `json:"tags"`
	Address streamAddress `json:"address"`
}

type streamAddress struct {
	Street string `json:"street"`
	City   string `json:"city"`
	Zip    string `json:"zip"`
}

// streamRecordCount ( Records and SkipRecords ) and streamLineCount ( NDJSON ) are the numbers of the records of
// the generated payloads, about 4 MB and 2 MB.
const (
	streamRecordCount = 20000
	streamLineCount   = 10000
)

// streamRecords returns n records, which are the same at every call.
func streamRecords(n int) []streamRecord {
	records := make([]streamRecord, n)
	for i := range records {
		s := strconv.Itoa(i)
		records[i] = streamRecord{
			ID:     int64(i) * 7919,
			Name:   "user " + s,
			Email:  "user" + s + "@example.com",
			Score:  float64(i%1000) / 8,
			Active: i%3 == 0,
			Tags:   []string{"tag" + strconv.Itoa(i%10), "group" + strconv.Itoa(i%7)},
			Address: streamAddress{
				Street: s + " Main Street",
				City:   "City " + strconv.Itoa(i%50),
				Zip:    strconv.Itoa(10000 + i%90000),
			},
		}
	}
	return records
}

func mustStdMarshal(v any) []byte {
	data, err := stdjson.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

var streamRecordsJSON = sync.OnceValue(func() []byte {
	return mustStdMarshal(streamRecords(streamRecordCount))
})

var streamNDJSON = sync.OnceValue(func() []byte {
	var buf bytes.Buffer
	for _, r := range streamRecords(streamLineCount) {
		buf.Write(mustStdMarshal(r))
		buf.WriteByte('\n')
	}
	return buf.Bytes()
})

// streamSkipJSON has the count after the records, so that the records are skipped before the field is found.
var streamSkipJSON = sync.OnceValue(func() []byte {
	return []byte(`{"items":` + string(streamRecordsJSON()) + `,"count":` + strconv.Itoa(streamRecordCount) + `}`)
})

type streamCount struct {
	Count int `json:"count"`
}

func streamCodeJSON() []byte {
	if codeJSON == nil {
		codeInit()
	}
	return codeJSON
}

func init() {
	sonicTypes = append(sonicTypes,
		reflect.TypeOf(&[]streamRecord{}), reflect.TypeOf(&streamRecord{}), reflect.TypeOf(&streamCount{}))
}

var (
	streamSmallPayload = &streamPayload{
		data:     func() []byte { return SmallFixture },
		newValue: func() any { return &SmallPayload{} },
	}
	streamMediumPayload = &streamPayload{
		data:     func() []byte { return MediumFixture },
		newValue: func() any { return &MediumPayload{} },
	}
	streamTwitterPayload = &streamPayload{
		data:     func() []byte { return []byte(TwitterJson) },
		newValue: func() any { return &TwitterStruct{} },
	}
	streamCodePayload = &streamPayload{
		data:     streamCodeJSON,
		newValue: func() any { return &codeResponse{} },
		collect:  true,
	}
	streamCodeAnyPayload = &streamPayload{
		data:     streamCodeJSON,
		newValue: func() any { return new(any) },
		collect:  true,
	}
	streamRecordsPayload = &streamPayload{
		data:     streamRecordsJSON,
		newValue: func() any { return &[]streamRecord{} },
		collect:  true,
	}
	streamNDJSONPayload = &streamPayload{
		data:     streamNDJSON,
		newValue: func() any { return &streamRecord{} },
		many:     true,
		collect:  true,
	}
	streamSkipPayload = &streamPayload{
		data:     streamSkipJSON,
		newValue: func() any { return &streamCount{} },
		collect:  true,
	}
)

func Benchmark_Decode_SmallStruct_ChunkedStream_EncodingJson(b *testing.B) {
	benchStream(b, streamSmallPayload, newEncodingJsonStreamDecoder)
}

func Benchmark_Decode_SmallStruct_ChunkedStream_GoJson(b *testing.B) {
	benchStream(b, streamSmallPayload, newGoJsonStreamDecoder)
}

func Benchmark_Decode_SmallStruct_ChunkedStream_Sonic(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamSmallPayload, newSonicStreamDecoder)
}

func Benchmark_Decode_SmallStruct_ChunkedStream_SonicStd(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamSmallPayload, newSonicStdStreamDecoder)
}

func Benchmark_Decode_MediumStruct_ChunkedStream_EncodingJson(b *testing.B) {
	benchStream(b, streamMediumPayload, newEncodingJsonStreamDecoder)
}

func Benchmark_Decode_MediumStruct_ChunkedStream_GoJson(b *testing.B) {
	benchStream(b, streamMediumPayload, newGoJsonStreamDecoder)
}

func Benchmark_Decode_MediumStruct_ChunkedStream_Sonic(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamMediumPayload, newSonicStreamDecoder)
}

func Benchmark_Decode_MediumStruct_ChunkedStream_SonicStd(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamMediumPayload, newSonicStdStreamDecoder)
}

func Benchmark_Decode_Twitter_ChunkedStream_EncodingJson(b *testing.B) {
	benchStream(b, streamTwitterPayload, newEncodingJsonStreamDecoder)
}

func Benchmark_Decode_Twitter_ChunkedStream_GoJson(b *testing.B) {
	benchStream(b, streamTwitterPayload, newGoJsonStreamDecoder)
}

func Benchmark_Decode_Twitter_ChunkedStream_Sonic(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamTwitterPayload, newSonicStreamDecoder)
}

func Benchmark_Decode_Twitter_ChunkedStream_SonicStd(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamTwitterPayload, newSonicStdStreamDecoder)
}

func Benchmark_Decode_Code_ChunkedStream_EncodingJson(b *testing.B) {
	benchStream(b, streamCodePayload, newEncodingJsonStreamDecoder)
}

func Benchmark_Decode_Code_ChunkedStream_GoJson(b *testing.B) {
	benchStream(b, streamCodePayload, newGoJsonStreamDecoder)
}

func Benchmark_Decode_Code_ChunkedStream_Sonic(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamCodePayload, newSonicStreamDecoder)
}

func Benchmark_Decode_Code_ChunkedStream_SonicStd(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamCodePayload, newSonicStdStreamDecoder)
}

func Benchmark_Decode_CodeAny_ChunkedStream_EncodingJson(b *testing.B) {
	benchStream(b, streamCodeAnyPayload, newEncodingJsonStreamDecoder)
}

func Benchmark_Decode_CodeAny_ChunkedStream_GoJson(b *testing.B) {
	benchStream(b, streamCodeAnyPayload, newGoJsonStreamDecoder)
}

func Benchmark_Decode_CodeAny_ChunkedStream_Sonic(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamCodeAnyPayload, newSonicStreamDecoder)
}

func Benchmark_Decode_CodeAny_ChunkedStream_SonicStd(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamCodeAnyPayload, newSonicStdStreamDecoder)
}

func Benchmark_Decode_Records_ChunkedStream_EncodingJson(b *testing.B) {
	benchStream(b, streamRecordsPayload, newEncodingJsonStreamDecoder)
}

func Benchmark_Decode_Records_ChunkedStream_GoJson(b *testing.B) {
	benchStream(b, streamRecordsPayload, newGoJsonStreamDecoder)
}

func Benchmark_Decode_Records_ChunkedStream_Sonic(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamRecordsPayload, newSonicStreamDecoder)
}

func Benchmark_Decode_Records_ChunkedStream_SonicStd(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamRecordsPayload, newSonicStdStreamDecoder)
}

func Benchmark_Decode_NDJSON_ChunkedStream_EncodingJson(b *testing.B) {
	benchStream(b, streamNDJSONPayload, newEncodingJsonStreamDecoder)
}

func Benchmark_Decode_NDJSON_ChunkedStream_GoJson(b *testing.B) {
	benchStream(b, streamNDJSONPayload, newGoJsonStreamDecoder)
}

func Benchmark_Decode_NDJSON_ChunkedStream_Sonic(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamNDJSONPayload, newSonicStreamDecoder)
}

func Benchmark_Decode_NDJSON_ChunkedStream_SonicStd(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamNDJSONPayload, newSonicStdStreamDecoder)
}

func Benchmark_Decode_SkipRecords_ChunkedStream_EncodingJson(b *testing.B) {
	benchStream(b, streamSkipPayload, newEncodingJsonStreamDecoder)
}

func Benchmark_Decode_SkipRecords_ChunkedStream_GoJson(b *testing.B) {
	benchStream(b, streamSkipPayload, newGoJsonStreamDecoder)
}

func Benchmark_Decode_SkipRecords_ChunkedStream_Sonic(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamSkipPayload, newSonicStreamDecoder)
}

func Benchmark_Decode_SkipRecords_ChunkedStream_SonicStd(b *testing.B) {
	pretouchSonic()
	benchStream(b, streamSkipPayload, newSonicStdStreamDecoder)
}

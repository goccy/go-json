package json_test

import (
	"bytes"
	"io"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	json "github.com/goccy/go-json"
)

// A Decoder keeps in memory what the value it decodes needs, whatever the reader.

// A reader which tells its length ( bytes.Reader, strings.Reader, bytes.Buffer ) is read as any reader: its input is
// not copied whole for its first value.
func TestDecodeStreamLenReaderNotReadWhole(t *testing.T) {
	var b strings.Builder
	// a first value larger than the first buffer, which the buffer grows for
	b.WriteString(`{"a":0,"b":"` + strings.Repeat("x", 2000) + "\"}\n")
	for i := 0; i < 100000; i++ {
		b.WriteString(`{"a":` + strconv.Itoa(i) + "}\n")
	}
	in := b.String()
	for _, r := range []interface {
		io.Reader
		Len() int
	}{strings.NewReader(in), bytes.NewReader([]byte(in)), bytes.NewBufferString(in)} {
		var v struct {
			A int
			B string
		}
		if err := json.NewDecoder(r).Decode(&v); err != nil {
			t.Fatal(err)
		}
		if r.Len() < len(in)/2 {
			t.Fatalf("%T: %d bytes read of %d for the first value", r, len(in)-r.Len(), len(in))
		}
	}
}

// zeroReader gives one byte of data per read, and nothing, without an error, every other read: a reader may return
// nothing, and more at the next read.
type zeroReader struct {
	data []byte
	zero bool
}

func (r *zeroReader) Read(p []byte) (int, error) {
	if r.zero = !r.zero; r.zero {
		return 0, nil
	}
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p[:1], r.data)
	r.data = r.data[n:]
	return n, nil
}

// A read which returns nothing is followed by another one.
func TestDecodeStreamZeroReads(t *testing.T) {
	for _, in := range []string{`null`, `true`, `false`, `12 `, `"s"`, `{"a":null}`, `[true,false]`, `{"a":[1,` + strings.Repeat(" ", 1000) + `null]}`} {
		var want, got any
		if err := json.Unmarshal([]byte(in), &want); err != nil {
			t.Fatal(err)
		}
		if err := json.NewDecoder(&zeroReader{data: []byte(in)}).Decode(&got); err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		gotJSON, _ := json.Marshal(got)
		wantJSON, _ := json.Marshal(want)
		if !bytes.Equal(gotJSON, wantJSON) {
			t.Fatalf("%q: got %s, want %s", in, gotJSON, wantJSON)
		}
	}
}

// A NUL byte of the input is an invalid character wherever the reads leave it, as in the input read at once.
func TestDecodeStreamNulByte(t *testing.T) {
	for _, in := range []string{"{\"a\":{\"b\":\x00}}", "{\"a\":[1,\x002]}", "{\"a\":\"x\x00\"}", "[\x00]"} {
		var v any
		want := json.Unmarshal([]byte(in), &v)
		for _, r := range []io.Reader{&zeroReader{data: []byte(in)}, &oneByteReader{data: in}, strings.NewReader(in)} {
			got := json.NewDecoder(r).Decode(&v)
			if want == nil || got == nil || got.Error() != want.Error() {
				t.Fatalf("%q from %T: got %v, want %v", in, r, got, want)
			}
		}
	}
}

// The elements of a long slice are decoded into a buffer which doubles, and which becomes the slice, so that the slice
// leaves little garbage: the growth of append, and the copy into a slice of the exact length, left several times its
// size.
func TestDecodeSliceBufferGrowth(t *testing.T) {
	const n = 200000
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(i))
	}
	b.WriteByte(']')
	in := []byte(b.String())
	var v []int64
	// the first decoding compiles the decoder
	if err := json.Unmarshal(in, &v); err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	v = nil
	if err := json.Unmarshal(in, &v); err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&after)
	if len(v) != n {
		t.Fatalf("%d elements", len(v))
	}
	// the input copied for the decoding, and the buffer doubled up to the length, which becomes the slice
	sliceSize := uint64(n * 8)
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > uint64(len(in))+4*sliceSize {
		t.Fatalf("%d bytes allocated for a slice of %d bytes from %d bytes", allocated, sliceSize, len(in))
	}
}

// A slice decoded after a failed decoding of a longer one keeps no more room than its own elements need: the buffer of
// the longer one, which the pool keeps, is copied from.
func TestDecodeSliceRoomAfterLongerSlice(t *testing.T) {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < 1<<20; i++ {
		b.WriteString(strconv.Itoa(i) + ",")
	}
	b.WriteString("x]")
	var long []int
	if err := json.Unmarshal([]byte(b.String()), &long); err == nil {
		t.Fatal("no error")
	}
	short := "[" + strings.TrimSuffix(strings.Repeat("1,", 70), ",") + "]"
	var v []int
	if err := json.Unmarshal([]byte(short), &v); err != nil {
		t.Fatal(err)
	}
	if len(v) != 70 || cap(v) > 3*70 {
		t.Fatalf("len %d cap %d", len(v), cap(v))
	}
}

// A truncated value of a reader which tells it has nothing left, as a bytes.Buffer which may be written to later, ends
// with the end of its input, whatever the size of the value.
func TestDecodeStreamTruncatedValueOfLenReader(t *testing.T) {
	for _, in := range []string{"[" + strings.Repeat("1,", 300), `"` + strings.Repeat("a", 2000), `{"a":` + strings.Repeat(`"b",`, 500)} {
		for _, r := range []io.Reader{bytes.NewBufferString(in), strings.NewReader(in), bytes.NewReader([]byte(in))} {
			done := make(chan error, 1)
			go func() { done <- json.NewDecoder(r).Decode(new(any)) }()
			select {
			case err := <-done:
				if err == nil {
					t.Fatalf("%.10q from %T: no error", in, r)
				}
			case <-time.After(10 * time.Second):
				t.Fatalf("%.10q from %T: Decode doesn't end", in, r)
			}
		}
	}
}

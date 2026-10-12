package json_test

import (
	"bytes"
	"fmt"
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

// roomReader is a strings.Reader which tells its length, and keeps the room of every read, len(p). It reads most
// bytes at most per read, if most is set, as a pipe or a connection may.
type roomReader struct {
	*strings.Reader
	most  int
	rooms []int
}

func (r *roomReader) Read(p []byte) (int, error) {
	r.rooms = append(r.rooms, len(p))
	if r.most > 0 {
		p = p[:min(len(p), r.most)]
	}
	return r.Reader.Read(p)
}

// envelopeItems is a value of an envelope which an interface value points to.
type envelopeItems struct {
	Items []map[string]any `json:"items"`
}

func (*envelopeItems) String() string { return "envelopeItems" }

// A stream is read by reads of the size of its first buffer at most, whatever the size of a value, whatever the reader
// tells and however few bytes it gives a read, also while and after the buffer grows for a token longer than it, or a
// value which is kept whole.
func TestDecodeStreamFixedReads(t *testing.T) {
	rec := `{"name":"alice","age":30,"tags":["a","b"],"address":{"city":"Springfield"}}`
	records := "[" + strings.Repeat(rec+",", 20000) + rec + "]"
	long := strings.Repeat("x", 100000)
	// a value of 2.5 MB in what a pointer held by an interface points to
	envelope := `{"Type":"records","Data":{"items":[` + strings.Repeat(rec+",", 32000) + rec + `]}}`
	for _, c := range []struct {
		name     string
		in       string
		newValue func() any
	}{
		{"records", records, func() any { return new([]map[string]any) }},
		{"records into any", records, func() any { return new(any) }},
		{"array into []byte", `{"b":[` + strings.Repeat("1,", 100000) + `1]}`, func() any { return new(struct{ B []byte }) }},
		{"skipped", `{"skip":` + records + `,"name":"x"}`, func() any { return new(struct{ Name string }) }},
		{"lines", strings.Repeat(rec+"\n", 20000), func() any { return new(any) }},
		{"interface holding a pointer", envelope, func() any {
			return &struct {
				Type string
				Data any
			}{Data: &envelopeItems{}}
		}},
		{"interface with methods holding a pointer", envelope, func() any {
			return &struct {
				Type string
				Data fmt.Stringer
			}{Data: &envelopeItems{}}
		}},
		{"long string", `["` + long + `",` + records[1:], func() any { return new([]any) }},
		{"raw message", records, func() any { return new(json.RawMessage) }},
		{"raw message, then records", `{"raw":` + records + `,"rest":` + records + `}`, func() any {
			return new(struct {
				Raw  json.RawMessage
				Rest []any
			})
		}},
	} {
		for _, most := range []int{0, 100} {
			r := &roomReader{Reader: strings.NewReader(c.in), most: most}
			dec := json.NewDecoder(r)
			for {
				if err := dec.Decode(c.newValue()); err == io.EOF {
					break
				} else if err != nil {
					t.Fatalf("%s: %v", c.name, err)
				}
			}
			// the first read is given the whole buffer
			for i, room := range r.rooms {
				if room > r.rooms[0] {
					t.Fatalf("%s, %d bytes at most a read: read %d of %d bytes, the buffer has %d", c.name, most, i, room, r.rooms[0])
				}
			}
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

// A slice decoded after a longer one keeps no more room than its own elements need: the buffer of the longer one, which
// the pool keeps, is copied from. The pool keeps it when a stream which decodes the elements of an array one by one
// into a slice with room for them fails at the first element past that room: the elements go back to the slice.
func TestDecodeSliceRoomAfterLongerSlice(t *testing.T) {
	const n = 1 << 20
	var b strings.Builder
	b.WriteString("[" + strings.Repeat(" ", 1000))
	for i := 0; i < n; i++ {
		b.WriteString(strconv.Itoa(i) + ",")
	}
	b.WriteString("x]")
	long := make([]int, 0, n)
	if err := json.NewDecoder(strings.NewReader(b.String())).Decode(&long); err == nil || len(long) != n {
		t.Fatalf("%d elements, %v", len(long), err)
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

// A long number is read in a time linear in its length, by reads of the size of the buffer: the digits read are
// scanned once. Skipping it takes a few times as long as skipping a string of its length, where a scan of the whole
// number at every read takes hundreds of times as long, however fast or busy the machine.
func TestDecodeStreamLongNumber(t *testing.T) {
	const n = 1 << 20
	number, str := "1"+strings.Repeat("2", n-1), `"`+strings.Repeat("s", n-2)+`"`
	// skipTime returns the least time of a few decodings of an object whose member, value, is skipped: the least of
	// enough of them is one which no other work of the machine interrupted
	skipTime := func(value string) time.Duration {
		in := `{"x":` + value + `}`
		least := time.Duration(1<<63 - 1)
		for i := 0; i < 7; i++ {
			var v struct{}
			start := time.Now()
			if err := json.NewDecoder(struct{ io.Reader }{strings.NewReader(in)}).Decode(&v); err != nil {
				t.Fatal(err)
			}
			least = min(least, time.Since(start))
		}
		return least
	}
	if numberTime, stringTime := skipTime(number), skipTime(str); numberTime > 50*stringTime {
		t.Fatalf("%v to skip a number of %d bytes, %v to skip a string of as many", numberTime, n, stringTime)
	}
	var v []json.Number
	if err := json.NewDecoder(struct{ io.Reader }{strings.NewReader("[" + number + "]")}).Decode(&v); err != nil || len(v) != 1 {
		t.Fatalf("len(v) = %d and err = %v, want 1 and nil", len(v), err)
	}
	if string(v[0]) != number {
		t.Fatalf("got a number of %d bytes, which differs from the one of %d bytes", len(v[0]), n)
	}
}

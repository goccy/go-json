package json_test

import (
	"errors"
	"io"
	"reflect"
	"runtime"
	"strings"
	"testing"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// roomReader is a strings.Reader which keeps the room of every read, len(p). It reads most bytes at most per read,
// if most is set, as a pipe or a connection may.
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

// readRecord is a member of the inputs of the tests of UnmarshalRead.
const readRecord = `{"name":"alice","age":30,"tags":["a","b"],"address":{"city":"Springfield"}}`

// readRecords is an array of the records of 1.5 MB.
var readRecords = "[" + strings.Repeat(readRecord+",", 20000) + readRecord + "]"

// readCounter is a value of a method which reads a jsontext.Decoder: it counts the elements of an array, whatever
// its size.
type readCounter struct{ n int }

func (c *readCounter) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	for dec.PeekKind() != ']' {
		if err := dec.SkipValue(); err != nil {
			return err
		}
		c.n++
	}
	_, err := dec.ReadToken()
	return err
}

// readCount is readCounter of a type of a number, whose pointer has a decoder of its own.
type readCount int

func (c *readCount) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var counter readCounter
	err := counter.UnmarshalJSONFrom(dec)
	*c = readCount(counter.n)
	return err
}

// The input is read by reads of the size of the first one at most, whatever the size of a value and however few bytes
// a read gives, also while and after the buffer grows for a token longer than it or a value which is held whole, and
// for the decoder given to a method.
func TestUnmarshalReadFixedReads(t *testing.T) {
	long := strings.Repeat("x", 100000)
	for _, c := range []struct {
		name     string
		in       string
		newValue func() any
	}{
		{"records", readRecords, func() any { return new([]map[string]any) }},
		{"records into any", readRecords, func() any { return new(any) }},
		{"records into structs", readRecords, func() any {
			return new([]struct {
				Name string
				Tags []string
			})
		}},
		{"skipped", `{"skip":` + readRecords + `,"name":"x"}`, func() any { return new(struct{ Name string }) }},
		{"long string", `["` + long + `",` + readRecords[1:], func() any { return new([]any) }},
		{"raw value", readRecords, func() any { return new(jsontext.Value) }},
		{"method", `{"records":` + readRecords + `}`, func() any { return new(struct{ Records readCounter }) }},
	} {
		for _, most := range []int{0, 100} {
			r := &roomReader{Reader: strings.NewReader(c.in), most: most}
			if err := json.UnmarshalRead(r, c.newValue()); err != nil {
				t.Fatalf("%s: %v", c.name, err)
			}
			for i, room := range r.rooms {
				if room > r.rooms[0] {
					t.Fatalf("%s, %d bytes at most a read: read %d of %d bytes, the first has %d", c.name, most, i, room, r.rooms[0])
				}
			}
		}
	}
}

// A value which is decoded element by element is not held whole: what is allocated is what the value takes, not
// what the input takes.
func TestUnmarshalReadMemory(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own")
	}
	for _, c := range []struct {
		name     string
		in       string
		newValue func() any
		opts     []json.Options
	}{
		{"skipped", `{"skip":` + readRecords + `,"name":"x"}`, func() any { return new(struct{ Name string }) }, nil},
		// the names of no field, which are checked for duplicates by sets of their names unless the option is set
		{"skipped members", readRecords, func() any { return new([]struct{}) }, []json.Options{jsontext.AllowDuplicateNames(true)}},
		// a JSON array read into bytes with FormatBytesWithLegacySemantics, whose elements past the length are skipped,
		// as long as the records
		{"bytes of a JSON array", `{"A":[` + strings.Repeat("1,", len(readRecords)/len("1,")) + `1]}`, func() any { return new(struct{ A [2]byte }) },
			[]json.Options{gojson.FormatBytesWithLegacySemantics(true), gojson.UnmarshalArrayFromAnyLength(true)}},
		{"counted by a method", `{"records":` + readRecords + `}`, func() any { return new(struct{ Records readCounter }) }, nil},
		{"counted by a method through a pointer", `{"records":` + readRecords + `}`, func() any {
			return new(struct{ Records *readCounter })
		}, nil},
		{"counted by a method of a number through a pointer", `{"records":` + readRecords + `}`, func() any {
			return new(struct{ Records *readCount })
		}, nil},
	} {
		// the type is compiled before, which is cached
		if err := json.UnmarshalRead(strings.NewReader(c.in), c.newValue(), c.opts...); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		v := c.newValue()
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		if err := json.UnmarshalRead(strings.NewReader(c.in), v, c.opts...); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		runtime.ReadMemStats(&after)
		if allocated := after.TotalAlloc - before.TotalAlloc; allocated > uint64(len(c.in))/20 {
			t.Fatalf("%s: %d bytes allocated for an input of %d", c.name, allocated, len(c.in))
		}
	}
}

func TestUnmarshalReadOf(t *testing.T) {
	type record struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var got []record
	if err := json.UnmarshalReadOf(strings.NewReader(readRecords), &got); err != nil {
		t.Fatal(err)
	}
	var want []record
	if err := json.Unmarshal([]byte(readRecords), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %d records, want %d", len(got), len(want))
	}
	// the errors of UnmarshalRead, with what was decoded before them
	for _, in := range []string{`{"name":1}`, `{"name":"a"} x`, `{"name":`, ``} {
		var a, b record
		errOf := json.UnmarshalReadOf(strings.NewReader(in), &a)
		err := json.UnmarshalRead(strings.NewReader(in), &b)
		if (errOf == nil) != (err == nil) || err != nil && errOf.Error() != err.Error() || a != b {
			t.Fatalf("%s: UnmarshalReadOf: %+v, %v, UnmarshalRead: %+v, %v", in, a, errOf, b, err)
		}
	}
	var serr *json.SemanticError
	if err := json.UnmarshalReadOf[record](strings.NewReader(`{}`), nil); !errors.As(err, &serr) {
		t.Fatalf("expected a SemanticError, got %v", err)
	}
}

// UnmarshalReadOf decodes a value which needs no allocation of its own without one: the stream, its buffer and the
// value are reused.
func TestUnmarshalReadOfZeroAllocation(t *testing.T) {
	if raceEnabled {
		t.Skip("sync.Pool drops items at random under the race detector")
	}
	r := strings.NewReader(unmarshalOfSmallInput)
	var v unmarshalOfSmall
	allocs := testing.AllocsPerRun(1000, func() {
		r.Reset(unmarshalOfSmallInput)
		var w unmarshalOfSmall
		if err := json.UnmarshalReadOf(io.Reader(r), &w); err != nil {
			t.Fatal(err)
		}
		v = w
	})
	// The strings share a buffer of the context, which is allocated once in a while.
	if allocs >= 0.1 {
		t.Fatalf("%v allocations per call", allocs)
	}
	if v.Ua != "user_agent" {
		t.Fatalf("got %+v", v)
	}
}

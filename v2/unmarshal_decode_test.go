package json_test

import (
	"bytes"
	"errors"
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// endReader gives its input at the first read, and fails the test at a read after it: the input of a reader which
// would block there.
type endReader struct {
	t   *testing.T
	in  string
	end bool
}

func (r *endReader) Read(p []byte) (int, error) {
	if r.end {
		r.t.Errorf("a read after the input %q, which a reader which blocks would not return", r.in)
		return 0, io.EOF
	}
	r.end = true
	return copy(p, r.in), nil
}

// The syntax error of a name twice in an object, before its colon, is returned without a read of the input after
// the colon, which a reader which blocks would not give.
func TestUnmarshalDecodeNameErrorReadsNoMore(t *testing.T) {
	in := `{"A":1,"A":`
	dec := jsontext.NewDecoder(&endReader{t: t, in: in})
	var v struct{ A int }
	if err := json.UnmarshalDecode(dec, &v); err == nil {
		t.Fatalf("%s: no error", in)
	}
}

// failingAtOnce fails without reading its value.
type failingAtOnce struct{}

var errFailingAtOnce = errors.New("failed without reading")

func (*failingAtOnce) UnmarshalJSONFrom(*jsontext.Decoder) error { return errFailingAtOnce }

// After the error of a method which read nothing of its value, a member of an object which the first read of the
// decoder ends in, the decoder is before the object, which it still has: the stream which decodes the object member
// by member, which gives the method a decoder of its own, read the rest of the object into the room of the buffer of
// the decoder, after the bytes it had, if it read more. After the error of a method of an object which the first read
// holds whole, the decoder is after the object.
func TestUnmarshalDecodeMethodErrorKeepsValue(t *testing.T) {
	const object = `{"B":3,"A":1,"C":2}`
	in := object + ` 5`
	for i := 1; i < len(in); i++ {
		for _, opts := range [][]json.Options{nil, {jsontext.AllowDuplicateNames(true), jsontext.AllowInvalidUTF8(true)}} {
			dec := jsontext.NewDecoder(io.MultiReader(strings.NewReader(in[:i]), strings.NewReader(in[i:])))
			var v struct {
				A    failingAtOnce
				B, C int
			}
			if err := json.UnmarshalDecode(dec, &v, opts...); !errors.Is(err, errFailingAtOnce) {
				t.Fatalf("first read of %d bytes: got %v", i, err)
			}
			want := object
			if i >= len(object) {
				want = "5"
			}
			if next, err := dec.ReadValue(); err != nil || string(next) != want {
				t.Fatalf("first read of %d bytes, options %v: after the error: %s, %v, want %s", i, opts, next, err, want)
			}
		}
	}
}

func TestUnmarshalDecodeOf(t *testing.T) {
	type record struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	in := `{"name":"a","age":1} {"name":"b","age":"x"} {"name":"c","age":3} {"name":`
	ours, theirs := jsontext.NewDecoder(strings.NewReader(in)), jsontext.NewDecoder(strings.NewReader(in))
	for i := 0; i < 5; i++ {
		var a, b record
		errOf := json.UnmarshalDecodeOf(ours, &a)
		err := json.UnmarshalDecode(theirs, &b)
		if (errOf == nil) != (err == nil) || err != nil && errOf.Error() != err.Error() || a != b {
			t.Fatalf("value %d: UnmarshalDecodeOf: %+v, %v, UnmarshalDecode: %+v, %v", i, a, errOf, b, err)
		}
	}
	var serr *json.SemanticError
	if err := json.UnmarshalDecodeOf[record](jsontext.NewDecoder(strings.NewReader(`{}`)), nil); !errors.As(err, &serr) {
		t.Fatalf("expected a SemanticError, got %v", err)
	}
}

// UnmarshalDecodeOf decodes a value which needs no allocation of its own without one, from a reader and from a
// bytes.Buffer, which the decoder reads in place.
func TestUnmarshalDecodeOfZeroAllocation(t *testing.T) {
	if raceEnabled {
		t.Skip("sync.Pool drops items at random under the race detector")
	}
	in := strings.Repeat(unmarshalOfSmallInput+"\n", 3000)
	for _, r := range []io.Reader{strings.NewReader(in), bytes.NewBufferString(in)} {
		dec := jsontext.NewDecoder(r)
		var v unmarshalOfSmall
		allocs := testing.AllocsPerRun(1000, func() {
			var w unmarshalOfSmall
			if err := json.UnmarshalDecodeOf(dec, &w); err != nil {
				t.Fatal(err)
			}
			v = w
		})
		// The strings share a buffer of the context, which is allocated once in a while.
		if allocs >= 0.1 {
			t.Fatalf("%T: %v allocations per call", r, allocs)
		}
		if v.Ua != "user_agent" {
			t.Fatalf("%T: got %+v", r, v)
		}
	}
}

// A value which a decoder of a reader holds no part of is decoded with bounded memory, as UnmarshalRead decodes it:
// what is allocated is what the value takes, not what the input takes.
func TestUnmarshalDecodeMemory(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own")
	}
	in := `{"skip":` + readRecords + `,"name":"x"}`
	newDecoder := func() *jsontext.Decoder {
		dec := jsontext.NewDecoder(strings.NewReader(in))
		for range 2 { // the bracket and the name
			if _, err := dec.ReadToken(); err != nil {
				t.Fatal(err)
			}
		}
		return dec
	}
	// the type is compiled before, which is cached
	if err := json.UnmarshalDecode(newDecoder(), new([]struct{}), jsontext.AllowDuplicateNames(true)); err != nil {
		t.Fatal(err)
	}
	dec := newDecoder()
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	var skipped []struct{}
	if err := json.UnmarshalDecode(dec, &skipped, jsontext.AllowDuplicateNames(true)); err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&after)
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > uint64(len(readRecords))/20 {
		t.Fatalf("%d bytes allocated for a value of %d", allocated, len(readRecords))
	}
	if len(skipped) != 20001 {
		t.Fatalf("got %d elements", len(skipped))
	}
	// the decoder goes on after the value
	var name string
	if _, err := dec.ReadToken(); err != nil {
		t.Fatal(err)
	}
	if err := json.UnmarshalDecode(dec, &name); err != nil || name != "x" {
		t.Fatalf("got %q, %v", name, err)
	}
}

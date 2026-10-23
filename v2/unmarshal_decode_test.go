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

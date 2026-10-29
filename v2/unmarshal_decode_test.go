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

// arrayOfAnyPlace is the place where arrayOfAny met null: the pointer and the input offset of its decoder.
var arrayOfAnyPlace struct {
	ptr    jsontext.Pointer
	offset int64
}

// arrayOfAny decodes an array into interface{} by UnmarshalDecode of the decoder of the function, which applies the
// function to its elements again, and fails for null, recording the place of its decoder there.
var arrayOfAny = json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *any) error {
	switch dec.PeekKind() {
	case 'n':
		arrayOfAnyPlace.ptr, arrayOfAnyPlace.offset = dec.StackPointer(), dec.InputOffset()
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		return errors.New("null")
	case '[':
		var a []any
		if err := json.UnmarshalDecode(dec, &a); err != nil {
			return err
		}
		*v = a
		return nil
	}
	return errors.ErrUnsupported
}))

// A function of a value which a stream holds whole in its buffer, which decodes it by UnmarshalDecode of its decoder,
// whose elements the function decodes again, finds its place in the whole input, where the buffer, whose start is
// not the start of the input, is: the levels around the buffer, and the offsets. The places are the ones of
// encoding/json/v2.
func TestUnmarshalDecodeFunctionInFunctionPlace(t *testing.T) {
	for _, c := range []struct {
		in      string
		ptr     jsontext.Pointer // of the error
		funcPtr jsontext.Pointer // of the decoder of the function, before null
		offset  int64
	}{
		{`[[null]]`, "/a/0/0", "/a/0", 616},
		{`[1,[2,null]]`, "/a/1/1", "/a/1/0", 619},
		{`[[[null]]]`, "/a/0/0/0", "/a/0/0", 617},
	} {
		in := `{"pad":"` + strings.Repeat("x", 600) + `","a":` + c.in + `}`
		for _, n := range []int{64, 512, 0} {
			arrayOfAnyPlace.ptr, arrayOfAnyPlace.offset = "", 0
			var v struct {
				Pad string `json:"pad"`
				A   any    `json:"a"`
			}
			dec := jsontext.NewDecoder(&roomReader{Reader: strings.NewReader(in), most: n})
			err := json.UnmarshalDecode(dec, &v, arrayOfAny)
			var serr *json.SemanticError
			if !errors.As(err, &serr) || serr.JSONPointer != c.ptr {
				t.Fatalf("%s, reads of %d: got %v, want the error within %q", c.in, n, err, c.ptr)
			}
			if arrayOfAnyPlace.ptr != c.funcPtr || arrayOfAnyPlace.offset != c.offset {
				t.Fatalf("%s, reads of %d: the function at %q@%d, want %q@%d", c.in, n, arrayOfAnyPlace.ptr, arrayOfAnyPlace.offset, c.funcPtr, c.offset)
			}
		}
	}
}

// stringThenFail decodes a string into interface{} by UnmarshalDecode of the decoder of the function, and fails
// after it; it decodes an array as stringThenFail of its elements.
var stringThenFail = json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *any) error {
	switch dec.PeekKind() {
	case '"':
		var s string
		if err := json.UnmarshalDecode(dec, &s); err != nil {
			return err
		}
		return errors.New("after the string")
	case '[':
		var a []any
		if err := json.UnmarshalDecode(dec, &a); err != nil {
			return err
		}
		*v = a
		return nil
	}
	return errors.ErrUnsupported
}))

// The error of a function which read its whole value by UnmarshalDecode is at the start of the value, also when the
// buffer of the decoder no longer has the start, a long value read in small reads. The places are the ones of
// encoding/json/v2.
func TestUnmarshalDecodeFunctionErrorAfterValue(t *testing.T) {
	long := `"` + strings.Repeat("x", 3000) + `"`
	for _, c := range []struct {
		in     string
		ptr    jsontext.Pointer
		offset int64
	}{
		{`"bad"`, "", 0},
		{`["bad"]`, "/0", 1},
		{`[["bad"]]`, "/0/0", 2},
		{long, "", 0},
		{`[` + long + `]`, "/0", 1},
		{`[1,` + long + `]`, "/1", 3},
	} {
		for _, n := range []int{1, 64, 0} {
			for _, read := range []bool{true, false} {
				var v any
				var err error
				if r := (&roomReader{Reader: strings.NewReader(c.in), most: n}); read {
					err = json.UnmarshalRead(r, &v, stringThenFail)
				} else {
					err = json.UnmarshalDecode(jsontext.NewDecoder(r), &v, stringThenFail)
				}
				var serr *json.SemanticError
				if !errors.As(err, &serr) || serr.JSONPointer != c.ptr || serr.ByteOffset != c.offset {
					t.Fatalf("%.20s, reads of %d, UnmarshalRead %v: got %v, want the error at %q@%d", c.in, n, read, err, c.ptr, c.offset)
				}
			}
		}
	}
}

// After the error of an output which is not a non-nil pointer, which reads nothing, the decoder is before its next
// value, also after the error of a value before it, which has a place in the input.
func TestUnmarshalDecodeInvalidOutputAfterError(t *testing.T) {
	for _, out := range []any{nil, 5, (*int)(nil)} {
		dec := jsontext.NewDecoder(strings.NewReader(`["x", 1, {"a":[2,3]}, 4]`))
		if _, err := dec.ReadToken(); err != nil {
			t.Fatal(err)
		}
		var i int
		if err := json.UnmarshalDecode(dec, &i); err == nil {
			t.Fatal("no error of a string decoded into an int")
		}
		if err := json.UnmarshalDecode(dec, out); err == nil {
			t.Fatalf("%#v: no error", out)
		}
		if next, err := dec.ReadValue(); err != nil || string(next) != "1" {
			t.Fatalf("%#v: after the error: %s, %v", out, next, err)
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
func TestUnmarshalDecodeOfAllocations(t *testing.T) {
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
		// a string is a copy of its own, and nothing else is allocated
		if allocs != unmarshalOfSmallStrings {
			t.Fatalf("%T: %v allocations per call, want %v", r, allocs, unmarshalOfSmallStrings)
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
	// The GC is run before the decoding which compiles the type and pools the state of the decoding, not before the
	// measured one. A GC moves the pooled state aside and drops the per-P arrays of the pools; the decoding after it
	// allocates the arrays again, by the number of Ps, and the state too when it runs on another P than the one which
	// pooled it, as it may after waiting for the GC: neither by the input, but more than the bound with many Ps. The
	// measured decoding takes the state which the one before pooled, unless the goroutine moved to another P since:
	// the decoding then allocates the state alone, which is within the bound.
	runtime.GC()
	if err := json.UnmarshalDecode(newDecoder(), new([]struct{}), jsontext.AllowDuplicateNames(true)); err != nil {
		t.Fatal(err)
	}
	dec := newDecoder()
	var before, after runtime.MemStats
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

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// The errors of a value decoded element by element are the ones of the value read at once, whatever the size of the
// value, and a string which doesn't end fails at its control character wherever it is.

type elementsMethods interface{ M() }

// A string which is not closed fails at the end of its line, whatever the destination, without reading the rest of
// the input, as encoding/json does ( issue 701 ).
func TestDecodeStreamUnterminatedStringAnyDestination(t *testing.T) {
	rest := strings.Repeat(`{"name":"x","age":1}`+"\n", 1<<16)
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"name":"unterminated,"age":1}` + "\n" + rest,
		`[` + pad + `{"name":"x"},{"name":"unterminated,"age":1}` + "\n" + rest,
		`{"a":[` + pad + `1],"name":"unterminated,"age":1}` + "\n" + rest,
	} {
		for _, target := range []struct {
			name string
			new  func() any
		}{
			{"raw message", func() any { return new(json.RawMessage) }},
			{"map", func() any { return new(map[string]string) }},
			{"map of any", func() any { return new(map[string]any) }},
			{"interface with methods", func() any { return new(elementsMethods) }},
			{"slice of structs", func() any { return new([]struct{ Name string }) }},
			{"struct", func() any { return new(struct{ Name string }) }},
			{"string", func() any { return new(string) }},
		} {
			r := &countingReader{r: &chunkReader{data: []byte(in), chunk: 512}}
			err := json.NewDecoder(r).Decode(target.new())
			var serr *json.SyntaxError
			if !errors.As(err, &serr) {
				t.Fatalf("%.30q into %s: got %v", in, target.name, err)
			}
			if r.n > 256<<10 {
				t.Fatalf("%.30q into %s: %d bytes read for %v", in, target.name, r.n, err)
			}
		}
	}
}

// decodeAtOnceAndByChunks decodes in into a new value by newValue, read at once and read by chunks of chunk bytes, and
// returns the two errors.
func decodeAtOnceAndByChunks(in string, chunk int, newValue func() any, setUp func(*json.Decoder)) (error, error) {
	whole := json.NewDecoder(bytes.NewReader([]byte(in)))
	chunked := json.NewDecoder(&chunkReader{data: []byte(in), chunk: chunk})
	setUp(whole)
	setUp(chunked)
	return whole.Decode(newValue()), chunked.Decode(newValue())
}

// The depth of a value is counted from the root for a syntax error in an element, and the depth limit of encoding/json
// applies to the value as it does to a value read at once.
func TestDecodeStreamElementsDepthErrors(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"a":` + pad + strings.Repeat(`{"a":`, 19999) + `null` + strings.Repeat("}", 20000),
		`{"a":` + pad + strings.Repeat(`{"a":`, 10001),
		`[` + pad + strings.Repeat(`[`, 9998) + `x` + strings.Repeat(`]`, 9999),
	} {
		for _, chunk := range []int{1000, 65536} {
			want, got := decodeAtOnceAndByChunks(in, chunk, func() any { return new(any) }, func(*json.Decoder) {})
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Fatalf("%d bytes, chunk %d: got %v, want %v", len(in), chunk, got, want)
			}
			var ws, gs *json.SyntaxError
			if errors.As(want, &ws) && errors.As(got, &gs) && ws.Offset != gs.Offset {
				t.Fatalf("%d bytes, chunk %d: offset %d, want %d", len(in), chunk, gs.Offset, ws.Offset)
			}
		}
	}
}

// An unknown field is an error once the value is read, as the one of a value read at once is: the end of the input
// before the end of the value is the error, as encoding/json reports it.
func TestDecodeStreamElementsUnknownFieldError(t *testing.T) {
	type S struct{ A, B int }
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"a":1,` + pad + `"zz":1,"b":2`,
		`{"a":1,` + pad + `"zz":1,"b":2}`,
		`{"a":1,` + pad + `"zz":"unterminated` + "\n" + strings.Repeat(`{"b":2}`, 1<<16),
		// a syntax error after the unknown field is the error, as encoding/json reads the value first
		`{"a":1,` + pad + `"zz":1,"b":2 x`,
		`{"a":1,` + pad + `"zz":1,"b":[1 2`,
		`{"a":1,` + pad + `"zz":1,"b":"x` + "\x01",
		`{"a":1,` + pad + `"zz":{"c":[1,{"d":2}]},"b":2}`,
	} {
		var stdErr error
		{
			dec := stdjson.NewDecoder(strings.NewReader(in))
			dec.DisallowUnknownFields()
			stdErr = dec.Decode(new(S))
		}
		for _, chunk := range []int{1000, 65536} {
			want, got := decodeAtOnceAndByChunks(in, chunk, func() any { return new(S) }, (*json.Decoder).DisallowUnknownFields)
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Fatalf("%.20q, chunk %d: got %v, want %v (encoding/json: %v)", in, chunk, got, want, stdErr)
			}
			var stdSyntax *stdjson.SyntaxError
			var syntax *json.SyntaxError
			if errors.Is(stdErr, io.ErrUnexpectedEOF) != errors.Is(got, io.ErrUnexpectedEOF) ||
				errors.As(stdErr, &stdSyntax) != errors.As(got, &syntax) {
				t.Fatalf("%.20q, chunk %d: got %v, encoding/json %v", in, chunk, got, stdErr)
			}
		}
	}
}

type elementsFailing struct{}

func (*elementsFailing) UnmarshalJSON([]byte) error { return errors.New("elementsFailing") }

// The error of a method is returned once the value is read, as the one of a value read at once is.
func TestDecodeStreamElementsMethodError(t *testing.T) {
	type S struct {
		A []elementsFailing
		B int
	}
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"A":[` + pad + `{},{}],"B":1}`,
		`{"A":[` + pad + `{},{}],"B":1 x`,
		`{"A":[` + pad + `{},{}],"B":[1 2`,
		`{"A":[` + pad + `{},{}],"B":1`,
		`[` + pad + `{"A":[{}]},{"B":2}]`,
	} {
		newValue := func() any { return new(S) }
		if in[0] == '[' {
			newValue = func() any { return new([]S) }
		}
		for _, chunk := range []int{1000, 65536} {
			want, got := decodeAtOnceAndByChunks(in, chunk, newValue, func(*json.Decoder) {})
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Fatalf("%.20q, chunk %d: got %v, want %v", in, chunk, got, want)
			}
		}
	}
}

type elementsBadKey string

func (k *elementsBadKey) UnmarshalText(b []byte) error {
	if string(b) == "bad" {
		return errors.New("elementsBadKey")
	}
	*k = elementsBadKey(b)
	return nil
}

// The error of the method of a key is returned once the value is read, as the one of a value read at once is: a syntax
// error or the end of the input after it is the error.
func TestDecodeStreamElementsKeyError(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	for _, in := range []struct {
		in  string
		new func() any
	}{
		{`{"C":{` + pad + `"bad":1,"ok":2}}`, func() any { return new(struct{ C map[elementsBadKey]int }) }},
		{`{"C":{` + pad + `"bad":1,"ok":2 x}}`, func() any { return new(struct{ C map[elementsBadKey]int }) }},
		{`{` + pad + `"bad":{},"ok":{} x}`, func() any { return new(map[elementsBadKey]any) }},
		{`{` + pad + `"bad":1,"x":2,`, func() any { return new(map[elementsBadKey]int) }},
		{`{` + pad + `"bad":1,"x":2}`, func() any { return new(map[elementsBadKey]int) }},
	} {
		var stdErr error
		{
			v := in.new()
			stdErr = stdjson.NewDecoder(strings.NewReader(in.in)).Decode(v)
		}
		for _, chunk := range []int{1000, 65536} {
			want, got := decodeAtOnceAndByChunks(in.in, chunk, in.new, func(*json.Decoder) {})
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Fatalf("%.30q, chunk %d: got %v, want %v (encoding/json: %v)", in.in, chunk, got, want, stdErr)
			}
			var stdSyntax *stdjson.SyntaxError
			var syntax *json.SyntaxError
			if errors.Is(stdErr, io.ErrUnexpectedEOF) != errors.Is(got, io.ErrUnexpectedEOF) ||
				errors.As(stdErr, &stdSyntax) != errors.As(got, &syntax) {
				t.Fatalf("%.30q, chunk %d: got %v, encoding/json %v", in.in, chunk, got, stdErr)
			}
		}
	}
}

// The index of the brackets of a long array keeps only the part not decoded yet, whatever the length of the array.
func TestDecodeStreamElementsIndexBounded(t *testing.T) {
	element := "," + strings.Repeat(" ", 61) + "{}"
	measure := func(n int) uint64 {
		r := &chunkReader{data: []byte("[{}" + strings.Repeat(element, n) + "]"), chunk: 4096}
		var v []struct{}
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		if err := json.NewDecoder(r).Decode(&v); err != nil {
			t.Fatal(err)
		}
		runtime.ReadMemStats(&after)
		if len(v) != n+1 {
			t.Fatalf("got %d elements", len(v))
		}
		return after.TotalAlloc - before.TotalAlloc
	}
	small, large := measure(1<<12), measure(1<<17)
	if large > small+(1<<20) {
		t.Fatalf("%d bytes allocated for %d elements, %d for %d", large, 1<<17, small, 1<<12)
	}
}

// The length of a slice decoded element by element counts the elements decoded before an error, never the element
// which failed. A slice decoded whole from the buffer may count it ( see sliceDecoder.Decode ): the value after an error
// is not specified.
func TestDecodeStreamElementsSliceAfterError(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	in := `[` + pad + `1,2,3,{]`
	for _, prior := range [][]int{nil, {9, 9, 9, 9, 9}} {
		v := append([]int(nil), prior...)
		if err := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 4096}).Decode(&v); err == nil {
			t.Fatal("no error")
		}
		if !reflect.DeepEqual(v, []int{1, 2, 3}) {
			t.Fatalf("prior %v: got %v", prior, v)
		}
	}
}

// A number which a read cuts is decoded once it is whole: its first digits are not written to the field, which the
// type error of the whole number leaves as it was.
func TestDecodeStreamElementsCutNumber(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	in := `{"A":1,` + pad + `"X":1e400,"B":2}`
	type target struct{ A, X, B int }
	var whole, chunked target
	werr := json.NewDecoder(bytes.NewReader([]byte(in))).Decode(&whole)
	cerr := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 1}).Decode(&chunked)
	if werr == nil || cerr == nil || werr.Error() != cerr.Error() || whole != chunked {
		t.Fatalf("got %+v, %v, want %+v, %v", chunked, cerr, whole, werr)
	}
}

// An input which ends in an integer of a value which is skipped, or read whole, ends unexpectedly, whatever the reads:
// the integer which a read cuts is read on, and the end of the input after it is the error.
func TestDecodeStreamElementsEndInInteger(t *testing.T) {
	pad := strings.Repeat(" ", 1000)
	for _, in := range []string{
		`{"unknown":[` + pad + `1,12`,
		`{"unknown":{"a":` + pad + `12`,
		`[` + pad + `1,12`,
	} {
		for _, chunk := range []int{1, 7, 512} {
			for _, target := range []func() any{
				func() any { return new(struct{ Name string }) },
				func() any { return new(json.RawMessage) },
			} {
				whole, chunked := decodeAtOnceAndByChunks(in, chunk, target, func(*json.Decoder) {})
				if chunked == nil || whole == nil || chunked.Error() != whole.Error() {
					t.Fatalf("%q by chunks of %d: got %v, want %v", in[:12], chunk, chunked, whole)
				}
			}
		}
	}
}

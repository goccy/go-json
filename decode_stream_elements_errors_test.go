package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
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
			if errors.Is(stdErr, io.ErrUnexpectedEOF) != errors.Is(got, io.ErrUnexpectedEOF) {
				t.Fatalf("%.20q, chunk %d: got %v, encoding/json %v", in, chunk, got, stdErr)
			}
		}
	}
}

// The length of a slice counts the elements decoded before an error, never the element which failed.
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

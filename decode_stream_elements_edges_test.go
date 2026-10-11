package json_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	json "github.com/goccy/go-json"
)

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

// A read which returns nothing is followed by another one: a literal is read on, whatever the reads give.
func TestDecodeStreamZeroReads(t *testing.T) {
	for _, in := range []string{`null`, `true`, `{"a":null}`, `[true,false]`, `{"a":[1,` + strings.Repeat(" ", 1000) + `null]}`} {
		var want, got any
		if err := json.Unmarshal([]byte(in), &want); err != nil {
			t.Fatal(err)
		}
		if err := json.NewDecoder(&zeroReader{data: []byte(in)}).Decode(&got); err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if !bytes.Equal(mustMarshal(t, got), mustMarshal(t, want)) {
			t.Fatalf("%q: got %v, want %v", in, got, want)
		}
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// A NUL byte of the input is an invalid character wherever a read leaves it, not the end of an object or an array.
func TestDecodeStreamElementsNulByte(t *testing.T) {
	for _, in := range []string{"{\"a\":{\"b\":\x00}}", "{\"a\":[1,\x002]}", "{\"a\":{\"b\":1,\x00}}"} {
		for _, chunk := range []int{1, 2, 3} {
			whole, chunked := decodeAtOnceAndByChunks(in, chunk, func() any { return new(struct{}) }, func(*json.Decoder) {})
			if whole == nil || chunked == nil || whole.Error() != chunked.Error() {
				t.Fatalf("%q by chunks of %d: got %v, want %v", in, chunk, chunked, whole)
			}
		}
	}
}

type edgesInner struct {
	Name string
	A    int
}

// A pointer which has a value before the decoding keeps it for a value of another kind, as for the value read at once,
// and keeps pointing to it for a value which fails, whose members decoded before the error are in it.
func TestDecodeStreamElementsPrefilledPointer(t *testing.T) {
	pad := strings.Repeat(" ", 10000)
	for i, in := range []string{`[1,` + pad + `2]`, `{"Name":"x",` + pad} {
		decode := func(r io.Reader) (*edgesInner, *edgesInner, error) {
			prior := &edgesInner{Name: "keep", A: 42}
			p := prior
			err := json.NewDecoder(r).Decode(&p)
			return prior, p, err
		}
		_, wantP, wantErr := decode(bytes.NewReader([]byte(in)))
		prior, p, err := decode(&chunkReader{data: []byte(in), chunk: 1})
		if wantErr == nil || err == nil || err.Error() != wantErr.Error() {
			t.Fatalf("%q: got %v, want %v", in[:8], err, wantErr)
		}
		if p != prior || (i == 0 && *p != *wantP) {
			t.Fatalf("%q: got %p %+v, want the pointer before, with %+v", in[:8], p, p, *wantP)
		}
	}

	// a field and an embedded pointer
	type outer struct {
		P *edgesInner
		*edgesInner
	}
	in := `{"P":[1,` + pad + `2]}`
	v := outer{P: &edgesInner{Name: "keep", A: 42}}
	prior := v.P
	if err := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 1}).Decode(&v); err == nil {
		t.Fatal("no error")
	}
	if v.P != prior || *v.P != (edgesInner{Name: "keep", A: 42}) {
		t.Fatalf("got %p %+v", v.P, v.P)
	}
}

// A syntax error is returned without a read more, wherever the last read cut a character after it: the character is
// read on only when the error is at it.
func TestDecodeStreamNoReadAfterError(t *testing.T) {
	cases := []string{
		"[\"a\x01bé\"]",
		`[1,2,x,"é"]`,
		"{\"a\":tru\x01e,\"b\":\"é\"}",
		`["\q","é"]`,
		`[1,x] ["é"]`,
		`{"a":[1,` + strings.Repeat(" ", 1000) + `x,"é"]}`,
		// an escape whose message shows two or six bytes, which have been read
		`"a\é" 1`,
		`"\u1xaé"`,
		`["a\é"]`,
		`{"a":["a\é"],` + strings.Repeat(" ", 1000) + `"b":1}`,
	}
	if !errorsShowWholeChar {
		// an invalid character which the message shows by its first byte
		cases = append(cases, `[1,é]`, `{"a":[1,`+strings.Repeat(" ", 1000)+`é]}`, `{"a":é}`)
	}
	for _, in := range cases {
		for _, tokens := range []bool{false, true} {
			// the input up to the first byte of é, after which the writer waits
			cut := strings.Index(in, "é") + 1
			pr, pw := io.Pipe()
			go func() { _, _ = pw.Write([]byte(in[:cut])) }()
			done := make(chan error, 1)
			go func() {
				dec := json.NewDecoder(pr)
				if !tokens {
					var v any
					done <- dec.Decode(&v)
					return
				}
				for {
					if _, err := dec.Token(); err != nil {
						done <- err
						return
					}
				}
			}()
			select {
			case err := <-done:
				if err == nil {
					t.Fatalf("%q: no error", in)
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("%q ( tokens: %v ): waits for a read after the error", in, tokens)
			}
			pw.Close()
		}
	}
}

// An element of a slice which has an error which is not one of the grammar, as an unknown field, is in the slice, as
// in the slice of the value read at once.
func TestDecodeStreamElementsSliceUnknownField(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	in := `[{"a":1},` + pad + `{"a":2},{"x":null},{"a":3}]`
	decode := func(r io.Reader) ([]struct{ A int }, error) {
		var v []struct{ A int }
		dec := json.NewDecoder(r)
		dec.DisallowUnknownFields()
		err := dec.Decode(&v)
		return v, err
	}
	want, wantErr := decode(bytes.NewReader([]byte(in)))
	for _, chunk := range []int{1, 4096} {
		got, err := decode(&chunkReader{data: []byte(in), chunk: chunk})
		if err == nil || wantErr == nil || err.Error() != wantErr.Error() || len(got) != len(want) {
			t.Fatalf("by chunks of %d: got %v, %v, want %v, %v", chunk, got, err, want, wantErr)
		}
	}
}

package json_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
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
	for _, in := range []string{`null`, `true`, `false`, `12 `, `"s"`, `{"a":null}`, `[true,false]`, `{"a":[1,` + strings.Repeat(" ", 1000) + `null]}`} {
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
			whole, chunked := decodeAtOnceAndByChunks(in, chunk, func() any { return new(struct{}) }, false)
			if msg := sameError(whole, chunked, 0); whole == nil || msg != "" {
				t.Fatalf("%q by chunks of %d: %s", in, chunk, msg)
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
		wantP := &edgesInner{Name: "keep", A: 42}
		wantErr := unmarshalReference([]byte(in), &wantP, false, false)
		if json.IsAtEnd(wantErr) {
			wantErr = io.ErrUnexpectedEOF
		}
		prior := &edgesInner{Name: "keep", A: 42}
		p := prior
		err := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 1}).Decode(&p)
		if msg := sameError(wantErr, err, 0); wantErr == nil || msg != "" {
			t.Fatalf("%q: %s", in[:8], msg)
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

// A syntax error is returned without a read more, wherever the last read cut a character, after the error or at it:
// the message of an error at a character which has been read in part quotes its first byte.
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
		// an invalid character
		`[1,é]`,
		`{"a":[1,` + strings.Repeat(" ", 1000) + `é]}`,
		`{"a":é}`,
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

// An invalid character of several bytes which the writer stops in is a syntax error without a wait for the rest of
// it: the error of the bytes written, as Unmarshal of them returns it.
func TestDecodeStreamCutInvalidCharacter(t *testing.T) {
	for _, c := range []string{"é", "あ", "😀"} {
		for n := 1; n < len(c); n++ {
			for _, prefix := range []string{"", `[1,`, `{"a":`, `{"a":[1,` + strings.Repeat(" ", 1000)} {
				in := prefix + c[:n]
				want := json.Unmarshal([]byte(in), new(any))
				pr, pw := io.Pipe()
				go func() { _, _ = pw.Write([]byte(in)) }()
				done := make(chan error, 1)
				go func() {
					var v any
					done <- json.NewDecoder(pr).Decode(&v)
				}()
				select {
				case err := <-done:
					if want == nil || err == nil || err.Error() != want.Error() || sameError(want, err, 0) != "" {
						t.Fatalf("%.40q: got %v, want %v", in, err, want)
					}
				case <-time.After(2 * time.Second):
					t.Fatalf("%.40q: waits for the rest of the character", in)
				}
				pw.Close()
			}
		}
	}
}

// An element of a slice which has an error which is not one of the grammar, as an unknown field, is in the slice, as
// in the slice of the value read at once.
func TestDecodeStreamElementsSliceUnknownField(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	in := `[{"a":1},` + pad + `{"a":2},{"x":null},{"a":3}]`
	var want []struct{ A int }
	wantErr := unmarshalReference([]byte(in), &want, true, false)
	for _, chunk := range []int{1, 4096} {
		var got []struct{ A int }
		dec := json.NewDecoder(&chunkReader{data: []byte(in), chunk: chunk})
		dec.DisallowUnknownFields()
		err := dec.Decode(&got)
		if msg := sameError(wantErr, err, 0); wantErr == nil || msg != "" || len(got) != len(want) {
			t.Fatalf("by chunks of %d: got %v, %v, want %v, %v", chunk, got, err, want, wantErr)
		}
	}
}

// The strings of the elements of an array decoded element by element are the ones of the input, after the buffer they
// were decoded from is read into again: they are copied out of it, or, under DecodeNoCopyString, the buffer is not read
// into again.
func TestDecodeStreamElementsStringsKept(t *testing.T) {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < 2000; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"x":%d,"y":"value-%d-abcdefghij"}`, i, i)
	}
	b.WriteByte(']')
	in := b.String()
	type elem struct {
		X int
		Y string
	}
	var want []elem
	if err := json.Unmarshal([]byte(in), &want); err != nil {
		t.Fatal(err)
	}
	for _, chunk := range []int{1, 100, 512} {
		for _, opts := range [][]json.DecodeOptionFunc{nil, {json.DecodeNoCopyString()}} {
			var got []elem
			if err := json.NewDecoder(&chunkReader{data: []byte(in), chunk: chunk}).DecodeWithOption(&got, opts...); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("by %d, %d options: got %.200s", chunk, len(opts), fmt.Sprint(got))
			}
		}
	}
}

// EmbeddedPointerInner is promoted through an embedded pointer, which a field of it allocates.
type EmbeddedPointerInner struct {
	Q struct {
		A int `json:"a"`
	} `json:"q"`
}

type embeddedPointerOuter struct {
	*EmbeddedPointerInner
	Name string
}

// unsettableInner is promoted through an embedded pointer to an unexported struct, which can't be allocated.
type unsettableInner struct {
	Z int
}

type unsettableOuter struct {
	*unsettableInner
	Name string
}

// The values decoded element by element are the ones of Unmarshal, with its options, and so are their errors: a
// field promoted through an embedded pointer, an array shorter than the Go array, a map key which is not of its
// type, and the fields of FirstWin, which skips the rest of an object once every field is decoded.
func TestDecodeStreamElementsValuesOfUnmarshal(t *testing.T) {
	pad := strings.Repeat(" ", 1000)
	firstWin := []json.DecodeOptionFunc{json.DecodeFieldPriorityFirstWin()}
	for _, c := range []struct {
		in       string
		newValue func() any
		opts     []json.DecodeOptionFunc
		disallow bool
	}{
		{`{"q":{` + pad + `"a":1},"Name":"x"}`, func() any { return new(embeddedPointerOuter) }, nil, false},
		{`[` + pad + `1]`, func() any { v := [3]any{7, 8, 9}; return &v }, nil, false},
		{`{` + pad + `"x":1,"2":3}`, func() any { return new(map[int]int) }, nil, false},
		{`{` + pad + `"Z":[1,` + pad + `2],"Name":"x"}`, func() any { return new(unsettableOuter) }, nil, false},
		{`{` + pad + `"Z":1,"Name":"x","unknown":1}`, func() any { return new(unsettableOuter) }, firstWin, true},
		{`{` + pad + `"Name":"x","unknown":1}`, func() any { return new(struct{ Name string }) }, firstWin, true},
	} {
		want := c.newValue()
		wantErr := json.UnmarshalLikeDecoder([]byte(c.in), want, c.disallow, false, c.opts...)
		for _, chunk := range []int{1, 512, 4096} {
			got := c.newValue()
			dec := json.NewDecoder(&chunkReader{data: []byte(c.in), chunk: chunk})
			if c.disallow {
				dec.DisallowUnknownFields()
			}
			err := dec.DecodeWithOption(got, c.opts...)
			if msg := sameError(wantErr, err, 0); msg != "" {
				t.Fatalf("%.40q into %T by %d: %s", c.in, got, chunk, msg)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%.40q into %T by %d: got %+v, want %+v", c.in, got, chunk, got, want)
			}
		}
	}
}

// The value of a field which can't be set is skipped by reads of the size of the buffer, and its error is the one of
// Unmarshal.
func TestDecodeStreamUnsettableFieldReads(t *testing.T) {
	in := `{"Z":[` + strings.Repeat("1,", 100000) + `1],"Name":"x"}`
	var want unsettableOuter
	wantErr := json.Unmarshal([]byte(in), &want)
	r := &roomReader{Reader: strings.NewReader(in)}
	var got unsettableOuter
	err := json.NewDecoder(r).Decode(&got)
	if msg := sameError(wantErr, err, 0); wantErr == nil || msg != "" || got != want {
		t.Fatalf("got %+v, %v, want %+v, %v", got, err, want, wantErr)
	}
	for _, room := range r.rooms {
		if room > r.rooms[0] {
			t.Fatalf("a read of %d bytes, the buffer has %d", room, r.rooms[0])
		}
	}
}

// A string member of a value decoded element by element is the one of Unmarshal, whatever the reads and whatever the
// string: a long one of ASCII, which is taken as it is scanned, one with an escape, one with bytes which are not
// ASCII, valid or not, and one with a control byte, which is an error.
func TestDecodeStreamElementsStrings(t *testing.T) {
	pad := strings.Repeat(" ", 1000)
	ascii := strings.Repeat("abcdefghij", 70)
	type record struct {
		ID   int
		Text string
	}
	for _, text := range []string{
		`""`,
		`"` + ascii + `"`,
		`"` + ascii + `\n` + ascii + `"`,
		`"` + ascii + `\u3042"`,
		`"` + ascii + `あ` + ascii + `"`,
		`"` + ascii + "\xff" + `"`,
		`"` + ascii + "\x01" + `"`,
	} {
		in := `[` + pad + `{"ID":1,"Text":` + text + `},{"ID":2,"Text":"x"}]`
		for _, opts := range [][]json.DecodeOptionFunc{nil, {json.DecodeNoCopyString()}} {
			var want []record
			wantErr := json.UnmarshalLikeDecoder([]byte(in), &want, false, false, opts...)
			for _, chunk := range []int{1, 64, 512, 4096} {
				var got []record
				err := json.NewDecoder(&chunkReader{data: []byte(in), chunk: chunk}).DecodeWithOption(&got, opts...)
				if msg := sameError(wantErr, err, 0); msg != "" {
					t.Fatalf("%.30q by %d, %d options: %s", text, chunk, len(opts), msg)
				}
				if wantErr == nil && !reflect.DeepEqual(got, want) {
					t.Fatalf("%.30q by %d, %d options: got %.200s, want %.200s", text, chunk, len(opts), fmt.Sprint(got), fmt.Sprint(want))
				}
			}
		}
	}
}

// The values decoded from the buffer of a stream refer to it, which is not read into again: values decoded whole
// between values decoded element by element keep their strings, numbers, keys and interface values, whatever the
// reads.
func TestDecodeStreamBufferKeptForTopValues(t *testing.T) {
	type small struct {
		S string
		N json.Number
		M map[string]string
		A any
	}
	long := `[` + strings.Repeat(" ", 600) + strings.Repeat(`"`+strings.Repeat("a", 100)+`",`, 30) + `"z"]`
	value := `{"S":"keep-s","N":12345,"M":{"key-m":"val-m"},"A":"keep-a"}`
	want := small{S: "keep-s", N: "12345", M: map[string]string{"key-m": "val-m"}, A: "keep-a"}
	in := strings.Repeat(long+value, 8) + long
	for _, chunk := range []int{1, 64, 512, 4096} {
		dec := json.NewDecoder(&chunkReader{data: []byte(in), chunk: chunk})
		dec.UseNumber()
		var kept []small
		for i := 0; i < 17; i++ {
			if i%2 == 0 {
				var v []string
				if err := dec.Decode(&v); err != nil || len(v) != 31 {
					t.Fatalf("by %d: value %d: %d strings, %v", chunk, i, len(v), err)
				}
				continue
			}
			var v small
			if err := dec.Decode(&v); err != nil {
				t.Fatalf("by %d: value %d: %v", chunk, i, err)
			}
			kept = append(kept, v)
		}
		for i, v := range kept {
			if !reflect.DeepEqual(v, want) {
				t.Fatalf("by %d: value %d is %+v after the values after it, want %+v", chunk, 2*i+1, v, want)
			}
		}
	}
}

// The strings of a value decoded element by element under DecodeNoCopyString refer to the buffer, which is not read
// into again, also after an option clears it for the next values: a value long or short after one which made the
// buffer which the stream reuses.
func TestDecodeStreamBufferKeptForNoCopyStrings(t *testing.T) {
	long := `[` + strings.Repeat(`"`+strings.Repeat("a", 100)+`",`, 30) + `"z"]`
	clear := func(o *json.DecodeOption) { o.Flags = 0 }
	for _, value := range []string{long, `["short-value-under-nocopy"]`} {
		in := long + value + strings.Repeat(long, 3)
		var want []string
		if err := json.Unmarshal([]byte(value), &want); err != nil {
			t.Fatal(err)
		}
		for _, chunk := range []int{1, 7, 64, 512, 4096} {
			dec := json.NewDecoder(&chunkReader{data: []byte(in), chunk: chunk})
			var first, kept []string
			if err := dec.Decode(&first); err != nil {
				t.Fatal(err)
			}
			if err := dec.DecodeWithOption(&kept, json.DecodeNoCopyString()); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				var v []string
				if err := dec.DecodeWithOption(&v, clear); err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(kept, want) {
				t.Fatalf("%.30q by %d: the strings decoded under DecodeNoCopyString changed: %.100q", value, chunk, strings.Join(kept, ","))
			}
		}
	}
}

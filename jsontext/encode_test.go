package jsontext_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/goccy/go-json/jsontext"
)

// TestWriteValueInAvailableBuffer writes values twice which are built in the AvailableBuffer of the encoder, as
// its documentation shows, to a writer and to a bytes.Buffer, under options whose output is longer than the
// value: the value must not change, and the output must be the one of a copy of it.
func TestWriteValueInAvailableBuffer(t *testing.T) {
	for _, opts := range [][]jsontext.Options{
		nil,
		{jsontext.Multiline(true)},
		{jsontext.EscapeForHTML(true), jsontext.SpaceAfterComma(true), jsontext.SpaceAfterColon(true)},
	} {
		for _, in := range []string{
			`[1,2,"<>&",{"a":[true,false,null]}]`,
			`{ "a" : 1 , "b" : [ 1 , 2 , "\/" ] }`,
			`"<<<>>>&&&"`,
		} {
			var want bytes.Buffer
			we := jsontext.NewEncoder(&want, opts...)
			we.WriteToken(jsontext.BeginArray)
			we.WriteValue(jsontext.Value(in))
			we.WriteValue(jsontext.Value(in))
			we.WriteToken(jsontext.EndArray)
			// a writer whose output the encoder buffers, and a bytes.Buffer, whose room the encoder writes to
			for _, w := range []string{"writer", "bytes.Buffer"} {
				var got bytes.Buffer
				got.Grow(1024)
				e := jsontext.NewEncoder(struct{ io.Writer }{&got}, opts...)
				if w == "bytes.Buffer" {
					e = jsontext.NewEncoder(&got, opts...)
				}
				e.WriteToken(jsontext.BeginArray) // the output which is not written yet
				b := append(e.AvailableBuffer(), in...)
				if err := e.WriteValue(b); err != nil {
					t.Fatal(err)
				}
				if string(b) != in {
					t.Fatalf("WriteValue(%q) in AvailableBuffer, to a %s, changed it to %q", in, w, b)
				}
				if err := e.WriteValue(b); err != nil {
					t.Fatal(err)
				}
				e.WriteToken(jsontext.EndArray)
				if got.String() != want.String() {
					t.Errorf("WriteValue(%q) in AvailableBuffer, to a %s = %q, want %q", in, w, got.String(), want.String())
				}
			}
		}
	}
}

// TestWriteValueInBytesBufferSpace writes top-level values built in the free space of the bytes.Buffer which the
// encoder writes to, where the output goes before the value is read.
func TestWriteValueInBytesBufferSpace(t *testing.T) {
	for _, in := range []string{
		`[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12]`,
		`{"a": 1, "bb": [true, false, null], "c":  "dddddddddd" }`,
		`[ 1,  2,  3,  4,  5 ]`,
	} {
		var want bytes.Buffer
		jsontext.NewEncoder(&want).WriteValue(jsontext.Value(in))
		var got bytes.Buffer
		got.Grow(1024)
		if err := jsontext.NewEncoder(&got).WriteValue(append(got.AvailableBuffer(), in...)); err != nil {
			t.Fatalf("WriteValue(%q) in the free space of the bytes.Buffer: %v", in, err)
		}
		if got.String() != want.String() {
			t.Errorf("WriteValue(%q) in the free space of the bytes.Buffer = %q, want %q", in, got.String(), want.String())
		}
	}
}

// TestResetCopiedEncoder resets with the same options an encoder which was copied after it was used, as an
// element of a slice which grew is: its values go to its own stack.
func TestResetCopiedEncoder(t *testing.T) {
	type holder struct{ enc jsontext.Encoder }
	var b bytes.Buffer
	hs := []holder{{}}
	hs[0].enc.Reset(&b)
	hs = append(hs, holder{}) // the first element is copied
	e := &hs[0].enc
	e.Reset(&b)
	e.WriteToken(jsontext.BeginObject)
	for _, v := range []string{`"a"`, `1`, `"b"`, `{"c":2}`} {
		if err := e.WriteValue(jsontext.Value(v)); err != nil {
			t.Fatalf("WriteValue(%s): %v", v, err)
		}
	}
	if err := e.WriteValue(jsontext.Value(`"a"`)); err == nil {
		t.Errorf("WriteValue of a duplicate name: no error")
	}
	e.WriteToken(jsontext.EndObject)
	if want := "{\"a\":1,\"b\":{\"c\":2}}\n"; b.String() != want {
		t.Errorf("output = %q, want %q", b.String(), want)
	}
}

// TestAppendFloatIntegers checks integers below 1e21, which are formatted by their shortest digits without an
// exponent, against strconv, in both precisions.
func TestAppendFloatIntegers(t *testing.T) {
	values := []float64{1, 1 << 24, 1<<24 + 2, 123456789, 1e15, 1<<53 - 1, 1 << 53, 1<<53 + 2, 1e20}
	r := rand.New(rand.NewPCG(1, 2))
	for range 10000 {
		values = append(values, math.Trunc(math.Ldexp(r.Float64(), r.IntN(70))))
	}
	for _, v := range values {
		for _, f := range []float64{v, -v, v - 1, v + 1} {
			for _, bits := range []int{32, 64} {
				rounded := f
				if bits == 32 {
					rounded = float64(float32(f))
				}
				if math.Abs(rounded) >= 1e21 {
					continue
				}
				want := strconv.AppendFloat(nil, rounded, 'f', -1, bits)
				if got := jsontext.AppendFloat(nil, f, bits); !bytes.Equal(got, want) {
					t.Errorf("AppendFloat(%v, %d) = %s, want %s", f, bits, got, want)
				}
			}
		}
	}
}

// TestResetCopiedCoders resets a copy of a Decoder and of an Encoder, which reads or writes something else, in turn
// with the coder it was copied from: each one reads and writes as a coder which was not copied. (The coders of
// encoding/json/jsontext share their buffers with their copies, which corrupts the reads and writes of both.)
func TestResetCopiedCoders(t *testing.T) {
	const in, in2 = `{"a":[1,{"b":2}],"c":"d"} [3]`, `[true,{"x":null}]`
	readAll := func(d *jsontext.Decoder) string {
		var s string
		for range 12 {
			tok, err := d.ReadToken()
			s += fmt.Sprintf("[%s %v %q] ", tok, err, d.StackPointer())
		}
		return s
	}
	ref := jsontext.NewDecoder(strings.NewReader(in))
	ref.ReadToken()
	ref.ReadToken()
	d := jsontext.NewDecoder(strings.NewReader(in))
	d.ReadToken()
	d.ReadToken()
	c := *d
	c.Reset(strings.NewReader(in2))
	var got, gotCopy string
	for range 12 {
		tok, err := d.ReadToken()
		got += fmt.Sprintf("[%s %v %q] ", tok, err, d.StackPointer())
		tok, err = c.ReadToken()
		gotCopy += fmt.Sprintf("[%s %v %q] ", tok, err, c.StackPointer())
	}
	if want := readAll(ref); got != want {
		t.Errorf("a Decoder whose copy was reset:\ngot:  %s\nwant: %s", got, want)
	}
	if want := readAll(jsontext.NewDecoder(strings.NewReader(in2))); gotCopy != want {
		t.Errorf("a copy of a Decoder, which was reset:\ngot:  %s\nwant: %s", gotCopy, want)
	}

	var b, b2 bytes.Buffer
	e := jsontext.NewEncoder(&b)
	e.WriteToken(jsontext.BeginObject)
	e.WriteToken(jsontext.String("a"))
	ec := *e
	ec.Reset(&b2)
	for _, v := range []string{`[1]`, `"b"`, `{"c":2}`} {
		if err := e.WriteValue(jsontext.Value(v)); err != nil {
			t.Fatal(err)
		}
		if err := ec.WriteValue(jsontext.Value(v)); err != nil {
			t.Fatal(err)
		}
	}
	e.WriteToken(jsontext.EndObject)
	if want := "{\"a\":[1],\"b\":{\"c\":2}}\n"; b.String() != want {
		t.Errorf("output of an Encoder whose copy was reset = %q, want %q", b.String(), want)
	}
	if want := "[1]\n\"b\"\n{\"c\":2}\n"; b2.String() != want {
		t.Errorf("output of a copy of an Encoder, which was reset = %q, want %q", b2.String(), want)
	}
}

// TestInterleavedCopiedCoders uses a Decoder or an Encoder and a copy of it in turn, in random orders, which
// encoding/json/jsontext doesn't define either: their reads and writes may be wrong, but none panics.
func TestInterleavedCopiedCoders(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	values := []string{`"a"`, `"b"`, `1`, `[1,2]`, `{"x":1}`, `{"a":{"b":[1]}}`, `"a"`, `true`, `{`, `"`}
	tokens := []jsontext.Token{jsontext.BeginObject, jsontext.EndObject, jsontext.BeginArray, jsontext.EndArray,
		jsontext.String("a"), jsontext.String("b"), jsontext.String("a"), jsontext.Int(1), jsontext.Null}
	const in = `{"a":[1,{"b":2,"c":[3,4]}],"d":{"e":"f","g":[{"h":null}]}} [5,6] {"i":true}`
	for range 3000 {
		opts := []jsontext.Options{jsontext.AllowDuplicateNames(r.IntN(2) == 0), jsontext.Multiline(r.IntN(2) == 0),
			jsontext.ReorderRawObjects(r.IntN(4) == 0)}
		var b bytes.Buffer
		var w io.Writer = &b
		if r.IntN(2) == 0 {
			w = struct{ io.Writer }{&b}
		}
		var rd io.Reader = strings.NewReader(in)
		switch r.IntN(3) {
		case 1:
			rd = bytes.NewBufferString(in)
		case 2:
			rd = iotest.OneByteReader(strings.NewReader(in))
		}
		encoders := []*jsontext.Encoder{jsontext.NewEncoder(w, opts...)}
		decoders := []*jsontext.Decoder{jsontext.NewDecoder(rd, opts...)}
		copyAt := r.IntN(20)
		for step := range 40 {
			if step == copyAt {
				ec, dc := *encoders[0], *decoders[0]
				encoders, decoders = append(encoders, &ec), append(decoders, &dc)
			}
			e, d := encoders[r.IntN(len(encoders))], decoders[r.IntN(len(decoders))]
			func() {
				defer func() {
					if p := recover(); p != nil {
						t.Fatalf("panic at step %d: %v\n%s", step, p, debug.Stack())
					}
				}()
				switch r.IntN(9) {
				case 0, 1:
					e.WriteToken(tokens[r.IntN(len(tokens))])
				case 2, 3:
					e.WriteValue(jsontext.Value(values[r.IntN(len(values))]))
				case 4:
					d.ReadToken()
				case 5:
					d.ReadValue()
				case 6:
					d.SkipValue()
				case 7:
					d.PeekKind()
				default:
					switch r.IntN(4) {
					case 0:
						e.Reset(w, opts...)
					case 1:
						d.Reset(strings.NewReader(in), opts...)
					case 2:
						e.WriteValue(append(e.AvailableBuffer(), values[r.IntN(len(values))]...))
					}
				}
				_, _ = e.StackPointer(), d.StackPointer()
				for i := range e.StackDepth() + 1 {
					e.StackIndex(i)
				}
				for i := range d.StackDepth() + 1 {
					d.StackIndex(i)
				}
			}()
		}
	}
}

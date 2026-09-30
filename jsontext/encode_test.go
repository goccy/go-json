package jsontext_test

import (
	"bytes"
	"io"
	"math"
	"math/rand/v2"
	"strconv"
	"testing"

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

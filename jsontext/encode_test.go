package jsontext_test

import (
	"bytes"
	"math"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/goccy/go-json/jsontext"
)

// TestWriteValueInAvailableBuffer writes values which are built in the buffer of AvailableBuffer, as its
// documentation shows, under options whose output is longer than the value: the output must be the one of a
// copy of the value.
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
			var got bytes.Buffer
			got.Grow(1024)
			e := jsontext.NewEncoder(&got, opts...)
			if err := e.WriteValue(append(e.AvailableBuffer(), in...)); err != nil {
				t.Fatal(err)
			}
			var want bytes.Buffer
			if err := jsontext.NewEncoder(&want, opts...).WriteValue(jsontext.Value(in)); err != nil {
				t.Fatal(err)
			}
			if got.String() != want.String() {
				t.Errorf("WriteValue(%q) in AvailableBuffer = %q, want %q", in, got.String(), want.String())
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

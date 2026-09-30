//go:build go1.27

package jsontext

import (
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

// The tests here compare appendFloat with strconv of Go 1.27, which encoding/json/jsontext formats the floats
// by: strconv of another version of Go may write other digits in rare cases ( Go 1.26 writes 2^-12 as a
// float32 as 0.00024414063, where Go 1.25 and Go 1.27 write 0.00024414062 ).

// appendFloatByStrconv is how encoding/json/jsontext writes a float.
func appendFloatByStrconv(dst []byte, f float64, size int) []byte {
	if size == 32 {
		f = float64(float32(f))
	}
	abs := math.Abs(f)
	format := byte('f')
	if abs != 0 && (size == 64 && (abs < 1e-6 || abs >= 1e21) || size == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21)) {
		format = 'e'
	}
	dst = strconv.AppendFloat(dst, f, format, -1, size)
	if format == 'e' {
		if n := len(dst); dst[n-4] == 'e' && dst[n-2] == '0' {
			dst[n-2] = dst[n-1]
			dst = dst[:n-1]
		}
	}
	return dst
}

func checkFloat(t *testing.T, f float64, size int) bool {
	t.Helper()
	got, want := appendFloat(nil, f, size), appendFloatByStrconv(nil, f, size)
	if string(got) != string(want) {
		t.Errorf("appendFloat(%v (%#x), %d) = %s, want %s", f, math.Float64bits(f), size, got, want)
		return false
	}
	return true
}

func TestAppendFloatStrconv(t *testing.T) {
	var floats []float64
	// the powers of two and of ten, and the floats next to them
	for e := -1075; e <= 1024; e++ {
		floats = append(floats, math.Ldexp(1, e))
	}
	for e := -325; e <= 309; e++ {
		f, _ := strconv.ParseFloat("1e"+strconv.Itoa(e), 64)
		floats = append(floats, f)
	}
	for _, f := range append([]float64(nil), floats...) {
		floats = append(floats, math.Nextafter(f, 0), math.Nextafter(f, math.Inf(1)))
	}
	floats = append(floats, 0, math.SmallestNonzeroFloat64, math.MaxFloat64, math.SmallestNonzeroFloat32,
		math.MaxFloat32, 1e21, 1e-6, 1e-7, 123456789, 0.1, 0.2, 0.3, 1.0/3, 2.0/3, 5e-324, 1.7976931348623157e308)
	for _, f := range floats {
		if math.IsInf(f, 0) {
			continue
		}
		for _, f := range []float64{f, -f} {
			checkFloat(t, f, 64)
			if f32 := float32(f); !math.IsInf(float64(f32), 0) {
				checkFloat(t, float64(f32), 32)
				checkFloat(t, float64(math.Nextafter32(f32, 0)), 32)
				if next := math.Nextafter32(f32, float32(math.Inf(1))); !math.IsInf(float64(next), 0) {
					checkFloat(t, float64(next), 32)
				}
			}
		}
	}
	// floats of random bits, and of random bits in their significand only, which are near the integers
	r := rand.New(rand.NewPCG(1, 2))
	n := 1 << 20
	if testing.Short() {
		n = 1 << 14
	}
	for range n {
		if f := math.Float64frombits(r.Uint64()); !math.IsNaN(f) && !math.IsInf(f, 0) {
			checkFloat(t, f, 64)
		}
		if f := math.Float32frombits(r.Uint32()); !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0) {
			checkFloat(t, float64(f), 32)
		}
		checkFloat(t, math.Float64frombits(0x4330000000000000|r.Uint64()>>12), 64)
		checkFloat(t, math.Float64frombits(uint64(r.IntN(0x7ff))<<52|r.Uint64()>>61), 64)
	}
}

// TestAppendFloatStrconvMany compares every float32, and as many float64s of random bits, which takes minutes:
// it runs where JSONTEXT_MANY_FLOATS is set.
func TestAppendFloatStrconvMany(t *testing.T) {
	if os.Getenv("JSONTEXT_MANY_FLOATS") == "" {
		t.Skip("JSONTEXT_MANY_FLOATS is not set")
	}
	var wg sync.WaitGroup
	workers := runtime.GOMAXPROCS(0)
	var failed sync.Mutex
	for w := range workers {
		wg.Go(func() {
			var got, want []byte
			check := func(f float64, size int) {
				if math.IsNaN(f) || math.IsInf(f, 0) {
					return
				}
				got, want = appendFloat(got[:0], f, size), appendFloatByStrconv(want[:0], f, size)
				if string(got) != string(want) {
					failed.Lock()
					t.Errorf("appendFloat(%v (%#x), %d) = %s, want %s", f, math.Float64bits(f), size, got, want)
					failed.Unlock()
				}
			}
			r := rand.New(rand.NewPCG(uint64(w), 3))
			for b := uint64(w); b < 1<<32; b += uint64(workers) {
				check(float64(math.Float32frombits(uint32(b))), 32)
				check(math.Float64frombits(r.Uint64()), 64)
			}
		})
	}
	wg.Wait()
}

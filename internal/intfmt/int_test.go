package intfmt

import (
	"math"
	"math/rand/v2"
	"strconv"
	"testing"
)

// AppendInt and AppendUint write the integers as strconv does, around every power of 10 and at the limits, after
// the bytes of the buffer and into a buffer of any capacity.
func TestAppendIntAndUint(t *testing.T) {
	var values []uint64
	for _, p := range pow10 {
		values = append(values, p-1, p, p+1, -(p - 1), -p, -(p + 1))
	}
	values = append(values, 0, math.MaxInt64, math.MaxUint64, 1<<63, 1<<63+1)
	r := rand.New(rand.NewPCG(1, 2))
	for range 100000 {
		values = append(values, r.Uint64()>>r.IntN(64))
	}
	for _, v := range values {
		for _, prefix := range [][]byte{nil, []byte("x"), make([]byte, 3, 4)} {
			got := AppendUint(append([]byte(nil), prefix...), v)
			if want := string(prefix) + strconv.FormatUint(v, 10); string(got) != want {
				t.Fatalf("AppendUint(%d) = %q, want %q", v, got, want)
			}
			got = AppendInt(append([]byte(nil), prefix...), int64(v))
			if want := string(prefix) + strconv.FormatInt(int64(v), 10); string(got) != want {
				t.Fatalf("AppendInt(%d) = %q, want %q", int64(v), got, want)
			}
		}
	}
}

func TestDecimalDigits(t *testing.T) {
	for i, p := range pow10 {
		if got := decimalDigits(p); got != i+1 {
			t.Fatalf("digits of %d: got %d, want %d", p, got, i+1)
		}
		if p > 1 {
			if got := decimalDigits(p - 1); got != i {
				t.Fatalf("digits of %d: got %d, want %d", p-1, got, i)
			}
		}
	}
	if got := decimalDigits(0); got != 1 {
		t.Fatalf("digits of 0: got %d", got)
	}
	if got := decimalDigits(math.MaxUint64); got != 20 {
		t.Fatalf("digits of the largest uint64: got %d", got)
	}
}

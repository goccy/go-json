package encoder

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
	"unsafe"
)

// The integers of every size are written as strconv writes them, around every power of 10 and at the limits of
// their sizes, after the bytes of the buffer and into a buffer of any capacity.
func TestAppendIntegers(t *testing.T) {
	var values []uint64
	for p := uint64(1); ; p *= 10 {
		values = append(values, p-1, p, p+1)
		if p > math.MaxUint64/10 {
			break
		}
	}
	values = append(values, 0, math.MaxInt8, math.MaxUint8, math.MaxInt16, math.MaxUint16, math.MaxInt32,
		math.MaxUint32, math.MaxInt64, math.MaxUint64, 1<<63)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 10000; i++ {
		values = append(values, r.Uint64()>>uint(r.Intn(64)))
	}
	for _, v := range values {
		for _, size := range []uint8{8, 16, 32, 64} {
			code := &Opcode{NumBitSize: size}
			mask := numMask(size)
			// the value of the size as a signed and as an unsigned integer
			signed := int64(v&mask) << (64 - size) >> (64 - size)
			for _, prefix := range [][]byte{nil, []byte("x"), make([]byte, 3, 4)} {
				buf := append([]byte(nil), prefix...)
				got := AppendInt(nil, buf, valueOfSize(v, size), code)
				if want := string(prefix) + strconv.FormatInt(signed, 10); string(got) != want {
					t.Fatalf("AppendInt of %d as int%d: got %q, want %q", v, size, got, want)
				}
				buf = append([]byte(nil), prefix...)
				got = AppendUint(nil, buf, valueOfSize(v, size), code)
				if want := string(prefix) + strconv.FormatUint(v&mask, 10); string(got) != want {
					t.Fatalf("AppendUint of %d as uint%d: got %q, want %q", v, size, got, want)
				}
			}
		}
	}
}

// valueOfSize returns a pointer to the lowest size bits of v as a variable of that size, as the encoder is given
// an integer of that size: they are the first bytes of v only on a little-endian machine.
func valueOfSize(v uint64, size uint8) unsafe.Pointer {
	switch size {
	case 8:
		u := uint8(v)
		return unsafe.Pointer(&u)
	case 16:
		u := uint16(v)
		return unsafe.Pointer(&u)
	case 32:
		u := uint32(v)
		return unsafe.Pointer(&u)
	}
	return unsafe.Pointer(&v)
}

func BenchmarkAppendInt(b *testing.B) {
	values := []int64{0, 7, 42, -6, 486, 123, -456, 98765, 123456, -12345678, 1234567890, -9876543210123, math.MinInt64}
	code := &Opcode{NumBitSize: 64}
	buf := make([]byte, 0, 64)
	for _, v := range values {
		v := v
		b.Run(strconv.FormatInt(v, 10), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf = AppendInt(nil, buf[:0], unsafe.Pointer(&v), code)
			}
		})
	}
}

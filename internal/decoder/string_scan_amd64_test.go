package decoder

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

func TestIndexStringSpecialAVX2(t *testing.T) {
	// The index and the non-ASCII bytes before it are the ones of a scan byte by byte, for every length and
	// every position of the special byte, including in the last block, which overlaps the previous one.
	if !runtime.HasAVX2 {
		t.Skip("AVX2 is not supported")
	}
	r := rand.New(rand.NewSource(1))
	plain := []byte("azAZ09 !#~\x7f\x80\xe3\xff")
	special := []byte{'"', '\\', 0x00, 0x1f, '\n'}
	for n := 32; n <= 200; n++ {
		for trial := 0; trial < 60; trial++ {
			buf := make([]byte, n+1)
			for i := 0; i < n; i++ {
				buf[i] = plain[r.Intn(len(plain))]
				if r.Intn(4) != 0 && buf[i] >= 0x80 {
					buf[i] = 'a'
				}
			}
			if trial%3 != 0 {
				buf[r.Intn(n)] = special[r.Intn(len(special))]
			}
			wantIndex, wantHigh := n, false
			for i := 0; i < n; i++ {
				c := buf[i]
				if c == '"' || c == '\\' || c < 0x20 {
					wantIndex = i
					break
				}
				if c >= 0x80 {
					wantHigh = true
				}
			}
			index, high := indexStringSpecialAVX2(unsafe.Pointer(&buf[0]), n)
			if index != wantIndex || (high != 0) != wantHigh {
				t.Fatalf("%q: got %d %v, want %d %v", buf[:n], index, high != 0, wantIndex, wantHigh)
			}
		}
	}
}

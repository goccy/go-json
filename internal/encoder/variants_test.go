package encoder

import (
	"bytes"
	"math/rand"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"unsafe"
)

// The candidates of the optimizations, which are measured against what is used now on the machine of the CI:
// amd64 can't be measured on the machine they are developed on.

//go:noinline
func variantReserve(b []byte, n int) []byte {
	grown := make([]byte, len(b), 2*cap(b)+n)
	copy(grown, b)
	return grown
}

const variantKeyChunk = 16

func variantPaddedKey(key string) string {
	buf := make([]byte, (len(key)/variantKeyChunk+1)*variantKeyChunk)
	copy(buf, key)
	return unsafe.String(unsafe.SliceData(buf), len(key))
}

// appendKeyByChunks copies the key, whose memory is padded, by chunks of a fixed size instead of memmove.
func appendKeyByChunks(b []byte, key string) []byte {
	n := len(b)
	if cap(b)-n < len(key)+variantKeyChunk {
		b = variantReserve(b, len(key)+variantKeyChunk)
	}
	src := unsafe.Pointer(unsafe.StringData(key))
	dst := unsafe.Pointer(unsafe.SliceData(b[n:cap(b)]))
	for i := 0; i < len(key); i += variantKeyChunk {
		*(*[variantKeyChunk]byte)(unsafe.Add(dst, i)) = *(*[variantKeyChunk]byte)(unsafe.Add(src, i))
	}
	return b[:n+len(key)]
}

var variantKeys = []string{`"a":`, `"sid":`, `"user_agent":`, `"a_key_longer_than_a_chunk":`}

func BenchmarkVariant_Key(b *testing.B) {
	for _, key := range variantKeys {
		padded := variantPaddedKey(key)
		b.Run("Current/"+strconv.Itoa(len(key)), func(b *testing.B) {
			buf := make([]byte, 0, 4096)
			for i := 0; i < b.N; i++ {
				buf = buf[:0]
				for j := 0; j < 8; j++ {
					buf = append(buf, padded...)
				}
			}
		})
		b.Run("ByChunks/"+strconv.Itoa(len(key)), func(b *testing.B) {
			buf := make([]byte, 0, 4096)
			for i := 0; i < b.N; i++ {
				buf = buf[:0]
				for j := 0; j < 8; j++ {
					buf = appendKeyByChunks(buf, padded)
				}
			}
		})
	}
}

// where the insertion sort of Mapslice.Sort stops being faster than slices.SortFunc, with keys which are random
// in their content and in their length.
func BenchmarkVariant_MapSort(b *testing.B) {
	rnd := rand.New(rand.NewSource(1))
	for _, n := range []int{2, 4, 8, 12, 16, 24, 32, 48, 64, 128} {
		const sets = 64
		keys := make([][]MapItem, sets)
		for i := range keys {
			keys[i] = make([]MapItem, n)
			for j := range keys[i] {
				key := make([]byte, 3+rnd.Intn(20))
				for k := range key {
					key[k] = byte('a' + rnd.Intn(26))
				}
				keys[i][j].Key = key
			}
		}
		work := make([]MapItem, n)
		b.Run("Insertion/"+strconv.Itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(work, keys[i%sets])
				insertionSortMapItems(work)
			}
		})
		b.Run("SortFunc/"+strconv.Itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(work, keys[i%sets])
				slices.SortFunc(work, func(a, b MapItem) int { return bytes.Compare(a.Key, b.Key) })
			}
		})
	}
}

// the reading of the entries of a map: by a range over the map as a map of the same layout, or by
// reflect.MapIter, which sets the key and the value to interface values.
func BenchmarkVariant_MapIter(b *testing.B) {
	for _, n := range []int{1, 5, 20} {
		m := map[string]any{}
		for i := 0; i < n; i++ {
			m["key"+strconv.Itoa(i)] = i
		}
		typ := reflect.TypeOf(m)
		mp := *(*unsafe.Pointer)(unsafe.Pointer(&m))
		shape := NewMapLayout(typ)
		byReflect := &MapLayout{collect: newReflectCollector(typ), StringKey: true, keySize: 16, valueSize: 16}
		for _, variant := range []struct {
			name   string
			layout *MapLayout
		}{{"Shape", shape}, {"Reflect", byReflect}} {
			layout := variant.layout
			b.Run(variant.name+"/"+strconv.Itoa(n), func(b *testing.B) {
				c := &MapContext{}
				var sum int
				for i := 0; i < b.N; i++ {
					layout.Collect(mp, c)
					for j := 0; j < c.Len; j++ {
						sum += len(c.Keys[j]) + int(uintptr(c.ValueAt(j))&1)
					}
				}
				_ = sum
			})
		}
	}
}

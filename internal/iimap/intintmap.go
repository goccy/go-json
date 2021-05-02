package iimap

import (
	"math"
	"sync/atomic"
	"unsafe"
)

// TypeMap is a lockless copy-on-write map to use for type information cache.
// The fill factor used the TypeMap is 0.6.
// A TypeMap will grow as needed.
type TypeMap struct {
	m unsafe.Pointer // *iiMap
}

// NewTypeMap returns a new TypeMap with 8 as initial capacity.
func NewTypeMap() *TypeMap {
	capacity := 8
	iim := newIIMap(capacity)
	return &TypeMap{m: unsafe.Pointer(iim)}
}

func (m *TypeMap) Size() int {
	return (*iiMap)(atomic.LoadPointer(&m.m)).size
}

// Get returns the value if the key is found in the map.
func (m *TypeMap) Get(key uintptr) interface{} {
	return (*iiMap)(atomic.LoadPointer(&m.m)).Get(key)
}

// Set adds or updates key with value to the map, if the key value
// is not present in the underlying map, it will copy the map and
// add the key value to the copy, then swap to the new map using atomic
// operation.
func (m *TypeMap) Set(key uintptr, val interface{}) {
	mm := (*iiMap)(atomic.LoadPointer(&m.m))
	if v := mm.Get(key); v == val {
		return
	}

	newMap := mm.Copy()
	newMap.Set(key, val)
	atomic.StorePointer(&m.m, unsafe.Pointer(newMap))
}

// -------- int interface map -------- //

const (
	intPhi     = 0x9E3779B9
	freeKey    = 0
	fillFactor = 0.6

	ptrsize = unsafe.Sizeof(uintptr(0))
)

func phiMix(x uintptr) uint64 {
	h := uint64(x * intPhi)
	return h ^ (h >> 16)
}

func calcThreshold(capacity int) int {
	return int(math.Floor(float64(capacity) * fillFactor))
}

type iiMap struct {
	data    []iiEntry
	dataptr unsafe.Pointer

	threshold int
	size      int
	mask      uint64
}

type iiEntry struct {
	K uintptr
	V interface{}
}

func newIIMap(capacity int) *iiMap {
	if capacity&(capacity-1) != 0 {
		panic("capacity must be power of two")
	}
	threshold := calcThreshold(capacity)
	mask := capacity - 1
	data := make([]iiEntry, capacity)
	return &iiMap{
		data:      data,
		dataptr:   unsafe.Pointer(&data[0]),
		threshold: threshold,
		size:      0,
		mask:      uint64(mask),
	}
}

// getK helps to eliminate slice bounds checking
func (m *iiMap) getK(ptr uint64) *uintptr {
	return (*uintptr)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize))
}

// getV helps to eliminate slice bounds checking
func (m *iiMap) getV(ptr uint64) *interface{} {
	return (*interface{})(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize + ptrsize))
}

func (m *iiMap) Get(key uintptr) interface{} {
	// manually inline phiMix to help inlining
	h := uint64(key * intPhi)
	ptr := h ^ (h >> 16)

	for {
		ptr &= m.mask
		k := *m.getK(ptr)
		if k == key {
			return *m.getV(ptr)
		}
		if k == freeKey {
			return nil
		}
		ptr += 1
	}
}

func (m *iiMap) Set(key uintptr, val interface{}) {
	ptr := phiMix(key)
	for {
		ptr &= m.mask
		k := *m.getK(ptr)
		if k == freeKey {
			*m.getK(ptr) = key
			*m.getV(ptr) = val
			m.size++
			return
		}
		if k == key {
			*m.getV(ptr) = val
			return
		}
		ptr += 1
	}
}

func (m *iiMap) Copy() *iiMap {
	capacity := cap(m.data)
	if m.size >= m.threshold {
		capacity *= 2
	}
	newMap := newIIMap(capacity)
	for _, e := range m.data {
		if e.K == freeKey {
			continue
		}
		newMap.Set(e.K, e.V)
	}
	return newMap
}

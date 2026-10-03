package runtime

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// TypeCache is a table of the values for the types, which is looked up by the address of a type
// without a lock.
//
// A lookup costs what the lookup of an array does: the table is buckets of a cache line, one of which a type is
// hashed to, and the table keeps every type in its bucket, so a lookup reads the bucket and compares the types
// in it, and nothing else. It is made larger when a type would have no room in its bucket, which happens only
// when a type is stored: the types of a program are stored once each. The table used to be an array with a slot
// for every address a type of the program can be at, found from the type links of the runtime, and then a hash
// table with open addressing, whose lookup depended on where the other types were.
//
// Init must be called before the table is used: a lookup reads a table which is always there, without a check.
type TypeCache[T any] struct {
	cache typeCache
}

// Init makes the table empty. It is called once, before the table is used, and allocates nothing: the empty
// table is one which every TypeCache shares, and which a store replaces, as it replaces any table.
func (c *TypeCache[T]) Init() {
	atomic.StorePointer(&c.cache.table, emptyTypeTable)
}

// emptyTypeTable is the table of a TypeCache which has no type yet: its buckets are never written.
var emptyTypeTable = tableWord(newTypeBuckets(minTypeBucketBits), 64-minTypeBucketBits)

// Load returns the value for the type, or nil. It is inlined into the callers.
func (c *TypeCache[T]) Load(typ uintptr) *T {
	return (*T)(c.cache.load(typ))
}

// Store sets the value for the type, and returns the value which the table has for the type:
// it is the value which was stored first if the type was compiled by more than one goroutine at a time.
func (c *TypeCache[T]) Store(typ uintptr, v *T) *T {
	return (*T)(c.cache.store(typ, unsafe.Pointer(v)))
}

// typeBucketSize is the number of the types of a bucket: the types and their values fill a cache line.
const typeBucketSize = 4

// typeBucket is the types of a bucket and their values. The values are words which the GC doesn't scan, as the
// types are, so that the buckets cost the GC nothing however many they are: typeCache.refs keeps the values alive.
type typeBucket struct {
	types  [typeBucketSize]uintptr
	values [typeBucketSize]uintptr
}

// typeCache is TypeCache for the values as pointers, so that its lookups are not generic code, which the compiler
// doesn't inline.
//
// A table is never written after it is published: a type is stored by a new table with the type, which replaces
// the table. So a lookup reads the table with one atomic load, of the word which has both the address of the
// buckets and the shift which makes the index of a bucket from the hash, and the bucket with plain loads, which the
// atomic load orders after the writes of the table: they cost what the loads of an array do. The buckets are
// aligned to their size, a cache line, so the shift fits in the low bits of their address, and the word points
// into the first bucket, which keeps the buckets alive.
type typeCache struct {
	table unsafe.Pointer // the first bucket plus the shift ( see tableWord ), read and written atomically
	// mu is for the writers, and count is the number of the types and refs the values, which it guards.
	mu    sync.Mutex
	count int
	refs  []unsafe.Pointer
}

// TypeHashMultiplier is the multiplier of the Fibonacci hashing of the address of a type: the top bits of the
// product spread the addresses, which are aligned and close to each other, over a table.
const TypeHashMultiplier = 0x9E3779B97F4A7C15

// minTypeBucketBits is the log2 of the number of the buckets of a new table.
const minTypeBucketBits = 4

func (c *typeCache) load(typ uintptr) unsafe.Pointer {
	table := atomic.LoadPointer(&c.table)
	shift := uintptr(table) % typeBucketAlign
	b := (*typeBucket)(unsafe.Add(table, uintptr((uint64(typ)*TypeHashMultiplier)>>shift)*unsafe.Sizeof(typeBucket{})-shift))
	for i := range b.types {
		if b.types[i] == typ {
			// the value is the address it is: it is kept alive by typeCache.refs, and a value in the heap doesn't
			// move.
			return *(*unsafe.Pointer)(unsafe.Pointer(&b.values[i]))
		}
	}
	return nil
}

// typeBucketAlign is the alignment of the buckets, which is their size: newTypeBuckets aligns them.
const typeBucketAlign = unsafe.Sizeof(typeBucket{})

// tableWord is the word of the table of the buckets: the address of the first bucket plus the shift.
func tableWord(buckets []typeBucket, shift uint64) unsafe.Pointer {
	return unsafe.Add(unsafe.Pointer(&buckets[0]), shift)
}

func (c *typeCache) store(typ uintptr, v unsafe.Pointer) unsafe.Pointer {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing := c.load(typ); existing != nil {
		return existing
	}
	c.refs = append(c.refs, v)
	table := atomic.LoadPointer(&c.table)
	shift := uintptr(table) % typeBucketAlign
	old := unsafe.Slice((*typeBucket)(unsafe.Add(table, -shift)), 1<<(64-shift))
	bits := uint64(64 - shift)
	// the new table has the types of the table and the type, each in its bucket: it is larger when a bucket has no
	// room for them, and when there are more types than buckets, so that most of the types are the first of their
	// buckets and a lookup finds them by its first comparison.
	for c.count+1 > 1<<bits {
		bits++
	}
	for ; ; bits++ {
		buckets := newTypeBuckets(bits)
		if rehashTypeBuckets(old, buckets, 64-bits) && insertTypeBucket(buckets, 64-bits, typ, v) {
			atomic.StorePointer(&c.table, tableWord(buckets, 64-bits))
			break
		}
	}
	c.count++
	return v
}

// newTypeBuckets returns 1<<bits buckets, aligned to their size: one more is allocated, of which the part before
// the alignment is left unused.
func newTypeBuckets(bits uint64) []typeBucket {
	buckets := make([]typeBucket, 1+1<<bits)
	skip := (typeBucketAlign - uintptr(unsafe.Pointer(&buckets[0]))%typeBucketAlign) % typeBucketAlign
	return unsafe.Slice((*typeBucket)(unsafe.Add(unsafe.Pointer(&buckets[0]), skip)), 1<<bits)
}

// rehashTypeBuckets puts the types of the buckets into the other buckets, and reports whether every one had room
// in its bucket.
func rehashTypeBuckets(from, to []typeBucket, toShift uint64) bool {
	for i := range from {
		for j, typ := range from[i].types {
			if typ != 0 && !insertTypeBucketWord(to, toShift, typ, from[i].values[j]) {
				return false
			}
		}
	}
	return true
}

func insertTypeBucket(buckets []typeBucket, shift uint64, typ uintptr, v unsafe.Pointer) bool {
	return insertTypeBucketWord(buckets, shift, typ, uintptr(v))
}

func insertTypeBucketWord(buckets []typeBucket, shift uint64, typ, v uintptr) bool {
	b := &buckets[(uint64(typ)*TypeHashMultiplier)>>shift]
	for i := range b.types {
		if b.types[i] == 0 {
			b.types[i], b.values[i] = typ, v
			return true
		}
	}
	return false
}

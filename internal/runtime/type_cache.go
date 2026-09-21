package runtime

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// TypeCache is a table of values which is indexed by the address of a type.
//
// The table has a slot for every address a type can be at, so it has hundreds of thousands of slots, almost all
// of which are empty. If the slots were pointers, the GC would scan all of them in every cycle: with a small
// heap, which makes the cycles frequent, that took more than a third of the CPU time of encoding.
// So a slot holds the address of a value as uintptr, which the GC doesn't scan, and the values are kept alive by
// refs, which has only the values.
type TypeCache[T any] struct {
	slots []uintptr
	mu    sync.Mutex
	refs  []*T
}

func NewTypeCache[T any](length uintptr) *TypeCache[T] {
	return &TypeCache[T]{slots: make([]uintptr, length)}
}

// Load returns the value of the slot, or nil.
func (c *TypeCache[T]) Load(index uintptr) *T {
	// the slot is read as a pointer, not converted from uintptr: the value it refers to is alive, and never moves.
	return (*T)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.slots[index]))))
}

// Store sets the value of the slot. A value is never removed.
func (c *TypeCache[T]) Store(index uintptr, v *T) {
	c.mu.Lock()
	c.refs = append(c.refs, v)
	c.mu.Unlock()
	// the value is referred to by refs before it gets visible.
	atomic.StoreUintptr(&c.slots[index], uintptr(unsafe.Pointer(v)))
}

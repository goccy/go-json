package encoder

import (
	"context"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

type compileContext struct {
	opcodeIndex       uint32
	ptrIndex          int
	indent            uint32
	escapeKey         bool
	structTypeToCodes map[uintptr]Opcodes
	recursiveCodes    *Opcodes
}

func (c *compileContext) incIndent() {
	c.indent++
}

func (c *compileContext) decIndent() {
	c.indent--
}

func (c *compileContext) incIndex() {
	c.incOpcodeIndex()
	c.incPtrIndex()
}

func (c *compileContext) decIndex() {
	c.decOpcodeIndex()
	c.decPtrIndex()
}

func (c *compileContext) incOpcodeIndex() {
	c.opcodeIndex++
}

func (c *compileContext) decOpcodeIndex() {
	c.opcodeIndex--
}

func (c *compileContext) incPtrIndex() {
	c.ptrIndex++
}

func (c *compileContext) decPtrIndex() {
	c.ptrIndex--
}

const (
	bufSize = 1024

	maxSlotsToClearByLoop = 16
)

var (
	runtimeContextPool = sync.Pool{
		New: func() interface{} {
			return &RuntimeContext{
				Buf:    make([]byte, 0, bufSize),
				Slots:  make([]Slot, 128),
				Option: &Option{},
			}
		},
	}
)

// Slot is a slot of the VM. The opcodes refer to a half of a slot by its offset from the head of the frame:
// Ptr is for a pointer ( the address of a value, the context of a map, the opcode to return to ) and
// Int is for the other values ( an index, a length, the offset of a frame, an indent ).
//
// A pointer is always held as unsafe.Pointer, so the GC sees every value being encoded, and the address stays
// valid wherever the value is.
type Slot struct {
	Ptr unsafe.Pointer
	Int uintptr
}

type RuntimeContext struct {
	Context    context.Context
	Buf        []byte
	MarshalBuf []byte
	Slots      []Slot
	SeenPtr    []unsafe.Pointer
	BaseIndent uint32
	Prefix     []byte
	IndentStr  []byte
	Option     *Option
	// usedSlots is the number of the slots which may hold a pointer.
	usedSlots int
	// nested is whether a frame was added by ReserveSlots: only such a frame uses SeenPtr and valueSlots.
	nested bool
	// topValue and valueSlots hold the values which are stored directly in an interface value:
	// topValue is for the value passed to Marshal, and a slot per nesting level is for the values
	// held by the interface values. A slot is never moved.
	topValue   unsafe.Pointer
	valueSlots []*unsafe.Pointer
}

// ValueAddr returns the address of the value passed to Marshal, which the data word of its interface value
// represents.
//
// The opcodes always take the address of a value. The data word of an interface value is the address
// for most of the types, but it is the value itself if the type is stored directly ( a pointer, a map,
// a struct of a single pointer, ... ). Such a value is copied to the context, and the address of the copy
// is returned.
func (c *RuntimeContext) ValueAddr(codeSet *OpcodeSet, dataWord unsafe.Pointer) unsafe.Pointer {
	if codeSet.DataWordIsAddr {
		return dataWord
	}
	c.topValue = dataWord
	return unsafe.Pointer(&c.topValue)
}

// InterfaceValueAddr is ValueAddr for a value held by an interface value at the nesting level.
//
// It is never inlined, and the VM doesn't check OpcodeSet.IfaceIndir by itself either,
// because a branch added to the VM changes the register allocation of the whole VM.
//
//go:noinline
func (c *RuntimeContext) InterfaceValueAddr(codeSet *OpcodeSet, dataWord unsafe.Pointer, level int) unsafe.Pointer {
	if codeSet.DataWordIsAddr {
		return dataWord
	}
	for len(c.valueSlots) <= level {
		c.valueSlots = append(c.valueSlots, new(unsafe.Pointer))
	}
	slot := c.valueSlots[level]
	*slot = dataWord
	return unsafe.Pointer(slot)
}

func (c *RuntimeContext) Init(p unsafe.Pointer, codelen int) {
	if len(c.Slots) < codelen {
		c.Slots = make([]Slot, codelen)
	}
	c.Slots[0].Ptr = p
	c.usedSlots = codelen
	c.SeenPtr = c.SeenPtr[:0]
	c.BaseIndent = 0
}

// ReserveSlots makes the context have the slots of the frames up to the length.
//
//go:noinline
func (c *RuntimeContext) ReserveSlots(length uintptr) {
	n := int(length)
	if len(c.Slots) < n {
		c.Slots = append(c.Slots, make([]Slot, n-len(c.Slots))...)
	}
	if c.usedSlots < n {
		c.usedSlots = n
	}
	c.nested = true
}

// Ptr returns the pointer to the slots.
// It is unsafe.Pointer, not uintptr, so that the address of a slot is calculated by unsafe.Add,
// which the compiler folds into the addressing mode of the load / store of the slot.
func (c *RuntimeContext) Ptr() unsafe.Pointer {
	header := (*runtime.SliceHeader)(unsafe.Pointer(&c.Slots))
	return header.Data
}

func TakeRuntimeContext() *RuntimeContext {
	return runtimeContextPool.Get().(*RuntimeContext)
}

func ReleaseRuntimeContext(ctx *RuntimeContext) {
	// The context of a call must neither be kept by the pool nor be seen by the next call,
	// which may not be given a context at all.
	ctx.Option.Context = nil
	ctx.releaseValues()
	runtimeContextPool.Put(ctx)
}

// releaseValues clears every pointer to the values which were encoded, so that the pool doesn't keep them alive.
func (c *RuntimeContext) releaseValues() {
	if slots := c.Slots[:c.usedSlots]; len(slots) <= maxSlotsToClearByLoop {
		// a call to clear the memory costs more than the loop for a few slots.
		for i := range slots {
			slots[i].Ptr = nil
		}
	} else {
		clear(slots)
	}
	c.usedSlots = 0
	c.topValue = nil
	if c.nested {
		// what only the frames of an interface value and of a recursive type use.
		clear(c.SeenPtr[:cap(c.SeenPtr)])
		for _, slot := range c.valueSlots {
			*slot = nil
		}
		c.nested = false
	}
}

// marshalerContext returns the context to call MarshalJSON(context.Context) with.
// It is never nil, also for a call which is not given a context.
func (c *RuntimeContext) marshalerContext() context.Context {
	if c.Option.Context == nil {
		return context.Background()
	}
	return c.Option.Context
}

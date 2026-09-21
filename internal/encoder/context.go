package encoder

import (
	"context"
	"sync"
	"unsafe"
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
)

var (
	runtimeContextPool = sync.Pool{
		New: func() interface{} {
			return &RuntimeContext{
				Buf:    make([]byte, 0, bufSize),
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

// SlotsLength is the number of the slots which the VM has on the stack at a time.
const SlotsLength = 32

// Slots is the slots of the VM, which is a local variable of it: see Run.
type Slots [SlotsLength]Slot

type RuntimeContext struct {
	Context    context.Context
	Buf        []byte
	MarshalBuf []byte
	SeenPtr    []unsafe.Pointer
	BaseIndent uint32
	Prefix     []byte
	IndentStr  []byte
	Option     *Option
	// usedValueSlots is whether valueSlots may hold a value.
	usedValueSlots bool
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
	c.usedValueSlots = true
	return unsafe.Pointer(slot)
}

func (c *RuntimeContext) Init() {
	c.SeenPtr = c.SeenPtr[:0]
	c.BaseIndent = 0
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
	c.topValue = nil
	if cap(c.SeenPtr) > 0 {
		clear(c.SeenPtr[:cap(c.SeenPtr)])
	}
	if c.usedValueSlots {
		for _, slot := range c.valueSlots {
			*slot = nil
		}
		c.usedValueSlots = false
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

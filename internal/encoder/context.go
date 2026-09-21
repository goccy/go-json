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
)

var (
	runtimeContextPool = sync.Pool{
		New: func() interface{} {
			return &RuntimeContext{
				Buf:      make([]byte, 0, bufSize),
				Ptrs:     make([]uintptr, 128),
				KeepRefs: make([]unsafe.Pointer, 0, 8),
				Option:   &Option{},
			}
		},
	}
)

type RuntimeContext struct {
	Context    context.Context
	Buf        []byte
	MarshalBuf []byte
	Ptrs       []uintptr
	KeepRefs   []unsafe.Pointer
	SeenPtr    []uintptr
	BaseIndent uint32
	Prefix     []byte
	IndentStr  []byte
	Option     *Option
	// topValue and valueSlots hold the values which are stored directly in an interface value:
	// topValue is for the value passed to Marshal, and a slot per nesting level is for the values
	// held by the interface values. A slot is never moved.
	topValue   uintptr
	valueSlots []*uintptr
}

// ValueAddr returns the address of the value passed to Marshal, which the data word of its interface value
// represents.
//
// The opcodes always take the address of a value. The data word of an interface value is the address
// for most of the types, but it is the value itself if the type is stored directly ( a pointer, a map,
// a struct of a single pointer, ... ). Such a value is copied to the context, and the address of the copy
// is returned.
//
// The copy is held as uintptr so that the value doesn't escape:
// the caller has to keep the value alive while it is encoded.
func (c *RuntimeContext) ValueAddr(codeSet *OpcodeSet, dataWord uintptr) uintptr {
	if codeSet.DataWordIsAddr {
		return dataWord
	}
	c.topValue = dataWord
	return uintptr(unsafe.Pointer(&c.topValue))
}

// InterfaceValueAddr is ValueAddr for a value held by an interface value at the nesting level.
//
// It is never inlined, and the VM doesn't check OpcodeSet.IfaceIndir by itself either,
// because a branch added to the VM changes the register allocation of the whole VM.
//
//go:noinline
func (c *RuntimeContext) InterfaceValueAddr(codeSet *OpcodeSet, dataWord uintptr, level int) uintptr {
	if codeSet.DataWordIsAddr {
		return dataWord
	}
	for len(c.valueSlots) <= level {
		c.valueSlots = append(c.valueSlots, new(uintptr))
	}
	slot := c.valueSlots[level]
	*slot = dataWord
	return uintptr(unsafe.Pointer(slot))
}

func (c *RuntimeContext) Init(p uintptr, codelen int) {
	if len(c.Ptrs) < codelen {
		c.Ptrs = make([]uintptr, codelen)
	}
	c.Ptrs[0] = p
	c.KeepRefs = c.KeepRefs[:0]
	c.SeenPtr = c.SeenPtr[:0]
	c.BaseIndent = 0
}

// Ptr returns the pointer to the slots of the pointers.
// It is unsafe.Pointer, not uintptr, so that the address of a slot is calculated by unsafe.Add,
// which the compiler folds into the addressing mode of the load / store of the slot.
// What the slots hold stays uintptr: they are not seen by the GC, and the values don't escape.
func (c *RuntimeContext) Ptr() unsafe.Pointer {
	header := (*runtime.SliceHeader)(unsafe.Pointer(&c.Ptrs))
	return header.Data
}

func TakeRuntimeContext() *RuntimeContext {
	return runtimeContextPool.Get().(*RuntimeContext)
}

func ReleaseRuntimeContext(ctx *RuntimeContext) {
	// The context of a call must neither be kept by the pool nor be seen by the next call,
	// which may not be given a context at all.
	ctx.Option.Context = nil
	runtimeContextPool.Put(ctx)
}

// marshalerContext returns the context to call MarshalJSON(context.Context) with.
// It is never nil, also for a call which is not given a context.
func (c *RuntimeContext) marshalerContext() context.Context {
	if c.Option.Context == nil {
		return context.Background()
	}
	return c.Option.Context
}

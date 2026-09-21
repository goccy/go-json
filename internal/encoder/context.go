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
	if codeSet.IfaceIndir {
		return dataWord
	}
	c.topValue = dataWord
	return uintptr(unsafe.Pointer(&c.topValue))
}

// InterfaceValueAddr is ValueAddr for a value held by an interface value at the nesting level.
// The caller checks OpcodeSet.IfaceIndir by itself, so that it is called only for the type stored directly.
//
//go:noinline
func (c *RuntimeContext) InterfaceValueAddr(dataWord uintptr, level int) uintptr {
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

func (c *RuntimeContext) Ptr() uintptr {
	header := (*runtime.SliceHeader)(unsafe.Pointer(&c.Ptrs))
	return uintptr(header.Data)
}

func TakeRuntimeContext() *RuntimeContext {
	return runtimeContextPool.Get().(*RuntimeContext)
}

func ReleaseRuntimeContext(ctx *RuntimeContext) {
	runtimeContextPool.Put(ctx)
}

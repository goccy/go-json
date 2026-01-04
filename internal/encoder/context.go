package encoder

import (
	"context"
	"reflect"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

type compileContext struct {
	opcodeIndex       uint32
	ptrIndex          int
	indent            uint32
	escapeKey         bool
	structTypeToCodes map[reflect.Type]Opcodes
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
				Ptrs:   make([]unsafe.Pointer, 128),
				Option: &Option{},
			}
		},
	}
)

type RuntimeContext struct {
	Context    context.Context
	Buf        []byte
	MarshalBuf []byte
	Ptrs       []unsafe.Pointer
	SeenPtr    []unsafe.Pointer
	BaseIndent uint32
	Prefix     []byte
	IndentStr  []byte
	Option     *Option
}

func (c *RuntimeContext) Init(p unsafe.Pointer, codelen int) {
	if len(c.Ptrs) < codelen {
		c.Ptrs = make([]unsafe.Pointer, codelen)
	}
	c.Ptrs[0] = p
	c.SeenPtr = c.SeenPtr[:0]
	c.BaseIndent = 0
}

func (c *RuntimeContext) Ptr() unsafe.Pointer {
	header := (*runtime.SliceHeader)(unsafe.Pointer(&c.Ptrs))
	return unsafe.Pointer(header.Data)
}

func TakeRuntimeContext() *RuntimeContext {
	return runtimeContextPool.Get().(*RuntimeContext)
}

func ReleaseRuntimeContext(ctx *RuntimeContext) {
	runtimeContextPool.Put(ctx)
}

package decoder

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

// The functions of unmarshaling of WithUnmarshalers of encoding/json/v2: a value of a type which a function takes is
// decoded by the function, before its method or its kind. The decoders of a type depend on the functions, so they
// are compiled for the functions and cached with them, apart from the tables shared by every goroutine.

// UnmarshalFunc is a function of unmarshaling, for the values of the types castable to Type ( see
// runtime.CastableTo ).
type UnmarshalFunc struct {
	Type reflect.Type
	// Decode decodes the value at cursor into the value at p, of the type from, and returns the cursor after it.
	Decode func(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer, from reflect.Type) (int64, error)
}

// UnmarshalFuncs are the functions of unmarshaling of a call, in the order they are tried, with the decoders
// compiled for them.
type UnmarshalFuncs struct {
	Funcs    []UnmarshalFunc
	decoders [2]sync.Map // map[uintptr]Decoder by the pointer type, without and with StringifyNumbers
	castable sync.Map    // map[reflect.Type][]*UnmarshalFunc
}

// decoderOf returns the decoder of the pointer type typ for the functions, compiling it if the type is new.
func (f *UnmarshalFuncs) decoderOf(typ unsafe.Pointer, stringifyNumbers bool) (Decoder, error) {
	cache := &f.decoders[0]
	if stringifyNumbers {
		cache = &f.decoders[1]
	}
	if v, ok := cache.Load(uintptr(typ)); ok {
		return v.(Decoder), nil
	}
	c := newCompiler(true)
	c.stringifyNumbers, c.funcs = stringifyNumbers, f
	dec, err := compileHead(runtime.TypeOfPtr(typ), c)
	if err != nil {
		return nil, err
	}
	v, _ := cache.LoadOrStore(uintptr(typ), dec)
	return v.(Decoder), nil
}

// funcsOf returns the functions which take a value of the type, in their order.
func (f *UnmarshalFuncs) funcsOf(typ reflect.Type) []*UnmarshalFunc {
	if f == nil {
		return nil
	}
	if v, ok := f.castable.Load(typ); ok {
		return v.([]*UnmarshalFunc)
	}
	var funcs []*UnmarshalFunc
	for i := range f.Funcs {
		if runtime.CastableTo(typ, f.Funcs[i].Type) {
			funcs = append(funcs, &f.Funcs[i])
		}
	}
	v, _ := f.castable.LoadOrStore(typ, funcs)
	return v.([]*UnmarshalFunc)
}

// SetV2Funcs makes ctx decode by the functions of unmarshaling, or by none if funcs is nil.
func (ctx *RuntimeContext) SetV2Funcs(funcs *UnmarshalFuncs) {
	ctx.v2Funcs = funcs
}

// v2FuncDecoder is the decoder of a value of a type which a function of unmarshaling takes, which decodes it.
type v2FuncDecoder struct {
	typ reflect.Type
	fn  *UnmarshalFunc
}

func (d *v2FuncDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	return d.fn.Decode(ctx, cursor, depth, p, d.typ)
}

func (d *v2FuncDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

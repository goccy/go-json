package decoder

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/internal/runtime"
)

// The functions of unmarshaling of WithUnmarshalers of encoding/json/v2: a value of a type which a function takes is
// decoded by the function, before its method or its kind. The decoders of a type depend on the functions, so they
// are compiled for the functions and cached with them, apart from the tables shared by every goroutine.

// UnmarshalFunc is a function of unmarshaling, for the values of the types castable to Type ( see
// runtime.CastableTo ).
type UnmarshalFunc struct {
	Type reflect.Type
	// Decode decodes the value at cursor into the value at p, of the type from, and returns the cursor after it, or
	// ErrUseDefault if the function declined the value, which it may if Declines is set.
	Decode   func(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer, from reflect.Type) (int64, error)
	Declines bool
}

// ErrUseDefault is the error of a method or a function of unmarshaling which declined a value without reading it:
// the next function, or the next method or the kind of the type, decodes it.
var ErrUseDefault = errors.New("json: the value is decoded by the next representation")

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

// V2Call is the state of a call of the v2 json package which its decoders read, which a context keeps from a call to
// the next once a call made it ( see RuntimeContext.V2Call ).
type V2Call struct {
	// Funcs are the functions of unmarshaling of the call, which the decoders are compiled for ( see
	// DecoderOfV2 ), or nil.
	Funcs *UnmarshalFuncs
	// Config are the options of the call, which a method or a function of unmarshaling which reads a
	// jsontext.Decoder is given.
	Config *options.Config
	// From is the state of the v2 json package for the methods and the functions which read a jsontext.Decoder,
	// which the decoder doesn't know.
	From unsafe.Pointer
}

// V2Call returns the state of the call of the v2 json package which decodes by ctx, which it makes the first time.
func (ctx *RuntimeContext) V2Call() *V2Call {
	if ctx.v2 == nil {
		ctx.v2 = new(V2Call)
	}
	return ctx.v2
}

// v2Funcs returns the functions of unmarshaling of the call which decodes by ctx, or nil.
func (ctx *RuntimeContext) v2Funcs() *UnmarshalFuncs {
	if v2 := ctx.v2; v2 != nil {
		return v2.Funcs
	}
	return nil
}

// v2FuncDecoder is the decoder of a value of a type which the functions of unmarshaling take, which decode it in
// their order: a function which declines it leaves it to the next one, and the last one to next, the decoder of
// the type without them, which is nil if the last one can't decline.
type v2FuncDecoder struct {
	typ   reflect.Type
	funcs []*UnmarshalFunc
	next  Decoder
}

func (d *v2FuncDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	for _, fn := range d.funcs {
		next, err := fn.Decode(ctx, cursor, depth, p, d.typ)
		if err != ErrUseDefault {
			return next, err
		}
	}
	return d.next.Decode(ctx, cursor, depth, p)
}

func (d *v2FuncDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

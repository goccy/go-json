package encoder

import (
	"reflect"
	"sync"
	"unsafe"
)

// The functions of marshaling of WithMarshalers of encoding/json/v2: a value of a type which a function takes is
// written by the function, before its method or its kind. The opcodes of a type depend on the functions, so they
// are compiled for the functions and cached with them, apart from the tables shared by every goroutine.

// MarshalFunc is a function of marshaling, for the values of the types castable to Type ( see castableTo ).
type MarshalFunc struct {
	Type reflect.Type
	// Append appends the value at p, of the type from: ErrUseDefault if the function declined to write it, which
	// lets the next function, the method or the kind of the type write it.
	Append func(ctx *RuntimeContext, b []byte, p unsafe.Pointer, from reflect.Type) ([]byte, error)
}

// MarshalFuncs are the functions of marshaling of a call, in the order they are tried, with the opcodes compiled
// for them.
type MarshalFuncs struct {
	Funcs    []MarshalFunc
	codeSets sync.Map // map[funcsKey]*OpcodeSet
	castable sync.Map // map[reflect.Type][]*MarshalFunc
	defaults sync.Map // map[defaultKey]*OpcodeSet: the default representations ( see appendDefault )
}

type funcsKey struct {
	typeptr uintptr
	mode    compileMode
}

// codeSet returns the opcodes of the type for the functions, compiling them if the type is new.
func (f *MarshalFuncs) codeSet(typeptr uintptr, mode compileMode) (*OpcodeSet, error) {
	key := funcsKey{typeptr: typeptr, mode: mode}
	if v, ok := f.codeSets.Load(key); ok {
		return v.(*OpcodeSet), nil
	}
	c := newCompiler(mode)
	c.funcs = f
	codeSet, err := c.compile(typeptr)
	if err != nil {
		return nil, err
	}
	v, _ := f.codeSets.LoadOrStore(key, codeSet)
	return v.(*OpcodeSet), nil
}

// funcsOf returns the functions which take a value of the type, in their order.
func (f *MarshalFuncs) funcsOf(typ reflect.Type) []*MarshalFunc {
	if f == nil {
		return nil
	}
	if v, ok := f.castable.Load(typ); ok {
		return v.([]*MarshalFunc)
	}
	var funcs []*MarshalFunc
	for i := range f.Funcs {
		if castableTo(typ, f.Funcs[i].Type) {
			funcs = append(funcs, &f.Funcs[i])
		}
	}
	v, _ := f.castable.LoadOrStore(typ, funcs)
	return v.([]*MarshalFunc)
}

// castableTo reports whether a value of the type from is given to a function of the type to: as itself, as its
// address if to is the pointer to it, or as its address in an interface value if its pointer type implements to.
func castableTo(from, to reflect.Type) bool {
	switch to.Kind() {
	case reflect.Interface:
		return reflect.PointerTo(from).Implements(to)
	case reflect.Pointer:
		return reflect.PointerTo(from) == to
	default:
		return from == to
	}
}

// appendByFuncs appends the value at p, of typ, by the first of the functions which doesn't decline it, or
// returns ErrUseDefault if they all do.
func appendByFuncs(ctx *RuntimeContext, b []byte, p unsafe.Pointer, typ reflect.Type, funcs []*MarshalFunc) ([]byte, error) {
	for _, f := range funcs {
		out, err := f.Append(ctx, b, p, typ)
		if err != ErrUseDefault {
			return out, err
		}
	}
	return b, ErrUseDefault
}

// v2FuncsCode returns the code of a value of typ written by the functions of marshaling, and true, if a function
// takes it. A value which they all decline is written by its default representation ( see appendDefault ).
func (c *Compiler) v2FuncsCode(typ reflect.Type) (Code, bool) {
	funcs := c.funcs.funcsOf(typ)
	if len(funcs) == 0 {
		return nil, false
	}
	return c.appendFuncCode(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
		out, err := appendByFuncs(ctx, b, p, typ, funcs)
		if err == ErrUseDefault {
			return appendDefault(ctx, b, typ, p, noMethod)
		}
		return out, err
	}), true
}

package decoder

import (
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

// The decoders allocate and assign values through the reflect package, never through the internal
// functions of the runtime: the pointers to the type descriptors they need are computed once, when
// a decoder is compiled.

// valueAt returns the addressable reflect.Value of the value at p, whose pointer type is ptrType
// ( the type descriptor of *T for a value of T ). It is reflect.NewAt(T, p).Elem() without the lookup
// of the pointer type which reflect.NewAt does for every call.
func valueAt(ptrType, p unsafe.Pointer) reflect.Value {
	return reflect.ValueOf(*(*any)(unsafe.Pointer(&emptyInterface{typ: ptrType, ptr: p}))).Elem()
}

// ptrTypeOf returns the type descriptor of *typ.
func ptrTypeOf(typ reflect.Type) unsafe.Pointer {
	return runtime.TypePtr(reflect.PointerTo(typ))
}

// newValue allocates a zero value of typ.
func newValue(typ reflect.Type) unsafe.Pointer {
	return reflect.New(typ).UnsafePointer()
}

// zeroBase is the address of the data of an empty slice which is not nil.
var zeroBase [0]byte

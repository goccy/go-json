package encoder

import (
	"context"
	"encoding"
	"encoding/json"
	"reflect"
	"unsafe"
)

// A method of a marshaler is called directly, by the code pointer of the method taken from the itab of the
// type and the interface when the type is compiled, instead of through the interface: making the interface
// value, asserting the interface and calling through it cost more than the call.
//
// A call through an interface passes the data word of the interface value as the receiver, and the opcodes of
// a marshaler carry that very word: the address of the value, or the pointer itself for a pointer type.
// A func value is a pointer to a closure whose first word is the code, and the first argument of a func value
// is passed as the receiver of a method is, so the code of the method is called as a func value.

// itab is the layout of the itab of a non-empty interface value.
type itab struct {
	inter unsafe.Pointer
	typ   unsafe.Pointer
	hash  uint32
	_     uint32
	fun   [1]uintptr
}

// MarshalerCall is the method of a marshaler, as it is called with the data word of the interface value.
type MarshalerCall struct {
	// fn is the code of the method: it is the first word, so that a pointer to MarshalerCall is a func value.
	fn uintptr
	// nilIsNull is whether a nil receiver is encoded as null without a call, which is so for a pointer.
	nilIsNull bool
	// recv is the type of the receiver, for an error.
	recv reflect.Type
}

func (m *MarshalerCall) call(recv unsafe.Pointer) ([]byte, error) {
	f := *(*func(unsafe.Pointer) ([]byte, error))(unsafe.Pointer(&m))
	return f(recv)
}

func (m *MarshalerCall) callContext(recv unsafe.Pointer, ctx context.Context) ([]byte, error) {
	f := *(*func(unsafe.Pointer, context.Context) ([]byte, error))(unsafe.Pointer(&m))
	return f(recv, ctx)
}

// newMarshalerCall returns the call of the method of the interface which the type of the receiver implements.
// It is decided when the type is compiled.
func newMarshalerCall(recv reflect.Type, iface reflect.Type) *MarshalerCall {
	// a zero value of the type is converted to the interface, which gives the itab.
	v := reflect.Zero(recv).Interface()
	var fn uintptr
	switch iface {
	case marshalJSONType:
		m, ok := v.(json.Marshaler)
		if !ok {
			return nil
		}
		fn = (*nonEmptyInterface)(unsafe.Pointer(&m)).methodCode(0)
	case marshalJSONContextType:
		m, ok := v.(marshalerContext)
		if !ok {
			return nil
		}
		fn = (*nonEmptyInterface)(unsafe.Pointer(&m)).methodCode(0)
	case marshalTextType:
		m, ok := v.(encoding.TextMarshaler)
		if !ok {
			return nil
		}
		fn = (*nonEmptyInterface)(unsafe.Pointer(&m)).methodCode(0)
	default:
		return nil
	}
	return &MarshalerCall{fn: fn, nilIsNull: recv.Kind() == reflect.Ptr, recv: recv}
}

// methodCode returns the code of the method of the interface value.
func (i *nonEmptyInterface) methodCode(index int) uintptr {
	return (*itab)(unsafe.Pointer(i.itab)).fun[index]
}

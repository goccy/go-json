package runtime

import "reflect"

// The names of the methods of the marshalers and the unmarshalers, which MethodByName looks up.
const (
	MarshalJSON   = "MarshalJSON"
	MarshalText   = "MarshalText"
	AppendText    = "AppendText"
	UnmarshalJSON = "UnmarshalJSON"
)

// MethodByName is reflect.Type.MethodByName for a method of one of the names above, which it passes as a constant:
// the linker then keeps the methods of that name only. A call of Method or MethodByName of reflect.Type or
// reflect.Value with an argument the compiler can't see as a constant marks the calling function, and the linker
// keeps every exported method of every type of the program ( cmd/compile/internal/walk.usemethod ).
func MethodByName(typ reflect.Type, name string) (reflect.Method, bool) {
	switch name {
	case MarshalJSON:
		return typ.MethodByName(MarshalJSON)
	case MarshalText:
		return typ.MethodByName(MarshalText)
	case AppendText:
		return typ.MethodByName(AppendText)
	case UnmarshalJSON:
		return typ.MethodByName(UnmarshalJSON)
	}
	panic("runtime: MethodByName: " + name + " is not a method of a marshaler")
}

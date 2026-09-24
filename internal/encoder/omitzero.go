package encoder

import (
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

// IsZeroer is the interface used to check custom zero values.
type IsZeroer interface {
	IsZero() bool
}

// IsZeroForOmitZero reports whether the field at p is the zero value of its type: nil for the
// pointer kinds, the primitive zero, or what a custom IsZero method reports. p is the address of
// the field, which the generic field opcode computes as the struct address plus the offset, and
// code.Type is the type of the field.
func IsZeroForOmitZero(code *Opcode, p unsafe.Pointer) bool {
	if p == nil {
		return true
	}
	switch runtime.TypeOfPtr(code.Type).Kind() {
	// the pointer kinds are zero when their first word, the pointer, is nil: an empty but
	// non-nil slice or map is not zero.
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Interface:
		return *(*unsafe.Pointer)(p) == nil

	case reflect.Bool:
		return !*(*bool)(p)

	case reflect.Int:
		return *(*int)(p) == 0
	case reflect.Int8:
		return *(*int8)(p) == 0
	case reflect.Int16:
		return *(*int16)(p) == 0
	case reflect.Int32:
		return *(*int32)(p) == 0
	case reflect.Int64:
		return *(*int64)(p) == 0

	case reflect.Uint:
		return *(*uint)(p) == 0
	case reflect.Uint8:
		return *(*uint8)(p) == 0
	case reflect.Uint16:
		return *(*uint16)(p) == 0
	case reflect.Uint32:
		return *(*uint32)(p) == 0
	case reflect.Uint64:
		return *(*uint64)(p) == 0
	case reflect.Uintptr:
		return *(*uintptr)(p) == 0

	case reflect.Float32:
		return *(*float32)(p) == 0
	case reflect.Float64:
		return *(*float64)(p) == 0

	case reflect.Complex64:
		return *(*complex64)(p) == 0
	case reflect.Complex128:
		return *(*complex128)(p) == 0

	case reflect.String:
		return len(*(*string)(p)) == 0

	// a struct or an array: reflect, which honors a custom IsZero method ( e.g. time.Time ).
	default:
		return isZeroValue(reflect.NewAt(runtime.TypeOfPtr(code.Type), p).Elem())
	}
}

// isZeroValue reports whether rv is zero, with a custom IsZero method preferred over the
// judgment of reflect.
func isZeroValue(rv reflect.Value) bool {
	if rv.CanInterface() {
		if iz, ok := rv.Interface().(IsZeroer); ok {
			return iz.IsZero()
		}
	}
	return rv.IsZero()
}

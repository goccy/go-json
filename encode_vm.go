package json

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

const startDetectingCyclesAfter = 1000

func load(base uintptr, idx uintptr) uintptr {
	addr := base + idx
	return **(**uintptr)(unsafe.Pointer(&addr))
}

func store(base uintptr, idx uintptr, p uintptr) {
	addr := base + idx
	**(**uintptr)(unsafe.Pointer(&addr)) = p
}

func ptrToUint64(p uintptr) uint64      { return **(**uint64)(unsafe.Pointer(&p)) }
func ptrToFloat32(p uintptr) float32    { return **(**float32)(unsafe.Pointer(&p)) }
func ptrToFloat64(p uintptr) float64    { return **(**float64)(unsafe.Pointer(&p)) }
func ptrToBool(p uintptr) bool          { return **(**bool)(unsafe.Pointer(&p)) }
func ptrToBytes(p uintptr) []byte       { return **(**[]byte)(unsafe.Pointer(&p)) }
func ptrToNumber(p uintptr) Number      { return **(**Number)(unsafe.Pointer(&p)) }
func ptrToString(p uintptr) string      { return **(**string)(unsafe.Pointer(&p)) }
func ptrToSlice(p uintptr) *sliceHeader { return *(**sliceHeader)(unsafe.Pointer(&p)) }
func ptrToPtr(p uintptr) uintptr {
	return uintptr(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
}
func ptrToUnsafePtr(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}
func ptrToInterface(code *opcode, p uintptr) interface{} {
	return *(*interface{})(unsafe.Pointer(&emptyInterface{
		typ: code.typ,
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&p)),
	}))
}

func errUnsupportedValue(code *opcode, ptr uintptr) *UnsupportedValueError {
	v := *(*interface{})(unsafe.Pointer(&emptyInterface{
		typ: code.typ,
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&ptr)),
	}))
	return &UnsupportedValueError{
		Value: reflect.ValueOf(v),
		Str:   fmt.Sprintf("encountered a cycle via %s", code.typ),
	}
}

func errUnsupportedFloat(v float64) *UnsupportedValueError {
	return &UnsupportedValueError{
		Value: reflect.ValueOf(v),
		Str:   strconv.FormatFloat(v, 'g', -1, 64),
	}
}

func errMarshalerWithCode(code *opcode, err error) *MarshalerError {
	return &MarshalerError{
		Type: rtype2type(code.typ),
		Err:  err,
	}
}

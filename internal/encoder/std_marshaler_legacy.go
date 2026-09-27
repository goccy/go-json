//go:build !go1.24

package encoder

import (
	"reflect"
	"unsafe"
)

// stdMarshalerAppender returns nil: the appending methods of the types of the standard library are of Go 1.24.
func stdMarshalerAppender(_, _ reflect.Type) func([]byte, unsafe.Pointer) ([]byte, bool) {
	return nil
}

package decoder

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

// heldTypeStringer is a type of a value which interfaces of methods hold.
type heldTypeStringer int

func (heldTypeStringer) String() string { return "" }

// The type which heldType reads of an interface value is the one reflect finds: the first word of an interface{},
// and the type of the itab of an interface value of an interface type with methods.
func TestHeldType(t *testing.T) {
	n := heldTypeStringer(1)
	empty := []any{&n, n, new(int), "s", []int{1}, map[string]int{}, struct{ A, B int }{}}
	for _, v := range empty {
		if got, want := runtime.TypeOfPtr(heldType(unsafe.Pointer(&v), true)), reflect.TypeOf(v); got != want {
			t.Fatalf("interface{} of %T: got %v", v, got)
		}
	}
	stringers := []fmt.Stringer{&n, n, &strings.Builder{}}
	for _, v := range stringers {
		if got, want := runtime.TypeOfPtr(heldType(unsafe.Pointer(&v), false)), reflect.TypeOf(v); got != want {
			t.Fatalf("fmt.Stringer of %T: got %v", v, got)
		}
	}
	readers := []io.Reader{strings.NewReader(""), &strings.Reader{}}
	for _, v := range readers {
		if got, want := runtime.TypeOfPtr(heldType(unsafe.Pointer(&v), false)), reflect.TypeOf(v); got != want {
			t.Fatalf("io.Reader of %T: got %v", v, got)
		}
	}
}

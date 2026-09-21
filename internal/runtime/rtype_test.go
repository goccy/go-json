package runtime

import (
	"reflect"
	"testing"
	"unsafe"
)

// dataWord returns the data word of the interface value.
func dataWord(v interface{}) unsafe.Pointer {
	return (*emptyInterface)(unsafe.Pointer(&v)).ptr
}

// The expectation is not written by hand: it is what the compiler actually did to the value.
// If the type is stored directly, the data word of the interface is the pointer which the value consists of.
// Otherwise the data word points to a copy of the value, so it never equals the pointer.
func TestIfaceIndir(t *testing.T) {
	type (
		ptrStruct       struct{ p *int }
		nestedPtrStruct struct{ s ptrStruct }
		mapStruct       struct{ m map[string]int }
		funcStruct      struct{ f func() }
		arrayStruct     struct{ a [1]*int }
		twoFieldStruct  struct {
			p *int
			n int
		}
		blankFieldStruct struct {
			_ struct{}
			p *int
		}
		ifaceStruct struct{ v interface{} }
	)
	var (
		n  = 1
		p  = &n
		m  = map[string]int{"a": 1}
		ch = make(chan int)
		fn = func() {}
		up = unsafe.Pointer(p)
	)
	word := func(v unsafe.Pointer) unsafe.Pointer { return *(*unsafe.Pointer)(v) }

	for _, test := range []struct {
		name string
		v    interface{}
		// pointer is the pointer which the value consists of, or nil if the value is not a single pointer.
		pointer unsafe.Pointer
	}{
		{"pointer", p, up},
		{"map", m, word(unsafe.Pointer(&m))},
		{"chan", ch, word(unsafe.Pointer(&ch))},
		{"func", fn, word(unsafe.Pointer(&fn))},
		{"unsafe.Pointer", up, up},
		{"struct of a pointer", ptrStruct{p}, up},
		{"struct of a struct of a pointer", nestedPtrStruct{ptrStruct{p}}, up},
		{"struct of a map", mapStruct{m}, word(unsafe.Pointer(&m))},
		{"struct of a func", funcStruct{fn}, word(unsafe.Pointer(&fn))},
		{"struct of an array of a pointer", arrayStruct{[1]*int{p}}, up},
		{"array of a pointer", [1]*int{p}, up},
		{"array of a struct of a pointer", [1]ptrStruct{{p}}, up},
		{"struct of a pointer and an int", twoFieldStruct{p, 1}, up},
		{"struct of a blank field and a pointer", blankFieldStruct{p: p}, up},
		{"struct of an interface", ifaceStruct{p}, up},
		{"array of two pointers", [2]*int{p, p}, up},
		{"empty array of pointers", [0]*int{}, nil},
		{"empty struct", struct{}{}, nil},
		{"int", 1, nil},
		{"string", "a", nil},
		{"slice", []*int{p}, nil},
		{"float64", 1.5, nil},
		{"bool", true, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			wantIndir := test.pointer == nil || dataWord(test.v) != test.pointer
			if got := IfaceIndir(Type2RType(reflect.TypeOf(test.v))); got != wantIndir {
				t.Fatalf("IfaceIndir(%T) = %v, but the compiler stores it with indirect = %v", test.v, got, wantIndir)
			}
		})
	}
}

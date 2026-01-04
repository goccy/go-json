package runtime

import (
	"reflect"
	"unsafe"
)

// Type representing reflect.rtype for noescape trick
type Type struct{}

//go:linkname rtype_NumMethod reflect.(*rtype).NumMethod
//go:noescape
func rtype_NumMethod(*Type) int

func (t *Type) NumMethod() int {
	return rtype_NumMethod(t)
}

//go:linkname rtype_Name reflect.(*rtype).Name
//go:noescape
func rtype_Name(*Type) string

func (t *Type) Name() string {
	return rtype_Name(t)
}

//go:linkname rtype_Size reflect.(*rtype).Size
//go:noescape
func rtype_Size(*Type) uintptr

func (t *Type) Size() uintptr {
	return rtype_Size(t)
}

//go:linkname rtype_String reflect.(*rtype).String
//go:noescape
func rtype_String(*Type) string

func (t *Type) String() string {
	return rtype_String(t)
}

//go:linkname rtype_Kind reflect.(*rtype).Kind
//go:noescape
func rtype_Kind(*Type) reflect.Kind

func (t *Type) Kind() reflect.Kind {
	return rtype_Kind(t)
}

//go:linkname rtype_Implements reflect.(*rtype).Implements
//go:noescape
func rtype_Implements(*Type, reflect.Type) bool

func (t *Type) Implements(u reflect.Type) bool {
	return rtype_Implements(t, u)
}

//go:linkname rtype_Elem reflect.(*rtype).Elem
//go:noescape
func rtype_Elem(*Type) reflect.Type

func (t *Type) Elem() *Type {
	return Type2RType(rtype_Elem(t))
}

//go:linkname rtype_Field reflect.(*rtype).Field
//go:noescape
func rtype_Field(*Type, int) reflect.StructField

func (t *Type) Field(i int) reflect.StructField {
	return rtype_Field(t, i)
}

//go:linkname rtype_FieldByName reflect.(*rtype).FieldByName
//go:noescape
func rtype_FieldByName(*Type, string) (reflect.StructField, bool)

func (t *Type) FieldByName(name string) (reflect.StructField, bool) {
	return rtype_FieldByName(t, name)
}

//go:linkname rtype_Key reflect.(*rtype).Key
//go:noescape
func rtype_Key(*Type) reflect.Type

func (t *Type) Key() *Type {
	return Type2RType(rtype_Key(t))
}

//go:linkname rtype_Len reflect.(*rtype).Len
//go:noescape
func rtype_Len(*Type) int

func (t *Type) Len() int {
	return rtype_Len(t)
}

//go:linkname rtype_NumField reflect.(*rtype).NumField
//go:noescape
func rtype_NumField(*Type) int

func (t *Type) NumField() int {
	return rtype_NumField(t)
}

//go:linkname PtrTo reflect.(*rtype).ptrTo
//go:noescape
func PtrTo(*Type) *Type

//go:linkname RType2Type reflect.toType
//go:noescape
func RType2Type(t *Type) reflect.Type

type emptyInterface struct {
	_   *Type
	ptr unsafe.Pointer
}

func Type2RType(t reflect.Type) *Type {
	return (*Type)(((*emptyInterface)(unsafe.Pointer(&t))).ptr)
}

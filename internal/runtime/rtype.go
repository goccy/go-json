package runtime

import (
	"reflect"
	"unsafe"
)

// Type representing reflect.rtype.
// It is a thin pointer to the runtime type descriptor, so it can be used as a cache key
// and as the type word of an interface value without going through reflect.Type.
type Type struct{}

func (t *Type) Align() int {
	return RType2Type(t).Align()
}

func (t *Type) FieldAlign() int {
	return RType2Type(t).FieldAlign()
}

func (t *Type) Method(a0 int) reflect.Method {
	return RType2Type(t).Method(a0)
}

func (t *Type) MethodByName(a0 string) (reflect.Method, bool) {
	return RType2Type(t).MethodByName(a0)
}

func (t *Type) NumMethod() int {
	return RType2Type(t).NumMethod()
}

func (t *Type) Name() string {
	return RType2Type(t).Name()
}

func (t *Type) PkgPath() string {
	return RType2Type(t).PkgPath()
}

func (t *Type) Size() uintptr {
	return RType2Type(t).Size()
}

func (t *Type) String() string {
	return RType2Type(t).String()
}

func (t *Type) Kind() reflect.Kind {
	return RType2Type(t).Kind()
}

func (t *Type) Implements(u reflect.Type) bool {
	return RType2Type(t).Implements(u)
}

func (t *Type) AssignableTo(u reflect.Type) bool {
	return RType2Type(t).AssignableTo(u)
}

func (t *Type) ConvertibleTo(u reflect.Type) bool {
	return RType2Type(t).ConvertibleTo(u)
}

func (t *Type) Comparable() bool {
	return RType2Type(t).Comparable()
}

func (t *Type) Bits() int {
	return RType2Type(t).Bits()
}

func (t *Type) ChanDir() reflect.ChanDir {
	return RType2Type(t).ChanDir()
}

func (t *Type) IsVariadic() bool {
	return RType2Type(t).IsVariadic()
}

func (t *Type) Elem() *Type {
	return Type2RType(RType2Type(t).Elem())
}

func (t *Type) Field(i int) reflect.StructField {
	return RType2Type(t).Field(i)
}

func (t *Type) FieldByIndex(index []int) reflect.StructField {
	return RType2Type(t).FieldByIndex(index)
}

func (t *Type) FieldByName(name string) (reflect.StructField, bool) {
	return RType2Type(t).FieldByName(name)
}

func (t *Type) FieldByNameFunc(match func(string) bool) (reflect.StructField, bool) {
	return RType2Type(t).FieldByNameFunc(match)
}

func (t *Type) In(i int) reflect.Type {
	return RType2Type(t).In(i)
}

func (t *Type) Key() *Type {
	return Type2RType(RType2Type(t).Key())
}

func (t *Type) Len() int {
	return RType2Type(t).Len()
}

func (t *Type) NumField() int {
	return RType2Type(t).NumField()
}

func (t *Type) NumIn() int {
	return RType2Type(t).NumIn()
}

func (t *Type) NumOut() int {
	return RType2Type(t).NumOut()
}

func (t *Type) Out(i int) reflect.Type {
	return RType2Type(t).Out(i)
}

func PtrTo(t *Type) *Type {
	return Type2RType(reflect.PointerTo(RType2Type(t)))
}

//go:linkname IfaceIndir reflect.ifaceIndir
//go:noescape
func IfaceIndir(*Type) bool

type emptyInterface struct {
	typ *Type
	ptr unsafe.Pointer
}

// RType2Type converts the type descriptor pointer to reflect.Type.
// reflect.TypeOf reads only the type word of the interface value, so the data word is left nil.
func RType2Type(t *Type) reflect.Type {
	return reflect.TypeOf(*(*interface{})(unsafe.Pointer(&emptyInterface{typ: t})))
}

func Type2RType(t reflect.Type) *Type {
	return (*Type)(((*emptyInterface)(unsafe.Pointer(&t))).ptr)
}

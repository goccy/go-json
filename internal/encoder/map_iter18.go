//go:build go1.18
// +build go1.18

package encoder

import (
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

var hiterField, _ = reflect.TypeOf(reflect.MapIter{}).FieldByName("hiter")

func MapIterInit(mapType *runtime.Type, m unsafe.Pointer) unsafe.Pointer {
	iter := reflect.New(hiterField.Type).UnsafePointer()
	mapIterInit(mapType, m, iter)
	return iter
}

//go:linkname mapIterInit reflect.mapiterinit
//go:noescape
func mapIterInit(mapType *runtime.Type, m, hiter unsafe.Pointer)

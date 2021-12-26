//go:build !go1.18
// +build !go1.18

package encoder

import (
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

//go:linkname MapIterInit reflect.mapiterinit
//go:noescape
func MapIterInit(mapType *runtime.Type, m unsafe.Pointer) unsafe.Pointer

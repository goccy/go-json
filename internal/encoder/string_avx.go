package encoder

import "unsafe"

//go:nosplit
//go:noescape
func _findEscapeIndex64(buf unsafe.Pointer, len int) (ret int)

//go:nosplit
//go:noescape
func _findEscapeIndex128(buf unsafe.Pointer, len int) (ret int)

//go:nosplit
//go:noescape
func _findEscapeIndex256(buf unsafe.Pointer, len int) (ret int)

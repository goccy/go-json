package runtime

import (
	"reflect"
	"unsafe"
)

type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

const (
	maxAcceptableTypeAddrRange = 1024 * 1024 * 2 // 2 Mib
)

type TypeAddr struct {
	BaseTypeAddr uintptr
	MaxTypeAddr  uintptr
	AddrRange    uintptr
}

var (
	typeAddr        *TypeAddr
	alreadyAnalyzed bool
)

//go:linkname typelinks reflect.typelinks
func typelinks() ([]unsafe.Pointer, [][]int32)

//go:linkname rtypeOff reflect.rtypeOff
func rtypeOff(unsafe.Pointer, int32) unsafe.Pointer

func AnalyzeTypeAddr() *TypeAddr {
	defer func() {
		alreadyAnalyzed = true
	}()
	if alreadyAnalyzed {
		return typeAddr
	}
	sections, offsets := typelinks()
	if len(sections) != 1 {
		return nil
	}
	if len(offsets) != 1 {
		return nil
	}
	section := sections[0]
	offset := offsets[0]
	var (
		min uintptr = uintptr(^uint(0))
		max uintptr = 0
	)
	for i := 0; i < len(offset); i++ {
		typ := (*Type)(rtypeOff(section, offset[i]))
		addr := uintptr(unsafe.Pointer(typ))
		if min > addr {
			min = addr
		}
		if max < addr {
			max = addr
		}
		if typ.Kind() == reflect.Ptr {
			addr = uintptr(unsafe.Pointer(typ.Elem()))
			if min > addr {
				min = addr
			}
			if max < addr {
				max = addr
			}
		}
	}
	addrRange := max - min
	if addrRange == 0 {
		return nil
	}
	if addrRange > maxAcceptableTypeAddrRange {
		return nil
	}
	typeAddr = &TypeAddr{
		BaseTypeAddr: min,
		MaxTypeAddr:  max,
		AddrRange:    addrRange,
	}
	return typeAddr
}

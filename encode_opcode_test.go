package json

import (
	"testing"
	"unsafe"
)

func TestDumpOpcode(t *testing.T) {
	var v interface{} = 1
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	typ := header.typ
	typeptr := uintptr(unsafe.Pointer(typ))
	codeSet, err := encodeCompileToGetCodeSet(typeptr)
	if err != nil {
		t.Fatal(err)
	}
	codeSet.code.dump()
}

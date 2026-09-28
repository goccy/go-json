package json_test

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/goccy/go-json/internal/decoder"
)

// probeDecode prints the low bits of the addresses of the objects which a decode of the targets reads, to see
// where the history of the process put them.
func probeDecode(name string, targets []any) {
	if os.Getenv("ZZPROBE") == "" {
		return
	}
	ctx := decoder.TakeRuntimeContext()
	ctxAddr := uintptr(unsafe.Pointer(ctx))
	decoder.ReleaseRuntimeContext(ctx)
	v := targets[0]
	e := (*[2]unsafe.Pointer)(unsafe.Pointer(&v))
	dec, _ := decoder.CompileToGetDecoder(e[0])
	d := (*[2]uintptr)(unsafe.Pointer(&dec))[1]
	fmt.Printf("probe %s ctx=%03x dec=%03x target=%03x input=%03x\n", name, ctxAddr&4095, d&4095, uintptr(e[1])&4095, uintptr(unsafe.Pointer(&lookupInput[0]))&4095)
}

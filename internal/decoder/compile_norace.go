//go:build !race
// +build !race

package decoder

import (
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

func CompileToGetDecoder(typ *runtime.Type) (Decoder, error) {
	initDecoder()
	typeptr := uintptr(unsafe.Pointer(typ))
	if typeptr > typeAddr.MaxTypeAddr {
		return compileToGetDecoderSlowPath(typeptr, typ)
	}

	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if dec := loadCachedDecoder(index); dec != nil {
		return dec, nil
	}

	dec, err := compileHead(typ, map[uintptr]Decoder{})
	if err != nil {
		return nil, err
	}
	return storeCachedDecoder(index, dec), nil
}

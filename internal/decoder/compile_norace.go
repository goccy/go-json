//go:build !race
// +build !race

package decoder

import (
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

func CompileToGetDecoder(typ *runtime.Type, tagName string) (Decoder, error) {
	initDecoder()
	typeptr := uintptr(unsafe.Pointer(typ))
	if typeptr > typeAddr.MaxTypeAddr {
		return compileToGetDecoderSlowPath(typeptr, tagName, typ)
	}

	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if dec := cachedDecoder[index]; dec != nil {
		return dec, nil
	}

	dec, err := compileHead(typ, tagName, map[uintptr]Decoder{})
	if err != nil {
		return nil, err
	}
	cachedDecoder[index] = dec
	return dec, nil
}

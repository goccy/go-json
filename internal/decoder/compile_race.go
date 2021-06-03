// +build race

package decoder

import (
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

var decMu sync.RWMutex

func CompileToGetDecoder(typ *runtime.Type) (decoder, error) {
	typeptr := uintptr(unsafe.Pointer(typ))
	if typeptr > maxTypeAddr {
		return decodeCompileToGetDecoderSlowPath(typeptr, typ)
	}

	index := (typeptr - baseTypeAddr) >> typeAddrShift
	decMu.RLock()
	if dec := cachedDecoder[index]; dec != nil {
		decMu.RUnlock()
		return dec, nil
	}
	decMu.RUnlock()

	dec, err := decodeCompileHead(typ, map[uintptr]decoder{})
	if err != nil {
		return nil, err
	}
	decMu.Lock()
	cachedDecoder[index] = dec
	decMu.Unlock()
	return dec, nil
}

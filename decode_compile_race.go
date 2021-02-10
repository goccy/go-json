// +build race

package json

import (
	"sync"
)

var decMu sync.RWMutex

func decodeCompileToGetDecoder(typeptr uintptr, typ *rtype) (decoder, error) {
	if typeptr > maxTypeAddr {
		return decodeCompileToGetDecoderSlowPath(typeptr, typ)
	}

	index := typeptr - baseTypeAddr
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

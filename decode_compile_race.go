// +build race

package json

import (
	"sync"
)

var decMu sync.RWMutex

func (d *Decoder) compileToGetDecoder(typeptr uintptr, typ *rtype) (decoder, error) {
	if typeptr > maxTypeAddr {
		return d.compileToGetDecoderSlowPath(typeptr, typ)
	}

	index := typeptr - baseTypeAddr
	decMu.RLock()
	if dec := cachedDecoder[index]; dec != nil {
		decMu.RUnlock()
		return dec, nil
	}
	decMu.RUnlock()

	d.structTypeToDecoder = map[uintptr]decoder{}
	dec, err := d.compileHead(typ)
	if err != nil {
		return nil, err
	}
	decMu.Lock()
	cachedDecoder[index] = dec
	decMu.Unlock()
	return dec, nil
}

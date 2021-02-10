// +build !race

package json

func decodeCompileToGetDecoder(typeptr uintptr, typ *rtype) (decoder, error) {
	if typeptr > maxTypeAddr {
		return decodeCompileToGetDecoderSlowPath(typeptr, typ)
	}

	index := typeptr - baseTypeAddr
	if dec := cachedDecoder[index]; dec != nil {
		return dec, nil
	}

	dec, err := decodeCompileHead(typ, map[uintptr]decoder{})
	if err != nil {
		return nil, err
	}
	cachedDecoder[index] = dec
	return dec, nil
}

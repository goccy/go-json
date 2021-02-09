// +build !race

package json

func (d *Decoder) compileToGetDecoder(typeptr uintptr, typ *rtype) (decoder, error) {
	if typeptr > maxTypeAddr {
		return d.compileToGetDecoderSlowPath(typeptr, typ)
	}

	index := typeptr - baseTypeAddr
	if dec := cachedDecoder[index]; dec != nil {
		return dec, nil
	}

	d.structTypeToDecoder = map[uintptr]decoder{}
	dec, err := d.compileHead(typ)
	if err != nil {
		return nil, err
	}
	cachedDecoder[index] = dec
	return dec, nil
}

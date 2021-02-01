// +build !race

package json

import "unsafe"

func encodeCompileToGetCodeSet(typeptr uintptr) (*opcodeSet, error) {
	if !existsCachedOpcodeSets {
		return encodeCompileToGetCodeSetSlowPath(typeptr)
	}
	if codeSet := cachedOpcodeSets[typeptr-baseTypeAddr]; codeSet != nil {
		return codeSet, nil
	}

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := *(**rtype)(unsafe.Pointer(&typeptr))

	code, err := encodeCompileHead(&encodeCompileContext{
		typ:                      copiedType,
		root:                     true,
		structTypeToCompiledCode: map[uintptr]*compiledCode{},
	})
	if err != nil {
		return nil, err
	}
	code = copyOpcode(code)
	codeLength := code.totalLength()
	codeSet := &opcodeSet{
		code:       code,
		codeLength: codeLength,
	}
	cachedOpcodeSets[int(typeptr-baseTypeAddr)] = codeSet
	return codeSet, nil
}

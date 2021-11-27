//go:build race
// +build race

package encoder

import (
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

var setsMu sync.RWMutex

func CompileToGetCodeSet(typeptr uintptr) (*OpcodeSet, error) {
	if typeptr > typeAddr.MaxTypeAddr {
		return compileToGetCodeSetSlowPath(typeptr)
	}
	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	setsMu.RLock()
	if codeSet := cachedOpcodeSets[index]; codeSet != nil {
		setsMu.RUnlock()
		return codeSet, nil
	}
	setsMu.RUnlock()

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := *(**runtime.Type)(unsafe.Pointer(&typeptr))

	noescapeKeyCode, err := compile(&compileContext{
		typ:               copiedType,
		structTypeToCode:  map[uintptr]*StructCode{},
		structTypeToCodes: map[uintptr]Opcodes{},
		recursiveCodes:    &Opcodes{},
	})
	if err != nil {
		return nil, err
	}
	escapeKeyCode, err := compile(&compileContext{
		typ:               copiedType,
		structTypeToCode:  map[uintptr]*StructCode{},
		structTypeToCodes: map[uintptr]Opcodes{},
		recursiveCodes:    &Opcodes{},
		escapeKey:         true,
	})
	if err != nil {
		return nil, err
	}

	noescapeKeyCode = copyOpcode(noescapeKeyCode)
	escapeKeyCode = copyOpcode(escapeKeyCode)
	setTotalLengthToInterfaceOp(noescapeKeyCode)
	setTotalLengthToInterfaceOp(escapeKeyCode)
	interfaceNoescapeKeyCode := copyToInterfaceOpcode(noescapeKeyCode)
	interfaceEscapeKeyCode := copyToInterfaceOpcode(escapeKeyCode)
	codeLength := noescapeKeyCode.TotalLength()
	codeSet := &OpcodeSet{
		Type:                     copiedType,
		NoescapeKeyCode:          noescapeKeyCode,
		EscapeKeyCode:            escapeKeyCode,
		InterfaceNoescapeKeyCode: interfaceNoescapeKeyCode,
		InterfaceEscapeKeyCode:   interfaceEscapeKeyCode,
		CodeLength:               codeLength,
		EndCode:                  ToEndCode(interfaceNoescapeKeyCode),
	}
	setsMu.Lock()
	cachedOpcodeSets[index] = codeSet
	setsMu.Unlock()
	return codeSet, nil
}

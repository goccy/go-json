// +build !race

package compiler

import (
	"github.com/goccy/go-json/internal/encoder"
)

func CompileToGetCodeSet(typeptr uintptr) (*encoder.OpcodeSet, error) {
	return nil, nil
}

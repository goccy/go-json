#!/bin/bash
# Writes the test which prints the sets of the recent opcodes of the types of the map benchmark.
set -e
cat > /tmp/zz_sets_test.go <<'GOEOF'
package benchmark

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestZZSetsOfMapValues(t *testing.T) {
	m := benchMapValue()
	vs := []any{m}
	for _, v := range m {
		vs = append(vs, v)
	}
	for _, v := range vs {
		typ := (*[2]uintptr)(unsafe.Pointer(&v))[0]
		fmt.Printf("%T set=%d\n", v, (uint64(typ)*0x9E3779B97F4A7C15)>>60)
	}
}
GOEOF

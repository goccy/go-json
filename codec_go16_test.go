// +build go1.16

package json

import (
	"testing"
)

func TestTypeAddressAligned32(t *testing.T) {
	if typeAddrShift != 5 {
		t.Fatalf("unexpected type address shift %d, want 5", typeAddrShift)
	}
}

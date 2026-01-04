package encoder

import (
	"reflect"
	"testing"
)

func TestDumpOpcode(t *testing.T) {
	ctx := TakeRuntimeContext()
	defer ReleaseRuntimeContext(ctx)
	var v interface{} = 1
	reflectValue := reflect.ValueOf(v)
	codeSet, err := CompileToGetCodeSetFromValue(ctx, reflectValue)
	if err != nil {
		t.Fatal(err)
	}
	codeSet.EscapeKeyCode.Dump()
}

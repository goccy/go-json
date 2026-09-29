package runtime

import (
	"reflect"
	"testing"
	"time"
)

// Each lookup finds the method of its name, and nothing in a type without it.
func TestMethodLookups(t *testing.T) {
	timePtr := reflect.TypeOf(&time.Time{})
	for name, lookup := range map[string]MethodLookup{
		"MarshalJSON": MarshalJSONMethod, "MarshalText": MarshalTextMethod, "AppendText": AppendTextMethod,
		"UnmarshalJSON": UnmarshalJSONMethod,
	} {
		want, wantOK := timePtr.MethodByName(name)
		got, ok := lookup(timePtr)
		if ok != wantOK || got.Name != want.Name || got.Func.Pointer() != want.Func.Pointer() {
			t.Errorf("%s of *time.Time: got %v %v, want %v %v", name, got.Name, ok, want.Name, wantOK)
		}
		if _, ok := lookup(reflect.TypeOf(0)); ok {
			t.Errorf("%s of int: found", name)
		}
	}
}

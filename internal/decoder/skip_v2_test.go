package decoder

import (
	"strconv"
	"strings"
	"testing"
)

// The sets of the names of the objects which skipValueV2 skips keep nothing of a value after it, whether it ends, an
// error stops it or a name twice: the maps of the objects of many names, which the context of the call would keep,
// are dropped, and the names of the call are the ones before the value.
func TestSkipValueV2KeepsNoNames(t *testing.T) {
	names := make([]string, 40)
	for i := range names {
		names[i] = `"n` + strconv.Itoa(i) + `":1`
	}
	many := "{" + strings.Join(names, ",")
	for _, tc := range []struct {
		name, in string
		valid    bool
	}{
		{"an object of many names", many + "}", true},
		{"an object of many names in an object", `{"a":` + many + `},"b":1}`, true},
		{"an object of many names and a syntax error", many + ",}", false},
		{"an object of many names and a name twice", many + `,"n0":1}`, false},
		{"an object of many names in an object and a syntax error", `{"a":` + many + `,"x":}}`, false},
	} {
		ctx := TakeRuntimeContext()
		ctx.Option.V2 = V2Semantics
		ctx.SetInput([]byte(tc.in))
		_, err := ctx.skipValueV2(0, 0)
		if valid := err == nil; valid != tc.valid {
			t.Errorf("%s: error %v", tc.name, err)
		}
		call := ctx.V2Call()
		for level, s := range call.skipNames {
			if s.set != nil {
				t.Errorf("%s: the map of %d names of the level %d is kept", tc.name, len(s.set), level)
			}
		}
		if n := len(call.names.entries); n != 0 {
			t.Errorf("%s: %d names of the call are kept", tc.name, n)
		}
		ctx.Option.V2 = 0
		ReleaseRuntimeContext(ctx)
	}
}

package decoder

import (
	"strconv"
	"strings"
	"testing"
)

// The sets of the names of the objects of a call share the names of the call: an object's names are its own,
// whatever the objects in it left, and are dropped at its end.
func TestNameSetsOfNestedObjects(t *testing.T) {
	call := new(V2Call)
	add := func(s *nameSet, name string, want bool) {
		t.Helper()
		if got := s.add(call, []byte(name)); got != want {
			t.Fatalf("add %q: got %v, want %v", name, got, want)
		}
	}
	var outer nameSet
	add(&outer, "a", true)
	add(&outer, "b", true)
	add(&outer, "xby", true)

	// an object in it, of the same names, which ends: its names of the same bits, of the same length and the same
	// first and last bytes, are compared with its own names only
	var inner nameSet
	add(&inner, "a", true)
	add(&inner, "c", true)
	add(&inner, "c", false)
	add(&inner, "b", true)
	add(&inner, "xay", true)
	add(&inner, "xby", true)
	inner.end(call)
	if len(call.names.ends) != 3 {
		t.Fatalf("after the end of the inner object: %d names, want 3", len(call.names.ends))
	}

	// an object in it which an error ends: its names are dropped by the next name of the outer one
	var failed nameSet
	add(&failed, "d", true)
	add(&failed, "e", true)
	add(&outer, "c", true)
	add(&outer, "d", true)
	add(&outer, "e", true)
	add(&outer, "a", false)
	if got := len(call.names.ends); got != 6 {
		t.Fatalf("names of the outer object: %d, want 6", got)
	}

	// an object after the names of the outer one, of its names, which are not of the outer one
	var leftAfter nameSet
	add(&outer, "f", true)
	add(&leftAfter, "a", true)
	add(&leftAfter, "x", true)
	leftAfter.end(call)
	outer.end(call)
	if len(call.names.ends) != 0 || len(call.names.buf) != 0 {
		t.Fatalf("after the end of the outer object: %d names of %d bytes", len(call.names.ends), len(call.names.buf))
	}
}

// Past maxListedNames, the names of an object are the keys of a map too, which the names of the call keep: objects
// in it add their names after them, and the array of the names of the call may be left for a larger one.
func TestNameSetOfManyNames(t *testing.T) {
	call := new(V2Call)
	add := func(s *nameSet, name string, want bool) {
		t.Helper()
		if got := s.add(call, []byte(name)); got != want {
			t.Fatalf("add %q: got %v, want %v", name, got, want)
		}
	}
	var outer, many nameSet
	add(&outer, "z", true)
	for i := 0; i <= maxListedNames+8; i++ {
		add(&many, "n"+strconv.Itoa(i), true)
	}
	if many.set == nil || len(call.names.ends) != maxListedNames+10 {
		t.Fatalf("set %v, %d names of the call, want a map and %d", many.set != nil, len(call.names.ends), maxListedNames+10)
	}
	for _, name := range []string{"n0", "n5", "n" + strconv.Itoa(maxListedNames+8)} {
		add(&many, name, false)
	}
	// an object in it, of a long name, which leaves the array of the names for a larger one
	var inner nameSet
	add(&inner, "n0", true)
	add(&inner, strings.Repeat("x", 4*cap(call.names.buf)), true)
	inner.end(call)
	add(&many, "z", true)
	for _, name := range []string{"n0", "n1", "n" + strconv.Itoa(maxListedNames+8), "z"} {
		add(&many, name, false)
	}
	many.end(call)
	if len(call.names.ends) != 1 {
		t.Fatalf("after the end of the object of many names: %d names, want 1", len(call.names.ends))
	}
	add(&outer, "z", false)
	add(&outer, "n0", true)
}

// truncate back to no name, as the release of a context does, drops the names which objects that errors ended left,
// and the arrays which long names made large.
func TestCallNamesRelease(t *testing.T) {
	call := new(V2Call)
	var s nameSet
	s.add(call, []byte("a"))
	call.names.truncate(0)
	if len(call.names.ends) != 0 || len(call.names.buf) != 0 {
		t.Fatalf("after release: %d names of %d bytes", len(call.names.ends), len(call.names.buf))
	}
	long := make([]byte, 70<<10)
	var big nameSet
	big.add(call, long)
	call.names.truncate(0)
	if call.names.buf != nil || call.names.ends != nil {
		t.Fatalf("after release of a long name: arrays of %d bytes and %d ends kept", cap(call.names.buf), cap(call.names.ends))
	}
}

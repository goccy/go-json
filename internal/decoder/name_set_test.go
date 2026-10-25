package decoder

import (
	"strconv"
	"strings"
	"testing"
)

// The sets of the names of the objects of a call share the names of the call: an object's names are its own, after
// the ones of the objects which it is in, and are dropped at its end.
func TestNameSetsOfNestedObjects(t *testing.T) {
	call := new(V2Call)
	add := func(s *nameSet, name string, want bool) {
		t.Helper()
		if got := s.addCopy(call, nil, []byte(name)); got != want {
			t.Fatalf("add %q: got %v, want %v", name, got, want)
		}
	}
	var outer nameSet
	add(&outer, "a", true)
	add(&outer, "b", true)
	add(&outer, "xby", true)

	// an object in it, of the same names, which ends
	var inner nameSet
	add(&inner, "a", true)
	add(&inner, "c", true)
	add(&inner, "c", false)
	add(&inner, "b", true)
	add(&inner, "xay", true)
	add(&inner, "xby", true)
	inner.end(call)
	if got := len(call.names.entries); got != 3 {
		t.Fatalf("after the end of the inner object: %d names, want 3", got)
	}

	add(&outer, "c", true)
	add(&outer, "d", true)
	add(&outer, "e", true)
	add(&outer, "a", false)
	if got := len(call.names.entries); got != 6 {
		t.Fatalf("names of the outer object: %d, want 6", got)
	}

	// an object after the names of the outer one, of its names, which are not of the outer one
	var leftAfter nameSet
	add(&outer, "f", true)
	add(&leftAfter, "a", true)
	add(&leftAfter, "x", true)
	leftAfter.end(call)
	outer.end(call)
	if len(call.names.entries) != 0 || len(call.names.copies) != 0 {
		t.Fatalf("after the end of the outer object: %d entries, %d bytes of copies", len(call.names.entries), len(call.names.copies))
	}
}

// Past maxListedNames, the names of an object are the keys of a map too, which the names of the call keep: objects
// in it add their names after them, and the array of the copies of the call may be left for a larger one.
func TestNameSetOfManyNames(t *testing.T) {
	call := new(V2Call)
	add := func(s *nameSet, name string, want bool) {
		t.Helper()
		if got := s.addCopy(call, nil, []byte(name)); got != want {
			t.Fatalf("add %q: got %v, want %v", name, got, want)
		}
	}
	var outer, many nameSet
	add(&outer, "z", true)
	const n = 1000
	for i := range n {
		add(&many, "n"+strconv.Itoa(i), true)
	}
	if many.set == nil || many.n != n || len(call.names.entries) != n+1 {
		t.Fatalf("set %v, %d names, %d names of the call, want a map, %d and %d", many.set != nil, many.n, len(call.names.entries), n, n+1)
	}
	for i := range n {
		add(&many, "n"+strconv.Itoa(i), false)
	}
	var inner nameSet
	add(&inner, "n0", true)
	add(&inner, strings.Repeat("x", 4*cap(call.names.copies)), true)
	inner.end(call)
	for i := n; i < 2*n; i++ {
		add(&many, "n"+strconv.Itoa(i), true)
	}
	add(&many, "z", true)
	for _, name := range []string{"n0", "n1", "n" + strconv.Itoa(2*n-1), "z"} {
		add(&many, name, false)
	}
	many.end(call)
	if got := len(call.names.entries); got != 1 {
		t.Fatalf("after the end of the object of many names: %d names, want 1", got)
	}
	add(&outer, "z", false)
	add(&outer, "n0", true)
}

// A name of the input and a copy of the same name are the same name, and names which differ by any of their bytes,
// or by their lengths, are not, whatever their lengths: the bytes between the first and the last eight of a long
// name too.
func TestNameSetOfNamesOfTheInput(t *testing.T) {
	call := new(V2Call)
	var names []string
	for n := 0; n <= 40; n++ {
		base := []byte(strings.Repeat("a", n))
		names = append(names, string(base))
		for i := range n {
			b := append([]byte(nil), base...)
			b[i] = 'b'
			names = append(names, string(b))
		}
	}
	names = append(names, "\x00", "\x00\x00", "a\x00")
	var input []byte
	var at []int64
	for _, name := range names {
		at = append(at, int64(len(input)))
		input = append(input, name...)
	}
	// sets of a few names, which are compared by their hashes, and one of all, of a map
	sets := [][2]int{{0, len(names)}}
	for first := 0; first < len(names); first += 7 {
		sets = append(sets, [2]int{first, min(first+7, len(names))})
	}
	for _, set := range sets {
		names, at := names[set[0]:set[1]], at[set[0]:set[1]]
		var s nameSet
		for i, name := range names {
			if !s.add(call, input, at[i], at[i]+int64(len(name))) {
				t.Fatalf("add %q of the input: had it", name)
			}
		}
		for _, name := range names {
			if s.addCopy(call, input, []byte(name)) {
				t.Fatalf("add a copy of %q: didn't have it", name)
			}
		}
		s.end(call)
		var copies nameSet
		for _, name := range names {
			if !copies.addCopy(call, input, []byte(name)) {
				t.Fatalf("add a copy of %q to a set of copies: had it", name)
			}
		}
		for i, name := range names {
			if copies.add(call, input, at[i], at[i]+int64(len(name))) {
				t.Fatalf("add %q of the input to the set of its copies: didn't have it", name)
			}
		}
		copies.end(call)
	}
}

// truncate back to no name, as the release of a context does, drops the names which objects that errors ended left,
// and the arrays which long names made large.
func TestCallNamesRelease(t *testing.T) {
	call := new(V2Call)
	var s nameSet
	s.addCopy(call, nil, []byte("a"))
	call.names.truncate(0, 0)
	if len(call.names.entries) != 0 || len(call.names.copies) != 0 {
		t.Fatalf("after release: %d entries, %d bytes of copies", len(call.names.entries), len(call.names.copies))
	}
	var big nameSet
	big.addCopy(call, nil, make([]byte, 70<<10))
	call.names.truncate(0, 0)
	if call.names.entries != nil || call.names.copies != nil {
		t.Fatalf("after release of a long name: arrays of %d entries and %d bytes kept", cap(call.names.entries), cap(call.names.copies))
	}
	var many nameSet
	for i := range 10000 {
		many.addCopy(call, nil, []byte(strconv.Itoa(i)))
	}
	many.end(call)
	if call.names.entries != nil || call.names.copies != nil {
		t.Fatalf("after the end of an object of many names: arrays of %d entries and %d bytes kept", cap(call.names.entries), cap(call.names.copies))
	}
}

package decoder

import (
	"encoding/binary"
	"math/bits"
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
// name too. The names of the input are added by add, and by addListed before add, as the loops of the decoding add
// most names: the input has no room after it, so that its last names are added by add.
func TestNameSetOfNamesOfTheInput(t *testing.T) {
	for _, listed := range []bool{false, true} {
		t.Run(map[bool]string{false: "add", true: "addListed"}[listed], func(t *testing.T) {
			testNameSetOfNamesOfTheInput(t, listed)
		})
	}
}

func testNameSetOfNamesOfTheInput(t *testing.T, listed bool) {
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
	input = input[:len(input):len(input)]
	add := func(s *nameSet, start, end int64) bool {
		if listed && s.addListed(call, input, start, end) {
			return true
		}
		return s.add(call, input, start, end)
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
			if !add(&s, at[i], at[i]+int64(len(name))) {
				t.Fatalf("add %q of the input: had it", name)
			}
		}
		for i, name := range names {
			if add(&s, at[i], at[i]+int64(len(name))) {
				t.Fatalf("add %q of the input again: didn't have it", name)
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
			if add(&copies, at[i], at[i]+int64(len(name))) {
				t.Fatalf("add %q of the input to the set of its copies: didn't have it", name)
			}
		}
		copies.end(call)
	}
}

// A set which addListed starts, in an object after the names of the object around it, begins after them: its end
// drops its names alone.
func TestNameSetOfAddListedAfterAnObject(t *testing.T) {
	call := new(V2Call)
	// room for the entries, so that addListed takes the name
	call.names.entries = make([]nameEntry, 0, 4)
	input := []byte(`"b"` + strings.Repeat(" ", 16))
	var outer, inner nameSet
	outer.addCopy(call, nil, []byte("a"))
	if !inner.addListed(call, input, 1, 2) {
		t.Fatal("addListed didn't take the name")
	}
	inner.end(call)
	if len(call.names.entries) != 1 {
		t.Fatalf("after the end of the inner object: %d names, want 1", len(call.names.entries))
	}
	if outer.addCopy(call, nil, []byte("a")) {
		t.Fatal("the outer object lost its name")
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

// oneHashName returns the name i of 16 bytes of ASCII, whose hash is the one of every such name ( see nameHash ): its
// first word differs by the bits of i in its bytes, and its last word by the same bits rotated back by the 29 bits
// which nameHash rotates the last word by, so that the rotation and the xor of the two words cancel them.
func oneHashName(i int) []byte {
	var d uint64
	for k := range 8 {
		d |= uint64(i>>(3*k)&7) << (8 * k)
	}
	name := make([]byte, 16)
	binary.LittleEndian.PutUint64(name, 0x4040404040404040^d)
	binary.LittleEndian.PutUint64(name[8:], 0x4040404040404040^bits.RotateLeft64(d, -29))
	return name
}

// The names of an object of many names which have one hash are told apart by their bytes, and not compared with one
// another: past maxListedNames, they are keys of the map of the set, whose hash the input doesn't choose, so that the
// time of an object of many names of one hash is not the square of their number.
func TestNameSetOfNamesOfOneHash(t *testing.T) {
	const n = 4 * maxListedNames
	if h := nameHash(oneHashName(0)); nameHash(oneHashName(n-1)) != h || nameHash(oneHashName(n/2)) != h {
		t.Fatal("the names have other hashes")
	}
	var input []byte
	at := make([]int64, n)
	for i := range n {
		at[i] = int64(len(input))
		input = append(input, oneHashName(i)...)
	}
	call := new(V2Call)
	var s nameSet
	for i := range n {
		if s.addListed(call, input, at[i], at[i]+16) || s.add(call, input, at[i], at[i]+16) {
			continue
		}
		t.Fatalf("name %d: had it", i)
	}
	if len(s.set) != n {
		t.Fatalf("%d names of one hash, of which %d are keys of the map of the set", n, len(s.set))
	}
	for i := range n {
		if _, ok := s.set[string(oneHashName(i))]; !ok {
			t.Fatalf("name %d is not a key of the map of the set", i)
		}
	}
	if s.add(call, input, at[n/2], at[n/2]+16) {
		t.Fatal("didn't have a name twice")
	}
}

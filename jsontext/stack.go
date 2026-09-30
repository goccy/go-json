package jsontext

import (
	"hash/maphash"
	"strconv"
)

// maxDepth is the deepest nesting of objects and arrays which an encoder or decoder takes.
const maxDepth = 10000

// level is a level of the nesting: the top level of the stream, or an open object or array. Its count is the
// number of tokens in it: the values of the top level or an array, and the names and values of an object.
type level struct {
	count  int64
	object bool
	first  int // for an object, the index in names.ends of its first name
	last   int // for an object, the index in names.ends of its last name
}

// stack is the state of the grammar: the levels of the nesting, the top level first, and the names of the open
// objects.
type stack struct {
	levels []level
	names  names
}

func (s *stack) reset() {
	s.levels = append(s.levels[:0], level{})
	s.names.reset()
}

// depth is the number of open objects and arrays.
func (s *stack) depth() int { return len(s.levels) - 1 }

func (s *stack) last() *level { return &s.levels[len(s.levels)-1] }

// push opens an object or an array, whose begin token is a value of the level where it is.
func (s *stack) push(object bool) error {
	if len(s.levels) > maxDepth {
		return errMaxDepth
	}
	s.last().count++
	s.levels = append(s.levels, level{object: object, first: len(s.names.ends)})
	return nil
}

// pop closes the innermost object or array.
func (s *stack) pop() {
	if l := s.last(); l.object {
		s.names.truncate(l.first)
	}
	s.levels = s.levels[:len(s.levels)-1]
}

// Where the innermost level is in its grammar.
func (l *level) needName() bool  { return l.object && l.count%2 == 0 }
func (l *level) needValue() bool { return l.object && l.count%2 == 1 }

// The pointers of a stack point to a value in the innermost level: pointLast to the last one which was read or
// written, pointNext to the one after it where a value is next, and pointAt to the level itself, or to the
// value of the last name in an object whose value is next.
const (
	pointLast = iota
	pointNext
	pointAt
	pointOut  // pointAt of the level which contains the innermost one, as if it was closed
	pointHere // the innermost object or array itself
)

// pointer is the JSON pointer of the stack at the value which where chooses.
func (s *stack) pointer(where int) Pointer {
	if len(s.levels) == 1 {
		return ""
	}
	return Pointer(s.pointerBytes(where))
}

// namePointer is the pointer to the name of a member of the innermost object, which is not inserted.
func (s *stack) namePointer(name []byte) Pointer {
	return Pointer(appendPointerToken(s.pointerBytes(pointAt), name))
}

func (s *stack) pointerBytes(where int) []byte {
	var b []byte
	n := len(s.levels)
	switch where {
	case pointOut:
		n--
		where = pointAt
	case pointHere:
		n--
		where = pointLast
	}
	for i := 1; i < n; i++ {
		l := &s.levels[i]
		innermost := i == n-1
		switch {
		case l.object:
			// the last name, while its value or a value in it is read, and after it was read if pointLast.
			if l.count > 0 && (l.count%2 == 1 || !innermost || where == pointLast) {
				b = appendPointerToken(b, s.names.get(l.last))
			}
		case !innermost || where == pointLast:
			if l.count > 0 {
				b = strconv.AppendInt(append(b, '/'), l.count-1, 10)
			}
		case where == pointNext:
			b = strconv.AppendInt(append(b, '/'), l.count, 10)
		}
	}
	return b
}

// insertName adds the unquoted name to the innermost object, which must need a name, and reports false if check
// is set and the object has it already. Without check, the object keeps only its last name, which the pointers
// need.
func (s *stack) insertName(name []byte, check bool) bool {
	l := s.last()
	n := &s.names
	if !check {
		n.truncate(l.first)
	} else if n.has(l.first, name) {
		return false
	}
	n.add(l.first, name)
	l.last = len(n.ends) - 1
	return true
}

// names are the names of the open objects, back to back, an object after the objects which contain it.
type names struct {
	buf  []byte // the names, unquoted
	ends []int  // the end of each name in buf
	// the indexes of the objects of many names, the innermost last.
	indexes []nameIndex
}

// A name is looked for in the names of its object one by one while they are few, and over maxLinearNames or
// maxLinearBytes of them, through an index of their hashes: the limits over which encoding/json/jsontext takes
// a map.
const (
	maxLinearNames = 64
	maxLinearBytes = 1024
)

// nameIndex is an index of the names of an object: a table of open addressing, whose size is a power of two,
// of the positions of its names plus one, by their hashes.
type nameIndex struct {
	first  int // the index in ends of the first name of the object
	slots  []int32
	hashes []uint64 // the hash of each name of the object
}

// nameSeed is the seed of the hashes, chosen at random as the one of a Go map is, so that no input can have
// many names of the same hash.
var nameSeed = maphash.MakeSeed()

func (n *names) reset() {
	n.buf = n.buf[:0]
	n.ends = n.ends[:0]
	n.indexes = n.indexes[:0]
}

// truncate drops the names from the index first on: the names of an object which is closed.
func (n *names) truncate(first int) {
	if k := len(n.indexes); k > 0 && n.indexes[k-1].first >= first {
		n.indexes = n.indexes[:k-1]
	}
	n.buf = n.buf[:n.start(first)]
	n.ends = n.ends[:first]
}

// start is the position in buf of the name at the index i.
func (n *names) start(i int) int {
	if i == 0 {
		return 0
	}
	return n.ends[i-1]
}

func (n *names) get(i int) []byte { return n.buf[n.start(i):n.ends[i]] }

// index is the index of the object whose first name is at first, or nil.
func (n *names) index(first int) *nameIndex {
	if k := len(n.indexes); k > 0 && n.indexes[k-1].first == first {
		return &n.indexes[k-1]
	}
	return nil
}

// has reports whether the object whose first name is at first has the name.
func (n *names) has(first int, name []byte) bool {
	if x := n.index(first); x != nil {
		h := maphash.Bytes(nameSeed, name)
		mask := uint64(len(x.slots) - 1)
		for slot := h & mask; x.slots[slot] != 0; slot = (slot + 1) & mask {
			i := int(x.slots[slot]) - 1
			if x.hashes[i] == h && string(n.get(first+i)) == string(name) {
				return true
			}
		}
		return false
	}
	start := n.start(first)
	for _, end := range n.ends[first:] {
		if string(n.buf[start:end]) == string(name) {
			return true
		}
		start = end
	}
	return false
}

// add appends a name to the innermost object, whose first name is at first.
func (n *names) add(first int, name []byte) {
	n.buf = append(n.buf, name...)
	n.ends = append(n.ends, len(n.buf))
	if x := n.index(first); x != nil {
		x.hashes = append(x.hashes, maphash.Bytes(nameSeed, name))
		if 2*len(x.hashes) > len(x.slots) {
			x.rebuild(2 * len(x.slots))
		} else {
			x.insert(len(x.hashes) - 1)
		}
		return
	}
	if len(n.ends)-first > maxLinearNames || len(n.buf)-n.start(first) > maxLinearBytes {
		n.indexObject(first)
	}
}

// indexObject makes the index of the names of the innermost object, whose first name is at first.
func (n *names) indexObject(first int) {
	n.indexes = append(n.indexes, nameIndex{first: first})
	x := &n.indexes[len(n.indexes)-1]
	for i := first; i < len(n.ends); i++ {
		x.hashes = append(x.hashes, maphash.Bytes(nameSeed, n.get(i)))
	}
	x.rebuild(4 * maxLinearNames)
}

// rebuild makes the table of the given size, a power of two, of all the names.
func (x *nameIndex) rebuild(size int) {
	x.slots = make([]int32, size)
	for i := range x.hashes {
		x.insert(i)
	}
}

// insert puts the name at the index i of the object in the table.
func (x *nameIndex) insert(i int) {
	mask := uint64(len(x.slots) - 1)
	slot := x.hashes[i] & mask
	for x.slots[slot] != 0 {
		slot = (slot + 1) & mask
	}
	x.slots[slot] = int32(i + 1)
}

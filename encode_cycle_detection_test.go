package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"runtime"
	"testing"

	"github.com/goccy/go-json"
)

type cycleDetectionHolder struct {
	V any
}

type cycleDetectionNode struct {
	Next *cycleDetectionNode
	Any  any
}

// nestedInterfaceValue wraps the value by the slices of an interface value.
func nestedInterfaceValue(depth int, v any) any {
	for i := 0; i < depth; i++ {
		v = []any{v}
	}
	return v
}

// firstFieldMapList is a list whose first field is a map.
type firstFieldMapList struct {
	M    map[string]firstFieldMapList
	Tail *firstFieldMapList
}

// linkedItem is an item of a list linked through interface values, whose first field is not a pointer.
type linkedItem struct {
	N    int
	Next any
}

// firstFieldShared is a struct whose first field is a map, which another field holds again.
type firstFieldShared struct {
	M map[string]any
	X any
}

type outerOfFirstField struct{ In innerOfFirstField }

type innerOfFirstField struct{ X any }

// directNode is stored directly in an interface value: its data word is the pointer it holds.
type directNode struct{ Next *directNode }

type directHolder struct{ P *cycleDetectionNode }

// The values are checked for a cycle after the nesting gets deep.
// A value which is referred to twice is not a cycle, whatever it holds.
func TestEncodeSharedValueNestedDeeply(t *testing.T) {
	const depth = 1100
	shared := &cycleDetectionHolder{V: (*int)(nil)}
	sharedMap := map[string]any{"a": 1}
	sharedValues := map[string]cycleMapOfValues{"a": {}}
	sharedValuesOfList := map[string]firstFieldMapList{"a": {}}
	var linked any
	for range depth {
		linked = &linkedItem{N: 5000, Next: linked}
	}
	outer := &outerOfFirstField{In: innerOfFirstField{X: 1}}
	for _, test := range []struct {
		name string
		v    any
	}{
		{"typed nil in interface", []*cycleDetectionHolder{shared, shared}},
		{"value in interface", []any{shared, shared}},
		{"recursive type", []*cycleDetectionNode{{Any: shared}, {Any: shared}}},
		{"map", []any{sharedMap, sharedMap}},
		{"map of values", []cycleMapOfValues{{M: sharedValues}, {M: sharedValues}}},
		// a pointer in an interface value, to a struct whose first field is a map which is shared
		{"pointer to a list of a map", &firstFieldMapList{M: sharedValuesOfList, Tail: &firstFieldMapList{M: sharedValuesOfList}}},
		// pointers in interface values, which are the values, not the first words they point to
		{"list through interfaces", linked},
		{"pointer to a struct of a map", &firstFieldShared{M: sharedMap, X: sharedMap}},
		// values of other types at the same address
		{"struct and its first field", []any{outer, &outer.In}},
		// a value stored directly in an interface value, and the value which the pointer it holds points to
		{"struct in interface", directNode{Next: &directNode{Next: &directNode{}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			v := nestedInterfaceValue(depth, test.v)
			expected, err := stdjson.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(expected, got) {
				t.Fatalf("expected %d bytes but got %d bytes", len(expected), len(got))
			}
		})
	}
}

func TestEncodeCycleNestedDeeply(t *testing.T) {
	t.Run("interface", func(t *testing.T) {
		cycle := []any{nil}
		cycle[0] = cycle
		for _, depth := range []int{0, 1100} {
			if _, err := json.Marshal(nestedInterfaceValue(depth, cycle)); err == nil {
				t.Fatalf("depth %d: expected an error", depth)
			}
		}
	})
	t.Run("recursive type", func(t *testing.T) {
		node := &cycleDetectionNode{}
		node.Next = &cycleDetectionNode{Next: node}
		for _, depth := range []int{0, 1100} {
			if _, err := json.Marshal(nestedInterfaceValue(depth, node)); err == nil {
				t.Fatalf("depth %d: expected an error", depth)
			}
		}
	})
	// a pointer in an interface value is recorded by itself, not by the first word it points to, which is nil here.
	t.Run("pointer in interface", func(t *testing.T) {
		node := &cycleDetectionNode{}
		node.Any = node
		if _, err := json.Marshal(node); err == nil {
			t.Fatal("expected an error")
		}
	})
	// a struct stored directly in an interface value, which holds the pointer of the cycle.
	t.Run("struct in interface", func(t *testing.T) {
		node := &cycleDetectionNode{}
		node.Any = directHolder{P: node}
		if _, err := json.Marshal(node); err == nil {
			t.Fatal("expected an error")
		}
	})
	// the value of a map is a copy, but the map it holds is the same: it is the cycle.
	t.Run("map of interface values", func(t *testing.T) {
		m := map[string]any{}
		m["k"] = m
		if _, err := json.Marshal(m); err == nil {
			t.Fatal("expected an error")
		}
	})
	// two slices of the same array which differ in length are not the same value.
	t.Run("slices of an array", func(t *testing.T) {
		v := []any{nil, nil}
		v[1] = v[:1]
		got, err := json.Marshal(nestedInterfaceValue(1100, v))
		want, wantErr := stdjson.Marshal(nestedInterfaceValue(1100, v))
		if err != nil || wantErr != nil || !bytes.Equal(got, want) {
			t.Fatalf("got %d bytes, %v; want %d bytes, %v", len(got), err, len(want), wantErr)
		}
	})
}

// A slice, a map or a pointer type which is a value of itself, not through a struct.
type (
	recursiveSliceType []recursiveSliceType
	recursiveMapType   map[string]recursiveMapType
	recursivePtrType   *recursivePtrType
	recursiveMapField  map[string]struct {
		X recursiveMapField `json:",omitempty"`
		Y recursiveMapField
	}
)

func TestEncodeRecursiveNonStructType(t *testing.T) {
	var p recursivePtrType
	p2 := recursivePtrType(&p)
	for _, v := range []any{
		recursiveSliceType{nil, recursiveSliceType{}, recursiveSliceType{recursiveSliceType{nil}}},
		recursiveMapType{"a": nil, "b": recursiveMapType{}, "c": recursiveMapType{"d": recursiveMapType{}}},
		p, p2, &p2,
		recursiveMapField{"a": {X: recursiveMapField{}, Y: nil}, "b": {X: recursiveMapField{"c": {}}}},
		[]recursiveSliceType{nil, {}},
		map[string]recursiveMapType{"x": {"y": nil}},
	} {
		got, err := json.Marshal(v)
		want, wantErr := stdjson.Marshal(v)
		if !bytes.Equal(got, want) || (err == nil) != (wantErr == nil) {
			t.Errorf("%T: got %s, %v; want %s, %v", v, got, err, want, wantErr)
		}
	}
	m := recursiveMapType{}
	m["k"] = m
	if _, err := json.Marshal(m); err == nil {
		t.Error("map: expected the error of a cycle")
	}
	s := recursiveSliceType{nil}
	s[0] = s
	if _, err := json.Marshal(s); err == nil {
		t.Error("slice: expected the error of a cycle")
	}
}

// Pointer types which are values of themselves through what they point to. The opcodes of a pointer type are the
// ones of the value it points to, which the code of such a value in itself follows the pointer to.
type (
	recursivePtrToSlice      *[]recursivePtrToSlice
	recursivePtrToMap        *map[string]recursivePtrToMap
	recursivePtrToArray      *[1]recursivePtrToArray
	recursivePtrToPair       *[2]recursivePtrToPair
	recursivePtrToPtrToSlice **[]recursivePtrToPtrToSlice
	recursivePtrToStruct     *recursiveStructOfPtr
	recursiveStructOfPtr     struct {
		A int
		F recursivePtrToStruct
	}
	recursivePtrToStructOfSlice *struct{ F []recursivePtrToStructOfSlice }
	recursivePtrToTagged        *recursiveTaggedStruct
	recursiveTaggedStruct       struct {
		A int
		F recursivePtrToTagged `json:",omitempty"`
		G recursivePtrToTagged `json:"g,string"`
		H []any
	}
)

func ptrToSlice(elems ...recursivePtrToSlice) recursivePtrToSlice {
	s := append([]recursivePtrToSlice{}, elems...)
	return &s
}

func ptrToMap(m map[string]recursivePtrToMap) recursivePtrToMap { return &m }

func ptrToPtrToSlice(elems ...recursivePtrToPtrToSlice) recursivePtrToPtrToSlice {
	s := append([]recursivePtrToPtrToSlice{}, elems...)
	p := &s
	return &p
}

func ptrToStructOfSlice(elems ...recursivePtrToStructOfSlice) recursivePtrToStructOfSlice {
	return &struct{ F []recursivePtrToStructOfSlice }{F: elems}
}

// throughContainers returns each value and the values which hold it in a pointer, a slice, an array, a field of a
// struct, a map and interface values: each of them is a level of opcodes before the ones of the value.
func throughContainers[T any](values ...T) []any {
	var out []any
	for _, v := range values {
		out = append(out, v, &v, []T{v}, [1]T{v}, struct{ F T }{v}, struct {
			F T `json:",omitempty"`
		}{v}, map[string]T{"a": v}, []any{v}, map[string]any{"a": &v})
	}
	return out
}

func TestEncodeRecursivePointerType(t *testing.T) {
	var nilSlice []recursivePtrToSlice
	s1 := ptrToSlice()
	s2 := ptrToSlice(nil, s1)
	s3 := ptrToSlice(s2, &nilSlice, nil)
	s4 := ptrToSlice(s3, s1)
	// the opcodes of the type are compiled from the slice it points to, where the type is not yet on the path, so
	// they hold two levels and a frame of the type is entered every second level: 2100 levels are 1050 frames, past
	// StartDetectingCyclesAfter.
	deepSlice := ptrToSlice()
	for range 2100 {
		deepSlice = ptrToSlice(deepSlice)
	}
	cycleSlice := []recursivePtrToSlice{nil}
	cycleSlice[0] = &cycleSlice

	m1 := ptrToMap(map[string]recursivePtrToMap{})
	m2 := ptrToMap(map[string]recursivePtrToMap{"a": nil, "b": m1})
	m3 := ptrToMap(map[string]recursivePtrToMap{"c": m2, "d": ptrToMap(nil)})
	m4 := ptrToMap(map[string]recursivePtrToMap{"e": m3, "f": m1})
	cycleMap := map[string]recursivePtrToMap{}
	cycleMap["a"] = &cycleMap

	a1 := recursivePtrToArray(&[1]recursivePtrToArray{})
	a2 := recursivePtrToArray(&[1]recursivePtrToArray{a1})
	a3 := recursivePtrToArray(&[1]recursivePtrToArray{a2})
	a4 := recursivePtrToArray(&[1]recursivePtrToArray{a3})
	var cycleArray [1]recursivePtrToArray
	cycleArray[0] = &cycleArray

	p1 := recursivePtrToPair(&[2]recursivePtrToPair{})
	p2 := recursivePtrToPair(&[2]recursivePtrToPair{nil, p1})
	p3 := recursivePtrToPair(&[2]recursivePtrToPair{p2, p1})
	p4 := recursivePtrToPair(&[2]recursivePtrToPair{p3, nil})

	var nilInner *[]recursivePtrToPtrToSlice
	q1 := ptrToPtrToSlice()
	q2 := ptrToPtrToSlice(nil, q1)
	q3 := ptrToPtrToSlice(q2, &nilInner)
	q4 := ptrToPtrToSlice(q3, q1)

	r1 := recursivePtrToStruct(&recursiveStructOfPtr{A: 1})
	r2 := recursivePtrToStruct(&recursiveStructOfPtr{A: 2, F: r1})
	r3 := recursivePtrToStruct(&recursiveStructOfPtr{A: 3, F: r2})
	r4 := recursivePtrToStruct(&recursiveStructOfPtr{A: 4, F: r3})
	deepStruct := recursivePtrToStruct(&recursiveStructOfPtr{})
	for i := range 1100 {
		deepStruct = &recursiveStructOfPtr{A: i, F: deepStruct}
	}
	cycleStruct := &recursiveStructOfPtr{A: 1}
	cycleStruct.F = cycleStruct

	l1 := ptrToStructOfSlice()
	l2 := ptrToStructOfSlice(l1, nil)
	l3 := ptrToStructOfSlice(l2, ptrToStructOfSlice(ptrToStructOfSlice()))
	l4 := ptrToStructOfSlice(l3, l1)

	g1 := recursivePtrToTagged(&recursiveTaggedStruct{A: 1})
	g2 := recursivePtrToTagged(&recursiveTaggedStruct{A: 2, F: g1, H: []any{g1, recursivePtrToSlice(nil), s2}})
	g3 := recursivePtrToTagged(&recursiveTaggedStruct{A: 3, F: g2, G: g2})
	g4 := recursivePtrToTagged(&recursiveTaggedStruct{A: 4, G: g3, H: []any{g3, a3}})

	var values []any
	values = append(values, throughContainers(nil, s1, s2, s3, s4)...)
	values = append(values, throughContainers(nil, m1, m2, m3, m4)...)
	values = append(values, throughContainers(nil, a1, a2, a3, a4)...)
	values = append(values, throughContainers(nil, p1, p2, p3, p4)...)
	values = append(values, throughContainers(nil, q1, q2, q3, q4)...)
	values = append(values, throughContainers(nil, r1, r2, r3, r4)...)
	values = append(values, throughContainers(nil, l1, l2, l3, l4)...)
	values = append(values, throughContainers(nil, g1, g2, g3, g4)...)
	// the values nested past the start of the detection of cycles, and the cycles, which fail as with encoding/json.
	values = append(values, deepSlice, []recursivePtrToStruct{deepStruct}, &cycleSlice, &cycleMap, &cycleArray,
		[]recursivePtrToStruct{cycleStruct})
	checkStdEncoding(t, values...)
}

// cycleMapOfValues is a cycle through the values of a map, which are copies: the map alone is reached again.
type cycleMapOfValues struct{ M map[string]cycleMapOfValues }

// cycleMap is a map which holds itself, through its values of interface{} and of its own type.
type cycleMap map[string]cycleMap

func TestEncodeCycleThroughMapValues(t *testing.T) {
	m := map[string]cycleMapOfValues{}
	m["a"] = cycleMapOfValues{M: m}
	self := cycleMap{}
	self["a"] = self
	anyMap := map[string]any{}
	anyMap["a"] = anyMap
	for _, v := range []any{m, &cycleMapOfValues{M: m}, self, anyMap} {
		_, err := json.Marshal(v)
		_, stdErr := stdjson.Marshal(v)
		if err == nil || stdErr == nil || err.Error() != stdErr.Error() {
			t.Errorf("%T: %v; want %v", v, err, stdErr)
		}
	}
}

// retainedMap is a list of maps, whose last value holds memory which the encoding doesn't write.
type retainedMap struct {
	M map[string]retainedMap
	P []byte `json:"-"`
}

// The records of the maps for the detection of cycles are not kept by the pooled context after the call.
func TestEncodeMapRecordsNotRetained(t *testing.T) {
	if raceEnabled {
		t.Skip("the runtime context, which keeps the records, is dropped from its pool at random")
	}
	const size = 32 << 20
	heap := func() uint64 {
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return m.HeapAlloc
	}
	before := heap()
	func() {
		v := map[string]retainedMap{"a": {P: make([]byte, size)}}
		for range 1100 {
			v = map[string]retainedMap{"a": {M: v}}
		}
		if _, err := json.Marshal(v); err != nil {
			t.Fatal(err)
		}
	}()
	if after := heap(); after > before+size/2 {
		t.Errorf("%d MiB is kept alive after the call", (after-before)>>20)
	}
}

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

// The values are checked for a cycle after the nesting gets deep.
// A value which is referred to twice is not a cycle, whatever it holds.
func TestEncodeSharedValueNestedDeeply(t *testing.T) {
	const depth = 1100
	shared := &cycleDetectionHolder{V: (*int)(nil)}
	sharedMap := map[string]any{"a": 1}
	sharedValues := map[string]cycleMapOfValues{"a": {}}
	sharedValuesOfList := map[string]firstFieldMapList{"a": {}}
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

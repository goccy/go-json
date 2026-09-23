package json_test

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

type marshalOfInner struct {
	ID   int
	Tags []string
}

type marshalOfStruct struct {
	Name  string
	Inner marshalOfInner
	Ptr   *marshalOfInner
	Any   any
	Map   map[string]int
	List  []marshalOfInner
	Skip  string `json:"-"`
	Omit  string `json:"omit,omitempty"`
}

// marshalOfPtrReceiver has MarshalJSON on the pointer receiver, which is not called for a value which is not
// addressable, such as the argument of Marshal.
type marshalOfPtrReceiver struct{ V int }

func (v *marshalOfPtrReceiver) MarshalJSON() ([]byte, error) {
	return []byte(`"pointer receiver"`), nil
}

type marshalOfValueReceiver struct{ V int }

func (v marshalOfValueReceiver) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"value receiver %d"`, v.V)), nil
}

type marshalOfHolder struct {
	P marshalOfPtrReceiver
	V marshalOfValueReceiver
}

type marshalOfSinglePointer struct{ P *int }

// marshalOfNested calls MarshalOf for the same type while it is encoded.
type marshalOfNested struct{ Depth int }

func (v marshalOfNested) MarshalJSON() ([]byte, error) {
	if v.Depth == 0 {
		return []byte(`"leaf"`), nil
	}
	b, err := json.MarshalOf(marshalOfNestedHolder{Child: marshalOfNested{Depth: v.Depth - 1}, Pad: "p"})
	if err != nil {
		return nil, err
	}
	return b, nil
}

type marshalOfNestedHolder struct {
	Child marshalOfNested
	Pad   string
}

func assertMarshalOf[T any](t *testing.T, name string, v T, optFuncs ...json.EncodeOptionFunc) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Helper()
		expected, expectedErr := json.MarshalWithOption(v, optFuncs...)
		// twice, so that the value which is reused is used.
		for i := 0; i < 2; i++ {
			got, err := json.MarshalOf(v, optFuncs...)
			if (expectedErr == nil) != (err == nil) {
				t.Fatalf("expected error %v but got %v", expectedErr, err)
			}
			if !bytes.Equal(expected, got) {
				t.Fatalf("expected %s but got %s", expected, got)
			}
		}
	})
}

func TestMarshalOf(t *testing.T) {
	n := 10
	inner := marshalOfInner{ID: 1, Tags: []string{"a", "b"}}
	full := marshalOfStruct{
		Name:  "name",
		Inner: inner,
		Ptr:   &inner,
		Any:   inner,
		Map:   map[string]int{"b": 2, "a": 1},
		List:  []marshalOfInner{inner, {}},
		Skip:  "skip",
	}
	var nilAny any
	var nilStringer fmt.Stringer
	var nilPtr *marshalOfStruct
	var nilMap map[string]int
	var nilSlice []int

	assertMarshalOf(t, "struct", full)
	assertMarshalOf(t, "zero struct", marshalOfStruct{})
	assertMarshalOf(t, "pointer to struct", &full)
	assertMarshalOf(t, "nil pointer", nilPtr)
	assertMarshalOf(t, "pointer to pointer", func() **marshalOfStruct { p := &full; return &p }())
	assertMarshalOf(t, "int", 42)
	assertMarshalOf(t, "float", 1.5)
	assertMarshalOf(t, "string", "<str>")
	assertMarshalOf(t, "bool", true)
	assertMarshalOf(t, "bytes", []byte("bytes"))
	assertMarshalOf(t, "slice", []marshalOfInner{inner, inner})
	assertMarshalOf(t, "nil slice", nilSlice)
	assertMarshalOf(t, "array", [3]int{1, 2, 3})
	assertMarshalOf(t, "empty array", [0]int{})
	assertMarshalOf(t, "array of a pointer", [1]*int{&n})
	assertMarshalOf(t, "large array", [1024]int{1, 2, 3})
	assertMarshalOf(t, "map", map[string]marshalOfInner{"x": inner})
	assertMarshalOf(t, "nil map", nilMap)
	assertMarshalOf(t, "struct of a pointer", marshalOfSinglePointer{P: &n})
	assertMarshalOf(t, "struct of a nil pointer", marshalOfSinglePointer{})
	assertMarshalOf(t, "interface", any(full))
	assertMarshalOf(t, "nil interface", nilAny)
	assertMarshalOf(t, "nil non-empty interface", nilStringer)
	assertMarshalOf(t, "non-empty interface", fmt.Stringer(time.Second))
	assertMarshalOf(t, "pointer receiver marshaler", marshalOfPtrReceiver{V: 1})
	assertMarshalOf(t, "pointer to pointer receiver marshaler", &marshalOfPtrReceiver{V: 1})
	assertMarshalOf(t, "value receiver marshaler", marshalOfValueReceiver{V: 1})
	assertMarshalOf(t, "marshalers in a struct", marshalOfHolder{V: marshalOfValueReceiver{V: 2}})
	assertMarshalOf(t, "marshalers in a slice", []marshalOfHolder{{}, {}})
	assertMarshalOf(t, "marshaler which calls MarshalOf", marshalOfNestedHolder{Child: marshalOfNested{Depth: 5}})
	assertMarshalOf(t, "time", time.Date(2026, 9, 21, 1, 2, 3, 0, time.UTC))
	assertMarshalOf(t, "unsupported type", struct{ C chan int }{})
	assertMarshalOf(t, "unsupported value", struct{ F float64 }{F: nan()})

	assertMarshalOf(t, "option: unordered map", map[string]int{"a": 1}, json.UnorderedMap())
	assertMarshalOf(t, "option: disable html escape", "<str>", json.DisableHTMLEscape())
	assertMarshalOf(t, "option: colorize", full, json.Colorize(json.DefaultColorScheme))
}

func nan() float64 {
	var zero float64
	return zero / zero
}

func TestMarshalOfAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("the runtime context, which keeps the value to reuse, is dropped from its pool at random")
	}
	inner := marshalOfInner{ID: 1, Tags: []string{"a", "b"}}
	v := marshalOfStruct{Name: "name", Inner: inner, Ptr: &inner, List: []marshalOfInner{inner}}
	if _, err := json.MarshalOf(v); err != nil {
		t.Fatal(err)
	}
	// the result is the only allocation.
	if allocs := testing.AllocsPerRun(100, func() { _, _ = json.MarshalOf(v) }); allocs > 1 {
		t.Fatalf("MarshalOf: %v allocations", allocs)
	}
	if allocs := testing.AllocsPerRun(100, func() { _, _ = json.MarshalOf(inner) }); allocs > 1 {
		t.Fatalf("MarshalOf after another type: %v allocations", allocs)
	}
}

func TestMarshalOfConcurrently(t *testing.T) {
	inner := marshalOfInner{ID: 1, Tags: []string{"a", "b"}}
	expectedInner, _ := json.Marshal(inner)
	var wg sync.WaitGroup
	for g := 0; g < 32; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				// the types alternate, so that the value a runtime context keeps is replaced.
				v := marshalOfStruct{Name: fmt.Sprint(g, "-", i), Inner: inner, List: []marshalOfInner{inner}}
				expected, _ := json.Marshal(v)
				got, err := json.MarshalOf(v)
				if err != nil || !bytes.Equal(expected, got) {
					t.Errorf("expected %s but got %s ( %v )", expected, got, err)
					return
				}
				got, err = json.MarshalOf(inner)
				if err != nil || !bytes.Equal(expectedInner, got) {
					t.Errorf("expected %s but got %s ( %v )", expectedInner, got, err)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}

// MarshalOf copies the value to a value which is reused. It must not keep what the value refers to alive.
func TestMarshalOfDoesNotKeepValueAlive(t *testing.T) {
	for _, test := range []struct {
		name   string
		broken bool
	}{
		{name: "encoded"},
		{name: "error", broken: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			collected := make(chan struct{}, 1)
			func() {
				inner := &marshalOfInner{ID: 1}
				runtime.SetFinalizer(inner, func(*marshalOfInner) { collected <- struct{}{} })
				v := marshalOfStruct{Ptr: inner}
				if test.broken {
					v.Any = make(chan int)
				}
				if _, err := json.MarshalOf(v); (err != nil) != test.broken {
					t.Fatalf("unexpected error: %v", err)
				}
			}()
			// Only one GC: the pool releases what it has after a few, whatever it holds.
			runtime.GC()
			select {
			case <-collected:
			case <-time.After(10 * time.Second):
				t.Fatal("the value is kept alive after the encoding")
			}
		})
	}
}

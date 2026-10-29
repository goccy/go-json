//go:build go1.24

package json_test

import (
	"context"
	"io"
	"runtime"
	"strings"
	"testing"
	"unsafe"
	"weak"

	json "github.com/goccy/go-json"
)

// The decoders keep buffers and contexts in pools between the calls. Nothing in them may refer to
// what a call decoded: a decoded string refers to the copy of the input, so a pointer left behind
// keeps the whole input of a previous call alive, and a program which decodes large inputs at a high
// rate keeps as many of them alive as its pools hold objects.

type retentionNode struct {
	Name string
	Kids []*retentionNode
}

func retentionInput() []byte {
	var b strings.Builder
	b.WriteString(`{"Name":"root","Kids":[`)
	for i := 0; i < 64; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"Name":"` + strings.Repeat("n", 64) + `","Kids":[{"Name":"leaf","Kids":[]},{"Name":"leaf","Kids":[]}]}`)
	}
	b.WriteString(`]}`)
	return []byte(b.String())
}

func decodedIsCollected(t *testing.T, decode func() any) {
	t.Helper()
	var refs []weak.Pointer[retentionNode]
	for i := 0; i < 8; i++ {
		switch v := decode().(type) {
		case *retentionNode:
			refs = append(refs, weak.Make(v), weak.Make(v.Kids[0]), weak.Make(v.Kids[63].Kids[1]))
		case map[string]any:
			// the values of interface{} are checked through the node decoded beside them
			refs = append(refs, weak.Make(v["node"].(*retentionNode)))
		}
	}
	// A pool keeps its objects over one collection ( its victim cache ): what they refer to is alive
	// through that collection, so one collection tells whether a pooled object refers to a decoded value.
	runtime.GC()
	for i, r := range refs {
		if r.Value() != nil {
			t.Fatalf("a value decoded by a previous call is still alive ( %d )", i)
		}
	}
}

func TestDecodeKeepsNothingOfPreviousCall(t *testing.T) {
	input := retentionInput()
	t.Run("Unmarshal", func(t *testing.T) {
		decodedIsCollected(t, func() any {
			v := new(retentionNode)
			if err := json.Unmarshal(input, v); err != nil {
				t.Fatal(err)
			}
			return v
		})
	})
	t.Run("Decoder", func(t *testing.T) {
		decodedIsCollected(t, func() any {
			v := new(retentionNode)
			if err := json.NewDecoder(strings.NewReader(string(input))).Decode(v); err != nil {
				t.Fatal(err)
			}
			return v
		})
	})
}

// A decoded value refers to memory of its own only, as with encoding/json: a value kept out of a call keeps alive
// neither the other values of the call, nor the buffers of the decoder, nor what the caller gave to the call.

// retentionString is a string of c whose bytes are not in a block of the tiny allocator of the runtime, which packs
// objects of less than 16 bytes without pointers together: such a string keeps its neighbors alive, whatever
// allocated it.
func retentionString(c byte) string {
	return strings.Repeat(string(c), 32)
}

// aliveRef reports whether the object which it was made for is alive, without keeping it alive.
type aliveRef func() bool

// refOf returns the aliveRef of the object which p points into.
func refOf(p *byte) aliveRef {
	w := weak.Make(p)
	return func() bool { return w.Value() != nil }
}

// refOfString returns the aliveRef of the bytes of s.
func refOfString(s string) aliveRef {
	return refOf(unsafe.StringData(s))
}

// refOfBox returns the aliveRef of the memory which the interface value v holds its value in.
func refOfBox(v any) aliveRef {
	return refOf((*byte)((*[2]unsafe.Pointer)(unsafe.Pointer(&v))[1]))
}

// allCollected fails the test if an object of refs is alive after a collection: the pools keep their objects over
// one collection ( their victim caches ), so that a pooled object which refers to one of them keeps it alive.
func allCollected(t *testing.T, refs []aliveRef) {
	t.Helper()
	runtime.GC()
	for i, alive := range refs {
		if alive() {
			t.Fatalf("an object which the kept value is not made of is alive ( %d of %d )", i, len(refs))
		}
	}
}

// recordingReader is a reader which keeps an aliveRef of every buffer it reads into.
type recordingReader struct {
	r    io.Reader
	bufs []aliveRef
}

func (r *recordingReader) Read(p []byte) (int, error) {
	if len(p) != 0 {
		r.bufs = append(r.bufs, refOf(&p[0]))
	}
	return r.r.Read(p)
}

type retentionStrings struct {
	A, B, C string
	Pad     string
}

func TestDecodedValueKeepsOnlyItsOwnMemory(t *testing.T) {
	a, b := retentionString('a'), retentionString('b')
	escaped := strings.Repeat(`c\n`, 16)
	object := `{"A":"` + a + `","B":"` + b + `","C":"` + escaped + `"}`
	// an object which the read buffer of a Decoder doesn't hold whole, which is decoded element by element, with a
	// token longer than the buffer
	long := `{"A":"` + a + `","B":"` + b + `","C":"` + escaped + `","Pad":"` + strings.Repeat("p", 2000) + `"}`
	decodeBy := func(t *testing.T, input string) (*retentionStrings, []aliveRef) {
		v := new(retentionStrings)
		r := &recordingReader{r: strings.NewReader(input)}
		if err := json.NewDecoder(r).Decode(v); err != nil {
			t.Fatal(err)
		}
		// the buffers of the decoder, which is dropped
		return v, r.bufs
	}
	for _, tc := range []struct {
		name   string
		decode func(*testing.T) (*retentionStrings, []aliveRef)
	}{
		{"Unmarshal", func(t *testing.T) (*retentionStrings, []aliveRef) {
			v := new(retentionStrings)
			if err := json.Unmarshal([]byte(object), v); err != nil {
				t.Fatal(err)
			}
			return v, nil
		}},
		{"UnmarshalOf", func(t *testing.T) (*retentionStrings, []aliveRef) {
			v := new(retentionStrings)
			if err := json.UnmarshalOf([]byte(object), v); err != nil {
				t.Fatal(err)
			}
			return v, nil
		}},
		{"UnmarshalContext", func(t *testing.T) (*retentionStrings, []aliveRef) {
			v := new(retentionStrings)
			if err := json.UnmarshalContext(context.Background(), []byte(object), v); err != nil {
				t.Fatal(err)
			}
			return v, nil
		}},
		{"Decoder", func(t *testing.T) (*retentionStrings, []aliveRef) {
			return decodeBy(t, object)
		}},
		{"Decoder element by element", func(t *testing.T) (*retentionStrings, []aliveRef) {
			return decodeBy(t, long)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, refs := tc.decode(t)
			kept := v.B
			refs = append(refs, refOfString(v.A), refOfString(v.C))
			if v.Pad != "" {
				refs = append(refs, refOfString(v.Pad))
			}
			v = nil
			allCollected(t, refs)
			if kept != b {
				t.Fatalf("got %q", kept)
			}
		})
	}
}

func TestDecodedInterfaceValueKeepsOnlyItsOwnMemory(t *testing.T) {
	a, b := retentionString('a'), retentionString('b')
	digits := strings.Repeat("1", 32)
	unmarshal := func(input string, v *any) error {
		return json.Unmarshal([]byte(input), v)
	}
	decodeNumbers := func(input string, v *any) error {
		dec := json.NewDecoder(strings.NewReader(input))
		dec.UseNumber()
		return dec.Decode(v)
	}
	for _, tc := range []struct {
		name   string
		input  string
		decode func(string, *any) error
		// keep returns what is kept of the two elements of the array, and the memory which it must not keep alive
		keep func(elems []any) (any, []aliveRef)
		want any
	}{
		{"strings", `["` + a + `","` + b + `"]`, unmarshal, func(elems []any) (any, []aliveRef) {
			return elems[1].(string), []aliveRef{refOfString(elems[0].(string)), refOfBox(elems[0]), refOfBox(elems[1])}
		}, b},
		{"json.Number", `[` + digits + `,` + digits + `]`, decodeNumbers, func(elems []any) (any, []aliveRef) {
			first := string(elems[0].(json.Number))
			return elems[1].(json.Number), []aliveRef{refOfString(first), refOfBox(elems[0]), refOfBox(elems[1])}
		}, json.Number(digits)},
		{"keys", `[{"` + a + `":1},{"` + b + `":2}]`, unmarshal, func(elems []any) (any, []aliveRef) {
			var first, second string
			for k := range elems[0].(map[string]any) {
				first = k
			}
			for k := range elems[1].(map[string]any) {
				second = k
			}
			return second, []aliveRef{refOfString(first)}
		}, b},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var v any
			if err := tc.decode(tc.input, &v); err != nil {
				t.Fatal(err)
			}
			kept, refs := tc.keep(v.([]any))
			v = nil
			allCollected(t, refs)
			if kept != tc.want {
				t.Fatalf("got %v", kept)
			}
		})
	}
}

func TestExtractedValueKeepsOnlyItsOwnMemory(t *testing.T) {
	a, b := retentionString('a'), retentionString('b')
	data := []byte(`["` + a + `","` + b + `"]`)
	path, err := json.CreatePath("$[*]")
	if err != nil {
		t.Fatal(err)
	}
	values, err := path.Extract(data)
	if err != nil {
		t.Fatal(err)
	}
	kept := values[1]
	refs := []aliveRef{refOf(&values[0][0])}
	values = nil
	allCollected(t, refs)
	if string(kept) != `"`+b+`"` {
		t.Fatalf("got %s", kept)
	}
	// the values are copies: the input may be modified after the call, as the one of the root alone is
	root, err := json.CreatePath("$")
	if err != nil {
		t.Fatal(err)
	}
	all, err := root.Extract(data)
	if err != nil {
		t.Fatal(err)
	}
	clear(data)
	if string(kept) != `"`+b+`"` || string(all[0]) != `["`+a+`","`+b+`"]` {
		t.Fatalf("the values changed with the input: %s, %s", kept, all[0])
	}
}

type retentionContextKey struct{}

// contextOfLargeValue returns a context which holds a value of its own, and the aliveRef of that value.
func contextOfLargeValue() (context.Context, aliveRef) {
	held := new([1 << 16]byte)
	return context.WithValue(context.Background(), retentionContextKey{}, held), refOf(&held[0])
}

func TestDecodeKeepsNothingOfTheCaller(t *testing.T) {
	ctx, held := contextOfLargeValue()
	var v retentionStrings
	if err := json.UnmarshalContext(ctx, []byte(`{"A":"a"}`), &v); err != nil {
		t.Fatal(err)
	}
	ctx = nil
	allCollected(t, []aliveRef{held})
}

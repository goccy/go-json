package json_test

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	jsonv2 "github.com/goccy/go-json/v2"
)

// The memory which the values decoded by go-json keep alive is compared with the memory which the same values
// decoded by encoding/json keep alive, in the same process, for the ways of decoding which kept far more than
// encoding/json before issue 775: short strings copied into chunks shared by every call, the numbers and the strings
// of interface{} boxed in slabs, slices and maps made for the length which other calls saw, the strings of a Decoder
// referring to its read buffer, the fragments of Path.Extract referring to a copy of the whole input.
//
// A value of go-json may keep a little more than one of encoding/json: the two decoders grow a slice from other
// capacities, place the small objects of a value in other blocks of the allocator, and the encoding/json of Go 1.27
// shares equal strings between its calls ( the stringCache of encoding/json/v2 ). What a value must never keep is
// memory which other values, other calls or the buffers of the decoder take, which grows with the work a program
// does: the values kept 9 to 14000 times the bytes of encoding/json's then. A case fails when a value of go-json
// keeps more than footprintFactor times the bytes of one of encoding/json, plus footprintSlack bytes: a margin for
// the differences of layout, which leaves retention of that kind well outside.
const (
	footprintFactor = 2
	footprintSlack  = 64
	// footprintKept is the number of the values which a case keeps, so that the bytes which a collection leaves to
	// the runtime, which are not the values', are a small part of the bytes per value.
	footprintKept = 1000
)

// liveHeap returns the live heap after two collections: the second one empties the pools, which keep their objects
// over one collection, so that only what the program refers to counts.
func liveHeap() int64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.HeapAlloc)
}

// keptPerValue returns the bytes of the live heap per value which the footprintKept values that decode returns keep
// alive: the live heap with the values, minus the live heap once they are dropped.
func keptPerValue(decode func(i int) any) float64 {
	kept := make([]any, footprintKept)
	for i := range kept {
		kept[i] = decode(i)
	}
	with := liveHeap()
	clear(kept)
	without := liveHeap()
	runtime.KeepAlive(kept)
	return float64(with-without) / footprintKept
}

// compareFootprint fails the test if a value which goDecode returns keeps alive more than the margin over what a
// value which stdDecode returns does ( see footprintFactor ).
func compareFootprint(t *testing.T, goDecode, stdDecode func(i int) any) {
	t.Helper()
	std := keptPerValue(stdDecode)
	got := keptPerValue(goDecode)
	t.Logf("a value keeps %.0f bytes alive, %.0f with encoding/json", got, std)
	if got > footprintFactor*std+footprintSlack {
		t.Errorf("a value keeps %.0f bytes alive, %.0f with encoding/json", got, std)
	}
}

// footprintItems returns {"items":[...]} of n objects of two strings of 12 to 14 bytes each, which differ by salt:
// 300 items are about 15 KB.
func footprintItems(n, salt int) []byte {
	var b strings.Builder
	b.WriteString(`{"items":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"A":"some text %d","B":"more text %d"}`, i+salt, i)
	}
	b.WriteString(`]}`)
	return []byte(b.String())
}

// footprintArray returns [...] of n elements, the element of each index i made by elem.
func footprintArray(n int, elem func(i int) string) []byte {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(elem(i))
	}
	b.WriteByte(']')
	return []byte(b.String())
}

// footprintObject returns an object of n members of numbers.
func footprintObject(n int) []byte {
	var b strings.Builder
	b.WriteByte('{')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `"key%d":%d`, i, i)
	}
	b.WriteByte('}')
	return []byte(b.String())
}

// footprintSmall returns a small object, which differs by i.
func footprintSmall(i int) []byte {
	return []byte(fmt.Sprintf(`{"id":"%d","city":"Sao Paulo"}`, i))
}

// footprintStream returns a stream of a string of long bytes, if long is not 0, and then n small objects.
func footprintStream(long, n int) []byte {
	var b bytes.Buffer
	if long != 0 {
		b.WriteString(`"` + strings.Repeat("x", long) + `"` + "\n")
	}
	for i := 0; i < n; i++ {
		b.Write(footprintSmall(i))
		b.WriteByte('\n')
	}
	return b.Bytes()
}

type footprintItem struct {
	A string `json:"A"`
	B string `json:"B"`
}

type footprintLarge struct {
	Items []footprintItem `json:"items"`
}

type footprintSmallValue struct {
	ID   string `json:"id"`
	City string `json:"city"`
}

// footprintCases are the ways of decoding which a value is kept out of, made for an unmarshal function.
func footprintCases(t *testing.T) []struct {
	name   string
	decode func(unmarshal func([]byte, any) error) func(int) any
} {
	large := footprintItems(300, 0)
	numbers := footprintArray(300, func(i int) string { return fmt.Sprintf("%d.5", i) })
	slice := footprintArray(100, func(i int) string { return fmt.Sprintf("%d.5", i) })
	object := footprintObject(100)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	return []struct {
		name   string
		decode func(unmarshal func([]byte, any) error) func(int) any
	}{
		{"a small value after a large document", func(u func([]byte, any) error) func(int) any {
			return func(i int) any {
				var l footprintLarge
				var s footprintSmallValue
				must(u(large, &l))
				must(u(footprintSmall(i), &s))
				return s
			}
		}},
		{"a string of each large document", func(u func([]byte, any) error) func(int) any {
			return func(i int) any {
				var l footprintLarge
				must(u(footprintItems(300, i*300), &l))
				return l.Items[(i*37)%len(l.Items)].A
			}
		}},
		{"a string of each large array decoded into interface{}", func(u func([]byte, any) error) func(int) any {
			return func(i int) any {
				var v any
				must(u(footprintArray(300, func(j int) string { return fmt.Sprintf(`"string number %d"`, i*300+j) }), &v))
				return v.([]any)[(i*37)%300]
			}
		}},
		{"a number of each large array decoded into interface{}", func(u func([]byte, any) error) func(int) any {
			return func(i int) any {
				var v any
				must(u(numbers, &v))
				return v.([]any)[(i*37)%300]
			}
		}},
		{"a small value decoded into interface{} after a large one", func(u func([]byte, any) error) func(int) any {
			return func(i int) any {
				var l, v any
				must(u(large, &l))
				must(u(footprintSmall(i), &v))
				return v
			}
		}},
		{"a slice of 1 element after one of 100", func(u func([]byte, any) error) func(int) any {
			return func(int) any {
				var big, s []float64
				must(u(slice, &big))
				must(u([]byte(`[1.5]`), &s))
				return s
			}
		}},
		{"a map of 1 member after one of 100", func(u func([]byte, any) error) func(int) any {
			return func(int) any {
				var big, m map[string]int
				must(u(object, &big))
				must(u([]byte(`{"a":1}`), &m))
				return m
			}
		}},
	}
}

func TestDecodedFootprintAsEncodingJSON(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own")
	}
	for _, u := range []struct {
		name      string
		unmarshal func([]byte, any) error
	}{
		{"Unmarshal", json.Unmarshal},
		{"UnmarshalContext", func(data []byte, v any) error {
			return json.UnmarshalContext(context.Background(), data, v)
		}},
		{"UnmarshalNoEscape", func(data []byte, v any) error {
			return json.UnmarshalNoEscape(data, v)
		}},
		{"v2.Unmarshal", func(data []byte, v any) error {
			return jsonv2.Unmarshal(data, v)
		}},
		{"v2.UnmarshalRead", func(data []byte, v any) error {
			return jsonv2.UnmarshalRead(bytes.NewReader(data), v)
		}},
		{"v2.UnmarshalDecode", func(data []byte, v any) error {
			return jsonv2.UnmarshalDecode(jsontext.NewDecoder(bytes.NewReader(data)), v)
		}},
		{"Decoder", func(data []byte, v any) error {
			return json.NewDecoder(bytes.NewReader(data)).Decode(v)
		}},
	} {
		for _, tc := range footprintCases(t) {
			t.Run(u.name+"/"+tc.name, func(t *testing.T) {
				compareFootprint(t, tc.decode(u.unmarshal), tc.decode(stdjson.Unmarshal))
			})
		}
	}
}

func TestUnmarshalOfFootprintAsEncodingJSON(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own")
	}
	large := footprintItems(300, 0)
	t.Run("a small value after a large document", func(t *testing.T) {
		compareFootprint(t, func(i int) any {
			var l footprintLarge
			var s footprintSmallValue
			if err := json.UnmarshalOf(large, &l); err != nil {
				t.Fatal(err)
			}
			if err := json.UnmarshalOf(footprintSmall(i), &s); err != nil {
				t.Fatal(err)
			}
			return s
		}, func(i int) any {
			var l footprintLarge
			var s footprintSmallValue
			if err := stdjson.Unmarshal(large, &l); err != nil {
				t.Fatal(err)
			}
			if err := stdjson.Unmarshal(footprintSmall(i), &s); err != nil {
				t.Fatal(err)
			}
			return s
		})
	})
}

// footprintDecoder is the Decoder of encoding/json or of go-json.
type footprintDecoder interface {
	Decode(v any) error
}

// keptOfStream returns the function of keptPerValue which keeps one value in every ten of the stream of dec.
func keptOfStream(t *testing.T, dec footprintDecoder) func(int) any {
	return func(int) any {
		var v footprintSmallValue
		for j := 0; j < 10; j++ {
			v = footprintSmallValue{}
			if err := dec.Decode(&v); err != nil {
				t.Fatal(err)
			}
		}
		return v
	}
}

// The values decoded by a Decoder one after another, which share its read buffer, which a long string grew.
func TestDecoderFootprintAsEncodingJSON(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own")
	}
	for _, long := range []int{0, 1 << 20} {
		t.Run(fmt.Sprintf("after a string of %d bytes", long), func(t *testing.T) {
			stream := footprintStream(long, 10*footprintKept)
			newDecoders := func(newDecoder func([]byte) footprintDecoder) func(int) any {
				dec := newDecoder(stream)
				if long != 0 {
					var s string
					if err := dec.Decode(&s); err != nil {
						t.Fatal(err)
					}
				}
				return keptOfStream(t, dec)
			}
			compareFootprint(t, newDecoders(func(b []byte) footprintDecoder {
				return json.NewDecoder(bytes.NewReader(b))
			}), newDecoders(func(b []byte) footprintDecoder {
				return stdjson.NewDecoder(bytes.NewReader(b))
			}))
		})
	}
}

// encoding/json has no Path: a fragment is compared with a copy of its bytes, which is all it needs.
func TestExtractFootprint(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own")
	}
	path, err := json.CreatePath("$.items[0].A")
	if err != nil {
		t.Fatal(err)
	}
	compareFootprint(t, func(i int) any {
		fragments, err := path.Extract(footprintItems(300, i*300))
		if err != nil {
			t.Fatal(err)
		}
		return fragments[0]
	}, func(i int) any {
		return []byte(fmt.Sprintf(`"some text %d"`, i*300))
	})
}

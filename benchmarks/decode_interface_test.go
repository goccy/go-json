package benchmark

import (
	"testing"

	gojson "github.com/goccy/go-json"
)

// The values of interface types which are set before the decoding: a value of interface{} which holds a pointer
// is decoded into what it points to, and a value of an interface type with methods is decoded by its
// unmarshaler.

type ifaceItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ifacePointerPayload struct {
	A, B, C, D, E, F, G, H any
}

func newIfacePointerPayload() *ifacePointerPayload {
	return &ifacePointerPayload{
		A: &ifaceItem{}, B: &ifaceItem{}, C: &ifaceItem{}, D: &ifaceItem{},
		E: &ifaceItem{}, F: &ifaceItem{}, G: &ifaceItem{}, H: &ifaceItem{},
	}
}

// ifaceUnmarshaler is an interface type with methods, whose values are decoded by their unmarshaler.
type ifaceUnmarshaler interface {
	UnmarshalJSON([]byte) error
}

// ifaceCount is decoded into the length of its value.
type ifaceCount int

func (c *ifaceCount) UnmarshalJSON(b []byte) error {
	*c = ifaceCount(len(b))
	return nil
}

type ifaceUnmarshalerPayload struct {
	A, B, C, D, E, F, G, H ifaceUnmarshaler
}

func newIfaceUnmarshalerPayload() *ifaceUnmarshalerPayload {
	return &ifaceUnmarshalerPayload{
		A: new(ifaceCount), B: new(ifaceCount), C: new(ifaceCount), D: new(ifaceCount),
		E: new(ifaceCount), F: new(ifaceCount), G: new(ifaceCount), H: new(ifaceCount),
	}
}

func ifaceDocument(value string) []byte {
	b := []byte{'{'}
	for i, name := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, `"`+name+`":`+value...)
	}
	return append(b, '}')
}

func Benchmark_Decode_InterfacePointer_Unmarshal_GoJson(b *testing.B) {
	data := ifaceDocument(`{"id":1,"name":"item"}`)
	v := newIfacePointerPayload()
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := gojson.Unmarshal(data, v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_InterfaceUnmarshaler_Unmarshal_GoJson(b *testing.B) {
	data := ifaceDocument(`[1,2,3]`)
	v := newIfaceUnmarshalerPayload()
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := gojson.Unmarshal(data, v); err != nil {
			b.Fatal(err)
		}
	}
}

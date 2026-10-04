//go:build go1.25

package benchmark

import (
	"strconv"
	"testing"

	gojson "github.com/goccy/go-json"
	gojsontext "github.com/goccy/go-json/jsontext"
	gojsonv2 "github.com/goccy/go-json/v2"
)

// The methods of the types of a program, which write their values themselves: MarshalJSON, which v1 calls, and
// MarshalJSONTo, which github.com/goccy/go-json/v2 calls instead, for the same values. MarshalJSON returns a new
// slice, which the encoder checks and copies; MarshalJSONTo writes to the encoder, which checks what it writes.
// The types of the *JSON names have MarshalJSON only, which v2 calls when a type has no MarshalJSONTo.

// orderID is a string of a prefix and a number: a value written by its bytes.
type orderID int64

func (id orderID) MarshalJSON() ([]byte, error) {
	b := append([]byte(nil), `"order-`...)
	b = strconv.AppendInt(b, int64(id), 10)
	return append(b, '"'), nil
}

func (id orderID) MarshalJSONTo(enc *gojsontext.Encoder) error {
	b := append(enc.AvailableBuffer(), `"order-`...)
	b = strconv.AppendInt(b, int64(id), 10)
	return enc.WriteValue(append(b, '"'))
}

// point is an object of two numbers: a value written by its tokens.
type point struct{ X, Y float64 }

func (p point) MarshalJSON() ([]byte, error) {
	b := append([]byte(nil), `{"x":`...)
	b = strconv.AppendFloat(b, p.X, 'g', -1, 64)
	b = append(b, `,"y":`...)
	b = strconv.AppendFloat(b, p.Y, 'g', -1, 64)
	return append(b, '}'), nil
}

func (p point) MarshalJSONTo(enc *gojsontext.Encoder) error {
	if err := enc.WriteToken(gojsontext.BeginObject); err != nil {
		return err
	}
	if err := enc.WriteToken(gojsontext.String("x")); err != nil {
		return err
	}
	if err := enc.WriteToken(gojsontext.Float(p.X)); err != nil {
		return err
	}
	if err := enc.WriteToken(gojsontext.String("y")); err != nil {
		return err
	}
	if err := enc.WriteToken(gojsontext.Float(p.Y)); err != nil {
		return err
	}
	return enc.WriteToken(gojsontext.EndObject)
}

type orderIDJSON orderID

func (id orderIDJSON) MarshalJSON() ([]byte, error) { return orderID(id).MarshalJSON() }

type pointJSON point

func (p pointJSON) MarshalJSON() ([]byte, error) { return point(p).MarshalJSON() }

// order has the values of the types, among plain fields.
type order[ID, Point any] struct {
	ID       ID      `json:"id"`
	Customer string  `json:"customer"`
	From     Point   `json:"from"`
	To       Point   `json:"to"`
	Related  []ID    `json:"related"`
	Total    float64 `json:"total"`
}

func newOrders[ID ~int64, Point ~struct{ X, Y float64 }]() []order[ID, Point] {
	orders := make([]order[ID, Point], 100)
	for i := range orders {
		orders[i] = order[ID, Point]{
			ID: ID(1000 + i), Customer: "customer " + strconv.Itoa(i),
			From: Point{X: 35.68 + float64(i)/100, Y: 139.76}, To: Point{X: 34.69, Y: 135.5 + float64(i)/100},
			Related: []ID{ID(i), ID(i + 1)}, Total: 1234.5 + float64(i),
		}
	}
	return orders
}

func Benchmark_Encode_MarshalerMethods_GoJson(b *testing.B) {
	orders := newOrders[orderID, point]()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.MarshalWithOption(orders, gojson.DisableHTMLEscape(), gojson.UnorderedMap()); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MarshalerMethods_GoJsonV2(b *testing.B) {
	orders := newOrders[orderID, point]()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojsonv2.Marshal(orders); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_MarshalerMethods_GoJsonV2MarshalJSON(b *testing.B) {
	orders := newOrders[orderIDJSON, pointJSON]()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojsonv2.Marshal(orders); err != nil {
			b.Fatal(err)
		}
	}
}

// The methods write the same JSON.
func TestMarshalerMethods(t *testing.T) {
	v1, err := gojson.MarshalWithOption(newOrders[orderID, point](), gojson.DisableHTMLEscape(), gojson.UnorderedMap())
	if err != nil {
		t.Fatal(err)
	}
	v2, err := gojsonv2.Marshal(newOrders[orderID, point]())
	if err != nil {
		t.Fatal(err)
	}
	v2JSON, err := gojsonv2.Marshal(newOrders[orderIDJSON, pointJSON]())
	if err != nil {
		t.Fatal(err)
	}
	if string(v1) != string(v2) || string(v1) != string(v2JSON) {
		t.Fatalf("outputs differ:\nv1         %.200s\nv2         %.200s\nMarshalJSON %.200s", v1, v2, v2JSON)
	}
}

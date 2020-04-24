package json_test

import (
	"testing"

	gojay "github.com/francoispqt/gojay"
	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

type T struct {
	A int     `json:"a"`
	B float64 `json:"b"`
	C string  `json:"c"`
}

func (t *T) MarshalJSONObject(enc *gojay.Encoder) {
	enc.IntKey("a", t.A)
	enc.FloatKey("b", t.B)
	enc.StringKey("c", t.C)
}

func (t *T) IsNil() bool {
	return t == nil
}

func newT() *T {
	return &T{A: 1, B: 3.14, C: `hello"world`}
}

func Benchmark_Encode_jsoniter(b *testing.B) {
	v := newT()
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_gojay(b *testing.B) {
	v := newT()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojay.MarshalJSONObject(v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Encode_gojson(b *testing.B) {
	v := newT()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := gojson.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

var fixture = []byte(`{"st": 1,"sid": 486,"tt": "active","gr": 0,"uuid": "de305d54-75b4-431b-adb2-eb6b9e546014","ip": "127.0.0.1","ua": "user_agent","tz": -6,"v": 1}`)

type SmallPayload struct {
	St   int
	Sid  int
	Tt   string
	Gr   int
	Uuid string
	Ip   string
	Ua   string
	Tz   int
	V    int
}

func (t *SmallPayload) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "st":
		return dec.AddInt(&t.St)
	case "sid":
		return dec.AddInt(&t.Sid)
	case "gr":
		return dec.AddInt(&t.Gr)
	case "tz":
		return dec.AddInt(&t.Tz)
	case "v":
		return dec.AddInt(&t.V)
	case "tt":
		return dec.AddString(&t.Tt)
	case "uuid":
		return dec.AddString(&t.Uuid)
	case "ip":
		return dec.AddString(&t.Ip)
	case "ua":
		return dec.AddString(&t.Ua)
	}
	return nil
}

func (t *SmallPayload) NKeys() int {
	return 9
}

func Benchmark_Decode_jsoniter(b *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v SmallPayload
		if err := json.Unmarshal(fixture, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_gojay(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v SmallPayload
		if err := gojay.UnmarshalJSONObject(fixture, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_gojson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v SmallPayload
		if err := gojson.Unmarshal(fixture, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_gojson_noescape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v SmallPayload
		if err := gojson.UnmarshalNoEscape(fixture, &v); err != nil {
			b.Fatal(err)
		}
	}
}

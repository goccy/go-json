package benchmark

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	gojson "github.com/goccy/go-json"
)

type coordinate struct {
	X    float64                `json:"x"`
	Y    float64                `json:"y"`
	Z    float64                `json:"z"`
	Name string                 `json:"name"`
	Opts map[string]interface{} `json:"opts"`
}

type testStruct struct {
	Coordinates []coordinate `json:"coordinates"`
}

func Benchmark_DecodeHugeDataGoJSON(b *testing.B) {
	dat, err := ioutil.ReadFile("./bench.json")
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	jobj := testStruct{}
	if err := gojson.Unmarshal(dat, &jobj); err != nil {
		b.Fatal(err)
	}
}

func Benchmark_DecodeHugeDataEncodingJSON(b *testing.B) {
	dat, err := ioutil.ReadFile("./bench.json")
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	jobj := testStruct{}
	if err := json.Unmarshal(dat, &jobj); err != nil {
		b.Fatal(err)
	}
}

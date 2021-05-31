package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

func TestColorize(t *testing.T) {
	v := struct {
		A int
		B uint
		C float32
		D string
		E bool
		F []byte
		G []int
	}{
		A: 123,
		B: 456,
		C: 3.14,
		D: "hello",
		E: true,
		F: []byte("binary"),
		G: []int{1, 2, 3, 4},
	}
	b, err := json.MarshalWithOption(v, json.Colorize(json.DefaultColorScheme))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}

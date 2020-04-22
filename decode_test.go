package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

func Test_Decoder(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		var v struct {
			A int    `json:"abcd"`
			B string `json:"str"`
		}
		assertErr(t, json.Unmarshal([]byte(`{ "abcd" : 123 , "str" : "hello" }`), &v))
		assertEq(t, "struct.A", 123, v.A)
		assertEq(t, "struct.B", "hello", v.B)
	})
}

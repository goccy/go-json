package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

func Test_Decoder(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		type T struct {
			AA int    `json:"aa"`
			BB string `json:"bb"`
			CC bool   `json:"cc"`
		}
		var v struct {
			A int    `json:"abcd"`
			B string `json:"str"`
			C bool
			D *T
		}
		content := []byte(`
{
  "abcd": 123,
  "str" : "hello",
  "c"   : true,
  "d"   : {
    "aa": 2,
    "bb": "world",
    "cc": true
  }
}`)
		assertErr(t, json.Unmarshal(content, &v))
		assertEq(t, "struct.A", 123, v.A)
		assertEq(t, "struct.B", "hello", v.B)
		assertEq(t, "struct.C", true, v.C)
		assertEq(t, "struct.D.AA", 2, v.D.AA)
	})
}

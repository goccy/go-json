package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

func Test_Decoder(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		var v int
		assertErr(t, json.Unmarshal([]byte(`-1`), &v))
		assertEq(t, "int", int(-1), v)
	})
	t.Run("int8", func(t *testing.T) {
		var v int8
		assertErr(t, json.Unmarshal([]byte(`-2`), &v))
		assertEq(t, "int8", int8(-2), v)
	})
	t.Run("int16", func(t *testing.T) {
		var v int16
		assertErr(t, json.Unmarshal([]byte(`-3`), &v))
		assertEq(t, "int16", int16(-3), v)
	})
	t.Run("int32", func(t *testing.T) {
		var v int32
		assertErr(t, json.Unmarshal([]byte(`-4`), &v))
		assertEq(t, "int32", int32(-4), v)
	})
	t.Run("int64", func(t *testing.T) {
		var v int64
		assertErr(t, json.Unmarshal([]byte(`-5`), &v))
		assertEq(t, "int64", int64(-5), v)
	})
	t.Run("uint", func(t *testing.T) {
		var v uint
		assertErr(t, json.Unmarshal([]byte(`1`), &v))
		assertEq(t, "uint", uint(1), v)
	})
	t.Run("uint8", func(t *testing.T) {
		var v uint8
		assertErr(t, json.Unmarshal([]byte(`2`), &v))
		assertEq(t, "uint8", uint8(2), v)
	})
	t.Run("uint16", func(t *testing.T) {
		var v uint16
		assertErr(t, json.Unmarshal([]byte(`3`), &v))
		assertEq(t, "uint16", uint16(3), v)
	})
	t.Run("uint32", func(t *testing.T) {
		var v uint32
		assertErr(t, json.Unmarshal([]byte(`4`), &v))
		assertEq(t, "uint32", uint32(4), v)
	})
	t.Run("uint64", func(t *testing.T) {
		var v uint64
		assertErr(t, json.Unmarshal([]byte(`5`), &v))
		assertEq(t, "uint64", uint64(5), v)
	})
	t.Run("bool", func(t *testing.T) {
		t.Run("true", func(t *testing.T) {
			var v bool
			assertErr(t, json.Unmarshal([]byte(`true`), &v))
			assertEq(t, "bool", true, v)
		})
		t.Run("false", func(t *testing.T) {
			v := true
			assertErr(t, json.Unmarshal([]byte(`false`), &v))
			assertEq(t, "bool", false, v)
		})
	})
	t.Run("string", func(t *testing.T) {
		var v string
		assertErr(t, json.Unmarshal([]byte(`"hello"`), &v))
		assertEq(t, "string", "hello", v)
	})
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
		assertEq(t, "struct.D.BB", "world", v.D.BB)
		assertEq(t, "struct.D.CC", true, v.D.CC)
	})
}

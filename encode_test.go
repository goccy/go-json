package json_test

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
)

func Test_Marshal(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		bytes, err := json.Marshal(-10)
		assertErr(t, err)
		assertEq(t, "int", `-10`, string(bytes))
	})
	t.Run("int8", func(t *testing.T) {
		bytes, err := json.Marshal(int8(-11))
		assertErr(t, err)
		assertEq(t, "int8", `-11`, string(bytes))
	})
	t.Run("int16", func(t *testing.T) {
		bytes, err := json.Marshal(int16(-12))
		assertErr(t, err)
		assertEq(t, "int16", `-12`, string(bytes))
	})
	t.Run("int32", func(t *testing.T) {
		bytes, err := json.Marshal(int32(-13))
		assertErr(t, err)
		assertEq(t, "int32", `-13`, string(bytes))
	})
	t.Run("int64", func(t *testing.T) {
		bytes, err := json.Marshal(int64(-14))
		assertErr(t, err)
		assertEq(t, "int64", `-14`, string(bytes))
	})
	t.Run("uint", func(t *testing.T) {
		bytes, err := json.Marshal(uint(10))
		assertErr(t, err)
		assertEq(t, "uint", `10`, string(bytes))
	})
	t.Run("uint8", func(t *testing.T) {
		bytes, err := json.Marshal(uint8(11))
		assertErr(t, err)
		assertEq(t, "uint8", `11`, string(bytes))
	})
	t.Run("uint16", func(t *testing.T) {
		bytes, err := json.Marshal(uint16(12))
		assertErr(t, err)
		assertEq(t, "uint16", `12`, string(bytes))
	})
	t.Run("uint32", func(t *testing.T) {
		bytes, err := json.Marshal(uint32(13))
		assertErr(t, err)
		assertEq(t, "uint32", `13`, string(bytes))
	})
	t.Run("uint64", func(t *testing.T) {
		bytes, err := json.Marshal(uint64(14))
		assertErr(t, err)
		assertEq(t, "uint64", `14`, string(bytes))
	})
	t.Run("float32", func(t *testing.T) {
		bytes, err := json.Marshal(float32(3.14))
		assertErr(t, err)
		assertEq(t, "float32", `3.14`, string(bytes))
	})
	t.Run("float64", func(t *testing.T) {
		bytes, err := json.Marshal(float64(3.14))
		assertErr(t, err)
		assertEq(t, "float64", `3.14`, string(bytes))
	})
	t.Run("bool", func(t *testing.T) {
		bytes, err := json.Marshal(true)
		assertErr(t, err)
		assertEq(t, "bool", `true`, string(bytes))
	})
	t.Run("string", func(t *testing.T) {
		bytes, err := json.Marshal("hello world")
		assertErr(t, err)
		assertEq(t, "string", `"hello world"`, string(bytes))
	})
	t.Run("struct", func(t *testing.T) {
		bytes, err := json.Marshal(struct {
			A int    `json:"a"`
			B uint   `json:"b"`
			C string `json:"c"`
			D int    `json:"-"`  // ignore field
			a int    `json:"aa"` // private field
		}{
			A: -1,
			B: 1,
			C: "hello world",
		})
		assertErr(t, err)
		assertEq(t, "struct", `{"a":-1,"b":1,"c":"hello world"}`, string(bytes))
		t.Run("null", func(t *testing.T) {
			type T struct {
				A *struct{} `json:"a"`
			}
			var v T
			bytes, err := json.Marshal(&v)
			assertErr(t, err)
			assertEq(t, "struct", `{"a":null}`, string(bytes))
		})
		t.Run("omitempty", func(t *testing.T) {
			type T struct {
				A int                    `json:",omitempty"`
				B int8                   `json:",omitempty"`
				C int16                  `json:",omitempty"`
				D int32                  `json:",omitempty"`
				E int64                  `json:",omitempty"`
				F uint                   `json:",omitempty"`
				G uint8                  `json:",omitempty"`
				H uint16                 `json:",omitempty"`
				I uint32                 `json:",omitempty"`
				J uint64                 `json:",omitempty"`
				K float32                `json:",omitempty"`
				L float64                `json:",omitempty"`
				O string                 `json:",omitempty"`
				P bool                   `json:",omitempty"`
				Q []int                  `json:",omitempty"`
				R map[string]interface{} `json:",omitempty"`
				S *struct{}              `json:",omitempty"`
				T int                    `json:"t,omitempty"`
			}
			var v T
			v.T = 1
			bytes, err := json.Marshal(&v)
			assertErr(t, err)
			assertEq(t, "struct", `{"t":1}`, string(bytes))
		})
		t.Run("head_omitempty", func(t *testing.T) {
			type T struct {
				A *struct{} `json:"a,omitempty"`
			}
			var v T
			bytes, err := json.Marshal(&v)
			assertErr(t, err)
			assertEq(t, "struct", `{}`, string(bytes))
		})
		t.Run("pointer_head_omitempty", func(t *testing.T) {
			type V struct{}
			type U struct {
				B *V `json:"b,omitempty"`
			}
			type T struct {
				A *U `json:"a"`
			}
			bytes, err := json.Marshal(&T{A: &U{}})
			assertErr(t, err)
			assertEq(t, "struct", `{"a":{}}`, string(bytes))
		})
		t.Run("head_int_omitempty", func(t *testing.T) {
			type T struct {
				A int `json:"a,omitempty"`
			}
			var v T
			bytes, err := json.Marshal(&v)
			assertErr(t, err)
			assertEq(t, "struct", `{}`, string(bytes))
		})
	})
	t.Run("slice", func(t *testing.T) {
		t.Run("[]int", func(t *testing.T) {
			bytes, err := json.Marshal([]int{1, 2, 3, 4})
			assertErr(t, err)
			assertEq(t, "[]int", `[1,2,3,4]`, string(bytes))
		})
		t.Run("[]interface{}", func(t *testing.T) {
			bytes, err := json.Marshal([]interface{}{1, 2.1, "hello"})
			assertErr(t, err)
			assertEq(t, "[]interface{}", `[1,2.1,"hello"]`, string(bytes))
		})
	})

	t.Run("array", func(t *testing.T) {
		bytes, err := json.Marshal([4]int{1, 2, 3, 4})
		assertErr(t, err)
		assertEq(t, "array", `[1,2,3,4]`, string(bytes))
	})
	t.Run("map", func(t *testing.T) {
		t.Run("map[string]int", func(t *testing.T) {
			bytes, err := json.Marshal(map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
				"d": 4,
			})
			assertErr(t, err)
			assertEq(t, "map", len(`{"a":1,"b":2,"c":3,"d":4}`), len(string(bytes)))
		})
		t.Run("map[string]interface{}", func(t *testing.T) {
			type T struct {
				A int
			}
			v := map[string]interface{}{
				"a": 1,
				"b": 2.1,
				"c": &T{
					A: 10,
				},
				"d": 4,
			}
			bytes, err := json.Marshal(v)
			assertErr(t, err)
			assertEq(t, "map[string]interface{}", len(`{"a":1,"b":2.1,"c":{"A":10},"d":4}`), len(string(bytes)))
		})
	})
}

type marshalJSON struct{}

func (*marshalJSON) MarshalJSON() ([]byte, error) {
	return []byte(`1`), nil
}

func Test_MarshalJSON(t *testing.T) {
	t.Run("*struct", func(t *testing.T) {
		bytes, err := json.Marshal(&marshalJSON{})
		assertErr(t, err)
		assertEq(t, "MarshalJSON", "1", string(bytes))
	})
	t.Run("time", func(t *testing.T) {
		bytes, err := json.Marshal(time.Time{})
		assertErr(t, err)
		assertEq(t, "MarshalJSON", `"0001-01-01T00:00:00Z"`, string(bytes))
	})
}

func Test_MarshalIndent(t *testing.T) {
	prefix := "-"
	indent := "\t"
	t.Run("struct", func(t *testing.T) {
		bytes, err := json.MarshalIndent(struct {
			A int    `json:"a"`
			B uint   `json:"b"`
			C string `json:"c"`
			D int    `json:"-"`  // ignore field
			a int    `json:"aa"` // private field
		}{
			A: -1,
			B: 1,
			C: "hello world",
		}, prefix, indent)
		assertErr(t, err)
		result := "{\n-\t\"a\": -1,\n-\t\"b\": 1,\n-\t\"c\": \"hello world\"\n-}"
		assertEq(t, "struct", result, string(bytes))
	})
	t.Run("slice", func(t *testing.T) {
		t.Run("[]int", func(t *testing.T) {
			bytes, err := json.MarshalIndent([]int{1, 2, 3, 4}, prefix, indent)
			assertErr(t, err)
			result := "[\n-\t1,\n-\t2,\n-\t3,\n-\t4\n-]"
			assertEq(t, "[]int", result, string(bytes))
		})
		t.Run("[]interface{}", func(t *testing.T) {
			bytes, err := json.MarshalIndent([]interface{}{1, 2.1, "hello"}, prefix, indent)
			assertErr(t, err)
			result := "[\n-\t1,\n-\t2.1,\n-\t\"hello\"\n-]"
			assertEq(t, "[]interface{}", result, string(bytes))
		})
	})

	t.Run("array", func(t *testing.T) {
		bytes, err := json.MarshalIndent([4]int{1, 2, 3, 4}, prefix, indent)
		assertErr(t, err)
		result := "[\n-\t1,\n-\t2,\n-\t3,\n-\t4\n-]"
		assertEq(t, "array", result, string(bytes))
	})
	t.Run("map", func(t *testing.T) {
		t.Run("map[string]int", func(t *testing.T) {
			bytes, err := json.MarshalIndent(map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
				"d": 4,
			}, prefix, indent)
			assertErr(t, err)
			result := "{\n-\t\"a\": 1,\n-\t\"b\": 2,\n-\t\"c\": 3,\n-\t\"d\": 4\n-}"
			assertEq(t, "map", len(result), len(string(bytes)))
		})
		t.Run("map[string]interface{}", func(t *testing.T) {
			type T struct {
				E int
				F int
			}
			v := map[string]interface{}{
				"a": 1,
				"b": 2.1,
				"c": &T{
					E: 10,
					F: 11,
				},
				"d": 4,
			}
			bytes, err := json.MarshalIndent(v, prefix, indent)
			assertErr(t, err)
			result := "{\n-\t\"a\": 1,\n-\t\"b\": 2.1,\n-\t\"c\": {\n-\t\t\"E\": 10,\n-\t\t\"F\": 11\n-\t},\n-\t\"d\": 4\n-}"
			assertEq(t, "map[string]interface{}", len(result), len(string(bytes)))
		})
	})
}

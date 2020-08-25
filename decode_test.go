package json_test

import (
	"fmt"
	"reflect"
	"strings"
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
	t.Run("float32", func(t *testing.T) {
		var v float32
		assertErr(t, json.Unmarshal([]byte(`3.14`), &v))
		assertEq(t, "float32", float32(3.14), v)
	})
	t.Run("float64", func(t *testing.T) {
		var v float64
		assertErr(t, json.Unmarshal([]byte(`3.14`), &v))
		assertEq(t, "float64", float64(3.14), v)
	})
	t.Run("slice", func(t *testing.T) {
		var v []int
		assertErr(t, json.Unmarshal([]byte(` [ 1 , 2 , 3 , 4 ] `), &v))
		assertEq(t, "slice", fmt.Sprint([]int{1, 2, 3, 4}), fmt.Sprint(v))
	})
	t.Run("array", func(t *testing.T) {
		var v [4]int
		assertErr(t, json.Unmarshal([]byte(` [ 1 , 2 , 3 , 4 ] `), &v))
		assertEq(t, "array", fmt.Sprint([4]int{1, 2, 3, 4}), fmt.Sprint(v))
	})
	t.Run("map", func(t *testing.T) {
		var v map[string]int
		assertErr(t, json.Unmarshal([]byte(` { "a": 1, "b": 2, "c": 3, "d": 4 } `), &v))
		assertEq(t, "map.a", v["a"], 1)
		assertEq(t, "map.b", v["b"], 2)
		assertEq(t, "map.c", v["c"], 3)
		assertEq(t, "map.d", v["d"], 4)
		t.Run("nested map", func(t *testing.T) {
			// https://github.com/goccy/go-json/issues/8
			content := `
{
  "a": {
    "nestedA": "value of nested a"
  },  
  "b": {
    "nestedB": "value of nested b"
  },  
  "c": {
    "nestedC": "value of nested c"
  }
}`
			var v map[string]interface{}
			assertErr(t, json.Unmarshal([]byte(content), &v))
			assertEq(t, "length", 3, len(v))
		})
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
		t.Run("struct.field null", func(t *testing.T) {
			var v struct {
				A string
				B []string
				C []int
				D map[string]interface{}
				E [2]string
				F interface{}
			}
			assertErr(t, json.Unmarshal([]byte(`{"a":null,"b":null,"c":null,"d":null,"e":null,"f":null}`), &v))
			assertEq(t, "string", v.A, "")
			assertNeq(t, "[]string", v.B, nil)
			assertEq(t, "[]string", len(v.B), 0)
			assertNeq(t, "[]int", v.C, nil)
			assertEq(t, "[]int", len(v.C), 0)
			assertNeq(t, "map", v.D, nil)
			assertEq(t, "map", len(v.D), 0)
			assertNeq(t, "array", v.E, nil)
			assertEq(t, "array", len(v.E), 2)
			assertEq(t, "interface{}", v.F, nil)
		})
	})
	t.Run("interface", func(t *testing.T) {
		t.Run("number", func(t *testing.T) {
			var v interface{}
			assertErr(t, json.Unmarshal([]byte(`10`), &v))
			assertEq(t, "interface.kind", "float64", reflect.TypeOf(v).Kind().String())
			assertEq(t, "interface", `10`, fmt.Sprint(v))
		})
		t.Run("string", func(t *testing.T) {
			var v interface{}
			assertErr(t, json.Unmarshal([]byte(`"hello"`), &v))
			assertEq(t, "interface.kind", "string", reflect.TypeOf(v).Kind().String())
			assertEq(t, "interface", `hello`, fmt.Sprint(v))
		})
		t.Run("bool", func(t *testing.T) {
			var v interface{}
			assertErr(t, json.Unmarshal([]byte(`true`), &v))
			assertEq(t, "interface.kind", "bool", reflect.TypeOf(v).Kind().String())
			assertEq(t, "interface", `true`, fmt.Sprint(v))
		})
		t.Run("slice", func(t *testing.T) {
			var v interface{}
			assertErr(t, json.Unmarshal([]byte(`[1,2,3,4]`), &v))
			assertEq(t, "interface.kind", "slice", reflect.TypeOf(v).Kind().String())
			assertEq(t, "interface", `[1 2 3 4]`, fmt.Sprint(v))
		})
		t.Run("map", func(t *testing.T) {
			var v interface{}
			assertErr(t, json.Unmarshal([]byte(`{"a": 1, "b": "c"}`), &v))
			assertEq(t, "interface.kind", "map", reflect.TypeOf(v).Kind().String())
			m := v.(map[string]interface{})
			assertEq(t, "interface", `1`, fmt.Sprint(m["a"]))
			assertEq(t, "interface", `c`, fmt.Sprint(m["b"]))
		})
		t.Run("null", func(t *testing.T) {
			var v interface{}
			v = 1
			assertErr(t, json.Unmarshal([]byte(`null`), &v))
			assertEq(t, "interface", nil, v)
		})
	})
}

func Test_Decoder_UseNumber(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`{"a": 3.14}`))
	dec.UseNumber()
	var v map[string]interface{}
	assertErr(t, dec.Decode(&v))
	assertEq(t, "json.Number", "json.Number", fmt.Sprintf("%T", v["a"]))
}

func Test_Decoder_DisallowUnknownFields(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`{"x": 1}`))
	dec.DisallowUnknownFields()
	var v struct {
		x int
	}
	err := dec.Decode(&v)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
	if err.Error() != `json: unknown field "x"` {
		t.Fatal("expected unknown field error")
	}
}

type unmarshalJSON struct {
	v int
}

func (u *unmarshalJSON) UnmarshalJSON(b []byte) error {
	var v int
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	u.v = v
	return nil
}

func Test_UnmarshalJSON(t *testing.T) {
	t.Run("*struct", func(t *testing.T) {
		var v unmarshalJSON
		assertErr(t, json.Unmarshal([]byte(`10`), &v))
		assertEq(t, "unmarshal", v.v, 10)
	})
}

type unmarshalText struct {
	v int
}

func (u *unmarshalText) UnmarshalText(b []byte) error {
	var v int
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	u.v = v
	return nil
}

func Test_UnmarshalText(t *testing.T) {
	t.Run("*struct", func(t *testing.T) {
		var v unmarshalText
		assertErr(t, json.Unmarshal([]byte(`11`), &v))
		assertEq(t, "unmarshal", v.v, 11)
	})
}

func Test_InvalidUnmarshalError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var v *struct{}
		err := fmt.Sprint(json.Unmarshal([]byte(`{}`), v))
		assertEq(t, "invalid unmarshal error", "json: Unmarshal(nil *struct {})", err)
	})
	t.Run("non pointer", func(t *testing.T) {
		var v int
		err := fmt.Sprint(json.Unmarshal([]byte(`{}`), v))
		assertEq(t, "invalid unmarshal error", "json: Unmarshal(non-pointer int)", err)
	})
}

func Test_Token(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`{"a": 1, "b": true, "c": [1, "two", null]}`))
	cnt := 0
	for {
		if _, err := dec.Token(); err != nil {
			break
		}
		cnt++
	}
	if cnt != 12 {
		t.Fatal("failed to parse token")
	}
}

func Test_DecodeStream(t *testing.T) {
	const stream = `
	[
		{"Name": "Ed", "Text": "Knock knock."},
		{"Name": "Sam", "Text": "Who's there?"},
		{"Name": "Ed", "Text": "Go fmt."},
		{"Name": "Sam", "Text": "Go fmt who?"},
		{"Name": "Ed", "Text": "Go fmt yourself!"}
	]
`
	type Message struct {
		Name, Text string
	}
	dec := json.NewDecoder(strings.NewReader(stream))

	tk, err := dec.Token()
	assertErr(t, err)
	assertEq(t, "[", fmt.Sprint(tk), "[")

	elem := 0
	// while the array contains values
	for dec.More() {
		var m Message
		// decode an array value (Message)
		assertErr(t, dec.Decode(&m))
		if m.Name == "" || m.Text == "" {
			t.Fatal("failed to assign value to struct field")
		}
		elem++
	}
	assertEq(t, "decode count", elem, 5)

	tk, err = dec.Token()
	assertErr(t, err)
	assertEq(t, "]", fmt.Sprint(tk), "]")
}

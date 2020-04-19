package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

func Test_Encoder(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		bytes, err := json.Marshal(-10)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `-10` {
			t.Fatal("failed to encode int")
		}
	})
	t.Run("int8", func(t *testing.T) {
		bytes, err := json.Marshal(int8(-11))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `-11` {
			t.Fatal("failed to encode int8")
		}
	})
	t.Run("int16", func(t *testing.T) {
		bytes, err := json.Marshal(int16(-12))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `-12` {
			t.Fatal("failed to encode int16")
		}
	})
	t.Run("int32", func(t *testing.T) {
		bytes, err := json.Marshal(int32(-13))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `-13` {
			t.Fatal("failed to encode int32")
		}
	})
	t.Run("int64", func(t *testing.T) {
		bytes, err := json.Marshal(int64(-14))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `-14` {
			t.Fatal("failed to encode int64")
		}
	})
	t.Run("uint", func(t *testing.T) {
		bytes, err := json.Marshal(uint(10))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `10` {
			t.Fatal("failed to encode uint")
		}
	})
	t.Run("uint8", func(t *testing.T) {
		bytes, err := json.Marshal(uint8(11))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `11` {
			t.Fatal("failed to encode uint8")
		}
	})
	t.Run("uint16", func(t *testing.T) {
		bytes, err := json.Marshal(uint16(12))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `12` {
			t.Fatal("failed to encode uint16")
		}
	})
	t.Run("uint32", func(t *testing.T) {
		bytes, err := json.Marshal(uint32(13))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `13` {
			t.Fatal("failed to encode uint32")
		}
	})
	t.Run("uint64", func(t *testing.T) {
		bytes, err := json.Marshal(uint64(14))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `14` {
			t.Fatal("failed to encode uint64")
		}
	})
	t.Run("float32", func(t *testing.T) {
		bytes, err := json.Marshal(float32(3.14))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `3.14` {
			t.Fatal("failed to encode float32")
		}
	})
	t.Run("float64", func(t *testing.T) {
		bytes, err := json.Marshal(float64(3.14))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `3.14` {
			t.Fatal("failed to encode float64")
		}
	})
	t.Run("bool", func(t *testing.T) {
		bytes, err := json.Marshal(true)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `true` {
			t.Fatal("failed to encode bool")
		}
	})
	t.Run("string", func(t *testing.T) {
		bytes, err := json.Marshal("hello world")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `"hello world"` {
			t.Fatal("failed to encode string")
		}
	})
	t.Run("struct", func(t *testing.T) {
		bytes, err := json.Marshal(struct {
			A int    `json:"a"`
			B uint   `json:"b"`
			C string `json:"c"`
		}{
			A: -1,
			B: 1,
			C: "hello world",
		})
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `{"a":-1,"b":1,"c":"hello world"}` {
			t.Fatal("failed to encode struct")
		}
	})
	t.Run("slice", func(t *testing.T) {
		bytes, err := json.Marshal([]int{1, 2, 3, 4})
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(bytes) != `[1,2,3,4]` {
			t.Fatal("failed to encode slice of int")
		}
	})
}

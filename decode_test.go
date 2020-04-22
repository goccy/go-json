package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

func Test_Decoder(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		var v struct {
			A int
		}
		assertErr(t, json.Unmarshal([]byte(`{"a":123}`), &v))
		assertEq(t, "struct.A", v.A, 123)
	})
}

package json_test

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
)

type encodeMaxDepthList struct{ Next *encodeMaxDepthList }

// A value nested deeper than 10000 objects and arrays fails, as encoding/json of Go 1.27 fails for it.
func TestEncodeMaxDepth(t *testing.T) {
	values := map[string]func(n int) any{
		"interface values": func(n int) any {
			var v any = 1
			for range n {
				v = []any{v}
			}
			return v
		},
		"interface values of time.Time": func(n int) any {
			var v any = time.Time{}
			for range n {
				v = []any{v}
			}
			return v
		},
		"list": func(n int) any {
			var l *encodeMaxDepthList
			for range n {
				l = &encodeMaxDepthList{l}
			}
			return l
		},
	}
	for name, of := range values {
		t.Run(name, func(t *testing.T) {
			if _, err := json.Marshal(of(10000)); err != nil {
				t.Errorf("10000 levels: %v", err)
			}
			if _, err := json.Marshal(of(10001)); err == nil {
				t.Error("10001 levels: got no error")
			}
			if _, err := json.MarshalIndent(of(10001), "", " "); err == nil {
				t.Error("10001 levels indented: got no error")
			}
		})
	}
}

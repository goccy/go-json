package json_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/goccy/go-json"
)

func TestUnmarshalPath(t *testing.T) {
	t.Run("UnmarshalPath", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
			t.Run("success", func(t *testing.T) {
				var v int
				if err := json.UnmarshalPath("a.b", src, &v); err != nil {
					t.Fatal(err)
				}
				if v != 10 {
					t.Fatal("failed to unmarshal path")
				}
			})
			t.Run("failure", func(t *testing.T) {
				var v int
				if err := json.UnmarshalPath("a.c", src, &v); err == nil {
					t.Fatal("expected error")
				}
			})
		})
		t.Run("bool", func(t *testing.T) {
			src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
			t.Run("success", func(t *testing.T) {
				var v bool
				if err := json.UnmarshalPath("a.c", src, &v); err != nil {
					t.Fatal(err)
				}
				if !v {
					t.Fatal("failed to unmarshal path")
				}
			})
			t.Run("failure", func(t *testing.T) {
				var v bool
				if err := json.UnmarshalPath("a.b", src, &v); err == nil {
					t.Fatal("expected error")
				}
			})
		})
	})
}

func TestPathGet(t *testing.T) {
	t.Run("selector", func(t *testing.T) {
		var v interface{}
		if err := json.Unmarshal([]byte(`{"a":{"b":10,"c":true},"b":"text"}`), &v); err != nil {
			t.Fatal(err)
		}
		var b int
		if err := json.Get("a.b", v, &b); err != nil {
			t.Fatal(err)
		}
		if b != 10 {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("index", func(t *testing.T) {
		var v interface{}
		if err := json.Unmarshal([]byte(`{"a":[{"b":10,"c":true},{"b":"text"}]}`), &v); err != nil {
			t.Fatal(err)
		}
		var b int
		if err := json.Get("a[0].b", v, &b); err != nil {
			t.Fatal(err)
		}
		if b != 10 {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("indexAll", func(t *testing.T) {
		var v interface{}
		if err := json.Unmarshal([]byte(`{"a":[{"b":1,"c":true},{"b":2},{"b":3}]}`), &v); err != nil {
			t.Fatal(err)
		}
		var b []int
		if err := json.Get("a[*].b", v, &b); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(b, []int{1, 2, 3}) {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("recursive", func(t *testing.T) {
		var v interface{}
		if err := json.Unmarshal([]byte(`{"a":[{"b":1,"c":true},{"b":2},{"b":3}],"a2":{"b":4}}`), &v); err != nil {
			t.Fatal(err)
		}
		var b []int
		if err := json.Get("..b", v, &b); err != nil {
			t.Fatal(err)
		}
		sort.Ints(b)
		if !reflect.DeepEqual(b, []int{1, 2, 3, 4}) {
			t.Fatalf("failed to decode by json.Get")
		}
	})
}

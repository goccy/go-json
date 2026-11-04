package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

// TestDecodeNullIntoBytes decodes null into a []byte which already holds bytes — at the top level, in a struct
// field, in a slice and in a map value — by Unmarshal and by a Decoder, and compares the value it leaves with
// the one encoding/json leaves: "The JSON null value unmarshals into an interface, map, pointer, or slice by
// setting that Go value to nil."
func TestDecodeNullIntoBytes(t *testing.T) {
	type holder struct {
		B  []byte            `json:"b"`
		Bs [][]byte          `json:"bs"`
		M  map[string][]byte `json:"m"`
	}
	cases := []struct {
		name string
		in   string
		// ours returns a target holding bytes already, and a view of it: the view tells nil from empty.
		ours func() (any, func() string)
	}{
		{"top level", "null", func() (any, func() string) {
			b := []byte{1, 2}
			return &b, func() string { return fmt.Sprintf("%v nil=%v", b, b == nil) }
		}},
		{"struct field", `{"b":null}`, func() (any, func() string) {
			v := holder{B: []byte{1, 2}}
			return &v, func() string { return fmt.Sprintf("%v nil=%v", v.B, v.B == nil) }
		}},
		{"slice element", `[null]`, func() (any, func() string) {
			v := [][]byte{{1, 2}}
			return &v, func() string { return fmt.Sprintf("%v nil=%v", v, v[0] == nil) }
		}},
		{"slice of an array", `[null,1]`, func() (any, func() string) {
			v := make([]any, 2)
			v[0] = []byte{1, 2}
			return &v, func() string { return fmt.Sprintf("%v", v[0]) }
		}},
		{"struct slice field", `{"bs":[null]}`, func() (any, func() string) {
			v := holder{Bs: [][]byte{{1, 2}}}
			return &v, func() string { return fmt.Sprintf("%v nil=%v", v.Bs, v.Bs[0] == nil) }
		}},
		{"map value", `{"m":{"k":null}}`, func() (any, func() string) {
			v := holder{M: map[string][]byte{"k": {1, 2}}}
			return &v, func() string { return fmt.Sprintf("%v nil=%v", v.M["k"], v.M["k"] == nil) }
		}},
	}
	for _, tc := range cases {
		got, view := tc.ours()
		want, wantView := tc.ours()
		gotErr := json.Unmarshal([]byte(tc.in), got)
		wantErr := stdjson.Unmarshal([]byte(tc.in), want)
		if view() != wantView() {
			t.Errorf("Unmarshal of %s into %T (%s):\n got: %s\nwant: %s", tc.in, got, tc.name, view(), wantView())
		}
		if fmt.Sprint(gotErr) != fmt.Sprint(wantErr) {
			t.Errorf("Unmarshal of %s into %T (%s):\n got: %v\nwant: %v", tc.in, got, tc.name, gotErr, wantErr)
		}

		got, view = tc.ours()
		want, wantView = tc.ours()
		gotErr = json.NewDecoder(strings.NewReader(tc.in)).Decode(got)
		wantErr = stdjson.NewDecoder(strings.NewReader(tc.in)).Decode(want)
		if view() != wantView() {
			t.Errorf("Decode of %s into %T (%s):\n got: %s\nwant: %s", tc.in, got, tc.name, view(), wantView())
		}
		if fmt.Sprint(gotErr) != fmt.Sprint(wantErr) {
			t.Errorf("Decode of %s into %T (%s):\n got: %v\nwant: %v", tc.in, got, tc.name, gotErr, wantErr)
		}
	}
}

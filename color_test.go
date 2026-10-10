package json_test

import (
	stdjson "encoding/json"
	"regexp"
	"testing"

	"github.com/goccy/go-json"
)

func TestColorize(t *testing.T) {
	v := struct {
		A int
		B uint
		C float32
		D string
		E bool
		F []byte
		G []int
		H *struct{}
		I map[string]any
	}{
		A: 123,
		B: 456,
		C: 3.14,
		D: "hello",
		E: true,
		F: []byte("binary"),
		G: []int{1, 2, 3, 4},
		H: nil,
		I: map[string]any{
			"mapA": -10,
			"mapB": 10,
			"mapC": nil,
		},
	}
	t.Run("marshal with color", func(t *testing.T) {
		b, err := json.MarshalWithOption(v, json.Colorize(json.DefaultColorScheme))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(b))
	})
	t.Run("marshal indent with color", func(t *testing.T) {
		b, err := json.MarshalIndentWithOption(v, "", "\t", json.Colorize(json.DefaultColorScheme))
		if err != nil {
			t.Fatal(err)
		}
		t.Log("\n" + string(b))
	})
}

// The quoted string of the option ,string is a string, which is colored as one: the string inside it is not
// colored, so the output without the colors is the one of encoding/json.
func TestColorizeStringTag(t *testing.T) {
	s := "x"
	type value struct {
		S string  `json:"s,string"`
		O string  `json:"o,omitempty,string"`
		P *string `json:"p,string"`
		Q *string `json:"q,omitempty,string"`
		L string  `json:"l,string"`
	}
	values := []any{
		value{S: "a", O: "b", P: &s, Q: &s, L: "c"},
		&value{S: "a", P: &s, L: "c"},
		[]value{{}},
		struct {
			S string `json:"s,string"`
		}{"a"},
	}
	colors := regexp.MustCompile("\x1b\\[[0-9;]*m")
	for _, v := range values {
		want, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.MarshalWithOption(v, json.Colorize(json.DefaultColorScheme))
		if err != nil {
			t.Fatal(err)
		}
		if got := colors.ReplaceAll(got, nil); string(got) != string(want) {
			t.Errorf("got %s, want %s", got, want)
		}
		wantIndent, err := stdjson.MarshalIndent(v, "", "\t")
		if err != nil {
			t.Fatal(err)
		}
		gotIndent, err := json.MarshalIndentWithOption(v, "", "\t", json.Colorize(json.DefaultColorScheme))
		if err != nil {
			t.Fatal(err)
		}
		if got := colors.ReplaceAll(gotIndent, nil); string(got) != string(wantIndent) {
			t.Errorf("indented: got %s, want %s", got, wantIndent)
		}
	}
}

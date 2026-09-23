package json_test

import (
	stdjson "encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

// A map whose key is a plain string is encoded by fewer opcodes than a map with a key of another kind, which
// its own opcode writes: both must encode as encoding/json does, sorted and with an indent, and the option
// to leave the map unsorted must write every entry.

type mapKeyString string

type mapKeyText struct{ V int }

func (k mapKeyText) MarshalText() ([]byte, error) { return []byte(strings.Repeat("t", k.V)), nil }

func TestEncodeMapKeys(t *testing.T) {
	values := []any{
		map[string]int{},
		map[string]int{"": 0},
		map[string]int{"b": 2, "a": 1, "c": 3},
		map[string]any{"a": 1, "b": 2.5, "c": "s", "d": nil, "e": []int{1}, "f": map[string]int{"x": 1}},
		map[string]string{"k\"q": "v", "<html>": "&", "日本": "語", "\x00": "c"},
		map[mapKeyString]int{"y": 1, "x": 2},
		map[string]*int{"nil": nil},
		map[string]map[string]int{"outer": {"inner": 1}, "empty": {}, "nil": nil},
		map[int]string{2: "b", 10: "j", 1: "a"},
		map[mapKeyText]int{{1}: 1, {3}: 3, {2}: 2},
		struct {
			M  map[string]int  `json:"m"`
			P  *map[string]int `json:"p"`
			N  map[string]int  `json:"n,omitempty"`
			MM map[string]bool `json:"mm"`
		}{M: map[string]int{"b": 1, "a": 2}, P: &map[string]int{"z": 0}, MM: map[string]bool{"t": true, "f": false}},
		[]map[string]int{{"a": 1}, nil, {}},
		json.Number("1"),
	}
	for _, v := range values {
		expected, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T:\n got %s\nwant %s", v, got, expected)
		}
		expected, _ = stdjson.MarshalIndent(v, "", "  ")
		got, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(got) != string(expected) {
			t.Errorf("%T with indent:\n got %s\nwant %s", v, got, expected)
		}
	}
}

func TestEncodeMapUnordered(t *testing.T) {
	m := map[string]any{"a": 1, "b": 2.5, "c": "s", "d": nil, "e": []int{1}, "f": map[string]int{"x": 1}}
	got, err := json.MarshalWithOption(m, json.UnorderedMap())
	if err != nil {
		t.Fatal(err)
	}
	// every entry is written, in the order of the map: sorted here to compare.
	var decoded map[string]any
	if err := stdjson.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("%s: %v", got, err)
	}
	keys := make([]string, 0, len(decoded))
	for k := range decoded {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if strings.Join(keys, ",") != "a,b,c,d,e,f" {
		t.Fatalf("unexpected keys %v in %s", keys, got)
	}
	indented, err := json.MarshalIndentWithOption(m, "", " ", json.UnorderedMap())
	if err != nil {
		t.Fatal(err)
	}
	if err := stdjson.Unmarshal(indented, &decoded); err != nil || len(decoded) != 6 {
		t.Fatalf("%s: %v", indented, err)
	}
}

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// A key of json.Number is named by its string, and a key of a pointer by the value it points to, as encoding/json
// of Go 1.27 names them: they were written as numbers without quotes, which is not JSON.
func TestEncodeNumberAndPointerMapKeys(t *testing.T) {
	three, i, u, s := stdjson.Number("3"), 4, uint8(5), "a"
	pi, pthree := &i, &three
	tests := []struct {
		name string
		v    any
		want string
	}{
		{"json.Number", map[stdjson.Number]int{"1.5": 1, "10": 2, "2": 3}, `{"1.5":1,"10":2,"2":3}`},
		{"json.Number not a number", map[stdjson.Number]int{"": 1, "x": 2}, `{"":1,"x":2}`},
		{"in a struct", struct{ M map[stdjson.Number]string }{map[stdjson.Number]string{"7": "a"}}, `{"M":{"7":"a"}}`},
		{"pointer to json.Number", map[*stdjson.Number]int{&three: 1}, `{"3":1}`},
		{"pointer to pointer to json.Number", map[**stdjson.Number]int{&pthree: 1}, `{"3":1}`},
		{"pointer to int", map[*int]int{&i: 1}, `{"4":1}`},
		{"pointer to pointer to int", map[**int]int{&pi: 1}, `{"4":1}`},
		{"pointer to uint8", map[*uint8]int{&u: 1}, `{"5":1}`},
		{"pointer to string", map[*string]int{&s: 1}, `{"a":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := json.Marshal(tt.v); err != nil || string(got) != tt.want {
				t.Errorf("Marshal: got %s %v, want %s", got, err, tt.want)
			}
			var b bytes.Buffer
			if err := json.NewEncoder(&b).Encode(tt.v); err != nil || b.String() != tt.want+"\n" {
				t.Errorf("Encode: got %s %v, want %s", b.String(), err, tt.want)
			}
			var indented bytes.Buffer
			if err := stdjson.Indent(&indented, []byte(tt.want), "", " "); err != nil {
				t.Fatal(err)
			}
			if got, err := json.MarshalIndent(tt.v, "", " "); err != nil || string(got) != indented.String() {
				t.Errorf("MarshalIndent: got %s %v, want %s", got, err, indented.String())
			}
		})
	}
	// a nil pointer has no name, nor a value which can't be a key, nor a type which points to itself.
	var self numberKeySelf
	self = &self
	yes := true
	for _, v := range []any{map[*int]int{nil: 1}, map[*string]int{nil: 1}, map[**int]int{new(*int): 1}, map[*stdjson.Number]int{nil: 1},
		map[*bool]int{&yes: 1}, map[*struct{ A int }]int{{1}: 1}, map[numberKeySelf]int{self: 1}} {
		if got, err := json.Marshal(v); err == nil {
			t.Errorf("%T: got %s, want an error", v, got)
		}
		if got, err := json.MarshalIndent(v, "", " "); err == nil {
			t.Errorf("%T: MarshalIndent got %s, want an error", v, got)
		}
	}
}

// numberKeySelf is a type of pointers which points to itself.
type numberKeySelf *numberKeySelf

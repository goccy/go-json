package json_test

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

type stringTagUnmarshaler int

func (n *stringTagUnmarshaler) UnmarshalJSON(b []byte) error {
	v, err := strconv.Atoi(string(b))
	*n = stringTagUnmarshaler(v + 100)
	return err
}

// The option ,string applies to a field of a string, a number or a bool, through one pointer, named or not, and not
// to the value of a pointer to a pointer, nor to a value of another kind or with a method of unmarshaling, as
// encoding/json of Go 1.27 does. The string "null" is null.
func TestDecodeStringTagPointers(t *testing.T) {
	i, s := 5, "x"
	pi, ps := &i, &s
	date := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	unmarshaler := stringTagUnmarshaler(105)
	named := stringTagNamedPointer(&i)
	tests := []struct {
		name    string
		data    string
		v       any // a pointer to the value, which is decoded into
		want    any
		wantErr bool
	}{
		{"pointer", `{"A":"5"}`, &struct {
			A *int `json:",string"`
		}{}, &struct {
			A *int `json:",string"`
		}{&i}, false},
		{"named pointer", `{"A":"5"}`, &struct {
			A stringTagNamedPointer `json:",string"`
		}{}, &struct {
			A stringTagNamedPointer `json:",string"`
		}{named}, false},
		{"named pointer, number not in a string", `{"A":5}`, &struct {
			A stringTagNamedPointer `json:",string"`
		}{}, nil, true},
		{"pointer to a pointer", `{"A":5}`, &struct {
			A **int `json:",string"`
		}{}, &struct {
			A **int `json:",string"`
		}{&pi}, false},
		{"pointer to a pointer, number in a string", `{"A":"5"}`, &struct {
			A **int `json:",string"`
		}{}, nil, true},
		{"pointer to a pointer of a string", `{"A":"x"}`, &struct {
			A **string `json:",string"`
		}{}, &struct {
			A **string `json:",string"`
		}{&ps}, false},
		{"pointer to a named pointer", `{"A":5}`, &struct {
			A *stringTagNamedPointer `json:",string"`
		}{}, &struct {
			A *stringTagNamedPointer `json:",string"`
		}{&named}, false},
		{"null in a string into a number", `{"A":"null"}`, &struct {
			A int `json:",string"`
		}{7}, &struct {
			A int `json:",string"`
		}{7}, false},
		{"null in a string into a pointer to a number", `{"A":"null"}`, &struct {
			A *float64 `json:",string"`
		}{new(float64)}, &struct {
			A *float64 `json:",string"`
		}{}, false},
		{"pointer to a time", `{"A":"2020-01-02T00:00:00Z"}`, &struct {
			A *time.Time `json:",string"`
		}{}, &struct {
			A *time.Time `json:",string"`
		}{&date}, false},
		{"pointer to a struct", `{"A":{"B":1}}`, &struct {
			A *struct{ B int } `json:",string"`
		}{}, &struct {
			A *struct{ B int } `json:",string"`
		}{&struct{ B int }{1}}, false},
		{"pointer to a slice", `{"A":[1]}`, &struct {
			A *[]int `json:",string"`
		}{}, &struct {
			A *[]int `json:",string"`
		}{&[]int{1}}, false},
		{"pointer to a number with a method", `{"A":5}`, &struct {
			A *stringTagUnmarshaler `json:",string"`
		}{}, &struct {
			A *stringTagUnmarshaler `json:",string"`
		}{&unmarshaler}, false},
		{"null in a string as a key of a map of numbers", `{"null":1}`, &map[int]int{}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.data), tt.v)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want error %t", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.v, tt.want) {
				t.Errorf("got %+v, want %+v", tt.v, tt.want)
			}
		})
	}
}

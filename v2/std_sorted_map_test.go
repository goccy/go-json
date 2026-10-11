//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"fmt"
	"strings"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The values of a sorted map whose keys are not strings are encoded in the order of the names of the keys, as
// encoding/json/v2 encodes them: a method of a value sees the place of its entry in the output, and what the method
// writes from it is the same with Deterministic.

// stdPlaceLogger and placeLogger log and write the place they are encoded at, the first one to encoding/json/v2,
// the second one to the v2 json package.
type stdPlaceLogger struct{ log *[]string }

func (l stdPlaceLogger) MarshalJSONTo(enc *stdjsontext.Encoder) error {
	k, n := enc.StackIndex(enc.StackDepth())
	s := fmt.Sprintf("%s %v:%d %d", enc.StackPointer(), k, n, enc.OutputOffset())
	*l.log = append(*l.log, s)
	return enc.WriteToken(stdjsontext.String(s))
}

type placeLogger struct{ log *[]string }

func (l placeLogger) MarshalJSONTo(enc *jsontext.Encoder) error {
	k, n := enc.StackIndex(enc.StackDepth())
	s := fmt.Sprintf("%s %v:%d %d", enc.StackPointer(), k, n, enc.OutputOffset())
	*l.log = append(*l.log, s)
	return enc.WriteToken(jsontext.String(s))
}

// failingValue fails to be marshaled when it is odd.
type failingValue int

func (v failingValue) MarshalJSON() ([]byte, error) {
	if v%2 != 0 {
		return nil, fmt.Errorf("odd value %d", int(v))
	}
	return fmt.Appendf(nil, "%d", int(v)), nil
}

type sortedMapTextKey int

func (k sortedMapTextKey) MarshalText() ([]byte, error) { return fmt.Appendf(nil, "t%03d", int(k)), nil }

func TestSortedMapValuesSameAsStd(t *testing.T) {
	const n = 40
	maps := func(value func() any) map[string]any {
		ints := map[int]any{}
		texts := map[sortedMapTextKey]any{}
		nested := map[int]any{}
		for i := range n {
			ints[i*7%n-n/2] = value()
			texts[sortedMapTextKey(n-i)] = value()
			nested[i] = []any{value(), map[uint]any{uint(i): value(), uint(i + 100): value()}}
		}
		return map[string]any{"ints": ints, "texts": texts, "nested": nested}
	}
	var stdLog, ourLog []string
	stdIn := maps(func() any { return stdPlaceLogger{&stdLog} })
	ourIn := maps(func() any { return placeLogger{&ourLog} })
	want, wantErr := stdjson.Marshal(stdIn, stdjson.Deterministic(true))
	got, err := json.Marshal(ourIn, json.Deterministic(true))
	if err != nil || wantErr != nil {
		t.Fatalf("Marshal error = %v; std %v", err, wantErr)
	}
	if string(got) != string(want) {
		t.Errorf("Marshal:\ngot  %s\nwant %s", got, want)
	}
	if g, w := strings.Join(ourLog, "\n"), strings.Join(stdLog, "\n"); g != w {
		t.Errorf("the places which the methods saw:\ngot\n%s\nwant\n%s", g, w)
	}
}

// The error of a sorted map is the one of the first value which fails in the order of the names.
func TestSortedMapValueErrorSameAsStd(t *testing.T) {
	ints := map[int]failingValue{}
	texts := map[sortedMapTextKey]failingValue{}
	for i := range 40 {
		ints[i] = failingValue(i * 7 % 40)
		texts[sortedMapTextKey(i)] = failingValue(i * 7 % 40)
	}
	for _, v := range []any{ints, texts, struct{ A, B any }{0, ints}} {
		compareWithStd(t, fmt.Sprintf("%T", v), v, []json.Options{json.Deterministic(true)})
	}
}

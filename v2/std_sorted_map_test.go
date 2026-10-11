//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
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

// stdNameKey and nameKey are keys named by their method, the first one for encoding/json/v2, the second one for the
// v2 json package. A key whose name is "fail" fails, and one whose name is "int" writes a number. The method logs
// the place it sees.
type stdNameKey struct {
	name string
	log  *[]string
}

func (k stdNameKey) MarshalJSONTo(enc *stdjsontext.Encoder) error {
	kind, n := enc.StackIndex(enc.StackDepth())
	*k.log = append(*k.log, fmt.Sprintf("%q %s %v:%d %d", k.name, enc.StackPointer(), kind, n, enc.OutputOffset()))
	switch k.name {
	case "fail":
		return errors.New("fail")
	case "int":
		return enc.WriteToken(stdjsontext.Int(3))
	}
	return enc.WriteToken(stdjsontext.String(k.name))
}

type nameKey struct {
	name string
	log  *[]string
}

func (k nameKey) MarshalJSONTo(enc *jsontext.Encoder) error {
	kind, n := enc.StackIndex(enc.StackDepth())
	*k.log = append(*k.log, fmt.Sprintf("%q %s %v:%d %d", k.name, enc.StackPointer(), kind, n, enc.OutputOffset()))
	switch k.name {
	case "fail":
		return errors.New("fail")
	case "int":
		return enc.WriteToken(jsontext.Int(3))
	}
	return enc.WriteToken(jsontext.String(k.name))
}

// nameTextKey is a key named by its text.
type nameTextKey string

func (k nameTextKey) MarshalText() ([]byte, error) { return []byte(k), nil }

// The keys of a sorted map are named in the order of the names, as encoding/json/v2 names them, whatever the names
// have: escapes, or a space or a '!', which come before the quote that ends a name. The method of a key sees the
// place of a map which has no member yet, and its error is at that place.
func TestSortedMapKeyNamesSameAsStd(t *testing.T) {
	runes := []rune("ab5 !\"\\<\x01é")
	r := rand.New(rand.NewPCG(1, 2))
	for range 500 {
		var stdLog, ourLog []string
		texts := map[nameTextKey]int{}
		stdKeys, ourKeys := map[stdNameKey]int{}, map[nameKey]int{}
		for i := range 1 + r.IntN(12) {
			var name []rune
			for range r.IntN(4) {
				name = append(name, runes[r.IntN(len(runes))])
			}
			texts[nameTextKey(name)] = i
			stdKeys[stdNameKey{string(name), &stdLog}] = i
			ourKeys[nameKey{string(name), &ourLog}] = i
		}
		if r.IntN(4) == 0 {
			name := []string{"fail", "int"}[r.IntN(2)]
			stdKeys[stdNameKey{name, &stdLog}] = -1
			ourKeys[nameKey{name, &ourLog}] = -1
		}
		compareWithStd(t, "texts", texts, []json.Options{json.Deterministic(true)})
		compareWithStd(t, "texts", struct{ A, M any }{0, texts}, []json.Options{json.Deterministic(true)})
		for _, in := range []struct{ std, our any }{{stdKeys, ourKeys}, {struct{ A, M any }{0, stdKeys}, struct{ A, M any }{0, ourKeys}}} {
			stdLog, ourLog = stdLog[:0], ourLog[:0]
			want, wantErr := stdjson.Marshal(in.std, stdjson.Deterministic(true))
			got, err := json.Marshal(in.our, json.Deterministic(true))
			if (err == nil) != (wantErr == nil) || err == nil && string(got) != string(want) {
				t.Fatalf("Marshal = %s, %v; std %s, %v", got, err, want, wantErr)
			}
			if err != nil {
				var serr *json.SemanticError
				var stdSerr *stdjson.SemanticError
				if errors.As(err, &serr) != errors.As(wantErr, &stdSerr) || serr != nil && (string(serr.JSONPointer) != string(stdSerr.JSONPointer) || serr.ByteOffset != stdSerr.ByteOffset) {
					t.Fatalf("Marshal error = %v; std %v", err, wantErr)
				}
			}
			if err != nil {
				// the keys named before the one which fails depend on the order of the map.
				continue
			}
			// the keys are named in the order of the map, which the logs are sorted for.
			slices.Sort(stdLog)
			slices.Sort(ourLog)
			if g, w := strings.Join(ourLog, "\n"), strings.Join(stdLog, "\n"); g != w {
				t.Fatalf("the stacks which the methods saw:\ngot\n%s\nwant\n%s", g, w)
			}
		}
	}
}

package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// randomJSON returns a random JSON value of objects, arrays and scalars nested up to depth, whose objects may
// repeat a key.
func randomJSON(r *rand.Rand, depth int) string {
	switch k := r.Intn(8); {
	case depth > 0 && k < 3:
		var b strings.Builder
		b.WriteByte('{')
		for i := r.Intn(12); i > 0; i-- {
			fmt.Fprintf(&b, "%q:%s", fmt.Sprintf("k%d", r.Intn(6)), randomJSON(r, depth-1))
			if i > 1 {
				b.WriteByte(',')
			}
		}
		b.WriteByte('}')
		return b.String()
	case depth > 0 && k < 5:
		var elems []string
		for i := r.Intn(5); i > 0; i-- {
			elems = append(elems, randomJSON(r, depth-1))
		}
		return "[" + strings.Join(elems, ",") + "]"
	case k == 5:
		return fmt.Sprint(r.Intn(1000))
	case k == 6:
		return []string{"true", "false", "null"}[r.Intn(3)]
	}
	return fmt.Sprintf("%q", strings.Repeat("s", r.Intn(4)))
}

func TestDecodeAnyMaps(t *testing.T) {
	// Objects decoded into interface{} and into map[string]interface{} are the maps of encoding/json, with the
	// last of the repeated keys, including after a malformed input, whose entries must not be left to the next
	// call.
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 3000; i++ {
		in := randomJSON(r, 4)
		if i%3 == 0 {
			// a malformed input: the entries decoded before the error are dropped
			var v any
			if err := json.Unmarshal([]byte(in[:len(in)/2]+"}}]]"), &v); err == nil && len(in) > 2 {
				_ = v
			}
		}
		var want, got any
		if err := stdjson.Unmarshal([]byte(in), &want); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(in), &got); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s:\n got %v\nwant %v", in, got, want)
		}
		if strings.HasPrefix(in, "{") {
			var wantMap, gotMap map[string]any
			stdjson.Unmarshal([]byte(in), &wantMap)
			if err := json.Unmarshal([]byte(in), &gotMap); err != nil || !reflect.DeepEqual(gotMap, wantMap) {
				t.Fatalf("%s: into a map:\n got %v %v\nwant %v", in, gotMap, err, wantMap)
			}
		}
	}
}

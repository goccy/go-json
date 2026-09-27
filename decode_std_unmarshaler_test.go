package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

// json.RawMessage is given the bytes of the buffer of the decoder without a copy, since its UnmarshalJSON copies
// them: what it holds must not change when the buffer is written again by the next calls and the next values of
// a stream.
func TestDecodeRawMessageKeepsItsBytes(t *testing.T) {
	type message struct {
		ID    int
		Input stdjson.RawMessage
		Ptr   *stdjson.RawMessage
		List  []stdjson.RawMessage
		Map   map[string]stdjson.RawMessage
	}
	input := func(i int) string {
		v := `{"k":"` + strings.Repeat(strconv.Itoa(i), 40) + `","n":[` + strconv.Itoa(i) + `,null]}`
		return `{"ID":` + strconv.Itoa(i) + `,"Input":` + v + `,"Ptr":` + v + `,"List":[` + v + `,"s",null],"Map":{"a":` + v + `}}`
	}
	var stream strings.Builder
	var fromCalls, wants []message
	for i := 0; i < 200; i++ {
		var want, got message
		if err := stdjson.Unmarshal([]byte(input(i)), &want); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(input(i)), &got); err != nil {
			t.Fatal(err)
		}
		wants = append(wants, want)
		fromCalls = append(fromCalls, got)
		stream.WriteString(input(i) + "\n")
	}
	dec := json.NewDecoder(strings.NewReader(stream.String()))
	var fromStream []message
	for {
		var got message
		if err := dec.Decode(&got); err != nil {
			break
		}
		fromStream = append(fromStream, got)
	}
	if len(fromStream) != len(wants) {
		t.Fatalf("expected %d values from the stream but got %d", len(wants), len(fromStream))
	}
	for i, want := range wants {
		for _, got := range []message{fromCalls[i], fromStream[i]} {
			w, _ := stdjson.Marshal(want)
			g, _ := stdjson.Marshal(got)
			if !bytes.Equal(w, g) {
				t.Fatalf("value %d:\n got %s\nwant %s", i, g, w)
			}
		}
	}
	// a RawMessage decoded into keeps using its own array, as UnmarshalJSON does.
	m := make(stdjson.RawMessage, 0, 64)
	before := &m[:1][0]
	if err := json.Unmarshal([]byte(`{"a":1}`), &m); err != nil || string(m) != `{"a":1}` || &m[0] != before {
		t.Fatalf("got %s %v, reused: %v", m, err, &m[0] == before)
	}
}

func BenchmarkDecodeRawMessage(b *testing.B) {
	type message struct {
		Type  string
		Input stdjson.RawMessage
	}
	data := []byte(`{"Type":"tool_use","Input":{"path":"internal/decoder/decode.go","old_string":"` + strings.Repeat("x", 200) + `","new_string":"y"}}`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v message
		if err := json.Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

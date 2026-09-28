package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"log/slog"
	"math/big"
	"net"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

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

// The unmarshalers of the types of the standard library are given the bytes of the buffer without a copy: they
// keep nothing of them. The values must be the ones encoding/json decodes, and must not change when the buffer is
// written again by the next values of a stream.
func TestDecodeStdUnmarshalers(t *testing.T) {
	type record struct {
		Int    *big.Int
		Value  big.Int
		Float  *big.Float
		Rat    *big.Rat
		Level  slog.Level
		Time   time.Time
		Addr   netip.Addr
		Prefix netip.Prefix
		IP     net.IP
		Levels map[slog.Level]int
	}
	input := func(i int) string {
		n := strconv.Itoa(i)
		return `{"Int":1` + strings.Repeat(n, 30) + `,"Value":-` + n + `,"Float":"1.5e` + n + `","Rat":"` + n + `/7",` +
			`"Level":"WARN+` + n + `","Time":"2026-09-28T01:02:03.` + strings.Repeat("1", 1+i%9) + `+09:00",` +
			`"Addr":"2001:db8::` + strconv.FormatInt(int64(i), 16) + `","Prefix":"10.0.0.0/8","IP":"192.0.2.` + strconv.Itoa(i%255) + `",` +
			`"Levels":{"INFO":` + n + `,"DEBUG-2":1}}`
	}
	var stream strings.Builder
	var wants []record
	for i := 0; i < 100; i++ {
		var want, got record
		if err := stdjson.Unmarshal([]byte(input(i)), &want); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(input(i)), &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("value %d:\n got %+v\nwant %+v", i, got, want)
		}
		wants = append(wants, want)
		stream.WriteString(input(i) + "\n")
	}
	dec := json.NewDecoder(strings.NewReader(stream.String()))
	var fromStream []record
	for {
		var got record
		if err := dec.Decode(&got); err != nil {
			break
		}
		fromStream = append(fromStream, got)
	}
	if !reflect.DeepEqual(fromStream, wants) {
		t.Fatal("the values decoded from the stream differ from the ones of encoding/json")
	}
	// null, and the errors of the unmarshalers, as encoding/json has them.
	for _, in := range []string{`{"Int":null,"Time":null,"Addr":null}`, `{"Level":null}`, `{"Int":"x"}`, `{"Level":"LOUD"}`,
		`{"Addr":"1.2.3"}`, `{"Time":"2026-13-01T00:00:00Z"}`, `{"Value":1e3}`} {
		var want, got record
		wantErr := stdjson.Unmarshal([]byte(in), &want)
		gotErr := json.Unmarshal([]byte(in), &got)
		// the value left by an error is not compared: go-json leaves a pointer nil which encoding/json of Go 1.26
		// and earlier allocates before the unmarshaler fails.
		if (wantErr == nil) != (gotErr == nil) || (wantErr == nil && !reflect.DeepEqual(got, want)) {
			t.Fatalf("%s:\n got %+v %v\nwant %+v %v", in, got, gotErr, want, wantErr)
		}
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

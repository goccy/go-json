package json_test

import (
	stdjson "encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

// The outputs of the marshalers of the types of the standard library are written by their appending methods: they
// must be what the marshalers return, for every kind of value, in every place, and the errors must be the ones of
// the marshalers.
func TestEncodeStdMarshalers(t *testing.T) {
	tokyo := time.FixedZone("JST", 9*60*60)
	times := []time.Time{
		{},
		time.Date(2026, 9, 28, 1, 2, 3, 0, time.UTC),
		time.Date(2026, 9, 28, 1, 2, 3, 456789000, time.UTC),
		time.Date(2026, 9, 28, 1, 2, 3, 1, tokyo),
		time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("", -(3*60*60+30*60))),
		time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC),
		time.Date(2026, 9, 28, 1, 2, 3, 0, time.Local),
		time.Unix(1790000000, 5).In(time.FixedZone("", 59*60)),
	}
	type fields struct {
		T  time.Time            `json:"t"`
		P  *time.Time           `json:"p"`
		N  *time.Time           `json:"n"`
		O  time.Time            `json:"o,omitempty"`
		S  []time.Time          `json:"s"`
		M  map[time.Time]string `json:"m"`
		MP map[string]*time.Time
		I  any
	}
	for _, tm := range times {
		tm := tm
		values := []any{
			tm, &tm, []time.Time{tm, tm},
			fields{T: tm, P: &tm, O: tm, S: []time.Time{tm}, M: map[time.Time]string{tm: "v"}, MP: map[string]*time.Time{"a": &tm, "b": nil}, I: tm},
			map[time.Time]int{tm: 1, tm.Add(-time.Second): 2},
		}
		for _, v := range values {
			want, err := stdjson.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			for _, opts := range [][]json.EncodeOptionFunc{nil, {json.DisableHTMLEscape()}, {json.Colorize(&json.ColorScheme{})}} {
				got, err := json.MarshalWithOption(v, opts...)
				if err != nil || string(got) != string(want) {
					t.Fatalf("%v:\n got %s %v\nwant %s", v, got, err, want)
				}
			}
			wantIndent, _ := stdjson.MarshalIndent(v, "", "  ")
			got, err := json.MarshalIndent(v, "", "  ")
			if err != nil || string(got) != string(wantIndent) {
				t.Fatalf("%v with indent:\n got %s %v\nwant %s", v, got, err, wantIndent)
			}
		}
	}
	// a year which RFC 3339 can't have is the error of the marshaler.
	for _, tm := range []time.Time{time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(-1, 1, 1, 0, 0, 0, 0, time.UTC)} {
		_, wantErr := tm.MarshalJSON()
		_, textErr := tm.MarshalText()
		for _, v := range []any{tm, &tm, struct{ T time.Time }{tm}, map[time.Time]int{tm: 1}} {
			_, err := json.Marshal(v)
			var marshalerErr *json.MarshalerError
			if !errors.As(err, &marshalerErr) {
				t.Fatalf("%v: expected a MarshalerError but got %v", v, err)
			}
			if inner := marshalerErr.Unwrap(); inner.Error() != wantErr.Error() && inner.Error() != textErr.Error() {
				t.Fatalf("%v: expected the error of the marshaler but got %v", v, inner)
			}
		}
	}
}

func BenchmarkEncodeTime(b *testing.B) {
	type event struct {
		Created, Updated, Closed time.Time
		Due                      *time.Time
	}
	now := time.Date(2026, 9, 28, 1, 2, 3, 456789000, time.UTC)
	v := &event{Created: now, Updated: now, Closed: now, Due: &now}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"net"
	"net/netip"
	"regexp"
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

// checkStdEncoding checks that the values encode as encoding/json encodes them, with and without the escape of
// HTML and with indent, by the plain VMs and the colored ones.
func checkStdEncoding(t *testing.T, values ...any) {
	t.Helper()
	for _, v := range values {
		want, wantErr := stdjson.Marshal(v)
		var unescaped bytes.Buffer
		enc := stdjson.NewEncoder(&unescaped)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(v)
		for i, opts := range [][]json.EncodeOptionFunc{nil, {json.DisableHTMLEscape()}, {json.Colorize(&json.ColorScheme{})}} {
			want := want
			if i == 1 && wantErr == nil {
				want = bytes.TrimSuffix(unescaped.Bytes(), []byte("\n"))
			}
			got, err := json.MarshalWithOption(v, opts...)
			if (err != nil) != (wantErr != nil) || string(got) != string(want) {
				t.Fatalf("%#v ( options %d ):\n got %s %v\nwant %s %v", v, i, got, err, want, wantErr)
			}
		}
		wantIndent, _ := stdjson.MarshalIndent(v, "", "  ")
		got, err := json.MarshalIndent(v, "", "  ")
		if (err != nil) != (wantErr != nil) || string(got) != string(wantIndent) {
			t.Fatalf("%#v with indent:\n got %s %v\nwant %s %v", v, got, err, wantIndent, wantErr)
		}
	}
}

// The other types of the standard library: the outputs of MarshalJSON are not checked, and the texts are
// appended by AppendText.
func TestEncodeStdMarshalersOfOtherPackages(t *testing.T) {
	huge, _ := new(big.Int).SetString("-123456789012345678901234567890", 10)
	var levelVar slog.LevelVar
	levelVar.Set(slog.LevelWarn + 2)
	type fields struct {
		Int      *big.Int
		NilInt   *big.Int
		Float    *big.Float
		Rat      *big.Rat
		Level    slog.Level
		LevelVar *slog.LevelVar
		Addr     netip.Addr
		Zero     netip.Addr
		Port     netip.AddrPort
		Prefix   netip.Prefix
		IP       net.IP
		NilIP    net.IP
		Regexp   *regexp.Regexp
		Levels   map[slog.Level]int
		Addrs    map[netip.Addr]string
		Any      any
	}
	v := fields{
		Int: huge, Float: big.NewFloat(1.5e300), Rat: big.NewRat(-3, 7), Level: slog.LevelDebug - 1, LevelVar: &levelVar,
		Addr: netip.MustParseAddr("fe80::1%eth0"), Port: netip.MustParseAddrPort("[::1]:8080"),
		Prefix: netip.MustParsePrefix("10.0.0.0/8"), IP: net.ParseIP("192.0.2.1"), Regexp: regexp.MustCompile(`a<b>&c"\d+`),
		Levels: map[slog.Level]int{slog.LevelInfo: 1, slog.LevelError: 2},
		Addrs:  map[netip.Addr]string{netip.MustParseAddr("::1"): "a", netip.MustParseAddr("127.0.0.1"): "b"}, Any: huge,
	}
	checkStdEncoding(t, v, &v, huge, big.NewInt(0), slog.LevelInfo, netip.Addr{}, net.IP{}, []*big.Int{huge, nil})
	// an IP of an invalid length is the error of MarshalText.
	if _, err := json.Marshal(net.IP{1, 2, 3}); err == nil {
		t.Fatal("expected the error of MarshalText")
	}
}

// userTime has the MarshalJSON of time.Time overridden: it is a type of the user, whose output is checked.
type userTime struct{ time.Time }

func (userTime) MarshalJSON() ([]byte, error) { return []byte(`{"a":`), nil }

func TestEncodeUserMarshalerIsChecked(t *testing.T) {
	if _, err := json.Marshal(struct{ T userTime }{}); err == nil {
		t.Fatal("expected an error for the invalid output of the marshaler of a type of the user")
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

func BenchmarkEncodeStdMarshalers(b *testing.B) {
	type record struct {
		Amount *big.Int
		Level  slog.Level
		Addr   netip.Addr
		Prefix netip.Prefix
		IP     net.IP
	}
	v := &record{Amount: big.NewInt(1234567890123), Level: slog.LevelWarn, Addr: netip.MustParseAddr("2001:db8::1"),
		Prefix: netip.MustParsePrefix("192.168.0.0/16"), IP: net.ParseIP("192.0.2.1")}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

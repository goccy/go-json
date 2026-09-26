package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// randomEscapedString returns the JSON literal of a random string of n runes, some of which are escaped in the
// ways JSON has, with runs of plain bytes of every length between the escapes.
func randomEscapedString(r *rand.Rand, n int) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < n; i++ {
		switch k := r.Intn(20); {
		case k == 0:
			b.WriteString([]string{`\n`, `\r`, `\t`, `\"`, `\\`, `\/`, `\b`, `\f`}[r.Intn(8)])
		case k == 1:
			fmt.Fprintf(&b, `\u%04x`, []rune{'a', 'é', 'あ', 0x2028, 0}[r.Intn(5)])
		case k == 2:
			b.WriteString(`\ud83d\ude00`) // a surrogate pair
		case k == 3:
			b.WriteString([]string{"é", "あ", "😀"}[r.Intn(3)])
		default:
			b.WriteByte(byte('a' + r.Intn(26)))
		}
	}
	b.WriteByte('"')
	return b.String()
}

func TestDecodeEscapedStrings(t *testing.T) {
	// Escaped strings of every length, decoded into a string copied out of the buffer ( up to 512 bytes ) and
	// decoded in place ( longer ones ), into an interface{} and a map key, as encoding/json does.
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 3000; i++ {
		lit := randomEscapedString(r, r.Intn(i%700+1))
		for _, in := range []string{lit, `{"k":` + lit + `}`, `{` + lit + `:1}`, `[` + lit + `,` + lit + `]`} {
			var want, got any
			if err := stdjson.Unmarshal([]byte(in), &want); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(in), &got); err != nil {
				t.Fatalf("%s: %v", in, err)
			}
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Fatalf("%s:\n got %q\nwant %q", in, got, want)
			}
		}
		var want, got string
		stdjson.Unmarshal([]byte(lit), &want)
		if err := json.Unmarshal([]byte(lit), &got); err != nil || got != want {
			t.Fatalf("%s: got %q %v, want %q", lit, got, err, want)
		}
		var s struct{ S string }
		if err := json.Unmarshal([]byte(`{"S":`+lit+`}`), &s); err != nil || s.S != want {
			t.Fatalf("%s: got %q %v, want %q", lit, s.S, err, want)
		}
	}
}

func TestDecodeLongStringsMalformed(t *testing.T) {
	// A long string, whose rest is scanned apart ( see scanStringRest ), is valid or not as encoding/json has it,
	// whatever follows its runs of plain bytes.
	plain := strings.Repeat("abcdefgh", 12)
	for _, tail := range []string{`"`, `\n"`, `\u00e9"`, `\ud83d\ude00"`, `\q"`, "\x01\"", `\u12"`, `\u12zz"`, `\`, ``, `é"`, "\xff\""} {
		for _, n := range []int{64, 65, 70, 96} {
			in := `"` + plain[:n] + tail
			for _, doc := range []string{in, `{"k":` + in + `}`, `[` + in + `,1]`} {
				var want, got any
				errStd := stdjson.Unmarshal([]byte(doc), &want)
				err := json.Unmarshal([]byte(doc), &got)
				if (errStd == nil) != (err == nil) {
					t.Fatalf("%q: encoding/json: %v, go-json: %v", doc, errStd, err)
				}
				if errStd == nil && fmt.Sprint(got) != fmt.Sprint(want) {
					t.Fatalf("%q:\n got %q\nwant %q", doc, got, want)
				}
			}
		}
	}
}

func TestDecodeSkippedStrings(t *testing.T) {
	// The strings of the keys of no field are skipped: they are validated as encoding/json does, whatever their
	// length and wherever an escape, a control character or a byte which is not ASCII is in them. Each is
	// decoded from a buffer of its own and from a stream, which have different ends.
	type target struct {
		A int `json:"a"`
	}
	specials := []string{`\n`, `é`, `😀`, "\x01", "é", "\xff", `\x`, `\u12`}
	for n := 0; n < 140; n++ {
		for pos := -1; pos < n; pos++ {
			for _, special := range specials {
				s := strings.Repeat("a", n)
				if pos >= 0 {
					s = s[:pos] + special + s[pos:]
				} else if special != specials[0] {
					continue
				}
				for _, doc := range []string{`{"unknown":"` + s + `","a":1}`, `{"a":1,"unknown":"` + s + `"}`} {
					var want, got, gotStream target
					wantErr := stdjson.Unmarshal([]byte(doc), &want)
					gotErr := json.Unmarshal([]byte(doc), &got)
					streamErr := json.NewDecoder(strings.NewReader(doc)).Decode(&gotStream)
					if (gotErr != nil) != (wantErr != nil) || (streamErr != nil) != (wantErr != nil) {
						t.Fatalf("%q: got the errors %v and %v, want %v", doc, gotErr, streamErr, wantErr)
					}
					if wantErr == nil && (got != want || gotStream != want) {
						t.Fatalf("%q: got %+v and %+v, want %+v", doc, got, gotStream, want)
					}
				}
			}
		}
	}
}

func TestDecodeEscapedStringRuns(t *testing.T) {
	// A string with escapes between runs of every length, short ones and ones longer than the words which the
	// escapes are decoded by, is decoded as encoding/json decodes it, in every mode of the strings.
	escapes := []string{`\n`, `\"`, `\\`, `\/`, `\t`, `é`, `あ`, `😀`, `\ud83d`, `\u0000`}
	r := rand.New(rand.NewSource(13))
	for i := 0; i < 3000; i++ {
		var b strings.Builder
		b.WriteByte('"')
		for n := r.Intn(8); n >= 0; n-- {
			run := r.Intn(4)
			switch run {
			case 0:
				run = r.Intn(8)
			case 1:
				run = 8 + r.Intn(16)
			default:
				run = r.Intn(200)
			}
			for j := 0; j < run; j++ {
				b.WriteByte("abcdefghij é"[r.Intn(12)])
			}
			b.WriteString(escapes[r.Intn(len(escapes))])
		}
		b.WriteByte('"')
		doc := b.String()
		var want string
		if err := stdjson.Unmarshal([]byte(doc), &want); err != nil {
			t.Fatalf("encoding/json: %q: %v", doc, err)
		}
		for _, opts := range [][]json.DecodeOptionFunc{nil, {json.DecodeNoCopyString()}} {
			var got string
			if err := json.UnmarshalWithOption([]byte(doc), &got, opts...); err != nil || got != want {
				t.Fatalf("%q: got %q, %v, want %q", doc, got, err, want)
			}
			var gotOf string
			if err := json.UnmarshalOf([]byte(doc), &gotOf, opts...); err != nil || gotOf != want {
				t.Fatalf("UnmarshalOf %q: got %q, %v, want %q", doc, gotOf, err, want)
			}
		}
		var any1 any
		if err := json.NewDecoder(strings.NewReader(doc)).Decode(&any1); err != nil || any1 != want {
			t.Fatalf("Decode %q: got %q, %v, want %q", doc, any1, err, want)
		}
	}
}

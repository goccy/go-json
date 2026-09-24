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

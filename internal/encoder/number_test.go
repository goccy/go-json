package encoder

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/goccy/go-json/internal/jsonnum"
)

// AppendNumber takes a json.Number by the grammar of the JSON numbers, which its check of the bytes by words
// takes at once, and appends it as it is: random texts of the bytes of numbers, of every length up to 28, and
// after a buffer with and without room.
func TestAppendNumberAsGrammar(t *testing.T) {
	r := rand.New(rand.NewPCG(11, 12))
	alphabet := "0123456789000-+.eEeE x\xff\xfa"
	var inputs []string
	for range 300000 {
		var sb strings.Builder
		for k := 1 + r.IntN(28); k > 0; k-- {
			if r.IntN(4) == 0 {
				sb.WriteByte(alphabet[r.IntN(len(alphabet))])
			} else {
				sb.WriteByte(byte('0' + r.IntN(10)))
			}
		}
		inputs = append(inputs, sb.String())
	}
	inputs = append(inputs, "0", "-0", "-", ".", "0.", ".0", "-.5", "00", "01", "10", "-01", "0.0", "1.5.5", "1e5", "1.", "-1.0",
		"1234567.", "123456.7", "0.123456789012345", "-0.12345678901234", "12345678901234567")
	for _, s := range inputs {
		for _, prefix := range [][]byte{nil, []byte("x"), append(make([]byte, 0, 64), 'x')} {
			got, err := AppendNumber(&RuntimeContext{Option: &Option{}}, append([]byte(nil), prefix...), json.Number(s))
			want := jsonnum.IsValid([]byte(s))
			if (err == nil) != want {
				t.Fatalf("AppendNumber(%q): error %v, want valid %v", s, err, want)
			}
			if err == nil && string(got) != string(prefix)+s {
				t.Fatalf("AppendNumber(%q) = %q", s, got)
			}
		}
	}
}

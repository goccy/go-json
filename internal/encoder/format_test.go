package encoder

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/goccy/go-json/jsontext"
)

// randomFormatInput is random text of the pieces of JSON, which is valid JSON now and then.
func randomFormatInput(r *rand.Rand) string {
	pieces := []string{
		"{", "}", "[", "]", ",", ":", " ", "\n", "\t", `"a"`, `"b\n"`, `"é"`, `"\q"`, `"\u12zz"`, "\"\x01\"",
		"\"<&>\"", "\" \"", "\"\xff\"", `"`, `1`, `-0.5e+3`, `01`, `1.`, `-`, `1e`, `true`, `false`, `null`, `tru`,
		`x`, `{"k":`, `[1,`, `{"k":1}`, `[1,2]`,
	}
	var sb strings.Builder
	for k := r.IntN(12); k > 0; k-- {
		sb.WriteString(pieces[r.IntN(len(pieces))])
	}
	return sb.String()
}

// The pass formats what jsontext formats as encoding/json has it formatted, and rejects what it rejects: a
// rejection is then reported by jsontext, which must find an error.
func TestFormatAsJSONText(t *testing.T) {
	r := rand.New(rand.NewPCG(9, 10))
	inputs := []string{`{}`, `[]`, ` { "a" : [ 1 , { } , [ ] ] } `, "\"a\"\n", strings.Repeat("[", 100) + strings.Repeat("]", 100)}
	for range 200000 {
		inputs = append(inputs, randomFormatInput(r))
	}
	for _, s := range inputs {
		src := append([]byte(s), nul)
		for escape := range 2 {
			want, werr := jsontext.AppendFormat(nil, s, formatOptions[escape]...)
			got, _, ok := formatJSON(nil, src, escape == 1, nil)
			if ok != (werr == nil) || ok && string(got) != string(want) {
				t.Fatalf("compact of %q, escape %d: got %q, %v; want %q, %v", s, escape, got, ok, want, werr)
			}
			opts := append(append([]jsontext.Options{}, formatOptions[escape]...), jsontext.Multiline(true), jsontext.WithIndentPrefix("  "), jsontext.WithIndent("\t"))
			want, werr = jsontext.AppendFormat(nil, s, opts...)
			got, _, ok = formatJSON(nil, src, escape == 1, &formatIndent{prefix: "  ", indent: "\t"})
			if ok != (werr == nil) || ok && string(got) != string(want) {
				t.Fatalf("indent of %q, escape %d: got %q, %v; want %q, %v", s, escape, got, ok, want, werr)
			}
		}
	}
}

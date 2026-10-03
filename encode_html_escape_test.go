package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

// HTMLEscape escapes '<', '>', '&', U+2028 and U+2029 wherever they are and keeps everything else as it is, as
// encoding/json does: the order and the duplicates of the names of an object, the white space, and bytes which
// are not JSON at all.
func TestHTMLEscapeKeepsInput(t *testing.T) {
	inputs := []string{
		`{"b":"<x>","a":1,"a":2}`,
		"{ \"a\" : [ 1 , \"&\" ] }\n",
		"\"  \"",
		`not json <&>`,
		"\xe2\x80",
		"\xe2\x80\xa8",
		"",
	}
	r := rand.New(rand.NewPCG(7, 8))
	pieces := []string{"<", ">", "&", " ", " ", "‧", "\xe2", "\xe2\x80", "é", `"`, `\`, " ", "{", "}", "a"}
	for range 2000 {
		var sb strings.Builder
		for k := r.IntN(80); k > 0; k-- {
			sb.WriteString(pieces[r.IntN(len(pieces))])
		}
		inputs = append(inputs, sb.String())
	}
	for _, in := range inputs {
		var got, want bytes.Buffer
		got.WriteString("prefix")
		want.WriteString("prefix")
		json.HTMLEscape(&got, []byte(in))
		stdjson.HTMLEscape(&want, []byte(in))
		if got.String() != want.String() {
			t.Fatalf("HTMLEscape(%q):\n got %q\nwant %q", in, got.String(), want.String())
		}
	}
}

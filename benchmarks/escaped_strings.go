package benchmark

import (
	"bytes"
	"strconv"

	"github.com/goccy/go-json"
)

type EscapedStrings struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// EscapedStringsFixture returns a serialised version of a slice of the EscapedStrings struct, with the given number of entities in it.
func EscapedStringsFixture(n int) []byte {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i := range n {
		if i > 0 {
			buf.WriteByte(',')
		}
		b, _ := json.Marshal(EscapedStrings{
			ID:   strconv.Itoa(i),
			Text: "This is\ta string\twith some escape sequences\n in it",
		})
		buf.Write(b)
	}
	buf.WriteByte(']')
	return buf.Bytes()
}

package json_test

import (
	"io"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// chunkReader hands out at most n bytes per Read, so a multi-byte character can
// be made to straddle the decoder's buffer boundary the way a network read does.
// The original report reproduced this through an http.Request body; this keeps
// the test hermetic while exercising the same path.
type chunkReader struct {
	data []byte
	n    int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	size := r.n
	if size > len(r.data) {
		size = len(r.data)
	}
	if size > len(p) {
		size = len(p)
	}
	copy(p, r.data[:size])
	r.data = r.data[size:]
	return size, nil
}

// A rune split across a buffer refill used to be judged complete-but-invalid,
// because the check looked past the bytes actually read into nul padding. Each
// byte of the sequence then became U+FFFD, leaving the string 6 bytes longer
// than it was sent. https://github.com/goccy/go-json/issues/574
func TestDecodeStreamMultiByteAcrossBufferBoundary(t *testing.T) {
	type payload struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	for _, tc := range []struct {
		name string
		unit string
	}{
		{"3-byte", "日本語更新テスト "},
		{"2-byte", "återhämtning "},
		{"4-byte", "🎌🗾🈯 "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Sweep lengths so the split lands at every offset within a rune.
			for reps := 300; reps <= 340; reps++ {
				body := strings.Repeat(tc.unit, reps)
				src := `{"title": "boundary", "body": "` + body + `"}`

				var got payload
				if err := json.NewDecoder(&chunkReader{data: []byte(src), n: 1024}).Decode(&got); err != nil {
					t.Fatalf("reps=%d: %v", reps, err)
				}
				if got.Body != body {
					t.Fatalf("reps=%d: body corrupted\n got len=%d\nwant len=%d\nfirst U+FFFD at %d",
						reps, len(got.Body), len(body), strings.Index(got.Body, string(rune(0xFFFD))))
				}
			}
		})
	}
}

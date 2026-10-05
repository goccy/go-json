//go:build go1.24

package benchmark

import (
	"fmt"
	"time"

	gojsontext "github.com/goccy/go-json/jsontext"
)

// The method of github.com/goccy/go-json/v2, which it calls instead of UnmarshalJSON. The time is written by
// MarshalJSON of time.Time, which it embeds: the encoders write it without the call, as they know its output, and
// a method of its own to write it would only be slower ( see encode_marshaler_to_test.go for the methods ).

// UnmarshalJSONFrom reads the time as UnmarshalJSON does: an RFC 3339 string or Unix seconds.
func (t *GitHubTimestamp) UnmarshalJSONFrom(dec *gojsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case '"':
		t.Time, err = time.Parse(time.RFC3339, tok.String())
		return err
	case '0':
		sec, err := tok.Int()
		if err != nil {
			return err
		}
		t.Time = time.Unix(sec, 0)
		return nil
	case 'n':
		// UnmarshalJSON decodes null as the seconds 0.
		t.Time = time.Unix(0, 0)
		return nil
	}
	return fmt.Errorf("GitHubTimestamp: unexpected JSON %v", tok.Kind())
}

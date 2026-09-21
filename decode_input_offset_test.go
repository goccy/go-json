package json_test

import (
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestDecoderInputOffsetEscapedBytes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "escaped quote",
			input: `{"test":"\""}`,
		},
		{
			name:  "unicode escape",
			input: `{"test":"\u0041"}`,
		},
		{
			name:  "unicode surrogate pair",
			input: `{"test":"\uD83D\uDE00"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := json.NewDecoder(strings.NewReader(tt.input))
			var value map[string]string
			if err := dec.Decode(&value); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got, want := dec.InputOffset(), int64(len(tt.input)); got != want {
				t.Fatalf("InputOffset() = %d, want %d", got, want)
			}
		})
	}
}

func TestDecoderInputOffsetEscapedBytesAcrossValues(t *testing.T) {
	input := `{"test":"\""} 42`
	dec := json.NewDecoder(strings.NewReader(input))

	var value map[string]string
	if err := dec.Decode(&value); err != nil {
		t.Fatalf("first Decode() error = %v", err)
	}
	if got, want := dec.InputOffset(), int64(len(`{"test":"\""}`)); got != want {
		t.Fatalf("InputOffset() after first Decode() = %d, want %d", got, want)
	}

	var number int
	if err := dec.Decode(&number); err != nil {
		t.Fatalf("second Decode() error = %v", err)
	}
	if got, want := dec.InputOffset(), int64(len(input)); got != want {
		t.Fatalf("InputOffset() after second Decode() = %d, want %d", got, want)
	}
}

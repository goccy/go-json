package encoder

import (
	"fmt"
	"strings"
	"testing"
)

// AppendString handles a string which has nothing to escape by itself: by bytes under a word, by words with
// the last one overlapping, and with a copy which depends on the length. Every length and every position of
// a byte is compared with the function for the options, which is the one used for a string to escape.
func TestAppendStringFastPath(t *testing.T) {
	special := []string{
		"\x00", "\x1f", " ", "!", `"`, "#", "&", "'", "/", "<", "=", ">", "?", "[", `\`, "]", "\x7f",
		"\x80", "\xe3", "\xff", "あ", " ", " ", "\n", "\t",
	}
	var inputs []string
	for length := 0; length <= 40; length++ {
		base := strings.Repeat("a", length)
		inputs = append(inputs, base)
		for pos := 0; pos <= length; pos++ {
			for _, c := range special {
				inputs = append(inputs, base[:pos]+c+base[pos:])
			}
		}
	}
	for index := range stringEscapes {
		escape := &stringEscapes[index]
		var flag OptionFlag
		if index&stringEscapeHTML != 0 {
			flag |= HTMLEscapeOption
		}
		if index&stringEscapeNormalize != 0 {
			flag |= NormalizeUTF8Option
		}
		ctx := &RuntimeContext{Option: &Option{Flag: flag}}
		t.Run(fmt.Sprintf("html=%v,normalize=%v", index&stringEscapeHTML != 0, index&stringEscapeNormalize != 0), func(t *testing.T) {
			for _, s := range inputs {
				expected := string(escape.appendEscaped([]byte("x"), s))
				// without and with the capacity.
				for _, buf := range [][]byte{[]byte("x"), append(make([]byte, 0, 256), 'x')} {
					if got := string(AppendString(ctx, buf, s)); got != expected {
						t.Fatalf("AppendString(%q): expected %s but got %s", s, expected, got)
					}
				}
			}
		})
	}
}

// The string must not be read out of its memory: the strings here end at the end of their memory.
func TestAppendStringAtEndOfMemory(t *testing.T) {
	ctx := &RuntimeContext{Option: &Option{Flag: HTMLEscapeOption | NormalizeUTF8Option}}
	for length := 1; length <= 24; length++ {
		mem := []byte(strings.Repeat("a", length))
		for start := 0; start < length; start++ {
			s := string(mem[start:])
			if got := string(AppendString(ctx, nil, s)); got != `"`+s+`"` {
				t.Fatalf("AppendString(%q): got %s", s, got)
			}
		}
	}
}

// The buffer grows as append does: the capacity must at least double, or appending many strings is quadratic.
func TestAppendStringGrowth(t *testing.T) {
	ctx := &RuntimeContext{Option: &Option{Flag: HTMLEscapeOption | NormalizeUTF8Option}}
	var buf []byte
	grown := 0
	for i := 0; i < 10000; i++ {
		before := cap(buf)
		buf = AppendString(ctx, buf, "0123456789")
		if cap(buf) != before {
			grown++
		}
	}
	if grown > 32 {
		t.Fatalf("the buffer grew %d times", grown)
	}
	if len(buf) != 10000*12 {
		t.Fatalf("unexpected length %d", len(buf))
	}
}

func BenchmarkAppendString(b *testing.B) {
	ctx := &RuntimeContext{Option: &Option{Flag: HTMLEscapeOption | NormalizeUTF8Option}}
	for _, s := range []string{
		"", "abc", "test42", "127.0.0.1", "user_agent_long", "de305d54-75b4-431b-adb2-eb6b9e546014",
		strings.Repeat("abcdefghij", 5), strings.Repeat("abcdefghij", 10), strings.Repeat("abcdefghij", 30),
		strings.Repeat("abcdefghij", 100), strings.Repeat("abcdefghij", 1000),
	} {
		s := s
		b.Run(fmt.Sprint(len(s)), func(b *testing.B) {
			buf := make([]byte, 0, 4096)
			for i := 0; i < b.N; i++ {
				buf = AppendString(ctx, buf[:0], s)
			}
		})
	}
}

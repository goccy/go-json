package encoder

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"
)

// AppendString handles a string which has nothing to escape by itself: by bytes under a half word, by a word
// made of two overlapping halves under a word, by words with the last one overlapping, and with a copy which
// depends on the length. Every length and every position of a byte is compared with the function for the
// options, which is the one used for a string to escape.
func TestAppendStringFastPath(t *testing.T) {
	special := []string{
		"\x00", "\x1f", " ", "!", `"`, "#", "&", "'", "/", "<", "=", ">", "?", "[", `\`, "]", "\x7f",
		"\x80", "\xe3", "\xff", "あ", " ", " ", "\n", "\t",
	}
	var inputs []string
	for length := 0; length <= 72; length++ {
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
	options := []struct {
		name string
		flag OptionFlag
	}{
		{"std", HTMLEscapeOption | NormalizeUTF8Option},
		{"fastest", 0},
	}
	text := strings.Repeat("abcdefghij", 1000)
	for _, o := range options {
		ctx := &RuntimeContext{Option: &Option{Flag: o.flag}}
		for _, n := range []int{0, 3, 6, 8, 9, 12, 15, 16, 17, 20, 24, 28, 31, 32, 36, 50, 64, 100, 300, 1000, 10000} {
			s := text[:n]
			b.Run(fmt.Sprintf("%s/%d", o.name, n), func(b *testing.B) {
				buf := make([]byte, 0, 16384)
				for i := 0; i < b.N; i++ {
					buf = AppendString(ctx, buf[:0], s)
				}
			})
		}
	}
}

// commonRuneSize takes a character for valid and not escaped only if decodeRuneInString does, for every first
// byte which is not ASCII and every two bytes after it, and at the end of the string.
func TestCommonRuneSize(t *testing.T) {
	var accepted [3]int
	for c := 0x80; c <= 0xff; c++ {
		for s1 := 0; s1 <= 0xff; s1++ {
			for s2 := 0; s2 <= 0xff; s2++ {
				s := string([]byte{byte(c), byte(s1), byte(s2)})
				for n := 1; n <= 3; n++ {
					size := commonRuneSize(s[:n], 0)
					if size == 0 {
						continue
					}
					accepted[size-1]++
					if state, want := decodeRuneInString(s[:n]); state != validUTF8State || want != size {
						t.Fatalf("% x: commonRuneSize %d, decodeRuneInString %d %d", s[:n], size, state, want)
					}
				}
			}
		}
	}
	// the characters of two bytes and the ones of three bytes of 0xE3 to 0xEC, 0xEE and 0xEF are taken.
	if want := 30 * 64 * 256 * 2; accepted[1] != want {
		t.Errorf("characters of two bytes: %d, want %d", accepted[1], want)
	}
	if want := 12 * 64 * 64; accepted[2] != want {
		t.Errorf("characters of three bytes: %d, want %d", accepted[2], want)
	}
}

// BenchmarkAppendStringNotASCII measures strings of Japanese, and of Japanese and ASCII, by the options of the
// std configuration, of the fast one, which normalizes UTF-8 without escaping HTML, and of the fastest one.
func BenchmarkAppendStringNotASCII(b *testing.B) {
	options := []struct {
		name string
		flag OptionFlag
	}{
		{"std", HTMLEscapeOption | NormalizeUTF8Option},
		{"fast", NormalizeUTF8Option},
		{"fastest", 0},
	}
	ja := []rune(strings.Repeat("日本語の文章を書き出します。", 100))
	mixed := []rune(strings.Repeat("東京 Tokyo 2026年9月30日, weather: 晴れ; ", 40))
	texts := []struct {
		name string
		s    string
	}{
		{"ja3", string(ja[:3])}, {"ja10", string(ja[:10])}, {"ja30", string(ja[:30])}, {"ja100", string(ja[:100])},
		{"ja1000", string(ja[:1000])}, {"mixed40", string(mixed[:40])}, {"mixed400", string(mixed[:400])},
	}
	for _, o := range options {
		ctx := &RuntimeContext{Option: &Option{Flag: o.flag}}
		for _, text := range texts {
			s := text.s
			b.Run(o.name+"/"+text.name, func(b *testing.B) {
				buf := make([]byte, 0, 16384)
				b.SetBytes(int64(len(s)))
				for i := 0; i < b.N; i++ {
					buf = AppendString(ctx, buf[:0], s)
				}
			})
		}
	}
}

// referenceNormalizedString is what appendNormalizedString and appendNormalizedHTMLString append, written by
// decoding every character with unicode/utf8.
func referenceNormalizedString(s string, html bool) string {
	b := []byte{'"'}
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			if seq := escapeSequences[c]; seq != 0 && (html || (c != '<' && c != '>' && c != '&')) {
				for k := uint64(0); k < seq>>56; k++ {
					b = append(b, byte(seq>>(8*k)))
				}
			} else {
				b = append(b, c)
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			b = append(b, invalidUTF8...)
		case r == '\u2028':
			b = append(b, `\u2028`...)
		case r == '\u2029':
			b = append(b, `\u2029`...)
		default:
			b = append(b, s[i:i+size]...)
		}
		i += size
	}
	return string(append(b, '"'))
}

// The strings of every length, mostly of characters which are not ASCII, with bytes to escape, invalid UTF-8,
// U+2028 and U+2029 anywhere, are appended by the functions which normalize UTF-8 as they are when every
// character is decoded ( with the loop of the escapes by SIMD if the CPU has it; see also
// TestAppendNormalizedStringsWithoutEscapeLoop ).
func TestAppendNormalizedStrings(t *testing.T) {
	checkAppendNormalizedStrings(t)
}

func checkAppendNormalizedStrings(t *testing.T) {
	t.Helper()
	r := rand.New(rand.NewSource(1))
	pieces := []string{"a", " ", "\"", "\\", "\n", "\x00", "\x1f", "\x7f", "<", ">", "&", "é", "ß", "あ", "語", "。", "Ａ",
		"\u0800", "\uffff", "\ue000", "\ufffd", "😀", "\u2027", "\u2028", "\u2029", "\u202a", "\xff", "\x80", "\xc3",
		"\xe3\x81", "\xe3", "\xf0\x9f\x98", "\xc0\x80", "\xe0\x80\x80", "\xed\xa0\x80", "\xf4\x90\x80\x80"}
	common := []string{"あ", "い", "語", "。", "Ａ", "é", "ß"}
	for n := 0; n < 4000; n++ {
		var b strings.Builder
		for k := r.Intn(n%200 + 1); k > 0; k-- {
			if r.Intn(8) == 0 {
				b.WriteString(pieces[r.Intn(len(pieces))])
			} else {
				b.WriteString(common[r.Intn(len(common))])
			}
		}
		s := b.String()
		if got, want := string(appendNormalizedString(nil, s)), referenceNormalizedString(s, false); got != want {
			t.Fatalf("%q:\n got %q\nwant %q", s, got, want)
		}
		if got, want := string(appendNormalizedHTMLString(nil, s)), referenceNormalizedString(s, true); got != want {
			t.Fatalf("HTML, %q:\n got %q\nwant %q", s, got, want)
		}
	}
}

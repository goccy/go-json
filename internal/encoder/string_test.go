package encoder

import (
	"fmt"
	"strings"
	"testing"
)

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
		ctx.SetEscaper()
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
		ctx.SetEscaper()
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

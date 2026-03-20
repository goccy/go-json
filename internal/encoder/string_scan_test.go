package encoder

import (
	"strings"
	"testing"
	"unsafe"
)

func TestScanEscapeBasic(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"no_escape", "hello world", 11},
		{"quote_start", `"hello`, 0},
		{"quote_middle", `hello"world`, 5},
		{"backslash", `hello\world`, 5},
		{"control_char", "hello\nworld", 5},
		{"null_byte", "hello\x00world", 5},
		{"tab", "hello\tworld", 5},
		{"high_byte", "hello\x80world", 5},
		{"all_safe_short", "abcde", 5},
		{"all_safe_16", "abcdefghijklmnop", 16},
		{"all_safe_32", "abcdefghijklmnopqrstuvwxyz012345", 32},
		{"escape_at_16", "abcdefghijklmno\"rest", 15},
		{"escape_at_17", "abcdefghijklmnop\"rest", 16},
		{"long_no_escape", strings.Repeat("a", 1024), 1024},
		{"long_escape_end", strings.Repeat("a", 1023) + "\"", 1023},
		{"long_escape_middle", strings.Repeat("a", 512) + "\"" + strings.Repeat("b", 511), 512},
		{"only_spaces", strings.Repeat(" ", 64), 64},
		{"control_0x1f", "hello\x1fworld", 5},
		{"just_under_0x20", "\x1f", 0},
		{"exactly_0x20", " ", 1},
		{"del_char", "\x7f", 1}, // 0x7F is safe (not in needEscape)
		{"utf8_2byte", "hello\xc0world", 5},
		{"utf8_start", "\xc0hello", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p unsafe.Pointer
			if len(tt.s) > 0 {
				p = stringptr(tt.s)
			}
			got := scanEscapeBasic(p, len(tt.s))
			if got != tt.want {
				t.Errorf("scanEscapeBasic(%q, %d) = %d, want %d", tt.s, len(tt.s), got, tt.want)
			}
		})
	}
}

func TestScanEscapeHTML(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"no_escape", "hello world", 11},
		{"lt", "hello<world", 5},
		{"gt", "hello>world", 5},
		{"amp", "hello&world", 5},
		{"quote", `hello"world`, 5},
		{"backslash", `hello\world`, 5},
		{"html_entities", "abc<def>ghi&jkl", 3},
		{"safe_long", strings.Repeat("a", 1024), 1024},
		{"lt_at_16", "abcdefghijklmno<rest", 15},
		{"gt_at_32", strings.Repeat("a", 31) + ">rest", 31},
		{"amp_at_17", "abcdefghijklmnop&rest", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p unsafe.Pointer
			if len(tt.s) > 0 {
				p = stringptr(tt.s)
			}
			got := scanEscapeHTML(p, len(tt.s))
			if got != tt.want {
				t.Errorf("scanEscapeHTML(%q, %d) = %d, want %d", tt.s, len(tt.s), got, tt.want)
			}
		})
	}
}

func TestScanEscapeBasicASCIIOnly(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"no_escape", "hello world", 11},
		{"quote_start", `"hello`, 0},
		{"quote_middle", `hello"world`, 5},
		{"backslash", `hello\world`, 5},
		{"control_char", "hello\nworld", 5},
		{"null_byte", "hello\x00world", 5},
		{"tab", "hello\tworld", 5},
		// Key difference: high bytes are NOT flagged
		{"high_byte_safe", "hello\x80world", 11},
		{"utf8_safe", "hello\xc0world", 11},
		{"all_safe_short", "abcde", 5},
		{"all_safe_16", "abcdefghijklmnop", 16},
		{"all_safe_32", "abcdefghijklmnopqrstuvwxyz012345", 32},
		{"escape_at_16", "abcdefghijklmno\"rest", 15},
		{"escape_at_17", "abcdefghijklmnop\"rest", 16},
		{"long_no_escape", strings.Repeat("a", 1024), 1024},
		{"long_escape_end", strings.Repeat("a", 1023) + "\"", 1023},
		// Unicode strings should be fully safe
		{"unicode_all_safe", strings.Repeat("\xe4\xb8\x96\xe7\x95\x8c", 100), 600}, // "世界" * 100
		{"unicode_with_escape", strings.Repeat("\xe4\xb8\x96", 50) + "\"" + strings.Repeat("\xe7\x95\x8c", 50), 150},
		{"mixed_unicode_ascii", "hello " + strings.Repeat("\xe4\xb8\x96\xe7\x95\x8c", 50) + " world", 312},
		{"only_spaces", strings.Repeat(" ", 64), 64},
		{"control_0x1f", "hello\x1fworld", 5},
		{"just_under_0x20", "\x1f", 0},
		{"exactly_0x20", " ", 1},
		{"del_char", "\x7f", 1}, // 0x7F is safe
		{"all_high_bytes", strings.Repeat("\x80\x90\xa0\xb0\xc0\xd0\xe0\xf0", 128), 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p unsafe.Pointer
			if len(tt.s) > 0 {
				p = stringptr(tt.s)
			}
			got := scanEscapeBasicASCIIOnly(p, len(tt.s))
			if got != tt.want {
				t.Errorf("scanEscapeBasicASCIIOnly(%q, %d) = %d, want %d", tt.s, len(tt.s), got, tt.want)
			}
		})
	}
}

func TestScanEscapeHTMLASCIIOnly(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"no_escape", "hello world", 11},
		{"lt", "hello<world", 5},
		{"gt", "hello>world", 5},
		{"amp", "hello&world", 5},
		{"quote", `hello"world`, 5},
		{"backslash", `hello\world`, 5},
		{"html_entities", "abc<def>ghi&jkl", 3},
		{"safe_long", strings.Repeat("a", 1024), 1024},
		{"lt_at_16", "abcdefghijklmno<rest", 15},
		{"gt_at_32", strings.Repeat("a", 31) + ">rest", 31},
		{"amp_at_17", "abcdefghijklmnop&rest", 16},
		// High bytes are NOT flagged
		{"high_byte_safe", "hello\x80world", 11},
		{"unicode_all_safe", strings.Repeat("\xe4\xb8\x96\xe7\x95\x8c", 100), 600},
		{"unicode_with_html", strings.Repeat("\xe4\xb8\x96", 50) + "<" + strings.Repeat("\xe7\x95\x8c", 50), 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p unsafe.Pointer
			if len(tt.s) > 0 {
				p = stringptr(tt.s)
			}
			got := scanEscapeHTMLASCIIOnly(p, len(tt.s))
			if got != tt.want {
				t.Errorf("scanEscapeHTMLASCIIOnly(%q, %d) = %d, want %d", tt.s, len(tt.s), got, tt.want)
			}
		})
	}
}

// Benchmark the SIMD scan functions with various string sizes
func BenchmarkScanEscapeBasic(b *testing.B) {
	sizes := []int{8, 16, 32, 64, 128, 256, 512, 1024}
	for _, size := range sizes {
		s := strings.Repeat("a", size)
		p := stringptr(s)
		n := len(s)
		b.Run(sprintf("%d_no_escape", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				scanEscapeBasic(p, n)
			}
		})
	}

	// Benchmark with escape at the end
	for _, size := range sizes {
		s := strings.Repeat("a", size-1) + "\""
		p := stringptr(s)
		n := len(s)
		b.Run(sprintf("%d_escape_end", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				scanEscapeBasic(p, n)
			}
		})
	}
}

func BenchmarkScanEscapeHTML(b *testing.B) {
	sizes := []int{8, 16, 32, 64, 128, 256, 512, 1024}
	for _, size := range sizes {
		s := strings.Repeat("a", size)
		p := stringptr(s)
		n := len(s)
		b.Run(sprintf("%d_no_escape", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				scanEscapeHTML(p, n)
			}
		})
	}
}

func BenchmarkScanEscapeBasicASCIIOnly(b *testing.B) {
	sizes := []int{16, 32, 64, 128, 256, 512, 1024}

	// ASCII strings
	for _, size := range sizes {
		s := strings.Repeat("a", size)
		p := stringptr(s)
		n := len(s)
		b.Run(sprintf("%d_ascii_no_escape", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				scanEscapeBasicASCIIOnly(p, n)
			}
		})
	}

	// Unicode strings - this is where ASCIIOnly shines vs Basic
	for _, size := range sizes {
		// Each "世界" is 6 bytes
		reps := size / 6
		if reps < 1 {
			reps = 1
		}
		s := strings.Repeat("\xe4\xb8\x96\xe7\x95\x8c", reps)
		p := stringptr(s)
		n := len(s)
		b.Run(sprintf("%d_unicode_no_escape", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				scanEscapeBasicASCIIOnly(p, n)
			}
		})
	}
}

// Benchmark the full appendString function
func BenchmarkAppendString(b *testing.B) {
	opt := &Option{}
	ctx := &RuntimeContext{Option: opt}
	buf := make([]byte, 0, 4096)

	b.Run("short_no_escape", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AppendString(ctx, buf[:0], "hello world")
		}
	})

	b.Run("medium_no_escape", func(b *testing.B) {
		s := strings.Repeat("hello world ", 10)
		for i := 0; i < b.N; i++ {
			AppendString(ctx, buf[:0], s)
		}
	})

	b.Run("long_no_escape", func(b *testing.B) {
		s := strings.Repeat("hello world ", 100)
		for i := 0; i < b.N; i++ {
			AppendString(ctx, buf[:0], s)
		}
	})

	b.Run("with_escapes", func(b *testing.B) {
		s := `hello "world" with \backslash and "quotes"`
		for i := 0; i < b.N; i++ {
			AppendString(ctx, buf[:0], s)
		}
	})

	b.Run("html_entities", func(b *testing.B) {
		htmlOpt := &Option{Flag: HTMLEscapeOption}
		htmlCtx := &RuntimeContext{Option: htmlOpt}
		s := `<div class="test">hello & world</div>`
		for i := 0; i < b.N; i++ {
			AppendString(htmlCtx, buf[:0], s)
		}
	})

	b.Run("unicode", func(b *testing.B) {
		s := "こんにちは世界 hello world"
		for i := 0; i < b.N; i++ {
			AppendString(ctx, buf[:0], s)
		}
	})

	b.Run("unicode_long", func(b *testing.B) {
		s := strings.Repeat("こんにちは世界 ", 100)
		for i := 0; i < b.N; i++ {
			AppendString(ctx, buf[:0], s)
		}
	})

	b.Run("unicode_with_sparse_escapes", func(b *testing.B) {
		// Unicode string with occasional escapes
		s := strings.Repeat("こんにちは世界\"hello\"", 50)
		for i := 0; i < b.N; i++ {
			AppendString(ctx, buf[:0], s)
		}
	})

	b.Run("ascii_long_sparse_escapes", func(b *testing.B) {
		// Long ASCII string with escapes every ~100 chars
		s := strings.Repeat(strings.Repeat("a", 99)+"\"", 10)
		for i := 0; i < b.N; i++ {
			AppendString(ctx, buf[:0], s)
		}
	})
}

// Compare old vs new scanning approaches for Unicode strings
func BenchmarkScanUnicodeComparison(b *testing.B) {
	// This benchmark demonstrates the key improvement:
	// scanEscapeBasic stops at every non-ASCII byte (false positives)
	// scanEscapeBasicASCIIOnly skips non-ASCII bytes entirely
	s := strings.Repeat("\xe4\xb8\x96\xe7\x95\x8c", 170) // ~1020 bytes of Unicode
	p := stringptr(s)
	n := len(s)

	b.Run("Basic_unicode_1020", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scanEscapeBasic(p, n)
		}
	})

	b.Run("ASCIIOnly_unicode_1020", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scanEscapeBasicASCIIOnly(p, n)
		}
	})
}

func sprintf(format string, args ...interface{}) string {
	// Simple sprintf without importing fmt
	if len(args) == 1 {
		if v, ok := args[0].(int); ok {
			s := ""
			if v == 0 {
				return strings.Replace(format, "%d", "0", 1)
			}
			for v > 0 {
				s = string(rune('0'+v%10)) + s
				v /= 10
			}
			return strings.Replace(format, "%d", s, 1)
		}
	}
	return format
}

package benchmark

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkZZMultibyte encodes a struct with a string of Japanese text, of several lengths and mixed with ASCII,
// by go-json and sonic in each report configuration.
func BenchmarkZZMultibyte(b *testing.B) {
	type doc struct {
		Text string `json:"text"`
	}
	ja := strings.Repeat("日本語の文章を書き出します。", 100)
	mixed := strings.Repeat("東京 Tokyo 2026年9月30日, weather: 晴れ; ", 40)
	cases := []struct {
		name string
		s    string
	}{
		{"ja3", string([]rune(ja)[:3])},
		{"ja10", string([]rune(ja)[:10])},
		{"ja30", string([]rune(ja)[:30])},
		{"ja100", string([]rune(ja)[:100])},
		{"ja1000", string([]rune(ja)[:1000])},
		{"mixed40", string([]rune(mixed)[:40])},
		{"mixed400", string([]rune(mixed)[:400])},
		{"ascii300q", strings.Repeat(`say "hi" <b>&</b> `, 20)[:300]},
	}
	pretouchSonic()
	for _, cat := range []string{"std", "fast", "fastest"} {
		for _, lib := range []string{"go-json", "sonic"} {
			var c *reportConfig
			for _, x := range reportConfigs {
				if x.ID == cat+"/"+lib {
					c = x
				}
			}
			for _, tc := range cases {
				v := &doc{Text: tc.s}
				b.Run(fmt.Sprintf("%s/%s/%s", cat, tc.name, lib), func(b *testing.B) {
					if _, err := c.marshal(v); err != nil {
						b.Fatal(err)
					}
					b.SetBytes(int64(len(tc.s)))
					for i := 0; i < b.N; i++ {
						if _, err := c.marshal(v); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stdtest_test

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	. "github.com/goccy/go-json/jsontext"
)

func TestPointer(t *testing.T) {
	tests := []struct {
		in         Pointer
		wantParent Pointer
		wantLast   string
		wantTokens []string
		wantValid  bool
	}{
		{"", "", "", nil, true},
		{"a", "", "a", []string{"a"}, false},
		{"~", "", "~", []string{"~"}, false},
		{"/a", "", "a", []string{"a"}, true},
		{"/foo/bar", "/foo", "bar", []string{"foo", "bar"}, true},
		{"///", "//", "", []string{"", "", ""}, true},
		{"/~0~1", "", "~/", []string{"~/"}, true},
		{"/\xde\xad\xbe\xef", "", "\xde\xad\xbe\xef", []string{"\xde\xad\xbe\xef"}, false},
	}
	for _, tt := range tests {
		if got := tt.in.Parent(); got != tt.wantParent {
			t.Errorf("Pointer(%q).Parent = %q, want %q", tt.in, got, tt.wantParent)
		}
		if got := tt.in.LastToken(); got != tt.wantLast {
			t.Errorf("Pointer(%q).Last = %q, want %q", tt.in, got, tt.wantLast)
		}
		if strings.HasPrefix(string(tt.in), "/") {
			wantRoundtrip := tt.in
			if !utf8.ValidString(string(wantRoundtrip)) {
				// Replace bytes of invalid UTF-8 with Unicode replacement character.
				wantRoundtrip = Pointer([]rune(wantRoundtrip))
			}
			if got := tt.in.Parent().AppendToken(tt.in.LastToken()); got != wantRoundtrip {
				t.Errorf("Pointer(%q).Parent().AppendToken(LastToken()) = %q, want %q", tt.in, got, tt.in)
			}
			in := tt.in
			for {
				if (in + "x").Contains(tt.in) {
					t.Errorf("Pointer(%q).Contains(%q) = true, want false", in+"x", tt.in)
				}
				if !in.Contains(tt.in) {
					t.Errorf("Pointer(%q).Contains(%q) = false, want true", in, tt.in)
				}
				if in == in.Parent() {
					break
				}
				in = in.Parent()
			}
		}
		if got := slices.Collect(tt.in.Tokens()); !slices.Equal(got, tt.wantTokens) {
			t.Errorf("Pointer(%q).Tokens = %q, want %q", tt.in, got, tt.wantTokens)
		}
		if got := tt.in.IsValid(); got != tt.wantValid {
			t.Errorf("Pointer(%q).IsValid = %v, want %v", tt.in, got, tt.wantValid)
		}
	}
}

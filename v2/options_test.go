package json_test

import (
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The options of this package and of jsontext are of one type: they may be joined, and each property is read by
// its setter.
func TestGetOption(t *testing.T) {
	opts := json.JoinOptions(
		json.Deterministic(true),
		jsontext.Multiline(true),
		json.FormatNilSliceAsNull(false),
		jsontext.WithIndent("  "),
		json.Deterministic(false),
		json.StringifyNumbers(true),
	)
	for _, c := range []struct {
		name    string
		get     func() (any, bool)
		want    any
		wantSet bool
	}{
		{"Deterministic", func() (any, bool) { return json.GetOption(opts, json.Deterministic) }, false, true},
		{"StringifyNumbers", func() (any, bool) { return json.GetOption(opts, json.StringifyNumbers) }, true, true},
		{"FormatNilSliceAsNull", func() (any, bool) { return json.GetOption(opts, json.FormatNilSliceAsNull) }, false, true},
		{"FormatNilMapAsNull", func() (any, bool) { return json.GetOption(opts, json.FormatNilMapAsNull) }, false, false},
		{"Multiline", func() (any, bool) { return json.GetOption(opts, jsontext.Multiline) }, true, true},
		{"WithIndent", func() (any, bool) { return json.GetOption(opts, jsontext.WithIndent) }, "  ", true},
		{"WithIndentPrefix", func() (any, bool) { return json.GetOption(opts, jsontext.WithIndentPrefix) }, "", false},
		{"AllowInvalidUTF8", func() (any, bool) { return json.GetOption(opts, jsontext.AllowInvalidUTF8) }, false, false},
	} {
		got, set := c.get()
		if got != c.want || set != c.wantSet {
			t.Errorf("GetOption(%s) = %v, %v; want %v, %v", c.name, got, set, c.want, c.wantSet)
		}
	}
	if v, ok := json.GetOption(nil, json.Deterministic); v || ok {
		t.Errorf("GetOption(nil) = %v, %v", v, ok)
	}
}

// DefaultOptionsV2 sets the options of the v1 semantics, false.
func TestDefaultOptionsV2(t *testing.T) {
	opts := json.DefaultOptionsV2()
	for name, setter := range map[string]func(bool) json.Options{
		"AllowDuplicateNames":       jsontext.AllowDuplicateNames,
		"AllowInvalidUTF8":          jsontext.AllowInvalidUTF8,
		"EscapeForHTML":             jsontext.EscapeForHTML,
		"EscapeForJS":               jsontext.EscapeForJS,
		"PreserveRawStrings":        jsontext.PreserveRawStrings,
		"Deterministic":             json.Deterministic,
		"FormatNilMapAsNull":        json.FormatNilMapAsNull,
		"FormatNilSliceAsNull":      json.FormatNilSliceAsNull,
		"MatchCaseInsensitiveNames": json.MatchCaseInsensitiveNames,
	} {
		if v, ok := json.GetOption(opts, setter); v || !ok {
			t.Errorf("GetOption(DefaultOptionsV2(), %s) = %v, %v; want false, true", name, v, ok)
		}
	}
	if _, ok := json.GetOption(opts, json.StringifyNumbers); ok {
		t.Error("DefaultOptionsV2 sets StringifyNumbers")
	}
}

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

// An option of a property which is true or false is made without an allocation, as the constants which
// encoding/json/v2 returns are: an option made for each call costs nothing more. It sets its property.
func TestBoolOptions(t *testing.T) {
	for _, c := range []struct {
		name string
		of   func(bool) json.Options
	}{
		{"StringifyNumbers", json.StringifyNumbers},
		{"Deterministic", json.Deterministic},
		{"FormatNilSliceAsNull", json.FormatNilSliceAsNull},
		{"FormatNilMapAsNull", json.FormatNilMapAsNull},
		{"OmitZeroStructFields", json.OmitZeroStructFields},
		{"MatchCaseInsensitiveNames", json.MatchCaseInsensitiveNames},
		{"RejectUnknownMembers", json.RejectUnknownMembers},
		{"AllowDuplicateNames", jsontext.AllowDuplicateNames},
		{"AllowInvalidUTF8", jsontext.AllowInvalidUTF8},
		{"EscapeForHTML", jsontext.EscapeForHTML},
		{"EscapeForJS", jsontext.EscapeForJS},
		{"PreserveRawStrings", jsontext.PreserveRawStrings},
		{"CanonicalizeRawInts", jsontext.CanonicalizeRawInts},
		{"CanonicalizeRawFloats", jsontext.CanonicalizeRawFloats},
		{"ReorderRawObjects", jsontext.ReorderRawObjects},
		{"SpaceAfterColon", jsontext.SpaceAfterColon},
		{"SpaceAfterComma", jsontext.SpaceAfterComma},
		{"Multiline", jsontext.Multiline},
	} {
		for _, v := range []bool{true, false} {
			if allocs := testing.AllocsPerRun(10, func() { _ = c.of(v) }); allocs != 0 {
				t.Errorf("%s(%v): %v allocations", c.name, v, allocs)
			}
			if got, set := json.GetOption(c.of(v), c.of); got != v || !set {
				t.Errorf("GetOption(%s(%v)) = %v, %v", c.name, v, got, set)
			}
			if got, set := json.GetOption(json.JoinOptions(c.of(!v), json.Deterministic(true), c.of(v)), c.of); got != v || !set {
				t.Errorf("GetOption of %s(%v) after %v = %v, %v", c.name, v, !v, got, set)
			}
		}
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

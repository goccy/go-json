package json_test

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	jsonv2 "github.com/goccy/go-json/v2"
)

// legacyOptions are the options of the semantics of this package, by name.
var legacyOptions = []struct {
	name string
	of   func(bool) json.Options
}{
	{"CallMethodsWithLegacySemantics", json.CallMethodsWithLegacySemantics},
	{"FormatByteArrayAsArray", json.FormatByteArrayAsArray},
	{"FormatBytesWithLegacySemantics", json.FormatBytesWithLegacySemantics},
	{"FormatDurationAsNano", json.FormatDurationAsNano},
	{"MatchCaseSensitiveDelimiter", json.MatchCaseSensitiveDelimiter},
	{"MergeWithLegacySemantics", json.MergeWithLegacySemantics},
	{"OmitEmptyWithLegacySemantics", json.OmitEmptyWithLegacySemantics},
	{"ParseBytesWithLooseRFC4648", json.ParseBytesWithLooseRFC4648},
	{"ParseTimeWithLooseRFC3339", json.ParseTimeWithLooseRFC3339},
	{"ReportErrorsWithLegacySemantics", json.ReportErrorsWithLegacySemantics},
	{"StringifyWithLegacySemantics", json.StringifyWithLegacySemantics},
	{"UnmarshalArrayFromAnyLength", json.UnmarshalArrayFromAnyLength},
}

// DefaultOptionsV1 sets every option of the semantics of this package to true and DefaultOptionsV2 to false; an
// option given after them overrides them, and GetOption tells each, as for the options of the v2 json package.
func TestLegacyOptions(t *testing.T) {
	for _, o := range legacyOptions {
		if v, ok := jsonv2.GetOption(json.DefaultOptionsV1(), o.of); !v || !ok {
			t.Errorf("%s of DefaultOptionsV1: got %v, %v, want true, true", o.name, v, ok)
		}
		if v, ok := jsonv2.GetOption(jsonv2.DefaultOptionsV2(), o.of); v || !ok {
			t.Errorf("%s of DefaultOptionsV2: got %v, %v, want false, true", o.name, v, ok)
		}
		if v, ok := jsonv2.GetOption(jsonv2.JoinOptions(json.DefaultOptionsV1(), o.of(false)), o.of); v || !ok {
			t.Errorf("%s set false after DefaultOptionsV1: got %v, %v, want false, true", o.name, v, ok)
		}
		if v, ok := jsonv2.GetOption(jsonv2.JoinOptions(json.DefaultOptionsV1(), jsonv2.DefaultOptionsV2()), o.of); v || !ok {
			t.Errorf("%s of DefaultOptionsV2 after DefaultOptionsV1: got %v, %v, want false, true", o.name, v, ok)
		}
		if v, ok := jsonv2.GetOption(o.of(true), o.of); !v || !ok {
			t.Errorf("%s(true): got %v, %v, want true, true", o.name, v, ok)
		}
		if _, ok := jsonv2.GetOption(jsonv2.Deterministic(true), o.of); ok {
			t.Errorf("%s of Deterministic(true): set", o.name)
		}
		for _, other := range legacyOptions {
			if v, ok := jsonv2.GetOption(o.of(true), other.of); other.name != o.name && (v || ok) {
				t.Errorf("%s of %s(true): got %v, %v, want false, false", other.name, o.name, v, ok)
			}
		}
	}
	// the options of the v2 json package and jsontext which DefaultOptionsV1 sets
	for _, o := range []struct {
		name string
		of   func(bool) json.Options
	}{
		{"Deterministic", jsonv2.Deterministic},
		{"FormatNilMapAsNull", jsonv2.FormatNilMapAsNull},
		{"FormatNilSliceAsNull", jsonv2.FormatNilSliceAsNull},
		{"MatchCaseInsensitiveNames", jsonv2.MatchCaseInsensitiveNames},
		{"AllowDuplicateNames", jsontext.AllowDuplicateNames},
		{"AllowInvalidUTF8", jsontext.AllowInvalidUTF8},
		{"EscapeForHTML", jsontext.EscapeForHTML},
		{"EscapeForJS", jsontext.EscapeForJS},
		{"PreserveRawStrings", jsontext.PreserveRawStrings},
	} {
		if v, ok := jsonv2.GetOption(json.DefaultOptionsV1(), o.of); !v || !ok {
			t.Errorf("%s of DefaultOptionsV1: got %v, %v, want true, true", o.name, v, ok)
		}
	}
	if v, ok := jsonv2.GetOption(json.DefaultOptionsV1(), jsonv2.StringifyNumbers); v || ok {
		t.Errorf("StringifyNumbers of DefaultOptionsV1: got %v, %v, want false, false", v, ok)
	}
}

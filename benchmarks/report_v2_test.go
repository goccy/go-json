//go:build go1.27 && goexperiment.jsonv2

package benchmark

import (
	stdjson "encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"

	"benchmark/report"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
	segmentio "github.com/segmentio/encoding/json"
)

// The configurations of encoding/json/v2, where Go has it, and the category of its behavior.
//
// encoding/json/v2 is in every category: with the options of the behavior of encoding/json ( DefaultOptionsV1 ) in
// "std", with those and the ones which drop HTML escaping and key sorting in "fast", and with its default options,
// the baseline of "v2". There, the other libraries are configured as close to it as their options allow, and
// encoding/json/v2 is also shown at its fastest, with the checks of duplicate keys and of UTF-8 turned off.

func v2Marshal(opts ...jsonv2.Options) func(any) ([]byte, error) {
	return func(v any) ([]byte, error) { return jsonv2.Marshal(v, opts...) }
}

func v2Unmarshal(opts ...jsonv2.Options) func([]byte) func(any) error {
	return stdUnmarshal(func(data []byte, v any) error { return jsonv2.Unmarshal(data, v, opts...) })
}

var (
	sonicV2    = sonic.Config{CompactMarshaler: true, CopyString: true, ValidateString: true, NoNullSliceOrMap: true, CaseSensitive: true, UseUnicodeErrors: true}.Froze()
	jsoniterV2 = jsoniter.Config{EscapeHTML: false, SortMapKeys: false, CaseSensitive: true, ValidateJsonRawMessage: true}.Froze()
)

func init() {
	v1 := stdjson.DefaultOptionsV1()
	std := &reportConfig{
		Config: report.Config{ID: "std/encoding/json/v2", Library: "encoding/json/v2", Category: "std",
			Title: "encoding/json/v2", Setting: "encoding/json/v2 with json.DefaultOptionsV1()"},
		module: "std", marshal: v2Marshal(v1), unmarshal: v2Unmarshal(v1),
	}
	fast := &reportConfig{
		Config: report.Config{ID: "fast/encoding/json/v2", Library: "encoding/json/v2", Category: "fast",
			Title: "encoding/json/v2", Setting: "encoding/json/v2 with json.DefaultOptionsV1(), jsontext.EscapeForHTML( false ), json.Deterministic( false )"},
		module:    "std",
		marshal:   v2Marshal(v1, jsontext.EscapeForHTML(false), jsonv2.Deterministic(false)),
		unmarshal: v2Unmarshal(v1),
	}
	fastest := jsonv2.JoinOptions(jsontext.AllowDuplicateNames(true), jsontext.AllowInvalidUTF8(true))
	v2 := []*reportConfig{
		{
			Config: report.Config{ID: "v2/encoding/json/v2", Library: "encoding/json/v2", Category: "v2", Title: "encoding/json/v2",
				Setting: "json.Marshal / json.Unmarshal of encoding/json/v2"},
			module: "std", marshal: v2Marshal(), unmarshal: v2Unmarshal(),
		},
		{
			Config: report.Config{ID: "v2/encoding/json/v2/fastest", Library: "encoding/json/v2", Category: "v2", Title: "encoding/json/v2 ( fastest )",
				Setting: "encoding/json/v2 with jsontext.AllowDuplicateNames( true ), jsontext.AllowInvalidUTF8( true )"},
			module: "std", marshal: v2Marshal(fastest), unmarshal: v2Unmarshal(fastest),
		},
		{
			Config: report.Config{ID: "v2/encoding/json", Library: "encoding/json", Category: "v2", Title: "encoding/json",
				Setting: "json.Encoder with SetEscapeHTML( false ) / json.Unmarshal ( no option for the rest )"},
			module: "std", marshal: stdMarshalNoHTMLEscape, unmarshal: stdUnmarshal(stdjson.Unmarshal),
		},
		{
			Config: report.Config{ID: "v2/go-json", Library: "goccy/go-json", Category: "v2", Title: "goccy/go-json",
				Setting: "json.MarshalWithOption( DisableHTMLEscape, UnorderedMap ) / json.Unmarshal ( no option for the rest )"},
			module: goJSONModule, marshal: goJSONMarshal(gojson.DisableHTMLEscape(), gojson.UnorderedMap()), unmarshal: goJSONUnmarshal(),
		},
		{
			Config: report.Config{ID: "v2/go-json/of", Library: "goccy/go-json", Category: "v2", Title: "goccy/go-json ( UnmarshalOf )",
				Setting: "json.UnmarshalOf, which takes the value by its type ( decode only )"},
			module: goJSONModule, decodeOf: []gojson.DecodeOptionFunc{},
		},
		{
			Config: report.Config{ID: "v2/sonic", Library: "bytedance/sonic", Category: "v2", Title: "bytedance/sonic",
				Setting: "sonic.Config{ CompactMarshaler, CopyString, ValidateString, NoNullSliceOrMap, CaseSensitive, UseUnicodeErrors: true }"},
			module: sonicModule, marshal: sonicV2.Marshal, unmarshal: stdUnmarshal(sonicV2.Unmarshal),
		},
		{
			Config: report.Config{ID: "v2/jsoniter", Library: "json-iterator/go", Category: "v2", Title: "json-iterator/go",
				Setting: "jsoniter.Config{ EscapeHTML: false, SortMapKeys: false, CaseSensitive: true, ValidateJsonRawMessage: true }"},
			module: jsoniterModule, marshal: jsoniterV2.Marshal, unmarshal: stdUnmarshal(jsoniterV2.Unmarshal),
		},
		{
			Config: report.Config{ID: "v2/segmentio", Library: "segmentio/encoding", Category: "v2", Title: "segmentio/encoding",
				Setting: "json.Append( 0 ) / json.Parse( DontMatchCaseInsensitiveStructFields )"},
			module: segmentioModule, marshal: segmentioMarshal(0), unmarshal: segmentioUnmarshal(segmentio.DontMatchCaseInsensitiveStructFields),
		},
	}
	insertAfter("std/encoding/json", std)
	insertAfter("fast/encoding/json", fast)
	insertAfter("fastest/encoding/json", &reportConfig{
		Config: report.Config{ID: "fastest/encoding/json/v2", Library: "encoding/json/v2", Category: "fastest", Title: "encoding/json/v2",
			Setting: "encoding/json/v2 with jsontext.AllowDuplicateNames( true ), jsontext.AllowInvalidUTF8( true )"},
		module: "std", marshal: v2Marshal(fastest), unmarshal: v2Unmarshal(fastest),
	})
	reportConfigs = append(reportConfigs, v2...)
	// after the category of encoding/json
	reportCategories = append(reportCategories[:1], append([]report.Category{{
		ID:    "v2",
		Title: "Same behavior as encoding/json/v2",
		Description: "The default behavior of encoding/json/v2: HTML is not escaped, the keys of a map are in any order, " +
			"a nil slice or map is encoded as [] or {}, invalid UTF-8 and duplicate keys are errors, and the keys match the " +
			"fields by their case. The other libraries are configured as close to it as their options allow, and " +
			"encoding/json/v2 is also shown at its fastest, without the checks of duplicate keys and of UTF-8.",
		Baseline: "v2/encoding/json/v2",
	}}, reportCategories[1:]...)...)
}

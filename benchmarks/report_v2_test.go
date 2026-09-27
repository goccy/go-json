//go:build go1.27 && goexperiment.jsonv2

package benchmark

import (
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"

	"benchmark/report"
)

// The configurations of encoding/json/v2, where the Go version has it: with the options of the behavior of
// encoding/json ( DefaultOptionsV1 ), it is compared with the other libraries in "std"; with its default options,
// it behaves differently from all of them, and is shown alone in "v2".
func init() {
	v1 := jsonv1.DefaultOptionsV1()
	std := &reportConfig{
		Config: report.Config{ID: "encoding/json/v2+v1", Library: "encoding/json/v2", Category: "std",
			Title: "encoding/json/v2 ( DefaultOptionsV1 )", Setting: "json.Marshal / json.Unmarshal of encoding/json/v2 with json.DefaultOptionsV1()"},
		module:  "std",
		marshal: func(v any) ([]byte, error) { return jsonv2.Marshal(v, v1) },
		unmarshal: stdUnmarshal(func(data []byte, v any) error {
			return jsonv2.Unmarshal(data, v, v1)
		}),
	}
	v2 := &reportConfig{
		Config: report.Config{ID: "encoding/json/v2", Library: "encoding/json/v2", Category: "v2",
			Title: "encoding/json/v2", Setting: "json.Marshal / json.Unmarshal of encoding/json/v2"},
		module:    "std",
		marshal:   func(v any) ([]byte, error) { return jsonv2.Marshal(v) },
		unmarshal: stdUnmarshal(func(data []byte, v any) error { return jsonv2.Unmarshal(data, v) }),
	}
	// after encoding/json, the baseline of "std"
	reportConfigs = append(reportConfigs[:1], append([]*reportConfig{std}, reportConfigs[1:]...)...)
	reportConfigs = append(reportConfigs, v2)
}

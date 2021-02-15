package json_test

import (
	"bytes"
	stdjson "encoding/json"
)

func intptr(v int) *int             { return &v }
func int8ptr(v int8) *int8          { return &v }
func int16ptr(v int16) *int16       { return &v }
func int32ptr(v int32) *int32       { return &v }
func int64ptr(v int64) *int64       { return &v }
func uptr(v uint) *uint             { return &v }
func uint8ptr(v uint8) *uint8       { return &v }
func uint16ptr(v uint16) *uint16    { return &v }
func uint32ptr(v uint32) *uint32    { return &v }
func uint64ptr(v uint64) *uint64    { return &v }
func float32ptr(v float32) *float32 { return &v }
func float64ptr(v float64) *float64 { return &v }
func stringptr(v string) *string    { return &v }
func boolptr(v bool) *bool          { return &v }

func encodeByEncodingJSON(data interface{}, indent, escape bool) string {
	var buf bytes.Buffer
	enc := stdjson.NewEncoder(&buf)
	enc.SetEscapeHTML(escape)
	if indent {
		enc.SetIndent("", "  ")
	}
	enc.Encode(data)
	return buf.String()
}

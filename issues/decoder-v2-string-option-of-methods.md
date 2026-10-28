# v2: the `string` option of a field is not given to the decoding within its method

- Area: decoder (v2)
- Kind: defect
- Priority: 3
- Status: recorded 2026-10-28 on feat/v1-options (the same on master), found by the review of
  `FormatDurationAsNano`

## What

encoding/json/v2 documents that the `string` option "specifies that StringifyNumbers be set when marshaling or
unmarshaling a struct field value": a field whose type has an `UnmarshalJSONFrom` method, which decodes into a number
by `UnmarshalDecode`, reads the number from a JSON string. The decoder of go-json gives the method no such option:

```go
type W struct{ N int64 }

func (w *W) UnmarshalJSONFrom(dec *jsontext.Decoder) error { return json.UnmarshalDecode(dec, &w.N) }

var v struct {
	X W `json:",string"`
}
json.Unmarshal([]byte(`{"X":"6"}`), &v) // encoding/json/v2: 6; go-json: cannot unmarshal JSON string into int64
json.Unmarshal([]byte(`{"X":6}`), &v)   // encoding/json/v2: an error; go-json: 6
```

The encoder gives the option to `MarshalJSONTo` as encoding/json/v2 does.

A function of `UnmarshalFromFunc` which decodes by `UnmarshalDecode` is given no option either ( `{"F":"1"}` into
`struct{ F *int "json:\",string\"" }` with a function of `*int` which decodes into its value: encoding/json/v2 1,
go-json the error of a JSON string; found in the review of `ReportErrorsWithLegacySemantics`, 2026-10-28 ).

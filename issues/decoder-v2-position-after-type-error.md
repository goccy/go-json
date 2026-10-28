# v2: where a decoder goes on after a value of a type with no representation

- Area: decoder (v2)
- Kind: defect
- Priority: 4
- Status: recorded 2026-10-28 on feat/v1-options (the same on master), found by the review of
  `FormatDurationAsNano`

## What

`UnmarshalDecode` into a value of a type which has no representation, as a `time.Duration` without
`FormatDurationAsNano`, fails before it reads the value: encoding/json/v2 leaves the decoder before the value, which
the next call reads, as the place where a decoder goes on after a failed `UnmarshalDecode` was decided to be the one
of encoding/json/v2; go-json skips the value:

```go
dec := jsontext.NewDecoder(strings.NewReader("1 2"))
json.UnmarshalDecode(dec, new(time.Duration)) // fails in both
var n int
json.UnmarshalDecode(dec, &n) // encoding/json/v2: 1; go-json: 2
```

With `ReportErrorsWithLegacySemantics`, encoding/json/v2 skips the value too.

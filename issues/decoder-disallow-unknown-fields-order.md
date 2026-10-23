# DisallowUnknownFields returns the unknown field instead of an earlier error

- Area: decoder
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-11; checked on master on 2026-10-23

## What

With `Decoder.DisallowUnknownFields`, when an input has a type error (or a bad map key) and
after it an unknown field, go-json stops at the unknown field and returns its error.
encoding/json keeps the first error in input order, reads on, and returns that one: `{"A":"x","Z":1}`
into `struct{ A int }` is the type error of `A` in encoding/json and `unknown field "Z"` in go-json.

## Note

The v2 option `RejectUnknownMembers` decides the same thing; the v2 struct decoder
(`internal/decoder/decode_v2_struct.go`) stops at the first error, as encoding/json/v2 does.

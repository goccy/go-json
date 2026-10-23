# Errors of map keys of pointer types

- Area: decoder
- Kind: defect
- Priority: 3
- Status: recorded 2026-10-13; checked on master on 2026-10-23

## What

For a map whose key is a pointer to a kind that can't be a key, the `Type` of the
`UnmarshalTypeError` is the pointer type in go-json, the type pointed to in Go 1.27:
`{"a":1}` into `map[*bool]int` names `*bool` in go-json and `bool` in encoding/json.

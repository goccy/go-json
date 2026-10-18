# float32 is rounded twice

- Area: decoder
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-19 (found by the review of the first v2 decoding pull request); re-check on master
  before fixing

## What

A number decoded into a float32 is parsed as a float64 and converted to float32, so it is rounded twice, and the
range is checked on the float64 (`internal/decoder/float.go`: `overflowsFloat32`, and `float32(v)` in
`compileFloat32`):

- `3.4028235e38`, the largest float32 written as Go writes it, fails as out of range; encoding/json and
  encoding/json/v2 decode it.
- `1.0000000596046447753906250000000001` becomes 1; encoding/json/v2 gives 1.0000001.

The root package (v1) and `Unmarshal` of `github.com/goccy/go-json/v2` both have it. With `StringifyNumbers`, the
v2 decoding parses with 32 bits and is right, so a quoted number and a plain one differ.

## Fix

Parse with 32 bits for a float32, or take the fast parse only when its conversion to float32 is exact. The float
decoder is on the hot path of v1: measure every decode benchmark of floats.

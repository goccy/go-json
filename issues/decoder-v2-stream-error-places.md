# v2: places of syntax errors in streams which differ from encoding/json/v2

- Area: decoder (v2)
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-23 on master, found by the review of the lookup of struct member names

## What

The `ByteOffset` or the `JSONPointer` of a `jsontext.SyntacticError`, which are documented, differ
from encoding/json/v2's, for every size of the reads of the reader (1, 7 and 512 bytes):

- `UnmarshalDecode` of an object cut in an array, after the `[` was read with `ReadToken`:
  `[{"x":300,{"x":300]` into `struct{ X int }` is at 10 in go-json and at 18 in encoding/json/v2
  (the pointer is `/0` in both).
- `UnmarshalRead` of an empty name of a member which an inlined map fallback takes:
  `{"a":1,"":nul` into `struct{ A int; Rest map[string]any \`json:",inline"\` }` has the pointer
  `""` in go-json and `"/"` in encoding/json/v2 (the offset is 13 in both). `Unmarshal` gives
  `"/"`.

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
- `UnmarshalRead` from a reader of one byte at a time of a syntax error in the value of a member of an empty name
  of no field: `{"":[ fals, true],"p":1}` into `struct{ A int }` has the pointer `/0` in go-json and `//0` in
  encoding/json/v2 (the offset is 10 in both); a reader of the whole input gives `//0` in both. Found on master
  on 2026-10-24 by the review of the skip of the values of no field.
- The place of the error of a call around a method which decodes with other options: an `UnmarshalJSONFrom`
  method which calls `UnmarshalDecode` with `AllowDuplicateNames(true)`, in a struct decoded without it:
  `{"Z":{"a":1},"A":{"w":{"b":1,"b":2}},"V":{"c":1,"c":2}}`, where `A` is of the method, is an error of the name
  `"b"` within `/A/w` in go-json, which the method allows, and of `"c"` within `/V` in encoding/json/v2. The input
  is read again for the place of the error by the options of the call around the method. Found on master on
  2026-10-24 by the tests of the skippers of the values of no field.

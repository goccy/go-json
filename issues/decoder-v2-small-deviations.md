# Small deviations of the v2 decoder

- Area: decoder (v2)
- Kind: deviation
- Priority: 4
- Status: recorded 2026-10-20, found in the review of `UnmarshalJSONFrom`; left by the rule of
  matching the documented behavior only

## What

- A method of `UnmarshalJSONFrom` or a function of `UnmarshalFromFunc` which reads a token,
  calls `PeekKind` and then returns an error: encoding/json/v2 reports the error at the offset
  of the next value, as `PeekKind` drops the token it read last; go-json at the token it read.
  The pointer is the same. Example: `[ {} ]` into `[]T`, a method which reads `{`, peeks and
  fails: offset 3 in encoding/json/v2, 2 in go-json.
- A method which reads past the end of the object or array its value is in: encoding/json/v2
  reads it and reports that the method didn't read exactly one value (a `SemanticError`); the
  decoder which go-json gives the method can't read it and returns a `SyntacticError` from the
  read, as the encoder given to `MarshalJSONTo` can't write it.

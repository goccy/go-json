# Small deviations of the v2 decoder

- Area: decoder (v2)
- Kind: deviation
- Priority: 4
- Status: recorded 2026-10-20, found in the review of `UnmarshalJSONFrom`, and 2026-10-23 in the
  reviews of `UnmarshalRead` and `UnmarshalDecode`; left by the rule of matching the documented
  behavior only

## What

- A method of `UnmarshalJSONFrom` or a function of `UnmarshalFromFunc` which reads a token,
  calls `PeekKind` and then returns an error: encoding/json/v2 reports the error at the offset
  of the next value, as `PeekKind` drops the token it read last; go-json at the token it read.
  The pointer is the same. Example: `[ {} ]` into `[]T`, a method which reads `{`, peeks and
  fails: offset 3 in encoding/json/v2, 2 in go-json. The same after a value which the method
  read by `UnmarshalDecode`, where go-json holds the whole input ( `Unmarshal`, and
  `UnmarshalDecode` of a decoder whose first read holds it ): `{"k":"x..."}`, a function which
  reads the string, peeks and fails, is at the end of the string in encoding/json/v2, and at its
  start in go-json.
- A method or a function which reads its whole value by `UnmarshalDecode` of its decoder and
  then returns an error: go-json reports the error at the start of the value; encoding/json/v2
  at the last token which its decoding read, the start of a value it reads whole ( a string, a
  number, a literal, a `jsontext.Value` ), and the closing `]` or `}` of an object or an array
  it decodes token by token ( into a slice, a struct, a map or an `interface{}` ). Example:
  `[{"A":1}]`, a function of `*[]any` which decodes its value into `[]struct{ A int }` and
  fails: offset 8 in encoding/json/v2, 0 in go-json. Both are offsets at or before the error,
  as `ByteOffset` is documented. The pointer is the same.
- A method which reads past the end of the object or array its value is in: encoding/json/v2
  reads it and reports that the method didn't read exactly one value (a `SemanticError`); the
  decoder which go-json gives the method can't read it and returns a `SyntacticError` from the
  read, as the encoder given to `MarshalJSONTo` can't write it.
- A value which is not a string where an object has a member name, as `{[` or `{{`: go-json
  reports that a member name must be a string, at the value; encoding/json/v2 reads the name as
  a whole value first, and reports what it meets reading it (the unexpected end of `{[` at offset
  2). `Unmarshal` and `UnmarshalRead` do the same.
- The end of an object or an array where a value of a method or a function which reads a
  `jsontext.Decoder` is, as `{"K":]}` into a map of such values: the decoder which go-json gives
  the method can't read the end of the level of its value, and reports it (`cannot read the end
  of the object or array which the value being read is in`, at `/K`); encoding/json/v2 reports
  the invalid character (`invalid character ']' after object value`, at the offset, pointer
  `""`). The offset is the same.
- `UnmarshalDecode` at the name of a member, into a type of an `UnmarshalJSONFrom` method or of a
  function of `UnmarshalFromFunc`: the decoder which go-json gives the method reads the name alone,
  as a value at the top level, which the decoder of the caller read with `ReadValue`, so that
  `InputOffset`, `StackPointer` and `StackDepth` are the ones of the name alone (0, `""` and 0
  before it, for `{"abc":1}`); encoding/json/v2 gives the method the decoder of the caller (1,
  `""` and 1 before the name, 6, `/abc` and 1 after it). The value decoded and the place of a
  `SemanticError` are the same.
- The values left by an error (recorded 2026-10-27, found in the review of the decoding of floats
  out of range as Go 1.27.2; encoding/json/v2 documents none of them):
  - an element of a slice which is decoded again, within its capacity, keeps what it held where
    the decoding of a scalar into it fails ( `[1e400]` or `[1.5]` into `[]float64{3}` or
    `[]int{3}`: go-json `[3]`, encoding/json/v2 `[0]` ); encoding/json/v2 zeroes the element
    first. The decodes which succeed are the same.
  - a pointer which is set is set to nil where the decoding of the value it points to fails
    ( `1.5` into a `**int` whose `*int` is set, `{"P":1.5}` into `struct{ P *int }` whose `P` is
    set: go-json nil, encoding/json/v2 keeps the pointer ).

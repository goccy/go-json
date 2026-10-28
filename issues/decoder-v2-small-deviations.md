# Small deviations of the v2 decoder

- Area: decoder (v2)
- Kind: deviation
- Priority: 4
- Status: recorded 2026-10-20, found in the review of `UnmarshalJSONFrom`, and 2026-10-23 in the
  reviews of `UnmarshalRead` and `UnmarshalDecode`; left by the rule of matching the documented
  behavior only

## What

- An `encoding/json.Number` ( the `Number` of the root package ) is decoded by its kind, a string, which a JSON number
  is an error of: encoding/json/v2 decodes it by its `UnmarshalJSONFrom`, which takes a JSON number ( `{"N":5}` into
  `struct{ N json.Number }`; recorded 2026-10-28, found in the review of `MergeWithLegacySemantics` ). Its methods
  take the `jsontext` of the standard library, which the decoder of go-json gives no method.
- An `encoding/json.Number` of a field of the `string` option fails as a type which the option doesn't apply to:
  encoding/json/v2 reads the number in the JSON string ( `{"N":"1"}` into `struct{ N json.Number "json:\",string\"" }`
  is `"1"` in encoding/json/v2, the error of the option in go-json; recorded 2026-10-28, found in the review of
  `StringifyWithLegacySemantics` ).
- With `ReportErrorsWithLegacySemantics` and without `StringifyWithLegacySemantics`, a field of an interface type of
  the `string` option is decoded without the option: encoding/json decodes the value which the interface value holds
  by it ( a JSON string `"1"` into the `*int` which it holds; recorded 2026-10-28, found in the work on
  `ReportErrorsWithLegacySemantics` ). The decoders of the values which interface values hold are the ones of their
  types, which have no option of a field.
- With `ReportErrorsWithLegacySemantics` and `StringifyWithLegacySemantics`, the cause of the `UnmarshalTypeError` of a
  value of the `string` option whose JSON string holds no JSON string ( `{"S":"x","N":1}` into
  `struct{ S *string "json:\",string\""; N int }` ): its `SyntaxError` is at the offset 1 in go-json, 0 in
  encoding/json/v2, with the same message ( recorded 2026-10-28, found in the review of
  `ReportErrorsWithLegacySemantics` ).
- With `ReportErrorsWithLegacySemantics`, a member of no field of a struct whose embedded `jsontext.Value` holds a value
  which is not an object: encoding/json/v2 panics, as it skips the rest of the object where the value of the member is
  not read yet ( `SkipValueRemainder` of `newUnmarshalErrorAfterWithSkipping`, Go 1.27.3; `{"A":1,"b":2,"C":3}` into
  a struct of `A`, `C` and the embedded value `1` ); go-json reports the error of the member, as encoding/json/v2
  reports it without the option, and decodes on after its value ( recorded 2026-10-28, found in the work on
  `ReportErrorsWithLegacySemantics` ).
- With `StringifyWithLegacySemantics`, a quoted null of an escape into a pointer of the `string` option sets the
  pointer nil, as a quoted null does, which the documentation of the option makes a substitute for a JSON null:
  encoding/json/v2 tells a quoted null of a pointer by the raw value, and decodes the escape into the value which
  the pointer points to, as the quoted null of a value of no pointer ( `{"P":"nul\u006c"}` into
  `struct{ P *string "json:\",string\"" }` whose `P` is set: go-json nil, encoding/json/v2 the pointer to
  `""`; recorded 2026-10-28, found by the tests of `StringifyWithLegacySemantics` ).
- No value where a value of a type which has no representation is ( a `chan`, a `func`, a `complex128`, a
  `time.Duration` of no format ): go-json reports what it finds there, encoding/json/v2 the `SemanticError` of the
  type, which it reports before it reads anything. The input is empty or white space only ( `Unmarshal` and
  `UnmarshalRead` of `""` or `"  "` into a `*chan int`: the unexpected end of the input in go-json ), a delimiter or
  the end of an object or an array ( `,`, `:`, `]`, `}` at the top level of `UnmarshalDecode`, `[0,]` as an element:
  a syntax error ), or the end of the stream ( `UnmarshalDecode`: `io.EOF` ). Recorded 2026-10-28, found in the
  review of `StringifyWithLegacySemantics`. A value, whole or cut, is the error of the type in both.
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
    first. The decodes which succeed are the same. With `ReportErrorsWithLegacySemantics` and without
    `MergeWithLegacySemantics`, which decode on, the element is left in the result ( `[1,"x",3]` into
    `[]int{9,9,9,9}`: go-json `[1 9 3]`, encoding/json/v2 `[1 0 3]` ).
  - the element of a slice whose decoding fails is not in the slice of `UnmarshalRead` with small reads, which
    decodes the elements one after another ( `{"S":[1,"x"]}` into `struct{ S []*int }` in reads of 1 or 7 bytes:
    go-json `[&1]`, `Unmarshal` and encoding/json/v2 `[&1 &0]`; recorded 2026-10-28, found in the review of
    `ReportErrorsWithLegacySemantics` ).

# Errors of ReportErrorsWithLegacySemantics in calls of misuse

- Area: v2 (options of encoding/json)
- Kind: deviation
- Priority: 3
- Status: recorded 2026-10-28, found in the review of the options of encoding/json

## What

With `ReportErrorsWithLegacySemantics`, encoding/json/v2 reports the errors of encoding/json; go-json
reports the ones of encoding/json/v2 in three calls which are not of normal use:

- `UnmarshalDecode` of the name of an object (the decoder is at a name): `{"a":1}` after the `{`,
  into an `int`, is a `*json.UnmarshalTypeError` in encoding/json/v2 and a `SemanticError` in go-json;
  a name twice (`{"a":1,"a":2}` after three tokens, into a string) is a `*json.SyntaxError` and a
  `*jsontext.SyntacticError` (`decodeCall.decode` leaves names to `decodeName`).
- `MarshalEncode` of a value whose name the encoder has already written in the object
  (`{`, `"a"`, `1`, then `"a"`): a `*json.SyntaxError` in encoding/json/v2 and a
  `*jsontext.SyntacticError` in go-json, which converts only the errors of the writer of
  `out.WriteValue`.
- `UnmarshalDecodeOf` into a nil pointer: a `SemanticError` in go-json, where `UnmarshalDecode`,
  `UnmarshalOf` and `UnmarshalReadOf` report an `InvalidUnmarshalError` with the option, which may also
  be an option of the decoder itself (`Decoder.Options`).

## Fix

Convert the errors of these paths by the rules of `unmarshalWithV1Errors` and `v1MarshalError` when the
option is set: a few lines each.

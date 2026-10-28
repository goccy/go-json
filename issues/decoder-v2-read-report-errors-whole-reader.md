# UnmarshalRead with ReportErrorsWithLegacySemantics reads the whole reader first

- Area: decoder (v2, options of encoding/json)
- Kind: deviation
- Priority: 3
- Status: recorded 2026-10-28, found in the review of the options of encoding/json

## What

With `ReportErrorsWithLegacySemantics`, `UnmarshalRead` and `UnmarshalReadOf` read the whole input of
the reader (`Stream.ReadAllV2`) before they check its syntax. encoding/json/v2 checks the value as it
reads it (`CheckNextValue`), and stops at the first syntax error:

- A syntax error followed by an error of the reader: encoding/json/v2 reports the `SyntaxError`
  (`{"a":}`, `[1,]`, `[1] x`, `{"a":1}{`, each followed by a reader which fails); go-json the error of
  the reader.
- A syntax error followed by a long or an endless input: encoding/json/v2 returns at once; go-json
  reads on, into memory, up to the end of the reader.

A valid input is held whole by both, as the option checks the whole value before it is decoded.

## Fix

Read the value by a `jsontext.Decoder` over the reader, as `UnmarshalDecode` with the option does
(`unmarshalDecodeWithV1Errors`), which stops at the first syntax error, then read up to the end of
the input for the data after the value; `Stream.ReadAllV2` and the stream parameter of
`unmarshalWithV1Errors` would go.

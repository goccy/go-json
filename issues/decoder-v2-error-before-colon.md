# v2: a syntax error after a member name comes before an error which needs no value

- Area: decoder (v2)
- Kind: deviation
- Priority: 4
- Status: recorded 2026-10-20, found in the review of the v2 struct decoding

## What

encoding/json/v2 reads the name of a member alone, and the colon after it only with the
value. An error which needs no value of the member is found before the colon:

- the embedded `jsontext.Value` fallback already holds a value which is not an object;
- a nil embedded pointer to an unexported struct type on the way to the field;
- a field of a type with no default representation (`time.Duration`);
- the `string` option on a type which it doesn't apply to (`[]int`).

The v2 decoder of go-json checks the colon before the value is decoded, so when the colon is
missing or the input ends after the name, it reports the `jsontext.SyntacticError` instead of
the `SemanticError`. Shape, with `Rest jsontext.Value \`json:",embed"\`` holding `1`:

```go
// Unmarshal([]byte(`{"B" 2}`), &v)
// go-json:            SyntacticError at 5: invalid character '2' after object name (expecting ':')
// encoding/json/v2:   SemanticError at 1, "/B": embedded raw value must be a JSON object
```

Both fail on such input; only the error which is reported differs. The checks of the name (an
unknown name with `RejectUnknownMembers`, a duplicate name) already come before the colon.

## Direction

The checks which read no value could move before the colon check of `v2StructDecoder.Decode`
(internal/decoder/decode_v2_struct.go); the errors of the decoders of the field types which read
nothing would need a way to be asked for before the colon. Do it only if it removes code or
costs none.

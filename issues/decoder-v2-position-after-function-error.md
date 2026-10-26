# v2: where a decoder goes on after the error of a function which read its value

- Area: decoder (v2)
- Kind: defect
- Priority: 4
- Status: recorded 2026-10-26 on master, found by the review of the offsets of the errors of functions

## What

`UnmarshalDecode` of a decoder of the caller, into a value of a function of `UnmarshalFromFunc` which reads its
whole value by `UnmarshalDecode` of its own decoder and then returns an error: encoding/json/v2 returns the
`SemanticError` with the decoder after the value, which the next read goes on from; go-json leaves the decoder
failed, or before the value:

```go
o := json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
	var s string
	json.UnmarshalDecode(dec, &s, json.WithUnmarshalers(nil))
	return errors.New("x")
}))
dec := jsontext.NewDecoder(r) // `["x...x","b"]`, a string of 3000 bytes
dec.ReadToken()               // [
err := json.UnmarshalDecode(dec, new(string), o)
tok, err2 := dec.ReadToken()
```

encoding/json/v2: `tok` is `"b"` at the input offset 3007. go-json: by reads of 1 or 64 bytes, every read after
the error fails with the same `SemanticError` at the input offset 1; by a read of the whole input, `tok` is the
long string again ( the decoder went back before the value ). The `SemanticError` is the same.

The state of a decoder after a failed call is not documented, and go-json doesn't follow encoding/json/v2's for
it; but go-json's own design goes on after the value of a `SemanticError` ( see the design of the decoding of
streams ), which this state breaks, and the state depends on the size of the reads.

## How to fix

After the `SemanticError` of a function which read its value whole, the decoder of the caller goes on after the
value, as after a `SemanticError` of a value which the stream decoded ( see the design of the decoding of
streams ): find where the stream of the call puts its decoder back for an error of a function, which it treats as
an error within the value.

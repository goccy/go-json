# v2: a method which ignores the error of its UnmarshalDecode succeeds

- Area: decoder (v2)
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-24 on master, found by the tests of the names of no field of nested objects

## What

An `UnmarshalJSONFrom` method which calls `UnmarshalDecode` on the decoder it is given, and returns nil after
a `SemanticError` of it, makes the call fail in encoding/json/v2, and succeed in go-json:

```go
type T struct{ V struct{ X int } }

func (t *T) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	_ = json.UnmarshalDecode(dec, &t.V) // `{"X":"x"}`: a SemanticError of X
	return nil
}
```

`Unmarshal` of `[{"A":{"X":"x"}}]` into `[]struct{ A T }` returns nil in go-json, and in encoding/json/v2
`json: cannot unmarshal into Go main.T within "/0/A": must read or write exactly one value`, for `Unmarshal`,
`UnmarshalRead` and `UnmarshalDecode`, and every size of the reads of the reader. go-json skips the rest of the
value after the `SemanticError` ( see the design of the decoding of streams ), so that the method has read one
value; encoding/json/v2 leaves its decoder in the value.

## How to fix

Find where encoding/json/v2 leaves the decoder of a method after a `SemanticError` of `UnmarshalDecode`, and
leave the attached decoder of go-json there, so that the method fails for not reading one value as it does in
encoding/json/v2; compare with it by a test of a method which ignores the error, for the three functions and the
sizes of the reads.

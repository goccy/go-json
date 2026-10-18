# The root package calls UnmarshalJSONFrom and MarshalJSONTo

- Area: decoder, encoder
- Kind: defect (missing Go 1.27 behavior)
- Priority: 1, right after [decoder-v2-api.md](decoder-v2-api.md)
- Status: open, must not be forgotten

## What

Go 1.27's `encoding/json` calls the `encoding/json/v2` methods `UnmarshalJSONFrom(*jsontext.Decoder)`
and `MarshalJSONTo(*jsontext.Encoder)`. go-json's root package calls neither:

- a type with only `MarshalJSONTo` is written as `{}` by go-json; std writes the method's output;
- a type with only `UnmarshalJSONFrom` is decoded field by field by go-json.

`MarshalJSONTo` is reached today only through the hooks that `github.com/goccy/go-json/v2`
sets (`internal/encoder/method_v2.go`).

Checked on go1.27.1 on 2026-10-14.

## Notes

- The README's Features section states this gap; remove that sentence when this is done.
- The decoding side needs the attached `jsontext.Decoder` of the v2 decoder, so it comes after it.

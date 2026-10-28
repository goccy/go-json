# UseNumber of a Decoder is not given to UnmarshalJSONFrom of encoding/json/jsontext

- Area: decoder (root package)
- Kind: deviation from encoding/json
- Priority: 3
- Status: recorded 2026-10-28, found in the review of the calls of `UnmarshalJSONFrom` from the root
  package

## What

A `Decoder` of the root package gives the decoder of an `UnmarshalJSONFrom` method the options of
encoding/json and its `DisallowUnknownFields` and `UseNumber`, as encoding/json gives them its methods:
`UnmarshalDecode` of v2 within the method decodes as the `Decoder` does. For a method which takes the
`Decoder` of encoding/json/jsontext, `UseNumber` is not given: `UnmarshalDecode` of encoding/json/v2
within the method decodes a number into an empty interface as a `float64`, where encoding/json decodes a
`json.Number`.

    type wrapped struct{ V any }

    func (w *wrapped) UnmarshalJSONFrom(dec *jsontext.Decoder) error { // encoding/json/jsontext
        return jsonv2.UnmarshalDecode(dec, w.V) // encoding/json/v2
    }

    dec := json.NewDecoder(strings.NewReader(`42`))
    dec.UseNumber()
    dec.Decode(&wrapped{V: new(any)}) // float64(42); encoding/json: json.Number("42")

The methods which take the `Decoder` of github.com/goccy/go-json/jsontext are given it.

## Why

encoding/json gives `UseNumber` by an option which it doesn't export (`unmarshalAnyWithRawNumber`, a
flag of its internal `jsonflags`), and no exported option of encoding/json/v2 sets it. The ways to give
it which were found are not taken: reading the unexported field of the options of a `Decoder` of
encoding/json by its name, and the options of a decoder which encoding/json gives a method, which carry
its internal flag of the call (`WithinArshalCall`), with which a decoder reset by them can't be reset
again.

## Fix

A decoder of encoding/json/jsontext with the option, made by encoding/json: a `Decoder` of encoding/json
with `UseNumber` which decodes the value into a value whose `UnmarshalJSONFrom` calls the method with
the decoder it is given, and reports whether the method declined. It costs a `Decoder` of encoding/json
for each value, or a way to reuse one.

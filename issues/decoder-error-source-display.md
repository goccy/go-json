# Show the place in the input for a decode error

- Area: decoder
- Kind: feature
- Priority: 4
- Status: open

## What

A decode error gives only an offset. goccy/go-yaml's `FormatError(err, colored, inclSource)`
prints the lines of the source around the error with a marker; decode errors of go-json should
be able to do the same.

## Constraint

The `Error()` strings stay as encoding/json's. The feature is a separate function (or data the
errors carry, such as the offset of `SyntaxError` and `UnmarshalTypeError`), given the input.

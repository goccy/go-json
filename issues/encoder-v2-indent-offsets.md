# v2 with indentation: offsets leave out the white space

- Area: encoder
- Kind: defect
- Priority: 3
- Status: recorded 2026-10-11; re-check on master before fixing

## What

With `jsontext.Multiline` or `WithIndent`:

- the `OutputOffset` that a method of a map value sees leaves out the white space before it,
  unlike encoding/json/v2 (string-keyed maps too);
- the offsets of errors point before the indentation; encoding/json/v2 points after it.

The output is written compact and indented afterwards, so the offsets are those of the compact
output. The fix is engine-wide.

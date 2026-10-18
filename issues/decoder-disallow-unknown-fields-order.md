# DisallowUnknownFields returns the unknown field instead of an earlier error

- Area: decoder
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-11; re-check on master before fixing

## What

With `Decoder.DisallowUnknownFields`, when an input has a type error (or a bad map key) and
after it an unknown field, go-json stops at the unknown field and returns its error.
encoding/json keeps the first error in input order, reads on, and returns that one.

## Note

The v2 option `RejectUnknownMembers` decides the same thing, so the fix may share code with
[decoder-v2-api.md](decoder-v2-api.md).

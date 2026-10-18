# Raise the pin of Go 1.27 in CI past 1.27.1

- Area: ci
- Kind: maintenance
- Priority: 3
- Status: open

## What

CI pins Go 1.27 to 1.27.1 (#711) because master failed with go1.27.2 without any change of ours.
Go 1.25 and 1.26 are not pinned.

What go1.27.2 changed:

- golangci-lint v2.13.2 can't read go1.27.2's export data; v2.14.0 (built with go1.27.2)
  reports no issue.
- encoding/json decoding a float out of range (`1e400`) leaves 0 with the error, where it set
  ±Inf (`TestIssue555NumbersOutOfRange`).
- encoding/json writes a key of a string type with `MarshalText` in a `map[any]` as its string
  value, not through `MarshalText` (`TestEncodeMapInterfaceKeys/named_text_marshaler`).

## Plan

One pull request: raise the pin, bump golangci-lint, and follow the new behavior on every Go
version. Check against the latest 1.27 patch release at the time.

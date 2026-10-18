# Small encoder deviations

- Area: encoder
- Kind: defect
- Priority: 3
- Status: recorded; re-check each on master before fixing

## What

- v2 `Marshal` of `map[any]int` with number keys writes them; encoding/json/v2 fails because an
  object member name must be a string (recorded 2026-10-11).
- v1 with HTML escaping off: U+2028 and U+2029 in the output of a `MarshalJSON` method are not
  escaped as encoding/json escapes them (recorded 2026-10-10).
- v2 `PreserveRawStrings` with `Deterministic` keeps the raw escapes in names where
  encoding/json/v2 sorts and writes them decoded (recorded 2026-10-11).
- `Colorize` with a color format that contains `"` or `\` breaks the order of sorted map keys
  (recorded 2026-10-11).
- v2 `MarshalOf` drops its output on an error, so the error loses its place compared with
  `Marshal` (recorded 2026-10-08).
- Recorded 2026-09-21 as needing a change of the VM:
  - an `omitempty` field whose type has `MarshalText` is written when it is the zero value;
  - a nil map whose type has a marshaler is written as `null` instead of calling the marshaler;
  - the indentation of a recursive type held by an interface value is wrong.

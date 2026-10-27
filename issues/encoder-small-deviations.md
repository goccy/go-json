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
- v2 `Marshal` of a map whose key is not a name, behind a pointer or held by an interface:
  the `GoType` of the `SemanticError` is the type of the value the key points to or holds, where
  encoding/json/v2 reports the type of the key ( `map[*[2]int]int`: `[2]int` for `*[2]int`;
  `map[any]int{[2]int{1, 2}: 1}`: `[2]int` for `interface {}`; recorded 2026-10-28, found in the
  review of `FormatByteArrayAsArray`, which makes `[N]byte` such a key too ).
- v2 `Marshal` of a field of `omitempty` of a `[0]T`, whose value a function of `WithMarshalers` writes as a
  value which is not empty: the field is always omitted, where encoding/json/v2 writes what the function writes
  ( ``struct{ A [0]int `json:",omitempty"` }`` with a function which writes `1`: `{}` against `{"A":1}`;
  recorded 2026-10-28, found in the review of `OmitEmptyWithLegacySemantics`, under which the field is omitted
  as encoding/json omits it ).
- v2 `MarshalOf` drops its output on an error, so the error loses its place compared with
  `Marshal` (recorded 2026-10-08).
- v2 `MarshalEncode` of a value which is not a string, where the encoder is at the name of a
  member, writes the value before it fails: its methods run, and an unsupported value in it is a
  `SemanticError`, where encoding/json/v2 fails at once with a `SyntacticError` at the place
  (recorded 2026-10-25, found by the review of the names "" of the places of encoders).
- v1 with Go 1.27.2 and later: a map key of a pointer type ( `map[*S]int`, `map[**S]int`, a
  named type of pointers ) which leads to a value of a string type with `MarshalText` is written
  by the method, as encoding/json documents it for a key which is not of a string type;
  encoding/json of those releases writes the string it leads to, and "" for a nil one (recorded
  2026-10-27). Writing the string needs an opcode of a pointer to a string whose nil is "", which
  the VM doesn't have. Such a pointer in a key of an interface type is written as the string.
- Recorded 2026-09-21 as needing a change of the VM:
  - an `omitempty` field whose type has `MarshalText` is written when it is the zero value;
  - a nil map whose type has a marshaler is written as `null` instead of calling the marshaler;
  - the indentation of a recursive type held by an interface value is wrong.

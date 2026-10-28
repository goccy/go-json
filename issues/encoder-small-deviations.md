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
- v2: the error of a key of an interface type which holds a value which can't be a name, as a nil pointer
  ( `map[encoding.TextMarshaler]int{(*T)(nil): 1}` ), is a `SemanticError` of the type of the held value; the one of
  encoding/json/v2 is of the type of the key, the interface type (recorded 2026-10-28, found in the review of
  `CallMethodsWithLegacySemantics`). The text and the place are the same otherwise.
- v2 with `CallMethodsWithLegacySemantics`: a nil pointer held by an interface value of a method of marshaling is
  written by the method, called for the nil pointer, but not by a function of an interface type which the interface
  type implements ( `MarshalFunc(func(encoding.TextMarshaler) ...)` with `[]encoding.TextMarshaler{(*T)(nil)}` or
  such a key ): encoding/json/v2 wraps the nil pointer in a struct of the interface type, whose pointer such a
  function takes, while a nil pointer is not a value which the function takes by the rule of the functions (its
  pointer doesn't implement the interface). A method of the value which a nil pointer points to is not called, where
  encoding/json/v2 panics: such a value is null, such a key the error of a nil pointer (recorded 2026-10-28, found in
  the review of `CallMethodsWithLegacySemantics`).
- v2: an empty slice or map one level past the max depth ( 10001 levels of `type R []R` whose innermost value is
  `R{}`, of `[]any`, a nil slice field or an inline map ) fails as deeper than the max depth in a compact output;
  encoding/json/v2 writes `[]` or `{}` without counting its level (recorded 2026-10-28, found in the review of
  `CallMethodsWithLegacySemantics`, without the option too).
- v2: the error of a value deeper than the max depth is at the start of the value whose opcodes go deeper, a level or
  two above the token where encoding/json/v2 reports it ( `type C [1]struct{ V int; Next []C }` nested 3333 times:
  offset 53328 and a pointer one token shorter, 53329 in encoding/json/v2; recorded 2026-10-28, found in the review of
  `CallMethodsWithLegacySemantics` ).
- v2 `MarshalWrite`: the error of the writer is returned as it is; encoding/json/v2 wraps it as the error of the writer
  of its encoder ( `jsontext: write error: ...`, which `errors.Is` sees through; recorded 2026-10-28, found in the work
  on `ReportErrorsWithLegacySemantics`, with which both return it as it is).
- v2 with `ReportErrorsWithLegacySemantics`: an error of a name of a key of a map which a method or a function wrote,
  a name which is not a string or a name which another key has, is an `UnsupportedValueError` or a `SyntaxError`, as
  the ones of the other names; encoding/json wraps it in a `MarshalerError` of the method, as it writes the name in its
  call. The kind of a function, `MarshalFunc` or `MarshalToFunc`, is not known where the names are checked (recorded
  2026-10-28).
- v2 with `ReportErrorsWithLegacySemantics`: a `jsontext.SyntacticError` which a method or a function returned is a
  `SyntaxError` at its offset plus one, as the ones of the output, which encoding/json reports after the text of the
  error; encoding/json reports it at its own offset, as only the errors which its jsontext makes with the option are
  after their text (recorded 2026-10-28, found in the review of `ReportErrorsWithLegacySemantics`).
- Recorded 2026-09-21 as needing a change of the VM:
  - an `omitempty` field whose type has `MarshalText` is written when it is the zero value;
  - a nil map whose type has a marshaler is written as `null` instead of calling the marshaler;
  - the indentation of a recursive type held by an interface value is wrong.

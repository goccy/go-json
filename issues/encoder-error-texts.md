# Encoder error texts and types that differ from encoding/json

- Area: encoder
- Kind: defect
- Priority: 4
- Status: recorded 2026-10-14; re-check on master before fixing

## What

- The error for a cycle names the type as `via main.Cyc`; encoding/json names it `*main.Cyc`.
- v2: the `SemanticError` of a cycle through a slice or an array of pointers to its own type ( `type R []*R`,
  `type A [1]*A` ) names the type `main.R`; encoding/json/v2 names the pointer `*main.R`, and the one of a cycle
  through a slice of an array type `A` names `main.A`, where encoding/json/v2 names `[]main.A` (recorded 2026-10-28,
  found in the review of `CallMethodsWithLegacySemantics`).
- When a value-receiver `MarshalJSON` fails, the type that `MarshalerError` names differs from
  encoding/json's.

## Note

Error texts are not documented API; change them only where it costs nothing.

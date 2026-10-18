# Encoder error texts and types that differ from encoding/json

- Area: encoder
- Kind: defect
- Priority: 4
- Status: recorded 2026-10-14; re-check on master before fixing

## What

- The error for a cycle names the type as `via main.Cyc`; encoding/json names it `*main.Cyc`.
- When a value-receiver `MarshalJSON` fails, the type that `MarshalerError` names differs from
  encoding/json's.

## Note

Error texts are not documented API; change them only where it costs nothing.

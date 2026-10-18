# Errors of map keys of pointer types, and maps after a failing key

- Area: decoder
- Kind: defect
- Priority: 3
- Status: recorded 2026-10-13; re-check on master before fixing

## What

- For a map whose key is a pointer to a kind that can't be a key, the `Type` of the
  `UnmarshalTypeError` is the pointer type in go-json, the type pointed to in Go 1.27.
- When a key's `UnmarshalText` fails, what the map holds afterwards differs from encoding/json
  (recorded: an entry with an empty key was stored where encoding/json keeps the map without it).

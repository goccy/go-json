# Decoding API of github.com/goccy/go-json/v2

- Area: decoder
- Kind: feature
- Priority: 1
- Status: plan to be approved

## What

`github.com/goccy/go-json/v2` has only the encoding half of the `encoding/json/v2` API
(`Marshal`, `MarshalWrite`, `MarshalEncode`, `MarshalOf`, `MarshalWriteOf`, the options,
`MarshalFunc` / `MarshalToFunc` and `MarshalJSONTo` methods). The decoding half is missing:

- `Unmarshal`, `UnmarshalRead`, `UnmarshalDecode`;
- `UnmarshalFunc`, `UnmarshalFromFunc`, `WithUnmarshalers`, and the `UnmarshalJSONFrom` method;
- the decoding options (`RejectUnknownMembers`, `MatchCaseInsensitiveNames`,
  `AllowDuplicateNames`, `AllowInvalidUTF8`, `StringifyNumbers`, the `format` tag option, ...);
- go-json's own allocation-free generic forms, as on the encoding side
  (`UnmarshalOf`, `UnmarshalReadOf`, `UnmarshalDecodeOf`).

## Decisions

- Build on master in small pull requests, one goal each, on go-json's v1 decoder engine
  (`internal/decoder`). The branch `feat/json-v2-decoder` (64766b5) is a reference only and is
  not merged: it depends on emulation of the state of `encoding/json/v2` after failed or misused
  calls, which go-json does not imitate.
- Match the documented behavior of Go 1.27's `encoding/json/v2` on every Go version.
  Not imitated: where std stops after a failed `UnmarshalDecode`, misuse states, error texts.
- Partial values after an error: only what the engine keeps by itself, no code for it.
- The running Go's `unicode` package is used; no Unicode tables for older Go.
- `UnmarshalRead` and `UnmarshalDecode` stream with bounded memory, on the element-wise stream
  decoder of #736 (fixed 512-byte reads). Reading a whole value first is not acceptable.
- Speed: v2 is at least as fast as go-json v1 on every payload, and faster than sonic.
  v1 must not get slower from code shared with v2.
- The functions are exported from the first pull request, with the gaps documented;
  the release waits until the API is complete.

## Lessons from the reference branch

- v2 code inside functions shared with v1 slows v1 even when it is skipped: a reassigned local
  across a call, one more live variable in a scan loop, more fields in `RuntimeContext`
  (the stream allocates one per `NewDecoder`; 272 to 336 bytes crossed a size class).
  The branch kept v2 state behind one pointer and v2 checks in tail calls.
- `decodeStringField` must stay as it is: even cold v2 branches in it cost v1
  `LargeSlice_EscapedString` 3-4% on arm64.
- With `MatchCaseInsensitiveNames`, the branch was 2-5x slower than v1 on untagged structs
  (the folding path for other members).
- `encoding/json/v2` reports some errors in an order that follows how it reads: type errors of
  string, number and bytes kinds after the whole value, those of bool, struct, map, slice and
  array at the first token; unknown names before the `:`.

## Next step

Write the plan of the pull requests and have it approved before any code.
Then [v1-calls-v2-methods.md](v1-calls-v2-methods.md) follows right after the series.

# Decoding API of github.com/goccy/go-json/v2

- Area: decoder
- Kind: feature
- Priority: 1
- Status: plan approved on 2026-10-19; pull requests 1 (#746), 2 and 3 merged, pull request 4 in progress

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

## Plan

The reference is the first commit of the branch, c7d2227, before the emulation was added.

- A v2 value is decoded by the v1 decoder of its type where the documented behavior is the same.
  Types of their own only where it differs: structs, maps, `[]byte` (strict base64), the UTF-8
  checks of strings, `time.Time`, methods and functions. The hot functions of v1 get no v2
  branch.
- The v2 state is one pointer in `RuntimeContext`, which must stay in its size class.
- The JSON pointer of an error is built from the path that the type errors of master already
  record (the steps of `pendingTypeError`), never by reading the input again from its start, so
  that streams can use the same.
- `UnmarshalRead` and `UnmarshalDecode` add v2 paths for structs, maps and slices to the
  element-wise stream decoding of #736. Their design is shown for approval before they start.
- The tests of encoding/json/v2 are ported to `v2/internal/stdtest` (tests only, as for
  marshaling), and each pull request enables the part it covers.
- Benchmarks of v2 decoding come with the first pull request, so the gate measures v2 from then.

Pull requests, in order (sizes without tests are estimates):

| # | Content | Size |
|---|---|---|
| 1 | `Unmarshal` of bool, numbers, strings, `[]byte`, `any`, slices, arrays and pointers; null as the zero value; `AllowInvalidUTF8`, `StringifyNumbers`, `AllowDuplicateNames`; `SemanticError` and `SyntacticError`; benchmarks of `any` payloads | 1,200-1,800 |
| 2 | Structs and maps: fields chosen by `jsonfields.Of` (shared with the encoder), exact names, `case:ignore`, `MatchCaseInsensitiveNames`, duplicate names, unknown members (skipped, `RejectUnknownMembers`, inline and unknown fields), `format`, map keys; benchmarks of struct payloads | 1,500-2,000 |
| 3 | Methods and functions: `UnmarshalJSON`, `UnmarshalText`, `time.Time`, `time.Duration`, `UnmarshalFunc`, `WithUnmarshalers`, `JoinUnmarshalers`, the exported `Unmarshaler` | 600-900 |
| 4 | `UnmarshalJSONFrom` and `UnmarshalFromFunc`: a `jsontext.Decoder` attached to the input of the engine, the counterpart of the encoder's | 400-600 |
| 5 | `UnmarshalOf`, the rows of the benchmark report, and v2 at least as fast as v1 | 150 and speed work |
| 6 | `UnmarshalRead` with bounded memory, and `UnmarshalReadOf` | 600-1,200 |
| 7 | `UnmarshalDecode` and `UnmarshalDecodeOf` | 500-900 |

Then [v1-calls-v2-methods.md](v1-calls-v2-methods.md).

Speed of v2 against v1 measured so far, for pull request 5 (arm64, Apple M5):

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `LargeSlice_EscapedString` (`[]string`) | 181-197 µs | 215 µs | each decoded string is scanned again for U+FFFD |
| `Code_Unmarshal` | 1.53 ms | 2.07 ms | not looked into |
| `TwitterBinding_Unmarshal` | 5.0 µs | 7.7 µs | not looked into |
| `SmallStruct_Unmarshal`, case-insensitive | 106 ns | 316 ns | names of another case go through `jsonfields.FoldName` and a map |
| `MediumStruct_Unmarshal`, case-insensitive | 776 ns | 3.65 µs | the same |
| `LargeStruct_Unmarshal`, case-insensitive | 10.0 µs | 66 µs | the same |
| `InterfaceUnmarshaler_Unmarshal` (`UnmarshalJSON` of the values of interfaces) | 185 ns | 565 ns | not looked into |
| `Times_Unmarshal` (`[]time.Time`) | 4.1 µs | 6.0 µs | not looked into |
| `GitHubREST_Unmarshal` (times by `UnmarshalJSONFrom` in v2, `UnmarshalJSON` in v1) | 98 µs, 2,703 allocations | 130 µs, 2,840 allocations | not looked into |
| `GitHubGraphQL_Unmarshal`, case-insensitive | 57 µs, 135 allocations | 105 µs, 1,749 allocations | not looked into; maybe the names of another case, as for the struct payloads |

`TwitterGeneric_Unmarshal` and `CodeAny_Unmarshal` (`any`) are as fast as v1.

Risks:

- Code shared with v1 can slow v1; the gate measures it, and a pull request that slows v1
  fixes it itself. v2 runs through functions of its own where v1's hot ones are kept as they
  are (the reference branch was 5-7% slower than v1 from it), which pull request 5 wins back.
- Until the series ends, master has an incomplete v2 decoding API, with the gaps documented;
  no release is tagged before it is complete.
- Pull requests 6 and 7 are new designs of unknown size and touch the stream decoding of v1
  (`stream_elements.go`).
- Where the positions built from the path differ from std's documented `ByteOffset` or
  `JSONPointer`, the pull request that finds it makes them equal.
- Where the reference branch and master differ (its own fix of recursive types, field choice),
  master's way is used: #742 for recursion, `jsonfields.Of` for the v2 field rules.

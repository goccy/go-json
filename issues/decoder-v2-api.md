# Decoding API of github.com/goccy/go-json/v2

- Area: decoder
- Kind: feature
- Priority: 1
- Status: plan approved on 2026-10-19; pull requests 1 (#746) to 6 merged, pull request 7 in progress

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
- The JSON pointer and the offset of an error are found by jsontext, which reads the input again
  as encoding/json/v2 reads it: from its start for `Unmarshal`, and for a stream from the place of
  the error in the buffer, under the levels of the path ( see below ).
- `UnmarshalRead` and `UnmarshalDecode` add v2 paths for structs, maps and slices to the
  element-wise stream decoding of #736, by the design approved on 2026-10-23:
  - v1's stream code is not changed: v2 has its own element functions
    (`internal/decoder/stream_elements_v2.go`), which call v1's read and scan functions. A value
    which the buffer holds whole is decoded by the decoders of `Unmarshal`, with the strings copied,
    so that the buffer is reused. The decoding stops at the first error.
  - Values which encoding/json/v2 reads whole (`UnmarshalJSON`, `UnmarshalText`, `UnmarshalFunc`,
    the types of strings, numbers and bytes, `jsontext.Value`) are read whole. `UnmarshalJSONFrom`
    and `UnmarshalFromFunc` are given a `jsontext.Decoder` attached at the value, which reads on
    from the reader of the stream, 512 bytes at a time.
  - The path, the levels of the objects and arrays decoded element by element with their counts
    and last names, is kept in the state of the call ( `V2Call.Path` ). The innermost function
    records the part of the buffer which an error is found again in: from the end of the token
    before the member, the element or the value which failed, which the buffer keeps until the
    value is decoded or begins to be decoded element by element.
  - A member which no field takes and which the buffer doesn't hold whole is skipped by the
    attached decoder, which checks it as encoding/json/v2 does (duplicate names, UTF-8).
  - After a failed `UnmarshalDecode`, a `SemanticError` leaves the decoder after the value, whose
    rest is read and checked by its grammar; any other error leaves the decoder before the value, as
    an error of `Decoder.ReadValue` does, which encoding/json/v2's `TestUnmarshalDecodeOptions` reads
    again with other options (decided first as "every later read fails", changed for that test),
    but if the stream dropped the start of the value, which every later read fails with then. The
    place where encoding/json/v2 stops within a value is not imitated.
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
| 5 | `UnmarshalOf` and the rows of the benchmark report; v2 at least as fast as v1 after pull request 7 | 150, then speed work |
| 6 | `UnmarshalRead` with bounded memory, and `UnmarshalReadOf` | 600-1,200 |
| 7 | `UnmarshalDecode` and `UnmarshalDecodeOf` | 500-900 |

The speed work of pull request 5 comes after pull request 7: the API is completed first, as decided on
2026-10-19 (see the order in [README.md](README.md)).

Then [v1-calls-v2-methods.md](v1-calls-v2-methods.md).

Speed of v2 against v1 measured so far, for the speed work after pull request 7 (arm64, Apple M5):

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `LargeSlice_EscapedString` (`[]string`) | 181-197 µs | 215 µs | each decoded string is scanned again for U+FFFD |
| `Code_Unmarshal` | 1.53 ms | 2.07 ms | `memberOf` for each name (15%: the scan for a byte which is not ASCII or an escape, and the comparison with the name of the field), the loop of `v2StructDecoder` (5% more than the one of v1), the check of the fields met twice (`fieldSeen`, 2%), and `v2ScalarDecoder` before the decoder of v1 of each number (3%) |
| `TwitterBinding_Unmarshal` | 5.0 µs | 7.7 µs | `memberOf` (23%: 14% the scan of the names, 6% the comparison), and the scan of each decoded string for U+FFFD (3%) |
| `SmallStruct_Unmarshal`, case-insensitive | 106 ns | 316 ns | names of another case go through `jsonfields.FoldName` and a map |
| `MediumStruct_Unmarshal`, case-insensitive | 776 ns | 3.65 µs | the same |
| `LargeStruct_Unmarshal`, case-insensitive | 10.0 µs | 66 µs | the same |
| `InterfaceUnmarshaler_Unmarshal` (`UnmarshalJSON` of the values of interfaces) | 185 ns | 565 ns | `decodeHeld` decodes the value which an interface holds through reflect (about 55%): `reflect.New` for a copy, and `reflect.Value.Set` back into the interface (29%, of which 15% checks that the type implements it); `DecoderOfV2` looks the decoder up for each value (4%) |
| `Times_Unmarshal` (`[]time.Time`) | 4.1 µs | 6.0 µs | each string is read three times: skipped by `ReadValue`, scanned for a byte which is not ASCII or an escape and for `{` (17%), and read again by `unquotedName` (11%) |
| `GitHubREST_Unmarshal` (times by `UnmarshalJSONFrom` in v2, `UnmarshalJSON` in v1) | 98 µs, 2,703 allocations | 130 µs, 2,840 allocations | the scan of each decoded string for U+FFFD (6%), and `memberOf` (9%); the times are not slower by `UnmarshalJSONFrom` than by `UnmarshalJSON` in v1 |
| `GitHubGraphQL_Unmarshal`, case-insensitive | 57 µs, 135 allocations | 105 µs, 1,749 allocations | the names of another case: `jsonfields.FoldName` allocates each folded name (88% of the allocations), and `memberOf` is 37% (`FoldName` 20%, the map of the folded names 9%) |

| 1.5 MB of records into `[]struct{}`, by `UnmarshalRead` (names of no field) | | 8.8 MB allocated | each name of no field is copied into the set of the names of its object, for the check of names met twice ( `nameSet` in `v2StructDecoder` ) |

`TwitterGeneric_Unmarshal` and `CodeAny_Unmarshal` (`any`) are as fast as v1.

`UnmarshalRead` against v1's `Decoder`, from a reader of 4 KiB at a time (the `ChunkedStream` benchmarks, the stream
reading 512 bytes at a time in both): the most memory kept alive at once and the allocations are v1's, but for
`CodeAny` (8.0 MB allocated against 6.2 MB, 78,000 allocations against 29,500).

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `Code_ChunkedStream` | 2.34 ms | 3.24 ms | the ones of `Code_Unmarshal` |
| `Records_ChunkedStream` | 4.26 ms | 6.32 ms | not looked into |
| `SkipRecords_ChunkedStream` | 1.93 ms | 4.64 ms | a value of no field which the buffer doesn't hold whole is skipped by a `jsontext.Decoder`, token by token, which checks it as encoding/json/v2 does |
| `Twitter_ChunkedStream` | 10.8 µs | 15.1 µs | the ones of `TwitterBinding_Unmarshal` |
| `CodeAny_ChunkedStream` | 4.18 ms | 4.91 ms | not looked into; the objects decoded element by element are merged into maps of no size |
| `NDJSON_ChunkedStream` (a record by `UnmarshalDecode`, against `Decoder.Decode`) | 2.38 ms | 4.33 ms | not looked into; each call sets up its state and its stream, which copies 512 bytes of the buffer of the decoder for a record of 180 |

The causes come from CPU and allocation profiles in a linux/arm64 container (2026-10-23), and the percentages are
of the time of v2. Each is the work of one function of a decoder of v2; none is in the attached decoder of
`UnmarshalJSONFrom` or in the way v2 is lowered onto the engine of v1, so the speed work changes those functions,
not the design which the next pull requests build on.

Risks:

- Code shared with v1 can slow v1; the gate measures it, and a pull request that slows v1
  fixes it itself. v2 runs through functions of its own where v1's hot ones are kept as they
  are (the reference branch was 5-7% slower than v1 from it), which the speed work wins back.
- Until the series ends, master has an incomplete v2 decoding API, with the gaps documented;
  no release is tagged before it is complete.
- Pull requests 6 and 7 are new designs of unknown size and touch the stream decoding of v1
  (`stream_elements.go`).
- Where the positions built from the path differ from std's documented `ByteOffset` or
  `JSONPointer`, the pull request that finds it makes them equal.
- Where the reference branch and master differ (its own fix of recursive types, field choice),
  master's way is used: #742 for recursion, `jsonfields.Of` for the v2 field rules.

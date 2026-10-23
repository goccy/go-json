# Decoding of v2 at least as fast as v1

- Area: decoder (v2)
- Kind: speed
- Priority: 1
- Status: open; the decoding API of v2 is complete (#746 to #754); struct member names are found by
  their words, as v1 finds them, since 2026-10-23

## Goal

`github.com/goccy/go-json/v2` decodes every payload at least as fast as the root package (v1), and
faster than sonic. v1 must not get slower from code shared with v2.

## Gaps

Medians of the gate on the decode benchmarks, arm64 (Apple M5), 2026-10-23. The causes come from CPU
and allocation profiles in a linux/arm64 container; the percentages are of the time of v2.

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `LargeSlice_EscapedString` (`[]string`) | 177 µs | 202 µs | each decoded string is scanned again for U+FFFD (`v2StringDecoder`), which v1 makes of invalid UTF-8 and of an escape of a lone surrogate, and v2 rejects |
| `Code_Unmarshal` | 1.51 ms | 1.62 ms | `v2ScalarDecoder` around the decoder of v1 of each number (4%), the scan of the decoded strings for U+FFFD (2.5%), and the loop of `v2StructDecoder` |
| `TwitterBinding_Unmarshal` | 5.3 µs | 6.1 µs | the scan of the decoded strings for U+FFFD (5%), and `v2ScalarDecoder` |
| `GitHubREST_Unmarshal` | 103 µs, 2,703 allocations | 117 µs, 2,711 allocations | the scan of the decoded strings for U+FFFD (6%); the lists of the sets of the names of no field ( `nameSet` ), which v1 doesn't check for duplicates |
| `SmallStruct_Unmarshal`, case-insensitive | 113 ns | 321 ns | a name of another case is folded by `jsonfields.FoldName`, which allocates it, and looked up in a map ( `memberOf` ); so is a name of no field |
| `MediumStruct_Unmarshal`, case-insensitive | 769 ns | 3.5 µs | the same |
| `LargeStruct_Unmarshal`, case-insensitive | 10.1 µs | 62 µs | the same |
| `GitHubGraphQL_Unmarshal`, case-insensitive | 60 µs, 135 allocations | 107 µs, 1,749 allocations | the same: `FoldName` makes 88% of the allocations, `memberOf` is 37% of the time |
| `InterfaceUnmarshaler_Unmarshal` (`UnmarshalJSON` of the values of interfaces) | 194 ns | 576 ns | `decodeHeld` decodes the value which an interface holds through reflect (about 55%): `reflect.New` for a copy, and `reflect.Value.Set` back into the interface, which checks that the type implements it; `DecoderOfV2` looks the decoder up for each value (4%) |
| `Times_Unmarshal` (`[]time.Time`) | 4.2 µs | 6.3 µs | each string is read three times: skipped by `ReadValue`, scanned for a byte which is not ASCII or an escape and for `{` (17%), and read again by `unquotedName` (11%) |
| 1.5 MB of records into `[]struct{}`, by `UnmarshalRead` | | 2.4 MB allocated | the lists of the sets of the names of no field, which grow for each object, and the names copied in the objects which a read cuts |

`CodeAny_Unmarshal` and `TwitterGeneric_Unmarshal` (`any`) are within 3% of v1.

`UnmarshalRead` and `UnmarshalDecode` against v1's `Decoder`, from a reader of 4 KiB at a time (the
`ChunkedStream` benchmarks; the stream reads 512 bytes at a time in both):

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `Code_ChunkedStream` | 2.35 ms | 2.81 ms | the ones of `Code_Unmarshal`, and the objects which a read cuts, decoded member by member ( `decodeStructElementsV2` ), each scalar through `elementEnd` and `decodeWholeV2` |
| `Records_ChunkedStream` | 4.94 ms | 6.10 ms | about 7,300 records, one at each read of 512 bytes, are cut by the read and decoded member by member; each whole record pays `nextElementV2`, `decodeNestedV2` and `decodeWholeV2` apart, where v1 decodes a run of them under one setting of the buffer ( `decodeRun` ) |
| `SkipRecords_ChunkedStream` | 1.93 ms | 4.63 ms | a value of no field which the buffer doesn't hold whole is skipped by a `jsontext.Decoder`, token by token, which checks it as encoding/json/v2 does |
| `CodeAny_ChunkedStream` | 4.16 ms | 5.50 ms | `decodeAnyObjectV2` allocates the value of each member ( moved to the heap ), and the list of its `nameSet` grows for each object: 78,000 allocations against 29,500 |
| `Twitter_ChunkedStream` | 11.0 µs | 13.8 µs | the ones of `TwitterBinding_Unmarshal`, and the objects which a read cuts |
| `NDJSON_ChunkedStream` (a record by `UnmarshalDecode`, against `Decoder.Decode`) | 2.48 ms | 4.03 ms | each call takes its state from three pools and applies its options, calls the hooks of jsontext, sets the path of the stream, and copies 512 bytes of the buffer of the decoder for a record of 180 |
| `SmallStruct_ChunkedStream`, case-insensitive | 284 ns | 384 ns | the ones of `SmallStruct_Unmarshal` |
| `MediumStruct_ChunkedStream`, case-insensitive | 1.7 µs | 4.5 µs | the ones of `MediumStruct_Unmarshal` |

## Constraints

- v2 code inside the functions of v1 slows v1 even when it is skipped: a reassigned local across a
  call, one more live variable in a scan loop, more fields in `RuntimeContext`. v2 runs through
  functions of its own; `decodeStringField` stays as it is (even cold v2 branches in it cost v1's
  `LargeSlice_EscapedString` 3-4% on arm64).
- Helpers of the hot types ( `RuntimeContext`, `Stream` ) are functions, not methods: a method
  moves the data of the binary, which slowed `CodeAny_Unmarshal` by 2%.
- Streams read 512 bytes at a time at most, with bounded memory; skipped values are checked as
  encoding/json/v2 checks them.

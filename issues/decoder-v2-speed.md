# Decoding of v2 at parity with v1

- Area: decoder (v2)
- Kind: speed
- Priority: 1
- Status: open; the decoding API of v2 is complete (#746 to #754); struct member names are found by
  their words, as v1 finds them, strings and the texts of methods and times are read once, and
  names are matched by their folded names without an allocation, since 2026-10-23; the names of no
  field are added to the set of the names of their object by their bits, kept in one stack of the
  call, and their values skipped and read for their text in one scan of jsontext, since 2026-10-24

## Goal

`github.com/goccy/go-json/v2` decodes every payload as fast as the root package (v1), within the
noise of the measures; it doesn't have to be faster. v1 must not get slower from code shared with
v2. Then faster than sonic, where v1 is (decided 2026-10-23: the payloads which v1 decodes slower
than sonic are not a target of this work). Being faster than master is not the measure: master
has the gaps below.

## Gaps

Medians of the gate on the decode benchmarks, arm64 (Apple M5), 2026-10-23; for the structs of many
names of no field, the medians of 8 rounds of each binary in turn, 2026-10-24. The causes come from CPU
and allocation profiles in a linux/arm64 container; the percentages are of the time of v2.

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `Code_Unmarshal` | 1.51 ms | 1.57 ms | `v2ScalarDecoder` around the decoder of v1 of each number (4%), and the loop of `v2StructDecoder` |
| `TwitterBinding_Unmarshal` | 5.3 µs | 5.6 µs | `v2ScalarDecoder`, and the loop of `v2StructDecoder` |
| `GitHubREST_Unmarshal` | 103 µs, 2,703 allocations | 108 µs, 2,586 allocations | the same |
| `SmallStruct_Unmarshal`, case-insensitive | 114 ns | 190 ns | `memberOf` for each name of another case (17%: the comparison with the name of the field first, the scan for a byte which is not ASCII or an escape), the loops of `v2StructDecoder` and of `v2StringDecoder`, and `v2ScalarDecoder` |
| `MediumStruct_Unmarshal`, case-insensitive | 770 ns, 7 allocations | 2.05 µs, 7 allocations | the values of no field, which jsontext reads for their names and their UTF-8 ( `skipValueV2`, in one scan of a skipper of the call for an object ); each name of no field is folded and looked up in a map of the folded names ( `memberOf` ) |
| `LargeStruct_Unmarshal`, case-insensitive | 10.1 µs, 66 allocations | 44.3 µs, 66 allocations | the same as `MediumStruct_Unmarshal` |
| `GitHubGraphQL_Unmarshal`, case-insensitive | 60 µs, 135 allocations | 73 µs, 136 allocations | not looked into since the folded names are found without an allocation |
| `InterfaceUnmarshaler_Unmarshal` (`UnmarshalJSON` of the values of interfaces) | 191 ns | 576 ns | `decodeHeld` decodes the value which an interface holds through reflect (about 55%): `reflect.New` for a copy, and `reflect.Value.Set` back into the interface, which checks that the type implements it; `DecoderOfV2` looks the decoder up for each value (4%) |
| 1.5 MB of records into `[]struct{}`, by `UnmarshalRead` | | 1.6 MB and 48 allocations | not looked into since the names of no field are kept in one stack of the call (3.9 MB and 71,903 allocations before) |

`LargeSlice_EscapedString`, `Times_Unmarshal`, `CodeAny_Unmarshal` and `TwitterGeneric_Unmarshal` are
within 2% of v1, and `SmallStruct_ChunkedStream` is faster than v1.

`UnmarshalRead` and `UnmarshalDecode` against v1's `Decoder`, from a reader of 4 KiB at a time (the
`ChunkedStream` benchmarks; the stream reads 512 bytes at a time in both):

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `Code_ChunkedStream` | 2.34 ms | 2.76 ms | the ones of `Code_Unmarshal`, and the objects which a read cuts, decoded member by member ( `decodeStructElementsV2` ), each scalar through `elementEnd` and `decodeWholeV2` |
| `Records_ChunkedStream` | 4.91 ms | 5.64 ms | about 7,300 records, one at each read of 512 bytes, are cut by the read and decoded member by member; each whole record pays `nextElementV2`, `decodeNestedV2` and `decodeWholeV2` apart, where v1 decodes a run of them under one setting of the buffer ( `decodeRun` ) |
| `SkipRecords_ChunkedStream` | 1.94 ms | 4.63 ms | a value of no field which the buffer doesn't hold whole is skipped by a `jsontext.Decoder`, token by token, which checks it as encoding/json/v2 does |
| `CodeAny_ChunkedStream` | 4.17 ms | 5.48 ms | `decodeAnyObjectV2` allocates the value of each member ( moved to the heap ): 61,000 allocations against 29,500 |
| `Twitter_ChunkedStream` | 10.9 µs | 13.4 µs | the ones of `TwitterBinding_Unmarshal`, and the objects which a read cuts |
| `NDJSON_ChunkedStream` (a record by `UnmarshalDecode`, against `Decoder.Decode`) | 2.47 ms | 3.78 ms | each call takes its state from three pools and applies its options, calls the hooks of jsontext, sets the path of the stream, and copies 512 bytes of the buffer of the decoder for a record of 180 |
| `MediumStruct_ChunkedStream`, case-insensitive | 1.7 µs | 3.2 µs | the ones of `MediumStruct_Unmarshal` |

## Constraints

- v2 code inside the functions of v1 slows v1 even when it is skipped: a reassigned local across a
  call, one more live variable in a scan loop, more fields in `RuntimeContext`. v2 runs through
  functions of its own; `decodeStringField` stays as it is (even cold v2 branches in it cost v1's
  `LargeSlice_EscapedString` 3-4% on arm64).
- Helpers of the hot types ( `RuntimeContext`, `Stream` ) are functions, not methods: a method
  moves the data of the binary, which slowed `CodeAny_Unmarshal` by 2%.
- Streams read 512 bytes at a time at most, with bounded memory; skipped values are checked as
  encoding/json/v2 checks them.

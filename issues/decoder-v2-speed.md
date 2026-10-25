# Decoding of v2 at parity with v1

- Area: decoder (v2)
- Kind: speed
- Priority: 1
- Status: open; the decoding API of v2 is complete (#746 to #754); struct member names are found by
  their words, as v1 finds them, strings and the texts of methods and times are read once, and
  names are matched by their folded names without an allocation, since 2026-10-23; the names of no
  field are added to the set of the names of their object by their bits, kept in one stack of the
  call, and their values skipped and read for their text in one scan of jsontext, the folded
  names looked up in a table as v1's of the names of the fields, and matched by the field which
  the folded name of each field matches, and bools and numbers decoded by the decoders of v1
  without another around them, plain names of folded names found by v1's table as they are read,
  numbers and literals of no field skipped without another read, pointers which interface
  values hold decoded in place, the scalars of the objects which a read of a stream cuts decoded
  in its buffer, the options of `UnmarshalDecode` without a pool, and the names of no field
  compared by their lengths and last bytes first, since 2026-10-24

## Goal

`github.com/goccy/go-json/v2` decodes every payload as fast as the root package (v1), within the
noise of the measures; it doesn't have to be faster. v1 must not get slower from code shared with
v2. Then faster than sonic, where v1 is (decided 2026-10-23: the payloads which v1 decodes slower
than sonic are not a target of this work). Being faster than master is not the measure: master
has the gaps below.

## Gaps

Medians of the gate on the decode benchmarks, arm64 (Apple M5), 2026-10-23; for the rows of 2026-10-24,
the medians of 8 rounds of each binary in turn. The causes come from CPU
and allocation profiles in a linux/arm64 container; the percentages are of the time of v2.

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `SmallStruct_Unmarshal`, case-insensitive | 114 ns | 149 ns | the config of the options of each call, which comes from a pool and is cleared after it (6%), and the loop of `v2StructDecoder`, which keeps more values across its calls than v1's, in the stack |
| `MediumStruct_Unmarshal`, case-insensitive | 770 ns, 7 allocations | 1.57 µs, 7 allocations | the values of no field, which jsontext reads for their names and their UTF-8 ( `skipValueV2`, 48%, a third of it the white space of the indented input at each token ), and the names of no field, kept for their duplicates (10%): work which v1 doesn't do |
| `LargeStruct_Unmarshal`, case-insensitive | 10.1 µs, 66 allocations | 24.6 µs, 66 allocations | the names of no field, kept for their duplicates ( `nameSet.add`, 29%: the copy of each name, and the comparison with the names of its object where its bit is one of theirs ), the values of no field (30%), and the length of the names without '_' and '-' ( `foldedLength`, 9% ) |
| `GitHubGraphQL_Unmarshal`, case-insensitive | 60 µs, 135 allocations | 61.5 µs, 135 allocations | not looked into since the plain names of another case are found by v1's table as they are read |
| `InterfaceUnmarshaler_Unmarshal` (`UnmarshalJSON` of the values of interfaces) | 191 ns, 8 allocations | 269 ns, no allocation | the value of each method read and checked by `ReadValue` (40%), and `reflect.PointerTo` of the type of the pointer which the interface holds (14%), which the program has no pointer type of; `DecoderOfV2` (7%) |
| 1.5 MB of records into `[]struct{}`, by `UnmarshalRead` | | 1.6 MB and 48 allocations | not looked into since the names of no field are kept in one stack of the call (3.9 MB and 71,903 allocations before) |

`LargeSlice_EscapedString`, `Times_Unmarshal`, `CodeAny_Unmarshal`, `TwitterGeneric_Unmarshal`,
`TwitterBinding_Unmarshal` and `Code_Unmarshal` are within 2% of v1, and `GitHubREST_Unmarshal` and
`SmallStruct_ChunkedStream` are faster than v1.

`UnmarshalRead` and `UnmarshalDecode` against v1's `Decoder`, from a reader of 4 KiB at a time (the
`ChunkedStream` benchmarks; the stream reads 512 bytes at a time in both):

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `Code_ChunkedStream` | 2.37 ms | 2.67 ms | the objects which a read cuts, decoded member by member ( `decodeStructElementsV2` ): for each member the name copied for the path of its errors, the bytes kept from the token before its value, and the end of a value which is not a scalar looked for before it is decoded; small costs, none of which is more than 5% |
| `Records_ChunkedStream` | 4.27 ms | 4.88 ms | the records which a read cuts, about one in three, decoded member by member as in `Code_ChunkedStream`; the records which the buffer holds whole are decoded as fast as v1's run of them ( `decodeRun` ) |
| `SkipRecords_ChunkedStream` | 1.89 ms | 4.64 ms | a value of no field which the buffer doesn't hold whole is skipped by a `jsontext.Decoder`, token by token as encoding/json/jsontext's `SkipValue` reads it, which reads it for its names and its UTF-8 as encoding/json/v2 does. A skip by the scan of `ReadValue`, which would drop what it read, takes 30-40% off by an estimate, as encoding/json/jsontext documents `SkipValue` ( "semantically equivalent to calling ReadValue and discarding the result" ); but after a `SyntacticError` the decoder would not be where encoding/json/jsontext's is ( `InputOffset`, `StackDepth` ), and the error may have another place. Left for later by decision (2026-10-24) |
| `CodeAny_ChunkedStream` | 4.14 ms | 5.37 ms | the objects which a read cuts, decoded member by member: each scalar of an interface{} looked for its end and decoded whole. The value of each member is allocated (61,000 allocations against 29,500); decoding it in a slot of the stack of the context instead took none of the time off (2026-10-24) |
| `Twitter_ChunkedStream` | 11.0 µs | 12.7 µs | the ones of `Code_ChunkedStream` |
| `NDJSON_ChunkedStream` (a record by `UnmarshalDecode`, against `Decoder.Decode`) | 2.48 ms | 3.63 ms | the setup of each call (about 14%): two pools, the hooks of jsontext, the path of the stream and the copy of the buffer of the decoder; and the records which a read cuts, as in `Code_ChunkedStream` |
| `MediumStruct_ChunkedStream`, case-insensitive | 1.68 µs | 2.64 µs | the ones of `MediumStruct_Unmarshal` |

## Constraints

- v2 code inside the functions of v1 slows v1 even when it is skipped: a reassigned local across a
  call, one more live variable in a scan loop, more fields in `RuntimeContext`. v2 runs through
  functions of its own; `decodeStringField` stays as it is (even cold v2 branches in it cost v1's
  `LargeSlice_EscapedString` 3-4% on arm64).
- Helpers of the hot types ( `RuntimeContext`, `Stream` ) are functions, not methods: a method
  moves the data of the binary, which slowed `CodeAny_Unmarshal` by 2%.
- Streams read 512 bytes at a time at most, with bounded memory; skipped values are checked as
  encoding/json/v2 checks them.

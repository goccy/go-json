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
  compared by their lengths and last bytes first, since 2026-10-24; v2 measured without options on
  structs with the tags of the names of their payloads, as encoding/json/v2's own benchmarks
  compare v2 with v1; the names of no field kept by their places in the input and compared by
  their hashes, the values of no field skipped in one read, by v1's skip when both
  `AllowDuplicateNames` and `AllowInvalidUTF8` are set, the members of no field skipped in a loop
  of their own, the '_' and '-' of folded names counted as they are read, and the options of
  `Unmarshal` and `UnmarshalRead` in the state of their call, since 2026-10-25; the members of a
  struct decoded by a loop of the options, chosen once for the object, which with
  `AllowDuplicateNames` and `AllowInvalidUTF8` is v1's, the setup of a call of `Unmarshal` without
  options with fewer calls, and most names of no field added to their set by `nameSet.addListed`,
  which makes no call, also since 2026-10-25; the members of the objects of a stream decoded by a
  loop chosen once and found by what `nextMember` read, the whole elements of a slice of a stream
  decoded as a run, the scalars of a stream, and of a nil `interface{}`, decoded in its buffer, the
  objects and the arrays of a stream into new `interface{}` values by stacks, a long top-level
  value decoded element by element at once, and the values of `UnmarshalDecode` read in place in
  the buffer of the `jsontext.Decoder`, decoded there whole, with the levels of their place found
  only where a call reads them, since 2026-10-26

## Goal

`github.com/goccy/go-json/v2` decodes every payload as fast as the root package (v1), within the
noise of the measures; it doesn't have to be faster. v1 must not get slower from code shared with
v2. Then faster than sonic, where v1 is (decided 2026-10-23: the payloads which v1 decodes slower
than sonic are not a target of this work). Being faster than master is not the measure: master
has the gaps below.

## Gaps

Medians of 6 rounds of each binary in turn, arm64 (Apple M5), 2026-10-25, where no other date is
given. The causes come from CPU and allocation profiles in a linux/arm64 container; the
percentages are of the time of v2.

v2 is measured without options, on structs with the tags of the names of their payloads, which v2
matches as they are: v1 matches the names to the fields without tags regardless of their case,
which v2 does only with `MatchCaseInsensitiveNames`, which also ignores '_' and '-'. The rows also
give v2 with `AllowDuplicateNames` and `AllowInvalidUTF8` ("as v1", benchmarks which are not in
the repository), which decodes these payloads as v1 does, and v2 with `MatchCaseInsensitiveNames`
("case-insensitive", the benchmarks of the suffix `CaseInsensitive`).

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `SmallStruct_Unmarshal` | 112 ns | 114 ns; 116 ns as v1 (2026-10-26); 151 ns case-insensitive | in a profile, against v1: the setup of the call, about 3 ns more (`beginCall`, `endCall` and `DecoderOfV2`, calls which are not inlined), the strings, about 2 ns more (`decodeStringV2` against `decodeStringField`), and the bits of the fields, kept for their duplicates, about 2 ns; as v1, the options which the call applies |
| `MediumStruct_Unmarshal` | 762 ns, 7 allocations | 930 ns; 775 ns as v1 (2026-10-26); 1.11 µs case-insensitive; 7 allocations | the names of no field, kept for their duplicates, and the values of no field, whose strings are read for their UTF-8 and whose objects for their names: work which v1 doesn't do |
| `LargeStruct_Unmarshal` | 10.0 µs, 66 allocations | 12.6 µs; 9.90 µs as v1 (2026-10-26); 16.6 µs case-insensitive; 66 allocations | the names of no field, kept for their duplicates (`nameSet.addListed`, 15%: 1,028 names, each about 1.9 ns, mostly the loads of the name, which the hash depends on, and of the word of the mask, which depends on the hash; a copy of its code in the loop of the members took about 0.45 µs off, not kept, so as not to have the code twice), and the values of no field (about 1.05 times v1's skip of them) |
| `TwitterBinding_Unmarshal` | 5.27 µs | 5.34 µs; 5.40 µs as v1, -0.8% in another layout (2026-10-26) | not looked into |
| 1.5 MB of records into `[]struct{}`, by `UnmarshalRead` | | 1.6 MB and 48 allocations, measured before 2026-10-25 | not looked into since the names of no field are kept in one stack of the call (3.9 MB and 71,903 allocations before) |

`CodeAny_Unmarshal`, `Code_Unmarshal`, `TwitterGeneric_Unmarshal`, `GitHubREST_Unmarshal` and
`Times_Unmarshal` are within 1% of v1 or faster, as v1 and without options (2026-10-26).
`GitHubGraphQL_Unmarshal` is 1.5-1.9% slower than v1 as v1, and `LargeSlice_EscapedString`
2.9-3.4% (183 µs against 178 µs; master 176 µs): commits which don't touch their paths move them
by up to 4% (measured at six commits of 2026-10-26: `LargeSlice_EscapedString` was 170 µs, faster
than v1, at one of them), the layout of the binary. `InterfaceUnmarshaler_Unmarshal` (167 ns;
154 ns as v1; v1 192.5 ns) and `SmallStruct_ChunkedStream` (172 ns; v1 286 ns) are faster than v1.

A deep input costs the collector more for each level: it scans the whole stack of the goroutine
in each of its cycles, and with `GOGC=1` its cycles grow with what the decoding allocates. The
input of `Benchmark_Decode_DeepDuplicateNames_Unmarshal_GoJsonV2`, whose deepest object has a name
twice, decoded into an `interface{}` again by the merging of v2 with `AllowDuplicateNames`, takes
32 times as long 4,000 levels deep as 250 levels deep with `GOGC=1`, and 16 times without the
collector; encoding/json/v2 takes 30 to 38 times. A level of the merging is three frames of about
600 bytes (`decodeElem`, `decodeMerging` and `decodeObjectInto`), which fewer frames would make
cheaper (measured 2026-10-25).

`UnmarshalRead` and `UnmarshalDecode` against v1's `Decoder`, from a reader of 4 KiB at a time (the
`ChunkedStream` benchmarks; the stream reads 512 bytes at a time in both):

| Benchmark | v1 | v2 | Cause found |
|---|---|---|---|
| `Code_ChunkedStream` | 2.36 ms | 2.57 ms; 2.36 ms as v1 (2026-10-26) | the objects which a read cuts, decoded member by member ( `decodeStructElementsV2` ): for each member the name copied for the path of its errors, the bytes kept from the token before its value, and the end of a value which is not a scalar looked for before it is decoded; small costs, none of which is more than 5% |
| `Records_ChunkedStream` | 4.86 ms | 5.44 ms; 4.89 ms as v1 (2026-10-26) | the records which a read cuts, about one in three, decoded member by member as in `Code_ChunkedStream`; the records which the buffer holds whole are decoded as fast as v1's run of them ( `decodeRun` ) |
| `SkipRecords_ChunkedStream` | 1.94 ms | 4.58 ms; 1.92 ms as v1 (2026-10-26) | a value of no field which the buffer doesn't hold whole is skipped by a `jsontext.Decoder`, token by token as encoding/json/jsontext's `SkipValue` reads it, which reads it for its names and its UTF-8 as encoding/json/v2 does. A skip by the scan of `ReadValue`, which would drop what it read, takes 30-40% off by an estimate, as encoding/json/jsontext documents `SkipValue` ( "semantically equivalent to calling ReadValue and discarding the result" ); but after a `SyntacticError` the decoder would not be where encoding/json/jsontext's is ( `InputOffset`, `StackDepth` ), and the error may have another place. Left for later by decision (2026-10-24) |
| `CodeAny_ChunkedStream` | 4.20 ms | 4.80 ms; 4.42 ms as v1 (2026-10-26) | as v1, the main goroutine's (+6% with `GOGC=off` too, not the collector's): about 2.5% is the lookup of the names of each object, which decodes a name twice into its first value with `AllowDuplicateNames` as encoding/json/v2 does, and errors for the kinds which don't merge, so that every member is looked up (`anyNames.index`, about 0.5 ns a name; without it, a measure only, +2.7-3.3%); the rest, about 3%, is spread over the member loop of `decodeNewAnyObjectV2` (two keeps of the bytes for the places of errors, the name and the count of the level of the path), the dispatch of the scalars, the wrapper `v2InterfaceDecoder.Decode` and the maps, in a native profile (xctrace); decoding the scalars by `decodeScalar` in the loops, without `decodeNestedV2`, took none of it off. Left for later by decision (2026-10-26); a way not tried is the places of errors in interface{} values made as the error goes up the levels, as v1 makes its paths, in place of the path kept for each member |
| `Twitter_ChunkedStream` | 11.0 µs | 12.1 µs; 9.90 µs as v1 (2026-10-26) | the ones of `Code_ChunkedStream` |
| `NDJSON_ChunkedStream` (a record by `UnmarshalDecode`, against `Decoder.Decode`) | 2.47 ms | 2.84 ms; 2.61 ms as v1 (2026-10-26) | as v1, the decoding of the records is at parity; the gap is the bridge of each call between the `jsontext.Decoder` and the stream of v2, which v1's `Decoder`, owning its stream, doesn't pay: about 23 ns a record for the records which a read cuts (a third of them: after a cut record the stream reads 512 bytes at a time through the decoder, about 2.7 records a read), decoded member by member (the pool of the state of the call, the levels of the place for the path of the stream, `BorrowV2`, `EndV2`, `EndValue`, the reads through `FetchInput`), and about 8 ns for the ones in the buffer of the decoder. Trims of the bridge are within the noise now. Left for later by decision (2026-10-26); the way found to close it is a top-level `jsontext.Decoder` owning the stream and the state of its calls, as v1's `Decoder` owns its stream, which costs 5-10 KB held by each decoder of two values or more (v1: about 1-2 KB). A state of the calls kept only by the decoder, through the pool, gave 1-1.5% and held what the pool held for two collections for the life of the decoder (tried and dropped 2026-10-26) |
| `MediumStruct_ChunkedStream` | 1.67 µs | 2.22 µs; 1.55 µs as v1 (2026-10-26); 2.35 µs case-insensitive | the ones of `MediumStruct_Unmarshal`, and the ones of `Code_ChunkedStream` |

## Constraints

- v2 code inside the functions of v1 slows v1 even when it is skipped: a reassigned local across a
  call, one more live variable in a scan loop, more fields in `RuntimeContext`. v2 runs through
  functions of its own; `decodeStringField` stays as it is (even cold v2 branches in it cost v1's
  `LargeSlice_EscapedString` 3-4% on arm64).
- Helpers of the hot types ( `RuntimeContext`, `Stream` ) are functions, not methods: a method
  moves the data of the binary, which slowed `CodeAny_Unmarshal` by 2%.
- Streams read 512 bytes at a time at most, with bounded memory; skipped values are checked as
  encoding/json/v2 checks them.

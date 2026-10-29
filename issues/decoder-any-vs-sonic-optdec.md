# Decoding into interface{} slower than sonic's optdec

- Area: decoder
- Kind: speed
- Priority: 2
- Status: recorded 2026-10-29; accepted by decision for the fix of issue 775

## What

Since the fix of issue 775, every string decoded into interface{} is a copy of its own and every number and string
is boxed by itself, as encoding/json does, so that a kept value keeps alive no memory but its own. Decoding into
interface{} allocates more then: Benchmark_Decode_TwitterGeneric_Unmarshal_GoJson makes 691 allocations per
operation instead of 135 ( 307 keys of 74 distinct names, 107 strings and 56 numbers, a key each, a string's bytes
and its box each, a box per number ), and takes 24% more time on an Apple M5 ( CodeAny_Unmarshal 25% more ).

On the amd64 machines of the CI ( 2026-10-29 ), GoJson against SonicStd with sonic's optdec decoder: 1.05 on an
EPYC 9V74, 1.09 on an EPYC 7763 with a live heap of 64 MB ( 0.90 and 0.80 before the fix ); 0.62-0.68 against the
other decoders of sonic. sonic's optdec counts the values of the document first and allocates one pool per document
for the numbers, one for the string headers and one for the elements of every array ( newEfacePool, sonic v1.15.4
internal/decoder/optdec/node.go ): a value kept out of a document keeps the decoded data of the whole document
alive, which is what the fix of issue 775 took out of go-json. Its copied strings are one allocation each, as
go-json's.

Benchmark_Decode_SmallStruct_ChunkedStream_GoJson, which makes a new Decoder for each value, copies the strings of
the value now: 1.02 of SonicStd on an EPYC 7763 ( 0.85 before ), 0.50-0.83 on the other machines; 1.03 on a Xeon
8370C, where it was 1.02 before.

## Measured ideas ( Apple M5, change of TwitterGeneric / CodeAny from before the fix )

- The fix as it is: +24% / +25%.
- A cache of 256 strings per runtime context kept from a call to the next, as encoding/json/v2's stringCache: +23%,
  and slower than the fix on inputs whose strings vary ( Code_Unmarshal +7.9% against +3.8%, GitHubREST +11.0%
  against +7.9% ): it pays only for an input decoded again and again.
- The plain keys of 2 to 16 bytes of map[string]interface{} looked up in a table of the call by their length and
  their first and last eight bytes, the table cleared at the end of the call: +21% / +18%. A key is as short as a
  string of the tiny allocator, whose allocation costs about as much as the lookup.
- The keys, and the boxes of the numbers and the strings in slabs shared by the values of a call: +17% / +6%; a kept
  value keeps the boxes of its call alive then.

The prototypes and their raw numbers are kept outside the repository with the other measurements of the fix.

## To do

Find speed elsewhere in the decoding into interface{} ( the building of the maps, the skip of white space of
indented inputs ) without sharing memory between the values.

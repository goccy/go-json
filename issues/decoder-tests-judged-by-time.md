# Tests which judge by elapsed time

- Area: decoder (tests)
- Kind: maintenance
- Priority: 3
- Status: recorded 2026-10-25, when the test of the depth of names twice of v2 was replaced by a
  benchmark

## What

Two unit tests of the root package pass or fail by a ratio of times, which the machine, its load
and the collector move as much as the code does: a test of decoding with `GOGC=1` on Windows once
measured 167 times for a ratio of 16 of code which was linear, as the collector scans the stack of
a deep decoding in each of its cycles.

- `TestDecodeStreamLongNumber` (`decode_stream_memory_test.go`): skipping a number of 1 MB must
  take at most 50 times as long as skipping a string of as many bytes, from a reader which has no
  `Len`.
- `TestDecodeTypeErrorPathLinear` (`decode_type_error_test.go`): a value 4,000 levels deep with a
  type error at its bottom must take at most 15 times as long as the same value without it, for
  `Unmarshal` and for decoders of 1-byte and 512-byte reads.

## Fix

What they guard is the time of a decoding against its input: a benchmark of each, which the CI
compares with master (`benchcheck`, a benchmark 15% slower fails), as
`Benchmark_Decode_DeepDuplicateNames_Unmarshal_GoJsonV2` replaced the test of the depth of names
twice. Keep in the tests what doesn't depend on time (the values decoded, the errors).

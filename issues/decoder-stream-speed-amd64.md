# Stream decoding slower than sonic on amd64 for values cut by reads

- Area: decoder
- Kind: speed
- Priority: 3
- Status: open; the speed work of #736 was ended by decision on 2026-10-13

## What

Since #736 the stream decoder reads 512 bytes at a time and decodes a value that a read cuts
element by element. On amd64 (measured on 97e3f4a, the last commit of #736):

- `NDJSON_ChunkedStream`: 11-16% slower than sonic on EPYC 7763 and the Xeons;
- `Twitter_ChunkedStream`: 5-45% slower than sonic on every CPU (master was already slower there);
- `LongStrings_ChunkedStream`: 4-10% slower than master (21% faster on arm64);
- `SmallStruct_ChunkedStream`: 5% slower than sonic on EPYC 9V45 (master 4%).

On arm64 every stream benchmark is faster than sonic.

## Causes found

- The element path costs about 1.8-2.1x the whole-value path per byte: key scan and lookup are
  split, numbers are parsed twice, and `bracketIndex.extend` adds 256 bytes per nested lookup.
- 37% of the NDJSON records are cut by a 512-byte read.
- The element path scans strings 8 bytes at a time, where the whole-value path uses AVX2 for
  strings over 64 bytes.

## Dropped approaches (measured)

- Read on once for a cut top-level value that fits the buffer: NDJSON 7-22% faster, but
  `LongStrings` 36-41% slower (slower than sonic).
- Decode runs of members up to the last comma in the buffer: NDJSON only 1.4% faster, other
  stream benchmarks 8-14% slower, `Unmarshal` 1-7% slower from the shared loop.

## Constraint

Reads stay at most 512 bytes (smaller is the goal), memory stays bounded.

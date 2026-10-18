# Map encoding: v2 slower than v1, and map[int]int slower than sonic

- Area: encoder
- Kind: speed
- Priority: 2
- Status: open

## What

The sorted-map benchmarks of #732 showed (2026-10-11):

- v2 maps with `MarshalText` keys about 3x slower than v1 (`Text20`: 5558 vs 1819 ns);
- v2 with `jsontext.Multiline` and `WithIndent` 1.3-2.9x slower than v1;
- `SonicStd` faster than go-json on `map[int]int` of 100 entries (7471 vs 8282 ns).

The first two break "v2 is at least as fast as v1", the last "never slower than sonic".

Re-measure on master before starting.

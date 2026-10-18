# Win back BenchmarkCodeUnmarshal

- Area: decoder
- Kind: speed
- Priority: 3
- Status: open; accepted in #657 on the condition that it is won back

## What

#657 (removing the runtime linknames) made `BenchmarkCodeUnmarshal` 22% slower on amd64
(1.11 to 1.35 ms) and 28% on arm64. #658 and #659 won part of it back; about 12% was left on
amd64 after #659 (2026-09-25).

## Cause

Not CPU at first: before #657 the pools kept the decoded values of the previous call alive, so
the heap grew and the GC ran less often. With the pools cleared, the benchmark is GC-bound:
13.5k allocations per operation, mostly one `*codeNode` per node.

## Ideas

- Allocate less: `Unmarshal` still copies the input for the trailing NUL and in-place unescaping.
- Struct key matching, float and int parsing were the largest CPU parts in a linux profile.
- Allocating pointees in slabs per type was not done: it breaks `runtime.SetFinalizer` on
  decoded pointers.

Re-measure on master before starting: #736 and #740 changed the decoder since.

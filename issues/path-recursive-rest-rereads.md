# Extract reads a value once for each name of .. around it which has a rest of the path

- Area: path
- Kind: speed
- Priority: 2
- Status: recorded 2026-10-28 (left by the fix of #774, which made .. search the values it selects)

## What

`Extract` ( and `Path.Unmarshal` ) of `..name` followed by more of the path ( `$..b.c`, `$..b.b`, `$..b[0]` ) reads
the value of each `name` twice: once with the rest of the path, and once more to search it for `name`. A value
nested in `n` of them is read `n` times, so the time grows with the size times the depth of the nesting: 165 ms for
`$..b.c` of 2000 objects nested by `"b"` around a string of 1 MB ( 1 MB in all ), where `$..b`, which has no rest and
reads a value once, takes 4 ms and `Path.Get` of the decoded value 1 ms ( Apple M5, 2026-10-28 ). The nesting is
up to 10000, the limit of the decoder.

`Path.Unmarshal` of any `..` also decodes each value it selects on its own: values nested in each other are decoded
once for each of them, which is what the result holds.

## Fix

Decode by the set of the nodes of the path which are at a value, instead of one: a member of an object takes the
nodes which each of them leads to ( `..` stays, and adds the rest of the path at its name ), and the value is read
once with all of them. Keep a path without `..` on the one node it has now, whose speed
`Benchmark_Decode_SmallStruct_UnmarshalPath_GoJson` measures.

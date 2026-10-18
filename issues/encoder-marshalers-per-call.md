# v2 WithMarshalers recompiles for every call

- Area: encoder
- Kind: speed
- Priority: 3
- Status: open; plan agreed on 2026-10-10

## What

Measured on master after #722 (arm64, `*struct{A, B, C int; S string}`, a function on the
string field):

| Case | Time | Allocations |
|---|---|---|
| no options | 27 ns | 1 |
| `MarshalJSON` method on the field | 38 ns | 2 |
| `WithMarshalers`, one value reused | 71 ns | |
| `WithMarshalers` built for each call, as std's examples write it | 10.5 µs | 377 |

The code sets are cached per `Marshalers` value (`internal/encoder/funcs_v2.go`), so a value
built per call compiles every time; a reused one costs the options path and a `sync.Map` lookup.

## Plan

- Compile `T` in the form "T has a function added at run time": the opcode refers to the
  function by its index, and calls the function at that index in the call's list
  (`ctx.Option.Funcs`).
- Key the opcodes by the argument types of the functions (and `MarshalFunc` vs
  `MarshalToFunc`), not by the `Marshalers` value. They live in the normal caches like any
  other compile mode; no separate map per `Marshalers`.
- Measure the four cases above before and after; the default path stays unchanged.

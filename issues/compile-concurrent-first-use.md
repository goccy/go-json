# Concurrent first use of a type compiles it once per goroutine

- Area: encoder, decoder
- Kind: speed
- Priority: 4
- Status: recorded 2026-10-29

## What

The opcodes of the encoder and the decoder of a type are compiled outside any lock, and the
first result stored in the shared table is the one every goroutine uses
(`compileToGetUnfilteredCodeSet` in internal/encoder/compiler.go, `CompileToGetDecoder` in
internal/decoder/compile.go, `typeCache.store` in internal/runtime/type_cache.go). When
goroutines use a new type at the same time, each of them compiles it, and all but one of the
results are dropped.

Measured on 2026-10-29 ( Apple M5, 1000 struct types used in the same order by every
goroutine, a fresh process ): 40.9 ms of wall time and 41 ms of CPU with one goroutine,
51.5 ms and 408 ms with ten. It is a cost of the start of a program only: once a type is
compiled, it is looked up without a lock. encoding/json compiles concurrently too, at a lower
cost per type.

## To decide

Whether the cost is worth a lock or a once per type, which would make the goroutines wait for
the compile of the first one instead.

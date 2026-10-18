# Valid is slower than sonic on long strings

- Area: decoder
- Kind: speed
- Priority: 2
- Status: open

## What

`Valid` of a value with long strings is slower than sonic (measured 2026-10-13):

| Machine | String length | go-json | sonic |
|---|---|---|---|
| amd64, Xeon 8573C | 604 B | 204 ns | 149 ns |
| amd64, Xeon 8573C | 8300 B | 1815 ns | 356 ns |
| arm64 | 8300 B | 800 ns | 429 ns |

## What was tried

The external pull request #702 (closed) made `skipFast` call a non-inlined `indexStringPlain`,
which uses AVX2 when at least 256 bytes are left in the buffer. Long strings got 33-77% faster,
but on EPYC 7763 `SkipRecords` got 13-21% slower, `Valid` of Code/Medium/Large/Twitter 17-27%
slower and `MediumStruct_Unmarshal` 6-7% slower: the call makes `skipFast` spill its registers
(stack stores on amd64 went from 5 to 23).

Other shapes measured:

- stop the word loop after 256 bytes of a string and hand over to `skipString`: 73% faster at
  8.3 KB, but 2-6% slower on payloads of short strings on amd64;
- a precomputed bound `min(length, start+256)`: same code on arm64, 1-6% slower on amd64 from
  one more live register;
- a bound of 128 bytes: 27-36% slower on 220-byte strings.

## Constraint

A shape that leaves the register use of `skipFast`'s string loop on amd64 unchanged; no
benchmark slower than master.

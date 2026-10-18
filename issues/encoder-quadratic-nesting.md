# Quadratic time for long pointer chains and nested MarshalEncode calls

- Area: encoder
- Kind: speed
- Priority: 3
- Status: recorded 2026-10-08 and 2026-10-11; re-check on master before fixing

## What

- A value that is a long chain of `*any` encodes in quadratic time, in v1 and v2:
  100,000 pointers took 1.2 s, encoding/json 43 ms.
- A chain of methods each writing `[`, calling `MarshalEncode` on the next value and writing
  `]`, nested N deep (arm64):

  | N | go-json | encoding/json/v2 |
  |---|---|---|
  | 2000 | 78-105 ms | 0.16 ms |
  | 4000 | 520 ms | 1.4 ms |
  | 1000, Multiline | 0.65 s | 2 ms |
  | 2000, Multiline | 4.1 s | 7 ms |

## Causes found for MarshalEncode

1. Each nested call writes its whole value into the method's encoder by `WriteValue`, which
   scans it again and grows the level stack: O(N²) in total.
2. With formatting, the base offset of the call is computed eagerly (`textcoder.Place`,
   `offsetAt`, `formattedOffset`), which formats the output so far on every call: O(N³).

## Ideas

- Append the engine output to an attached encoder without the scan when it is not at a name,
  the place is valid and `ctx.Deep` is unset.
- Compute the base offset only when an error or `OutputOffset` needs it.

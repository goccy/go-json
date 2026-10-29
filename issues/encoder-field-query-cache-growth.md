# Opcodes of every field query kept for good

- Area: encoder
- Kind: memory
- Priority: 2
- Status: recorded 2026-10-29; a bound needs a design decision

## What

`MarshalContext` with a field query compiles the opcodes of the query for the type and keeps
them in the type's `OpcodeSet` (`queryCache`, by the hash of the query), which is never
emptied. A program which builds its queries from its requests ( the fields a client asked
for ) keeps the opcodes of every query it ever met: about 6.2 KB for each distinct query of
a struct of 16 fields, measured on 2026-10-29 ( 4.5 KB with v0.10.6, which kept them so
too ).

The reads of the cache take no lock since the fix of issue 775; only its size is left.

## To decide

How the cache is bounded: by a number of the queries of a type, by their bytes, or by
dropping the ones which were not used for a while, and what a query which is not in the
cache costs then ( a compile of its opcodes ).

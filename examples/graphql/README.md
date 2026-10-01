# A GraphQL server on json.Selection

This example asks whether a GraphQL server built on go-json's selection ( `json.Selection` ) can compete with
the server which [gqlgen](https://github.com/99designs/gqlgen) generates. Both servers serve the same schema
( [schema.graphqls](schema.graphqls) ) and the same data ( [store](store) ), and `TestSameResponses` checks
that every variant writes the same bytes as gqlgen ( the errors, which gqlgen adds in the order they happen,
are compared as a set ).

## How gqlgen writes a response

Read in gqlgen v0.17.95 and its generated code:

1. Execution builds a tree of `graphql.Marshaler`: a `FieldSet` per object, and per field a `FieldContext`, the
   value boxed in `any` and a closure ( `WriterFunc` ) which writes it. A resolver field runs in a goroutine of
   its own ( `FieldSet.Concurrently`, `MarshalSliceConcurrently` ), which is what lets a data loader gather the
   fetches of a level.
2. The tree is written token by token through `io.Writer`, with gqlgen's own escaping of strings.
3. The written data is put in `Response{Data: json.RawMessage}` and encoded by `encoding/json`, which scans and
   copies the whole data again ( and escapes `&`, `<`, `>`, which gqlgen's own writer doesn't ).

## How the go-json engine works

[gql](gql) is a small GraphQL engine, a proof of concept: the types of the schema are plain Go structs whose
JSON keys are the names of the GraphQL fields, and a field with a resolver is a `gql.Field[R]`
( [models.go](gojsonserver/models.go) ). There is no code generation. A request is served in two phases:

1. **Resolve.** The query is parsed and validated by gqlparser, and turned into a `json.Selection` by the
   CollectFields algorithm of GraphQL ( fragments, inline fragments, `@skip` / `@include`, merged fields,
   aliases ), which is kept with the query. The fields to resolve are found from the selection level by level
   over the whole response, and the resolvers of a level run concurrently: a data loader gathers their
   fetches as with gqlgen. A `BatchResolver` resolves the field of all the values of a level by one call,
   since the whole level is known before it runs, without the wait of a data loader. A value reached twice
   by the same selected field is resolved once. The path of every field is kept, for the errors.
2. **Write.** The response is written by one call of `json.MarshalContext` with the selection: go-json
   compiles the code of a type for the selection once, and a `gql.Field` writes its resolved value where it is.

## Results

`go test -run '^$' -bench . .` on an Apple M5 ( arm64 ). Other processes loaded the machine ( a load average of
5-30 ), so the times are longer than on an idle machine and the ratios are what to read. The variants:

- `gqlgen`, and `gqlgen+dataloader` with [dataloadgen](https://github.com/vikstrous/dataloadgen), which the
  documents of gqlgen recommend;
- `go-json+batch`: the engine with the batch resolvers;
- `go-json`: the engine with every field resolved by a call of its own, concurrently;
- `go-json+dataloader`: the engine with the same data loaders as gqlgen.

**Resolvers fetching from a store of a latency of 1 ms** ( `BenchmarkServersWithLatency`, the data loaders
wait 1 ms for a batch ):

| query | server | time | fetches | allocations |
| --- | --- | ---: | ---: | ---: |
| `three_levels`: 10 users, 2 posts each, 2 comments each | gqlgen | 3.82 ms | 31 | 3,059 |
| | gqlgen+dataloader | 6.47 ms | 3 | 3,078 |
| | go-json+batch | 3.88 ms | 3 | 326 |
| | go-json | 3.89 ms | 31 | 480 |
| | go-json+dataloader | 6.33 ms | 3 | 469 |
| `friends_deep`: 20 users, their friends, the friends of those, their posts | gqlgen | 6.09 ms | 621 | 68,694 |
| | gqlgen+dataloader | 9.57 ms | 5 | 67,673 |
| | go-json+batch | 5.20 ms | 4 | 2,880 |
| | go-json | 5.25 ms | 145 | 3,705 |
| | go-json+dataloader | 8.98 ms | 4 | 3,484 |

- The resolve phase is as concurrent as gqlgen: every variant takes the time of gqlgen with the same way of
  fetching, and a data loader gathers the same batches.
- The batch resolvers fetch as few times as a data loader, in the time of the unbatched concurrent resolvers:
  the wait of the data loader is not needed when the level is known.

**The CPU of a request** ( `BenchmarkServers`, one request at a time, means of 3 runs, no latency ):

| query | gqlgen | go-json+batch | faster | allocations gqlgen → go-json+batch |
| --- | ---: | ---: | ---: | ---: |
| `user`: one user, 7 scalar fields | 5.55 µs | 1.05 µs | 5.3x | 170 → 21 |
| `users_100`: 100 users, 5 scalar fields | 174 µs | 10.2 µs | 17x | 7,098 → 24 |
| `users_1000_wide`: 1000 users, 7 scalar fields | 1.88 ms | 0.24 ms | 8.0x | 90,360 → 25 |
| `posts_nested`: 50 posts, their authors, 5 comments with their authors | 577 µs | 50.4 µs | 11x | 25,112 → 465 |
| `three_levels` | 85.5 µs | 17.2 µs | 5.0x | 3,029 → 326 |
| `friends_deep` | 1.32 ms | 199 µs | 6.6x | 68,067 → 2,880 |
| `search_union`: a union with fragments and variables | 104 µs | 26.6 µs | 3.9x | 3,757 → 369 |
| `aliases_and_order` | 16.3 µs | 2.69 µs | 6.1x | 284 → 53 |
| `skip` ( `@skip` by a variable ) | 9.51 µs | 1.28 µs | 7.4x | 215 → 33 |

`BenchmarkServersParallel` ( the requests from all the CPUs ) keeps these ratios or larger ( `users_100` 24x,
`posts_nested` 22x ), as gqlgen's allocations cost it the GC.

## What this shows

A GraphQL server on go-json can compete with gqlgen: it answers the same, with the same concurrency and data
loaders, 3.9x-17x faster in the CPU and with 5x-3600x fewer allocations, and without code generation. The part
which only a JSON library can do is writing the response for a selection known in advance.

What go-json needed for it, added with this example:

- the order of a selection and aliases; the selection of the elements of a list, an array and a map ( issue #390 );
- `FieldOf` and `Select()` of no field; `SelectionField.With` and `SelectedFieldFromContext`;
- a marshaler whose method is on the pointer is called with the address of the value for a type of the size of
  a pointer too, as `encoding/json` does: the engine finds a resolved value by the address of its field.

What the JSON library could still do better, planned on the `encoding/json/v2` API ( `MarshalJSONTo` ):

1. A `gql.Field` writes its value by a nested `MarshalContext`, whose output the outer encoding checks again.
   Writing it in place ( `MarshalJSONTo(*jsontext.Encoder)` with the selection in the options ) removes the copy
   and the check: it is what separates `friends_deep` from `users_1000_wide` in the output rate.
2. A union is written value by value, each by `MarshalContext` with the selection of its type. A selection with
   type conditions would let the encoder switch the code by the type.
3. Errors are written as `null` with their paths, but the `null` of a non-null field is not propagated to its
   parent: the encoder could rewind its buffer to the nullable parent.

What a product would add besides: introspection ( `__schema` ), mutations and subscriptions, custom scalars and
enums, `@defer`, and the extensions of a server ( limits of complexity, tracing, persisted queries ).

## Running it

```
go test .                                  # the servers write the same responses
go test -run '^$' -bench . .               # the benchmarks
go run ./cmd/server -server go-json        # or -server gqlgen; -loader-wait 1ms for the data loaders
curl -s localhost:8080/query -H 'Content-Type: application/json' -d '{"query":"{ users(first: 2) { name } }"}'
```

The gqlgen server is generated by `go run github.com/99designs/gqlgen@v0.17.95 generate --config gqlgen.yml` in
[gqlgenserver](gqlgenserver). This module needs Go 1.26, as gqlgen does.

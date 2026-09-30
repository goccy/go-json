# A GraphQL server on json.Selection

This example asks whether a GraphQL server built on go-json's selection ( `json.Selection` ) is worth more
than the server which [gqlgen](https://github.com/99designs/gqlgen) generates. Both servers serve the same
schema ( [schema.graphqls](schema.graphqls) ) and the same data ( [store](store) ), and
`TestSameResponses` checks that they write byte-for-byte the same responses to the same queries.

## How the go-json server works

[gojsonserver](gojsonserver) is written from scratch, without code generation:

1. A query is parsed and validated by [gqlparser](https://github.com/vektah/gqlparser), as gqlgen does.
2. The operation is turned into a `json.Selection` by the CollectFields algorithm of GraphQL: the fragments
   and the inline fragments whose type condition applies are merged in, `@skip` and `@include` are
   applied, the fields of the same response key are merged, and an alias becomes the key the field is
   written with. The selection is kept with the parsed query ( and by the values of the variables of
   `@skip` / `@include`, if they use one ), so that go-json compiles the code for it once.
3. The types of the schema are plain Go structs whose JSON keys are the names of the GraphQL fields
   ( [models.go](gojsonserver/models.go) ). A field without arguments is a Go field. A field with arguments
   is a resolver: a value whose `MarshalJSON(context.Context)` is called only if the field is selected. It
   gets its arguments by `json.SelectedFieldFromContext`, which the selection carries
   ( `json.SelectionField.With` ), and writes what it resolves with `json.MarshalContext`, which applies the
   selection of the field.
4. The whole response is written by one call: `json.MarshalContext(ctx, query, json.WithSelection(sel))`.
5. The whole selection is known before anything is written, so the resolver of a list fetches the fields
   with resolvers that its elements select, once for the list, as the batch of a data loader does
   ( `requestState.prefetch` ).

A union is written by its resolver, which writes each value with the selection of its type.

## Results

`go test -run '^$' -bench . -count 5 .` on an Apple M5, as a native arm64 binary, while other processes kept
the load average around 15: the times of both servers are longer than on an idle machine, and the ratios are
what to read. A run on a quieter machine gave ratios of the same size ( `user` 8.5x, `users_100` 18x ).
The times are the means of the 5 runs.

One request at a time ( `BenchmarkServers` ):

| query | what it selects | gqlgen | go-json | faster | allocations gqlgen → go-json |
| --- | --- | ---: | ---: | ---: | ---: |
| `user` | one user, 7 scalar fields | 5.47 µs | 1.45 µs | 3.8x | 170 → 11 |
| `missing_user` | a user which is not found | 6.64 µs | 0.88 µs | 7.6x | 82 → 11 |
| `users_100` | 100 users, 5 scalar fields | 401 µs | 19.3 µs | 20.8x | 7,099 → 13 |
| `users_1000_wide` | 1000 users, 7 scalar fields | 4.17 ms | 0.44 ms | 9.4x | 90,362 → 14 |
| `posts_nested` | 50 posts with their author and 5 comments with their authors | 1.49 ms | 71.9 µs | 20.7x | 25,062 → 215 |
| `friends_deep` | 20 users, their friends, the friends of those, and their posts | 3.77 ms | 326 µs | 11.6x | 67,469 → 2,495 |
| `search_union` | a union with fragments on both types and variables | 207 µs | 13.5 µs | 15.4x | 2,160 → 64 |
| `search_posts_only` | a union with a fragment on one type | 31.9 µs | 2.98 µs | 10.7x | 293 → 29 |
| `aliases_and_order` | aliases, and fields in another order than the struct | 29.7 µs | 1.71 µs | 17.4x | 283 → 19 |
| `skip` / `include` | `@skip` / `@include` by a variable | 20.9 µs | 1.86 µs | 11.2x | 215 → 22 |
| `typename` | `__typename` at every level | 28.7 µs | 1.86 µs | 15.5x | 242 → 23 |
| `default_first` | the default value of an argument | 32.3 µs | 1.26 µs | 25.7x | 362 → 13 |
| `variable_first` | an argument given by a variable | 20.0 µs | 2.10 µs | 9.5x | 219 → 22 |
| `merged_fields` | the same field selected twice, merged | 20.0 µs | 1.45 µs | 13.8x | 230 → 15 |

The requests served from all the CPUs ( `BenchmarkServersParallel`, time per request ), where the GC of the
allocations runs beside the requests:

| query | gqlgen | go-json | faster |
| --- | ---: | ---: | ---: |
| `user` | 5.73 µs | 0.69 µs | 8.3x |
| `users_100` | 211 µs | 3.43 µs | 61x |
| `posts_nested` | 453 µs | 15.4 µs | 29x |
| `friends_deep` | 1.32 ms | 78.0 µs | 17x |
| `search_union` | 36.4 µs | 4.16 µs | 8.7x |

The resolvers fetching from a store of a latency of 1 ms, as from a database
( `BenchmarkServersWithLatency`, `{ users(first: 10) { name posts(first: 2) { title } } }` ):

| server | time | fetches |
| --- | ---: | ---: |
| gqlgen | 2.59 ms | 11 |
| go-json | 2.54 ms | 2 |

gqlgen runs the resolvers of the 10 users concurrently, 11 fetches in 2 rounds. The go-json server runs the
resolvers one by one while it writes the response, which took 13.8 ms, 11 fetches in a row, before the list
fetched the posts of its users in one batch: the selection tells the list what its elements need before they
are written.

## What this shows

- **The output is where a GraphQL server spends its time, and go-json writes it 4x-26x faster, 8x-61x under
  load, with 15x-6500x fewer allocations.** gqlgen writes every field of every object by a generated function
  through `graphql.Marshaler` values and an ordered field set. With a selection, go-json compiles the code of a
  type for the query once, and a value is written by the same VM as a plain `Marshal`: a scalar field costs no
  call and no allocation.
- **No code generation.** The schema is bound by the JSON keys of plain Go structs, and a resolver is a field
  whose type has `MarshalJSON(context.Context)`. The server is about 600 lines.
- **The GraphQL features the example uses fit the selection:** the order of the fields, aliases, fragments,
  inline fragments, unions, `@skip` / `@include`, `__typename`, arguments by literals, defaults and
  variables, and merged fields. `TestSameResponses` compares the bytes of the two servers.

What go-json needed for it, added with this example:

- the order of a selection and aliases ( GraphQL writes the fields in the order of the query );
- the selection of the elements of a list, an array and a map ( issue #390 );
- `FieldOf` and `Select()` of no field, for a type which a query selects nothing of ( `{}` );
- `SelectionField.With` and `SelectedFieldFromContext`, by which a resolver knows its arguments.

What a product built on it would still have to solve, found by writing the example:

1. **A resolver writes its value by a nested `MarshalContext`,** whose output the outer encoding checks
   again. It halves the speed of `friends_deep` compared with `users_1000_wide`, which has no resolvers. An
   interface by which a resolver returns a value for the encoder to write in place would remove both.
2. **A union is selected by its resolver,** which calls `MarshalContext` for each value with the selection
   of its type. A selection with type conditions would let the encoder do it.
3. **Errors.** An error of a resolver stops the whole encoding, while GraphQL writes `null` and an error
   with the path of the field. The encoder knows the key of the field, not the path, and doesn't propagate
   the `null` of a non-null field to its parent.
4. **Latency.** The resolvers run one by one while the response is written, so they have to fetch in
   batches, as the example does from the selection ( data loaders do the same for gqlgen ).
5. **Mutations, subscriptions, custom scalars and enums** are out of the example.

So a GraphQL server built on go-json is worth a product of its own: the part which the selection does is
an order of magnitude faster than generated code, and the pieces above are what that product would add.

## Running it

```
go test .                                  # the servers write the same responses
go test -run '^$' -bench . .               # the benchmarks
go run ./cmd/server -server go-json        # or -server gqlgen
curl -s localhost:8080/query -H 'Content-Type: application/json' -d '{"query":"{ users(first: 2) { name } }"}'
```

The gqlgen server is generated by `go run github.com/99designs/gqlgen generate --config gqlgen.yml` in
[gqlgenserver](gqlgenserver). This module needs Go 1.26, as gqlgen does.

# Selection: replace FieldQuery, resolvers for GraphQL

- Area: selection
- Kind: feature
- Priority: 3
- Status: pull request #691 open

## What

#691 replaces `FieldQuery` with `json.Selection` (`Select`, `Field`, `FieldOf`,
`ParseSelection`, `WithSelection`, `SelectionFromContext`, `SelectedFieldFromContext`,
`SelectionField.With` for arguments; order and aliases; lists, maps and interfaces pass through,
which fixes #390). `examples/graphql` compares a server on `Selection` with gqlgen: the same
bytes on 21 queries, 4-26x faster serially, 8-61x in parallel.

## Decided design of resolvers

Resolvers support concurrency and data loaders from the start, in two phases:

1. resolve: walk the selection level by level over the whole response and call the resolvers of
   a level concurrently (bounded);
2. write: one pass of the VM writes the resolved values; custom writers use `MarshalJSONTo`
   with the sub-selection in `enc.Options()`; errors become `null` and a path.

It was to wait for the v2 encoder and `MarshalJSONTo`, which are on master now.

## Related

- [selection-promoted-field-key.md](selection-promoted-field-key.md)
- A sub-selection on a field of a pointer type that points to its own struct type is not
  applied, the same as on `interface{}` fields; to be handled in #691.

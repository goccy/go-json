# StackDepth within MarshalJSONTo builds the whole stack

- Area: encoder (v2)
- Kind: speed
- Priority: 4
- Status: recorded 2026-10-20, found while `UnmarshalJSONFrom` was added

## What

`StackDepth` and `StackIndex` of the `jsontext.Encoder` which a `MarshalJSONTo` method or a
function of `MarshalToFunc` is given build the stack of the whole output for every call
(`stack.full`), which allocates; encoding/json/v2 allocates nothing. The decoder given to
`UnmarshalJSONFrom` takes its depth from its own stack and the number of the levels around it
(`stack.base`), but the encoder can't yet: an encoder attached for the key of a sorted map is
given no number of the levels around it (`textcoder.Attachment.OuterDepth`), and the comparison
of its stack with encoding/json/v2 (`TestSortedMapKeyNamesSameAsStd`) fails with it.

## Direction

Give every attach of the encoder its `OuterDepth`, then take `StackDepth` from
`depth() + base`, as the decoder does.

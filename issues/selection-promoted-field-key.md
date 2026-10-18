# FieldQuery cannot select a promoted field by its JSON key

- Area: selection
- Kind: defect or documentation
- Priority: 3
- Status: needs a decision: fix it, or document it

## What

A field promoted from an embedded struct can't be selected in a `FieldQuery` by the JSON key it
is written under. Either make it selectable by that key, or document that it isn't.
#691 replaces `FieldQuery` with `Selection`; decide there.

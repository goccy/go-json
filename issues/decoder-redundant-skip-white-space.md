# Skip white space once per value

- Area: decoder
- Kind: speed
- Priority: 4
- Status: open

## What

Every `Decode` of a type starts with `skipWhiteSpace`, although its caller (a struct field after
`:`, an array element after `,`, a map value, the stream, `Unmarshal`) has most often skipped
the white space already. #661 added more such calls: the non-inlined slow paths of
int/uint/float skip again from the start, and bytes, number, invalid, map entries, wrapped
strings and the type-error walk call it too.

## Plan

Define the contract that the cursor given to `Decode` points at the first byte of the value
(the caller skips once), and make the slow paths take the cursor already skipped.
A check for robustness goes only where the input can need it, never as an unconditional call
at the top of a hot function.

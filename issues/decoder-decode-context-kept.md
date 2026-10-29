# DecodeContext leaves its context to the next calls of the Decoder

- Area: decoder
- Kind: defect
- Priority: 3
- Status: recorded 2026-10-29

## What

`Decoder.DecodeContext` sets the context and `ContextOption` in the options of the stream
( decode.go ), which stay after the call. So the Decoder keeps the context alive until the
next `DecodeContext`, with what the context holds, and a later `Decode` gives that context
to the `UnmarshalJSON(context.Context, []byte)` methods of the values it decodes. v0.10.6 does
the same.

The context of `UnmarshalContext`, which the pooled runtime context kept the same way, is
dropped when the runtime context goes back to the pool since the fix of issue 775.

## Fix

Make the context and `ContextOption` the ones of the call only: restore the options after
`DecodeWithOption` returns.

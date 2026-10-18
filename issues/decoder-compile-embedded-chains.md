# Decoder compile time of deep chains of embedded structs

- Area: decoder
- Kind: speed
- Priority: 4
- Status: open; accepted as the cost of #743

## What

Since #743 each struct walks its embedded structs to choose its fields as encoding/json does.
A chain `Tk{Tk+1; Sk []Tk}` of 16 levels compiles in 98 µs, where it took 39 µs before
(master then chose some fields wrongly). Compiling happens once per type.

## Idea

An allocation-free `StructTagFromField`, shared with the encoder, which parses the same tags
when it chooses its fields.

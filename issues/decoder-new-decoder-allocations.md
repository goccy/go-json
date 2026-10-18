# Allocations of NewDecoder compared with v0.10.5

- Area: decoder
- Kind: speed
- Priority: 3
- Status: recorded 2026-10-11; to be measured

## What

`NewDecoder` followed by a `Decode` of a small value was recorded to allocate more than
go-json v0.10.5. Measure allocations and bytes per operation of master, v0.10.5 and
encoding/json before deciding anything. #736 changed the stream decoder since.

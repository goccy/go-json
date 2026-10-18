# Lost error for an embedded pointer to an unexported struct below another embedded pointer

- Area: decoder
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-15; re-check on master before fixing

## What

encoding/json fails with `json: cannot set embedded pointer to unexported struct: ...` when a
value would have to be stored through a nil embedded pointer to an unexported struct type.
When that pointer is embedded in a struct which is itself reached through an embedded pointer
to an exported struct, go-json loses this error. Shape:

```go
type inner struct{ X int }
type Mid struct{ *inner }
type Outer struct{ *Mid }
// Unmarshal([]byte(`{"X":1}`), &Outer{})
```

Found while #743 was compared with encoding/json's field rules; master had it before #743 too.

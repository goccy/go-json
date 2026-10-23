# Choice among fields whose names differ only in case

- Area: decoder
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-15; checked on master on 2026-10-23

## What

encoding/json matches a key to a field without regard to case when no field has the key
exactly. When several fields have names that differ only in case, one of them in an embedded
struct, and the key matches none exactly, go-json can choose a different field than
encoding/json does.

Found while #743 was compared with encoding/json's field rules; master had it before #743 too.
Shape, checked on master on 2026-10-23:

```go
type Inner struct{ AB int }
type Outer struct {
	Inner
	Ab int `json:"ab"`
}
// Unmarshal([]byte(`{"aB":1}`), &v)
// encoding/json: Outer.Ab is 1
// go-json:       Outer.Inner.AB is 1
```

Fields of one struct whose tags differ only in case are chosen as encoding/json chooses them.

## Note

The v2 option `MatchCaseInsensitiveNames` and the `case:ignore` tag option choose among the
same fields, so the fix may share code with the v2 struct decoder
(`internal/decoder/decode_v2_struct.go`).

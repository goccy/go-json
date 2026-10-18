# Choice among fields whose names differ only in case

- Area: decoder
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-15; re-check on master before fixing

## What

encoding/json matches a key to a field without regard to case when no field has the key
exactly. When several fields have names that differ only in case (for example tags `"ab"` and
`"AB"`) and the key matches none exactly (`"Ab"`), go-json can choose a different field than
encoding/json does.

Found while #743 was compared with encoding/json's field rules; master had it before #743 too.

## Note

The v2 option `MatchCaseInsensitiveNames` and the `case:ignore` tag option choose among the
same fields, so the fix may share code with [decoder-v2-api.md](decoder-v2-api.md).

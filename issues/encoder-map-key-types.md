# Map key types that v1 encodes differently from encoding/json

- Area: encoder
- Kind: defect
- Priority: 3
- Status: recorded 2026-10-12 (left by #735); re-check on master before fixing

## What

v1 `Marshal`, compared with encoding/json:

- keys whose type has only `AppendText` (`map[*AT]`, `map[**AT]`) are unsupported;
  for `map[**AT]` go-json calls `MarshalText` where encoding/json calls `AppendText`;
- `map[*float64]` is unsupported; encoding/json writes `{"1.5":1}`;
- a named pointer to a type whose pointer has `MarshalText` (`type PT *T`; `map[PT]`,
  `map[*PT]`) is unsupported;
- `map[PIP]` (`type PIP *IP`, `IP int` with a pointer-receiver `MarshalText`) is named `"3"`
  where encoding/json uses the text;
- `map[*BJ]` (`BJ` has a value `MarshalJSON` and `MarshalText`) is named by `MarshalJSON`;
- keys of 255 or more levels of pointers fail with "unsupported type" where encoding/json names
  them;
- the error of a failing key marshaler names the value type where encoding/json names the
  pointer type (`map[EV]`, `map[*time.Time]` with the year 10000).

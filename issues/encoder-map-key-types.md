# Map key types that v1 encodes differently from encoding/json

- Area: encoder
- Kind: defect
- Priority: 3
- Status: recorded 2026-10-12 (left by #735); checked on master 9adc5cb on 2026-10-28

## What

v1 `Marshal`, compared with encoding/json:

- keys of a pointer to a type which has only `AppendText` (`map[*AT]`, `map[**AT]`, `AT int`) are
  written as the value they lead to (`{"1":1}`), where encoding/json calls `AppendText`
  (`{"at":1}`); with `MarshalText` too (`map[*ATM]`, `map[**ATM]`), go-json calls `MarshalText`
  where encoding/json calls `AppendText`;
- `map[*float64]` is unsupported; encoding/json writes `{"1.5":1}`;
- a named pointer to a type whose pointer has `MarshalText` (`type PT *T`; `map[PT]`,
  `map[*PT]`) is unsupported;
- `map[PIP]` (`type PIP *IP`, `IP int` with a pointer-receiver `MarshalText`) is named `"3"`
  where encoding/json uses the text;
- `map[*BJ]` (`BJ` has a value `MarshalJSON` and `MarshalText`) is named by `MarshalJSON`; when
  it writes a value which is not a string, the output is not JSON (`{{"json":1}:1}`, with no error;
  encoding/json writes `{"text":1}`), as in v0.11.2;
- keys of 256 or more levels of pointers fail with "unsupported type" where encoding/json names
  them;
- the error of a failing key marshaler names the value type where encoding/json names the
  pointer type (`map[EV]`, `map[*time.Time]` with the year 10000).

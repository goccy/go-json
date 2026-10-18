# Encode benchmarks of the report still slower than sonic

- Area: encoder
- Kind: speed
- Priority: 3
- Status: open

## What

On the benchmark report (amd64, https://goccy.github.io/go-json/), after #681, #685 and #686
(2026-09-30):

- `fastest` config, anthropic payload: 0.1-6.5% slower than sonic;
- `fast` config, mixed text with 40% non-ASCII on Xeon 8573C: 2.3% slower.

## Causes found

- anthropic, `fastest`: the VM and struct keys took about 0.22 s more than sonic's generated
  code over the same encodes, the final copy 0.12 s more; strings are faster than sonic's.
- sonic skips the validation of marshaler output in `fastest`.

Re-measure on the current report before starting.

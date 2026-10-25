# Work queue

The known work on go-json, one file per item. Reports and requests of users stay on GitHub
Issues; this directory holds what the maintainers have found and planned.

## Order

The decoding API of v2 is complete (2026-10-23). Its speed against v1 comes first, then the calls
of the v2 methods by the root package; the known defects and the other speed gaps wait until
then, since none of them is urgent.

Priority 1: now

- [decoder-v2-speed.md](decoder-v2-speed.md): v2 decoding at parity with v1
- [v1-calls-v2-methods.md](v1-calls-v2-methods.md): the root package calls `UnmarshalJSONFrom`
  and `MarshalJSONTo` (right after the speed of v2 decoding)

Priority 2: first after priority 1

- [decoder-v2-stream-error-places.md](decoder-v2-stream-error-places.md)
- [decoder-disallow-unknown-fields-order.md](decoder-disallow-unknown-fields-order.md)
- [decoder-unexported-embedded-pointer-error.md](decoder-unexported-embedded-pointer-error.md)
- [decoder-case-insensitive-field-choice.md](decoder-case-insensitive-field-choice.md)
- [decoder-float32-rounding.md](decoder-float32-rounding.md)
- [decoder-valid-long-strings.md](decoder-valid-long-strings.md)
- [encoder-sorted-map-speed.md](encoder-sorted-map-speed.md)

Priority 3

- [decoder-map-key-errors.md](decoder-map-key-errors.md)
- [decoder-new-decoder-allocations.md](decoder-new-decoder-allocations.md)
- [decoder-stream-speed-amd64.md](decoder-stream-speed-amd64.md)
- [decoder-code-unmarshal-debt.md](decoder-code-unmarshal-debt.md)
- [decoder-tests-judged-by-time.md](decoder-tests-judged-by-time.md)
- [encoder-v2-indent-offsets.md](encoder-v2-indent-offsets.md)
- [encoder-map-key-types.md](encoder-map-key-types.md)
- [encoder-small-deviations.md](encoder-small-deviations.md)
- [encoder-quadratic-nesting.md](encoder-quadratic-nesting.md)
- [encoder-marshalers-per-call.md](encoder-marshalers-per-call.md)
- [encoder-report-gaps-vs-sonic.md](encoder-report-gaps-vs-sonic.md)
- [selection-api.md](selection-api.md)
- [selection-promoted-field-key.md](selection-promoted-field-key.md)
- [ci-go-1.27-pin.md](ci-go-1.27-pin.md)

Priority 4

- [decoder-redundant-skip-white-space.md](decoder-redundant-skip-white-space.md)
- [decoder-compile-embedded-chains.md](decoder-compile-embedded-chains.md)
- [decoder-v2-error-before-colon.md](decoder-v2-error-before-colon.md)
- [decoder-v2-small-deviations.md](decoder-v2-small-deviations.md)
- [decoder-error-source-display.md](decoder-error-source-display.md)
- [encoder-error-texts.md](encoder-error-texts.md)
- [encoder-vm-structure.md](encoder-vm-structure.md)
- [encoder-marshal-to-stack-depth.md](encoder-marshal-to-stack-depth.md)
- [encoder-type-table-guards.md](encoder-type-table-guards.md)
- [ci-benchcheck-binary-placement.md](ci-benchcheck-binary-placement.md)

## Conventions

- One item per file, named by what it covers (`<area>-<subject>.md`), with the fields
  Area, Kind, Priority and Status at the top.
- "Recorded" items were seen on master at the date given and have not been checked since;
  reproduce them on master before fixing.
- The pull request that resolves an item deletes its file and its line here.
- What go-json follows: the documented behavior of Go 1.27's encoding/json and
  encoding/json/v2 on every Go version; not error texts, undocumented states after failed or
  misused calls, or bugs of the standard library.

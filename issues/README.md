# Work queue

The known work on go-json, one file per item. Reports and requests of users stay on GitHub
Issues; this directory holds what the maintainers have found and planned.

## Order

The decoding half of the v2 API comes first: the API is completed before the known defects and
speed gaps are worked on, since none of them is urgent. Everything below priority 1 waits until
the v2 decoder is done.

Priority 1: now

- [decoder-v2-api.md](decoder-v2-api.md): the decoding API of `github.com/goccy/go-json/v2`
- [v1-calls-v2-methods.md](v1-calls-v2-methods.md): the root package calls `UnmarshalJSONFrom`
  and `MarshalJSONTo` (right after the v2 decoder)

Priority 2: first after the v2 decoder

- [decoder-disallow-unknown-fields-order.md](decoder-disallow-unknown-fields-order.md)
- [decoder-unexported-embedded-pointer-error.md](decoder-unexported-embedded-pointer-error.md)
- [decoder-case-insensitive-field-choice.md](decoder-case-insensitive-field-choice.md)
- [decoder-valid-long-strings.md](decoder-valid-long-strings.md)
- [encoder-sorted-map-speed.md](encoder-sorted-map-speed.md)

Priority 3

- [decoder-map-key-errors.md](decoder-map-key-errors.md)
- [decoder-new-decoder-allocations.md](decoder-new-decoder-allocations.md)
- [decoder-stream-speed-amd64.md](decoder-stream-speed-amd64.md)
- [decoder-code-unmarshal-debt.md](decoder-code-unmarshal-debt.md)
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
- [decoder-error-source-display.md](decoder-error-source-display.md)
- [encoder-error-texts.md](encoder-error-texts.md)
- [encoder-vm-structure.md](encoder-vm-structure.md)
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

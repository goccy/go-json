# Small deviations of JSON Path

- Area: path
- Kind: defect
- Priority: 3
- Status: recorded 2026-10-28 (left by the fix of #774); checked on the branch of the fix, as on master 483eceb

## What

- `Path.Get` of `$` sets nothing and returns no error: `Extract` of `$` gives the whole text.
- `Path.Get` does not select a field promoted from an embedded struct by its name, where `Marshal` writes it as a
  member: `$.B` of `struct{ Inner; D int }` ( `Inner` has `B` ) fails, where `Marshal` writes `{"B":1,"D":0}`, and
  `$..X` of `struct{ inner; Y int }` ( an unexported `inner` with `X` ) selects nothing, where `Extract` of what
  `Marshal` writes, `{"X":1,"Y":2}`, selects `1`. `$.Inner.B` selects the field of an exported embedded struct.
- `PathString` is not the path that was created: `$.a` gives `.a`, `$['a.b'].c` and `$."a.b".c` give `.a.b.c`.
- `Extract` gives a value with the white space before it: `$.a` of `{"a": 1}` gives `" 1"`.

## Fix

Each is small: `Get` of `$` casts the source; the names of a struct are taken from the decoder of its type ( as
in path-cast-by-go-field-names.md ), with the promoted fields; `String` keeps the text of the path; the value starts
after its white space.

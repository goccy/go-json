# Path.Get and Path.Unmarshal fill a struct by the names of its Go fields

- Area: path
- Kind: defect
- Priority: 2
- Status: recorded 2026-10-28 (left by the fix of #774); checked on the branch of the fix, as on master 483eceb

## What

`Path.Get` and `Path.Unmarshal` cast the value they select to the destination ( `AssignValue`,
`internal/decoder/assign.go` ). Into a struct, they take a member of an object by the name of a Go field
( `reflect.Type.FieldByName` ), not by the names of the JSON object of the struct:

- the name in the tag is not used: `$.a` of `{"a":{"name":"x"}}` into ``struct{ Name string `json:"name"` }``
  leaves `Name` empty, with no error, where `Unmarshal` of `{"name":"x"}` sets it;
- the names are not matched without regard to case, as `Unmarshal` matches them: of
  `{"a":{"name":"x","Name":"y"}}`, only `"Name"` sets `Name`;
- a field tagged `"-"` is set by its Go name, where `Unmarshal` skips it;
- a struct cast from another struct takes its fields by their Go names, not the promoted ones: `$.a` of
  `{"a": outer}`, an `outer` of `struct{ inner; Y int }` with `inner{X: 1}` and `Y: 2`, into `struct{ X, Y int }` gives
  `{X: 0, Y: 2}`, where `Marshal` of `outer` writes `{"X":1,"Y":2}`.

`Path.Unmarshal` decodes each value it selects into `any` and casts the slice of them, so a struct destination takes
the first one.

## Fix

Decode into a struct as `Unmarshal` decodes: for `Path.Unmarshal`, decode the selected text into the destination
( its element for a slice ) with the decoder of its type, instead of the cast; for `Path.Get`, cast the members by
the names of the decoder of the struct. Keep the casts of the other kinds.

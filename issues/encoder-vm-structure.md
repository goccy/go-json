# Encoder VM: one huge function

- Area: encoder
- Kind: speed, maintenance
- Priority: 4
- Status: open

## What

`vm.Run` (generated from `internal/cmd/generator/vm.go.tmpl` into four VMs) is one function of
about 150 KB with a stack frame of about 9 KB. Go allocates registers for the whole function,
so a change in one opcode reshuffles thousands of instructions and moves unrelated benchmarks by
several percent. Changes have been kept to the old shapes and to non-inlined helpers because
of it.

Costs accepted in #626 for this reason:

- `OpInterface` calls the non-inlined `RuntimeContext.InterfaceValueAddr` once per interface
  value: 3-5% on `Encode_Interface_GoJson` and `Encode_MapInterface_GoJson`;
- direct types at the top level that are not pointers are copied to a context slot.

## Measured ceiling

Straight-line Go producing the same JSON as the VM (2026-09-21, arm64): 16 int fields 101 vs
106 ns, 8 string fields 70 vs 75 ns. The dispatch, spills and slots cost 5-7%; most time is in
the leaf functions and the fixed cost per call.

## Ideas

- Compile opcodes apart (functions per opcode, smaller dispatch loops), or the hot loop in
  assembly so that helpers can be inlined.
- Slot addresses by `unsafe.Add` on an `unsafe.Pointer` base, keeping slot values `uintptr`
  (one instruction less per access on amd64).
- Not to retry: replacing the `appendInt = encoder.AppendInt` function variables in
  `vm/util.go` with wrappers made it slower (106 to 124 ns, arm64).

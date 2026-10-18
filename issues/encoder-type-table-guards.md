# Shared type table: follow-ups left for when they are needed

- Area: encoder
- Kind: maintenance
- Priority: 4
- Status: deferred until needed

## What

#674 looks up the first entry of the shared table inline in the VM on a miss of the recent
code sets (`SharedCodeSets().LoadFirst`, then `RememberCodeSet`). Left for later:

- a test that fails when a new Go version stops inlining `LoadFirst` or `RememberCodeSet` into
  the VM;
- a lookup of two entries written directly in the VM template;
- the recent code sets have a fixed capacity per context; programs with more hot types fall off
  it (a growable table per context removes that cliff at the cost of memory).

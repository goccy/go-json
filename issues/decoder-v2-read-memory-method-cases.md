# Memory tests of UnmarshalRead which skip the methods

- Area: decoder (tests)
- Kind: maintenance
- Priority: 3
- Status: recorded 2026-10-28, found while adding a case of `StringifyWithLegacySemantics`

## What

Three cases of `TestUnmarshalReadMemory` (`v2/unmarshal_read_test.go`) are meant to measure the
memory of a value which a method decodes as the stream reads it: "counted by a method", "counted by
a method through a pointer" and "counted by a method of a number through a pointer". Their input
names the member `"records"`, and the field is `Records`, which encoding/json/v2 matches
case-sensitively by default: the member is of no field and is skipped, so the cases measure the
skipping of an unknown member, which "skipped" measures already, and the method is never called.

With the names matched (`"Records"`, or a tag `json:"records"`), the three cases pass: the memory of
the methods is bounded, only the tests don't show it.

## Fix

Name the member as the field, and assert that the method counted the elements, so that a case which
skips the value fails.

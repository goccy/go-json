# A power of ten of mantissa 1 is left to strconv

- Area: decoder
- Kind: speed
- Priority: 4
- Status: recorded 2026-10-28 (found by the review of the fix of #770); checked on master 5c8a7dd

## What

`mulPow10` (`internal/jsonnum/pow10.go`) never decides a number whose mantissa is 1, out of the range of Clinger's
fast path: `1e23`, `1e100`, `1e-23`, `1e-30`, and `0.` followed by 22 zeros and `1`. They are parsed by strconv,
correctly but slower. `2e100`, `10e99` and `1.0e23` are decided.

The product of a mantissa `w` and the 128-bit significand of `10^q`, whose top bit is set, has a top word of zero
only for `w == 1`, and `round192` declines a number whose top word is zero.

## Fix

Take the 128-bit significand itself for `w == 1` ( round it from its two words ), or double `w` and take one from
the binary exponent. Measure the float benchmarks, whose common path must not get slower.

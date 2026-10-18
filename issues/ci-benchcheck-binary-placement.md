# Benchmark gate: binaries of identical code fall in different speed clusters

- Area: ci
- Kind: maintenance
- Priority: 4
- Status: idea, to be measured before any proposal

## What

On AMD EPYC 7763 and 9V74 runners, small encode benchmarks have two clusters of results 20-50%
apart over the test binaries, while the rounds of one binary differ by about 1%. The content of
a binary does not decide its cluster: copies of one binary file fell in either cluster, and after
the page cache was dropped many binaries moved to the other one. The mechanism is unknown
(the runners have no PMU).

#744 made the gate judge a single benchmark as slower only when both its fastest layout and the
median of its layouts are beyond the tolerance, which removed the false failures measured.

## Idea

Write the base and head binaries of each layout one right after the other, so that both are
placed alike, and measure whether they then fall in the same cluster.

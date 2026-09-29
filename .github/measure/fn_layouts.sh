#!/bin/bash
# Measures the benchmarks of AppendString and AppendInt of master and the short strings and integers branch, each
# with the tests of the branch, in four function layouts and with a pad: the mean over the layouts of the median
# of the rounds, against master.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
git worktree add -q /tmp/wt-master origin/master
git worktree add -q /tmp/wt-head origin/perf/encoder-short-strings-ints
cp /tmp/wt-head/internal/encoder/string_fast_path_test.go /tmp/wt-head/internal/encoder/int_test.go /tmp/wt-master/internal/encoder/
for v in master head; do
  git -C /tmp/wt-$v worktree add -q /tmp/wt-${v}pad HEAD 2>/dev/null || true
done
for v in master head; do
  cp -r /tmp/wt-$v /tmp/wt-${v}pad-src
  cp $root/.github/measure/pad.go /tmp/wt-${v}pad-src/internal/encoder/zz_measure_pad.go
done
names="master masterpad head headpad"
for l in 1 2 3 4; do
  for n in master head; do
    (cd /tmp/wt-$n && go test -c -ldflags=-randlayout=$l -o /tmp/fn-$n-$l.test ./internal/encoder/)
    (cd /tmp/wt-${n}pad-src && go test -c -ldflags=-randlayout=$l -o /tmp/fn-${n}pad-$l.test ./internal/encoder/)
  done
done
for r in 1 2 3; do
  for l in 1 2 3 4; do
    for n in $names; do
      /tmp/fn-$n-$l.test -test.run '^$' -test.bench 'BenchmarkAppendString/|BenchmarkAppendInt/' -test.benchtime 60ms | awk -v n=$n -v l=$l '/ns\/op/ {print n, l, $1, $3}' >> /tmp/fl.txt
    done
  done
done
NAMES="$names" python3 - <<'PY'
import collections, os, statistics
names = os.environ['NAMES'].split()
d = collections.defaultdict(list); order = []
for line in open('/tmp/fl.txt'):
    n, l, b, v = line.split(); d[(n, l, b)].append(float(v))
    if b not in order: order.append(b)
mean = {}
for (n, l, b), vs in d.items():
    mean.setdefault((n, b), []).append(statistics.median(vs))
print(f"{'benchmark':48}" + ''.join(f"{n:>11}" for n in names))
for b in order:
    m = statistics.mean(mean[('master', b)])
    print(f"{b:48}" + ''.join(f"{(statistics.mean(mean[(n, b)]) / m - 1) * 100:+10.1f}%" for n in names))
PY

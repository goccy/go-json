#!/bin/bash
# Measures the benchmark of AppendString of master and of variants of the strings ( str_v3.patch, str_v5.patch, str_v6.patch ), each also
# with a pad, in four function layouts: the mean over the layouts of the median of the rounds, against master.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
names=""
for v in master v3 v5 v6; do
  for pad in "" pad; do
    n=$v$pad; names="$names $n"
    git worktree add -q /tmp/wt-$n origin/master
    cp $root/.github/measure/string_fast_path_test.go /tmp/wt-$n/internal/encoder/string_fast_path_test.go
    [ $v = master ] || (cd /tmp/wt-$n && git apply $root/.github/measure/str_$v.patch)
    [ -z "$pad" ] || cp $root/.github/measure/pad.go /tmp/wt-$n/internal/encoder/zz_measure_pad.go
    for l in 1 2 3 4; do
      (cd /tmp/wt-$n && go test -c -ldflags=-randlayout=$l -o /tmp/fn-$n-$l.test ./internal/encoder/)
    done
  done
done
for r in 1 2 3; do
  for l in 1 2 3 4; do
    for n in $names; do
      /tmp/fn-$n-$l.test -test.run '^$' -test.bench 'BenchmarkAppendString/' -test.benchtime 40ms | awk -v n=$n -v l=$l '/ns\/op/ {print n, l, $1, $3}' >> /tmp/sv.txt
    done
  done
done
NAMES="$names" python3 - <<'PY'
import collections, os, statistics
names = os.environ['NAMES'].split()
d = collections.defaultdict(list); order = []
for line in open('/tmp/sv.txt'):
    n, l, b, v = line.split(); d[(n, l, b)].append(float(v))
    if b not in order: order.append(b)
mean = {}
for (n, l, b), vs in d.items():
    mean.setdefault((n, b), []).append(statistics.median(vs))
print(f"{'benchmark':44}" + ''.join(f"{n:>10}" for n in names))
for b in order:
    m = statistics.mean(mean[('master', b)])
    print(f"{b:44}" + ''.join(f"{(statistics.mean(mean[(n, b)]) / m - 1) * 100:+9.1f}%" for n in names))
PY

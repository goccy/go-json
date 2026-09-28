#!/bin/bash
# Builds master and the string branch, each also with a pad, with several function layouts, and runs the encode
# benchmarks of the payloads without marshalers by turns: the mean over the layouts of the fastest round of each,
# against master.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
BENCH='^Benchmark_(Encode_(Small|Medium|Large)Struct_GoJson|TwitterBinding_GoJson|TwitterBinding_GoJsonLikeSonicFast)$'
LAYOUTS=4; ROUNDS=3
root=$PWD
variant() { # name ref [pad]
  git worktree add -q /tmp/wt-$1 $2
  if [ -n "$3" ]; then cp $root/.github/measure/pad.go /tmp/wt-$1/internal/encoder/zz_measure_pad.go; fi
  for l in $(seq 1 $LAYOUTS); do
    (cd /tmp/wt-$1/benchmarks && go test -c -ldflags=-randlayout=$l -o /tmp/bin-$1-$l.test .)
  done
}
variant master origin/master
variant masterpad origin/master pad
variant head origin/perf/encoder-string-escape
variant headpad origin/perf/encoder-string-escape pad
names="master masterpad head headpad"
cd benchmarks
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for n in $names; do
      /tmp/bin-$n-$l.test -test.run '^$' -test.bench "$BENCH" -test.benchtime 200ms | awk -v n=$n -v l=$l '/ns\/op/ {print n, l, $1, $3}' >> /tmp/ab.txt
    done
  done
done
NAMES="$names" python3 - <<'PY'
import collections, os, statistics
names = os.environ['NAMES'].split()
best = {}
for line in open('/tmp/ab.txt'):
    n, l, b, v = line.split()
    best[(n, l, b)] = min(best.get((n, l, b), 1e18), float(v))
mean = collections.defaultdict(list)
for (n, l, b), v in best.items():
    mean[(n, b)].append(v)
print(f"{'benchmark':52}" + ''.join(f"{n:>11}" for n in names))
for b in sorted({b for (_, b) in mean}):
    m = statistics.mean(mean[('master', b)])
    print(f"{b:52}" + ''.join(f"{(statistics.mean(mean[(n, b)]) / m - 1) * 100:+10.1f}%" for n in names))
PY

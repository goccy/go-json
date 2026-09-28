#!/bin/bash
# Builds the head and the candidate ( a patch ), each with pads of 0, 2, 4 and 6 dictionaries, which move the
# read-only data by 32 bytes each two, with two function layouts, and prints the decode of the types of each build
# against the head, to compare the spread of the layouts of the two.
set -e
root=$PWD; LAYOUTS=2; ROUNDS=3
lscpu | sed -n 's/^Model name: *//p' | head -1
pad() {
  echo 'package decoder'
  echo
  echo '//go:noinline'
  echo 'func measurePadOf[T any](n int) any { return make([]T, n) }'
  for i in $(seq 1 $1); do
    echo "type measurePad$i struct{ n [$i]int }"
    echo "var measurePadSink$i = measurePadOf[measurePad$i](1)"
  done
}
names=""
for v in head hp; do
  for k in 0 2 4 6; do
    n=$v$k; names="$names $n"
    git worktree add -q /tmp/wt-$n HEAD
    cp $root/.github/measure/type_lookup_test.go /tmp/wt-$n/
    case $v in hp) (cd /tmp/wt-$n && git apply $root/.github/measure/hp.patch);; esac
    [ $k = 0 ] || pad $k > /tmp/wt-$n/internal/decoder/zz_measure_pad.go
    for l in $(seq 1 $LAYOUTS); do
      (cd /tmp/wt-$n && go test -c -ldflags=-randlayout=$l -o /tmp/root-$n-$l.test .)
    done
  done
done
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for n in $names; do
      /tmp/root-$n-$l.test -test.run '^$' -test.bench 'BenchmarkTypeLookups$' -test.benchtime 200ms | awk -v n=$n -v l=$l '/ns\/op/ && /decode\// {print n, l, $1, $3}' >> /tmp/s.txt
    done
  done
done
NAMES="$names" python3 - <<'PY'
import collections, os, statistics
names = os.environ['NAMES'].split()
best = {}
order = []
for line in open('/tmp/s.txt'):
    n, l, b, ns = line.split()
    best[(n, l, b)] = min(best.get((n, l, b), 1e18), float(ns))
    if b not in order:
        order.append(b)
mean = collections.defaultdict(list)
for (n, l, b), ns in best.items():
    mean[(n, b)].append(ns)
print(f"{'benchmark':58}" + ''.join(f"{n:>7}" for n in names))
for b in order:
    m = statistics.mean(mean[('head0', b)])
    print(f"{b:58}" + ''.join(f"{(statistics.mean(mean[(n, b)]) / m - 1) * 100:+6.1f}%" for n in names))
PY

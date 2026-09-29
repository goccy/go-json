#!/bin/bash
# Measures the encode of Japanese strings by go-json and sonic in each report configuration, in three function
# layouts: the mean over the layouts of the best of the rounds, and go-json against sonic.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
git worktree add -q /tmp/wt origin/master
cp $root/.github/measure/zz_multibyte_test.go $root/.github/measure/zz_prof_test.go /tmp/wt/benchmarks/
cd /tmp/wt/benchmarks
for l in 1 2 3; do go test -c -ldflags=-randlayout=$l -o /tmp/mb-$l.test .; done
for r in 1 2 3; do
  for l in 1 2 3; do
    /tmp/mb-$l.test -test.run '^$' -test.bench 'BenchmarkZZMultibyte/' -test.benchtime 100ms | awk -v l=$l '/ns\/op/ {print l, $1, $3}' >> /tmp/mb.txt
  done
done
python3 - <<'PY'
import collections, statistics, re
best = {}
for line in open('/tmp/mb.txt'):
    l, b, v = line.split()
    b = re.sub(r'-\d+$', '', b.replace('BenchmarkZZMultibyte/', ''))
    best[(l, b)] = min(best.get((l, b), 1e18), float(v))
mean = collections.defaultdict(list)
for (l, b), v in best.items():
    mean[b].append(v)
keys = sorted({b.rsplit('/', 1)[0] for b in mean}, key=lambda k: (k.split('/')[0], k))
print(f"{'config/case':22} {'go-json ns':>11} {'sonic ns':>10} {'go-json/sonic':>14}")
for k in keys:
    g = statistics.mean(mean[k + '/go-json']); s = statistics.mean(mean[k + '/sonic'])
    print(f"{k:22} {g:11.1f} {s:10.1f} {(g / s - 1) * 100:+13.1f}%")
PY

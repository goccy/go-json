#!/bin/bash
# Counts the lookups of the code sets for an encode of each report payload by go-json on master, and how many miss
# the recent code sets of the context, in the default function layout and in three random ones.
set -e
root=$PWD
git worktree add -q /tmp/wt-count origin/master
cp $root/.github/measure/zz_count_test.go /tmp/wt-count/benchmarks/
(cd /tmp/wt-count && git apply $root/.github/measure/count.patch)
cd /tmp/wt-count/benchmarks
for l in 0 1 2 3; do
  flags=""
  [ $l = 0 ] || flags="-ldflags=-randlayout=$l"
  go test -c $flags -o /tmp/c-$l.test .
  /tmp/c-$l.test -test.run '^TestZZCodeSetCounts$' | awk -v l=$l '/^count/ {print "L" l, $2, $3, $5, $7, $9}' >> /tmp/count.txt
done
python3 - <<'PY'
import collections
rows = collections.defaultdict(dict)
for line in open('/tmp/count.txt'):
    l, cfg, p, lookups, misses, shared = line.split()
    rows[(cfg, p)][l] = (float(lookups), float(misses), float(shared))
print(f"{'config':16} {'payload':12}  lookups  misses of the recent sets by layout ( L0 is the default one )")
for (cfg, p), d in sorted(rows.items()):
    print(f"{cfg:16} {p:12} {d['L0'][0]:8.1f}  " + '  '.join(f"{l} {d[l][1]:6.1f}" for l in ['L0', 'L1', 'L2', 'L3']))
PY

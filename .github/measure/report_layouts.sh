#!/bin/bash
# Runs the report encode of the payloads without marshalers for master and the string branch, each also with a
# pad, with several function layouts: the median over the rounds of each layout, then the mean over the layouts,
# against master.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
LAYOUTS=3; ROUNDS=3
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
mkdir -p /tmp/rep
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for n in $names; do
      (cd /tmp/wt-$n/benchmarks && BENCH_REPORT_ONLY='^(std|fast|fastest)/go-json/encode/(small|medium|large|twitter|code)$' BENCH_REPORT_ROUNDS=1 BENCH_REPORT_OUT=/tmp/rep/$n-$l-$r.json /tmp/bin-$n-$l.test -test.run '^TestReport$' -test.count=1 -test.timeout 60m -test.benchtime=100ms > /dev/null)
    done
  done
done
NAMES="$names" python3 - <<'PY'
import collections, glob, json, os, statistics
names = os.environ['NAMES'].split()
res = collections.defaultdict(list)
for path in glob.glob('/tmp/rep/*.json'):
    n, l, r = os.path.basename(path)[:-5].split('-')
    for x in json.load(open(path))['results']:
        if x['condition'] == 'live-heap':
            res[(n, l, x['config'], x['payload'])].append(x['nsPerOp'])
mean = collections.defaultdict(list)
for (n, l, c, p), v in res.items():
    mean[(n, c, p)].append(statistics.median(v))
keys = sorted({(c, p) for (_, c, p) in mean})
print(f"{'config/payload':30}" + ''.join(f"{n:>11}" for n in names))
for c, p in keys:
    m = statistics.mean(mean[('master', c, p)])
    print(f"{c + ' ' + p:30}" + ''.join(f"{(statistics.mean(mean[(n, c, p)]) / m - 1) * 100:+10.1f}%" for n in names))
PY

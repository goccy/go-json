#!/bin/bash
# Measures the encode of the report payloads by master, the integer branch, and the integer branch encoding into
# a new buffer of the size of the last output instead of copying the output ( nocopy.patch ), with several
# function layouts, and sonic for reference; then profiles who calls memmove for many encodes of twitter.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD; LAYOUTS=3; ROUNDS=3
variant() { # name ref [patch]
  git worktree add -q /tmp/wt-$1 $2
  cp $root/.github/measure/zz_prof_test.go /tmp/wt-$1/benchmarks/
  if [ -n "$3" ]; then (cd /tmp/wt-$1 && git apply $root/.github/measure/$3); fi
  for l in $(seq 1 $LAYOUTS); do
    (cd /tmp/wt-$1/benchmarks && go test -c -ldflags=-randlayout=$l -o /tmp/bin-$1-$l.test .)
  done
}
variant master origin/master
variant head origin/perf/encoder-int-output
variant nocopy origin/perf/encoder-int-output nocopy.patch
names="master head nocopy"
cd /tmp/wt-master/benchmarks
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for cfg in fastest fast std; do
      for p in small medium twitter twitter-any github openai anthropic code; do
        for n in $names; do
          ZZ_CONFIG=$cfg/go-json ZZ_PAYLOAD=$p /tmp/bin-$n-$l.test -test.run '^$' -test.bench '^BenchmarkZZEncode$' -test.benchtime 200ms | awk -v n=$n -v l=$l -v k=$cfg/$p '/ns\/op/ {print n, l, k, $3}' >> /tmp/c.txt
        done
        if [ $l = 1 ]; then
          ZZ_CONFIG=$cfg/sonic ZZ_PAYLOAD=$p /tmp/bin-master-1.test -test.run '^$' -test.bench '^BenchmarkZZEncode$' -test.benchtime 200ms | awk -v k=$cfg/$p '/ns\/op/ {print "sonic", 1, k, $3}' >> /tmp/c.txt
        fi
      done
    done
  done
done
python3 - <<'PY'
import collections, statistics
best = {}
for line in open('/tmp/c.txt'):
    n, l, k, v = line.split()
    best[(n, l, k)] = min(best.get((n, l, k), 1e18), float(v))
mean = collections.defaultdict(list)
for (n, l, k), v in best.items():
    mean[(n, k)].append(v)
keys = sorted({k for (_, k) in mean})
print(f"{'config/payload':22} {'master':>10} {'head':>8} {'nocopy':>8} {'sonic':>10} {'head/sonic':>11} {'nocopy/sonic':>13}")
for k in keys:
    m = statistics.mean(mean[('master', k)]); h = statistics.mean(mean[('head', k)]); c = statistics.mean(mean[('nocopy', k)]); s = statistics.mean(mean[('sonic', k)])
    print(f"{k:22} {m:10.0f} {(h / m - 1) * 100:+7.1f}% {(c / m - 1) * 100:+7.1f}% {s:10.0f} {(h / s - 1) * 100:+10.1f}% {(c / s - 1) * 100:+12.1f}%")
PY
IGNORE='pretouchSonic|reportPayloads|newReportPayload|PretouchMany'
for v in "head go-json" "master sonic"; do
  set -- $v
  echo "== memmove of $2 fastest twitter 3000000x"
  ZZ_CONFIG=fastest/$2 ZZ_PAYLOAD=twitter /tmp/bin-$1-1.test -test.run '^$' -test.bench '^BenchmarkZZEncode$' -test.benchtime=3000000x -test.cpuprofile /tmp/m.prof | grep ns/op
  go tool pprof -ignore "$IGNORE" -peek 'runtime\.memmove$|runtime\.mallocgc$' /tmp/bin-$1-1.test /tmp/m.prof 2>/dev/null | sed -n '/flat  flat%/,$p' | head -30
done

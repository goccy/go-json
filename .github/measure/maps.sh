#!/bin/bash
# Measures the encode of the report payloads with maps by master and the map branch, with several function
# layouts, and sonic for reference; then profiles the map branch and sonic for the same number of encodes.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD; LAYOUTS=3; ROUNDS=3; BRANCH=${BRANCH:-perf/encoder-map-sort}
variant() { # name ref
  git worktree add -q /tmp/wt-$1 $2
  cp $root/.github/measure/zz_prof_test.go /tmp/wt-$1/benchmarks/
  for l in $(seq 1 $LAYOUTS); do
    (cd /tmp/wt-$1/benchmarks && go test -c -ldflags=-randlayout=$l -o /tmp/bin-$1-$l.test .)
  done
}
variant master origin/master
variant head origin/$BRANCH
names="master head"
cd /tmp/wt-master/benchmarks
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for cfg in std fast fastest; do
      for p in twitter-any openai github anthropic twitter small; do
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
print(f"{'config/payload':22} {'master':>10} {'head':>8} {'sonic':>10} {'master/sonic':>13} {'head/sonic':>11}")
for k in sorted({k for (_, k) in mean}):
    m = statistics.mean(mean[('master', k)]); h = statistics.mean(mean[('head', k)]); s = statistics.mean(mean[('sonic', k)])
    print(f"{k:22} {m:10.0f} {(h / m - 1) * 100:+7.1f}% {s:10.0f} {(m / s - 1) * 100:+12.1f}% {(h / s - 1) * 100:+10.1f}%")
PY
if [ "$MACHINE" = 1 ]; then
  IGNORE='pretouchSonic|reportPayloads|newReportPayload|PretouchMany'
  while read cfg p n; do
    for v in "head go-json" "master sonic"; do
      set -- $v
      echo "== profile $cfg/$2/$p ${n}x"
      ZZ_CONFIG=$cfg/$2 ZZ_PAYLOAD=$p /tmp/bin-$1-1.test -test.run '^$' -test.bench '^BenchmarkZZEncode$' -test.benchtime=${n}x -test.cpuprofile /tmp/p.prof | grep ns/op
      go tool pprof -ignore "$IGNORE" -top -cum -nodecount 45 /tmp/bin-$1-1.test /tmp/p.prof 2>/dev/null | sed -n '6,51p'
    done
  done <<'LIST'
std twitter-any 150000
fast openai 800000
LIST
fi

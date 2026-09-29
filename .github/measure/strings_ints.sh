#!/bin/bash
# Measures the encode of the report payloads by master and a branch ( BRANCH ), with several function layouts,
# and sonic for reference, and the benchmarks of AppendString and AppendInt of both; then profiles the branch and
# sonic for the same number of encodes.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD; LAYOUTS=3; ROUNDS=3; BRANCH=${BRANCH:-perf/encoder-short-strings-ints}
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
      for p in small medium large twitter twitter-any github openai anthropic code; do
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
# the benchmarks of the functions, with the tests of the branch in both
cp /tmp/wt-head/internal/encoder/string_fast_path_test.go /tmp/wt-head/internal/encoder/int_test.go /tmp/wt-master/internal/encoder/
for n in master head; do (cd /tmp/wt-$n && go test -c -o /tmp/fn-$n.test ./internal/encoder/); done
for r in 1 2 3 4 5; do
  for n in master head; do
    /tmp/fn-$n.test -test.run '^$' -test.bench 'BenchmarkAppendString/|BenchmarkAppendInt/' -test.benchtime 100ms | awk -v n=$n '/ns\/op/ {print n, $1, $3}' >> /tmp/fn.txt
  done
done
python3 - <<'PY2'
import collections, statistics
d = collections.defaultdict(list); order = []
for line in open('/tmp/fn.txt'):
    n, b, v = line.split(); d[(n, b)].append(float(v))
    if b not in order: order.append(b)
for b in order:
    m, h = statistics.median(d[('master', b)]), statistics.median(d[('head', b)])
    print(f"fn {b:48} {m:8.2f} -> {h:8.2f} ns {(h / m - 1) * 100:+6.1f}%")
PY2
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
fastest small 8000000
fastest twitter 600000
LIST
fi

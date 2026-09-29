#!/bin/bash
# Measures the encode of the report payloads by master and variants ( <v>.patch ), each also
# with a pad, in three function layouts, and sonic for reference: the mean over the layouts of the best of the rounds.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD; LAYOUTS=${LAYOUTS:-3}; ROUNDS=${ROUNDS:-3}; PAYLOADS=${PAYLOADS:-small medium large twitter twitter-any github openai anthropic code}
names=""
for v in master ${VARIANTS:-keys_v1}; do
  ref=origin/master
  for pad in "" pad; do
    n=$v$pad; names="$names $n"
    git worktree add -q /tmp/wt-$n $ref
    [ $v = master ] || (cd /tmp/wt-$n && git apply $root/.github/measure/$v.patch && go test -count=1 . ./internal/encoder/... >/dev/null)
    cp $root/.github/measure/zz_prof_test.go /tmp/wt-$n/benchmarks/
    [ -z "$pad" ] || cp $root/.github/measure/pad.go /tmp/wt-$n/internal/encoder/zz_measure_pad.go
    for l in $(seq 1 $LAYOUTS); do
      (cd /tmp/wt-$n/benchmarks && go test -c -ldflags=-randlayout=$l -o /tmp/bin-$n-$l.test .)
    done
  done
done
cd /tmp/wt-master/benchmarks
for p in $PAYLOADS; do echo "$p $(ZZ_PAYLOAD=$p /tmp/bin-master-1.test -test.run '^TestZZKeys$' -test.v | grep -o 'keys:.*')"; done
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for cfg in std fast fastest; do
      for p in $PAYLOADS; do
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
NAMES="$names" python3 - <<'PY'
import collections, os, statistics
names = os.environ['NAMES'].split()
best = {}
for line in open('/tmp/c.txt'):
    n, l, k, v = line.split()
    best[(n, l, k)] = min(best.get((n, l, k), 1e18), float(v))
mean = collections.defaultdict(list)
for (n, l, k), v in best.items():
    mean[(n, k)].append(v)
print(f"{'config/payload':22}" + ''.join(f"{n:>10}" for n in names) + f"{'master/sonic':>13}{'variant/sonic':>14}")
for k in sorted({k for (_, k) in mean}):
    m = statistics.mean(mean[('master', k)]); s = statistics.mean(mean[('sonic', k)]); h = statistics.mean(mean[(names[-2], k)])
    print(f"{k:22}" + ''.join(f"{(statistics.mean(mean[(n, k)]) / m - 1) * 100:+9.1f}%" for n in names) + f"{(m / s - 1) * 100:+12.1f}%{(h / s - 1) * 100:+10.1f}%")
PY

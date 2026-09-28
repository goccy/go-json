#!/bin/bash
# Profiles the encode of the GitHub REST issues on amd64 before and after the trust of the marshalers which a
# type has by embedding time.Time, and times the two by turns.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
for v in base:af286b6 new:1f108ff; do
  n=${v%%:*}; rev=${v#*:}
  git worktree add -q /tmp/wt-$n $rev
  (cd /tmp/wt-$n/benchmarks && go test -c -o /tmp/bench-$n.test .)
done
BENCH='^Benchmark_Encode_GitHubREST_(GoJson|GoJsonValidateString|GoJsonLikeSonic|SonicStd|SonicValidateString|SonicFastest)$'
for r in 1 2 3 4 5; do
  for n in base new; do
    (cd /tmp/wt-$n/benchmarks && /tmp/bench-$n.test -test.run '^$' -test.bench "$BENCH" -test.benchtime 300ms -test.benchmem) | awk -v n=$n '/ns\/op/ {print n, $1, $3, $(NF-1)}' >> /tmp/t.txt
  done
done
python3 - <<'PY'
best = {}; allocs = {}; order = []
for line in open('/tmp/t.txt'):
    n, b, ns, a = line.split()
    best[(n, b)] = min(best.get((n, b), 1e18), float(ns)); allocs[(n, b)] = a
    if b not in order:
        order.append(b)
for b in order:
    o, w = best[('base', b)], best[('new', b)]
    print(f"{b:56} {o:10.0f} -> {w:10.0f} ns/op {(w / o - 1) * 100:+6.1f}%  allocs {allocs[('base', b)]} -> {allocs[('new', b)]}")
PY
for n in base new; do
  echo "== profile $n"
  (cd /tmp/wt-$n/benchmarks && /tmp/bench-$n.test -test.run '^$' -test.bench '^Benchmark_Encode_GitHubREST_GoJson$' -test.benchtime 3s -test.cpuprofile /tmp/$n.prof >/dev/null)
  go tool pprof -top -nodecount 30 /tmp/bench-$n.test /tmp/$n.prof 2>/dev/null | sed -n '4,40p'
done

#!/bin/bash
# Builds the head and the candidate ( a patch ), and runs the decode of the types alone and after the other
# benchmarks, printing where the objects which it reads are, to see whether the history of the process moves it.
set -e
root=$PWD; LAYOUTS=2; ROUNDS=3
for v in head hp; do
  git worktree add -q /tmp/wt-$v HEAD
  cp $root/.github/measure/type_lookup_test.go $root/.github/measure/zz_probe_test.go /tmp/wt-$v/
  case $v in hp) (cd /tmp/wt-$v && git apply $root/.github/measure/hp.patch);; esac
  for l in $(seq 1 $LAYOUTS); do
    (cd /tmp/wt-$v && go test -c -ldflags=-randlayout=$l -o /tmp/root-$v-$l.test .)
  done
done
modes=(full 'BenchmarkTypeLookups$' decodes 'BenchmarkTypeLookups$/decode' pair 'BenchmarkTypeLookups$/(encode|decode)/top-level/1_types$' alone 'BenchmarkTypeLookups$/decode/top-level/1_types$')
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for ((m = 0; m < ${#modes[@]}; m += 2)); do
      for v in head hp; do
        ZZPROBE=1 /tmp/root-$v-$l.test -test.run '^$' -test.bench "${modes[m+1]}" -test.benchtime 300ms | awk -v v=$v -v l=$l -v m=${modes[m]} -v r=$r '/ns\/op/ && /decode\/top-level\/(1|8|128)_types-/ {print m, v, l, $1, $3} /^probe/ && r == 1 {print "probe", m, v, l, $2, $3, $4, $5, $6}' >> /tmp/o.txt
      done
    done
  done
done
grep '^probe' /tmp/o.txt | sort -u
grep -v '^probe' /tmp/o.txt | python3 -c '
import sys, collections, statistics
best = {}
for line in sys.stdin:
    m, v, l, b, ns = line.split()
    best[(m, v, l, b)] = min(best.get((m, v, l, b), 1e18), float(ns))
mean = collections.defaultdict(list)
for (m, v, l, b), ns in best.items():
    mean[(m, b, v)].append(ns)
for (m, b) in sorted({(m, b) for (m, b, v) in mean}):
    h, p = statistics.mean(mean[(m, b, "head")]), statistics.mean(mean[(m, b, "hp")])
    print(f"{m:8} {b:52} head={h:7.2f} hp={p:7.2f} {(p / h - 1) * 100:+5.1f}%")
'

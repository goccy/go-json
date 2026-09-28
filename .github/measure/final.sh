#!/bin/bash
# Builds the head and the strategies of the lookups of the types ( a patch each ), each also with a pad, with
# several function layouts, and runs them by turns; prints the mean over the layouts of the fastest round of each
# benchmark, against the head.
set -e
root=$PWD; LAYOUTS=3; ROUNDS=3
names=""
for v in head headpad hp hppad; do
  names="$names $v"
  git worktree add -q /tmp/wt-$v HEAD
  cp $root/.github/measure/type_lookup_test.go /tmp/wt-$v/
  case $v in hp|hppad) (cd /tmp/wt-$v && git apply $root/.github/measure/hp.patch);; esac
  case $v in *pad) cp $root/.github/measure/pad.go /tmp/wt-$v/internal/encoder/zz_measure_pad.go;; esac
  for l in $(seq 1 $LAYOUTS); do
    (cd /tmp/wt-$v && go test -c -ldflags=-randlayout=$l -o /tmp/root-$v-$l.test .)
    (cd /tmp/wt-$v/benchmarks && go test -c -ldflags=-randlayout=$l -o /tmp/bench-$v-$l.test .)
  done
done
SUITE='^Benchmark_(Encode_(Small|Medium|Large)StructCached_GoJson|Encode_Interface_GoJson|Encode_MapInterface_GoJson(LikeSonic)?|TwitterGeneric_GoJson|TwitterBinding_GoJson|Decode_(Small|Large)Struct_Unmarshal_GoJson|Decode_Twitter\w*_GoJson|Encode_OpenAIResponse_GoJson)$'
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for v in $names; do
      /tmp/root-$v-$l.test -test.run '^$' -test.bench 'BenchmarkTypeLookups$' -test.benchtime 150ms | awk -v v=$v -v l=$l '/ns\/op/ {print v, l, $1, $3}' >> /tmp/f.txt
      /tmp/root-$v-$l.test -test.run '^$' -test.bench 'BenchmarkTypeLookupsLargeProgram$' -test.benchtime 150ms | awk -v v=$v -v l=$l '/ns\/op/ {print v, l, $1, $3}' >> /tmp/f.txt
      (cd benchmarks && /tmp/bench-$v-$l.test -test.run '^$' -test.bench "$SUITE" -test.benchtime 200ms) | awk -v v=$v -v l=$l '/ns\/op/ {print v, l, $1, $3}' >> /tmp/f.txt
    done
  done
done
NAMES="$names" python3 - <<'PY'
import collections, os, statistics
names = os.environ['NAMES'].split()
best = {}
order = []
for line in open('/tmp/f.txt'):
    v, l, b, ns = line.split()
    best[(v, l, b)] = min(best.get((v, l, b), 1e18), float(ns))
    if b not in order:
        order.append(b)
mean = collections.defaultdict(list)
for (v, l, b), ns in best.items():
    mean[(v, b)].append(ns)
print(f"{'benchmark':66}" + ''.join(f"{n:>9}" for n in names))
for b in order:
    m = statistics.mean(mean[('head', b)])
    print(f"{b:66}" + ''.join(f"{(statistics.mean(mean[(n, b)]) / m - 1) * 100:+8.1f}%" for n in names))
PY

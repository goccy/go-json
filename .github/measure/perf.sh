#!/bin/bash
# Builds the head and the candidate ( a patch ), and compares the hardware counters and the profiles of the
# decode of one type, which the candidate slows although it runs the same instructions.
set -e
root=$PWD
sudo apt-get install -y -qq linux-tools-common "linux-tools-$(uname -r)" >/dev/null 2>&1 || sudo apt-get install -y -qq linux-tools-generic >/dev/null 2>&1 || true
PERF=$(command -v perf || ls /usr/lib/linux-tools/*/perf | head -1)
sudo sysctl -q kernel.perf_event_paranoid=-1 kernel.kptr_restrict=0
for v in head hp; do
  git worktree add -q /tmp/wt-$v HEAD
  cp $root/.github/measure/type_lookup_test.go /tmp/wt-$v/
  case $v in hp) (cd /tmp/wt-$v && git apply $root/.github/measure/hp.patch);; esac
  (cd /tmp/wt-$v && go test -c -o /tmp/root-$v.test .)
done
BENCH='BenchmarkTypeLookups$/decode/top-level/1_types$'
EVENTS=cycles,instructions,branches,branch-misses,L1-dcache-loads,L1-dcache-load-misses,L1-icache-load-misses,iTLB-load-misses,dTLB-load-misses,stalled-cycles-frontend,stalled-cycles-backend
for r in 1 2 3; do
  for v in head hp; do
    echo "== stat $v round $r"
    $PERF stat -x, -e $EVENTS /tmp/root-$v.test -test.run '^$' -test.bench "$BENCH" -test.benchtime 3s 2>&1 | grep -E 'ns/op|,' | grep -v '^#'
  done
done
for v in head hp; do
  echo "== record $v"
  $PERF record -q -o /tmp/perf-$v.data /tmp/root-$v.test -test.run '^$' -test.bench "$BENCH" -test.benchtime 3s >/dev/null 2>&1
  $PERF report -i /tmp/perf-$v.data --stdio --no-children --percent-limit 1.5 2>/dev/null | grep -v '^#' | grep -v '^$' | head -30
done
for f in 'github.com/goccy/go-json/internal/decoder.(*structDecoder).Decode' 'github.com/goccy/go-json/internal/decoder.(*intDecoder).Decode' 'github.com/goccy/go-json.unmarshal'; do
  for v in head hp; do
    echo "== annotate $v $f"
    $PERF annotate -i /tmp/perf-$v.data --stdio -s "$f" 2>/dev/null | awk '$1+0 >= 0.8' | head -25
  done
done

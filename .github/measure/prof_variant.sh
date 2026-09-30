#!/bin/bash
# Profiles the encode of a payload by master and by a variant ( VARIANT.patch ) for the same number of encodes.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
IGNORE='pretouchSonic|reportPayloads|newReportPayload|PretouchMany'
for v in master $VARIANT; do
  git worktree add -q /tmp/wt-$v origin/master
  [ $v = master ] || (cd /tmp/wt-$v && git apply $root/.github/measure/$v.patch)
  cp $root/.github/measure/zz_prof_test.go /tmp/wt-$v/benchmarks/
  (cd /tmp/wt-$v/benchmarks && go test -c -o /tmp/p-$v.test .)
done
cd /tmp/wt-master/benchmarks
export GOGC=off GOMEMLIMIT=2GiB
for c in ${CONFIGS:-std/openai}; do
  for v in master $VARIANT; do
    echo "== $c $v"
    ZZ_CONFIG=${c%%/*}/go-json ZZ_PAYLOAD=${c##*/} /tmp/p-$v.test -test.run '^$' -test.bench '^BenchmarkZZEncode$' -test.benchtime=${N:-1000000}x -test.cpuprofile /tmp/$v.prof | grep ns/op
    go tool pprof -ignore "$IGNORE" -top -nodecount 22 /tmp/p-$v.test /tmp/$v.prof 2>/dev/null | sed -n '6,28p' | awk '{printf "%8s %8s  %s\n", $1, $4, $6}'
    echo "-- lines of vm.Run"
    go tool pprof -ignore "$IGNORE" -lines -top -nodecount 600 /tmp/p-$v.test /tmp/$v.prof 2>/dev/null | grep -E 'encoder/vm/vm\.go:[0-9]+' | awk '{print $1, $7}' | sed -E 's#.*/vm\.go:##' | while read flat line; do
      c=$(awk -v l=$line 'NR<=l && /case encoder\.Op/ {c=$0} NR==l {print c; exit}' /tmp/wt-$v/internal/encoder/vm/vm.go | sed -E 's/^\s*case encoder\.//; s/, encoder\./,/g; s/:$//' | cut -c1-60)
      echo "$flat $c"
    done | awk '{f=$1; sub(/s$/,"",f); $1=""; a[$0]+=f} END {for (k in a) printf "%6.2fs %s\n", a[k], k}' | sort -rn | head -25
    echo "-- inlined helpers"
    go tool pprof -ignore "$IGNORE" -lines -top -nodecount 600 /tmp/p-$v.test /tmp/$v.prof 2>/dev/null | grep -E 'encoder/vm/util\.go' | awk '{n=split($6,f,"/"); m=split($7,g,"/"); printf "%8s  %s %s\n", $1, f[n], g[m]}' | head -10
  done
done

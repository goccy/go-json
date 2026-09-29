#!/bin/bash
# Profiles the encode of report payloads by go-json and sonic in the fastest configuration without GC, and the
# time of the inlined helpers of vm.Run which write the keys and the punctuation of the structs.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
git worktree add -q /tmp/wt origin/master
cp $root/.github/measure/zz_prof_test.go /tmp/wt/benchmarks/
cd /tmp/wt/benchmarks
go test -c -o /tmp/b.test .
IGNORE='pretouchSonic|reportPayloads|newReportPayload|PretouchMany'
export GOGC=off
for p in ${PAYLOADS:-anthropic openai small}; do
  for lib in go-json sonic; do
    echo "== $p fastest/$lib"
    ZZ_CONFIG=fastest/$lib ZZ_PAYLOAD=$p /tmp/b.test -test.run '^$' -test.bench '^BenchmarkZZEncode$' -test.benchtime=${N:-3000000}x -test.cpuprofile /tmp/$p-$lib.prof | grep ns/op
    go tool pprof -ignore "$IGNORE" -top -nodecount 25 /tmp/b.test /tmp/$p-$lib.prof 2>/dev/null | sed -n '6,31p' | awk '{printf "%8s %7s %8s %7s  %s\n", $1, $2, $4, $5, $6}'
  done
  echo "== $p go-json: the inlined helpers of the structs ( flat seconds by function )"
  go tool pprof -ignore "$IGNORE" -lines -top -nodecount 400 /tmp/b.test /tmp/$p-go-json.prof 2>/dev/null | awk '/vm\.(appendStruct|appendComma|appendColon|appendNull|store|load|ptrTo)/ {split($6,a,"."); f[$6]+=$1} END {for (k in f) printf "%8.2fs  %s\n", f[k], k}' | sort -rn
done

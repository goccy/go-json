#!/bin/bash
# Profiles the encode of a report payload by go-json and sonic in the fastest configuration, with the lines of
# vm.Run, to see what the keys of the structs cost.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
git worktree add -q /tmp/wt origin/master
cp $root/.github/measure/zz_prof_test.go /tmp/wt/benchmarks/
cd /tmp/wt/benchmarks
go test -c -o /tmp/b.test .
IGNORE='pretouchSonic|reportPayloads|newReportPayload|PretouchMany'
for p in ${PAYLOADS:-anthropic openai}; do
  for lib in go-json sonic; do
    echo "== $p fastest/$lib"
    ZZ_CONFIG=fastest/$lib ZZ_PAYLOAD=$p /tmp/b.test -test.run '^$' -test.bench '^BenchmarkZZEncode$' -test.benchtime=${N:-300000}x -test.cpuprofile /tmp/$p-$lib.prof | grep ns/op
    go tool pprof -ignore "$IGNORE" -top -nodecount 30 /tmp/b.test /tmp/$p-$lib.prof 2>/dev/null | sed -n '4,36p'
  done
  echo "== $p go-json: lines of vm.Run"
  go tool pprof -ignore "$IGNORE" -focus 'vm\.Run' -lines -top -nodecount 45 /tmp/b.test /tmp/$p-go-json.prof 2>/dev/null | sed -n '6,52p'
done

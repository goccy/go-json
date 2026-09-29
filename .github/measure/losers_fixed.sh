#!/bin/bash
# Profiles go-json on master for the payloads where it is behind sonic, with the encoder alone in the loop, and
# shows who calls memmove and the allocator, and the lines of AppendInt.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
git worktree add -q /tmp/wt-master origin/master
cp $root/.github/measure/zz_prof_test.go /tmp/wt-master/benchmarks/
cd /tmp/wt-master/benchmarks
go test -c -o /tmp/m.test .
IGNORE='pretouchSonic|reportPayloads|newReportPayload|PretouchMany'
while read cat p n; do
  for lib in go-json sonic; do
    echo "== profile $cat/$lib/encode/$p ${n}x"
    ZZ_CONFIG=$cat/$lib ZZ_PAYLOAD=$p /tmp/m.test -test.run '^$' -test.bench '^BenchmarkZZEncode$' -test.benchtime=${n}x -test.cpuprofile /tmp/p.prof | grep ns/op
    echo "-- peek memmove"
    go tool pprof -ignore "$IGNORE" -peek 'runtime\.memmove$' /tmp/m.test /tmp/p.prof 2>/dev/null | sed -n '/flat  flat%/,$p' | head -30
    echo "-- peek mallocgc"
    go tool pprof -ignore "$IGNORE" -peek 'runtime\.mallocgc$' /tmp/m.test /tmp/p.prof 2>/dev/null | sed -n '/flat  flat%/,$p' | head -20
  done
  if [ $p = small ]; then
    echo "-- list AppendInt"
    go tool pprof -ignore "$IGNORE" -list 'encoder\.AppendInt$' /tmp/m.test /tmp/p.prof 2>/dev/null | head -120
  fi
done <<'LIST'
fastest twitter 600000
fastest anthropic 2000000
fastest small 15000000
LIST

#!/bin/bash
# Profiles go-json and sonic on master for the same number of encodes of each payload where go-json is behind
# sonic, with the encoder alone in the loop, so that the seconds of the profiles are the costs of the encodes.
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
    go tool pprof -ignore "$IGNORE" -top -nodecount 30 /tmp/m.test /tmp/p.prof 2>/dev/null | sed -n '4,36p'
    echo "-- cum"
    go tool pprof -ignore "$IGNORE" -top -cum -nodecount 45 /tmp/m.test /tmp/p.prof 2>/dev/null | sed -n '6,51p'
  done
done <<'LIST'
fastest openai 500000
fast openai 500000
fastest anthropic 2000000
fastest small 15000000
fastest twitter 600000
fast twitter 600000
std twitter-any 90000
LIST

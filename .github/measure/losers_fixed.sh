#!/bin/bash
# Profiles go-json and sonic on master for the same number of encodes of each payload where go-json is behind
# sonic, so that the seconds of the profiles are the costs of an encode: the flat and cumulative tops of each.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
git worktree add -q /tmp/wt-master origin/master
cd /tmp/wt-master/benchmarks
go test -c -o /tmp/m.test .
while read cat p n; do
  for lib in go-json sonic; do
    echo "== profile $cat/$lib/encode/$p ${n}x"
    BENCH_REPORT_ONLY="^$cat/$lib/encode/$p\$" BENCH_REPORT_ROUNDS=1 BENCH_REPORT_OUT=/tmp/p.json \
      /tmp/m.test -test.run '^TestReport$' -test.count=1 -test.timeout 30m -test.benchtime=${n}x -test.cpuprofile /tmp/p.prof > /dev/null
    python3 -c 'import json; [print("result", r["condition"], r["nsPerOp"], r["allocsPerOp"], r["bytesPerOp"]) for r in json.load(open("/tmp/p.json"))["results"]]'
    go tool pprof -top -nodecount 30 /tmp/m.test /tmp/p.prof 2>/dev/null | sed -n '4,36p'
    echo "-- cum"
    go tool pprof -top -cum -nodecount 40 /tmp/m.test /tmp/p.prof 2>/dev/null | sed -n '6,46p'
  done
done <<'LIST'
fastest openai 40000
fast openai 40000
fastest anthropic 150000
fastest small 1500000
fastest twitter 60000
fast twitter 60000
std twitter-any 12000
LIST

#!/bin/bash
# Profiles the report encode of the payloads where go-json is behind sonic, on the string branch.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
git worktree add -q /tmp/wt-head origin/perf/encoder-string-escape
cd /tmp/wt-head/benchmarks
go test -c -o /tmp/head.test .
for sel in 'fast/go-json/encode/openai' 'fastest/go-json/encode/openai' 'fastest/go-json/encode/twitter' 'std/go-json/encode/twitter-any' 'fastest/sonic/encode/openai' 'fastest/sonic/encode/twitter'; do
  name=$(echo $sel | tr '/' '_')
  echo "== $sel"
  BENCH_REPORT_ONLY="^${sel}\$" BENCH_REPORT_ROUNDS=3 BENCH_REPORT_OUT=/tmp/$name.json /tmp/head.test -test.run '^TestReport$' -test.count=1 -test.timeout 60m -test.benchtime=2s -test.cpuprofile /tmp/$name.prof > /dev/null
  go tool pprof -top -nodecount 35 /tmp/head.test /tmp/$name.prof 2>/dev/null | sed -n '4,45p'
done

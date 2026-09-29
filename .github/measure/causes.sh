#!/bin/bash
# Finds the causes of the gaps of go-json to sonic on master: the report encode of openai, small and anthropic
# with several function layouts, then, on the first machine, profiles of the same number of encodes by each,
# limited to the encode ( -focus ), and the lines of go-json which take the time.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
git worktree add -q /tmp/wt-master origin/master
cd /tmp/wt-master/benchmarks
go test -c -o /tmp/m-0.test .
for l in 1 2 3; do go test -c -ldflags=-randlayout=$l -o /tmp/m-$l.test .; done
ONLY='^(fastest|fast|v2)/(go-json|sonic)/encode/(openai|small|anthropic)$'
for r in 1 2; do
  for l in 0 1 2 3; do
    BENCH_REPORT_ONLY="$ONLY" BENCH_REPORT_ROUNDS=3 BENCH_REPORT_OUT=/tmp/l-$l-$r.json /tmp/m-$l.test -test.run '^TestReport$' -test.count=1 -test.timeout 60m -test.benchtime=100ms > /dev/null
  done
done
python3 - <<'PY'
import collections, glob, json, os, statistics
res = collections.defaultdict(list)
for path in glob.glob('/tmp/l-*.json'):
    l = os.path.basename(path).split('-')[1]
    for x in json.load(open(path))['results']:
        cat, lib = x['config'].split('/')[:2]
        res[(l, x['condition'], cat, x['payload'], lib)].append(x['nsPerOp'])
print('== go-json against sonic by layout ( 0 is the default layout of the report )')
for cond in ['live-heap', 'no-live-heap']:
    for cat, p in [('fastest', 'openai'), ('fast', 'openai'), ('v2', 'openai'), ('fastest', 'small'), ('fastest', 'anthropic')]:
        row = f'{cond:13} {cat:8} {p:10}'
        for l in '0123':
            g, s = res.get((l, cond, cat, p, 'go-json')), res.get((l, cond, cat, p, 'sonic'))
            row += f'  L{l} {(statistics.median(g) / statistics.median(s) - 1) * 100:+6.1f}%' if g and s else '  -'
        print(row)
PY
if [ "$MACHINE" = 1 ]; then
  FOCUS='go-json\.marshal$|sonic\.frozenConfig\.Marshal$'
  while read cat p n; do
    for lib in go-json sonic; do
      echo "== profile $cat/$lib/encode/$p ${n}x"
      BENCH_REPORT_ONLY="^$cat/$lib/encode/$p\$" BENCH_REPORT_ROUNDS=1 BENCH_REPORT_OUT=/tmp/p.json \
        /tmp/m-0.test -test.run '^TestReport$' -test.count=1 -test.timeout 30m -test.benchtime=${n}x -test.cpuprofile /tmp/p.prof > /dev/null
      echo "-- flat"
      go tool pprof -focus "$FOCUS" -top -nodecount 30 /tmp/m-0.test /tmp/p.prof 2>/dev/null | sed -n '4,36p'
      echo "-- cum"
      go tool pprof -focus "$FOCUS" -top -cum -nodecount 30 /tmp/m-0.test /tmp/p.prof 2>/dev/null | sed -n '6,36p'
      if [ $lib = go-json ]; then
        echo "-- lines of vm.Run"
        go tool pprof -focus "$FOCUS" -lines -top -nodecount 40 /tmp/m-0.test /tmp/p.prof 2>/dev/null | grep -E 'vm\.Run|flat  flat%' | head -30
      fi
    done
  done <<'LIST'
fastest openai 300000
fast openai 300000
fastest small 8000000
fastest anthropic 1200000
LIST
fi

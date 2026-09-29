#!/bin/bash
# Measures go-json against sonic on master as the benchmark report does, lists where go-json is slower, and, on
# the first machine, profiles go-json and sonic on each of those.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
git worktree add -q /tmp/wt-master origin/master
cd /tmp/wt-master/benchmarks
go test -c -o /tmp/m.test .
BENCH_REPORT_ONLY='^(std|fast|v2|fastest)/(go-json|sonic)/(encode|decode)/' BENCH_REPORT_ROUNDS=5 BENCH_REPORT_OUT=/tmp/all.json \
  /tmp/m.test -test.run '^TestReport$' -test.count=1 -test.timeout 80m -test.benchtime=100ms > /dev/null
python3 - <<'PY'
import collections, json, statistics
res = collections.defaultdict(list)
for r in json.load(open('/tmp/all.json'))['results']:
    lib = r['config'].split('/')[1]
    res[(r['condition'], r['op'], r['payload'], r['config'].split('/')[0], lib)].append(r['nsPerOp'])
med = {k: statistics.median(v) for k, v in res.items()}
losers = {}
for op in ['encode', 'decode']:
    for cond in ['live-heap', 'no-live-heap']:
        print(f'== {op} {cond}: go-json against sonic')
        for cat in ['std', 'fast', 'v2', 'fastest']:
            row = f'{cat:8}'
            for p in ['small', 'medium', 'large', 'twitter', 'twitter-any', 'github', 'openai', 'anthropic', 'code']:
                g, s = med.get((cond, op, p, cat, 'go-json')), med.get((cond, op, p, cat, 'sonic'))
                if g and s:
                    d = (g / s - 1) * 100
                    row += f' {p}:{d:+.1f}%'
                    if d > 1:
                        k = (cat, op, p)
                        losers[k] = max(losers.get(k, -1e9), d)
            print(row)
with open('/tmp/losers.txt', 'w') as f:
    for (cat, op, p), d in sorted(losers.items(), key=lambda x: -x[1]):
        print(f'LOSER {cat} {op} {p} {d:+.1f}%')
        f.write(f'{cat} {op} {p}\n')
PY
if [ "$MACHINE" = 1 ]; then
  while read cat op p; do
    for lib in go-json sonic; do
      echo "== profile $cat/$lib/$op/$p"
      BENCH_REPORT_ONLY="^$cat/$lib/$op/$p\$" BENCH_REPORT_ROUNDS=3 BENCH_REPORT_OUT=/tmp/p.json \
        /tmp/m.test -test.run '^TestReport$' -test.count=1 -test.timeout 30m -test.benchtime=1s -test.cpuprofile /tmp/p.prof > /dev/null
      go tool pprof -top -nodecount 25 /tmp/m.test /tmp/p.prof 2>/dev/null | sed -n '6,32p'
    done
  done < /tmp/losers.txt
fi

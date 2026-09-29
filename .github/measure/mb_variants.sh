#!/bin/bash
# Tests and measures variants of the escape of strings with characters which are not ASCII ( utf8_<v>.patch ) against
# master, each also with a pad, in three function layouts, by the encode of Japanese strings; sonic for reference.
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD; LAYOUTS=${LAYOUTS:-3}; ROUNDS=${ROUNDS:-3}
names=""
for v in ${VARIANTS:-master u4 u5}; do
  for pad in "" pad; do
    n=$v$pad; names="$names $n"
    git worktree add -q /tmp/wt-$n origin/master
    [ $v = master ] || (cd /tmp/wt-$n && git apply $root/.github/measure/utf8_$v.patch)
    [ -z "$pad" ] || cp $root/.github/measure/pad.go /tmp/wt-$n/internal/encoder/zz_measure_pad.go
    if [ $v != master ] && [ -z "$pad" ]; then
      (cd /tmp/wt-$n && go test -count=1 -run 'TestAppendEscapedSIMD|TestAppendHTMLEscapedSIMD|TestScanStringAVX2|TestAppendString|TestUTF8|TestCommonRuneSize' ./internal/encoder/)
    fi
    cp $root/.github/measure/zz_multibyte_test.go $root/.github/measure/zz_prof_test.go /tmp/wt-$n/benchmarks/
    for l in $(seq 1 $LAYOUTS); do
      (cd /tmp/wt-$n/benchmarks && go test -c -ldflags=-randlayout=$l -o /tmp/mb-$n-$l.test .)
    done
  done
done
cd /tmp/wt-master/benchmarks
for r in $(seq 1 $ROUNDS); do
  for l in $(seq 1 $LAYOUTS); do
    for n in $names; do
      /tmp/mb-$n-$l.test -test.run '^$' -test.bench 'BenchmarkZZMultibyte/.*/.*/go-json$' -test.benchtime 100ms | awk -v n=$n -v l=$l '/ns\/op/ {print n, l, $1, $3}' >> /tmp/mb.txt
    done
    if [ $l = 1 ]; then
      /tmp/mb-master-1.test -test.run '^$' -test.bench 'BenchmarkZZMultibyte/.*/.*/sonic$' -test.benchtime 100ms | awk '/ns\/op/ {print "sonic", 1, $1, $3}' >> /tmp/mb.txt
    fi
  done
done
NAMES="$names" python3 - <<'PY'
import collections, os, re, statistics
names = os.environ['NAMES'].split()
best = {}
for line in open('/tmp/mb.txt'):
    n, l, b, v = line.split()
    b = re.sub(r'-\d+$', '', b.replace('BenchmarkZZMultibyte/', '')).rsplit('/', 1)[0]
    best[(n, l, b)] = min(best.get((n, l, b), 1e18), float(v))
mean = collections.defaultdict(list)
for (n, l, b), v in best.items():
    mean[(n, b)].append(v)
cases = sorted({b for (_, b) in mean})
last = names[-2]
print(f"{'config/case':20}" + ''.join(f"{n:>10}" for n in names) + f"{'master/sonic':>13}{last + '/sonic':>12}")
for b in cases:
    m = statistics.mean(mean[('master', b)]); s = statistics.mean(mean[('sonic', b)])
    print(f"{b:20}" + ''.join(f"{(statistics.mean(mean[(n, b)]) / m - 1) * 100:+9.1f}%" for n in names) +
          f"{(m / s - 1) * 100:+12.1f}%{(statistics.mean(mean[(last, b)]) / s - 1) * 100:+11.1f}%")
PY

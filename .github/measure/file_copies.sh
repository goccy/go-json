#!/bin/bash
# A throwaway script of a measure/* branch: does the cluster of a small encode benchmark follow the binary file,
# rather than the code in it? The gate of PR 744 measured identical binaries ( same code, same layout, the same
# addresses and sizes of every symbol ) 229 and 181 ns/op at their fastest.
#
# For 4 layouts, one binary is built and copied 5 times ( new files: new pages of the page cache ). Every copy is
# measured twice, then the page cache is dropped and every copy is measured again.
set -u
lscpu | grep 'Model name'
uname -r
echo "THP: $(cat /sys/kernel/mm/transparent_hugepage/enabled)"
B=/tmp/b
mkdir -p $B
cd "$GITHUB_WORKSPACE/benchmarks" || exit 1
for seed in 0 1 2 3; do
	F=""
	[ $seed -ne 0 ] && F="-ldflags=-randlayout=$seed"
	go test -c -o "$B/s$seed.c0.test" $F . || exit 1
	for c in 1 2 3 4 5; do
		cp "$B/s$seed.c0.test" "$B/s$seed.c$c.test"
	done
done
cmp "$B/s1.c0.test" "$B/s1.c5.test" && echo "copies are identical"
BENCH='^(Benchmark_Encode_SmallStructCached_GoJson(OptimizedFieldOrder|LikeSonic)?|Benchmark_Encode_SmallStruct_GoJsonMarshalValue)$'
measure() { # round
	for seed in 0 1 2 3; do
		for c in 0 1 2 3 4 5; do
			"$B/s$seed.c$c.test" -test.run '^$' -test.bench "$BENCH" -test.benchtime 100ms -test.count 1 |
				awk -v s=$seed -v c=$c -v r=$1 '/^Benchmark/ {sub(/-[0-9]+$/, "", $1); print "RAW", s, c, r, $1, $3}'
		done
	done
}
measure 1
measure 2
# how the file is mapped while it runs: huge pages of the text?
"$B/s1.c0.test" -test.run '^$' -test.bench 'Benchmark_Encode_SmallStructCached_GoJson$' -test.benchtime 3s > /dev/null &
pid=$!
sleep 1.5
grep -E 'FilePmdMapped|AnonHugePages|Rss' /proc/$pid/smaps_rollup | sed 's/^/SMAPS /'
wait $pid
sync
echo 3 | sudo tee /proc/sys/vm/drop_caches > /dev/null && echo "dropped the page cache"
measure 3

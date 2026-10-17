#!/bin/bash
# A throwaway script of a measure/* branch: how the benchmarks of the encode group of the gate spread over function
# layouts for identical code, to choose how the gate compares the layouts of the two sides.
#
# Builds the benchmarks of this commit with 24 layouts ( the seed 0 is the linker's own ), and measures each
# top-level function of the shard $1 of 5 with every layout in 2 rounds of 100ms. Prints every result as
# RAW <function> <layout> <round> <benchmark> <ns/op>.
set -u
lscpu | grep 'Model name'
SHARD=$1
LAYOUTS=24
B=/tmp/b
mkdir -p $B
cd "$GITHUB_WORKSPACE/benchmarks" || exit 1
seq 0 $((LAYOUTS - 1)) | xargs -P 4 -I{} sh -c 'if [ {} -eq 0 ]; then go test -c -o /tmp/b/l{}.test .; else go test -c -o /tmp/b/l{}.test -ldflags=-randlayout={} .; fi' || exit 1
echo "built"

# the functions of the shard, as the gate chooses them: go-json's benchmarks of the encode group, sorted, dealt
# to the shards one by one.
COMPARED='EasyJson|EncodingJson|FFJson|FastJson|GoJay|GoJayUnsafe|Jettison|JsonIter|SegmentioJson|Sonic|SonicFast|SonicFastest|SonicStd|StdLib'
$B/l0.test -test.list . | grep '^Benchmark' | grep -Ev "_($COMPARED)\$" | grep -v Decode | grep -v Unmarshal | LC_ALL=C sort |
	awk -v s=$SHARD '(NR - 1) % 5 == s' > $B/funcs.txt
echo "functions: $(wc -l < $B/funcs.txt)"

start=$(date +%s)
while read -r fn; do
	for round in 1 2; do
		for L in $(seq 0 $((LAYOUTS - 1))); do
			$B/l$L.test -test.run '^$' -test.bench "^${fn}\$" -test.benchtime 100ms -test.count 1 |
				awk -v f="$fn" -v l=$L -v r=$round '/^Benchmark/ && $4 == "ns/op" {sub(/-[0-9]+$/, "", $1); print "RAW", f, l, r, $1, $3}'
		done
	done
done < $B/funcs.txt
echo "measured in $(( $(date +%s) - start ))s"

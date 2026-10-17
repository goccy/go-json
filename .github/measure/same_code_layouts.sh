#!/bin/bash
# A throwaway script of a measure/* branch: how the benchmarks of the encode group of the gate spread over function
# layouts for identical code, with the default function alignment of the linker ( 32 bytes on amd64 ) and with
# -funcalign=64, to tell whether the alignment of the functions within a 64-byte line is what makes them bimodal.
#
# Builds the benchmarks of this commit with 24 layouts ( the seed 0 is the linker's own ) for each alignment, and
# measures each top-level function of the shard $1 of 5 with every layout of both in 2 rounds of 100ms.
# Prints every result as RAW <function> <layout> <round> <benchmark> <ns/op>, the layout being a32-<seed> or
# a64-<seed>, and the address modulo 64 of the hot functions of every a32 layout as ALIGN <layout> <name> <mod 64>.
set -u
lscpu | grep 'Model name'
SHARD=$1
LAYOUTS=24
B=/tmp/b
mkdir -p $B
cd "$GITHUB_WORKSPACE/benchmarks" || exit 1
build() { # align seed
	F="-funcalign=$1"
	[ "$2" -ne 0 ] && F="$F -randlayout=$2"
	go test -c -o "$B/a$1-$2.test" -ldflags="$F" .
}
export -f build
export B
for a in 32 64; do
	seq 0 $((LAYOUTS - 1)) | xargs -P 4 -I{} bash -c "build $a {}" || exit 1
done
echo "built"
for s in $(seq 0 $((LAYOUTS - 1))); do
	go tool nm -size -sort address "$B/a32-$s.test" |
		grep -E ' T github\.com/goccy/go-json/(internal/encoder/vm\.Run|internal/encoder\.AppendInt|internal/encoder\.AppendFloat64|internal/jsonstring\.AppendQuoted|internal/encoder/run\.Encode|marshal|internal/encoder\.CompileToGetCodeSet)$' |
		while read -r addr size kind name; do
			echo "ALIGN a32-$s $name $((16#$addr % 64)) $addr"
		done
done

# the functions of the shard, as the gate chooses them: go-json's benchmarks of the encode group, sorted, dealt
# to the shards one by one.
COMPARED='EasyJson|EncodingJson|FFJson|FastJson|GoJay|GoJayUnsafe|Jettison|JsonIter|SegmentioJson|Sonic|SonicFast|SonicFastest|SonicStd|StdLib'
$B/a32-0.test -test.list . | grep '^Benchmark' | grep -Ev "_($COMPARED)\$" | grep -v Decode | grep -v Unmarshal | LC_ALL=C sort |
	awk -v s=$SHARD '(NR - 1) % 5 == s' > $B/funcs.txt
echo "functions: $(wc -l < $B/funcs.txt)"

start=$(date +%s)
while read -r fn; do
	for round in 1 2; do
		for s in $(seq 0 $((LAYOUTS - 1))); do
			for a in 32 64; do
				$B/a$a-$s.test -test.run '^$' -test.bench "^${fn}\$" -test.benchtime 100ms -test.count 1 |
					awk -v f="$fn" -v l="a$a-$s" -v r=$round '/^Benchmark/ && $4 == "ns/op" {sub(/-[0-9]+$/, "", $1); print "RAW", f, l, r, $1, $3}'
			done
		done
	done
done < $B/funcs.txt
echo "measured in $(( $(date +%s) - start ))s"

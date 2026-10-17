#!/bin/bash
# A throwaway script of a measure/* branch: why the benchmark gate finds
# Benchmark_Encode_SmallStructCached_GoJsonOptimizedFieldOrder about +32% slower on PR 743, whose code doesn't run in it.
#
# The gate builds the base with the library in a temporary directory ( -modfile ) and the head with the library in
# the workspace. This builds both commits in both ways, with the 8 layouts of the gate, measures them interleaved,
# and profiles them, to tell whether the difference follows the commit or the way it is built.
set -u
lscpu | grep 'Model name'
W=$GITHUB_WORKSPACE
B=/tmp/b
mkdir -p $B
BASE=f551f75
HEAD=be1df5d
BENCH='^Benchmark_Encode_SmallStructCached_GoJson(OptimizedFieldOrder)?$'
git fetch -q origin fix/decoder-embedded-ancestor

# the library in a temporary directory of the same length as the gate's, the benchmarks of the workspace.
build_tmp() { # name commit dir
	git worktree add -q "$3/base" "$2"
	cp "$W/benchmarks/go.mod" "$W/benchmarks/go.sum" "$3/"
	(cd "$W/benchmarks" && go mod edit -replace "github.com/goccy/go-json=$3/base" "$3/go.mod")
	for L in 0 1 2 3 4 5 6 7; do
		F=""
		[ $L -ne 0 ] && F="-ldflags=-randlayout=$L"
		(cd "$W/benchmarks" && go test -c -o "$B/$1.$L.test" $F -modfile "$3/go.mod" .) || exit 1
	done
}
# the library in the workspace.
build_ws() { # name commit
	git checkout -q "$2"
	for L in 0 1 2 3 4 5 6 7; do
		F=""
		[ $L -ne 0 ] && F="-ldflags=-randlayout=$L"
		(cd "$W/benchmarks" && go test -c -o "$B/$1.$L.test" $F .) || exit 1
	done
}
build_ws head_ws $HEAD
build_ws base_ws $BASE
build_tmp base_tmp $BASE /tmp/benchcheck-1234567890
build_tmp head_tmp $HEAD /tmp/benchcheck-1234567891
VARIANTS="base_tmp head_ws base_ws head_tmp"
echo "built"

cd "$W/benchmarks"
for r in 1 2 3; do
	for L in 0 1 2 3 4 5 6 7; do
		for v in $VARIANTS; do
			"$B/$v.$L.test" -test.run '^$' -test.bench "$BENCH" -test.benchtime 200ms -test.count 1 -test.benchmem | grep '^Benchmark' | sed "s/^/$v $L /"
		done
	done
done > $B/results.txt
echo "== fastest of 3 rounds: variant layout benchmark ns/op"
awk '{k=$1" "$2" "$3; if(!(k in m)||$5<m[k])m[k]=$5} END{for(k in m)print k, m[k]}' $B/results.txt | sort -k3,3 -k1,1 -k2,2n
echo "== fastest of all layouts: variant benchmark ns/op"
awk '{k=$1" "$3; if(!(k in m)||$5<m[k])m[k]=$5} END{for(k in m)print k, m[k]}' $B/results.txt | sort -k2,2 -k1,1
echo "== allocations"
grep ' 0 ' $B/results.txt | awk '{print $1, $3, $7, $8, $9, $10}' | sort -u

for v in $VARIANTS; do
	"$B/$v.0.test" -test.run '^$' -test.bench '^Benchmark_Encode_SmallStructCached_GoJsonOptimizedFieldOrder$' -test.benchtime 3s -test.cpuprofile "$B/$v.prof" | grep '^Benchmark'
	echo "== $v profile"
	go tool pprof -top -nodecount 20 "$B/$v.0.test" "$B/$v.prof" 2>/dev/null | tail -n +5
done
for v in $VARIANTS; do
	for fn in 'vm\.Run$' 'runtime\.mallocgc' 'runtime\.memmove' 'encoder\.AppendInt$' 'jsonstring\.'; do
		echo "== $v disasm $fn ( instructions with samples )"
		go tool pprof -disasm "$fn" "$B/$v.0.test" "$B/$v.prof" 2>/dev/null | awk '$1 != "." || $2 != "." ' | head -80
	done
done
for v in $VARIANTS; do
	echo "== $v sections"
	readelf -SW "$B/$v.0.test" | grep -E ' \.(text|rodata|typelink|itablink|gopclntab|noptrdata|data|bss|noptrbss) '
	echo "== $v symbols"
	go tool nm -size -sort address "$B/$v.0.test" | grep -E 'SmallPayload|go:string\."(TestString|8f9a65eb|127\.0\.0\.1|Mozilla)|cachedOpcodeSets|runtimeContextPool|OptimizeFieldOrder|vm\.Run$|encoder\.AppendInt$'
done

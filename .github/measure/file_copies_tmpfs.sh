#!/bin/bash
# A throwaway script of a measure/* branch: the cluster of a small encode benchmark follows where the page cache holds
# the binary ( see file_copies.sh ). Does it go away when the text of the binary is mapped by huge pages?
#
# The binaries of 4 layouts are copied 4 times into the disk, a tmpfs with huge=always ( the text, at 0x400000, can
# be mapped by 2 MB pages ) and a tmpfs with huge=never, and every copy is measured twice, the file systems in turn.
set -u
lscpu | grep 'Model name'
uname -r
echo "THP: $(cat /sys/kernel/mm/transparent_hugepage/enabled), shmem: $(cat /sys/kernel/mm/transparent_hugepage/shmem_enabled)"
B=/tmp/b
mkdir -p $B /tmp/fs/disk /tmp/fs/huge /tmp/fs/none
sudo mount -t tmpfs -o size=2G,huge=always tmpfs /tmp/fs/huge || exit 1
sudo mount -t tmpfs -o size=2G,huge=never tmpfs /tmp/fs/none || exit 1
sudo chmod 777 /tmp/fs/huge /tmp/fs/none
cd "$GITHUB_WORKSPACE/benchmarks" || exit 1
for seed in 0 1 2 3; do
	F=""
	[ $seed -ne 0 ] && F="-ldflags=-randlayout=$seed"
	go test -c -o "$B/s$seed.test" $F . || exit 1
	for c in 0 1 2 3; do
		for fs in disk huge none; do
			cp "$B/s$seed.test" "/tmp/fs/$fs/s$seed.c$c.test"
		done
	done
done
# how the text is mapped while a binary runs.
for fs in disk huge none; do
	"/tmp/fs/$fs/s1.c0.test" -test.run '^$' -test.bench 'Benchmark_Encode_SmallStructCached_GoJson$' -test.benchtime 3s > /dev/null &
	pid=$!
	sleep 1.5
	echo "SMAPS $fs $(grep -E 'FilePmdMapped|ShmemPmdMapped' /proc/$pid/smaps_rollup | tr -s ' ' | tr '\n' ' ')"
	wait $pid
done
BENCH='^(Benchmark_Encode_SmallStructCached_GoJson(OptimizedFieldOrder|LikeSonic)?|Benchmark_Encode_SmallStruct_GoJsonMarshalValue)$'
for r in 1 2; do
	for seed in 0 1 2 3; do
		for c in 0 1 2 3; do
			for fs in disk huge none; do
				"/tmp/fs/$fs/s$seed.c$c.test" -test.run '^$' -test.bench "$BENCH" -test.benchtime 100ms -test.count 1 |
					awk -v f=$fs -v s=$seed -v c=$c -v r=$r '/^Benchmark/ {sub(/-[0-9]+$/, "", $1); print "RAW", f, s, c, r, $1, $3}'
			done
		done
	done
done

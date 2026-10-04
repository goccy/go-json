#!/bin/bash
# measure-stt-tree.sh <master|v3> <k>: makes /tmp/<name>-<k>, a tree of the commit with k unused types and
# functions added to the encoder package, which move the type descriptors and the code of the binary, and the test
# which prints the sets of the map benchmark types.
set -e
name=$1
k=$2
case $name in
master) sha=$MASTER ;;
v3) sha=$V3 ;;
esac
dir=/tmp/$name-$k
rm -rf "$dir"
git worktree add -f "$dir" "$sha" > /dev/null 2>&1
cp /tmp/zz_sets_test.go "$dir/benchmarks/"
if [ "$k" -gt 0 ]; then
	{
		echo 'package encoder'
		echo
		echo 'import "reflect"'
		echo
		for i in $(seq 1 "$k"); do
			echo "type aaPad$i struct{ A [$i]int64 }"
			echo
			echo "//go:noinline"
			echo "func aaPadFunc$i(n int) int { return n*$i + 1 }"
			echo
		done
		echo 'var aaPadSink = []any{'
		for i in $(seq 1 "$k"); do
			echo "	reflect.TypeOf(aaPad$i{}), aaPadFunc$i(1),"
		done
		echo '}'
	} > "$dir/internal/encoder/aa_pad.go"
fi

#!/bin/bash
# Temporary: measures the SmallStruct and MediumStruct decode benchmarks of master, of commits of the branch and of
# variants of the head ( a ref with +name is the ref with .github/workflows/variants/name.patch applied ) on one machine.
set -e
export SONIC_USE_OPTDEC=1
grep -m1 "model name" /proc/cpuinfo
refs="master HEAD HEAD+pad1 HEAD+pad2 HEAD+fuse3 HEAD+fuse3pad2 HEAD+all3"
layouts=3
out=$PWD/.bisect
mkdir -p $out
bench='^Benchmark_Decode_(OpenAIResponse|AnthropicMessage|OpenAIChatCompletion|GitHubREST|GitHubGraphQL|TwitterBinding|TwitterGeneric|LargeStruct|MediumStruct|SmallStruct)_Unmarshal_(GoJson|GoJsonLikeSonic|GoJsonUnmarshalOfNoCopyString|SonicFastestValidating)$|^BenchmarkCodeUnmarshal$|^BenchmarkUnicodeDecoder$|^Benchmark_Decode_LargeSlice_EscapedString_GoJson$|^Benchmark_Decode_(OpenAIResponse|AnthropicMessage)Stream_Unmarshal_GoJson$'
name() { local n=${1//^/_}; echo ${n//+/_}; }
for r in $refs; do
  d=$(name $r)
  base=${r%%+*}
  git worktree add -q $out/wt-$d ${base/master/origin/master}
  # the benchmarks of the head, which older refs may not have
  cp benchmarks/*.go $out/wt-$d/benchmarks/
  if [ "$r" != "$base" ]; then
    [ -s $PWD/.github/workflows/variants/${r#*+}.patch ] && git -C $out/wt-$d apply $PWD/.github/workflows/variants/${r#*+}.patch
  fi
  for l in $(seq 0 $((layouts-1))); do
    flags=""
    [ $l -ne 0 ] && flags="-ldflags=-randlayout=$l"
    (cd $out/wt-$d/benchmarks && go test -c -o $out/$d.$l.test $flags .)
  done
done
cd benchmarks
for round in 1 2 3 4; do
  for l in $(seq 0 $((layouts-1))); do
    for r in $refs; do
      $out/$(name $r).$l.test -test.run '^$' -test.bench "$bench" -test.benchtime 200ms \
        | awk -v c=$r -v l=$l '/ns\/op/ {print c, l, $1, $3}' >> $out/results.txt
    done
  done
done
# the fastest round of every layout, then the mean over the layouts
awk '{k=$1" "$2" "$3; if (!(k in m) || $4 < m[k]) m[k]=$4} END {for (k in m) {split(k,a," "); s[a[1]" "a[3]]+=m[k]; n[a[1]" "a[3]]++} for (k in s) printf "%s %.1f\n", k, s[k]/n[k]}' $out/results.txt | sort -k2,2 -k1,1 | awk '{printf "%-16s %-75s %s\n", $1, $2, $3}'


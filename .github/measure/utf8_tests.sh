#!/bin/bash
# Runs the tests of the escapes of strings on master with each variant ( utf8_<v>.patch ), and reports each result.
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
for v in ${VARIANTS:-u11}; do
  git worktree add -q /tmp/wt-$v origin/master
  (cd /tmp/wt-$v && git apply $root/.github/measure/utf8_$v.patch)
  echo "== $v"
  (cd /tmp/wt-$v && go test -count=1 -run 'TestEscapeUTF8AVX2AcrossBlocks|TestAppendEscapedSIMD|TestAppendHTMLEscapedSIMD|TestScanStringAVX2|TestAppendString|TestUTF8|TestCommonRuneSize|TestAppendNormalizedStrings' ./internal/encoder/ 2>&1 | grep -E '^(--- FAIL|ok|FAIL)|_test.go:[0-9]+:' | head -6 | cut -c1-200)
done
(cd /tmp/wt-u11 && go test -count=1 ./... 2>&1 | grep -E '^(--- FAIL|FAIL|ok)' | head -20)

#!/bin/bash
# Counts the paths of escapeUTF8AVX2 which the tests of the escapes take ( utf8_counts.patch ).
set -e
lscpu | sed -n 's/^Model name: *//p' | head -1
root=$PWD
git worktree add -q /tmp/wt-counts origin/master
cd /tmp/wt-counts && git apply $root/.github/measure/utf8_counts.patch
go test -count=1 -v -run 'TestEscapeUTF8AVX2AcrossBlocks|TestAppendEscapedSIMD|TestAppendHTMLEscapedSIMD|TestAppendNormalizedStrings$|TestAppendStringFastPath|TestZZZUTF8LoopCounts' ./internal/encoder/ | grep -E '^(--- |ok|FAIL)|zz_utf8_counts_test.go'

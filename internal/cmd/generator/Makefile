.PHONY: asm
asm:
	clang -S -O2 -mavx2 -masm=intel -mno-red-zone -mstackrealign -mllvm -inline-threshold=1000 -fno-asynchronous-unwind-tables -fno-exceptions -fno-rtti -c ./simd/string.c

.PHONY: generate
generate:
	docker build -f Dockerfile.simd -t go-json-simd .
	docker run --rm -v "$(CURDIR):/tmp" go-json-simd bash -c "cp /work/string_avx.s /tmp/string_avx.s"

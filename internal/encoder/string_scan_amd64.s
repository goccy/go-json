#include "textflag.h"

// func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)
TEXT ·cpuid(SB), NOSPLIT, $0-24
	MOVL eaxArg+0(FP), AX
	MOVL ecxArg+4(FP), CX
	CPUID
	MOVL AX, eax+8(FP)
	MOVL BX, ebx+12(FP)
	MOVL CX, ecx+16(FP)
	MOVL DX, edx+20(FP)
	RET

// func xgetbv() (eax, edx uint32)
TEXT ·xgetbv(SB), NOSPLIT, $0-8
	MOVL $0, CX
	XGETBV
	MOVL AX, eax+0(FP)
	MOVL DX, edx+4(FP)
	RET

// func scanStringAVX2(p unsafe.Pointer, n int, chars uint64, high uint64) int
//
// It returns the index of the first byte of the n bytes at p which may need an escape, or n. Such a byte is
// a control character, '"', '\', one of the three characters in chars ( a byte each ), and a byte which is not
// ASCII if high is not zero. n is 32 or more. The last block overlaps the previous one.
TEXT ·scanStringAVX2(SB), NOSPLIT, $8-40
	MOVQ p+0(FP), SI
	MOVQ n+8(FP), CX
	MOVQ chars+16(FP), AX
	MOVQ high+24(FP), R9

	// the constants, a byte in every lane, from memory: a move from a general register is a legacy SSE
	// instruction, and mixing it with the 256-bit instructions cost about 500 ns per call.
	VPBROADCASTB c1f<>(SB), Y8      // 0x1f: a byte less than 0x20 is a control character
	VPBROADCASTB cquote<>(SB), Y9
	VPBROADCASTB cbslash<>(SB), Y10
	MOVQ         AX, chars-8(SP)
	VPBROADCASTB chars-8(SP), Y11   // chars[0]
	VPBROADCASTB chars-7(SP), Y12   // chars[1]
	VPBROADCASTB chars-6(SP), Y13   // chars[2]

	// the mask of the lanes to keep from the sign bits: all or none.
	XORL  R10, R10
	TESTQ R9, R9
	JEQ   nohigh
	MOVL  $0xffffffff, R10
nohigh:

	XORQ  DI, DI                 // the index of the block
	LEAQ  -32(CX), R11           // the index of the last block
loop:
	VMOVDQU (SI)(DI*1), Y0
	VPMAXUB Y8, Y0, Y1
	VPCMPEQB Y8, Y1, Y1          // x <= 0x1f
	VPCMPEQB Y9, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPCMPEQB Y10, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPCMPEQB Y11, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPCMPEQB Y12, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPCMPEQB Y13, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPMOVMSKB Y1, AX
	VPMOVMSKB Y0, DX             // the bytes which are not ASCII
	ANDL      R10, DX
	ORL       DX, AX
	JNE       found
	ADDQ      $32, DI
	CMPQ      DI, R11
	JLT       loop
	// the last block, which overlaps the previous one unless the blocks ended exactly at n.
	CMPQ      DI, CX
	JEQ       none
	MOVQ      R11, DI
	VMOVDQU (SI)(DI*1), Y0
	VPMAXUB Y8, Y0, Y1
	VPCMPEQB Y8, Y1, Y1
	VPCMPEQB Y9, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPCMPEQB Y10, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPCMPEQB Y11, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPCMPEQB Y12, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPCMPEQB Y13, Y0, Y2
	VPOR     Y2, Y1, Y1
	VPMOVMSKB Y1, AX
	VPMOVMSKB Y0, DX
	ANDL      R10, DX
	ORL       DX, AX
	JNE       found
none:
	VZEROUPPER
	MOVQ CX, ret+32(FP)
	RET
found:
	VZEROUPPER
	TZCNTL AX, AX
	ADDQ   DI, AX
	MOVQ   AX, ret+32(FP)
	RET

DATA c1f<>+0(SB)/1, $0x1f
GLOBL c1f<>(SB), RODATA|NOPTR, $1
DATA cquote<>+0(SB)/1, $0x22
GLOBL cquote<>(SB), RODATA|NOPTR, $1
DATA cbslash<>+0(SB)/1, $0x5c
GLOBL cbslash<>(SB), RODATA|NOPTR, $1

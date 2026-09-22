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
TEXT ·scanStringAVX2(SB), NOSPLIT, $0-40
	MOVQ p+0(FP), SI
	MOVQ n+8(FP), CX
	MOVQ chars+16(FP), AX
	MOVQ high+24(FP), R9

	// the constants, a byte in every lane.
	MOVQ         $0x1f, DX
	MOVQ         DX, X8
	VPBROADCASTB X8, Y8          // 0x1f: a byte less than 0x20 is a control character
	MOVQ         $'"', DX
	MOVQ         DX, X9
	VPBROADCASTB X9, Y9
	MOVQ         $'\\', DX
	MOVQ         DX, X10
	VPBROADCASTB X10, Y10
	MOVQ         AX, X11
	VPBROADCASTB X11, Y11        // chars[0]
	MOVQ         AX, DX
	SHRQ         $8, DX
	MOVQ         DX, X12
	VPBROADCASTB X12, Y12        // chars[1]
	SHRQ         $8, DX
	MOVQ         DX, X13
	VPBROADCASTB X13, Y13        // chars[2]

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

// func scanStringSSE(p unsafe.Pointer, n int, chars uint64, high uint64) int
//
// scanStringAVX2 with 128-bit registers, VEX encoded: a variant to find why the 256-bit one costs a fixed
// time per call.
TEXT ·scanStringSSE(SB), NOSPLIT, $8-40
	MOVQ p+0(FP), SI
	MOVQ n+8(FP), CX
	MOVQ chars+16(FP), AX
	MOVQ high+24(FP), R9

	VPBROADCASTB c1f<>(SB), X8
	VPBROADCASTB cquote<>(SB), X9
	VPBROADCASTB cbslash<>(SB), X10
	MOVQ AX, chars-8(SP)
	VPBROADCASTB chars-8(SP), X11
	VPBROADCASTB chars-7(SP), X12
	VPBROADCASTB chars-6(SP), X13

	XORL  R10, R10
	TESTQ R9, R9
	JEQ   nohigh16
	MOVL  $0xffffffff, R10
nohigh16:
	XORQ  DI, DI
	LEAQ  -16(CX), R11
loop16:
	VMOVDQU (SI)(DI*1), X0
	VPMAXUB X8, X0, X1
	VPCMPEQB X8, X1, X1
	VPCMPEQB X9, X0, X2
	VPOR     X2, X1, X1
	VPCMPEQB X10, X0, X2
	VPOR     X2, X1, X1
	VPCMPEQB X11, X0, X2
	VPOR     X2, X1, X1
	VPCMPEQB X12, X0, X2
	VPOR     X2, X1, X1
	VPCMPEQB X13, X0, X2
	VPOR     X2, X1, X1
	VPMOVMSKB X1, AX
	VPMOVMSKB X0, DX
	ANDL      R10, DX
	ORL       DX, AX
	JNE       found16
	ADDQ      $16, DI
	CMPQ      DI, R11
	JLT       loop16
	CMPQ      DI, CX
	JEQ       none16
	MOVQ      R11, DI
	VMOVDQU (SI)(DI*1), X0
	VPMAXUB X8, X0, X1
	VPCMPEQB X8, X1, X1
	VPCMPEQB X9, X0, X2
	VPOR     X2, X1, X1
	VPCMPEQB X10, X0, X2
	VPOR     X2, X1, X1
	VPCMPEQB X11, X0, X2
	VPOR     X2, X1, X1
	VPCMPEQB X12, X0, X2
	VPOR     X2, X1, X1
	VPCMPEQB X13, X0, X2
	VPOR     X2, X1, X1
	VPMOVMSKB X1, AX
	VPMOVMSKB X0, DX
	ANDL      R10, DX
	ORL       DX, AX
	JNE       found16
none16:
	MOVQ CX, ret+32(FP)
	RET
found16:
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

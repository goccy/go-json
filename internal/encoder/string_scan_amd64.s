#include "textflag.h"

// MASK computes into Yd the mask of the bytes of Yx which may need an escape, using Yt as a temporary:
// the lane of such a byte is not zero. See nibbleTables: Y8 is the table of the low nibbles in both halves,
// Y9 the one of the high nibbles and Y10 has 0x0f in every lane.
//
// Only VEX instructions are used: a legacy SSE instruction, such as a move from a general register to an XMM
// register, mixed with them cost about 500 ns per call.
#define MASK(Yx, Yt, Yd) \
	VPAND   Y10, Yx, Yt \
	VPSRLW  $4, Yx, Yd \
	VPAND   Y10, Yd, Yd \
	VPSHUFB Yt, Y8, Yt \
	VPSHUFB Yd, Y9, Yd \
	VPAND   Yt, Yd, Yd

// func scanStringAVX2(p unsafe.Pointer, n int, tables *nibbleTables) int
//
// It returns 1 if a byte of the n bytes at p may need an escape by the tables, or 0. n is 32 or more.
// The blocks of 128 bytes are looked at first, then the blocks of 32 bytes, and the last block overlaps the
// previous one.
TEXT ·scanStringAVX2(SB), NOSPLIT, $0-32
	MOVQ p+0(FP), SI
	MOVQ n+8(FP), CX
	MOVQ tables+16(FP), AX

	VBROADCASTI128 (AX), Y8
	VBROADCASTI128 16(AX), Y9
	VPBROADCASTB   c0f<>(SB), Y10

	MOVQ SI, DI                  // the address of the block
	LEAQ -128(SI)(CX*1), R11     // the address of the last block of 128 bytes
	CMPQ DI, R11
	JGT  small
loop128:
	VMOVDQU (DI), Y0
	VMOVDQU 32(DI), Y1
	VMOVDQU 64(DI), Y2
	VMOVDQU 96(DI), Y3
	MASK(Y0, Y4, Y5)
	MASK(Y1, Y4, Y6)
	VPOR    Y6, Y5, Y5
	MASK(Y2, Y4, Y6)
	VPOR    Y6, Y5, Y5
	MASK(Y3, Y4, Y6)
	VPOR    Y6, Y5, Y5
	VPTEST  Y5, Y5
	JNE     found
	ADDQ    $128, DI
	CMPQ    DI, R11
	JLE     loop128
small:
	LEAQ -32(SI)(CX*1), R11      // the address of the last block of 32 bytes
	CMPQ DI, R11
	JGT  last
loop32:
	VMOVDQU (DI), Y0
	MASK(Y0, Y4, Y5)
	VPTEST  Y5, Y5
	JNE     found
	ADDQ    $32, DI
	CMPQ    DI, R11
	JLE     loop32
last:
	// the last block, which overlaps the previous one unless the blocks ended exactly at n.
	LEAQ (SI)(CX*1), DX
	CMPQ DI, DX
	JEQ  none
	VMOVDQU (R11), Y0
	MASK(Y0, Y4, Y5)
	VPTEST  Y5, Y5
	JNE     found
none:
	VZEROUPPER
	MOVQ $0, ret+24(FP)
	RET
found:
	VZEROUPPER
	MOVQ $1, ret+24(FP)
	RET

DATA c0f<>+0(SB)/1, $0x0f
GLOBL c0f<>(SB), RODATA|NOPTR, $1

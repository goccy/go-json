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

// func indexEscapeAVX2(p unsafe.Pointer, n int, tables *nibbleTables) int
//
// It returns the index of the first byte of the n bytes at p which may need an escape by the tables, or n if
// there is none. n is 32 or more. The blocks of 32 bytes are looked at in order, and the last block overlaps the
// previous one.
TEXT ·indexEscapeAVX2(SB), NOSPLIT, $0-32
	MOVQ p+0(FP), SI
	MOVQ n+8(FP), CX
	MOVQ tables+16(FP), AX

	VBROADCASTI128 (AX), Y8
	VBROADCASTI128 16(AX), Y9
	VPBROADCASTB   c0f<>(SB), Y10
	VPXOR          Y11, Y11, Y11

	MOVQ SI, DI                  // the address of the block
	LEAQ -32(SI)(CX*1), R11      // the address of the last block
iloop:
	CMPQ DI, R11
	JGT  ilast
	VMOVDQU (DI), Y0
	MASK(Y0, Y4, Y5)
	VPTEST  Y5, Y5
	JNE     ifound
	ADDQ    $32, DI
	JMP     iloop
ilast:
	// the last block, which overlaps the previous one unless the blocks ended exactly at n.
	LEAQ (SI)(CX*1), DX
	CMPQ DI, DX
	JEQ  inone
	MOVQ R11, DI
	VMOVDQU (DI), Y0
	MASK(Y0, Y4, Y5)
	VPTEST  Y5, Y5
	JNE     ifound
inone:
	VZEROUPPER
	MOVQ CX, ret+24(FP)
	RET
ifound:
	// the lanes which are zero are the ones of the bytes which need no escape.
	VPCMPEQB  Y11, Y5, Y5
	VPMOVMSKB Y5, DX
	NOTL      DX
	BSFL      DX, DX
	SUBQ      SI, DI
	ADDQ      DX, DI
	VZEROUPPER
	MOVQ      DI, ret+24(FP)
	RET

// func escapeStringAVX2(dst, src unsafe.Pointer, n int, tables *nibbleTables, seqs *[256]uint64) (consumed, written int)
//
// It appends the n bytes at src to dst, escaped, 32 bytes at a time: a block is stored to dst as it is, and dst
// is advanced to the first byte of it which may need an escape by the tables, whose escape, the sequence of seqs
// of the byte ( see escapeSequences ), is stored after it. It stops at a byte whose sequence is 0, which the
// caller escapes, or when fewer than 32 bytes remain, and returns the numbers of the bytes it read and wrote. dst
// has room for 6*n+32 bytes: a byte is escaped in 6 bytes at most, and a block is stored whole.
TEXT ·escapeStringAVX2(SB), NOSPLIT, $0-56
	MOVQ dst+0(FP), DI
	MOVQ src+8(FP), SI
	MOVQ n+16(FP), CX
	MOVQ tables+24(FP), AX
	MOVQ seqs+32(FP), R9

	VBROADCASTI128 (AX), Y8
	VBROADCASTI128 16(AX), Y9
	VPBROADCASTB   c0f<>(SB), Y10
	VPXOR          Y11, Y11, Y11

	MOVQ SI, R12                 // the start of src
	MOVQ DI, R13                 // the start of dst
	LEAQ (SI)(CX*1), R11         // the end of src
eloop:
	LEAQ    32(SI), DX
	CMPQ    DX, R11
	JGT     edone
	VMOVDQU (SI), Y0
	VMOVDQU Y0, (DI)
	MASK(Y0, Y4, Y5)
	VPTEST  Y5, Y5
	JNE     eescape
	ADDQ    $32, SI
	ADDQ    $32, DI
	JMP     eloop
eescape:
	// the lanes which are zero are the ones of the bytes which need no escape.
	VPCMPEQB  Y11, Y5, Y5
	VPMOVMSKB Y5, DX
	NOTL      DX
	BSFL      DX, DX
	ADDQ      DX, SI
	ADDQ      DX, DI
	MOVBLZX   (SI), AX
	MOVQ      (R9)(AX*8), BX
	TESTQ     BX, BX
	JEQ       edone
	// the sequence, whose length is in the top byte: the bytes after it are written over.
	MOVQ      BX, (DI)
	SHRQ      $56, BX
	ADDQ      BX, DI
	INCQ      SI
	JMP       eloop
edone:
	VZEROUPPER
	SUBQ R12, SI
	MOVQ SI, consumed+40(FP)
	SUBQ R13, DI
	MOVQ DI, written+48(FP)
	RET

DATA c0f<>+0(SB)/1, $0x0f
GLOBL c0f<>(SB), RODATA|NOPTR, $1

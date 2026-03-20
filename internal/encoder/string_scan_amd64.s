#include "textflag.h"

// SSE2 constants for character comparisons

DATA const_ctrl<>+0x00(SB)/8, $0x2020202020202020
DATA const_ctrl<>+0x08(SB)/8, $0x2020202020202020
GLOBL const_ctrl<>(SB), (NOPTR+RODATA), $16

DATA const_quote<>+0x00(SB)/8, $0x2222222222222222
DATA const_quote<>+0x08(SB)/8, $0x2222222222222222
GLOBL const_quote<>(SB), (NOPTR+RODATA), $16

DATA const_bslash<>+0x00(SB)/8, $0x5c5c5c5c5c5c5c5c
DATA const_bslash<>+0x08(SB)/8, $0x5c5c5c5c5c5c5c5c
GLOBL const_bslash<>(SB), (NOPTR+RODATA), $16

DATA const_lt<>+0x00(SB)/8, $0x3c3c3c3c3c3c3c3c
DATA const_lt<>+0x08(SB)/8, $0x3c3c3c3c3c3c3c3c
GLOBL const_lt<>(SB), (NOPTR+RODATA), $16

DATA const_gt<>+0x00(SB)/8, $0x3e3e3e3e3e3e3e3e
DATA const_gt<>+0x08(SB)/8, $0x3e3e3e3e3e3e3e3e
GLOBL const_gt<>(SB), (NOPTR+RODATA), $16

DATA const_amp<>+0x00(SB)/8, $0x2626262626262626
DATA const_amp<>+0x08(SB)/8, $0x2626262626262626
GLOBL const_amp<>(SB), (NOPTR+RODATA), $16

// func scanEscapeBasic(p unsafe.Pointer, n int) int
//
// Scans n bytes starting at p for characters needing JSON string escaping.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), non-ASCII (>= 0x80), '"', '\'
// Uses SSE2 to process 16 bytes per iteration.
TEXT ·scanEscapeBasic(SB), NOSPLIT, $0-24
	MOVQ p+0(FP), SI
	MOVQ n+8(FP), CX
	XORQ AX, AX

	// Load SSE2 constants
	MOVOU const_ctrl<>(SB), X0
	MOVOU const_quote<>(SB), X1
	MOVOU const_bslash<>(SB), X2

	// Calculate loop bound
	MOVQ CX, DX
	SUBQ $15, DX
	JLE  scanBasicTail

scanBasicLoop16:
	CMPQ AX, DX
	JGE  scanBasicTail

	// Load 16 bytes (unaligned)
	MOVOU (SI)(AX*1), X3

	// Check control chars (< 0x20) and non-ASCII (>= 0x80)
	// Signed comparison: 0x20 > byte catches both ranges
	MOVO X0, X4
	PCMPGTB X3, X4

	// Check for '"' (0x22)
	MOVO  X3, X5
	PCMPEQB X1, X5
	POR   X5, X4

	// Check for '\' (0x5C)
	MOVO  X3, X5
	PCMPEQB X2, X5
	POR   X5, X4

	// Extract bitmask
	PMOVMSKB X4, BX
	TESTL BX, BX
	JNZ   scanBasicFound16

	ADDQ $16, AX
	JMP  scanBasicLoop16

scanBasicFound16:
	BSFL BX, BX
	ADDQ BX, AX
	MOVQ AX, ret+16(FP)
	RET

scanBasicTail:
	CMPQ AX, CX
	JGE  scanBasicNotFound

	MOVBLZX (SI)(AX*1), BX
	CMPB BL, $0x20
	JB   scanBasicFoundTail
	TESTB $0x80, BL
	JNZ  scanBasicFoundTail
	CMPB BL, $0x22
	JE   scanBasicFoundTail
	CMPB BL, $0x5C
	JE   scanBasicFoundTail

	INCQ AX
	JMP  scanBasicTail

scanBasicFoundTail:
	MOVQ AX, ret+16(FP)
	RET

scanBasicNotFound:
	MOVQ CX, ret+16(FP)
	RET

// func scanEscapeHTML(p unsafe.Pointer, n int) int
//
// Same as scanEscapeBasic but also detects '<' (0x3C), '>' (0x3E), '&' (0x26).
// Used for HTML-safe JSON string escaping.
TEXT ·scanEscapeHTML(SB), NOSPLIT, $0-24
	MOVQ p+0(FP), SI
	MOVQ n+8(FP), CX
	XORQ AX, AX

	// Load SSE2 constants
	MOVOU const_ctrl<>(SB), X0
	MOVOU const_quote<>(SB), X1
	MOVOU const_bslash<>(SB), X2
	MOVOU const_lt<>(SB), X7
	MOVOU const_gt<>(SB), X8
	MOVOU const_amp<>(SB), X9

	// Calculate loop bound
	MOVQ CX, DX
	SUBQ $15, DX
	JLE  scanHTMLTail

scanHTMLLoop16:
	CMPQ AX, DX
	JGE  scanHTMLTail

	// Load 16 bytes
	MOVOU (SI)(AX*1), X3

	// Check control chars + non-ASCII
	MOVO X0, X4
	PCMPGTB X3, X4

	// Check '"'
	MOVO  X3, X5
	PCMPEQB X1, X5
	POR   X5, X4

	// Check '\'
	MOVO  X3, X5
	PCMPEQB X2, X5
	POR   X5, X4

	// Check '<'
	MOVO  X3, X5
	PCMPEQB X7, X5
	POR   X5, X4

	// Check '>'
	MOVO  X3, X5
	PCMPEQB X8, X5
	POR   X5, X4

	// Check '&'
	MOVO  X3, X5
	PCMPEQB X9, X5
	POR   X5, X4

	// Extract bitmask
	PMOVMSKB X4, BX
	TESTL BX, BX
	JNZ   scanHTMLFound16

	ADDQ $16, AX
	JMP  scanHTMLLoop16

scanHTMLFound16:
	BSFL BX, BX
	ADDQ BX, AX
	MOVQ AX, ret+16(FP)
	RET

scanHTMLTail:
	CMPQ AX, CX
	JGE  scanHTMLNotFound

	MOVBLZX (SI)(AX*1), BX
	CMPB BL, $0x20
	JB   scanHTMLFoundTail
	TESTB $0x80, BL
	JNZ  scanHTMLFoundTail
	CMPB BL, $0x22
	JE   scanHTMLFoundTail
	CMPB BL, $0x5C
	JE   scanHTMLFoundTail
	CMPB BL, $0x3C
	JE   scanHTMLFoundTail
	CMPB BL, $0x3E
	JE   scanHTMLFoundTail
	CMPB BL, $0x26
	JE   scanHTMLFoundTail

	INCQ AX
	JMP  scanHTMLTail

scanHTMLFoundTail:
	MOVQ AX, ret+16(FP)
	RET

scanHTMLNotFound:
	MOVQ CX, ret+16(FP)
	RET

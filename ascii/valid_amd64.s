// Code generated by command: go run valid_asm.go -out valid_amd64.s -stubs valid_amd64.go. DO NOT EDIT.

#include "textflag.h"

// func validPrint(s string) int
// Requires: SSE, SSE2, SSE4.1
TEXT ·validPrint(SB), NOSPLIT, $0-24
	MOVQ s_base+0(FP), AX
	MOVQ s_len+8(FP), CX
	MOVQ $0x0000000000000000, DX

	// Only initialize the 64 bits registers if there are more than 8 bytes.
	CMPQ CX, $0x08
	JL   loop1

	// Only initialize the 128 bits registers if there are more than 16 bytes.
	CMPQ   CX, $0x10
	JL     init8
	MOVQ   $0xffffffffffffffff, BX
	PINSRQ $0x00, BX, X0
	PINSRQ $0x01, BX, X0
	MOVQ   $0x2020202020202020, BX
	MOVQ   $0x2020202020202020, BP
	PINSRQ $0x00, BX, X1
	PINSRQ $0x01, BP, X1
	MOVQ   $0x0101010101010101, BX
	MOVQ   $0x0101010101010101, BP
	PINSRQ $0x00, BX, X3
	PINSRQ $0x01, BP, X3
	MOVQ   $0x8080808080808080, BX
	MOVQ   $0x8080808080808080, BP
	PINSRQ $0x00, BX, X2
	PINSRQ $0x01, BP, X2
	PINSRQ $0x00, BX, X4
	PINSRQ $0x01, BP, X4

loop32:
	// Loop until less than 32 bytes remain.
	CMPQ   CX, $0x20
	JL     loop16
	MOVUPS (AX), X5
	MOVUPS X5, X6
	MOVUPS X5, X7
	SUBSS  X1, X7
	XORPS  X0, X6
	ANDPS  X6, X7
	ANDPS  X2, X7
	MOVQ   X7, BX
	CMPQ   BX, $0x00
	JNE    done
	MOVUPS X5, X7
	ADDSS  X3, X7
	ORPS   X5, X7
	ANDPS  X4, X7
	MOVQ   X7, BX
	CMPQ   BX, $0x00
	JNE    done
	MOVUPS 16(AX), X5
	MOVUPS X5, X6
	MOVUPS X5, X7
	SUBSS  X1, X7
	XORPS  X0, X6
	ANDPS  X6, X7
	ANDPS  X2, X7
	MOVQ   X7, BX
	CMPQ   BX, $0x00
	JNE    done
	MOVUPS X5, X7
	ADDSS  X3, X7
	ORPS   X5, X7
	ANDPS  X4, X7
	MOVQ   X7, BX
	CMPQ   BX, $0x00
	JNE    done
	SUBQ   $0x20, CX
	ADDQ   $0x20, AX
	JMP    loop32

loop16:
	// Consume the next 16 bytes of input.
	CMPQ   CX, $0x10
	JL     init8
	MOVUPS (AX), X5
	MOVUPS X5, X6
	MOVUPS X5, X7
	SUBSS  X1, X7
	XORPS  X0, X6
	ANDPS  X6, X7
	ANDPS  X2, X7
	MOVQ   X7, BX
	CMPQ   BX, $0x00
	JNE    done
	MOVUPS X5, X7
	ADDSS  X3, X7
	ORPS   X5, X7
	ANDPS  X4, X7
	MOVQ   X7, BX
	CMPQ   BX, $0x00
	JNE    done
	SUBQ   $0x10, CX
	ADDQ   $0x10, AX

init8:
	MOVQ $0xffffffffffffffff, DI
	MOVQ $0x2020202020202020, R8
	MOVQ $0x8080808080808080, R9
	MOVQ $0x0101010101010101, R10
	MOVQ $0x8080808080808080, R11

	// Consume the next 8 bytes of input.
	CMPQ CX, $0x08
	JL   loop1
	MOVQ (AX), BX
	MOVQ BX, BP
	MOVQ BX, SI
	SUBQ R8, SI
	XORQ DI, BP
	ANDQ BP, SI
	ANDQ R9, SI
	CMPQ SI, $0x00
	JNE  done
	MOVQ BX, SI
	ADDQ R10, SI
	ORQ  BX, SI
	ANDQ R11, SI
	CMPQ SI, $0x00
	JNE  done
	SUBQ $0x08, CX
	ADDQ $0x08, AX

loop1:
	// Loop until zero bytes remain.
	CMPQ    CX, $0x00
	JLE     valid
	MOVBQZX (AX), BX
	CMPQ    BX, $0x20
	JL      done
	CMPQ    BX, $0x7e
	JG      done
	DECQ    CX
	INCQ    AX
	JMP     loop1

valid:
	MOVQ $0x0000000000000001, DX

done:
	MOVQ DX, ret+16(FP)
	RET

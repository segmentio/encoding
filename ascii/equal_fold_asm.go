// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("equalFoldAVX2", NOSPLIT, "func(a *byte, b *byte, n uintptr) int")
	Doc("Case-insensitive comparison of two ASCII strings (equality).")

	p := Load(Param("a"), GP64())
	q := Load(Param("b"), GP64())
	n := Load(Param("n"), GP64())
	SHRQ(Imm(4), n) // n /= 16

	eq := GP64()
	MOVQ(U64(0), eq)

	mask64 := GP64()
	MOVQ(U64(0xDFDFDFDFDFDFDFDF), mask64)

	mask128 := XMM()
	mask256 := YMM()
	PINSRQ(Imm(0), mask64, mask128)
	PINSRQ(Imm(1), mask64, mask128)
	VPBROADCASTQ(mask128, mask256)

	cmpk := GP32()
	xmm0 := XMM()
	xmm1 := XMM()
	ymm0 := YMM()
	ymm1 := YMM()
	ymm2 := YMM()
	ymm3 := YMM()

	Label("loop64")
	CMPQ(n, Imm(4))
	JL(LabelRef("loop32"))

	VPAND(Mem{Base: p}, mask256, ymm0)
	VPAND(Mem{Base: q}, mask256, ymm1)
	VPCMPEQB(ymm1, ymm0, ymm0)

	VPAND((Mem{Base: p}).Offset(32), mask256, ymm2)
	VPAND((Mem{Base: q}).Offset(32), mask256, ymm3)
	VPCMPEQB(ymm3, ymm2, ymm2)

	VPAND(ymm2, ymm0, ymm0)
	VPMOVMSKB(ymm0, cmpk)
	CMPL(cmpk, U32(0xFFFFFFFF))
	JNE(LabelRef("done"))

	ADDQ(Imm(64), p)
	ADDQ(Imm(64), q)
	SUBQ(Imm(4), n)
	JMP(LabelRef("loop64"))

	Label("loop32")
	CMPQ(n, Imm(2))
	JL(LabelRef("loop16"))

	VPAND(Mem{Base: p}, mask256, ymm0)
	VPAND(Mem{Base: q}, mask256, ymm1)
	VPCMPEQB(ymm1, ymm0, ymm0)
	VPMOVMSKB(ymm0, cmpk)
	CMPL(cmpk, U32(0xFFFFFFFF))
	JNE(LabelRef("done"))

	ADDQ(Imm(32), p)
	ADDQ(Imm(32), q)
	SUBQ(Imm(2), n)

	Label("loop16")
	CMPQ(n, Imm(0))
	JE(LabelRef("equal"))

	VPAND(Mem{Base: p}, mask128, xmm0)
	VPAND(Mem{Base: q}, mask128, xmm1)
	VPCMPEQB(xmm1, xmm0, xmm0)
	VPMOVMSKB(xmm0, cmpk)
	CMPL(cmpk, U32(0xFFFF))
	JNE(LabelRef("done"))

	Label("equal")
	MOVQ(U64(1), eq)

	Label("done")
	Store(eq, ReturnIndex(0))
	RET()
	Generate()
}

// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("validPrintAVX2", NOSPLIT, "func(p *byte, n uintptr) int")
	Doc("Validates that the string only contains printable ASCII characters.")

	p := Load(Param("p"), GP64())
	n := Load(Param("n"), GP64())
	SHRQ(Imm(4), n) // n /= 16

	r := GP64()
	MOVQ(U64(0), r)

	Comment("Initialize 128 bits registers.")
	min := GP64()
	max := GP64()
	MOVQ(U64(0x1F1F1F1F1F1F1F1F), min)
	MOVQ(U64(0x7E7E7E7E7E7E7E7E), max)

	minXMM := XMM()
	PINSRQ(Imm(0), min, minXMM)
	PINSRQ(Imm(1), min, minXMM)

	maxXMM := XMM()
	PINSRQ(Imm(0), max, maxXMM)
	PINSRQ(Imm(1), max, maxXMM)

	xmm0 := XMM()
	xmm1 := XMM()
	msk0 := GP32()

	CMPQ(n, Imm(2)) // skip YMM register initialization if there are less than 32 bytes
	JL(LabelRef("loop16"))

	Comment("Initialize 256 bits registers.")
	minYMM := YMM()
	VPBROADCASTQ(minXMM, minYMM)

	maxYMM := YMM()
	VPBROADCASTQ(maxXMM, maxYMM)

	ymm0 := YMM()
	ymm1 := YMM()
	ymm2 := YMM()
	ymm3 := YMM()
	ymm4 := YMM()
	ymm5 := YMM()
	ymm6 := YMM()
	ymm7 := YMM()
	ymm8 := YMM()

	Label("loop64")
	Comment("Unroll two iterations of the loop operating on 32 bytes chunks.")
	CMPQ(n, Imm(4))
	JL(LabelRef("loop32"))

	VMOVUPS(Mem{Base: p}, ymm0)
	VMOVUPS((Mem{Base: p}).Offset(32), ymm1)
	VPCMPGTB(minYMM, ymm0, ymm2) // A = bytes that are greater than the min-1 (i.e. valid at lower end)
	VPCMPGTB(maxYMM, ymm0, ymm3) // B = bytes that are greater than the max (i.e. invalid at upper end)
	VPANDN(ymm2, ymm3, ymm4)     // A & ~B mask should be full unless there's an invalid byte
	VPCMPGTB(minYMM, ymm1, ymm5) // compute the same for the next 32 bytes
	VPCMPGTB(maxYMM, ymm1, ymm6)
	VPANDN(ymm5, ymm6, ymm7)
	VPAND(ymm4, ymm7, ymm8) // combine masks
	VPMOVMSKB(ymm8, msk0)
	XORL(U32(0xFFFFFFFF), msk0) // check for a zero somewhere
	JNE(LabelRef("done"))

	SUBQ(Imm(4), n)
	ADDQ(Imm(64), p)
	CMPQ(n, Imm(4)) // more than 64 bytes?
	JGE(LabelRef("loop64"))

	Label("loop32")
	Comment("Consume the next 32 bytes of input.")
	CMPQ(n, Imm(2))
	JL(LabelRef("loop16"))

	VMOVUPS(Mem{Base: p}, ymm0)
	VPCMPGTB(minYMM, ymm0, ymm2)
	VPCMPGTB(maxYMM, ymm0, ymm3)
	VPANDN(ymm2, ymm3, ymm4)
	VPMOVMSKB(ymm4, msk0)
	XORL(U32(0xFFFFFFFF), msk0)
	JNE(LabelRef("done"))

	SUBQ(Imm(2), n)
	ADDQ(Imm(32), p)

	Label("loop16")
	Comment("Consume the next 16 bytes of input.")
	CMPQ(n, Imm(0))
	JE(LabelRef("valid"))

	MOVUPS(Mem{Base: p}, xmm0)
	MOVUPS(xmm0, xmm1)
	PCMPGTB(minXMM, xmm0)
	PCMPGTB(maxXMM, xmm1)
	PANDN(xmm0, xmm1)
	PMOVMSKB(xmm0, msk0)
	XORL(U32(0xFFFF), msk0)
	JNE(LabelRef("done"))

	Label("valid")
	MOVQ(U32(1), r)

	Label("done")
	Store(r, ReturnIndex(0))
	RET()
	Generate()
}

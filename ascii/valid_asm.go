// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("validPrint16", NOSPLIT, "func(p *byte, n uintptr) int")
	Doc("Validates that the string only contains printable ASCII characters.")

	p := Load(Param("p"), GP64())
	n := Load(Param("n"), GP64())
	r := GP64()
	MOVQ(U64(0), r)

	Comment("Initialize 128 bits registers.")
	min := GP64()
	max := GP64()
	MOVQ(U64(0x1919191919191919), min)
	MOVQ(U64(0x7F7F7F7F7F7F7F7F), max)

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

	Label("loop64")
	Comment("Unroll two iterations of the loop operating on 32 bytes chunks.")
	CMPQ(n, Imm(4))
	JL(LabelRef("loop32"))

	VMOVUPS(Mem{Base: p}, ymm0)

	VMOVUPS(ymm0, ymm1)
	VPMINUB(minYMM, ymm1, ymm2)
	VPCMPEQB(minYMM, ymm2, ymm1)
	VPMOVMSKB(ymm1, msk0)
	XORL(U32(0xFFFFFFFF), msk0)
	JNE(LabelRef("done"))

	VMOVUPS(ymm0, ymm1)
	VPMAXUB(maxYMM, ymm1, ymm2)
	VPCMPEQB(maxYMM, ymm2, ymm1)
	VPMOVMSKB(ymm1, msk0)
	XORL(U32(0xFFFFFFFF), msk0)
	JNE(LabelRef("done"))

	VMOVUPS((Mem{Base: p}).Offset(32), ymm0)

	VMOVUPS(ymm0, ymm1)
	VPMINUB(minYMM, ymm1, ymm2)
	VPCMPEQB(minYMM, ymm2, ymm1)
	VPMOVMSKB(ymm1, msk0)
	XORL(U32(0xFFFFFFFF), msk0)
	JNE(LabelRef("done"))

	VMOVUPS(ymm0, ymm1)
	VPMAXUB(maxYMM, ymm1, ymm2)
	VPCMPEQB(maxYMM, ymm2, ymm1)
	VPMOVMSKB(ymm1, msk0)
	XORL(U32(0xFFFFFFFF), msk0)
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

	VMOVUPS(ymm0, ymm1)
	VPMINUB(minYMM, ymm1, ymm2)
	VPCMPEQB(minYMM, ymm2, ymm1)
	VPMOVMSKB(ymm1, msk0)
	XORL(U32(0xFFFFFFFF), msk0)
	JNE(LabelRef("done"))

	VMOVUPS(ymm0, ymm1)
	VPMAXUB(maxYMM, ymm1, ymm2)
	VPCMPEQB(maxYMM, ymm2, ymm1)
	VPMOVMSKB(ymm1, msk0)
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
	PMINUB(minXMM, xmm1)    // extract the max of 0x20(x16) and xmm1 in each byte
	PCMPEQB(minXMM, xmm1)   // check bytes of xmm1 for equality with 0x20
	PMOVMSKB(xmm1, msk0)    // move the most significant bits of xmm1 to msk0
	XORL(U32(0xFFFF), msk0) // invert all bits of msk0
	JNE(LabelRef("done"))   // if non-zero, some bytes were lower than 0x20

	MOVUPS(xmm0, xmm1)
	PMAXUB(maxXMM, xmm1)    // extract the min of 0x7E(x16) and xmm1 in each byte
	PCMPEQB(maxXMM, xmm1)   // check bytes of xmm1 for equality with 0x7F
	PMOVMSKB(xmm1, msk0)    // move the most significant bits of xmm1 to msk0
	XORL(U32(0xFFFF), msk0) // invert all bits of msk0
	JNE(LabelRef("done"))   // if non-zero, some bytes were greater than 0x7E

	Label("valid")
	MOVQ(U32(1), r)

	Label("done")
	Store(r, ReturnIndex(0))
	RET()
	Generate()
}

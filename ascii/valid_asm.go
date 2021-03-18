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

	Label("loop32")
	Comment("Loop until less than 32 bytes remain.")

	VMOVUPS(Mem{Base: p}, ymm0)

	VMOVUPS(ymm0, ymm1)
	VPMINUB(minYMM, ymm1, ymm2)  // extract the max of 0x20(x32) and ymm1 in each byte
	VPCMPEQB(minYMM, ymm2, ymm1) // check bytes of ymm1 for equality with 0x20
	VPMOVMSKB(ymm1, msk0)        // move the most significant bits of ymm1 to msk0
	XORL(U32(0xFFFFFFFF), msk0)  // invert all bits of msk0
	JNE(LabelRef("done"))        // if non-zero, some bytes were lower than 0x20

	VMOVUPS(ymm0, ymm1)
	VPMAXUB(maxYMM, ymm1, ymm2)  // extract the min of 0x7E(x32) and ymm1 in each byte
	VPCMPEQB(maxYMM, ymm2, ymm1) // check bytes of ymm1 for equality with 0x7F
	VPMOVMSKB(ymm1, msk0)        // move the most significant bits of ymm1 to msk0
	XORL(U32(0xFFFFFFFF), msk0)  // invert all bits of msk0
	JNE(LabelRef("done"))        // if non-zero, some bytes were greater than 0x7E

	SUBQ(Imm(2), n)
	ADDQ(Imm(32), p)
	CMPQ(n, Imm(2)) // more than 32 bytes?
	JGE(LabelRef("loop32"))

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

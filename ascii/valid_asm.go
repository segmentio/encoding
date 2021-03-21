// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("valid16", NOSPLIT, "func(p *byte, n uintptr) int")
	Doc("Validates that the string only contains ASCII characters.")

	p := Load(Param("p"), GP64())
	n := Load(Param("n"), GP64())
	r := GP64()
	MOVQ(U64(0), r)

	msk := GP32()
	xmm0 := XMM()
	ymm0 := YMM()

	Label("loop64")
	CMPQ(n, Imm(4))
	JL(LabelRef("loop32"))

	// TODO: it would be nice to be able to use VPMOVB2M here, but it appears
	// that the current version of avo hasn't yet released support for the
	// AVX.512 instruction set.
	VMOVUPS(Mem{Base: p}, ymm0)
	VPOR((Mem{Base: p}).Offset(32), ymm0, ymm0)

	VPMOVMSKB(ymm0, msk)
	CMPL(msk, Imm(0))
	JNE(LabelRef("done"))

	SUBQ(Imm(4), n)
	ADDQ(Imm(64), p)
	CMPQ(n, Imm(4))
	JGE(LabelRef("loop64"))

	Label("loop32")
	Comment("Consume the next 32 bytes of input.")
	CMPQ(n, Imm(2))
	JL(LabelRef("loop16"))

	VMOVUPS(Mem{Base: p}, ymm0)
	VPMOVMSKB(ymm0, msk)
	CMPL(msk, Imm(0))
	JNE(LabelRef("done"))

	SUBQ(Imm(2), n)
	ADDQ(Imm(32), p)

	Label("loop16")
	Comment("Consume the next 16 bytes of input.")
	CMPQ(n, Imm(0))
	JE(LabelRef("valid"))

	MOVUPS(Mem{Base: p}, xmm0)
	PMOVMSKB(xmm0, msk)
	CMPL(msk, Imm(0))
	JNE(LabelRef("done"))

	Label("valid")
	MOVQ(U32(1), r)

	Label("done")
	Store(r, ReturnIndex(0))
	RET()
	Generate()
}

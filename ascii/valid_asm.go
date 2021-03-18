// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

// https://graphics.stanford.edu/~seander/bithacks.html#HasLessInWord
//
// The masks have been simplified for the special case implemented by the
// validatePrint function, which tests if the words contain a byte less than
// 0x20 or more than 0x7E (bounds of printable ASCII characters).
const (
	mask0 = 0x0101010101010101 // (2^64 - 1) / 255
	mask1 = 0x8080808080808080 // mask0 * 128
	space = 0x2020202020202020 // mask0 * 0x20
)

func main() {
	TEXT("validPrint16", NOSPLIT, "func(p *byte, n uintptr) int")
	Doc("Validates that the string only contains printable ASCII characters.")

	p := Load(Param("p"), GP64())
	n := Load(Param("n"), GP64())
	r := GP64()
	MOVQ(U64(0), r)

	// =========================================================================
	// Loop optimized for strings with 16 bytes or more.
	Label("init16")

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

	// The first section unrolls two loop iterations, which amortizes the cost
	// of memory loads and loop management (pointer increment, counter decrement).
	/*
		Label("loop32")
		Comment("Loop until less than 32 bytes remain.")
		CMPQ(n, Imm(2)) // less than 2 x 16 bytes?
		JL(LabelRef("loop16"))

		SUBQ(Imm(2), n)
		ADDQ(Imm(32), p)
		JMP(LabelRef("loop32"))
	*/

	// The second part of the 16 bytes section is entered when there are less
	// than 32 bytes remaining.
	Label("loop16")
	//Comment("Consume the next 16 bytes of input.")
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

	DECQ(n)
	ADDQ(Imm(16), p)
	JMP(LabelRef("loop16"))
	// =========================================================================

	Label("valid")
	MOVQ(U32(1), r)

	Label("done")
	Store(r, ReturnIndex(0))
	RET()
	Generate()
}

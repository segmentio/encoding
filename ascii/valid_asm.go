// +build ignore

package main

import (
	"math"

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

	n := Load(Param("n"), GP64())
	CMPQ(n, Imm(0))
	JE(LabelRef("valid"))

	p := Load(Param("p"), GP64())
	r := GP64()
	x := GP64()
	y := GP64()
	MOVQ(U64(0), r)

	// =========================================================================
	// Loop optimized for strings with 16 bytes or more.
	Label("init16")
	maxUint128 := XMM()
	hasLessThan0x20Lxmm := XMM()
	hasLessThan0x20Rxmm := XMM()
	hasMoreThan0x7eLxmm := XMM()
	hasMoreThan0x7eRxmm := XMM()

	xmm0 := XMM()
	xmm1 := XMM()
	xmm2 := XMM()

	MOVQ(U64(math.MaxUint64), x)
	PINSRQ(Imm(0), x, maxUint128)
	PINSRQ(Imm(1), x, maxUint128)

	MOVQ(U64(space), x)
	MOVQ(U64(space), y)
	PINSRQ(Imm(0), x, hasLessThan0x20Lxmm)
	PINSRQ(Imm(1), y, hasLessThan0x20Lxmm)

	MOVQ(U64(mask0), x)
	MOVQ(U64(mask0), y)
	PINSRQ(Imm(0), x, hasMoreThan0x7eLxmm)
	PINSRQ(Imm(1), y, hasMoreThan0x7eLxmm)

	MOVQ(U64(mask1), x)
	MOVQ(U64(mask1), y)
	PINSRQ(Imm(0), x, hasLessThan0x20Rxmm)
	PINSRQ(Imm(1), y, hasLessThan0x20Rxmm)
	PINSRQ(Imm(0), x, hasMoreThan0x7eRxmm)
	PINSRQ(Imm(1), y, hasMoreThan0x7eRxmm)

	// The first section unrolls two loop iterations, which amortizes the cost
	// of memory loads and loop management (pointer increment, counter decrement).
	Label("loop32")
	Comment("Loop until less than 32 bytes remain.")
	CMPQ(n, Imm(2)) // less than 2 x 16 bytes?
	JL(LabelRef("loop16"))

	MOVUPS(Mem{Base: p}, xmm0)

	// hasLess
	MOVUPS(xmm0, xmm1)
	MOVUPS(xmm0, xmm2)
	SUBSS(hasLessThan0x20Lxmm, xmm2)
	XORPS(maxUint128, xmm1)
	ANDPS(xmm1, xmm2)
	ANDPS(hasLessThan0x20Rxmm, xmm2)
	MOVQ(xmm2, x)
	CMPQ(x, Imm(0))
	JNE(LabelRef("done"))

	// hasMore
	MOVUPS(xmm0, xmm2)
	ADDSS(hasMoreThan0x7eLxmm, xmm2)
	ORPS(xmm0, xmm2)
	ANDPS(hasMoreThan0x7eRxmm, xmm2)
	MOVQ(xmm2, x)
	CMPQ(x, Imm(0))
	JNE(LabelRef("done"))

	MOVUPS((Mem{Base: p}).Offset(16), xmm0)

	// hasLess
	MOVUPS(xmm0, xmm1)
	MOVUPS(xmm0, xmm2)
	SUBSS(hasLessThan0x20Lxmm, xmm2)
	XORPS(maxUint128, xmm1)
	ANDPS(xmm1, xmm2)
	ANDPS(hasLessThan0x20Rxmm, xmm2)
	MOVQ(xmm2, x)
	CMPQ(x, Imm(0))
	JNE(LabelRef("done"))

	// hasMore
	MOVUPS(xmm0, xmm2)
	ADDSS(hasMoreThan0x7eLxmm, xmm2)
	ORPS(xmm0, xmm2)
	ANDPS(hasMoreThan0x7eRxmm, xmm2)
	MOVQ(xmm2, x)
	CMPQ(x, Imm(0))
	JNE(LabelRef("done"))

	SUBQ(Imm(2), n)
	ADDQ(Imm(32), p)
	JMP(LabelRef("loop32"))

	// The second part of the 16 bytes section is entered when there are less
	// than 32 bytes remaining.
	Label("loop16")
	Comment("Consume the next 16 bytes of input.")
	CMPQ(n, Imm(0))
	JE(LabelRef("valid"))

	MOVUPS(Mem{Base: p}, xmm0)

	// hasLess
	MOVUPS(xmm0, xmm1)
	MOVUPS(xmm0, xmm2)
	SUBSS(hasLessThan0x20Lxmm, xmm2)
	XORPS(maxUint128, xmm1)
	ANDPS(xmm1, xmm2)
	ANDPS(hasLessThan0x20Rxmm, xmm2)
	MOVQ(xmm2, x)
	CMPQ(x, Imm(0))
	JNE(LabelRef("done"))

	// hasMore
	MOVUPS(xmm0, xmm2)
	ADDSS(hasMoreThan0x7eLxmm, xmm2)
	ORPS(xmm0, xmm2)
	ANDPS(hasMoreThan0x7eRxmm, xmm2)
	MOVQ(xmm2, x)
	CMPQ(x, Imm(0))
	JNE(LabelRef("done"))
	// =========================================================================

	Label("valid")
	MOVQ(U32(1), r)

	Label("done")
	Store(r, ReturnIndex(0))
	RET()
	Generate()
}

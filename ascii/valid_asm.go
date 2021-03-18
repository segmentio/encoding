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
	TEXT("validPrint", NOSPLIT, "func(s string) int")
	Doc("Validates that the string only contains printable ASCII characters.")

	p := Load(Param("s").Base(), GP64())
	n := Load(Param("s").Len(), GP64())
	r := GP64()
	x := GP64()
	y := GP64()
	z := GP64()
	MOVQ(U64(0), r)

	Comment("Only initialize the 64 bits registers if there are more than 8 bytes.")
	CMPQ(n, Imm(8))
	JL(LabelRef("loop1"))

	Comment("Only initialize the 128 bits registers if there are more than 16 bytes.")
	CMPQ(n, Imm(16))
	JL(LabelRef("init8"))

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
	CMPQ(n, Imm(32))
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

	SUBQ(Imm(32), n)
	ADDQ(Imm(32), p)
	JMP(LabelRef("loop32"))

	// The second part of the 16 bytes section is entered when there are less
	// than 32 bytes remaining.
	Label("loop16")
	Comment("Consume the next 16 bytes of input.")
	CMPQ(n, Imm(16))
	JL(LabelRef("init8"))

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

	SUBQ(Imm(16), n)
	ADDQ(Imm(16), p)
	// =========================================================================

	// =========================================================================
	// Section for the case where there is at least 8 bytes to read. The code
	// either jumps here from the end of "loop16", or starts here for strings
	// that are shorter than 16 bytes.
	Label("init8")
	maxUint64 := GP64()
	hasLessThan0x20L := GP64()
	hasLessThan0x20R := GP64()
	hasMoreThan0x7eL := GP64()
	hasMoreThan0x7eR := GP64()

	MOVQ(U64(math.MaxUint64), maxUint64)
	MOVQ(U64(mask0*uint64(0x20)), hasLessThan0x20L)
	MOVQ(U64(mask1), hasLessThan0x20R)
	MOVQ(U64(mask0*(uint64(127)-0x7e)), hasMoreThan0x7eL)
	MOVQ(U64(mask1), hasMoreThan0x7eR)

	Label("loop8")
	Comment("Consume the next 8 bytes of input.")
	CMPQ(n, Imm(8))
	JL(LabelRef("loop1"))

	MOVQ(Mem{Base: p}, x)

	// func hasLess64(x, n uint64) bool {
	// 	return ((x - (constL64 * n)) & ^x & constR64) != 0
	// }
	//
	// With n=0x20 to test for characters before the space ASCII code.
	MOVQ(x, y)
	MOVQ(x, z)
	SUBQ(hasLessThan0x20L, z)
	XORQ(maxUint64, y)
	ANDQ(y, z)
	ANDQ(hasLessThan0x20R, z)
	CMPQ(z, Imm(0))
	JNE(LabelRef("done"))

	// func hasMore64(x, n uint64) bool {
	// 	return (((x + (hasMoreConstL64 * (127 - n))) | x) & hasMoreConstR64) != 0
	// }
	//
	// With n=0x7e to test for characters after the `~` ASCII code.
	MOVQ(x, z)
	ADDQ(hasMoreThan0x7eL, z)
	ORQ(x, z)
	ANDQ(hasMoreThan0x7eR, z)
	CMPQ(z, Imm(0))
	JNE(LabelRef("done"))

	SUBQ(Imm(8), n)
	ADDQ(Imm(8), p)
	// =========================================================================

	// =========================================================================
	// Last step, iterate over the remaining bytes.
	Label("loop1")
	Comment("Loop until zero bytes remain.")
	CMPQ(n, Imm(0))
	JLE(LabelRef("valid"))

	Label("enterLoop1")
	MOVBQZX(Mem{Base: p}, x)
	CMPQ(x, Imm(0x20))
	JL(LabelRef("done"))

	CMPQ(x, Imm(0x7e))
	JG(LabelRef("done"))

	DECQ(n)
	INCQ(p)
	JMP(LabelRef("loop1"))
	// =========================================================================

	Label("valid")
	MOVQ(U64(1), r)

	Label("done")
	Store(r, ReturnIndex(0))
	RET()
	Generate()
}

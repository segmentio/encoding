// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("equalFoldAVX2", NOSPLIT, "func(a *byte, b *byte, n uintptr) int")
	Doc("Case-insensitive comparison of two ASCII strings (equality).")

	i := GP64()
	p := Mem{Base: Load(Param("a"), GP64()), Index: i, Scale: 1}
	q := Mem{Base: Load(Param("b"), GP64()), Index: i, Scale: 1}
	n := Load(Param("n"), GP64())
	XORQ(i, i)
	SHRQ(Imm(4), n) // n /= 16

	eq := GP64()
	XORQ(eq, eq)

	cmpk := GP32()
	mask256 := [4]Register{}
	mask128 := [4]Register{}

	for i, b := range [4]byte{0x20, 0x1F, 0x9A, 0x01} {
		y := YMM()
		g := GP32()

		MOVB(U8(b), g.As8())
		PINSRB(U8(0), g, y.AsX())
		VPBROADCASTB(y.AsX(), y)

		mask256[i] = y
		mask128[i] = y.AsX()
	}

	Label("loop64")
	CMPQ(n, Imm(4))
	JB(LabelRef("cmp32"))

	VMOVDQU(p, Y0)
	VMOVDQU(p.Offset(32), Y3)
	VMOVDQU(q, Y1)
	VMOVDQU(q.Offset(32), Y4)

	gen(Y0, Y1, Y2, mask256)
	gen(Y3, Y4, Y5, mask256)
	VPAND(Y3, Y0, Y0) // merge results together

	ADDQ(Imm(64), i)
	SUBQ(Imm(4), n)

	VPMOVMSKB(Y0, cmpk)
	CMPL(cmpk, U32(0xFFFFFFFF))
	JNE(LabelRef("done"))

	JMP(LabelRef("loop64"))

	Label("cmp32")
	CMPQ(n, Imm(2))
	JB(LabelRef("cmp16"))

	VMOVDQU(p, Y0)
	VMOVDQU(q, Y1)

	gen(Y0, Y1, Y2, mask256)

	ADDQ(Imm(32), i)
	SUBQ(Imm(2), n)

	VPMOVMSKB(Y0, cmpk)
	CMPL(cmpk, U32(0xFFFFFFFF))
	JNE(LabelRef("done"))

	Label("cmp16")
	CMPQ(n, Imm(1))
	JB(LabelRef("equal"))

	VMOVDQU(p, X0)
	VMOVDQU(q, X1)

	gen(X0, X1, X2, mask128)

	VPMOVMSKB(X0, cmpk)
	CMPL(cmpk, U32(0x0000FFFF))
	JNE(LabelRef("done"))

	Label("equal")
	MOVQ(U64(1), eq)

	Label("done")
	Store(eq, ReturnIndex(0))
	RET()
	Generate()
}

func gen(v0, v1, v2 Register, mask [4]Register) {
	VXORPD(v0, v1, v1)        // calculate difference between v0 and v1
	VPCMPEQB(mask[0], v1, v2) // check if above difference is the 6th bit
	VORPD(mask[0], v0, v0)    // set the 6th bit for v0
	VPADDB(mask[1], v0, v0)   // add 0x1f to each byte to set top bit for letters
	VPCMPGTB(v0, mask[2], v0) // compare if not letter: v - 'a' < 'z' - 'a' + 1
	VPAND(v2, v0, v0)         // combine 6th-bit difference with letter range
	VPAND(mask[3], v0, v0)    // merge test mask
	VPSLLW(Imm(5), v0, v0)    // shift into case bit position
	VPCMPEQB(v1, v0, v0)      // compare original difference with case-only difference
}

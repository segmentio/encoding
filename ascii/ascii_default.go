// +build !amd64

package ascii

// Placeholders for AVX2 function names so compilation works on non-amd64
// platforms.

func equalFoldAVX2(*byte, *byte, uintptr) int {
	return 0
}

func validAVX2(*byte, uintptr) int {
	return 0
}

func validPrintAVX2(*byte, uintptr) int {
	return 0
}

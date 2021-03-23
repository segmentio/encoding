package ascii

import (
	"unsafe"

	. "github.com/klauspost/cpuid/v2"
)

var asm struct {
	equalFoldAVX2  func(*byte, *byte, uintptr) int
	validAVX2      func(*byte, uintptr) int
	validPrintAVX2 func(*byte, uintptr) int
}

func init() {
	if CPU.Supports(AVX, AVX2) {
		asm.equalFoldAVX2 = equalFoldAVX2
		asm.validAVX2 = validAVX2
		asm.validPrintAVX2 = validPrintAVX2
	}
}

// https://graphics.stanford.edu/~seander/bithacks.html#HasLessInWord
const (
	hasLessConstL64 = (^uint64(0)) / 255
	hasLessConstR64 = hasLessConstL64 * 128

	hasLessConstL32 = (^uint32(0)) / 255
	hasLessConstR32 = hasLessConstL32 * 128

	hasMoreConstL64 = (^uint64(0)) / 255
	hasMoreConstR64 = hasMoreConstL64 * 128

	hasMoreConstL32 = (^uint32(0)) / 255
	hasMoreConstR32 = hasMoreConstL32 * 128
)

//go:nosplit
func hasLess64(x, n uint64) bool {
	return ((x - (hasLessConstL64 * n)) & ^x & hasLessConstR64) != 0
}

//go:nosplit
func hasLess32(x, n uint32) bool {
	return ((x - (hasLessConstL32 * n)) & ^x & hasLessConstR32) != 0
}

//go:nosplit
func hasMore64(x, n uint64) bool {
	return (((x + (hasMoreConstL64 * (127 - n))) | x) & hasMoreConstR64) != 0
}

//go:nosplit
func hasMore32(x, n uint32) bool {
	return (((x + (hasMoreConstL32 * (127 - n))) | x) & hasMoreConstR32) != 0
}

func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

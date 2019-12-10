package ascii

import (
	"unsafe"
)

// Valid returns true if b contains only ASCII characters.
func Valid(b []byte) bool {
	return valid(unsafe.Pointer(&b), uintptr(len(b)))
}

// ValidString returns true if s contains only ASCII characters.
func ValidString(s string) bool {
	return valid(unsafe.Pointer(&s), uintptr(len(s)))
}

// ValidBytes returns true if b is an ASCII character.
func ValidByte(b byte) bool {
	return b <= 0x7f
}

// ValidBytes returns true if b is an ASCII character.
func ValidRune(r rune) bool {
	return r <= 0x7f
}

//go:nosplit
func valid(s unsafe.Pointer, n uintptr) bool {
	if n == 0 {
		return true
	}

	i := uintptr(0)
	p := *(*unsafe.Pointer)(s)

	for (n - i) >= 8 {
		if ((*(*uint64)(unsafe.Pointer(uintptr(p) + i))) & 0x8080808080808080) != 0 {
			return false
		}
		i += 8
	}

	if (n - i) >= 4 {
		if ((*(*uint32)(unsafe.Pointer(uintptr(p) + i))) & 0x80808080) != 0 {
			return false
		}
		i += 4
	}

	if (n - i) >= 2 {
		if ((*(*uint16)(unsafe.Pointer(uintptr(p) + i))) & 0x8080) != 0 {
			return false
		}
		i += 2
	}

	if i < n {
		if ((*(*uint8)(unsafe.Pointer(uintptr(p) + i))) & 0x80) != 0 {
			return false
		}
	}

	return true
}

// Valid returns true if b contains only printable ASCII characters.
func ValidPrint(b []byte) bool {
	return validPrint(unsafe.Pointer(&b), uintptr(len(b)))
}

// ValidString returns true if s contains only printable ASCII characters.
func ValidPrintString(s string) bool {
	return validPrint(unsafe.Pointer(&s), uintptr(len(s)))
}

// ValidBytes returns true if b is an ASCII character.
func ValidPrintByte(b byte) bool {
	return 0x20 <= b && b <= 0x7f
}

// ValidBytes returns true if b is an ASCII character.
func ValidPrintRune(r rune) bool {
	return 0x20 <= r && r <= 0x7f
}

//go:nosplit
func validPrint(s unsafe.Pointer, n uintptr) bool {
	if n == 0 {
		return true
	}

	i := uintptr(0)
	p := *(*unsafe.Pointer)(s)

	for (n - i) >= 8 {
		x := *(*uint64)(unsafe.Pointer(uintptr(p) + i))
		if ((x & 0x8080808080808080) != 0) || hasLess64(x, 0x20) {
			return false
		}
		i += 8
	}

	if (n - i) >= 4 {
		x := *(*uint32)(unsafe.Pointer(uintptr(p) + i))
		if ((x & 0x80808080) != 0) || hasLess32(x, 0x20) {
			return false
		}
		i += 4
	}

	if (n - i) >= 2 {
		x := *(*uint16)(unsafe.Pointer(uintptr(p) + i))
		if ((x & 0x8080) != 0) || hasLess16(x, 0x20) {
			return false
		}
		i += 2
	}

	if i < n {
		x := *(*uint8)(unsafe.Pointer(uintptr(p) + i))
		if ((x & 0x80) != 0) || x < 0x20 {
			return false
		}
	}

	return true
}

// https://graphics.stanford.edu/~seander/bithacks.html#HasLessInWord
const (
	hasLessConstL64 = (^uint64(0)) / 255
	hasLessConstR64 = hasLessConstL64 * 128

	hasLessConstL32 = (^uint32(0)) / 255
	hasLessConstR32 = hasLessConstL32 * 128
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
func hasLess16(x, n uint16) bool {
	return ((x >> 8) < n) || ((x & 0xff) < n)
}

//go:generate go run valid_asm.go -out valid_amd64.s -stubs valid_amd64.go
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
	i := uintptr(0)
	p := *(*unsafe.Pointer)(s)

	for n >= 8 {
		if ((*(*uint64)(unsafe.Pointer(uintptr(p) + i))) & 0x8080808080808080) != 0 {
			return false
		}
		i += 8
		n -= 8
	}

	if n >= 4 {
		if ((*(*uint32)(unsafe.Pointer(uintptr(p) + i))) & 0x80808080) != 0 {
			return false
		}
		i += 4
		n -= 4
	}

	var x uint32
	switch n {
	case 3:
		x = uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + i))) | uint32(*(*uint16)(unsafe.Pointer(uintptr(p) + i + 1)))<<8
	case 2:
		x = uint32(*(*uint16)(unsafe.Pointer(uintptr(p) + i)))
	case 1:
		x = uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + i)))
	default:
		return true
	}
	return (x & 0x80808080) == 0
}

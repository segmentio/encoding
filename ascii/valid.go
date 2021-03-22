//go:generate go run valid_asm.go -out valid_amd64.s -stubs valid_amd64.go
package ascii

import "unsafe"

// Valid returns true if b contains only ASCII characters.
func Valid(b []byte) bool {
	return len(b) == 0 || valid(unsafe.Pointer(&b), uintptr(len(b)))
}

// ValidString returns true if s contains only ASCII characters.
func ValidString(s string) bool {
	return len(s) == 0 || valid(unsafe.Pointer(&s), uintptr(len(s)))
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

	if n >= 16 {
		if asm.valid16((*byte)(p), n/16) == 0 {
			return false
		}
		i = (n / 16) * 16
	}

	if (n - i) >= 8 {
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

	var x uint32
	switch n - i {
	case 3:
		x = uint32(*(*uint16)(unsafe.Pointer(uintptr(p) + i))) | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + i + 2)))<<16
	case 2:
		x = uint32(*(*uint16)(unsafe.Pointer(uintptr(p) + i)))
	case 1:
		x = uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + i)))
	default:
		return true
	}
	return (x & 0x80808080) == 0
}

//go:nosplit
func valid16(s *byte, n uintptr) int {
	p := unsafe.Pointer(s)
	i := uintptr(0)

	for n > 0 {
		lo := *(*uint64)(unsafe.Pointer(uintptr(p) + i))
		hi := *(*uint64)(unsafe.Pointer(uintptr(p) + i + 8))

		if (lo&0x8080808080808080) != 0 || (hi&0x8080808080808080) != 0 {
			return 0
		}

		i += 16
		n--
	}

	return 1
}

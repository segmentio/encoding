//go:generate go run valid_asm.go -out valid_amd64.s -stubs valid_amd64.go
package ascii

import "unsafe"

// Valid returns true if b contains only ASCII characters.
func Valid(b []byte) bool {
	return ValidString(unsafeString(b))
}

// ValidBytes returns true if b is an ASCII character.
func ValidByte(b byte) bool {
	return b <= 0x7f
}

// ValidBytes returns true if b is an ASCII character.
func ValidRune(r rune) bool {
	return r <= 0x7f
}

// ValidString returns true if s contains only ASCII characters.
func ValidString(s string) bool {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&s))
	n := uintptr(len(s))

	if n >= 8 {
		if n > 32 && asm.validAVX2 != nil {
			if asm.validAVX2((*byte)(p), n) == 0 {
				return false
			}
			k := (n / 16) * 16
			p = unsafe.Pointer(uintptr(p) + k)
			n -= k
		}

		for n > 8 {
			if (*(*uint64)(p) & 0x8080808080808080) != 0 {
				return false
			}
			p = unsafe.Pointer(uintptr(p) + 8)
			n -= 8
		}

		if n == 8 {
			return (*(*uint64)(p) & 0x8080808080808080) == 0
		}
	}

	if n > 4 {
		if (*(*uint32)(p) & 0x80808080) != 0 {
			return false
		}
		p = unsafe.Pointer(uintptr(p) + 4)
		n -= 4
	}

	var x uint32
	switch n {
	case 4:
		x = *(*uint32)(p)
	case 3:
		x = uint32(*(*uint16)(p)) | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + 2)))<<16
	case 2:
		x = uint32(*(*uint16)(p))
	case 1:
		x = uint32(*(*uint8)(p))
	default:
		return true
	}
	return (x & 0x80808080) == 0
}

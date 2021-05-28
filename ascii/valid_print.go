//go:generate go run valid_print_asm.go -out valid_print_amd64.s -stubs valid_print_amd64.go
package ascii

import "unsafe"

// Valid returns true if b contains only printable ASCII characters.
func ValidPrint(b []byte) bool {
	return ValidPrintString(unsafeString(b))
}

// ValidBytes returns true if b is an ASCII character.
func ValidPrintByte(b byte) bool {
	return 0x20 <= b && b <= 0x7e
}

// ValidBytes returns true if b is an ASCII character.
func ValidPrintRune(r rune) bool {
	return 0x20 <= r && r <= 0x7e
}

// ValidString returns true if s contains only printable ASCII characters.
func ValidPrintString(s string) bool {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&s))
	n := uintptr(len(s))

	if n >= 8 {
		if n > 32 && asm.validPrintAVX2 != nil {
			if asm.validPrintAVX2((*byte)(p), n) == 0 {
				return false
			}
			if (n % 16) == 0 {
				return true
			}
			k := (n / 16) * 16
			p = unsafe.Pointer(uintptr(p) + k)
			n -= k
		}

		for n > 8 {
			if hasLess64(*(*uint64)(p), 0x20) || hasMore64(*(*uint64)(p), 0x7e) {
				return false
			}
			p = unsafe.Pointer(uintptr(p) + 8)
			n -= 8
		}

		if n == 8 {
			return !(hasLess64(*(*uint64)(p), 0x20) || hasMore64(*(*uint64)(p), 0x7e))
		}
	}

	if n > 4 {
		if hasLess32(*(*uint32)(p), 0x20) || hasMore32(*(*uint32)(p), 0x7e) {
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
		x = 0x20000000 | uint32(*(*uint16)(p)) | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + 2)))<<16
	case 2:
		x = 0x20200000 | uint32(*(*uint16)(p))
	case 1:
		x = 0x20202000 | uint32(*(*uint8)(p))
	default:
		return true
	}
	return !(hasLess32(x, 0x20) || hasMore32(x, 0x7e))
}
